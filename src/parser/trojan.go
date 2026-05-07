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
		Address:       host,
		Port:          port,
		Protocol:      "trojan",
		UUID:          password,
		Network:       q.Get("type"),
		TLS:           "tls",
		SNI:           q.Get("sni"),
		Fingerprint:   q.Get("fp"),
		AllowInsecure: q.Get("allowInsecure") == "1" || q.Get("allowInsecure") == "true",
		Host:          q.Get("host"),
		Path:          q.Get("path"),
		Flow:          q.Get("flow"),
		RawLink:       link,
	}

	// Trojan 支持非 TLS 模式
	if q.Get("security") == "tls" || q.Get("security") == "" {
		profile.TLS = "tls"
	} else if q.Get("security") == "reality" {
		profile.TLS = "reality"
		profile.PublicKey = q.Get("pbk")
		profile.ShortID = q.Get("sid")
	}

	if profile.Network == "" {
		profile.Network = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
