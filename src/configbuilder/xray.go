package configbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"v2rayn-go/database"
)

// ========== Xray 配置结构�?==========

// XrayConfig Xray 完整配置
type XrayConfig struct {
	Log       XrayLog        `json:"log"`
	API       *XrayAPI       `json:"api,omitempty"`
	DNS       *XrayDNS       `json:"dns,omitempty"`
	Inbounds  []XrayInbound  `json:"inbounds"`
	Outbounds []XrayOutbound `json:"outbounds"`
	Routing   *XrayRouting   `json:"routing,omitempty"`
	Policy    *XrayPolicy    `json:"policy,omitempty"`
}

type XrayLog struct {
	Access   string `json:"access,omitempty"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}

type XrayAPI struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

type XrayDNS struct {
	Servers []interface{} `json:"servers"`
	Tag     string        `json:"tag,omitempty"`
}

type XrayInbound struct {
	Listen         string               `json:"listen,omitempty"`
	Port           int                  `json:"port"`
	Protocol       string               `json:"protocol"`
	Settings       *XrayInboundSettings `json:"settings,omitempty"`
	StreamSettings *XrayStreamSettings  `json:"streamSettings,omitempty"`
	Tag            string               `json:"tag"`
	Sniffing       *XraySniffing        `json:"sniffing,omitempty"`
}

type XrayInboundSettings struct {
	Auth             string `json:"auth,omitempty"`
	UDP              bool   `json:"udp,omitempty"`
	AllowTransparent bool   `json:"allowTransparent,omitempty"`
}

type XraySniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

type XrayOutbound struct {
	Protocol       string                `json:"protocol"`
	Settings       *XrayOutboundSettings `json:"settings,omitempty"`
	StreamSettings *XrayStreamSettings   `json:"streamSettings,omitempty"`
	Tag            string                `json:"tag"`
	Mux            *XrayMux              `json:"mux,omitempty"`
}

type XrayOutboundSettings struct {
	Vnext    []XrayVnext   `json:"vnext,omitempty"`
	Servers  []XrayServer  `json:"servers,omitempty"`
	Response *XrayResponse `json:"response,omitempty"`
}

type XrayVnext struct {
	Address string     `json:"address"`
	Port    int        `json:"port"`
	Users   []XrayUser `json:"users"`
}

type XrayUser struct {
	ID         string `json:"id"`
	AlterID    int    `json:"alterId,omitempty"`
	Security   string `json:"security,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Flow       string `json:"flow,omitempty"`
}

type XrayServer struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Method   string `json:"method,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	OTA      bool   `json:"ota,omitempty"`
}

type XrayResponse struct {
	Type string `json:"type"`
}

type XrayStreamSettings struct {
	Network         string               `json:"network"`
	Security        string               `json:"security"`
	TLSSettings     *XrayTLSSettings     `json:"tlsSettings,omitempty"`
	RealitySettings *XrayRealitySettings `json:"realitySettings,omitempty"`
	WSSettings      *XrayWSSettings      `json:"wsSettings,omitempty"`
	HTTPSettings    *XrayHTTPSettings    `json:"httpSettings,omitempty"`
	GRPCSettings    *XrayGRPCSettings    `json:"grpcSettings,omitempty"`
	TCPSettings     *XrayTCPSettings     `json:"tcpSettings,omitempty"`
	SocketSettings  *XraySocketSettings  `json:"socketSettings,omitempty"`
}

type XrayTLSSettings struct {
	ServerName    string   `json:"serverName,omitempty"`
	AllowInsecure bool     `json:"allowInsecure,omitempty"`
	ALPN          []string `json:"alpn,omitempty"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
}

type XrayRealitySettings struct {
	Show        bool     `json:"show,omitempty"`
	Xver        int      `json:"xver,omitempty"`
	Server      string   `json:"server,omitempty"`
	ServerNames []string `json:"serverNames,omitempty"`
	PublicKey   string   `json:"publicKey,omitempty"`
	ShortID     string   `json:"shortID,omitempty"`
	SpiderX     string   `json:"spiderX,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

type XrayWSSettings struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type XrayHTTPSettings struct {
	Path string   `json:"path,omitempty"`
	Host []string `json:"host,omitempty"`
}

type XrayGRPCSettings struct {
	ServiceName string `json:"serviceName,omitempty"`
}

type XrayTCPSettings struct {
	Header *XrayTCPHeader `json:"header,omitempty"`
}

type XrayTCPHeader struct {
	Type    string           `json:"type"`
	Request *XrayHTTPRequest `json:"request,omitempty"`
}

type XrayHTTPRequest struct {
	Version string              `json:"version,omitempty"`
	Method  string              `json:"method,omitempty"`
	Path    []string            `json:"path,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
}

