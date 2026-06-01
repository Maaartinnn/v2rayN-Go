package web

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"v2rayn-go/service"
)

// contextKey 自定义 context key 类型，避免与其他包冲突
type contextKey string

// contextKeyUser 是存储在请求 context 中的用户信息的 key
const contextKeyUser contextKey = "auth_user"

// userContext 存储在请求 context 中的用户精简信息
// 仅包含下游 handler 需要的字段，避免暴露敏感数据（如密码哈希、JWTSecret）
type userContext struct {
	UUID        string // 用户 UUID
	Username    string // 用户名
	Role        int    // 角色
	TOTPEnabled bool   // 是否启用 TOTP
}

// authWhiteList 不需要 JWT 验证的 API 路径前缀
// 白名单中的路径将直接放行，由 handler 自行处理身份验证（如 /api/login）
var authWhiteList = []string{
	"/api/login",
}

// isAuthWhitelisted 检查请求路径是否在白名单中
func isAuthWhitelisted(path string) bool {
	for _, wl := range authWhiteList {
		if strings.HasPrefix(path, wl) {
			return true
		}
	}
	return false
}

// AuthMiddleware 创建 JWT 认证中间件
// 工作流程：
//  1. 非 /api/ 开头的请求 → 直接放行（静态资源）
//  2. 白名单路径 → 直接放行（如 /api/login）
//  3. 提取 Authorization: Bearer <token>
//  4. authSvc.ValidateToken 验证 Token
//  5. 成功 → 将精简用户信息存入 context，调用 next handler
//  6. 失败 → 返回 401 Unauthorized
func AuthMiddleware(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. 非 API 路径直接放行（静态资源、WebSocket 等）
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// 2. 白名单路径直接放行
			if isAuthWhitelisted(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 3. 提取 Bearer Token
			token := extractBearerToken(r)
			if token == "" {
				jsonError(w, "authentication required", http.StatusUnauthorized)
				return
			}

			// 4. 验证 Token（解析 → 查 DB 拿 secret → 验签）
			user, err := authSvc.ValidateToken(token)
			if err != nil {
				slog.Warn("auth failed", "path", r.URL.Path, "error", err)
				jsonError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// 5. 构造精简用户信息存入 context
			uc := &userContext{
				UUID:        user.UUID,
				Username:    user.Username,
				Role:        user.Role,
				TOTPEnabled: user.TOTPEnabled,
			}
			ctx := context.WithValue(r.Context(), contextKeyUser, uc)

			// 6. 调用下游 handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken 从请求头提取 Bearer Token
// 格式: "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	// 必须以 "Bearer " 开头（注意空格）
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
}
