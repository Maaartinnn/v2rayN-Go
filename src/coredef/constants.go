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
