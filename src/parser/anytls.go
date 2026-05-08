package parser

import (
	"fmt"

	"v2rayn-go/database"
)

// parseAnytls 解析 anytls:// 链接
// 格式: anytls://password@host:port?security=xxx&sni=xxx#name
func parseAnytls(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse anytls link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)
	q := u.Query()

	password := u.User.Username()

	profile := &database.Profile{
		Name:          name,
		Address:       host,
		Port:          port,
		Protocol:      "anytls",
		UUID:          password,
		TLS:           "tls",
		SNI:           q.Get("sni"),
		Fingerprint:   q.Get("fp"),
		AllowInsecure: q.Get("allowInsecure") == "1" || q.Get("allowInsecure") == "true",
		RawLink:       link,
	}

	if profile.SNI == "" {
		profile.SNI = host
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("anytls-%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
