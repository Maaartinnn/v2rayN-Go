// mihomo.go — Mihomo (Clash Meta) 配置生成器
//
// 将 database.Profile + RoutingRule 转换为 Mihomo 原生 YAML 配置。
// Mihomo 与 Xray/Singbox 不同，使用 YAML 格式而非 JSON，
// 因此 Build() 保存为 .yaml 文件，BuildBytes() 返回 YAML 字节。
//
// 结构体设计策略（平衡可维护性与覆盖度）：
//   - 基础通用字段（name, type, server, port 等）严格定义为强类型
//   - 协议专属参数（ws-opts, grpc-opts, reality-opts 等）通过 Extra map[string]any + yaml:",inline" 内联
//   - 指针类型字段（*bool, *int）配合 omitempty 标签，避免零值误判为"未设置"
//   - 利用 Go 1.26 new(expr) 语法创建指针字面量，如 new(true)

package configbuilder

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"v2rayn-go/database"
)

// ========== Mihomo 配置结构体 ==========

// MihomoConfig Mihomo 完整配置（根级别）
//
// 对应 Mihomo YAML 配置文件的顶层结构。
// 参考: https://wiki.metacubex.one/config/
type MihomoConfig struct {
	MixedPort   int    `yaml:"mixed-port"`             // 混合代理端口（HTTP + SOCKS5 同端口）
	AllowLan    bool   `yaml:"allow-lan,omitempty"`    // 允许局域网连接
	BindAddress string `yaml:"bind-address,omitempty"` // 绑定地址
	Mode        string `yaml:"mode"`                   // 运行模式: rule / global / direct
	LogLevel    string `yaml:"log-level"`              // 日志级别: silent / error / warning / info / debug
	IPv6        bool   `yaml:"ipv6,omitempty"`         // 是否启用 IPv6

	Proxies     []MihomoProxy      `yaml:"proxies,omitempty"`      // 代理节点列表
	ProxyGroups []MihomoProxyGroup `yaml:"proxy-groups,omitempty"` // 策略组列表
	Rules       []string           `yaml:"rules,omitempty"`        // 路由规则列表
}

// MihomoProxy 代理节点（基础字段严格定义，协议专属参数通过 Extra 内联）
//
// 基础字段覆盖所有 8 种协议的公共属性。
// 协议专属参数（ws-opts, grpc-opts, reality-opts, hysteria2-opts 等）
// 通过 Extra map + yaml:",inline" 标签序列化为 YAML 的内联键值对，
// 避免维护庞大而易变的完整结构体。
//
// 指针类型字段（*bool, *int）配合 omitempty 标签使用：
//   - nil → 字段不出现在 YAML 中（语义：未设置，由 Mihomo 使用默认值）
//   - 非 nil → 字段出现在 YAML 中（语义：用户显式指定了值）
//
// 创建指针字面量请使用 Go 1.26 的 new(expr) 语法：
//
//	proxy.TLS = new(true)          // *bool = true
//	proxy.SkipCertVerify = new(false)
//	proxy.AlterID = new(int(0))    // *int = 0
type MihomoProxy struct {
	// === 通用基础字段（所有协议共享）===
	Name     string `yaml:"name"`               // 节点名称（必须唯一）
	Type     string `yaml:"type"`               // 协议类型: vmess, vless, trojan, ss, hysteria2, tuic, socks5, http
	Server   string `yaml:"server"`             // 服务器地址
	Port     int    `yaml:"port"`               // 服务器端口
	UUID     string `yaml:"uuid,omitempty"`     // UUID（vmess/vless/tuic）
	Password string `yaml:"password,omitempty"` // 密码（trojan/ss/hysteria2/http/socks5）
	Cipher   string `yaml:"cipher,omitempty"`   // 加密方式（ss 必填，vmess 可选）

	// === 通用可选字段 ===
	UDP            *bool  `yaml:"udp,omitempty"`                // 启用 UDP 转发
	TLS            *bool  `yaml:"tls,omitempty"`                // 启用 TLS
	ServerName     string `yaml:"servername,omitempty"`         // TLS SNI
	SkipCertVerify *bool  `yaml:"skip-cert-verify,omitempty"`   // 跳过证书验证
	Fingerprint    string `yaml:"client-fingerprint,omitempty"` // TLS 指纹（chrome/firefox 等）
	Network        string `yaml:"network,omitempty"`            // 传输协议: tcp/ws/grpc/h2
	Flow           string `yaml:"flow,omitempty"`               // VLESS 流控（xtls-rprx-vision）

	// === 协议专属参数 ===
	// 通过 yaml:",inline" 将 map 内容展开为 YAML 同级键值对。
	// 例如 Extra["ws-opts"] = map[string]any{"path": "/api"}
	// 序列化为：
	//   ws-opts:
	//     path: /api
	Extra map[string]any `yaml:",inline,omitempty"`
}

