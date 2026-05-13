package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"v2rayn-go/config"
	"v2rayn-go/configbuilder"
	"v2rayn-go/core"
	"v2rayn-go/database"
	"v2rayn-go/parser"
	"v2rayn-go/subscription"
	"v2rayn-go/updater"

	"github.com/gorilla/websocket"
)

// wsConn 封装 WebSocket 连接，加写锁防止并发写入 panic
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsConn) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

func (w *wsConn) ReadMessage() (messageType int, p []byte, err error) {
	return w.conn.ReadMessage()
}

func (w *wsConn) Close() error {
	return w.conn.Close()
}

// downloadState 下载状态
type downloadState struct {
	CoreName   string `json:"core_name"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Percentage int    `json:"percentage"`
	Status     string `json:"status"` // "downloading", "complete", "error"
	Error      string `json:"error,omitempty"`
}

// Server Web 服务器
type Server struct {
	cfg             *config.AppConfig
	coreMgr         *core.CoreAdminManager
	subSvc          *subscription.Service
	pingSvc         *subscription.PingService
	updater         *updater.Updater
	upgrader        websocket.Upgrader
	wsClients       sync.Map
	activeDownloads sync.Map // map[string]*downloadState
}

// NewServer 创建 Web 服务器
func NewServer(cfg *config.AppConfig, coreMgr *core.CoreAdminManager) *Server {
	return &Server{
		cfg:     cfg,
		coreMgr: coreMgr,
		subSvc:  subscription.NewService(),
		pingSvc: subscription.NewPingService(),
		updater: updater.NewUpdater(cfg),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Start 启动 Web 服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/core/start", s.handleCoreStart)
	mux.HandleFunc("/api/core/stop", s.handleCoreStop)
	mux.HandleFunc("/api/core/status", s.handleCoreStatus)

	mux.HandleFunc("/api/profiles/import-image", s.handleProfileImportImage)
	mux.HandleFunc("/api/profiles/import-to-group", s.handleProfileImportToGroup)
	mux.HandleFunc("/api/profiles/import", s.handleProfileImport)
	mux.HandleFunc("/api/profiles/dedup", s.handleProfileDedup)
	mux.HandleFunc("/api/profiles/ping-all", s.handlePingAll)
	mux.HandleFunc("/api/profiles/reorder", s.handleProfilesReorder)
	mux.HandleFunc("/api/profiles/", s.handleProfileByID)
	mux.HandleFunc("/api/profiles", s.handleProfiles)

	mux.HandleFunc("/api/groups/reorder", s.handleGroupsReorder)
	mux.HandleFunc("/api/groups/", s.handleGroupByID)
	mux.HandleFunc("/api/groups", s.handleGroups)

	mux.HandleFunc("/api/strategy-groups/", s.handleStrategyGroupByID)
	mux.HandleFunc("/api/strategy-groups", s.handleStrategyGroups)

	mux.HandleFunc("/api/cores", s.handleCores)
	mux.HandleFunc("/api/cores/check-updates", s.handleCoresCheckUpdates)
	mux.HandleFunc("/api/cores/detect-versions", s.handleCoresDetectVersions)
	mux.HandleFunc("/api/cores/download-url", s.handleCoreDownloadURL)
	mux.HandleFunc("/api/cores/download", s.handleCoreDownload)
	mux.HandleFunc("/api/cores/upload", s.handleCoreUpload)

	mux.HandleFunc("/api/settings", s.handleSettings)

	mux.HandleFunc("/api/routing-rules/", s.handleRoutingRuleByID)
	mux.HandleFunc("/api/routing-rules", s.handleRoutingRules)

	mux.HandleFunc("/api/proxy/system", s.handleSystemProxy)

	mux.HandleFunc("/api/ws", s.handleWebSocket)

	// 静态文件服务 (go:embed)
	staticFS, err := fs.Sub(StaticFiles, "dist")
	if err != nil {
		return fmt.Errorf("failed to load embedded files: %w", err)
	}

	// 对于非 API 请求，返回前端页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 尝试从嵌入的文件系统中提供文件
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// SPA fallback: 返回 index.html
			http.ServeFileFS(w, r, staticFS, "index.html")
			return
		}
		f.Close()
		http.FileServerFS(staticFS).ServeHTTP(w, r)
	})

	// 启动日志广播器（将内核日志广播给所有 WebSocket 客户端）
	go s.logBroadcaster()

	// API 请求日志中间件
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("[API] %s %s", r.Method, r.URL.Path)
		}
		mux.ServeHTTP(w, r)
	})

	addr := s.cfg.GetListenAddr()
	log.Printf("Web server starting on http://%s", addr)
	return http.ListenAndServe(addr, handler)
}

// ========== Core API ==========

func (s *Server) handleCoreStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CoreType   string `json:"core_type"`
		ConfigPath string `json:"config_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 如果没有指定配置路径，生成一个
	if req.ConfigPath == "" {
		// 获取当前激活的节点
		var profile database.Profile
		if err := database.DB.Where("is_active = ?", true).First(&profile).Error; err != nil {
			jsonError(w, "No active profile selected", http.StatusBadRequest)
			return
		}

		// 如果请求未指定内核类型，使用节点保存的 core_type
		if req.CoreType == "" {
			req.CoreType = profile.CoreType
		}

		// 获取路由规则
		var rules []database.RoutingRule
		database.DB.Order("sort_order ASC").Find(&rules)

		// 生成配置
		var configPath string
		var configErr error
		switch req.CoreType {
		case "xray":
			configPath, configErr = configbuilder.SaveXrayConfig(&profile, rules, s.cfg.AppDir, s.cfg.SocksPort, s.cfg.HTTPPort)
		case "sing-box":
			configPath, configErr = configbuilder.SaveSingboxConfig(&profile, rules, s.cfg.AppDir, s.cfg.SocksPort)
		default:
			jsonError(w, "Unsupported core type", http.StatusBadRequest)
			return
		}
		if configErr != nil {
			jsonError(w, configErr.Error(), http.StatusInternalServerError)
			return
		}
		req.ConfigPath = configPath
	}

	if err := s.coreMgr.StartCore(core.CoreType(req.CoreType), req.ConfigPath); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "started"})
	go s.broadcastStatus()
}

