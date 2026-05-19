package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// AppConfig 应用全局配置
type AppConfig struct {
	// AppDir 应用工作目录（可执行文件所在目录）
	AppDir string `json:"-"`
	// DBPath 数据库文件路径
	DBPath string `json:"-"`
	// BinDir 外挂内核目录
	BinDir string `json:"-"`
	// LogDir 日志目录
	LogDir string `json:"-"`

	// === 可配置的网络参数 ===
	// 注意：以下字段不使用 omitzero，因为用户可能显式设置为零值
	//（如 WebPort=0 表示禁用 Web UI），omitzero 会导致这些设置
	// 在保存时丢失，下次加载时被 DefaultConfig() 覆盖。

	// WebPort Web 界面端口
	WebPort int `json:"web_port"`
	// ListenIP 监听 IP 地址
	ListenIP string `json:"listen_ip"`
	// SocksPort SOCKS5 代理端口
	SocksPort int `json:"socks_port"`
	// HTTPPort HTTP 代理端口
	HTTPPort int `json:"http_port"`
	// OutboundIP 出站绑定 IP
	OutboundIP string `json:"outbound_ip"`

	// GitHubMirror GitHub 下载镜像地址
	// 默认值为空字符串，omitzero 不会造成信息损失
	GitHubMirror string `json:"github_mirror,omitzero"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		WebPort:    2017,
		ListenIP:   "127.0.0.1",
		SocksPort:  10808,
		HTTPPort:   10809,
		OutboundIP: "0.0.0.0",
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

// LoadWithPriority 按优先级加载配置：CLI 参数 > JSON 配置文件 > 默认值
// 优先级：
//  1. 命令行启动参数（最高优先级）
//  2. 同级目录 config.json
//  3. 系统默认值
func (c *AppConfig) LoadWithPriority() error {
	// 第一步：加载 JSON 配置文件（覆盖默认值）
	c.loadJSONConfig()

	// 第二步：解析 CLI 参数（覆盖 JSON 配置）
	c.parseCLIFlags()

	return nil
}

// loadJSONConfig 从应用目录下的 config.json 加载配置
// json.Unmarshal 只覆盖 JSON 中存在的字段，缺失字段保留 DefaultConfig() 的值
func (c *AppConfig) loadJSONConfig() {
	configPath := filepath.Join(c.AppDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// config.json 不存在不是错误，使用默认值
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to read config.json: %v", err)
		}
		return
	}

	if err := json.Unmarshal(data, c); err != nil {
		log.Printf("Warning: failed to parse config.json: %v", err)
		return
	}

	log.Printf("Loaded config from %s", configPath)
}

// parseCLIFlags 解析命令行参数（最高优先级）
func (c *AppConfig) parseCLIFlags() {
	// 定义 CLI 标志
	listenIP := flag.String("listen-ip", "", "Listen IP address (e.g. 127.0.0.1, 0.0.0.0)")
	port := flag.Int("port", 0, "Web UI port")
	socksPort := flag.Int("socks-port", 0, "SOCKS5 proxy port")
	httpPort := flag.Int("http-port", 0, "HTTP proxy port")
	outboundIP := flag.String("outbound-ip", "", "Outbound bind IP")
	githubMirror := flag.String("github-mirror", "", "GitHub mirror URL for downloading cores")

	flag.Parse()

	// 仅覆盖非零值（CLI 参数优先）
	if *listenIP != "" {
		c.ListenIP = *listenIP
	}
	if *port > 0 {
		c.WebPort = *port
	}
	if *socksPort > 0 {
		c.SocksPort = *socksPort
	}
	if *httpPort > 0 {
		c.HTTPPort = *httpPort
	}
	if *outboundIP != "" {
		c.OutboundIP = *outboundIP
	}
	if *githubMirror != "" {
		c.GitHubMirror = *githubMirror
	}
}

// SaveJSONConfig 将当前配置保存到 config.json
// json:"-" 的字段自动隐藏；GitHubMirror 为零值时 omitzero 自动省略
func (c *AppConfig) SaveJSONConfig() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(c.AppDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}

	return nil
}

// GetListenAddr 返回完整的监听地址
func (c *AppConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.ListenIP, c.WebPort)
}
