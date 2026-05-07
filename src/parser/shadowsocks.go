package parser

import (
	"fmt"
	"strings"

	"v2rayn-go/database"
)

// parseShadowsocks 解析 ss:// 链接
// 格式1: ss://base64(method:password)@host:port#name
// 格式2: ss://base64(method:password@host:port)#name
// 格式3: ss://base64(method:password)@host:port?plugin=xxx#name
func parseShadowsocks(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	cleanLink := removeNameSuffix(link)

	// 移除 ss:// 前缀
	data := cleanLink[5:]

	// 尝试解析为 URI 格式: ss://base64(method:password)@host:port?params
	if idx := strings.Index(data, "@"); idx >= 0 {
		return parseShadowsocksURI(data, idx, name, link)
	}

	// 整体 base64 格式: ss://base64(method:password@host:port)
	decoded, err := base64Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ss link: %w", err)
	}

	// 解析 method:password@host:port
	return parseShadowsocksDecoded(decoded, name, link)
}

// parseShadowsocksURI 解析 URI 格式的 SS 链接
func parseShadowsocksURI(data string, atIdx int, name string, rawLink string) (*database.Profile, error) {
	// base64(method:password)@host:port
	userPart := data[:atIdx]
	hostPart := data[atIdx+1:]

	// 解码用户信息部分
	decoded, err := base64Decode(userPart)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ss userinfo: %w", err)
	}

	// method:password
	colonIdx := strings.Index(decoded, ":")
	if colonIdx < 0 {
		return nil, fmt.Errorf("invalid ss userinfo format")
	}
	method := decoded[:colonIdx]
	password := decoded[colonIdx+1:]

	// 解析 host:port?params
	fullHost := hostPart
	query := ""
	if qIdx := strings.Index(fullHost, "?"); qIdx >= 0 {
		query = fullHost[qIdx+1:]
		fullHost = fullHost[:qIdx]
	}

	host := fullHost
	port := 8388 // 默认 SS 端口
	if cIdx := strings.LastIndex(fullHost, ":"); cIdx >= 0 {
		host = fullHost[:cIdx]
		port = parseIntSafe(fullHost[cIdx+1:], 8388)
	}

	profile := &database.Profile{
		Name:     name,
		Address:  host,
		Port:     port,
		Protocol: "shadowsocks",
		Security: method,
		UUID:     password,
		RawLink:  rawLink,
	}

	// 解析插件参数
	if query != "" {
		params := parseQueryString(query)
		if plugin := params["plugin"]; plugin != "" {
			profile.Path = plugin
		}
	}

	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseShadowsocksDecoded 解析已解码的 SS 链接
func parseShadowsocksDecoded(decoded string, name string, rawLink string) (*database.Profile, error) {
	// 格式: method:password@host:port
	atIdx := strings.LastIndex(decoded, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("invalid ss format: missing @")
	}

	userPart := decoded[:atIdx]
	hostPart := decoded[atIdx+1:]

	colonIdx := strings.Index(userPart, ":")
	if colonIdx < 0 {
		return nil, fmt.Errorf("invalid ss format: missing method:password")
	}

	method := userPart[:colonIdx]
	password := userPart[colonIdx+1:]

	host := hostPart
	port := 8388
	if cIdx := strings.LastIndex(hostPart, ":"); cIdx >= 0 {
		host = hostPart[:cIdx]
		port = parseIntSafe(hostPart[cIdx+1:], 8388)
	}

	profile := &database.Profile{
		Name:     name,
		Address:  host,
		Port:     port,
		Protocol: "shadowsocks",
		Security: method,
		UUID:     password,
		RawLink:  rawLink,
	}

	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseShadowsocksR 解析 ssr:// 链接
// 格式: ssr://base64(host:port:protocol:method:obfs:base64pass/?params)
func parseShadowsocksR(link string) (*database.Profile, error) {
	name := extractNameFromLink(link)
	data := link[6:] // 移除 ssr://

	decoded, err := base64Decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ssr link: %w", err)
	}

	// 分离参数部分
	mainPart := decoded
	queryPart := ""
	if qIdx := strings.Index(decoded, "/?"); qIdx >= 0 {
		mainPart = decoded[:qIdx]
		queryPart = decoded[qIdx+2:]
	}

	// 解析主体: host:port:protocol:method:obfs:base64pass
	parts := strings.Split(mainPart, ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid ssr format: expected at least 6 parts, got %d", len(parts))
	}

	host := parts[0]
	port := parseIntSafe(parts[1], 443)
	protocol := parts[2]
	method := parts[3]
	obfs := parts[4]

	// 解码密码
	password, err := base64Decode(parts[5])
	if err != nil {
		password = parts[5]
	}

	profile := &database.Profile{
		Name:     name,
		Address:  host,
		Port:     port,
		Protocol: "shadowsocksr",
		Security: method,
		UUID:     password,
		Network:  protocol,
		Path:     obfs,
		RawLink:  link,
	}

	// 解析附加参数
	if queryPart != "" {
		params := parseQueryString(queryPart)
		if obfsParam := params["obfsparam"]; obfsParam != "" {
			decoded, _ := base64Decode(obfsParam)
			profile.Host = decoded
		}
		if protoParam := params["protoparam"]; protoParam != "" {
			decoded, _ := base64Decode(protoParam)
			profile.Seed = decoded
		}
		if remarks := params["remarks"]; remarks != "" {
			decoded, _ := base64Decode(remarks)
			if decoded != "" {
				profile.Name = decoded
			}
		}
	}

	if profile.Name == "" {
		profile.Name = fmt.Sprintf("%s:%d", profile.Address, profile.Port)
	}

	return profile, nil
}

// parseQueryString 解析简单的 key=value&key2=value2 格式
func parseQueryString(query string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}