type XraySocketSettings struct {
	Mark        int    `json:"mark,omitempty"`
	TCPFastOpen bool   `json:"tcpFastOpen,omitempty"`
	Tproxy      string `json:"tproxy,omitempty"`
}

type XrayMux struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency,omitempty"`
}

type XrayRouting struct {
	DomainStrategy string        `json:"domainStrategy"`
	DomainMatcher  string        `json:"domainMatcher,omitempty"`
	Rules          []XrayRule    `json:"rules"`
	Balancers      []interface{} `json:"balancers,omitempty"`
}

type XrayRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
	OutboundTag string   `json:"outboundTag"`
}

type XrayPolicy struct {
	Levels map[string]XrayPolicyLevel `json:"levels,omitempty"`
}

type XrayPolicyLevel struct {
	Handshake    int `json:"handshake,omitempty"`
	ConnIdle     int `json:"connIdle,omitempty"`
	UplinkOnly   int `json:"uplinkOnly,omitempty"`
	DownlinkOnly int `json:"downlinkOnly,omitempty"`
}

// BuildXrayConfigWithStrategy 根据选中的节点、路由规则和策略组生�?Xray 配置
func BuildXrayConfigWithStrategy(profile *database.Profile, rules []database.RoutingRule, strategyGroups []database.StrategyGroup, configDir string, socksPort, httpPort int) (*XrayConfig, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	cfg := &XrayConfig{
		Log: XrayLog{
			Error:    "warning",
			Loglevel: "warning",
		},
		Inbounds: []XrayInbound{
			{
				Listen:   "127.0.0.1",
				Port:     socksPort,
				Protocol: "socks",
				Settings: &XrayInboundSettings{},
				Tag:      "socks-in",
				Sniffing: &XraySniffing{
					Enabled:      true,
					DestOverride: []string{"http", "tls", "quic"},
				},
			},
			{
				Listen:   "127.0.0.1",
				Port:     httpPort,
				Protocol: "http",
				Settings: &XrayInboundSettings{},
				Tag:      "http-in",
			},
		},
	}

	// 构建�?outbound
	outbound, err := buildXrayOutbound(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to build outbound: %w", err)
	}
	cfg.Outbounds = []XrayOutbound{*outbound}

	// 构建策略�?outbounds �?balancers
	if len(strategyGroups) > 0 {
		balancers := buildXrayBalancers(strategyGroups)
		cfg.Routing = buildXrayRoutingWithBalancers(rules, balancers, configDir)
	} else {
		cfg.Routing = buildXrayRouting(rules, configDir)
	}

	return cfg, nil
}

// BuildXrayConfig 根据选中的节点和路由规则生成 Xray 配置
func BuildXrayConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, socksPort, httpPort int) (*XrayConfig, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	cfg := &XrayConfig{
		Log: XrayLog{
			Error:    "warning",
			Loglevel: "warning",
		},
		Inbounds: []XrayInbound{
			{
				Listen:   "127.0.0.1",
				Port:     socksPort,
				Protocol: "socks",
				Settings: &XrayInboundSettings{},
				Tag:      "socks-in",
				Sniffing: &XraySniffing{
					Enabled:      true,
					DestOverride: []string{"http", "tls", "quic"},
				},
			},
			{
				Listen:   "127.0.0.1",
				Port:     httpPort,
				Protocol: "http",
				Settings: &XrayInboundSettings{},
				Tag:      "http-in",
			},
		},
	}

	// 构建 outbound
	outbound, err := buildXrayOutbound(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to build outbound: %w", err)
	}
	cfg.Outbounds = []XrayOutbound{*outbound}

	// 构建路由规则
	cfg.Routing = buildXrayRouting(rules, configDir)

	return cfg, nil
}