// MihomoProxyGroup 策略组
//
// 支持 4 种策略组类型：
//   - select: 手动选择
//   - url-test: 自动测速
//   - fallback: 故障转移
//   - load-balance: 负载均衡
type MihomoProxyGroup struct {
	Name      string   `yaml:"name"`                // 策略组名称
	Type      string   `yaml:"type"`                // select / url-test / fallback / load-balance
	Proxies   []string `yaml:"proxies"`             // 成员节点名称列表
	URL       string   `yaml:"url,omitempty"`       // 测速 URL（url-test/fallback 专用）
	Interval  int      `yaml:"interval,omitempty"`  // 测速间隔（秒）
	Tolerance int      `yaml:"tolerance,omitempty"` // 容差（ms，url-test 专用）
}

// ========== Mihomo 配置构建函数 ==========

// BuildMihomoConfig 根据选中的节点和路由规则生成 Mihomo 配置
//
// 参数：
//   - profile: 当前激活的代理节点
//   - rules:   用户自定义路由规则列表（已按 sort_order 排序）
//   - configDir: 应用目录（用于查找 geo 数据文件）
//   - mixedPort: 混合代理端口（HTTP + SOCKS5 同端口）
//
// 返回的 MihomoConfig 可直接通过 yaml.Marshal 序列化为 Mihomo 原生 YAML。
func BuildMihomoConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, mixedPort int) (*MihomoConfig, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	cfg := &MihomoConfig{
		MixedPort: mixedPort,
		Mode:      "rule",
		LogLevel:  "info",
	}

	// 构建代理节点
	proxy, err := buildMihomoProxy(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to build proxy: %w", err)
	}
	cfg.Proxies = []MihomoProxy{*proxy}

	// 构建默认策略组（手动选择，包含唯一的代理节点 + DIRECT）
	cfg.ProxyGroups = []MihomoProxyGroup{
		{
			Name:    "Proxy",
			Type:    "select",
			Proxies: []string{proxy.Name, "DIRECT"},
		},
	}

	// 构建路由规则
	cfg.Rules = buildMihomoRules(rules, configDir)

	return cfg, nil
}

// buildMihomoProxy 根据节点信息构建 Mihomo 代理节点
//
// 协议映射（Profile.ProxyProtocol → Mihomo type）：
//
//	vmess          → vmess
//	vless          → vless
//	trojan         → trojan
//	shadowsocks    → ss
//	hysteria2      → hysteria2
//	tuic           → tuic
//	socks          → socks5
//	http           → http
func buildMihomoProxy(p *database.Profile) (*MihomoProxy, error) {
	proxy := &MihomoProxy{
		Name: "proxy",
	}

	switch p.ProxyProtocol {
	case "vmess":
		proxy.Type = "vmess"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.UUID = p.ProxyCredential
		proxy.Cipher = p.ProxySecurity
		if proxy.Cipher == "" {
			proxy.Cipher = "auto"
		}
		proxy.UDP = new(true)

	case "vless":
		proxy.Type = "vless"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.UUID = p.ProxyCredential
		proxy.Flow = p.ProxyFlow
		proxy.UDP = new(true)

	case "trojan":
		proxy.Type = "trojan"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.Password = p.ProxyCredential
		proxy.UDP = new(true)

	case "shadowsocks":
		proxy.Type = "ss"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.Cipher = p.ProxySecurity
		proxy.Password = p.ProxyCredential
		proxy.UDP = new(true)

	case "hysteria2":
		proxy.Type = "hysteria2"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.Password = p.ProxyCredential
		proxy.UDP = new(true)

	case "tuic":
		proxy.Type = "tuic"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.UUID = p.ProxyCredential
		proxy.Password = p.ProxySecurity
		proxy.UDP = new(true)

	case "socks":
		proxy.Type = "socks5"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.UUID = p.ProxyCredential   // Username
		proxy.Password = p.ProxySecurity // Password
		proxy.UDP = new(true)

	case "http":
		proxy.Type = "http"
		proxy.Server = p.ProxyAddress
		proxy.Port = p.ProxyPort
		proxy.UUID = p.ProxyCredential   // Username
		proxy.Password = p.ProxySecurity // Password

	default:
		return nil, fmt.Errorf("unsupported protocol for mihomo: %s", p.ProxyProtocol)
	}

	// TLS 设置
	buildMihomoTLS(proxy, p)

	// 传输层设置
	buildMihomoTransport(proxy, p)

	return proxy, nil
}

