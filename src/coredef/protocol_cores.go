// protocol_cores.go — 协议与内核兼容性映射
//
// 定义每种代理协议支持哪些内核，以及推荐优先级顺序。
// 此映射表是唯一事实来源（Single Source of Truth），前端 coreMap.ts 和
// 后端 CoreService.Start() 均依赖此表选择正确的内核。
//
// 注意事项：
//   - 每个协议的内核列表按推荐优先级排序（第一个是最佳推荐）
//   - Mihomo 支持 VLESS 协议（从 Clash.Meta 内核开始支持）
//   - 新增协议或内核时只需修改此文件

package coredef

// ProtocolCoreMap 协议 → 支持的内核列表（按推荐优先级排序）
//
// 用途：
//   - NodeEditForm 编辑页面：决定节点的可用内核列表 (core_list)
//   - CoreService.Start()：当用户未指定 coreType 时，选择最佳默认内核
//
// 推荐优先级策略：
//   - Xray 对 VMess/VLESS/Trojan 支持最成熟，排第一
//   - Sing-box 对 Hysteria2/TUIC/WireGuard/AnyTLS 原生支持最好，排第一
//   - Mihomo 对 ShadowsocksR 独家支持，排第一
//   - 多内核支持的协议，按社区使用率和稳定性排序
var ProtocolCoreMap = map[string][]CoreType{
	"vmess":        {TypeXray, TypeSingBox, TypeMihomo},
	"vless":        {TypeXray, TypeSingBox, TypeMihomo}, // Mihomo 从 Clash.Meta 起支持 VLESS
	"trojan":       {TypeXray, TypeSingBox, TypeMihomo},
	"shadowsocks":  {TypeXray, TypeSingBox, TypeMihomo},
	"shadowsocksr": {TypeMihomo}, // SSR 仅 Mihomo 支持
	"hysteria2":    {TypeSingBox, TypeMihomo},
	"hysteria":     {TypeSingBox, TypeMihomo},
	"tuic":         {TypeSingBox, TypeMihomo},
	"wireguard":    {TypeSingBox, TypeMihomo},
	"anytls":       {TypeSingBox}, // AnyTLS 仅 Sing-box 支持
	"socks":        {TypeXray, TypeSingBox},
	"http":         {TypeXray, TypeSingBox},
}

// GetSupportedCoresForProtocol 获取指定协议支持的内核列表（按推荐优先级排序）。
// 如果协议未知，返回 nil。
//
// 典型用法：
//
//	cores := coredef.GetSupportedCoresForProtocol("vmess")
//	// 返回: [xray sing-box mihomo]
func GetSupportedCoresForProtocol(protocol string) []CoreType {
	return ProtocolCoreMap[protocol]
}
