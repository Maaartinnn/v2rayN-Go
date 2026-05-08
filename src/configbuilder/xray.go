package configbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"v2rayn-go/database"
)

// ========== Xray 配置结构体 ==========

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

// BuildXrayConfig 根据选中的节点和路由规则生成 Xray 配置
func BuildXrayConfig(profile *database.Profile, rules []database.RoutingRule, socksPort, httpPort int) (*XrayConfig, error) {
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
	cfg.Routing = buildXrayRouting(rules)

	return cfg, nil
}

// buildXrayOutbound 根据节点信息构建 outbound
func buildXrayOutbound(p *database.Profile) (*XrayOutbound, error) {
	outbound := &XrayOutbound{
		Tag: "proxy",
	}

	switch p.Protocol {
	case "vmess":
		outbound.Protocol = "vmess"
		outbound.Settings = &XrayOutboundSettings{
			Vnext: []XrayVnext{
				{
					Address: p.Address,
					Port:    p.Port,
					Users: []XrayUser{
						{
							ID:       p.UUID,
							AlterID:  p.AlterID,
							Security: p.Security,
						},
					},
				},
			},
		}

	case "vless":
		outbound.Protocol = "vless"
		user := XrayUser{
			ID:         p.UUID,
			Encryption: "none",
			Flow:       p.Flow,
		}
		outbound.Settings = &XrayOutboundSettings{
			Vnext: []XrayVnext{
				{
					Address: p.Address,
					Port:    p.Port,
					Users:   []XrayUser{user},
				},
			},
		}

	case "trojan":
		outbound.Protocol = "trojan"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.Address,
					Port:     p.Port,
					Password: p.UUID,
				},
			},
		}

	case "shadowsocks":
		outbound.Protocol = "shadowsocks"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.Address,
					Port:     p.Port,
					Method:   p.Security,
					Password: p.UUID,
				},
			},
		}

	case "socks":
		outbound.Protocol = "socks"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.Address,
					Port:     p.Port,
					Password: p.UUID,
				},
			},
		}

	case "http":
		outbound.Protocol = "http"
		outbound.Settings = &XrayOutboundSettings{
			Servers: []XrayServer{
				{
					Address:  p.Address,
					Port:     p.Port,
					Password: p.UUID,
				},
			},
		}

	default:
		return nil, fmt.Errorf("unsupported protocol for xray: %s", p.Protocol)
	}

	// 构建 StreamSettings
	streamSettings := buildXrayStreamSettings(p)
	outbound.StreamSettings = streamSettings

	return outbound, nil
}

// buildXrayStreamSettings 构建传输层设置
func buildXrayStreamSettings(p *database.Profile) *XrayStreamSettings {
	ss := &XrayStreamSettings{
		Network: p.Network,
	}

	// TLS 设置
	switch p.TLS {
	case "tls":
		ss.Security = "tls"
		tlsSettings := &XrayTLSSettings{
			AllowInsecure: p.AllowInsecure,
		}
		if p.SNI != "" {
			tlsSettings.ServerName = p.SNI
		}
		if p.Fingerprint != "" {
			tlsSettings.Fingerprint = p.Fingerprint
		}
		ss.TLSSettings = tlsSettings

	case "reality":
		ss.Security = "reality"
		realitySettings := &XrayRealitySettings{
			PublicKey: p.PublicKey,
			ShortID:   p.ShortID,
		}
		if p.SNI != "" {
			realitySettings.ServerNames = []string{p.SNI}
		}
		if p.Fingerprint != "" {
			realitySettings.Fingerprint = p.Fingerprint
		}
		ss.RealitySettings = realitySettings

	default:
		ss.Security = "none"
	}

	// 传输层协议设置
	switch p.Network {
	case "ws":
		ss.WSSettings = &XrayWSSettings{
			Path: p.Path,
		}
		if p.Host != "" {
			ss.WSSettings.Headers = map[string]string{
				"Host": p.Host,
			}
		}

	case "h2":
		ss.HTTPSettings = &XrayHTTPSettings{
			Path: p.Path,
		}
		if p.Host != "" {
			ss.HTTPSettings.Host = []string{p.Host}
		}

	case "grpc":
		ss.GRPCSettings = &XrayGRPCSettings{
			ServiceName: p.Path,
		}

	case "tcp":
		if p.Host != "" {
			ss.TCPSettings = &XrayTCPSettings{
				Header: &XrayTCPHeader{
					Type: "http",
					Request: &XrayHTTPRequest{
						Path: []string{p.Path},
						Headers: map[string][]string{
							"Host": {p.Host},
						},
					},
				},
			}
		}
	}

	return ss
}

// buildXrayRouting 构建路由规则
func buildXrayRouting(rules []database.RoutingRule) *XrayRouting {
	routing := &XrayRouting{
		DomainStrategy: "IPIfNonMatch",
		Rules: []XrayRule{
			// 默认直连规则：局域网和中国大陆 IP
			{
				Type:        "field",
				IP:          []string{"geoip:private", "geoip:cn"},
				OutboundTag: "direct",
			},
			// 默认直连规则：中国大陆域名
			{
				Type:        "field",
				Domain:      []string{"geosite:cn"},
				OutboundTag: "direct",
			},
		},
	}

	// 添加用户自定义规则
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

// SaveXrayConfig 生成并保存 Xray 配置文件
func SaveXrayConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, socksPort, httpPort int) (string, error) {
	cfg, err := BuildXrayConfig(profile, rules, socksPort, httpPort)
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