// buildMihomoTLS 构建 Mihomo 代理节点的 TLS 相关字段
//
// 根据 Profile.ProxyTLS 值设置 TLS/Reality 参数：
//   - "tls"    → 启用 TLS + SNI + 指纹 + 跳过证书验证
//   - "reality" → 启用 TLS + reality-opts
//   - 其他协议（trojan/hysteria2/tuic）隐式启用 TLS
func buildMihomoTLS(proxy *MihomoProxy, p *database.Profile) {
	switch p.ProxyTLS {
	case "tls":
		proxy.TLS = new(true)
		if p.ProxySNI != "" {
			proxy.ServerName = p.ProxySNI
		}
		if p.ProxyFingerprint != "" {
			proxy.Fingerprint = p.ProxyFingerprint
		}
		if p.ProxyAllowInsecure {
			proxy.SkipCertVerify = new(true)
		}

	case "reality":
		proxy.TLS = new(true)
		if p.ProxySNI != "" {
			proxy.ServerName = p.ProxySNI
		}
		if p.ProxyFingerprint != "" {
			proxy.Fingerprint = p.ProxyFingerprint
		}
		// Reality 参数通过 Extra 内联
		realityOpts := map[string]any{
			"public-key": p.ProxyPublicKey,
			"short-id":   p.ProxyShortID,
		}
		if proxy.Extra == nil {
			proxy.Extra = make(map[string]any)
		}
		proxy.Extra["reality-opts"] = realityOpts

	default:
		// Trojan / Hysteria2 / TUIC 协议隐式要求 TLS
		if p.ProxyProtocol == "trojan" || p.ProxyProtocol == "hysteria2" || p.ProxyProtocol == "tuic" {
			proxy.TLS = new(true)
			if p.ProxySNI != "" {
				proxy.ServerName = p.ProxySNI
			}
			if p.ProxyAllowInsecure {
				proxy.SkipCertVerify = new(true)
			}
		}
	}
}

// buildMihomoTransport 构建 Mihomo 代理节点的传输层参数
//
// 根据 Profile.ProxyNetwork 设置对应的传输层选项（ws-opts, grpc-opts 等），
// 通过 Extra map 内联到 YAML 中。
//
// 传输协议映射：
//   - ws   → network: ws + ws-opts
//   - h2   → network: h2 + h2-opts（Mihomo 使用 h2-opts 而非 http-opts）
//   - grpc → network: grpc + grpc-opts
//   - tcp  → 若有 Host 则设置 http-opts（伪装头部）
func buildMihomoTransport(proxy *MihomoProxy, p *database.Profile) {
	switch p.ProxyNetwork {
	case "ws":
		proxy.Network = "ws"
		wsOpts := map[string]any{}
		if p.ProxyPath != "" {
			wsOpts["path"] = p.ProxyPath
		}
		if p.ProxyHost != "" {
			wsOpts["headers"] = map[string]string{
				"Host": p.ProxyHost,
			}
		}
		if len(wsOpts) > 0 {
			if proxy.Extra == nil {
				proxy.Extra = make(map[string]any)
			}
			proxy.Extra["ws-opts"] = wsOpts
		}

	case "h2":
		proxy.Network = "h2"
		h2Opts := map[string]any{}
		if p.ProxyPath != "" {
			h2Opts["path"] = p.ProxyPath
		}
		if p.ProxyHost != "" {
			h2Opts["host"] = []string{p.ProxyHost}
		}
		if len(h2Opts) > 0 {
			if proxy.Extra == nil {
				proxy.Extra = make(map[string]any)
			}
			proxy.Extra["h2-opts"] = h2Opts
		}

	case "grpc":
		proxy.Network = "grpc"
		if p.ProxyPath != "" {
			grpcOpts := map[string]any{
				"grpc-service-name": p.ProxyPath,
			}
			if proxy.Extra == nil {
				proxy.Extra = make(map[string]any)
			}
			proxy.Extra["grpc-opts"] = grpcOpts
		}

	case "tcp":
		// TCP 传输层伪装为 HTTP
		if p.ProxyHost != "" {
			httpOpts := map[string]any{
				"path": []string{p.ProxyPath},
				"headers": map[string][]string{
					"Host": {p.ProxyHost},
				},
			}
			if proxy.Extra == nil {
				proxy.Extra = make(map[string]any)
			}
			proxy.Extra["http-opts"] = httpOpts
		}
	}
}

