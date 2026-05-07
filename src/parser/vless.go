package parser

import (
	"fmt"

	"v2rayn-go/database"
)

// parseVless 解析 vless:// 链接
// 格式: vless://uuid@host:port?type=xxx&security=xxx&...#name
func parseVless(link string) (*database.Profile, error) {
	// 移除 #name 部分用于解析
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vless link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)
	q := u.Query()

	profile := &database.Profile{
		Name:        name,
		Address:     host,
		Port:        port,
		Protocol:    "vless",
		UUID:        u.User.Username(),
		Network:     q.Get("type"),
		TLS:         q.Get("security"),
		SNI:         q.Get("sni"),
		Fingerprint: q.Get("fp"),
		Flow:        q.Get("flow"),
		Host:        q.Get("host"),
		Path:        q.Get("path"),
		Seed:        q.Get("seed"),
		PublicKey:   q.Get("pbk"),
		ShortID:     q.Get("sid"),
		SiderSNI:    q.Get("sni"),
		RawLink:     link,
	}

	// Reality 模式下的特殊参数
	if q.Get("sni") != "" {
		profile.SNI = q.Get("sni")
	}

	if profile.Network == "" {
		profile.Network = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
