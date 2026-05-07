package config

import (
	"os"
	"path/filepath"
)

// AppConfig 应用全局配置
type AppConfig struct {
	// AppDir 应用工作目录（可执行文件所在目录）
	AppDir string
	// DBPath 数据库文件路径
	DBPath string
	// BinDir 外挂内核目录
	BinDir string
	// LogDir 日志目录
	LogDir string
	// WebPort Web 界面端口
	WebPort int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		WebPort: 2017,
	}
}

// Init 初始化配置，基于可执行文件所在目录
func (c *AppConfig) Init() error {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	c.AppDir = filepath.Dir(exePath)

	// 设置各路径
	c.DBPath = filepath.Join(c.AppDir, "config.db")
	c.BinDir = filepath.Join(c.AppDir, "bin")
	c.LogDir = filepath.Join(c.AppDir, "logs")

	// 确保必要目录存在
	dirs := []string{c.BinDir, c.LogDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
