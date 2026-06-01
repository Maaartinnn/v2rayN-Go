package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"

	"v2rayn-go/config"
	"v2rayn-go/database"
	"v2rayn-go/sysmgr"

	"golang.org/x/crypto/bcrypt"
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
  (no args)          Run in foreground mode (for development/debugging)
  install            Register as system service with auto-start
  uninstall          Remove system service
  start              Start the system service
  stop               Stop the system service
  restart            Restart the system service
  daemon             Run as system service (called by service manager)
  admin set <pwd>    Reset admin password (invalidates all sessions)
  admin random       Generate random admin password (invalidates all sessions)
  help               Show this help message

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
  v2rayN-Go start                        # Start the service
  v2rayN-Go admin set mypassword123       # Reset admin password
  v2rayN-Go admin random                 # Generate random admin password`)
}

// RunAdmin 处理 admin 子命令（在 flag.Parse() 之前拦截，执行完即退出）
// 用法: v2rayN-Go admin <subcommand> [args]
func RunAdmin(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: v2rayN-Go admin <subcommand>")
		fmt.Println()
		fmt.Println("Subcommands:")
		fmt.Println("  set <password>   Reset admin password (invalidates all sessions & disables TOTP)")
		fmt.Println("  random           Generate random admin password (invalidates all sessions & disables TOTP)")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "set":
		if len(args) < 2 {
			fmt.Println("Error: password argument required")
			fmt.Println("Usage: v2rayN-Go admin set <password>")
			os.Exit(1)
		}
		resetAdminPassword(args[1])

	case "random":
		// 生成 16 字节随机密码（hex 编码，32 字符）
		pwdBytes := make([]byte, 16)
		if _, err := rand.Read(pwdBytes); err != nil {
			slog.Error("failed to generate random password", "error", err)
			os.Exit(1)
		}
		plainPassword := hex.EncodeToString(pwdBytes)
		resetAdminPassword(plainPassword)
		fmt.Printf("Generated random password: %s\n", plainPassword)

	default:
		fmt.Printf("Unknown admin subcommand: %s\n", subcommand)
		fmt.Println("Valid subcommands: set, random")
		os.Exit(1)
	}
}

// resetAdminPassword 重置超管密码，同时刷新 JWTSecret（踢掉所有设备）并关闭 TOTP
func resetAdminPassword(newPassword string) {
	// 查找超管（Role=1）
	var admin database.User
	if err := database.DB.Where("role = ?", 1).First(&admin).Error; err != nil {
		fmt.Println("Error: admin user not found. Please start the application at least once first.")
		os.Exit(1)
	}

	// 生成新密码的 bcrypt 哈希
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	// 刷新 JWTSecret（使所有旧 Token 失效）
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		slog.Error("failed to generate JWT secret", "error", err)
		os.Exit(1)
	}
	newJWTSecret := hex.EncodeToString(secretBytes)

	// 原子更新：密码哈希 + JWTSecret + 关闭 TOTP
	updates := map[string]any{
		"password_hash": string(hashedPwd),
		"jwt_secret":    newJWTSecret,
		"totp_secret":   "",
		"totp_enabled":  false,
	}
	if err := database.DB.Model(&admin).Updates(updates).Error; err != nil {
		slog.Error("failed to update admin credentials", "error", err)
		os.Exit(1)
	}

	fmt.Println("Admin credentials updated successfully.")
	fmt.Println("  - Password: updated")
	fmt.Println("  - All active sessions: invalidated")
	fmt.Println("  - TOTP (2FA): disabled")
}
