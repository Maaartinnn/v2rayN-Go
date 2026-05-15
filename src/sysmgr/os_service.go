package sysmgr

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"v2rayn-go/config"
	"v2rayn-go/core"
	"v2rayn-go/database"
	"v2rayn-go/web"

	"github.com/kardianos/service"
)

// App 是系统服务管理的核心结构
type App struct {
	cfg    *config.AppConfig
	core   *core.CoreAdminManager
	logger service.Logger
}

// NewApp 创建新的 App 实例
func NewApp(cfg *config.AppConfig) *App {
	return &App{
		cfg:  cfg,
		core: core.NewCoreAdminManager(cfg),
	}
}

// Start 实现 service.Interface，服务启动时调用
func (a *App) Start(s service.Service) error {
	log.Println("Service starting...")
	go a.run()
	return nil
}

// run 服务主循环
func (a *App) run() {
	// 初始化数据库
	if err := database.Init(a.cfg); err != nil {
		log.Printf("Failed to init database: %v", err)
		return
	}
	log.Println("Database initialized successfully")

	// 启动 Web 服务器
	webServer := web.NewServer(a.cfg, a.core)
	if err := webServer.Start(); err != nil {
		log.Printf("Web server error: %v", err)
	}
}

// Stop 实现 service.Interface，服务停止时调用
func (a *App) Stop(s service.Service) error {
	log.Println("Service stopping...")

	// 停止所有内核
	a.core.StopAll()

	// 关闭数据库
	database.Close()

	log.Println("Service stopped")
	return nil
}

// GetCoreManager 获取内核管理器
func (a *App) GetCoreManager() *core.CoreAdminManager {
	return a.core
}

// GetServiceConfig 获取系统服务配置
func getServiceConfig() *service.Config {
	exePath, _ := os.Executable()
	return &service.Config{
		Name:        "v2rayn-go",
		DisplayName: "v2rayN-Go",
		Description: "v2rayN-Go - A lightweight proxy control center",
		Executable:  exePath,
		Arguments:   []string{"daemon"},
		Option: service.KeyValue{
			"WorkingDirectory": filepath.Dir(exePath),
		},
	}
}

// InstallService 注册系统服务
func InstallService(cfg *config.AppConfig) error {
	app := NewApp(cfg)
	svcConfig := getServiceConfig()

	s, err := service.New(app, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Install(); err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	log.Println("Service installed successfully")
	return nil
}

// UninstallService 卸载系统服务
func UninstallService(cfg *config.AppConfig) error {
	app := NewApp(cfg)
	svcConfig := getServiceConfig()

	s, err := service.New(app, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	log.Println("Service uninstalled successfully")
	return nil
}

// StartService 启动系统服务
func StartService(cfg *config.AppConfig) error {
	app := NewApp(cfg)
	svcConfig := getServiceConfig()

	s, err := service.New(app, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	log.Println("Service started successfully")
	return nil
}

// StopService 停止系统服务
func StopService(cfg *config.AppConfig) error {
	app := NewApp(cfg)
	svcConfig := getServiceConfig()

	s, err := service.New(app, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	log.Println("Service stopped successfully")
	return nil
}

// RestartService 重启系统服务
func RestartService(cfg *config.AppConfig) error {
	if err := StopService(cfg); err != nil {
		// 停止失败可能是因为服务未运行，忽略错误
		log.Printf("Warning: failed to stop service (may not be running): %v", err)
	}
	return StartService(cfg)
}

// RunAsService 以系统服务模式运行
func RunAsService(cfg *config.AppConfig) error {
	app := NewApp(cfg)
	svcConfig := getServiceConfig()

	s, err := service.New(app, svcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// 设置日志
	app.logger, err = s.Logger(nil)
	if err != nil {
		log.Printf("Warning: failed to create service logger: %v", err)
	}

	return s.Run()
}

// RunDirect 直接运行（非系统服务模式，用于前台调试）
func RunDirect(cfg *config.AppConfig) error {
	app := NewApp(cfg)

	// 初始化数据库
	if err := database.Init(cfg); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()

	log.Printf("v2rayN-Go started in direct mode")
	log.Printf("Web UI: http://127.0.0.1:%d", cfg.WebPort)
	log.Printf("Press Ctrl+C to stop")

	// 启动 Web 服务器
	webServer := web.NewServer(cfg, app.core)
	if err := webServer.Start(); err != nil {
		return fmt.Errorf("web server error: %w", err)
	}

	return nil
}
