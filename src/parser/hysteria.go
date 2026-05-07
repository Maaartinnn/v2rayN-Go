package parser

import (
	"fmt"

	"v2rayn-go/database"
)

// parseHysteria2 解析 hysteria2:// 或 hy2:// 链接
// 格式: hysteria2://password@host:port?insecure=1&sni=xxx&obfs=xxx&obfs-password=xxx#name
func parseHysteria2(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	// 统一为 hysteria2://
	if len(cleanLink) > 5 && cleanLink[:5] == "hy2:/" {
		cleanLink = "hysteria2" + cleanLink[2:]
	}

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hysteria2 link: %w", err)
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
		Protocol:      "hysteria2",
		UUID:          password,
		TLS:           "tls",
		SNI:           q.Get("sni"),
		AllowInsecure: q.Get("insecure") == "1" || q.Get("insecure") == "true",
		Fingerprint:   q.Get("fp"),
		RawLink:       link,
	}

	// OBFS 支持
	if obfs := q.Get("obfs"); obfs != "" {
		profile.Network = obfs
		profile.Path = q.Get("obfs-password")
	}

	if profile.SNI == "" {
		profile.SNI = host
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseHysteria 解析 hysteria:// 链接
// 格式: hysteria://host:port?auth=xxx&sni=xxx&insecure=1&upmbps=100&downmbps=100&obfs=xxx#name
func parseHysteria(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hysteria link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)
	q := u.Query()

	// Hysteria 1 的认证信息可能在 auth 参数或 userinfo 中
	auth := q.Get("auth")
	if auth == "" && u.User != nil {
		auth = u.User.Username()
	}

	profile := &database.Profile{
		Name:          name,
		Address:       host,
		Port:          port,
		Protocol:      "hysteria",
		UUID:          auth,
		TLS:           "tls",
		SNI:           q.Get("sni"),
		AllowInsecure: q.Get("insecure") == "1" || q.Get("insecure") == "true",
		Fingerprint:   q.Get("fp"),
		RawLink:       link,
	}

	if obfs := q.Get("obfs"); obfs != "" {
		profile.Network = obfs
		profile.Path = q.Get("obfs-password")
	}

	if profile.SNI == "" {
		profile.SNI = host
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseTuic 解析 tuic:// 链接
// 格式: tuic://uuid:password@host:port?congestion_control=xxx&sni=xxx&alpn=xxx&allow_insecure=1#name
func parseTuic(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tuic link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)
	q := u.Query()

	// 解析 uuid:password
	uuid := ""
	password := ""
	if u.User != nil {
		uuid = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	profile := &database.Profile{
		Name:          name,
		Address:       host,
		Port:          port,
		Protocol:      "tuic",
		UUID:          uuid,
		Security:      password,
		TLS:           "tls",
		SNI:           q.Get("sni"),
		AllowInsecure: q.Get("allow_insecure") == "1" || q.Get("allow_insecure") == "true",
		Fingerprint:   q.Get("fp"),
		Network:       q.Get("congestion_control"),
		RawLink:       link,
	}

	if profile.SNI == "" {
		profile.SNI = host
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
