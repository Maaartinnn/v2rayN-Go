package main

import (
	"log"

	"v2rayn-go/cmd"
	"v2rayn-go/config"
)

func main() {
	// 初始化配置
	cfg := config.DefaultConfig()
	if err := cfg.Init(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	// 按优先级加载配置：CLI 参数 > JSON 配置文件 > 默认值
	if err := cfg.LoadWithPriority(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 执行 CLI 命令
	cmd.Run(cfg)
}
