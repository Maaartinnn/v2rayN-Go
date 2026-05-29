package coredef

import "time"

// === 默认网络参数 ===
const (
	DefaultWebPort    = 2017
	DefaultSocksPort  = 10808
	DefaultHTTPPort   = 10809
	DefaultListenIP   = "127.0.0.1"
	DefaultOutboundIP = "0.0.0.0"
)

// === 内核管理 ===
const (
	CoreStopTimeout      = 5 * time.Second
	CoreLogChannelBuffer = 100
)

// === HTTP 业务限制 ===
const (
	MultipartMaxMemoryDefault = 10 << 20  // 10MB（图片上传）
	MultipartMaxMemoryCore    = 200 << 20 // 200MB（内核上传）
	PingAllConcurrency        = 20
)

// === 代理协议类型 ===
const (
	// 真实代理协议
	ProtocolVMess       = "vmess"
	ProtocolVLESS       = "vless"
	ProtocolTrojan      = "trojan"
	ProtocolShadowsocks = "shadowsocks"
	ProtocolHysteria2   = "hysteria2"
	ProtocolTUIC        = "tuic"
	ProtocolWireGuard   = "wireguard"
	ProtocolSocks       = "socks"
	ProtocolHTTP        = "http"

	// 策略组类型（虚拟节点）
	ProtocolSelector    = "selector"
	ProtocolURLTest     = "urltest"
	ProtocolFallback    = "fallback"
	ProtocolLoadBalance = "loadbalance"
)

// === 负载均衡策略 ===
const (
	StrategyRoundRobin = "round-robin"
	StrategyLeastLoad  = "least-load"
	StrategyRandom     = "random"
)

// IsStrategyProtocol 判断协议类型是否为策略组
func IsStrategyProtocol(proto string) bool {
	switch proto {
	case ProtocolSelector, ProtocolURLTest, ProtocolFallback, ProtocolLoadBalance:
		return true
	}
	return false
}

// IsProxyProtocol 判断协议类型是否为真实代理协议
func IsProxyProtocol(proto string) bool {
	return proto != "" && !IsStrategyProtocol(proto)
}
