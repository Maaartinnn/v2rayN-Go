package web

import (
	"log/slog"
	"net/http"

	"v2rayn-go/service"
)

// AuthHandler 处理认证相关 API 请求
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register 注册认证相关路由到 ServeMux
func (h *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", h.handleLogin)
	mux.HandleFunc("POST /api/change-password", h.handleChangePassword)
	mux.HandleFunc("POST /api/totp/enable", h.handleEnableTOTP)
	mux.HandleFunc("POST /api/totp/check", h.handleCheckTOTP)
	mux.HandleFunc("POST /api/totp/verify", h.handleVerifyTOTP)
	mux.HandleFunc("POST /api/totp/disable", h.handleDisableTOTP)
	mux.HandleFunc("POST /api/sessions/revoke-all", h.handleRevokeAll)
	mux.HandleFunc("GET /api/auth/me", h.handleMe)
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/login
// 请求体: { "username": "admin", "password": "xxx", "totp_code": "123456" }
// 成功响应: { "token": "eyJ...", "user": { "uuid", "username", "role", "totp_enabled" } }
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// 验证凭据
	user, err := h.authSvc.Login(req.Username, req.Password, req.TOTPCode)
	if err != nil {
		slog.Warn("login failed", "username", req.Username, "error", err)
		jsonError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 签发 JWT
	token, err := h.authSvc.GenerateToken(user)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	slog.Info("user logged in", "username", user.Username, "uuid", user.UUID)

	jsonOK(w, map[string]any{
		"token": token,
		"user": map[string]any{
			"uuid":         user.UUID,
			"username":     user.Username,
			"role":         user.Role,
			"totp_enabled": user.TOTPEnabled,
		},
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/auth/me
// 返回当前登录用户的基本信息（由中间件注入 user context）
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	jsonOK(w, map[string]any{
		"uuid":         user.UUID,
		"username":     user.Username,
		"role":         user.Role,
		"totp_enabled": user.TOTPEnabled,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/change-password
// 请求体: { "old_password": "xxx", "new_password": "yyy" }
// 成功后自动 RotateJWTSecret，其他设备 Token 失效
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	newToken, err := h.authSvc.ChangePassword(user.UUID, req.OldPassword, req.NewPassword)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("password changed", "username", user.Username, "uuid", user.UUID)
	jsonOK(w, map[string]string{
		"status": "password_changed",
		"token":  newToken,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/totp/enable
// 无请求体，生成随机 TOTP 密钥并返回 secret + otpauth URL（前端渲染二维码）
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleEnableTOTP(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	secret, otpauthURL, err := h.authSvc.EnableTOTP(user.UUID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]string{
		"secret":      secret,
		"otpauth_url": otpauthURL,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/totp/check
// 请求体: { "secret": "JBSWY3DPEHPK3PXP" }
// 校验自定义 TOTP 密钥格式（不写数据库，仅校验 + 清洗）
// 响应: { "valid": true, "secret": "CLEANED" } 或 { "valid": false, "secret": "original" }
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleCheckTOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"secret"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	valid, cleaned := h.authSvc.CheckTOTPSecret(req.Secret)
	jsonOK(w, map[string]any{
		"valid":  valid,
		"secret": cleaned,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/totp/verify
// 请求体: { "code": "123456", "secret": "JBSWY3D..." }（secret 可选）
// 验证通过后正式启用 TOTP
//   - secret 非空 → 使用自定义密钥（后端校验 + 写入 DB），再验证动态码
//   - secret 为空 → 使用 DB 中已有的密钥（由 enable 生成的随机密钥）验证
//
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleVerifyTOTP(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code   string `json:"code"`
		Secret string `json:"secret"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.authSvc.VerifyAndActivateTOTP(user.UUID, req.Code, req.Secret); err != nil {
		mapServiceError(w, err)
		return
	}

	slog.Info("TOTP enabled", "username", user.Username, "uuid", user.UUID)
	jsonOK(w, map[string]string{"status": "totp_enabled"})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/totp/disable
// 请求体: { "totp_code": "123456" }
// 需要 TOTP 验证码确认，关闭后清空 TOTP 密钥
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleDisableTOTP(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	var req struct {
		TOTPCode string `json:"totp_code"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.authSvc.DisableTOTP(user.UUID, req.TOTPCode); err != nil {
		mapServiceError(w, err)
		return
	}

	slog.Info("TOTP disabled", "username", user.Username, "uuid", user.UUID)
	jsonOK(w, map[string]string{"status": "totp_disabled"})
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/sessions/revoke-all
// 刷新 JWTSecret 使所有设备的旧 Token 失效
// ──────────────────────────────────────────────────────────────────────────────
func (h *AuthHandler) handleRevokeAll(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.authSvc.RotateJWTSecret(user.UUID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("all sessions revoked", "username", user.Username, "uuid", user.UUID)
	jsonOK(w, map[string]string{"status": "all_sessions_revoked"})
}

// ──────────────────────────────────────────────────────────────────────────────
// getUserFromContext 从请求 context 中提取由 Auth 中间件注入的 User 对象
// ──────────────────────────────────────────────────────────────────────────────
func getUserFromContext(r *http.Request) *userContext {
	v := r.Context().Value(contextKeyUser)
	if v == nil {
		return nil
	}
	u, ok := v.(*userContext)
	if !ok {
		return nil
	}
	return u
}
