package configbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"v2rayn-go/database"
)

// ========== Sing-box 配置结构�?==========

// SingboxConfig Sing-box 完整配置
type SingboxConfig struct {
	Log          *SingboxLog          `json:"log,omitempty"`
	DNS          *SingboxDNS          `json:"dns,omitempty"`
	Inbounds     []SingboxInbound     `json:"inbounds"`
	Outbounds    []SingboxOutbound    `json:"outbounds"`
	Route        *SingboxRoute        `json:"route,omitempty"`
	Experimental *SingboxExperimental `json:"experimental,omitempty"`
}

type SingboxLog struct {
	Level     string `json:"level,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

type SingboxDNS struct {
	Servers  []SingboxDNSServer `json:"servers,omitempty"`
	Rules    []interface{}      `json:"rules,omitempty"`
	Final    string             `json:"final,omitempty"`
	Strategy string             `json:"strategy,omitempty"`
}

type SingboxDNSServer struct {
	Tag     string `json:"tag"`
	Address string `json:"address"`
	Detour  string `json:"detour,omitempty"`
}

type SingboxInbound struct {
	Type                     string `json:"type"`
	Tag                      string `json:"tag"`
	Listen                   string `json:"listen,omitempty"`
	ListenPort               int    `json:"listen_port,omitempty"`
	Sniff                    bool   `json:"sniff,omitempty"`
	SniffOverrideDestination bool   `json:"sniff_override_destination,omitempty"`
}

type SingboxOutbound struct {
	Type       string            `json:"type"`
	Tag        string            `json:"tag"`
	Server     string            `json:"server,omitempty"`
	ServerPort int               `json:"server_port,omitempty"`
	UUID       string            `json:"uuid,omitempty"`
	Password   string            `json:"password,omitempty"`
	Method     string            `json:"method,omitempty"`
	AlterID    int               `json:"alter_id,omitempty"`
	Security   string            `json:"security,omitempty"`
	Network    string            `json:"network,omitempty"`
	TLS        *SingboxTLS       `json:"tls,omitempty"`
	Transport  *SingboxTransport `json:"transport,omitempty"`
	Flow       string            `json:"flow,omitempty"`
	Multiplex  *SingboxMultiplex `json:"multiplex,omitempty"`
}

type SingboxTLS struct {
	Enabled    bool            `json:"enabled"`
	ServerName string          `json:"server_name,omitempty"`
	Insecure   bool            `json:"insecure,omitempty"`
	ALPN       []string        `json:"alpn,omitempty"`
	UTLS       *SingboxUTLS    `json:"utls,omitempty"`
	Reality    *SingboxReality `json:"reality,omitempty"`
}

type SingboxUTLS struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type SingboxReality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type SingboxTransport struct {
	Type        string            `json:"type"`
	Path        string            `json:"path,omitempty"`
	Host        string            `json:"host,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
}

type SingboxMultiplex struct {
	Enabled    bool   `json:"enabled"`
	Protocol   string `json:"protocol,omitempty"`
	MaxStreams int    `json:"max_streams,omitempty"`
}

type SingboxRoute struct {
	Rules               []SingboxRule `json:"rules,omitempty"`
	Final               string        `json:"final,omitempty"`
	AutoDetectInterface bool          `json:"auto_detect_interface,omitempty"`
}

type SingboxRule struct {
	Protocol     string   `json:"protocol,omitempty"`
	Domain       []string `json:"domain,omitempty"`
	DomainSuffix []string `json:"domain_suffix,omitempty"`
	IPCIDR       []string `json:"ip_cidr,omitempty"`
	Outbound     string   `json:"outbound"`
}

type SingboxExperimental struct {
	ClashAPI *SingboxClashAPI `json:"clash_api,omitempty"`
}

type SingboxClashAPI struct {
	ExternalController string `json:"external_controller,omitempty"`
	ExternalUI         string `json:"external_ui,omitempty"`
}

// BuildSingboxConfig 根据选中的节点和路由规则生成 Sing-box 配置
func BuildSingboxConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, mixedPort int) (*SingboxConfig, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	cfg := &SingboxConfig{
		Log: &SingboxLog{
			Level:     "warn",
			Timestamp: true,
		},
		Inbounds: []SingboxInbound{
			{
				Type:                     "mixed",
				Tag:                      "mixed-in",
				Listen:                   "127.0.0.1",
				ListenPort:               mixedPort,
				Sniff:                    true,
				SniffOverrideDestination: false,
			},
		},
	}

	// 构建 outbound
	outbound, err := buildSingboxOutbound(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to build outbound: %w", err)
	}

	// 添加 direct outbound
	directOutbound := SingboxOutbound{
		Type: "direct",
		Tag:  "direct",
	}

	cfg.Outbounds = []SingboxOutbound{*outbound, directOutbound}

	// 构建路由
	cfg.Route = buildSingboxRoute(rules, configDir)

	return cfg, nil
}