// buildXrayOutbound 根据节点信息构建 outbound
func buildXrayOutbound(p *database.Profile) (*XrayOutbound, error) {
	outbound := &XrayOutbound{
		Tag: "proxy",
	}

	switch p.ProxyProtocol {
	case "vmess":
		outbound.Protocol = "vmess"
		outbound.Settings = &XrayOutboundSettings{
			Vnext: []XrayVnext{
				{
					Address: p.ProxyAddress,
					Port:    p.ProxyPort,
					Users: []XrayUser{
						{
							ID:       p.ProxyCredential,
							AlterID:  p.ProxyAlterID,
							Security: p.ProxySecurity,
						},
					},
				},
			},
		}

	case "vless":
		outbound.Protocol = "vless"
		user := XrayUser{
			ID:         p.ProxyCredential,
			Encryption: "none",
			Flow:       p.ProxyFlow,
		}
		outbound.Settings = &XrayOutboundSettings{
			Vnext: []XrayVnext{
				{
					Address: p.ProxyAddress,
					Port:    p.ProxyPort,
					Users:   []XrayUser{user},
				},
			},
		}

	case "trojan":
		outbound.Protocol = "trojan"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.ProxyAddress,
					Port:     p.ProxyPort,
					Password: p.ProxyCredential,
				},
			},
		}

	case "shadowsocks":
		outbound.Protocol = "shadowsocks"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.ProxyAddress,
					Port:     p.ProxyPort,
					Method:   p.ProxySecurity,
					Password: p.ProxyCredential,
				},
			},
		}

	case "socks":
		outbound.Protocol = "socks"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.ProxyAddress,
					Port:     p.ProxyPort,
					Password: p.ProxyCredential,
				},
			},
		}

	case "http":
		outbound.Protocol = "http"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.ProxyAddress,
					Port:     p.ProxyPort,
					Password: p.ProxyCredential,
				},
			},
		}

	default:
		return nil, fmt.Errorf("unsupported protocol for xray: %s", p.ProxyProtocol)
	}

	// 构建 StreamSettings
	streamSettings := buildXrayStreamSettings(p)
	outbound.StreamSettings = streamSettings

	return outbound, nil
}

// buildXrayStreamSettings 构建传输层设�?
func buildXrayStreamSettings(p *database.Profile) *XrayStreamSettings {
	ss := &XrayStreamSettings{
		Network: p.ProxyNetwork,
	}

	// TLS 设置
	switch p.ProxyTLS {
	case "tls":
		ss.Security = "tls"
		tlsSettings := &XrayTLSSettings{
			AllowInsecure: p.ProxyAllowInsecure,
		}
		if p.ProxySNI != "" {
			tlsSettings.ServerName = p.ProxySNI
		}
		if p.ProxyFingerprint != "" {
			tlsSettings.Fingerprint = p.ProxyFingerprint
		}
		ss.TLSSettings = tlsSettings

	case "reality":
		ss.Security = "reality"
		realitySettings := &XrayRealitySettings{
			PublicKey: p.ProxyPublicKey,
			ShortID:   p.ProxyShortID,
		}
		if p.ProxySNI != "" {
			realitySettings.ServerNames = []string{p.ProxySNI}
		}
		if p.ProxyFingerprint != "" {
			realitySettings.Fingerprint = p.ProxyFingerprint
		}
		ss.RealitySettings = realitySettings

	default:
		ss.Security = "none"
	}

	// 传输层协议设�?
	switch p.ProxyNetwork {
	case "ws":
		ss.WSSettings = &XrayWSSettings{
			Path: p.ProxyPath,
		}
		if p.ProxyHost != "" {
			ss.WSSettings.Headers = map[string]string{
				"Host": p.ProxyHost,
			}
		}

	case "h2":
		ss.HTTPSettings = &XrayHTTPSettings{
			Path: p.ProxyPath,
		}
		if p.ProxyHost != "" {
			ss.HTTPSettings.Host = []string{p.ProxyHost}
		}

	case "grpc":
		ss.GRPCSettings = &XrayGRPCSettings{
			ServiceName: p.ProxyPath,
		}

	case "tcp":
		if p.ProxyHost != "" {
			ss.TCPSettings = &XrayTCPSettings{
				Header: &XrayTCPHeader{
					Type: "http",
					Request: &XrayHTTPRequest{
						Path: []string{p.ProxyPath},
						Headers: map[string][]string{
							"Host": {p.ProxyHost},
						},
					},
				},
			}
		}
	}

	return ss
}

