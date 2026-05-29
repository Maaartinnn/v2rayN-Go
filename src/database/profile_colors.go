package database

// 协议徽标颜色映射（背景色 + 文字色）
// 修改颜色值只需改这里，前端零改动。
var protocolColorMap = map[string]ColorPair{
	"vmess":        {Bg: "rgba(106, 155, 204, 0.12)", Text: "#6A9BCC"},
	"vless":        {Bg: "rgba(217, 119, 87, 0.12)", Text: "#D97757"},
	"trojan":       {Bg: "rgba(201, 148, 58, 0.12)", Text: "#C9943A"},
	"shadowsocks":  {Bg: "rgba(107, 143, 71, 0.12)", Text: "#6B8F47"},
	"shadowsocksr": {Bg: "rgba(107, 143, 71, 0.12)", Text: "#6B8F47"},
	"hysteria2":    {Bg: "rgba(192, 69, 58, 0.12)", Text: "#C0453A"},
	"hysteria":     {Bg: "rgba(192, 69, 58, 0.12)", Text: "#C0453A"},
	"tuic":         {Bg: "rgba(120, 86, 188, 0.12)", Text: "#7856BC"},
	"wireguard":    {Bg: "rgba(86, 155, 132, 0.12)", Text: "#569B84"},
	"anytls":       {Bg: "rgba(140, 120, 200, 0.12)", Text: "#8C78C8"},

	// 策略组（虚拟节点）
	"selector":    {Bg: "rgba(59, 130, 246, 0.12)", Text: "#3B82F6"},
	"urltest":     {Bg: "rgba(16, 185, 129, 0.12)", Text: "#10B981"},
	"fallback":    {Bg: "rgba(245, 158, 11, 0.12)", Text: "#F59E0B"},
	"loadbalance": {Bg: "rgba(139, 92, 246, 0.12)", Text: "#8B5CF6"},
}

// 默认协议颜色（未知协议）
var defaultProtocolColor = ColorPair{
	Bg:   "var(--color-muted)",
	Text: "var(--color-muted-foreground)",
}

// 内核徽标颜色映射
var coreColorMap = map[string]ColorPair{
	"xray":     {Bg: "rgba(106, 155, 204, 0.12)", Text: "#6A9BCC"},
	"sing-box": {Bg: "rgba(140, 100, 200, 0.12)", Text: "#8C64C8"},
	"mihomo":   {Bg: "rgba(217, 119, 87, 0.12)", Text: "#D97757"},
}

// 默认内核颜色（自动/空）
var defaultCoreColor = ColorPair{
	Bg:   "var(--color-muted)",
	Text: "var(--color-muted-foreground)",
}

// GetProtocolColor 根据协议类型返回徽标颜色。
func GetProtocolColor(protocol string) ColorPair {
	if c, ok := protocolColorMap[protocol]; ok {
		return c
	}
	return defaultProtocolColor
}

// GetCoreColor 根据内核类型返回徽标颜色。
// coreType 为空时表示自动模式，使用默认颜色。
func GetCoreColor(coreType string) ColorPair {
	if coreType == "" {
		return defaultCoreColor
	}
	if c, ok := coreColorMap[coreType]; ok {
		return c
	}
	return defaultCoreColor
}

// GetLatencyColor 根据测速结果返回延迟指示灯颜色（CSS 变量名）。
//   - 空或 "timeout" → error
//   - < 100ms → success
//   - < 300ms → warning
//   - ≥ 300ms → error
func GetLatencyColor(result string) string {
	if result == "" || result == "timeout" {
		return "var(--color-error)"
	}
	// 简单整数解析，避免引入 strconv 依赖
	ms := 0
	for _, c := range result {
		if c >= '0' && c <= '9' {
			ms = ms*10 + int(c-'0')
		} else {
			// 非数字字符，视为无效
			return "var(--color-error)"
		}
	}
	if ms < 100 {
		return "var(--color-success)"
	}
	if ms < 300 {
		return "var(--color-warning)"
	}
	return "var(--color-error)"
}
