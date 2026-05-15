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
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			http.ServeFileFS(w, r, staticFS, "index.html")
			return
		}
		f.Close()
		http.FileServerFS(staticFS).ServeHTTP(w, r)
	})

	go s.logBroadcaster()

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

	if req.ConfigPath == "" {
		var profile database.Profile
		if err := database.DB.Where("is_active = ?", true).First(&profile).Error; err != nil {
			jsonError(w, "No active profile selected", http.StatusBadRequest)
			return
		}
		if req.CoreType == "" {
			req.CoreType = profile.CoreType
		}
		var rules []database.RoutingRule
		database.DB.Order("sort_order ASC").Find(&rules)

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
		if profile.GroupUUID == "" {
			jsonError(w, "group_uuid is required", http.StatusBadRequest)
			return
		}
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", profile.GroupUUID).First(&group).Error; err != nil {
			jsonError(w, "Group not found", http.StatusBadRequest)
			return
		}
		profile.SortOrder = database.SortNewScoped(&database.Profile{}, "group_uuid = ?", profile.GroupUUID)
		profile.UUID = database.GenerateUUID()
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
		Links     string `json:"links"`
		GroupUUID string `json:"group_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.GroupUUID == "" {
		jsonError(w, "group_uuid is required", http.StatusBadRequest)
		return
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", req.GroupUUID).First(&group).Error; err != nil {
		jsonError(w, "Group not found", http.StatusBadRequest)
		return
	}

	profiles, err := parser.ParseLinks(strings.Split(req.Links, "\n"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	seq := database.SortNewBatch(&database.Profile{}, "group_uuid = ?", len(profiles), req.GroupUUID)

	for i, profile := range profiles {
		profile.SortOrder = seq[i]
		profile.GroupUUID = req.GroupUUID
		database.DB.Create(profile)
	}

	jsonOK(w, map[string]int{"imported": len(profiles)})
}

func (s *Server) handleProfileImportToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Links     string `json:"links"`
		GroupUUID string `json:"group_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.GroupUUID == "" {
		jsonError(w, "group_uuid is required", http.StatusBadRequest)
		return
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", req.GroupUUID).First(&group).Error; err != nil {
		jsonError(w, "Group not found", http.StatusBadRequest)
		return
	}

	profiles, err := parser.ParseLinks(strings.Split(req.Links, "\n"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	seq := database.SortNewBatch(&database.Profile{}, "group_uuid = ?", len(profiles), req.GroupUUID)

	for i, profile := range profiles {
		profile.SortOrder = seq[i]
		profile.GroupUUID = req.GroupUUID
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

// ========== Groups API ==========

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var groups []database.NodeGroup
		database.DB.Order("sort_order ASC").Find(&groups)
		for i := range groups {
			var count int64
			database.DB.Model(&database.Profile{}).Where("group_uuid = ?", groups[i].UUID).Count(&count)
			groups[i].NodeCount = int(count)
		}
		jsonOK(w, groups)

	case http.MethodPost:
		var group database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if group.UUID == "" {
			group.UUID = database.GenerateUUID()
		}
		group.SortOrder = database.SortNew(&database.NodeGroup{})
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
		UUID       string `json:"uuid"`
		BeforeUUID string `json:"before_uuid"`
		AfterUUID  string `json:"after_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.UUID == "" {
		jsonError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	var beforeOrder, afterOrder *int

	if req.BeforeUUID != "" {
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", req.BeforeUUID).First(&group).Error; err == nil {
			v := group.SortOrder
			beforeOrder = &v
		}
	}
	if req.AfterUUID != "" {
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", req.AfterUUID).First(&group).Error; err == nil {
			v := group.SortOrder
			afterOrder = &v
		}
	}

	newOrder := database.SortInsert(beforeOrder, afterOrder)

	if err := database.DB.Model(&database.NodeGroup{}).Where("uuid = ?", req.UUID).Update("sort_order", newOrder).Error; err != nil {
		jsonError(w, "Failed to reorder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{"status": "reordered", "sort_order": newOrder})
}

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
		var count int64
		database.DB.Model(&database.Profile{}).Where("group_uuid = ?", group.UUID).Count(&count)
		group.NodeCount = int(count)
		jsonOK(w, group)

	case http.MethodPut:
		var updated database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		updated.ID = group.ID
		if updated.UUID == "" {
			updated.UUID = group.UUID
		}
		updated.SortOrder = group.SortOrder
		if err := database.DB.Save(&updated).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, updated)

	case http.MethodDelete:
		var count int64
		database.DB.Model(&database.NodeGroup{}).Count(&count)
		if count <= 1 {
			jsonError(w, "Cannot delete the last group", http.StatusBadRequest)
			return
		}
		// 先查出被删节点的 UUID 列表，用于清理策略组脏引用
		var deletedProfileUUIDs []string
		database.DB.Model(&database.Profile{}).Where("group_uuid = ?", group.UUID).Pluck("uuid", &deletedProfileUUIDs)
		// 删除该分组下的所有节点
		database.DB.Where("group_uuid = ?", group.UUID).Delete(&database.Profile{})
		// 清理 StrategyGroup 中的脏引用
		if len(deletedProfileUUIDs) > 0 {
			deletedSet := make(map[string]bool, len(deletedProfileUUIDs))
			for _, uid := range deletedProfileUUIDs {
				deletedSet[uid] = true
			}
			var strategyGroups []database.StrategyGroup
			database.DB.Find(&strategyGroups)
			for _, sg := range strategyGroups {
				if sg.ProfileUUIDs == "" {
					continue
				}
				var uuids []string
				if err := json.Unmarshal([]byte(sg.ProfileUUIDs), &uuids); err != nil {
					continue
				}
				var cleaned []string
				for _, uid := range uuids {
					if !deletedSet[uid] {
						cleaned = append(cleaned, uid)
					}
				}
				if len(cleaned) != len(uuids) {
					newJSON, _ := json.Marshal(cleaned)
					database.DB.Model(&sg).Update("profile_uuids", string(newJSON))
				}
			}
		}
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

	var req struct {
		GroupUUID string `json:"group_uuid"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var profiles []database.Profile
	query := database.DB.Order("sort_order ASC")
	if req.GroupUUID != "" {
		query = query.Where("group_uuid = ?", req.GroupUUID)
	}
	query.Find(&profiles)

	seen := make(map[string]bool)
	var duplicates []uint

	for _, p := range profiles {
		key := p.RawLink
		if idx := strings.LastIndex(key, "#"); idx != -1 {
			key = key[:idx]
		}
		if key == "" {
			key = fmt.Sprintf("%s:%d:%s", p.ProxyAddress, p.ProxyPort, p.ProxyProtocol)
			if p.ProxyCredential != "" {
				key += ":" + p.ProxyCredential
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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	groupUUID := r.FormValue("group_uuid")
	if groupUUID == "" {
		jsonError(w, "group_uuid is required", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			jsonError(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		links, decodeErr := parser.DecodeQRFromBytes(data)
		if decodeErr != nil {
			jsonError(w, "No QR code found in image: "+decodeErr.Error(), http.StatusBadRequest)
			return
		}
		importParsedLinksWithGroup(w, links, groupUUID)
		return
	}

	imageURL := r.FormValue("url")
	if imageURL == "" {
		jsonError(w, "No image file or URL provided", http.StatusBadRequest)
		return
	}

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

	importParsedLinksWithGroup(w, links, groupUUID)
}

func importParsedLinksWithGroup(w http.ResponseWriter, links []string, groupUUID string) {
	if groupUUID == "" {
		jsonError(w, "group_uuid is required", http.StatusBadRequest)
		return
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", groupUUID).First(&group).Error; err != nil {
		jsonError(w, "Group not found", http.StatusBadRequest)
		return
	}

	profiles, err := parser.ParseLinks(links)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	seq := database.SortNewBatch(&database.Profile{}, "group_uuid = ?", len(profiles), groupUUID)

	for i, profile := range profiles {
		profile.SortOrder = seq[i]
		profile.GroupUUID = groupUUID
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
	cores := s.updater.GetLocalCores()
	jsonOK(w, cores)
}

func (s *Server) handleCoresCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cores := s.updater.CheckAllUpdates()
	latestVersions := make(map[string]string)
	for _, c := range cores {
		if c.LatestVer != "" {
			ver := strings.TrimPrefix(c.LatestVer, "v")
			latestVersions[c.Name] = ver
		}
	}
	jsonOK(w, map[string]interface{}{"latest_versions": latestVersions})
}

func (s *Server) handleCoresDetectVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.detectCoreVersions()
	jsonOK(w, map[string]string{"status": "detecting"})
}

func (s *Server) detectCoreVersions() {
	go func() {
		cores := s.updater.GetLocalCoresWithVersions()
		versions := make(map[string]string)
		for _, c := range cores {
			if c.Version != "" {
				versions[c.Name] = c.Version
			}
		}
		s.broadcastToAll(map[string]interface{}{"type": "core_versions", "payload": versions})
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
	if _, exists := s.activeDownloads.Load(req.CoreName); exists {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

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
			s.broadcastToAll(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
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
	if req.CoreName == "" || req.DownloadURL == "" {
		jsonError(w, "Missing core_name or download_url", http.StatusBadRequest)
		return
	}
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
			s.broadcastToAll(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s from URL: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName, "url": req.DownloadURL})
}

func (s *Server) handleCoreUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	binName := coreName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	destPath := filepath.Join(coreDir, binName)

	fileName := strings.ToLower(header.Filename)
	isArchive := strings.HasSuffix(fileName, ".zip") || strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz")

	if isArchive {
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
		if err := s.updater.ExtractBinary(tmpPath, header.Filename, destPath, binName); err != nil {
			jsonError(w, "Failed to extract binary from archive: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
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

	if runtime.GOOS != "windows" {
		os.Chmod(destPath, 0755)
	}

	log.Printf("Uploaded core: %s (%s)", coreName, header.Filename)
	jsonOK(w, map[string]string{"status": "uploaded", "core": coreName, "path": destPath})
}

// ========== Profile by ID API ==========

func (s *Server) handleProfileByID(w http.ResponseWriter, r *http.Request) {
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

	if len(parts) > 1 {
		switch parts[1] {
		case "select":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			database.DB.Model(&database.Profile{}).Where("is_active = ?", true).Update("is_active", false)
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
		if updated.GroupUUID == "" {
			jsonError(w, "group_uuid is required", http.StatusBadRequest)
			return
		}
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", updated.GroupUUID).First(&group).Error; err != nil {
			jsonError(w, "Group not found", http.StatusBadRequest)
			return
		}
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
		rule.SortOrder = database.SortNew(&database.RoutingRule{})
		rule.UUID = database.GenerateUUID()
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
		jsonOK(w, map[string]interface{}{"enabled": false, "port": s.cfg.SocksPort})
	case http.MethodPost:
		var req struct {
			Enabled bool `json:"enabled"`
			Port    int  `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		jsonOK(w, map[string]interface{}{"enabled": req.Enabled, "port": req.Port})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Settings API ==========

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
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
		group.SortOrder = database.SortNew(&database.StrategyGroup{})
		group.UUID = database.GenerateUUID()
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

	s.sendStatusToClient(wc)

	for {
		_, _, err := wc.ReadMessage()
		if err != nil {
			break
		}
	}
}

// ========== WebSocket Broadcasting ==========

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

func (s *Server) broadcastStatus() {
	statuses := s.coreMgr.GetAllStatus()
	s.broadcastToAll(map[string]interface{}{"type": "status", "payload": statuses})
}

func (s *Server) sendStatusToClient(wc *wsConn) {
	statuses := s.coreMgr.GetAllStatus()
	wc.WriteJSON(map[string]interface{}{"type": "status", "payload": statuses})
}

func (s *Server) logBroadcaster() {
	logChan := s.coreMgr.LogChannel()
	for entry := range logChan {
		s.broadcastToAll(map[string]interface{}{"type": "log", "payload": entry})
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