func (s *Server) handleCoreStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CoreType string `json:"core_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CoreType == "" {
		req.CoreType = "xray"
	}

	if err := s.coreMgr.StopCore(core.CoreType(req.CoreType)); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "stopped"})
	go s.broadcastStatus()
}

func (s *Server) handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	statuses := s.coreMgr.GetAllStatus()
	jsonOK(w, statuses)
}

// ========== Profile API ==========

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var profiles []database.Profile
		database.DB.Order("sort_order ASC").Find(&profiles)
		jsonOK(w, profiles)

	case http.MethodPost:
		var profile database.Profile
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := database.DB.Create(&profile).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, profile)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProfileImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Links   string `json:"links"`
		GroupID uint   `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profiles, err := parser.ParseLinks(strings.Split(req.Links, "\n"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取分组信息
	var groupName string
	if req.GroupID > 0 {
		var group database.NodeGroup
		if err := database.DB.First(&group, req.GroupID).Error; err == nil {
			groupName = group.Alias
		}
	}

	// 获取当前最大排序号
	var maxOrder int
	database.DB.Model(&database.Profile{}).Where("group_id = ?", req.GroupID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	for i, profile := range profiles {
		profile.SortOrder = maxOrder + i + 1
		profile.GroupID = req.GroupID
		profile.GroupName = groupName
		database.DB.Create(profile)
	}

	jsonOK(w, map[string]int{"imported": len(profiles)})
}

// handleProfileImportToGroup 导入节点到指定分组
func (s *Server) handleProfileImportToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Links   string `json:"links"`
		GroupID uint   `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profiles, err := parser.ParseLinks(strings.Split(req.Links, "\n"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取分组信息
	var groupName string
	if req.GroupID > 0 {
		var group database.NodeGroup
		if err := database.DB.First(&group, req.GroupID).Error; err == nil {
			groupName = group.Alias
		}
	}

	// 获取当前最大排序号
	var maxOrder int
	database.DB.Model(&database.Profile{}).Where("group_id = ?", req.GroupID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	for i, profile := range profiles {
		profile.SortOrder = maxOrder + i + 1
		profile.GroupID = req.GroupID
		profile.GroupName = groupName
		database.DB.Create(profile)
	}

	jsonOK(w, map[string]int{"imported": len(profiles)})
}

func (s *Server) handlePingAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go s.pingSvc.PingAllProfiles(r.Context(), 20)
	jsonOK(w, map[string]string{"status": "pinging"})
}

func (s *Server) handleProfilesReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UUIDs []string `json:"uuids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tx := database.DB.Begin()
	for i, uuid := range req.UUIDs {
		if err := tx.Model(&database.Profile{}).Where("uuid = ?", uuid).Update("sort_order", i).Error; err != nil {
			tx.Rollback()
			jsonError(w, "Failed to reorder: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		jsonError(w, "Failed to commit reorder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "reordered"})
}

// ========== Groups API ==========

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var groups []database.NodeGroup
		database.DB.Order("sort_order ASC").Find(&groups)

		// 计算每个分组的节点数
		for i := range groups {
			var count int64
			database.DB.Model(&database.Profile{}).Where("group_id = ?", groups[i].ID).Count(&count)
			groups[i].NodeCount = int(count)
		}

		jsonOK(w, groups)

	case http.MethodPost:
		var group database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// 自动生成 UUID
		if group.UUID == "" {
			group.UUID = database.GenerateUUID()
		}

		// 设置排序
		var maxOrder int
		database.DB.Model(&database.NodeGroup{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)
		group.SortOrder = maxOrder + 1

		if err := database.DB.Create(&group).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, group)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGroupsReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UUIDs []string `json:"uuids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tx := database.DB.Begin()
	for i, uuid := range req.UUIDs {
		if err := tx.Model(&database.NodeGroup{}).Where("uuid = ?", uuid).Update("sort_order", i).Error; err != nil {
			tx.Rollback()
			jsonError(w, "Failed to reorder: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		jsonError(w, "Failed to commit reorder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "reordered"})
}

// handleGroupByID handles /api/groups/{id} and /api/groups/{id}/refresh
func (s *Server) handleGroupByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	parts := strings.SplitN(path, "/", 2)
	id := strings.TrimSpace(parts[0])

	if id == "" {
		jsonError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	var group database.NodeGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		jsonError(w, "Group not found", http.StatusNotFound)
		return
	}

	// 检查子操作
	if len(parts) > 1 {
		switch parts[1] {
		case "refresh":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !group.IsSubscription {
				jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
				return
			}
			go func() {
				if err := s.subSvc.UpdateGroupSubscription(&group, false); err != nil {
					log.Printf("Failed to refresh group %s: %v", group.Alias, err)
				}
			}()
			jsonOK(w, map[string]string{"status": "refreshing"})
			return

		case "refresh-proxy":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !group.IsSubscription {
				jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
				return
			}
			go func() {
				if err := s.subSvc.UpdateGroupSubscription(&group, true); err != nil {
					log.Printf("Failed to refresh group %s via proxy: %v", group.Alias, err)
				}
			}()
			jsonOK(w, map[string]string{"status": "refreshing"})
			return

		default:
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		// 计算节点数
		var count int64
		database.DB.Model(&database.Profile{}).Where("group_id = ?", group.ID).Count(&count)
		group.NodeCount = int(count)
		jsonOK(w, group)

	case http.MethodPut:
		var updated database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		updated.ID = group.ID
		// 保留 UUID 不变
		if updated.UUID == "" {
			updated.UUID = group.UUID
		}
		// 保留原始 sort_order，编辑操作不改变排序
		updated.SortOrder = group.SortOrder
		if err := database.DB.Save(&updated).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, updated)

	case http.MethodDelete:
		// 检查是否只剩一个分组
		var count int64
		database.DB.Model(&database.NodeGroup{}).Count(&count)
		if count <= 1 {
			jsonError(w, "Cannot delete the last group", http.StatusBadRequest)
			return
		}
		// 清除该分组下节点的分组引用
		database.DB.Model(&database.Profile{}).Where("group_id = ?", group.ID).Updates(map[string]interface{}{"group_id": 0, "group_name": ""})
		database.DB.Delete(&group)
		jsonOK(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Profile Dedup ==========

func (s *Server) handleProfileDedup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 接收可选的 group_id 参数
	var req struct {
		GroupID uint `json:"group_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var profiles []database.Profile
	query := database.DB.Order("sort_order ASC")
	if req.GroupID > 0 {
		query = query.Where("group_id = ?", req.GroupID)
	}
	query.Find(&profiles)

	seen := make(map[string]bool)
	var duplicates []uint

	for _, p := range profiles {
		// 基于 raw_link 去重（去掉 #名称 部分，只比较配置）
		key := p.RawLink
		if idx := strings.LastIndex(key, "#"); idx != -1 {
			key = key[:idx]
		}
		if key == "" {
			// fallback: address + port + protocol + uuid
			key = fmt.Sprintf("%s:%d:%s", p.Address, p.Port, p.Protocol)
			if p.UUID != "" {
				key += ":" + p.UUID
			}
		}
		if seen[key] {
			duplicates = append(duplicates, p.ID)
		} else {
			seen[key] = true
		}
	}

	if len(duplicates) > 0 {
		database.DB.Delete(&database.Profile{}, duplicates)
	}

	jsonOK(w, map[string]interface{}{
		"removed": len(duplicates),
		"total":   len(profiles),
	})
}

// ========== Profile Import from Image ==========

func (s *Server) handleProfileImportImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var imageURL string

	// Get group_id from form
	groupIDStr := r.FormValue("group_id")
	var groupID uint
	fmt.Sscanf(groupIDStr, "%d", &groupID)

	// Check for uploaded file
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		// Read file bytes
		data, err := io.ReadAll(file)
		if err != nil {
			jsonError(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		// Try to decode QR from image bytes
		links, decodeErr := parser.DecodeQRFromBytes(data)
		if decodeErr != nil {
			jsonError(w, "No QR code found in image: "+decodeErr.Error(), http.StatusBadRequest)
			return
		}
		importParsedLinksWithGroup(w, links, groupID)
		return
	}

	// Check for image URL
	imageURL = r.FormValue("url")
	if imageURL == "" {
		jsonError(w, "No image file or URL provided", http.StatusBadRequest)
		return
	}

	// Download image from URL
	resp, err := http.Get(imageURL)
	if err != nil {
		jsonError(w, "Failed to download image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		jsonError(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	links, decodeErr := parser.DecodeQRFromBytes(data)
	if decodeErr != nil {
		jsonError(w, "No QR code found in image: "+decodeErr.Error(), http.StatusBadRequest)
		return
	}

	importParsedLinksWithGroup(w, links, groupID)
}

func importParsedLinks(w http.ResponseWriter, links []string) {
	importParsedLinksWithGroup(w, links, 0)
}

func importParsedLinksWithGroup(w http.ResponseWriter, links []string, groupID uint) {
	profiles, err := parser.ParseLinks(links)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var groupName string
	if groupID > 0 {
		var group database.NodeGroup
		if err := database.DB.First(&group, groupID).Error; err == nil {
			groupName = group.Alias
		}
	}

	var maxOrder int
	database.DB.Model(&database.Profile{}).Where("group_id = ?", groupID).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	for i, profile := range profiles {
		profile.SortOrder = maxOrder + i + 1
		profile.GroupID = groupID
		profile.GroupName = groupName
		database.DB.Create(profile)
	}

	jsonOK(w, map[string]int{"imported": len(profiles)})
}

// ========== Core Hub API ==========

func (s *Server) handleCores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 只返回本地信息，不访问网络，毫秒级响应
	cores := s.updater.GetLocalCores()
	jsonOK(w, cores)
}

func (s *Server) handleCoresCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查所有内核的最新版本（访问 GitHub API，支持镜像降级）
	cores := s.updater.CheckAllUpdates()
	latestVersions := make(map[string]string)
	for _, c := range cores {
		if c.LatestVer != "" {
			// 统一去掉 v 前缀，确保前端比较一致
			ver := strings.TrimPrefix(c.LatestVer, "v")
			latestVersions[c.Name] = ver
		}
	}
	jsonOK(w, map[string]interface{}{
		"latest_versions": latestVersions,
	})
}

func (s *Server) handleCoresDetectVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 触发异步版本检测，结果通过 WebSocket 推送
	s.detectCoreVersions()
	jsonOK(w, map[string]string{"status": "detecting"})
}

