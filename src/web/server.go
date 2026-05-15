package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"

	"v2rayn-go/config"
	"v2rayn-go/core"
	"v2rayn-go/database"
	"v2rayn-go/parser"
	"v2rayn-go/service"
	"v2rayn-go/subscription"

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
	cfg *config.AppConfig

	// 业务 Service 层
	profileSvc  *service.ProfileService
	groupSvc    *service.GroupService
	strategySvc *service.StrategyGroupService
	routingSvc  *service.RoutingRuleService
	coreSvc     *service.CoreService
	settingsSvc *service.SettingsService

	// 保留的直接依赖
	coreMgr *core.CoreAdminManager
	pingSvc service.PingServiceInterface

	upgrader        websocket.Upgrader
	wsClients       sync.Map
	activeDownloads sync.Map // map[string]*downloadState
}

// PingServiceInterface 是 ping 服务的接口（由 subscription 包实现）
type PingServiceInterface = service.PingServiceInterface

// NewServer 创建 Web 服务器
func NewServer(cfg *config.AppConfig, coreMgr *core.CoreAdminManager) *Server {
	coreSvc := service.NewCoreService(cfg, coreMgr)
	return &Server{
		cfg:         cfg,
		profileSvc:  service.NewProfileService(),
		groupSvc:    service.NewGroupService(),
		strategySvc: service.NewStrategyGroupService(),
		routingSvc:  service.NewRoutingRuleService(),
		coreSvc:     coreSvc,
		settingsSvc: service.NewSettingsService(cfg),
		coreMgr:     coreMgr,
		pingSvc:     service.NewPingService(),
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

	if err := s.coreSvc.Start(req.CoreType, req.ConfigPath); err != nil {
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

	if err := s.coreSvc.Stop(req.CoreType); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "stopped"})
	go s.broadcastStatus()
}

func (s *Server) handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	statuses := s.coreSvc.GetAllStatus()
	jsonOK(w, statuses)
}

