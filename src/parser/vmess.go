package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"v2rayn-go/database"
)

// vmessJSON VMess 标准格式 (V2RayN 标准)
type vmessJSON struct {
	V    string      `json:"v"`
	Ps   string      `json:"ps"`
	Add  string      `json:"add"`
	Port interface{} `json:"port"`
	ID   string      `json:"id"`
	Aid  interface{} `json:"aid"`
	Scy  string      `json:"scy"`
	Net  string      `json:"net"`
	Type string      `json:"type"`
	Host string      `json:"host"`
	Path string      `json:"path"`
	TLS  string      `json:"tls"`
	SNI  string      `json:"sni"`
	ALPN string      `json:"alpn"`
	Fp   string      `json:"fp"`
}

// parseVmess 解析 vmess:// 链接
// 支持两种格式：
// 1. vmess://base64(JSON) - V2RayN 标准格式
// 2. vmess://base64(uuid@host:port?params)#name - URI 格式
func parseVmess(link string) (*database.Profile, error) {
	// 移除 vmess:// 前缀
	data := link[8:]

	// 尝试标准 JSON 格式
	decoded, err := base64Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vmess link: %w", err)
	}

	// 检查是否是 JSON 格式
	decoded = strings.TrimSpace(decoded)
	if strings.HasPrefix(decoded, "{") {
		return parseVmessJSON(decoded, link)
	}

	// URI 格式: vmess://base64(uuid@host:port?params)#name
	return parseVmessURI(decoded, link)
}

// parseVmessJSON 解析 VMess JSON 格式
func parseVmessJSON(jsonStr string, rawLink string) (*database.Profile, error) {
	var vj vmessJSON
	if err := json.Unmarshal([]byte(jsonStr), &vj); err != nil {
		return nil, fmt.Errorf("failed to parse vmess json: %w", err)
	}

	port := 0
	switch p := vj.Port.(type) {
	case float64:
		port = int(p)
	case string:
		port = parseIntSafe(p, 443)
	}

	aid := 0
	switch a := vj.Aid.(type) {
	case float64:
		aid = int(a)
	case string:
		aid = parseIntSafe(a, 0)
	}

	profile := &database.Profile{
		Name:        vj.Ps,
		Address:     vj.Add,
		Port:        port,
		Protocol:    "vmess",
		UUID:        vj.ID,
		AlterID:     aid,
		Security:    vj.Scy,
		Network:     vj.Net,
		TLS:         vj.TLS,
		SNI:         vj.SNI,
		Fingerprint: vj.Fp,
		Host:        vj.Host,
		Path:        vj.Path,
		RawLink:     rawLink,
	}

	if profile.Security == "" {
		profile.Security = "auto"
	}
	if profile.Network == "" {
		profile.Network = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseVmessURI 解析 VMess URI 格式
func parseVmessURI(decoded string, rawLink string) (*database.Profile, error) {
	// 格式: uuid@host:port?security=xxx&type=xxx&host=xxx&path=xxx&tls=xxx&sni=xxx#name
	// 补全为标准 URI
	if !strings.Contains(decoded, "://") {
		decoded = "vmess://" + decoded
	}

	u, err := parseURL(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vmess uri: %w", err)
	}

	name := extractNameFromLink(rawLink)
	if name == "" {
		name = extractNameFromFragment(u)
	}

	host := u.Hostname()
	port := parseIntSafe(u.Port(), 443)

	q := u.Query()
	profile := &database.Profile{
		Name:        name,
		Address:     host,
		Port:        port,
		Protocol:    "vmess",
		UUID:        u.User.Username(),
		Security:    q.Get("security"),
		Network:     q.Get("type"),
		TLS:         q.Get("tls"),
		SNI:         q.Get("sni"),
		Fingerprint: q.Get("fp"),
		Host:        q.Get("host"),
		Path:        q.Get("path"),
		RawLink:     rawLink,
	}

	if profile.Security == "" {
		profile.Security = "auto"
	}
	if profile.Network == "" {
		profile.Network = "tcp"
	}
	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}
