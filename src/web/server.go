package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"v2rayn-go/config"
	"v2rayn-go/core"
	"v2rayn-go/service"
)

// Server Web 服务器 — 纯 DI 容器与路由总线
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
	}
}

// Start 启动 Web 服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 1. 创建 WSHandler（它同时实现 StatusBroadcaster 接口）
	wsHandler := NewWSHandler(s.coreSvc, s.coreMgr)

	// 2. 实例化各业务 Handler 并显式注入依赖
	coreHandler := NewCoreHandler(s.coreSvc, wsHandler)
	profileHandler := NewProfileHandler(s.profileSvc, s.pingSvc)
	groupHandler := NewGroupHandler(s.groupSvc)
	strategyHandler := NewStrategyGroupHandler(s.strategySvc)
	routingHandler := NewRoutingRuleHandler(s.routingSvc)
	settingsHandler := NewSettingsHandler(s.settingsSvc, s.cfg)

	// 3. 注册路由
	coreHandler.Register(mux)
	profileHandler.Register(mux)
	groupHandler.Register(mux)
	strategyHandler.Register(mux)
	routingHandler.Register(mux)
	settingsHandler.Register(mux)
	wsHandler.Register(mux)

	// 4. 静态文件服务 (go:embed)
	staticFS, err := fs.Sub(StaticFiles, "dist")
	if err != nil {
		return fmt.Errorf("failed to load embedded files: %w", err)
	}

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

	// 5. 启动日志广播（使用 context.Background 支持优雅退出）
	go wsHandler.LogBroadcaster(context.Background())

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
	json.NewEncoder(w).Encode(map[string]interface{}{"error": msg, "code": code})
}

// decodeJSON 从请求体解码 JSON 到 v，失败时自动写入 400 错误响应并返回 false。
// 调用方只需 `if !decodeJSON(w, r, &req) { return }`。
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return false
	}
	return true
}