// hasGeoDatFiles 检�?xray 目录下是否存�?geoip.dat �?geosite.dat
func hasGeoDatFiles(configDir string) (bool, bool) {
	binDir := filepath.Join(configDir, "bin", "xray")
	_, geoipErr := os.Stat(filepath.Join(binDir, "geoip.dat"))
	_, geositeErr := os.Stat(filepath.Join(binDir, "geosite.dat"))
	return geoipErr == nil, geositeErr == nil
}

// buildDefaultRoutingRules 构建默认路由规则（根�?dat 文件是否存在决定是否包含 geo 规则�?
func buildDefaultRoutingRules(configDir string) []XrayRule {
	var rules []XrayRule
	hasGeoIP, hasGeoSite := hasGeoDatFiles(configDir)

	if hasGeoIP {
		rules = append(rules, XrayRule{
			Type:        "field",
			IP:          []string{"geoip:private", "geoip:cn"},
			OutboundTag: "direct",
		})
	}
	if hasGeoSite {
		rules = append(rules, XrayRule{
			Type:        "field",
			Domain:      []string{"geosite:cn"},
			OutboundTag: "direct",
		})
	}

	return rules
}

// buildXrayRouting 构建路由规则
func buildXrayRouting(rules []database.RoutingRule, configDir string) *XrayRouting {
	routing := &XrayRouting{
		DomainStrategy: "IPIfNonMatch",
		Rules:          buildDefaultRoutingRules(configDir),
	}

	// 添加用户自定义规�?
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		xrayRule := XrayRule{
			Type:        "field",
			OutboundTag: rule.Type,
		}

		if rule.Domain != "" {
			xrayRule.Domain = splitAndTrim(rule.Domain)
		}
		if rule.IP != "" {
			xrayRule.IP = splitAndTrim(rule.IP)
		}
		if rule.Port != "" {
			xrayRule.Port = rule.Port
		}

		routing.Rules = append(routing.Rules, xrayRule)
	}

	return routing
}

// buildXrayBalancers 根据策略组构�?Xray balancers
func buildXrayBalancers(groups []database.StrategyGroup) []interface{} {
	var balancers []interface{}
	for _, g := range groups {
		if !g.Enabled {
			continue
		}
		balancer := map[string]interface{}{
			"tag":      g.Name,
			"selector": []string{g.Name},
		}
		switch g.Type {
		case "urltest":
			balancer["strategy"] = map[string]interface{}{
				"type": "urlTest",
			}
		case "fallback":
			balancer["strategy"] = map[string]interface{}{
				"type": "fallback",
			}
		case "loadbalance":
			strategyType := "random"
			switch g.Strategy {
			case "round-robin":
				strategyType = "roundRobin"
			case "least-load":
				strategyType = "leastLoad"
			}
			balancer["strategy"] = map[string]interface{}{
				"type": strategyType,
			}
		default: // selector
			balancer["strategy"] = map[string]interface{}{
				"type": "random",
			}
		}
		balancers = append(balancers, balancer)
	}
	return balancers
}

// buildXrayRoutingWithBalancers 构建�?balancer 的路由规�?
func buildXrayRoutingWithBalancers(rules []database.RoutingRule, balancers []interface{}, configDir string) *XrayRouting {
	routing := &XrayRouting{
		DomainStrategy: "IPIfNonMatch",
		Balancers:      balancers,
		Rules:          buildDefaultRoutingRules(configDir),
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		xrayRule := XrayRule{
			Type:        "field",
			OutboundTag: rule.Type,
		}
		if rule.Domain != "" {
			xrayRule.Domain = splitAndTrim(rule.Domain)
		}
		if rule.IP != "" {
			xrayRule.IP = splitAndTrim(rule.IP)
		}
		if rule.Port != "" {
			xrayRule.Port = rule.Port
		}
		routing.Rules = append(routing.Rules, xrayRule)
	}

	return routing
}

// SaveXrayConfig 生成并保�?Xray 配置文件
func SaveXrayConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, socksPort, httpPort int) (string, error) {
	cfg, err := BuildXrayConfig(profile, rules, configDir, socksPort, httpPort)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, "xray_config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return configPath, nil
}
