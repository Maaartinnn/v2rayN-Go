// settings_service.go — 配置管理业务逻辑层
//
// 职责：
//   - GetSettings(): 返回当前配置快照（只读）
//   - UpdateSettings(): 接收前端单字段/多字段局部更新，带校验和脏标记
//
// 设计要点：
//   - UpdateSettingsRequest 使用指针类型（*int, *string, *bool），区分"未传"和"传了零值"
//   - dirty flag 模式：只有字段值真正改变时才触发 AtomicWriteFile，避免不必要的 Sync() 开销
//   - 三步拦截（判空→判变→判合法）：确保非法数据永远不触碰内存和磁盘
//   - 前端通过失焦（Blur）触发保存，每次只发送变更的字段

package service

import (
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"sync"

	"v2rayn-go/config"
	"v2rayn-go/database"

	"gorm.io/gorm/clause"
)

// basePathPattern 路由前缀合法性校验：仅允许字母、数字、下划线、连字符
// 不允许斜杠（"/"）、点号（"."）或其他特殊字符
var basePathPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// SettingsService 配置管理业务逻辑层
type SettingsService struct {
	cfg *config.AppConfig

	// 应用设置的内存缓存，避免每次 JWT 签发都查库
	// 由 GetSettingFast 读取，由 UpdateSettings 写入更新
	cacheMu   sync.RWMutex
	cacheData map[string]string
}

// NewSettingsService 创建配置服务
func NewSettingsService(cfg *config.AppConfig) *SettingsService {
	return &SettingsService{
		cfg:       cfg,
		cacheData: make(map[string]string), // 提前分配，避免 nil 判断
	}
}

// GetSettingFast 带双重检查锁定（DCL）的缓存优先读取。
//
// 首次调用时从 DB 回填缓存，后续调用 RWMutex.RLock + map 读取（纳秒级，零 DB I/O）。
// DCL 防止并发请求瞬间穿透到 DB：只有第一个穿透请求会查库，其余等待并从缓存读取。
//
// 使用场景：JWT 签发等高频调用路径，避免每次都查 app_settings 表。
func (s *SettingsService) GetSettingFast(key string) string {
	// 第一次检查：快速路径，无锁读取
	s.cacheMu.RLock()
	if v, ok := s.cacheData[key]; ok {
		s.cacheMu.RUnlock()
		return v
	}
	s.cacheMu.RUnlock()

	// 缓存未命中，加写锁
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// 第二次检查（DCL）：加锁期间可能已被其他 goroutine 回填
	if v, ok := s.cacheData[key]; ok {
		return v
	}

	// 穿透到 DB，回填缓存
	var setting database.AppSetting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return ""
	}
	s.cacheData[key] = setting.Value
	return setting.Value
}

// GetSettings 获取当前配置（合并 config.json 字段 + app_settings 数据库表）
func (s *SettingsService) GetSettings() map[string]any {
	// 从 app_settings 表读取服务器级配置
	appSettings := getAllAppSettings()

	return map[string]any{
		// config.json 字段
		"listen_ip":         s.cfg.ListenIP,
		"web_port":          s.cfg.WebPort,
		"socks_port":        s.cfg.SocksPort,
		"http_port":         s.cfg.HTTPPort,
		"outbound_ip":       s.cfg.OutboundIP,
		"github_mirror":     s.cfg.GitHubMirror,
		"core_config_debug": s.cfg.CoreConfigDebug,
		// app_settings 字段（服务器级，需重启生效）
		"force_https":      appSettings["force_https"],
		"custom_base_path": appSettings["custom_base_path"],
		"jwt_expire_hours": appSettings["jwt_expire_hours"],
	}
}

// UpdateSettingsRequest 配置更新请求（局部更新，所有字段均为指针类型）。
//
// 指针类型设计意图：
//   - nil   → 前端未传此字段，跳过处理
//   - 非nil → 前端显式设置了值（包括零值），需要处理
//
// 前端每次只发送一个字段（失焦保存），因此大部分字段为 nil。
//
// 存储归属：
//   - config.json 字段：listen_ip, socks_port, http_port, outbound_ip, github_mirror, core_config_debug
//   - app_settings 数据库表：force_https, custom_base_path（服务器级配置，需重启生效）
type UpdateSettingsRequest struct {
	// config.json 字段
	ListenIP        *string `json:"listen_ip"`
	SocksPort       *int    `json:"socks_port"`
	HTTPPort        *int    `json:"http_port"`
	OutboundIP      *string `json:"outbound_ip"`
	GitHubMirror    *string `json:"github_mirror"`
	CoreConfigDebug *bool   `json:"core_config_debug"`
	// app_settings 数据库表字段（服务器级，需重启生效）
	ForceHTTPS     *string `json:"force_https"`
	CustomBasePath *string `json:"custom_base_path"`
	JwtExpireHours *string `json:"jwt_expire_hours"`
}

