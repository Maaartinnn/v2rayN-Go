package main

import (
	"log/slog"
	"os"

	"v2rayn-go/cmd"
	"v2rayn-go/config"
	"v2rayn-go/coredef"
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

	// 按优先级加载配置：CLI 参数 > JSON 配置文件 > 默认值
	if err := cfg.LoadWithPriority(); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 执行 CLI 命令
	cmd.Run(cfg)
}
