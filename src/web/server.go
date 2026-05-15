package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"

	"v2rayn-go/config"
	"v2rayn-go/core"
	"v2rayn-go/service"

	"github.com/gorilla/websocket"
)

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

	// 按业务域注册路由
	s.RegisterCoreRoutes(mux)
	s.RegisterProfileRoutes(mux)
	s.RegisterGroupRoutes(mux)
	s.RegisterStrategyGroupRoutes(mux)
	s.RegisterRoutingRuleRoutes(mux)
	s.RegisterSettingsRoutes(mux)
	s.RegisterWebSocketRoutes(mux)

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
