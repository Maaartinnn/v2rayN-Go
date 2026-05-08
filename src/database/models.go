package database

import (
	"time"

	"gorm.io/gorm"
)

// Profile 代理节点配置
type Profile struct {
	gorm.Model

	// 基本信息
	Name     string `gorm:"size:256" json:"name"`    // 节点名称
	Address  string `gorm:"size:256" json:"address"` // 服务器地址
	Port     int    `json:"port"`                    // 服务器端口
	Protocol string `gorm:"size:64" json:"protocol"` // 协议类型: vmess, vless, trojan, shadowsocks, hysteria2 等

	// 认证信息
	UUID     string `gorm:"size:128" json:"uuid"`    // UUID / 密码
	AlterID  int    `json:"alter_id"`                // VMess alterId
	Security string `gorm:"size:64" json:"security"` // 加密方式
	Network  string `gorm:"size:64" json:"network"`  // 传输协议: tcp, ws, grpc, h2 等

	// TLS 设置
	TLS           string `gorm:"size:32" json:"tls"`         // tls, reality, ""
	SNI           string `gorm:"size:256" json:"sni"`        // Server Name Indication
	Fingerprint   string `gorm:"size:64" json:"fingerprint"` // TLS 指纹
	AllowInsecure bool   `json:"allow_insecure"`             // 跳过证书验证

	// 传输层设置
	Host      string `gorm:"size:256" json:"host"`       // WS/H2 主机头
	Path      string `gorm:"size:512" json:"path"`       // WS/H2 路径
	Seed      string `gorm:"size:256" json:"seed"`       // QUIC seed
	Flow      string `gorm:"size:64" json:"flow"`        // VLESS flow
	PublicKey string `gorm:"size:128" json:"public_key"` // Reality 公钥
	ShortID   string `gorm:"size:128" json:"short_id"`   // Reality shortId
	SiderSNI  string `gorm:"size:256" json:"sider_sni"`  // Reality serverName

	// 原始链接
	RawLink string `gorm:"type:text" json:"raw_link"` // 原始分享链接

	// 状态信息
	TestResult   string    `gorm:"size:128" json:"test_result"`    // 测速结果 (延迟 ms)
	LastTestTime time.Time `json:"last_test_time"`                 // 最后测速时间
	SortOrder    int       `json:"sort_order"`                     // 排序顺序
	IsActive     bool      `gorm:"default:false" json:"is_active"` // 是否为当前激活节点

	// 订阅与分组信息
	SubscriptionID uint   `json:"subscription_id"`            // 所属订阅 ID
	GroupID        uint   `json:"group_id"`                   // 所属分组 ID
	GroupName      string `gorm:"size:256" json:"group_name"` // 分组名称（冗余字段，便于查询）
}

// Subscription 订阅源
type Subscription struct {
	gorm.Model

	Name    string `gorm:"size:256" json:"name"` // 订阅名称
	URL     string `gorm:"type:text" json:"url"` // 订阅地址
	Enabled bool   `gorm:"default:true" json:"enabled"`

	// 自动更新设置
	AutoUpdate     bool      `gorm:"default:true" json:"auto_update"`      // 是否自动更新
	UpdateInterval int       `gorm:"default:86400" json:"update_interval"` // 更新间隔（秒）
	LastUpdateTime time.Time `json:"last_update_time"`                     // 最后更新时间

	// 请求设置
	UserAgent string `gorm:"size:512" json:"user_agent"` // 自定义 User-Agent

	// 分组
	GroupID uint `json:"group_id"` // 所属分组 ID

	// 订阅信息
	UserInfo  string `gorm:"type:text" json:"user_info"` // 用户信息（流量等）
	NodeCount int    `json:"node_count"`                 // 节点数量（计算字段）
}

// NodeGroup 节点分组
type NodeGroup struct {
	gorm.Model

	Name        string `gorm:"size:256;uniqueIndex" json:"name"` // 分组名称
	Description string `gorm:"size:512" json:"description"`      // 分组描述
	SortOrder   int    `json:"sort_order"`                       // 排序顺序
	Color       string `gorm:"size:32" json:"color"`             // 分组颜色标识
}

// RoutingRule 路由规则
type RoutingRule struct {
	gorm.Model

	Name      string `gorm:"size:256" json:"name"`    // 规则名称
	Type      string `gorm:"size:64" json:"type"`     // 规则类型: direct, proxy, block
	Domain    string `gorm:"type:text" json:"domain"` // 域名规则（逗号分隔）
	IP        string `gorm:"type:text" json:"ip"`     // IP 规则（CIDR，逗号分隔）
	Port      string `gorm:"size:128" json:"port"`    // 端口规则
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	SortOrder int    `json:"sort_order"` // 排序顺序（越小优先级越高）
}

// AppSetting 应用设置（KV 存储）
type AppSetting struct {
	gorm.Model

	Key   string `gorm:"uniqueIndex;size:128" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
