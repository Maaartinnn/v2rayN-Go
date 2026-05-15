package httpclient

import (
	"net/http"
	"time"
)

const DefaultUserAgent = "v2rayN-Go/1.0"

// userAgentTransport 包装底层 Transport，在 RoundTrip 中自动注入 User-Agent。
// 遵循 Go 标准库规范：不修改原始 *http.Request，
// 而是通过 req.Clone() 创建副本后修改，避免 Data Race。
type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		// req.Clone() 会浅拷贝 Request 并深拷贝 Header（新建 map 和 []string），
		// Go string 不可变，因此共享底层字符串是安全的。
		// 在 clone 上 Set 不会影响原始请求。
		clonedReq := req.Clone(req.Context())
		clonedReq.Header.Set("User-Agent", t.userAgent)
		return t.base.RoundTrip(clonedReq)
	}
	return t.base.RoundTrip(req)
}

// cloneDefaultTransport 基于 http.DefaultTransport 克隆，
// 保留原生连接池优化（MaxIdleConns=100, TLS 配置等）。
func cloneDefaultTransport() *http.Transport {
	return http.DefaultTransport.(*http.Transport).Clone()
}

// NewClient 创建统一的 HTTP 客户端（自动注入 User-Agent）。
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &userAgentTransport{
			base:      cloneDefaultTransport(),
			userAgent: DefaultUserAgent,
		},
	}
}

// NewClientWithProxy 创建带代理支持的 HTTP 客户端（自动注入 User-Agent）。
func NewClientWithProxy(timeout time.Duration) *http.Client {
	transport := cloneDefaultTransport()
	transport.Proxy = http.ProxyFromEnvironment
	return &http.Client{
		Timeout: timeout,
		Transport: &userAgentTransport{
			base:      transport,
			userAgent: DefaultUserAgent,
		},
	}
}