// ========== Profile API ==========

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		profiles, err := s.profileSvc.List()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, profiles)

	case http.MethodPost:
		var profile database.Profile
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.profileSvc.Create(&profile); err != nil {
			if strings.Contains(err.Error(), "not found") {
				jsonError(w, err.Error(), http.StatusBadRequest)
			} else {
				jsonError(w, err.Error(), http.StatusInternalServerError)
			}
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

	count, err := s.profileSvc.ImportLinks(req.Links, req.GroupUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonOK(w, map[string]int{"imported": count})
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
		groups, err := s.groupSvc.List()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, groups)

	case http.MethodPost:
		var group database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.groupSvc.Create(&group); err != nil {
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

	newOrder, err := s.groupSvc.Reorder(req.UUID, req.BeforeUUID, req.AfterUUID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{"status": "reordered", "sort_order": newOrder})
}

func (s *Server) handleGroupByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	parts := strings.SplitN(path, "/", 2)
	uuid := strings.TrimSpace(parts[0])

	if uuid == "" {
		jsonError(w, "Missing group UUID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "refresh":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			group, err := s.groupSvc.Get(uuid)
			if err != nil {
				jsonError(w, err.Error(), http.StatusNotFound)
				return
			}
			if !group.IsSubscription {
				jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
				return
			}
			go func() {
				// 订阅刷新仍由 subscription 包处理
				subSvc := subscription.NewService()
				if err := subSvc.UpdateGroupSubscription(group, false); err != nil {
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
			group, err := s.groupSvc.Get(uuid)
			if err != nil {
				jsonError(w, err.Error(), http.StatusNotFound)
				return
			}
			if !group.IsSubscription {
				jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
				return
			}
			go func() {
				subSvc := subscription.NewService()
				if err := subSvc.UpdateGroupSubscription(group, true); err != nil {
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
		group, err := s.groupSvc.Get(uuid)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, group)

	case http.MethodPut:
		var updated database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := s.groupSvc.Update(uuid, &updated)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, result)

	case http.MethodDelete:
		if err := s.groupSvc.Delete(uuid); err != nil {
			if strings.Contains(err.Error(), "last group") || strings.Contains(err.Error(), "not found") {
				jsonError(w, err.Error(), http.StatusBadRequest)
			} else {
				jsonError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := s.profileSvc.Dedup(req.GroupUUID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"removed": result.Removed,
		"total":   result.Total,
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
		s.importParsedLinks(w, links, groupUUID)
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

	s.importParsedLinks(w, links, groupUUID)
}

// importParsedLinks 将已解析的链接导入到指定分组
func (s *Server) importParsedLinks(w http.ResponseWriter, links []string, groupUUID string) {
	count, err := s.profileSvc.ImportParsedLinks(links, groupUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonOK(w, map[string]int{"imported": count})
}

// ========== Core Hub API ==========

func (s *Server) handleCores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cores := s.coreSvc.GetLocalCores()
	jsonOK(w, cores)
}

func (s *Server) handleCoresCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	latestVersions := s.coreSvc.CheckUpdates()
	jsonOK(w, map[string]interface{}{"latest_versions": latestVersions})
}

func (s *Server) handleCoresDetectVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.coreSvc.DetectVersions(func(versions map[string]string) {
		s.broadcastToAll(map[string]interface{}{"type": "core_versions", "payload": versions})
	})
	jsonOK(w, map[string]string{"status": "detecting"})
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
		err := s.coreSvc.Download(req.CoreName, func(downloaded, total int64) {
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
		err := s.coreSvc.DownloadFromURL(req.CoreName, req.DownloadURL, func(downloaded, total int64) {
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

	destPath, err := s.coreSvc.Upload(coreName, header.Filename, file)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]string{"status": "uploaded", "core": coreName, "path": destPath})
}

// ========== Profile by UUID API ==========

func (s *Server) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	parts := strings.SplitN(path, "/", 2)
	uuid := strings.TrimSpace(parts[0])

	if uuid == "" {
		jsonError(w, "Missing profile UUID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "select":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if err := s.profileSvc.Select(uuid); err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			jsonOK(w, map[string]string{"status": "selected"})
			return

		case "ping":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			profile, err := s.profileSvc.Get(uuid)
			if err != nil {
				jsonError(w, err.Error(), http.StatusNotFound)
				return
			}
			go s.pingSvc.PingSingleProfile(profile)
			jsonOK(w, map[string]string{"status": "pinging"})
			return

		default:
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		profile, err := s.profileSvc.Get(uuid)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, profile)

	case http.MethodPut:
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		profile, err := s.profileSvc.Update(uuid, req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
				jsonError(w, err.Error(), http.StatusBadRequest)
			} else {
				jsonError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		jsonOK(w, profile)

	case http.MethodDelete:
		if err := s.profileSvc.Delete(uuid); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== Routing Rules API ==========

func (s *Server) handleRoutingRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rules, err := s.routingSvc.List()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, rules)

	case http.MethodPost:
		var rule database.RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.routingSvc.Create(&rule); err != nil {
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
	uuid := strings.TrimSpace(path)

	switch r.Method {
	case http.MethodPut:
		var updated database.RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := s.routingSvc.Update(uuid, &updated)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, result)

	case http.MethodDelete:
		if err := s.routingSvc.Delete(uuid); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		settings := s.settingsSvc.GetSettings()
		jsonOK(w, settings)

	case http.MethodPost:
		var req service.UpdateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.settingsSvc.UpdateSettings(&req); err != nil {
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
		groups, err := s.strategySvc.List()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, groups)

	case http.MethodPost:
		var group database.StrategyGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := s.strategySvc.Create(&group); err != nil {
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
	uuid := strings.TrimSpace(path)

	if uuid == "" {
		jsonError(w, "Missing strategy group UUID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		group, err := s.strategySvc.Get(uuid)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, group)

	case http.MethodPut:
		var updated database.StrategyGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := s.strategySvc.Update(uuid, &updated)
		if err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		jsonOK(w, result)

	case http.MethodDelete:
		if err := s.strategySvc.Delete(uuid); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	statuses := s.coreSvc.GetAllStatus()
	s.broadcastToAll(map[string]interface{}{"type": "status", "payload": statuses})
}

func (s *Server) sendStatusToClient(wc *wsConn) {
	statuses := s.coreSvc.GetAllStatus()
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