// detectCoreVersions 异步检测所有已安装内核的版本号并通过 WebSocket 推送
func (s *Server) detectCoreVersions() {
	go func() {
		cores := s.updater.GetLocalCoresWithVersions()
		versions := make(map[string]string)
		for _, c := range cores {
			if c.Version != "" {
				versions[c.Name] = c.Version
			}
		}
		s.broadcastToAll(map[string]interface{}{
			"type":    "core_versions",
			"payload": versions,
		})
	}()
}

func (s *Server) handleCoreDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CoreName string `json:"core_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CoreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}

	// 检查是否已在下载中
	if _, exists := s.activeDownloads.Load(req.CoreName); exists {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	// 初始化下载状态
	state := &downloadState{CoreName: req.CoreName, Status: "downloading"}
	s.activeDownloads.Store(req.CoreName, state)

	go func() {
		defer s.activeDownloads.Delete(req.CoreName)

		err := s.updater.DownloadCore(req.CoreName, func(downloaded, total int64) {
			state.Downloaded = downloaded
			state.Total = total
			if total > 0 {
				state.Percentage = int(downloaded * 100 / total)
			}
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_progress",
				"payload": state,
			})
		})

		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_complete",
				"payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()},
			})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_complete",
				"payload": map[string]interface{}{"core_name": req.CoreName, "success": true},
			})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName})
}

func (s *Server) handleCoreDownloadURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CoreName    string `json:"core_name"`
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CoreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}
	if req.DownloadURL == "" {
		jsonError(w, "Missing download_url", http.StatusBadRequest)
		return
	}

	// 检查是否已在下载中
	if _, exists := s.activeDownloads.Load(req.CoreName); exists {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	state := &downloadState{CoreName: req.CoreName, Status: "downloading"}
	s.activeDownloads.Store(req.CoreName, state)

	go func() {
		defer s.activeDownloads.Delete(req.CoreName)

		err := s.updater.DownloadCoreFromURL(req.CoreName, req.DownloadURL, func(downloaded, total int64) {
			state.Downloaded = downloaded
			state.Total = total
			if total > 0 {
				state.Percentage = int(downloaded * 100 / total)
			}
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_progress",
				"payload": state,
			})
		})

		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s from URL: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_complete",
				"payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()},
			})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{
				"type":    "download_complete",
				"payload": map[string]interface{}{"core_name": req.CoreName, "success": true},
			})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName, "url": req.DownloadURL})
}

func (s *Server) handleCoreUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 200MB)
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	coreName := r.FormValue("core_name")
	if coreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("binary")
	if err != nil {
		jsonError(w, "Missing binary file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Map core name to sub-directory
	subDir := coreName
	switch coreName {
	case "xray":
		subDir = "xray"
	case "sing-box":
		subDir = "sing_box"
	case "mihomo":
		subDir = "mihomo"
	}
	coreDir := filepath.Join(s.cfg.BinDir, subDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		jsonError(w, "Failed to create core directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine binary name
	binName := coreName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	destPath := filepath.Join(coreDir, binName)

	// Check if uploaded file is an archive
	fileName := strings.ToLower(header.Filename)
	isArchive := strings.HasSuffix(fileName, ".zip") || strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz")

	if isArchive {
		// Save archive to temp file first
		tmpFile, err := os.CreateTemp("", "v2rayn-upload-*.tmp")
		if err != nil {
			jsonError(w, "Failed to create temp file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, file); err != nil {
			tmpFile.Close()
			jsonError(w, "Failed to save temp file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpFile.Close()

		// Extract binary from archive
		if err := s.updater.ExtractBinary(tmpPath, header.Filename, destPath, binName); err != nil {
			jsonError(w, "Failed to extract binary from archive: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Save as raw binary
		dst, err := os.Create(destPath)
		if err != nil {
			jsonError(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			jsonError(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Set executable permission on non-Windows
	if runtime.GOOS != "windows" {
		os.Chmod(destPath, 0755)
	}

	log.Printf("Uploaded core: %s (%s)", coreName, header.Filename)
	jsonOK(w, map[string]string{"status": "uploaded", "core": coreName, "path": destPath})
}

// ========== Profile by ID API ==========

func (s *Server) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/profiles/{id} or /api/profiles/{id}/select or /api/profiles/{id}/ping
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	parts := strings.SplitN(path, "/", 2)
	id := strings.TrimSpace(parts[0])

	if id == "" {
		jsonError(w, "Missing profile ID", http.StatusBadRequest)
		return
	}

	var profile database.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		jsonError(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Check for sub-action
	if len(parts) > 1 {
		switch parts[1] {
		case "select":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			// Deactivate all profiles first
			database.DB.Model(&database.Profile{}).Where("is_active = ?", true).Update("is_active", false)
			// Activate selected profile
			database.DB.Model(&profile).Update("is_active", true)
			jsonOK(w, map[string]string{"status": "selected"})
			return

		case "ping":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			go s.pingSvc.PingSingleProfile(&profile)
			jsonOK(w, map[string]string{"status": "pinging"})
			return

		default:
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		jsonOK(w, profile)

	case http.MethodPut:
		var updated database.Profile
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		updated.ID = profile.ID
		// 保留原始 sort_order，编辑操作不改变排序
		updated.SortOrder = profile.SortOrder
		if err := database.DB.Save(&updated).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, updated)

	case http.MethodDelete:
		database.DB.Delete(&profile)
		jsonOK(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Routing Rules API ==========

func (s *Server) handleRoutingRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var rules []database.RoutingRule
		database.DB.Order("sort_order ASC").Find(&rules)
		jsonOK(w, rules)

	case http.MethodPost:
		var rule database.RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := database.DB.Create(&rule).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, rule)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRoutingRuleByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/routing-rules/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/routing-rules/")
	var rule database.RoutingRule
	if err := database.DB.First(&rule, path).Error; err != nil {
		jsonError(w, "Rule not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var updated database.RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		updated.ID = rule.ID
		if err := database.DB.Save(&updated).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, updated)

	case http.MethodDelete:
		database.DB.Delete(&rule)
		jsonOK(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== System Proxy API ==========

func (s *Server) handleSystemProxy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOK(w, map[string]interface{}{
			"enabled": false,
			"port":    s.cfg.SocksPort,
		})

	case http.MethodPost:
		var req struct {
			Enabled bool `json:"enabled"`
			Port    int  `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// System proxy management is platform-specific and handled at OS level
		jsonOK(w, map[string]interface{}{
			"enabled": req.Enabled,
			"port":    req.Port,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Settings API ==========

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 返回当前配置
		settings := map[string]interface{}{
			"listen_ip":     s.cfg.ListenIP,
			"web_port":      s.cfg.WebPort,
			"socks_port":    s.cfg.SocksPort,
			"http_port":     s.cfg.HTTPPort,
			"outbound_ip":   s.cfg.OutboundIP,
			"github_mirror": s.cfg.GitHubMirror,
		}
		jsonOK(w, settings)

	case http.MethodPost:
		// 更新配置
		var req struct {
			ListenIP     *string `json:"listen_ip"`
			SocksPort    *int    `json:"socks_port"`
			HTTPPort     *int    `json:"http_port"`
			OutboundIP   *string `json:"outbound_ip"`
			GitHubMirror *string `json:"github_mirror"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.ListenIP != nil {
			s.cfg.ListenIP = *req.ListenIP
		}
		if req.SocksPort != nil && *req.SocksPort > 0 {
			s.cfg.SocksPort = *req.SocksPort
		}
		if req.HTTPPort != nil && *req.HTTPPort > 0 {
			s.cfg.HTTPPort = *req.HTTPPort
		}
		if req.OutboundIP != nil {
			s.cfg.OutboundIP = *req.OutboundIP
		}
		if req.GitHubMirror != nil {
			s.cfg.GitHubMirror = *req.GitHubMirror
		}

		// 保存到 config.json
		if err := s.cfg.SaveJSONConfig(); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonOK(w, map[string]string{"status": "saved"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Strategy Groups API ==========

func (s *Server) handleStrategyGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var groups []database.StrategyGroup
		database.DB.Order("sort_order ASC").Find(&groups)
		jsonOK(w, groups)

	case http.MethodPost:
		var group database.StrategyGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := database.DB.Create(&group).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, group)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStrategyGroupByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/strategy-groups/")
	id := strings.TrimSpace(path)

	if id == "" {
		jsonError(w, "Missing strategy group ID", http.StatusBadRequest)
		return
	}

	var group database.StrategyGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		jsonError(w, "Strategy group not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		jsonOK(w, group)

	case http.MethodPut:
		var updated database.StrategyGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		updated.ID = group.ID
		if err := database.DB.Save(&updated).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, updated)

	case http.MethodDelete:
		database.DB.Delete(&group)
		jsonOK(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== WebSocket ==========

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	wc := &wsConn{conn: conn}
	defer wc.Close()

	clientID := fmt.Sprintf("%p", conn)
	s.wsClients.Store(clientID, wc)
	defer s.wsClients.Delete(clientID)

	// 连接时立即发送当前内核状态
	s.sendStatusToClient(wc)

	// 保持连接
	for {
		_, _, err := wc.ReadMessage()
		if err != nil {
			break
		}
	}
}

// ========== WebSocket Broadcasting ==========

// broadcastToAll 向所有 WebSocket 客户端广播消息
func (s *Server) broadcastToAll(msg interface{}) {
	s.wsClients.Range(func(key, value interface{}) bool {
		if wc, ok := value.(*wsConn); ok {
			if err := wc.WriteJSON(msg); err != nil {
				s.wsClients.Delete(key)
			}
		}
		return true
	})
}

// broadcastStatus 广播当前所有内核状态给 WebSocket 客户端
func (s *Server) broadcastStatus() {
	statuses := s.coreMgr.GetAllStatus()
	s.broadcastToAll(map[string]interface{}{
		"type":    "status",
		"payload": statuses,
	})
}

// sendStatusToClient 向单个 WebSocket 客户端发送当前状态
func (s *Server) sendStatusToClient(wc *wsConn) {
	statuses := s.coreMgr.GetAllStatus()
	wc.WriteJSON(map[string]interface{}{
		"type":    "status",
		"payload": statuses,
	})
}

// logBroadcaster 从内核日志通道读取日志并广播给所有 WebSocket 客户端
func (s *Server) logBroadcaster() {
	logChan := s.coreMgr.LogChannel()
	for entry := range logChan {
		s.broadcastToAll(map[string]interface{}{
			"type":    "log",
			"payload": entry,
		})
	}
}

// ========== Helpers ==========

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
