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
		UUID:          database.GenerateUUID(),
		Name:      name,
		ProxyAddress:   host,
		ProxyPort:      port,
		ProxyProtocol:  "wireguard",
		ProxyCredential:      privateKey, // Store private key in UUID field
		ProxyPublicKey: q.Get("public_key"),
		// WireGuard specific: store reserved bytes and address in extra fields
		ProxyPath:    q.Get("reserved"), // Reserved bytes stored in Path
		ProxyHost:    q.Get("address"),  // WireGuard interface address stored in Host
		RawLink: link,
	}

	// Also check for shorter param names
	if profile.ProxyPublicKey == "" {
		profile.ProxyPublicKey = q.Get("pk")
	}
	if profile.ProxyPath == "" {
		profile.ProxyPath = q.Get("reserved")
	}
	if profile.ProxyHost == "" {
		profile.ProxyHost = q.Get("addr")
	}

	if profile.Name == "" {
		profile.Name = fmt.Sprintf("wg-%s:%d", profile.ProxyAddress, profile.ProxyPort)
	}

	return profile, nil
}
