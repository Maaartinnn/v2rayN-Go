package database

// ColorPair 颜色对（背景色 + 文字色），用于前端徽标渲染。
type ColorPair struct {
	Bg   string `json:"bg"`   // 背景色（rgba 或 CSS 变量）
	Text string `json:"text"` // 文字色（hex 或 CSS 变量）
}

// ProfileListItem 节点列表精简数据，仅包含前端列表展示和操作所需的字段。
// 编辑节点时通过 GET /api/profiles/{uuid} 按需获取完整 Profile。
type ProfileListItem struct {
	// 用 uuid 作为前端唯一标识（列表 key、多选、激活、删除等操作均依赖 uuid）
	UUID string `json:"uuid"`

	// 展示字段
	Name          string `json:"name"`           // 节点名称
	ProxyProtocol string `json:"proxy_protocol"` // 协议类型
	ProxyAddress  string `json:"proxy_address"`  // 服务器地址
	ProxyPort     int    `json:"proxy_port"`     // 服务器端口
	CoreType      string `json:"core_type"`      // 内核类型（空 = 自动）

	// 状态字段
	TestResult string `json:"test_result"` // 测速结果
	IsActive   bool   `json:"is_active"`   // 是否激活
	GroupUUID  string `json:"group_uuid"`  // 所属分组 UUID

	// 策略组相关字段
	NodeType    string `json:"node_type"`    // "proxy" | 策略组类型（selector/urltest/fallback/loadbalance）
	MemberCount int    `json:"member_count"` // 策略组成员数量（普通节点 = 0）

	// 后端计算的颜色字段（前端直接使用，无需自行判断）
	ProtocolColor ColorPair `json:"protocol_color"` // 协议徽标颜色
	CoreColor     ColorPair `json:"core_color"`     // 内核徽标颜色
	LatencyColor  string    `json:"latency_color"`  // 延迟指示灯颜色（CSS 变量）
}