// buildSingboxOutbound 根据节点信息构建 Sing-box outbound
func buildSingboxOutbound(p *database.Profile) (*SingboxOutbound, error) {
	outbound := &SingboxOutbound{
		Tag: "proxy",
	}

	switch p.ProxyProtocol {
	case "vmess":
		outbound.Type = "vmess"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential
		outbound.AlterID = p.ProxyAlterID
		outbound.Security = p.ProxySecurity
		if outbound.Security == "" {
			outbound.Security = "auto"
		}

	case "vless":
		outbound.Type = "vless"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential
		outbound.Flow = p.ProxyFlow

	case "trojan":
		outbound.Type = "trojan"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.Password = p.ProxyCredential

	case "shadowsocks":
		outbound.Type = "shadowsocks"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.Method = p.ProxySecurity
		outbound.Password = p.ProxyCredential

	case "hysteria2":
		outbound.Type = "hysteria2"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.Password = p.ProxyCredential

	case "hysteria":
		outbound.Type = "hysteria"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.Password = p.ProxyCredential

	case "tuic":
		outbound.Type = "tuic"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential
		outbound.Password = p.ProxySecurity

	case "wireguard":
		outbound.Type = "wireguard"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential     // Private key
		outbound.Password = p.ProxyHost // Interface address (e.g. 10.0.0.2/32)
		// PublicKey and Reserved stored in extra fields
		outbound.Security = p.ProxyPublicKey // Reuse Security field for public_key
		outbound.Network = p.ProxyPath       // Reuse Network field for reserved bytes

	case "anytls":
		outbound.Type = "anytls"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.Password = p.ProxyCredential

	case "socks":
		outbound.Type = "socks"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential         // Username
		outbound.Password = p.ProxySecurity // Password (reuse Security field)

	case "http":
		outbound.Type = "http"
		outbound.Server = p.ProxyAddress
		outbound.ServerPort = p.ProxyPort
		outbound.UUID = p.ProxyCredential         // Username
		outbound.Password = p.ProxySecurity // Password (reuse Security field)

	default:
		return nil, fmt.Errorf("unsupported protocol for sing-box: %s", p.ProxyProtocol)
	}

	// TLS 设置
	tls := buildSingboxTLS(p)
	if tls != nil {
		outbound.TLS = tls
	}

	// 传输层设�?
	transport := buildSingboxTransport(p)
	if transport != nil {
		outbound.Transport = transport
	}

	return outbound, nil
}

// buildSingboxTLS 构建 Sing-box TLS 设置
func buildSingboxTLS(p *database.Profile) *SingboxTLS {
	switch p.ProxyTLS {
	case "tls":
		tls := &SingboxTLS{
			Enabled:    true,
			ServerName: p.ProxySNI,
			Insecure:   p.ProxyAllowInsecure,
		}
		if p.ProxyFingerprint != "" {
			tls.UTLS = &SingboxUTLS{
				Enabled:     true,
				Fingerprint: p.ProxyFingerprint,
			}
		}
		return tls

	case "reality":
		tls := &SingboxTLS{
			Enabled:    true,
			ServerName: p.ProxySNI,
		}
		if p.ProxyFingerprint != "" {
			tls.UTLS = &SingboxUTLS{
				Enabled:     true,
				Fingerprint: p.ProxyFingerprint,
			}
		}
		tls.Reality = &SingboxReality{
			Enabled:   true,
			PublicKey: p.ProxyPublicKey,
			ShortID:   p.ProxyShortID,
		}
		return tls

	default:
		if p.ProxyProtocol == "trojan" || p.ProxyProtocol == "hysteria" || p.ProxyProtocol == "hysteria2" || p.ProxyProtocol == "tuic" {
			return &SingboxTLS{
				Enabled:    true,
				ServerName: p.ProxySNI,
				Insecure:   p.ProxyAllowInsecure,
			}
		}
	}

	return nil
}

// buildSingboxTransport 构建 Sing-box 传输层设�?
func buildSingboxTransport(p *database.Profile) *SingboxTransport {
	switch p.ProxyNetwork {
	case "ws":
		transport := &SingboxTransport{
			Type: "ws",
			Path: p.ProxyPath,
		}
		if p.ProxyHost != "" {
			transport.Headers = map[string]string{
				"Host": p.ProxyHost,
			}
		}
		return transport

	case "h2":
		return &SingboxTransport{
			Type: "http",
			Path: p.ProxyPath,
			Host: p.ProxyHost,
		}

	case "grpc":
		return &SingboxTransport{
			Type:        "grpc",
			ServiceName: p.ProxyPath,
		}

	case "tcp":
		if p.ProxyHost != "" {
			return &SingboxTransport{
				Type: "http",
				Path: p.ProxyPath,
				Host: p.ProxyHost,
			}
		}
	}

	return nil
}

// buildSingboxRoute 构建 Sing-box 路由规则
func buildSingboxRoute(rules []database.RoutingRule, configDir string) *SingboxRoute {
	route := &SingboxRoute{
		AutoDetectInterface: true,
		Final:               "proxy",
		Rules: []SingboxRule{
			// 默认直连规则：局域网 IP
			{
				IPCIDR:   []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
				Outbound: "direct",
			},
		},
	}

	// 仅在 dat 文件存在时添�?geo 规则
	hasGeoIP, hasGeoSite := hasGeoDatFiles(configDir)
	if hasGeoSite {
		route.Rules = append(route.Rules, SingboxRule{
			Domain:   []string{"geosite:cn"},
			Outbound: "direct",
		})
	}
	if hasGeoIP {
		route.Rules = append(route.Rules, SingboxRule{
			IPCIDR:   []string{"geoip:cn"},
			Outbound: "direct",
		})
	}

	// 添加用户自定义规�?
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		sbRule := SingboxRule{
			Outbound: rule.Type,
		}

		if rule.Domain != "" {
			sbRule.Domain = splitAndTrim(rule.Domain)
		}
		if rule.IP != "" {
			sbRule.IPCIDR = splitAndTrim(rule.IP)
		}

		route.Rules = append(route.Rules, sbRule)
	}

	return route
}

// SaveSingboxConfig 生成并保�?Sing-box 配置文件
func SaveSingboxConfig(profile *database.Profile, rules []database.RoutingRule, configDir string, mixedPort int) (string, error) {
	cfg, err := BuildSingboxConfig(profile, rules, configDir, mixedPort)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, "singbox_config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return configPath, nil
}
