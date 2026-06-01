package database

import (
	"time"

	"gorm.io/gorm"
)

// User 系统用户（鉴权与两步验证）
type User struct {
	gorm.Model

	// 记录标识
	UUID string `gorm:"size:36;uniqueIndex;not null" json:"uuid"` // 用户唯一标识（JWT sub 字段）

	// 认证信息
	Username     string `gorm:"size:64;uniqueIndex;not null" json:"username"` // 登录用户名
	PasswordHash string `gorm:"size:256;not null" json:"-"`                   // bcrypt 密码哈希（JSON 不输出）
	JWTSecret    string `gorm:"size:256;not null" json:"-"`                   // 用户专属 JWT 签名密钥

	// 两步验证 (TOTP)
	TOTPSecret  string `gorm:"size:64;default:''" json:"-"`       // TOTP 密钥（Base32 编码）
	TOTPEnabled bool   `gorm:"default:false" json:"totp_enabled"` // 是否已启用两步验证

	// 权限
	Role int `gorm:"default:0" json:"role"` // 角色: 0=普通用户, 1=超管（全局唯一）
}

// Profile 代理节点配置
type Profile struct {
	gorm.Model

	// 记录标识
	UUID string `gorm:"size:36;uniqueIndex" json:"uuid"` // 记录唯一标识

	// 基本信息
	Name string `gorm:"size:256" json:"name"` // 节点名称

	// 代理配置（proxy_ 前缀，与业务字段隔离）
	ProxyAddress       string `gorm:"size:256" json:"proxy_address"`      // 服务器地址
	ProxyPort          int    `json:"proxy_port"`                         // 服务器端口
	ProxyProtocol      string `gorm:"size:64" json:"proxy_protocol"`      // 协议类型: vmess, vless, trojan, shadowsocks, hysteria2 等
	ProxyCredential    string `gorm:"size:128" json:"proxy_credential"`   // UUID / 密码（代理认证凭证）
	ProxyAlterID       int    `json:"proxy_alter_id"`                     // VMess alterId
	ProxySecurity      string `gorm:"size:64" json:"proxy_security"`      // 加密方式
	ProxyNetwork       string `gorm:"size:64" json:"proxy_network"`       // 传输协议: tcp, ws, grpc, h2 等
	ProxyTLS           string `gorm:"size:32" json:"proxy_tls"`           // tls, reality, ""
	ProxySNI           string `gorm:"size:256" json:"proxy_sni"`          // Server Name Indication
	ProxyFingerprint   string `gorm:"size:64" json:"proxy_fingerprint"`   // TLS 指纹
	ProxyAllowInsecure bool   `json:"proxy_allow_insecure"`               // 跳过证书验证
	ProxyHost          string `gorm:"size:256" json:"proxy_host"`         // WS/H2 主机头
	ProxyPath          string `gorm:"size:512" json:"proxy_path"`         // WS/H2 路径
	ProxySeed          string `gorm:"size:256" json:"proxy_seed"`         // QUIC seed
	ProxyFlow          string `gorm:"size:64" json:"proxy_flow"`          // VLESS flow
	ProxyPublicKey     string `gorm:"size:128" json:"proxy_public_key"`   // Reality 公钥
	ProxyShortID       string `gorm:"size:128" json:"proxy_short_id"`     // Reality shortId
	ProxySiderSNI      string `gorm:"size:256" json:"proxy_sider_sni"`    // Reality serverName
	ProxyDialerProxy   string `gorm:"size:256" json:"proxy_dialer_proxy"` // 前置代理 tag（链式代理）

	// 内核设置
	CoreType string `gorm:"size:64;default:''" json:"core_type"` // 内核类型: xray, sing-box, mihomo, ""(自动)

	// 原始链接
	RawLink string `gorm:"type:text" json:"raw_link"` // 原始分享链接

	// 状态信息
	TestResult   string    `gorm:"size:128" json:"test_result"`    // 测速结果 (延迟 ms)
	LastTestTime time.Time `json:"last_test_time"`                 // 最后测速时间
	SortOrder    int       `json:"sort_order"`                     // 排序顺序
	IsActive     bool      `gorm:"default:false" json:"is_active"` // 是否为当前激活节点

	// 分组信息
	GroupUUID string `gorm:"size:36;not null;index" json:"group_uuid"` // 所属分组 UUID

	// === 策略组字段（strategy_ 前缀，仅 proxy_protocol 为策略类型时使用）===
	StrategyMemberUUIDs  string `gorm:"type:text" json:"strategy_member_uuids"`    // 成员 Profile UUID 列表（JSON 数组）
	StrategyTestURL      string `gorm:"size:512" json:"strategy_test_url"`         // 测试 URL
	StrategyTestInterval int    `gorm:"default:300" json:"strategy_test_interval"` // 测试间隔（秒）
	StrategyType         string `gorm:"size:64" json:"strategy_type"`              // 负载均衡策略: round-robin, least-load, random
}

// NodeGroup 节点分组（统一管理：普通分组 + 订阅分组）
type NodeGroup struct {
	gorm.Model

	// 基本信息
	UUID        string `gorm:"size:36;uniqueIndex" json:"uuid"` // 唯一标识
	Alias       string `gorm:"size:256" json:"alias"`           // 别名
	Description string `gorm:"size:512" json:"description"`     // 描述（兼容）
	SortOrder   int    `json:"sort_order"`                      // 排序顺序
	Color       string `gorm:"size:32" json:"color"`            // 颜色标识

	// 订阅相关（仅订阅分组有效）
	IsSubscription bool   `gorm:"default:false" json:"is_subscription"` // 是否为订阅分组
	URL            string `gorm:"type:text" json:"url"`                 // 订阅地址
	Enabled        bool   `gorm:"default:true" json:"enabled"`          // 启用
	EnableUpdate   bool   `gorm:"default:false" json:"enable_update"`   // 启用更新
	UpdateInterval int    `gorm:"default:0" json:"update_interval"`     // 自动更新间隔（分钟），≤0禁用
	AliasRegex     string `gorm:"size:512" json:"alias_regex"`          // 别名正则过滤
	UserAgent      string `gorm:"size:512" json:"user_agent"`           // User-Agent

	// 备注与状态
	Notes          string    `gorm:"type:text" json:"notes"`     // 备注
	LastUpdateTime time.Time `json:"last_update_time"`           // 最后更新时间
	UserInfo       string    `gorm:"type:text" json:"user_info"` // 用户信息（流量等）
	NodeCount      int       `json:"node_count"`                 // 节点数量
}

// RoutingRule 路由规则
type RoutingRule struct {
	gorm.Model

	UUID      string `gorm:"size:36;uniqueIndex" json:"uuid"` // 唯一标识
	Name      string `gorm:"size:256" json:"name"`            // 规则名称
	Type      string `gorm:"size:64" json:"type"`             // 规则类型: direct, proxy, block
	Domain    string `gorm:"type:text" json:"domain"`         // 域名规则（逗号分隔）
	IP        string `gorm:"type:text" json:"ip"`             // IP 规则（CIDR，逗号分隔）
	Port      string `gorm:"size:128" json:"port"`            // 端口规则
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	SortOrder int    `json:"sort_order"` // 排序顺序（越小优先级越高）
}

// AppSetting 应用设置（KV 存储）
type AppSetting struct {
	gorm.Model

	Key   string `gorm:"uniqueIndex;size:128" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
