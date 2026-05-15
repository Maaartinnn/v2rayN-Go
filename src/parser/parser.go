package parser

import (
	"fmt"
	"strings"

	"v2rayn-go/database"
)

// ParseLink 解析分享链接并返回 Profile
func ParseLink(link string) (*database.Profile, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil, fmt.Errorf("empty link")
	}

	// 根据协议前缀分发到对应的解析器
	switch {
	case strings.HasPrefix(link, "vmess://"):
		return parseVmess(link)
	case strings.HasPrefix(link, "vless://"):
		return parseVless(link)
	case strings.HasPrefix(link, "trojan://"):
		return parseTrojan(link)
	case strings.HasPrefix(link, "ss://"):
		return parseShadowsocks(link)
	case strings.HasPrefix(link, "ssr://"):
		return parseShadowsocksR(link)
	case strings.HasPrefix(link, "hysteria2://") || strings.HasPrefix(link, "hy2://"):
		return parseHysteria2(link)
	case strings.HasPrefix(link, "hysteria://"):
		return parseHysteria(link)
	case strings.HasPrefix(link, "tuic://"):
		return parseTuic(link)
	case strings.HasPrefix(link, "wireguard://") || strings.HasPrefix(link, "wg://"):
		return parseWireGuard(link)
	case strings.HasPrefix(link, "anytls://"):
		return parseAnytls(link)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", link[:min(len(link), 20)])
	}
}

// ParseLinks 批量解析分享链接
func ParseLinks(links []string) ([]*database.Profile, error) {
	var profiles []*database.Profile
	var errs []string

	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}
		profile, err := ParseLink(link)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to parse %s: %v", truncate(link, 50), err))
			continue
		}
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all links failed to parse:\n%s", strings.Join(errs, "\n"))
	}

	return profiles, nil
}

// ParseSubscriptionContent 解析订阅内容（可能是 base64 编码的链接列表）
func ParseSubscriptionContent(content string) ([]*database.Profile, error) {
	content = strings.TrimSpace(content)

	// 尝试 base64 解码
	decoded, err := base64Decode(content)
	if err == nil {
		content = decoded
	}

	// 按行分割
	lines := strings.Split(content, "\n")
	var links []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			links = append(links, line)
		}
	}

	return ParseLinks(links)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
