package parser

import (
	"fmt"

	"v2rayn-go/database"
)

// parseTrojan 解析 trojan:// 链接
// 格式: trojan://password@host:port?security=xxx&type=xxx&...#name
func parseTrojan(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trojan link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)
	q := u.Query()

	password := ""
	if u.User != nil {
		password = u.User.Username()
	}

	profile := &database.Profile{
		Name:          name,
		ProxyAddress:       host,
		ProxyPort:          port,
		ProxyProtocol:      "trojan",
		ProxyCredential:          password,
		ProxyNetwork:       q.Get("type"),
		ProxyTLS:           "tls",
		ProxySNI:           q.Get("sni"),
		ProxyFingerprint:   q.Get("fp"),
		ProxyAllowInsecure: q.Get("allowInsecure") == "1" || q.Get("allowInsecure") == "true",
		ProxyHost:          q.Get("host"),
		ProxyPath:          q.Get("path"),
		ProxyFlow:          q.Get("flow"),
		RawLink:       link,
	}

	// Trojan 支持非 TLS 模式
	if q.Get("security") == "tls" || q.Get("security") == "" {
		profile.ProxyTLS = "tls"
	} else if q.Get("security") == "reality" {
		profile.ProxyTLS = "reality"
		profile.ProxyPublicKey = q.Get("pbk")
		profile.ProxyShortID = q.Get("sid")
	}

	if profile.ProxyNetwork == "" {
		profile.ProxyNetwork = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.ProxyAddress, profile.ProxyPort)
	}

	return profile, nil
}
