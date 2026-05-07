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

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 执行 CLI 命令
	cmd.Run(cfg)
}