// UpdateSettings 更新配置（局部更新 + 脏标记 + 白名单校验）。
//
// 三步拦截策略（每个字段）：
//  1. 判空：req.Field != nil → 前端传了此字段
//  2. 判变：*req.Field != s.cfg.Field → 值确实有变化
//  3. 判合法：端口 1-65535，IP 通过 net.ParseIP 校验
//
// dirty flag 模式：
//   - 任意字段通过三步拦截后，标记 changed = true
//   - 函数末尾只有 changed == true 时才调用 SaveJSONConfig()
//   - 如果所有字段都没变，直接返回 nil，零磁盘 I/O
func (s *SettingsService) UpdateSettings(req *UpdateSettingsRequest) error {
	changed := false

	// ListenIP：判空 → 判变 → net.ParseIP 校验
	if req.ListenIP != nil && *req.ListenIP != s.cfg.ListenIP {
		if net.ParseIP(*req.ListenIP) == nil {
			return NewValidation("invalid listen_ip: must be a valid IP address", nil)
		}
		s.cfg.ListenIP = *req.ListenIP
		changed = true
	}

	// SocksPort：判空 → 判变 → 1-65535 校验
	if req.SocksPort != nil && *req.SocksPort != s.cfg.SocksPort {
		if *req.SocksPort < 1 || *req.SocksPort > 65535 {
			return NewValidation("socks_port must be between 1 and 65535", nil)
		}
		s.cfg.SocksPort = *req.SocksPort
		changed = true
	}

	// HTTPPort：判空 → 判变 → 1-65535 校验
	if req.HTTPPort != nil && *req.HTTPPort != s.cfg.HTTPPort {
		if *req.HTTPPort < 1 || *req.HTTPPort > 65535 {
			return NewValidation("http_port must be between 1 and 65535", nil)
		}
		s.cfg.HTTPPort = *req.HTTPPort
		changed = true
	}

	// OutboundIP：判空 → 判变 → net.ParseIP 校验
	if req.OutboundIP != nil && *req.OutboundIP != s.cfg.OutboundIP {
		if net.ParseIP(*req.OutboundIP) == nil {
			return NewValidation("invalid outbound_ip: must be a valid IP address", nil)
		}
		s.cfg.OutboundIP = *req.OutboundIP
		changed = true
	}

	// GitHubMirror：判空 → 判变（允许空字符串清空，无需格式校验）
	if req.GitHubMirror != nil && *req.GitHubMirror != s.cfg.GitHubMirror {
		s.cfg.GitHubMirror = *req.GitHubMirror
		changed = true
	}

	// CoreConfigDebug：判空 → 判变（布尔值无需格式校验）
	if req.CoreConfigDebug != nil && *req.CoreConfigDebug != s.cfg.CoreConfigDebug {
		s.cfg.CoreConfigDebug = *req.CoreConfigDebug
		changed = true
	}

	// ── app_settings 数据库表字段（服务器级配置，需重启生效）──────────
	// 使用 SQLite 原生 Upsert（ON CONFLICT DO UPDATE），一条 SQL 完成插入或更新
	// 避免先 SELECT 再决定 INSERT/UPDATE 的并发竞争问题

	// ForceHTTPS：判空 → 写入 app_settings
	if req.ForceHTTPS != nil {
		if err := upsertAppSetting("force_https", *req.ForceHTTPS); err != nil {
			return fmt.Errorf("failed to save force_https: %w", err)
		}
	}

	// JwtExpireHours：判空 → 正整数校验（1-8760）→ 写入 app_settings
	if req.JwtExpireHours != nil {
		val := strings.TrimSpace(*req.JwtExpireHours)
		if val == "" {
			val = "24" // 空值回退到默认 24 小时
		}
		// 正整数校验
		hours := 0
		if _, err := fmt.Sscanf(val, "%d", &hours); err != nil || hours < 1 || hours > 8760 {
			return NewValidation("jwt_expire_hours must be a positive integer between 1 and 8760", nil)
		}
		if err := upsertAppSetting("jwt_expire_hours", val); err != nil {
			return fmt.Errorf("failed to save jwt_expire_hours: %w", err)
		}
		// 立即更新缓存，无需重启即可对新 JWT 生效
		s.cacheMu.Lock()
		s.cacheData["jwt_expire_hours"] = val
		s.cacheMu.Unlock()
	}

	// CustomBasePath：判空 → trim → 正则校验 → 写入 app_settings
	//
	// 存储规范（纯路径名，全链路统一，无斜杠）：
	//   - 空字符串 ""  → 无前缀（默认值）
	//   - "my-path"     → 访问地址 http://host:port/my-path/...
	//   - "/"、"/my/"   → 非法输入，前端负责拦截，后端正则兜底拒绝
	if req.CustomBasePath != nil {
		val := strings.TrimSpace(*req.CustomBasePath)
		// 正则校验：仅允许字母、数字、下划线、连字符（无斜杠）
		if val != "" && !basePathPattern.MatchString(val) {
			return NewValidation("custom_base_path: only letters, digits, hyphens and underscores allowed, no slashes", nil)
		}
		if err := upsertAppSetting("custom_base_path", val); err != nil {
			return fmt.Errorf("failed to save custom_base_path: %w", err)
		}
	}

	// dirty flag：只有 config.json 字段真正改变时才落盘
	// app_settings 字段已直接写入数据库，无需触发文件保存
	if !changed {
		return nil
	}

	if err := s.cfg.SaveJSONConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// upsertAppSetting 使用 SQLite 原生 Upsert 写入 app_settings 键值对。
//
// SQL 语义：INSERT INTO app_settings (key, value) VALUES (?, ?)
//
//	ON CONFLICT(key) DO UPDATE SET value = excluded.value
//
// 一条 SQL 完成插入或更新，避免先 SELECT 再决定 INSERT/UPDATE 的并发竞争。
// GORM 通过 Clauses(clause.OnConflict{...}) 翻译为数据库原生语法。
func upsertAppSetting(key, value string) error {
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},              // 冲突检测列（唯一索引）
		DoUpdates: clause.AssignmentColumns([]string{"value"}), // 冲突时更新 value
	}).Create(&database.AppSetting{Key: key, Value: value}).Error
}

// getAllAppSettings 从 app_settings 表读取所有键值对，返回 map。
// 用于 GetSettings() 合并返回给前端。
func getAllAppSettings() map[string]string {
	var settings []database.AppSetting
	result := make(map[string]string)
	if err := database.DB.Find(&settings).Error; err != nil {
		slog.Warn("failed to read app_settings", "error", err)
		return result
	}
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result
}
