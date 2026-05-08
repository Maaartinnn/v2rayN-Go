package parser

import (
	"fmt"

	"v2rayn-go/database"
)

// parseWireGuard 解析 wireguard:// 链接
// 格式: wireguard://privateKey@host:port?publicKey=xxx&reserved=xxx&address=xxx#name
func parseWireGuard(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	u, err := parseURL(cleanLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wireguard link: %w", err)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 51820)
	q := u.Query()

	// privateKey is in the userinfo part
	privateKey := u.User.Username()

	profile := &database.Profile{
		Name:      name,
		Address:   host,
		Port:      port,
		Protocol:  "wireguard",
		UUID:      privateKey, // Store private key in UUID field
		PublicKey: q.Get("public_key"),
		// WireGuard specific: store reserved bytes and address in extra fields
		Path:    q.Get("reserved"), // Reserved bytes stored in Path
		Host:    q.Get("address"),  // WireGuard interface address stored in Host
		RawLink: link,
	}

	// Also check for shorter param names
	if profile.PublicKey == "" {
		profile.PublicKey = q.Get("pk")
	}
	if profile.Path == "" {
		profile.Path = q.Get("reserved")
	}
	if profile.Host == "" {
		profile.Host = q.Get("addr")
	}

	if profile.Name == "" {
		profile.Name = fmt.Sprintf("wg-%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
