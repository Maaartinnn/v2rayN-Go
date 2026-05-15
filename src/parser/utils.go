package parser

import (
	"encoding/base64"
	"net/url"
	"strconv"
	"strings"
)

// base64Decode base64 解码
func base64Decode(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(s)
	}
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(s)
	}
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(s)
	}
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// parseURL 解析 URL 并处理可能的错误
func parseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// getQueryParam 获取 URL 查询参数
func getQueryParam(u *url.URL, key string) string {
	return u.Query().Get(key)
}

// parseIntSafe 安全地解析整数，失败返回默认值
func parseIntSafe(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// extractNameFromFragment 从 URL fragment 中提取名称
func extractNameFromFragment(u *url.URL) string {
	frag := u.Fragment
	if frag == "" {
		return ""
	}
	// URL decode
	decoded, err := url.QueryUnescape(frag)
	if err != nil {
		return frag
	}
	return decoded
}

// extractNameFromLink 从链接末尾提取名称（#fragment）
func extractNameFromLink(link string) string {
	idx := strings.LastIndex(link, "#")
	if idx < 0 {
		return ""
	}
	name := link[idx+1:]
	decoded, err := url.QueryUnescape(name)
	if err != nil {
		return name
	}
	return decoded
}

// removeNameSuffix 移除链接中的 #name 部分
func removeNameSuffix(link string) string {
	idx := strings.LastIndex(link, "#")
	if idx < 0 {
		return link
	}
	return link[:idx]
}
