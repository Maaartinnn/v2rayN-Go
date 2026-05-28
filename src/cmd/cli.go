package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"v2rayn-go/config"
	"v2rayn-go/sysmgr"
)

// Run 解析命令行参数并执行相应操作
func Run(cfg *config.AppConfig) {
	// 如果没有参数，直接以前台模式运行
	if len(os.Args) < 2 {
		if err := sysmgr.RunDirect(cfg); err != nil {
			slog.Error("failed to run", "error", err)
			os.Exit(1)
		}
		return
	}

	command := os.Args[1]

	switch command {
	case "install":
		if err := sysmgr.InstallService(cfg); err != nil {
			slog.Error("failed to install service", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service installed. Use 'v2rayN-Go start' to start it.")

	case "uninstall":
		if err := sysmgr.UninstallService(cfg); err != nil {
			slog.Error("failed to uninstall service", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled.")

	case "start":
		if err := sysmgr.StartService(cfg); err != nil {
			slog.Error("failed to start service", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service started.")

	case "stop":
		if err := sysmgr.StopService(cfg); err != nil {
			slog.Error("failed to stop service", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service stopped.")

	case "restart":
		if err := sysmgr.RestartService(cfg); err != nil {
			slog.Error("failed to restart service", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service restarted.")

	case "daemon":
		// 以系统服务模式运行（由服务管理器调用）
		if err := sysmgr.RunAsService(cfg); err != nil {
			slog.Error("failed to run as service", "error", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`v2rayN-Go - A lightweight proxy control center

Usage:
  v2rayN-Go [command] [flags]

Commands:
  (no args)     Run in foreground mode (for development/debugging)
  install       Register as system service with auto-start
  uninstall     Remove system service
  start         Start the system service
  stop          Stop the system service
  restart       Restart the system service
  daemon        Run as system service (called by service manager)
  help          Show this help message

Flags (highest priority, override config.json):
  --listen-ip string     Listen IP address (e.g. 127.0.0.1, 0.0.0.0)
  --port int             Web UI port (default 2017)
  --socks-port int       SOCKS5 proxy port (default 10808)
  --http-port int        HTTP proxy port (default 10809)
  --outbound-ip string   Outbound bind IP (default 0.0.0.0)
  --github-mirror string GitHub mirror URL for downloading cores

Config priority (highest to lowest):
  1. CLI flags (--listen-ip, --port, etc.)
  2. config.json (in the same directory as the executable)
  3. System defaults

Examples:
  v2rayN-Go                              # Run with defaults
  v2rayN-Go --listen-ip 0.0.0.0 --port 8080
  v2rayN-Go install                      # Install as system service
  v2rayN-Go start                        # Start the service`)
}
