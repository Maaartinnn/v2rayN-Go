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

	// 链式代理
	DialerProxy string `gorm:"size:256" json:"dialer_proxy"` // 前置代理 tag（链式代理）

	// 原始链接
	RawLink string `gorm:"type:text" json:"raw_link"` // 原始分享链接

	// 状态信息
	TestResult   string    `gorm:"size:128" json:"test_result"`    // 测速结果 (延迟 ms)
	LastTestTime time.Time `json:"last_test_time"`                 // 最后测速时间
	SortOrder    int       `json:"sort_order"`                     // 排序顺序
	IsActive     bool      `gorm:"default:false" json:"is_active"` // 是否为当前激活节点

	// 分组信息
	GroupID   uint   `json:"group_id"`                   // 所属分组 ID
	GroupName string `gorm:"size:256" json:"group_name"` // 分组名称（冗余字段，便于查询）
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

	Name      string `gorm:"size:256" json:"name"`    // 规则名称
	Type      string `gorm:"size:64" json:"type"`     // 规则类型: direct, proxy, block
	Domain    string `gorm:"type:text" json:"domain"` // 域名规则（逗号分隔）
	IP        string `gorm:"type:text" json:"ip"`     // IP 规则（CIDR，逗号分隔）
	Port      string `gorm:"size:128" json:"port"`    // 端口规则
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	SortOrder int    `json:"sort_order"` // 排序顺序（越小优先级越高）
}

// StrategyGroup 策略组
type StrategyGroup struct {
	gorm.Model

	Name        string `gorm:"size:256;uniqueIndex" json:"name"` // 策略组名称
	Type        string `gorm:"size:64" json:"type"`              // 策略组类型: selector, urltest, fallback, loadbalance
	Description string `gorm:"size:512" json:"description"`      // 描述

	// 成员节点（通过关联表）
	ProfileIDs string `gorm:"type:text" json:"profile_ids"` // 成员节点 ID 列表（JSON 数组）

	// 测试设置
	TestURL      string `gorm:"size:512" json:"test_url"`         // 测试 URL
	TestInterval int    `gorm:"default:300" json:"test_interval"` // 测试间隔（秒）

	// 负载均衡设置
	Strategy string `gorm:"size:64" json:"strategy"` // 负载均衡策略: round-robin, least-load, random

	SortOrder int  `json:"sort_order"` // 排序顺序
	Enabled   bool `gorm:"default:true" json:"enabled"`
}

// AppSetting 应用设置（KV 存储）
type AppSetting struct {
	gorm.Model

	Key   string `gorm:"uniqueIndex;size:128" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
