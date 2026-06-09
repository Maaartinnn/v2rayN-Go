package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"v2rayn-go/config"
	"v2rayn-go/core"
	"v2rayn-go/database"
	"v2rayn-go/service"
)

// Server Web 服务器 — 纯 DI 容器与路由总线
type Server struct {
	cfg *config.AppConfig

	// 业务 Service 层
	profileSvc  *service.ProfileService
	groupSvc    *service.GroupService
	routingSvc  *service.RoutingRuleService
	coreSvc     *service.CoreService
	settingsSvc *service.SettingsService
	authSvc     *service.AuthService

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
		routingSvc:  service.NewRoutingRuleService(),
		coreSvc:     coreSvc,
		settingsSvc: service.NewSettingsService(cfg),
		authSvc:     service.NewAuthService(),
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
	authHandler := NewAuthHandler(s.authSvc)
	coreHandler := NewCoreHandler(s.coreSvc, wsHandler)
	profileHandler := NewProfileHandler(s.profileSvc, s.coreSvc, s.pingSvc)
	groupHandler := NewGroupHandler(s.groupSvc)
	routingHandler := NewRoutingRuleHandler(s.routingSvc)
	settingsHandler := NewSettingsHandler(s.settingsSvc, s.cfg)

	// 3. 注册路由（auth 路由先注册，白名单在中间件中处理）
	authHandler.Register(mux)
	coreHandler.Register(mux)
	profileHandler.Register(mux)
	groupHandler.Register(mux)
	routingHandler.Register(mux)
	settingsHandler.Register(mux)
	wsHandler.Register(mux)

	// 4. 静态文件服务 (go:embed)
	staticFS, err := fs.Sub(StaticFiles, "dist")
	if err != nil {
		return fmt.Errorf("failed to load embedded files: %w", err)
	}

	// 5. 读取 app_settings 中的服务器配置
	basePath := getSettingFromDB("custom_base_path")

	// 6. 使用 html/template 解析 index.html 模板，注入 custom_base_path
	//
	// html/template 的优势：
	//   - 安全性：自动上下文感知转义（JS 字符串上下文中会安全处理）
	//   - 鲁棒性：不受 Vite 压缩/格式化影响，模板语法 {{ .BasePath }} 在任何位置都有效
	//   - 可扩展性：未来可注入更多环境变量，无需修改替换逻辑
	//   - Option("missingkey=error")：拼写错误时 Fail Fast，而非静默渲染为空字符串
	//
	// basePath 格式为纯路径名（如 "my-secret"），注入时需要加上 "/" 前缀
	injectVal := basePath
	if injectVal != "" {
		injectVal = "/" + injectVal
	}

	tmpl, err := template.New("index.html").Option("missingkey=error").ParseFS(StaticFiles, "dist/index.html")
	if err != nil {
		return fmt.Errorf("failed to parse index.html template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"BasePath": injectVal}); err != nil {
		return fmt.Errorf("failed to execute index.html template: %w", err)
	}
	modifiedIndexHTML := buf.Bytes()
	slog.Info("index.html base path injected", "base_path", injectVal)

	// 7. 注册精确路由（Go 1.22+ {$} 语法）
	// 哈希路由模式下，浏览器只请求根路径和 /assets/... 静态资源，无需 SPA 路由降级
	//
	// indexHandler：精确匹配根路径，返回注入了 base path 的 index.html
	indexHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(modifiedIndexHTML)
	}

	// "{$}" 精确匹配根路径，不再匹配 /assets 等子路径
	mux.HandleFunc("GET {$}", indexHandler)
	if injectVal != "" {
		// 精确匹配带前缀的根路径（如 /my-secret/）
		mux.HandleFunc("GET "+injectVal+"/{$}", indexHandler)
	}
	// 静态文件 fallback（/assets/xxx.js、/favicon.svg 等）
	mux.Handle("GET /", http.FileServerFS(staticFS))

	// 8. 启动日志广播（使用 context.Background 支持优雅退出）
	go wsHandler.LogBroadcaster(context.Background())

	// 9. Auth 中间件包装（拦截非白名单的 /api/ 请求）
	authedMux := AuthMiddleware(s.authSvc)(mux)

	// 10. 日志中间件
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			slog.Info("API request", "method", r.Method, "path", r.URL.Path)
		}
		authedMux.ServeHTTP(w, r)
	})

	// 11. 读取 force_https 配置
	forceHTTPS := getSettingFromDB("force_https")

	// 12. 动态路由前缀包装
	handler := withBasePath(basePath, innerHandler)

	addr := s.cfg.GetListenAddr()

	// 13. 根据 force_https 选择启动模式
	if forceHTTPS == "true" {
		certDir := filepath.Join(s.cfg.AppDir, "certs")
		return s.startHTTPS(handler, certDir)
	}

	slog.Info("web server starting", "addr", addr)
	return http.ListenAndServe(addr, handler)
}

// ========== Helpers ==========

// ReorderRequest 通用重排序请求（三个列表共用）
type ReorderRequest struct {
	UUID       string `json:"uuid"`
	BeforeUUID string `json:"before_uuid"`
	AfterUUID  string `json:"after_uuid"`
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": msg, "code": code})
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

// getSettingFromDB 从 app_settings 表读取指定 key 的值
// 用于启动时读取服务器级配置（force_https, custom_base_path 等）
func getSettingFromDB(key string) string {
	var setting struct {
		Value string
	}
	if err := database.DB.Table("app_settings").Where("key = ?", key).Select("value").Scan(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

// withBasePath 为 HTTP handler 添加自定义路由前缀
// basePath 存储格式为纯路径名（无斜杠），如 "my-secret"，空字符串表示无前缀
// 当 basePath 为空时直接返回原 handler（无额外开销）
// 否则在匹配请求路径时自动加上 "/" 前缀进行比对，并剥离后转发给内部 handler
//
// 例如 basePath="my-secret" 时：
//
//	/my-secret/api/core/status → /api/core/status
//	/my-secret/ → /
//	/other → 404
func withBasePath(basePath string, handler http.Handler) http.Handler {
	// 规范化：去除首尾斜杠，新格式存储无斜杠纯路径名
	prefix := strings.Trim(basePath, "/")
	if prefix == "" {
		return handler
	}
	prefix = "/" + prefix // 统一为 "/my-secret" 格式进行路径比对

	slog.Info("route prefix enabled", "prefix", prefix)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 匹配前缀 + 前缀后紧跟 "/" 或精确匹配
		if strings.HasPrefix(r.URL.Path, prefix+"/") || r.URL.Path == prefix {
			// 剥离前缀：/my-secret/api/... → /api/...
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			handler.ServeHTTP(w, r)
		} else if r.URL.Path == "/" {
			// 根路径重定向到 basePath
			http.Redirect(w, r, prefix+"/", http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	})
}