// buildMihomoRules 构建 Mihomo 路由规则
//
// Mihomo 路由规则格式为纯字符串: "TYPE,PARAM,POLICY"
// 规则按顺序从上到下匹配，MATCH 兜底。
//
// 规则来源（按优先级）：
//  1. 直连规则：局域网 IP（始终优先）
//  2. Geo 规则：如果 geo 数据文件存在，添加国内域名/IP 直连规则
//  3. 用户自定义规则：来自 RoutingRule 表
//  4. 兜底规则：MATCH → Proxy
func buildMihomoRules(rules []database.RoutingRule, configDir string) []string {
	var mihomoRules []string

	// 局域网 IP 直连（始终优先）
	mihomoRules = append(mihomoRules, "IP-CIDR,10.0.0.0/8,DIRECT,no-resolve")
	mihomoRules = append(mihomoRules, "IP-CIDR,172.16.0.0/12,DIRECT,no-resolve")
	mihomoRules = append(mihomoRules, "IP-CIDR,192.168.0.0/16,DIRECT,no-resolve")

	// Geo 数据文件规则（检查 Mihomo 内核目录下是否存在 geo 数据文件）
	hasGeoIP, hasGeoSite := hasGeoDatFiles(configDir)
	if hasGeoSite {
		mihomoRules = append(mihomoRules, "GEOSITE,cn,DIRECT")
	}
	if hasGeoIP {
		mihomoRules = append(mihomoRules, "GEOIP,CN,DIRECT")
	}

	// 用户自定义规则
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		policy := rule.Type // direct / proxy / block
		switch policy {
		case "proxy":
			policy = "Proxy" // 对齐策略组名称
		case "direct":
			policy = "DIRECT"
		case "block":
			policy = "REJECT"
		}

		// 域名规则
		if rule.Domain != "" {
			domains := splitAndTrim(rule.Domain)
			for _, domain := range domains {
				mihomoRules = append(mihomoRules, fmt.Sprintf("DOMAIN-SUFFIX,%s,%s", domain, policy))
			}
		}

		// IP 规则
		if rule.IP != "" {
			ips := splitAndTrim(rule.IP)
			for _, ip := range ips {
				mihomoRules = append(mihomoRules, fmt.Sprintf("IP-CIDR,%s,%s,no-resolve", ip, policy))
			}
		}

		// 端口规则
		if rule.Port != "" {
			mihomoRules = append(mihomoRules, fmt.Sprintf("DST-PORT,%s,%s", rule.Port, policy))
		}
	}

	// 兜底规则：走代理
	mihomoRules = append(mihomoRules, "MATCH,Proxy")

	return mihomoRules
}

// SaveMihomoConfig 生成并保存 Mihomo 配置文件
//
// 输出路径: {configDir}/binConfig/mihomo_config.yaml
// 使用 YAML 格式（Mihomo 原生格式），而非 JSON。
// 与 SaveXrayConfig / SaveSingboxConfig 统一输出到 binConfig 目录。
func SaveMihomoConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, mixedPort int) (string, error) {
	cfg, err := BuildMihomoConfig(profile, rules, configDir, mixedPort)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// 确保 binConfig 目录存在
	binConfigDir := filepath.Join(configDir, "binConfig")
	if err := os.MkdirAll(binConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create binConfig directory: %w", err)
	}

	configPath := filepath.Join(binConfigDir, "mihomo_config.yaml")
	// mihomo_config.yaml 是运行时从数据库动态生成的派生数据，不需要原子写入
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return configPath, nil
}
