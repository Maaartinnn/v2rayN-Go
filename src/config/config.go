package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"v2rayn-go/coredef"
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

	// backupOnce 保证 BackupConfig 在多次调用时只执行一次
	backupOnce sync.Once
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		WebPort:    coredef.DefaultWebPort,
		ListenIP:   coredef.DefaultListenIP,
		SocksPort:  coredef.DefaultSocksPort,
		HTTPPort:   coredef.DefaultHTTPPort,
		OutboundIP: coredef.DefaultOutboundIP,
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
//
// 断电容灾逻辑：
//   - 文件不存在 → 使用默认值（正常情况）
//   - 文件为空或 JSON 解析失败 → 自动从 config.json.bak 恢复
//   - 恢复时使用原子写入，确保恢复操作本身不会引入新的损坏风险
func (c *AppConfig) loadJSONConfig() {
	configPath := filepath.Join(c.AppDir, "config.json")
	bakPath := configPath + ".bak"

	data, err := os.ReadFile(configPath)
	if err != nil || len(data) == 0 {
		// 文件不存在（首次运行）→ 使用默认值，不是错误
		if os.IsNotExist(err) {
			return
		}
		// 文件为空（0KB，可能由断电导致）或读取失败 → 尝试 .bak 恢复
		slog.Warn("config.json missing or empty, attempting backup restore", "error", err)
		if c.tryRestoreFromBackup(configPath, bakPath) {
			return
		}
		return
	}

	// 文件存在且非空，尝试解析
	if err := json.Unmarshal(data, c); err != nil {
		// JSON 格式损坏（可能由用户手动编辑出错或断电写入残缺导致）
		slog.Warn("config.json parse error, attempting backup restore", "error", err)
		if c.tryRestoreFromBackup(configPath, bakPath) {
			return
		}
		slog.Warn("no valid backup available, falling back to defaults")
		return
	}

	slog.Info("loaded config", "path", configPath)
}

// tryRestoreFromBackup 尝试从 .bak 文件恢复配置。
// 恢复成功时返回 true，同时将 .bak 内容原子写回 config.json。
// 恢复失败时返回 false（调用方应回退到默认配置）。
func (c *AppConfig) tryRestoreFromBackup(configPath, bakPath string) bool {
	bakData, err := os.ReadFile(bakPath)
	if err != nil || len(bakData) == 0 {
		slog.Warn("backup file not available", "path", bakPath, "error", err)
		return false
	}

	// 验证 .bak 内容是合法 JSON
	var probe AppConfig
	if err := json.Unmarshal(bakData, &probe); err != nil {
		slog.Warn("backup file also corrupted", "path", bakPath, "error", err)
		return false
	}

	// .bak 内容合法 → 应用到当前配置（逐字段赋值，避免复制 backupOnce 等不可复制字段）
	c.AppDir = probe.AppDir
	c.DBPath = probe.DBPath
	c.BinDir = probe.BinDir
	c.LogDir = probe.LogDir
	c.WebPort = probe.WebPort
	c.ListenIP = probe.ListenIP
	c.SocksPort = probe.SocksPort
	c.HTTPPort = probe.HTTPPort
	c.OutboundIP = probe.OutboundIP
	c.GitHubMirror = probe.GitHubMirror
	// 将 .bak 内容原子写回 config.json（统一使用原子写入，避免恢复操作本身引入断电风险）
	if err := AtomicWriteFile(configPath, bakData, 0644); err != nil {
		slog.Warn("failed to restore config.json from backup", "error", err)
		// 即使写回失败，内存中的配置已恢复，程序可以正常运行
	}

	slog.Warn("config restored from backup", "path", bakPath)
	return true
}

// BackupConfig 将当前生效的 config.json 备份为 config.json.bak。
// 此方法应在应用成功启动后调用（配置已通过完整验证），
// 使用 sync.Once 保证即使被多次调用也只执行一次。
//
// 设计意图：
//   - 只有经过完整验证的配置才会写入 .bak，避免脏数据污染备份
//   - 使用原子写入 .bak 文件，确保备份操作本身不会因断电产生损坏
func (c *AppConfig) BackupConfig() {
	c.backupOnce.Do(func() {
		configPath := filepath.Join(c.AppDir, "config.json")
		bakPath := configPath + ".bak"

		data, err := os.ReadFile(configPath)
		if err != nil {
			slog.Warn("failed to read config for backup", "error", err)
			return
		}
		if len(data) == 0 {
			return
		}

		if err := AtomicWriteFile(bakPath, data, 0644); err != nil {
			slog.Warn("failed to create config backup", "error", err)
			return
		}
		slog.Debug("config backup created", "path", bakPath)
	})
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

// SaveJSONConfig 将当前配置保存到 config.json（断电安全版本）。
//
// 采用原子写入策略：写临时文件 → 强制 Sync 落盘 → 重命名替换。
// 即使在写入过程中遭遇断电，原文件也不会损坏（os.Rename 在主流
// 文件系统中是原子操作）。
//
// json:"-" 的字段自动隐藏；GitHubMirror 为零值时 omitzero 自动省略。
func (c *AppConfig) SaveJSONConfig() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(c.AppDir, "config.json")
	if err := AtomicWriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config.json: %w", err)
	}

	return nil
}

// AtomicWriteFile 将数据以原子方式写入文件，保证在断电或崩溃等
// 异常情况下目标文件不会损坏（不会出现 0KB 或写入一半的情况）。
//
// 实现原理：
//  1. 在目标文件同目录下创建临时文件（同目录保证 os.Rename 不会跨分区失败）
//  2. 将数据写入临时文件
//  3. 调用 f.Sync() 强制将数据从操作系统 Page Cache 刷入物理磁盘
//     （这一步是防断电的关键：仅靠 os.WriteFile 的数据仍在内存缓冲区中，
//     断电时缓冲区内容会丢失）
//  4. 关闭临时文件（Windows 下必须先关闭才能 Rename）
//  5. 使用 os.Rename 将临时文件原子替换为目标文件
//
// 任何步骤失败时，临时文件会被 defer 自动清理，不会留下残留。
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	// 1. 获取目标目录，确保临时文件和目标文件在同一文件系统分区
	dir := filepath.Dir(filename)

	// 2. 创建临时文件
	f, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	tmpName := f.Name()

	// success 标志：只有在 rename 成功后才设为 true
	// defer 兜底清理：无论成功失败都关闭文件，但只有失败时才删除临时文件
	success := false
	defer func() {
		// f.Close() 允许多次调用（第二次返回 nil），此处兜底处理异常路径
		f.Close()
		if !success {
			os.Remove(tmpName)
		}
	}()

	// 3. 写入数据
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write data failed: %w", err)
	}

	// 4. 核心防御：强制将数据从 Page Cache 刷入物理磁盘
	// 没有这一步，数据可能仅存在于操作系统内存缓冲区中，
	// 此时断电会导致数据丢失，即使 os.Rename 本身是原子的。
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// 5. 设置文件权限（CreateTemp 默认权限是 0600，需要显式修改）
	if err := f.Chmod(perm); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	// 6. 必须先关闭文件，再进行重命名
	// 特别是 Windows 上，打开状态的文件无法被覆盖或重命名
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file failed: %w", err)
	}

	// 7. OS 级别的原子替换：在主流文件系统（ext4/xfs/ntfs）上，
	// os.Rename 是原子操作——要么完全成功，要么不发生，绝不会出现中间状态
	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	success = true
	return nil
}

// GetListenAddr 返回完整的监听地址
func (c *AppConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.ListenIP, c.WebPort)
}
