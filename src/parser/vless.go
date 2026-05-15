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
		UUID:          database.GenerateUUID(),
		Name:        name,
		ProxyAddress:     host,
		ProxyPort:        port,
		ProxyProtocol:    "vless",
		ProxyCredential:        u.User.Username(),
		ProxyNetwork:     q.Get("type"),
		ProxyTLS:         q.Get("security"),
		ProxySNI:         q.Get("sni"),
		ProxyFingerprint: q.Get("fp"),
		ProxyFlow:        q.Get("flow"),
		ProxyHost:        q.Get("host"),
		ProxyPath:        q.Get("path"),
		ProxySeed:        q.Get("seed"),
		ProxyPublicKey:   q.Get("pbk"),
		ProxyShortID:     q.Get("sid"),
		ProxySiderSNI:    q.Get("sni"),
		RawLink:     link,
	}

	// Reality 模式下的特殊参数
	if q.Get("sni") != "" {
		profile.ProxySNI = q.Get("sni")
	}

	if profile.ProxyNetwork == "" {
		profile.ProxyNetwork = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.ProxyAddress, profile.ProxyPort)
	}

	return profile, nil
}
