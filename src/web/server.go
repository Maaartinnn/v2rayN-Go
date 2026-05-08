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
	"v2rayn-go/configbuilder"
	"v2rayn-go/core"
	"v2rayn-go/database"
	"v2rayn-go/parser"
	"v2rayn-go/subscription"

	"github.com/gorilla/websocket"
)

// Server Web 服务器
type Server struct {
	cfg       *config.AppConfig
	coreMgr   *core.CoreAdminManager
	subSvc    *subscription.Service
	pingSvc   *subscription.PingService
	upgrader  websocket.Upgrader
	wsClients sync.Map
}

// NewServer 创建 Web 服务器
func NewServer(cfg *config.AppConfig, coreMgr *core.CoreAdminManager) *Server {
	return &Server{
		cfg:     cfg,
		coreMgr: coreMgr,
		subSvc:  subscription.NewService(),
		pingSvc: subscription.NewPingService(),
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

	mux.HandleFunc("/api/profiles", s.handleProfiles)
	mux.HandleFunc("/api/profiles/import", s.handleProfileImport)
	mux.HandleFunc("/api/profiles/ping-all", s.handlePingAll)

	mux.HandleFunc("/api/subscriptions", s.handleSubscriptions)
	mux.HandleFunc("/api/subscriptions/refresh-all", s.handleRefreshAll)

	mux.HandleFunc("/api/groups", s.handleGroups)
	mux.HandleFunc("/api/profiles/dedup", s.handleProfileDedup)
	mux.HandleFunc("/api/profiles/import-image", s.handleProfileImportImage)

	mux.HandleFunc("/api/settings", s.handleSettings)

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

	addr := s.cfg.GetListenAddr()
	log.Printf("Web server starting on http://%s", addr)
	return http.ListenAndServe(addr, mux)
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

	if req.CoreType == "" {
		req.CoreType = "xray"
	}

	// 如果没有指定配置路径，生成一个
	if req.ConfigPath == "" {
		// 获取当前激活的节点
		var profile database.Profile
		if err := database.DB.Where("is_active = ?", true).First(&profile).Error; err != nil {
			jsonError(w, "No active profile selected", http.StatusBadRequest)
			return
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
		Links string `json:"links"`
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

	// 获取当前最大排序号
	var maxOrder int
	database.DB.Model(&database.Profile{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	for i, profile := range profiles {
		profile.SortOrder = maxOrder + i + 1
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

// ========== Subscription API ==========

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var subs []database.Subscription
		database.DB.Find(&subs)
		jsonOK(w, subs)

	case http.MethodPost:
		var sub database.Subscription
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := database.DB.Create(&sub).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, sub)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRefreshAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go s.subSvc.UpdateAllSubscriptions()
	jsonOK(w, map[string]string{"status": "refreshing"})
}

// ========== Groups API ==========

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var groups []database.NodeGroup
		database.DB.Order("sort_order ASC").Find(&groups)
		jsonOK(w, groups)

	case http.MethodPost:
		var group database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := database.DB.Create(&group).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, group)

	case http.MethodPut:
		var group database.NodeGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if group.ID == 0 {
			jsonError(w, "Missing group ID", http.StatusBadRequest)
			return
		}
		if err := database.DB.Save(&group).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, group)

	case http.MethodDelete:
		var req struct {
			ID uint `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// Clear group references from profiles
		database.DB.Model(&database.Profile{}).Where("group_id = ?", req.ID).Updates(map[string]interface{}{"group_id": 0, "group_name": ""})
		database.DB.Delete(&database.NodeGroup{}, req.ID)
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

	var profiles []database.Profile
	database.DB.Order("sort_order ASC").Find(&profiles)

	seen := make(map[string]bool)
	var duplicates []uint

	for _, p := range profiles {
		// Create a unique key based on address + port + protocol
		key := fmt.Sprintf("%s:%d:%s", p.Address, p.Port, p.Protocol)
		if p.UUID != "" {
			key += ":" + p.UUID
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
		importParsedLinks(w, links)
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

	importParsedLinks(w, links)
}

func importParsedLinks(w http.ResponseWriter, links []string) {
	profiles, err := parser.ParseLinks(links)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var maxOrder int
	database.DB.Model(&database.Profile{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)

	for i, profile := range profiles {
		profile.SortOrder = maxOrder + i + 1
		database.DB.Create(profile)
	}

	jsonOK(w, map[string]int{"imported": len(profiles)})
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
	defer conn.Close()

	clientID := fmt.Sprintf("%p", conn)
	s.wsClients.Store(clientID, conn)
	defer s.wsClients.Delete(clientID)

	// 启动日志转发
	logChan := s.coreMgr.LogChannel()
	go func() {
		for entry := range logChan {
			msg := map[string]interface{}{
				"type":    "log",
				"payload": entry,
			}
			conn.WriteJSON(msg)
		}
	}()

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
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
