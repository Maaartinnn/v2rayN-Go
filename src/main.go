package main

import (
	"log/slog"
	"os"

	"v2rayn-go/cmd"
	"v2rayn-go/config"
	"v2rayn-go/coredef"
	"v2rayn-go/database"
)

func main() {
	// 初始化日志（默认 info 级别，输出到 stderr）
	coredef.InitLogger("info", os.Stderr)

	// 初始化配置
	cfg := config.DefaultConfig()
	if err := cfg.Init(); err != nil {
		slog.Error("failed to init config", "error", err)
		os.Exit(1)
	}

	// 注意：admin 子命令在 flag.Parse() 之前拦截，避免 flag 解析冲突
	// 如果是 admin 命令，提前初始化数据库并执行，然后退出
	if len(os.Args) >= 2 && os.Args[1] == "admin" {
		if err := database.Init(cfg); err != nil {
			slog.Error("failed to init database", "error", err)
			os.Exit(1)
		}
		defer database.Close()
		cmd.RunAdmin(os.Args[2:]) // 传入 admin 后面的子参数
		return
	}

	// 按优先级加载配置：CLI 参数 > JSON 配置文件 > 默认值
	if err := cfg.LoadWithPriority(); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 注意：数据库初始化（Init + SeedDefaults）在 sysmgr.RunDirect / App.run 中执行
	// 避免 main.go 与 sysmgr 重复调用 Init

	// 执行 CLI 命令
	cmd.Run(cfg)
}
