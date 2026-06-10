package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"v2rayn-go/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// base32Pattern Base32 格式校验（RFC 4648，大写字母 A-Z + 数字 2-7 + 可选的 = 填充）
var base32Pattern = regexp.MustCompile(`^[A-Z2-7]+=*$`)

// AuthService 认证服务（登录、JWT、TOTP、密码管理）
type AuthService struct {
	settingsSvc *SettingsService
}

// NewAuthService 创建 AuthService 实例，注入 SettingsService 以复用内存缓存
func NewAuthService(settingsSvc *SettingsService) *AuthService {
	return &AuthService{settingsSvc: settingsSvc}
}

// ──────────────────────────────────────────────────────────────────────────────
// JWT Claims
// ──────────────────────────────────────────────────────────────────────────────

// AuthClaims 自定义 JWT Claims，使用 UUID 作为用户标识
type AuthClaims struct {
	UserUUID string `json:"user_uuid"` // 用户 UUID（全局唯一标识）
	Username string `json:"username"`  // 用户名
	Role     int    `json:"role"`      // 角色: 0=普通用户, 1=超管
	jwt.RegisteredClaims
}

// ──────────────────────────────────────────────────────────────────────────────
// 登录验证
// ──────────────────────────────────────────────────────────────────────────────

// Login 验证用户名 + 密码 + 可选 TOTP 动态码
// 返回通过验证的 User 对象
func (s *AuthService) Login(username, password, totpCode string) (*database.User, error) {
	// 1. 按用户名查找用户
	var user database.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 2. 验证密码（bcrypt 常量时间比较，防时序攻击）
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 3. 如果启用了 TOTP 两步验证，校验动态码
	if user.TOTPEnabled {
		if totpCode == "" {
			return nil, fmt.Errorf("totp code required")
		}
		// 使用默认配置验证 TOTP（允许前后各 1 个时间窗口偏差，即 ±30 秒）
		// 足以应对手机时间与服务器时间的轻微漂移
		if !totp.Validate(totpCode, user.TOTPSecret) {
			return nil, fmt.Errorf("invalid totp code")
		}
	} else if totpCode != "" {
		// 没开 TOTP 但输入了动态码 → 拒绝登录（防止攻击者试探）
		return nil, fmt.Errorf("two-factor authentication is not enabled, but code was provided")
	}

	return &user, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// JWT 签发与验证
// ──────────────────────────────────────────────────────────────────────────────

// GenerateToken 为指定用户签发 JWT Token
// 使用用户专属的 JWTSecret（HS256），过期时间从 app_settings 读取
func (s *AuthService) GenerateToken(user *database.User) (string, error) {
	// 从 SettingsService 缓存读取 JWT 过期时间（小时）
	// 使用内存缓存（DCL），零 DB I/O，修改后无需重启即可对新 JWT 生效
	expireHours := 24 // 默认 24 小时
	if val := s.settingsSvc.GetSettingFast("jwt_expire_hours"); val != "" {
		if h, err := strconv.Atoi(val); err == nil && h > 0 {
			expireHours = h
		}
	}

	now := time.Now()
	claims := AuthClaims{
		UserUUID: user.UUID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.UUID,
		},
	}

	// 使用用户专属的 JWTSecret 签发（HS256）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(user.JWTSecret))
}

// ValidateToken 验证 JWT Token 并返回对应的 User
// 流程：解析 JWT（不验签）→ 提取 UUID → 从 DB 拿 JWTSecret → 重新验签
// 这样 RotateJWTSecret 后旧 Token 立即失效
func (s *AuthService) ValidateToken(tokenStr string) (*database.User, error) {
	// 1. 解析 JWT（不验签），仅提取 claims 中的 UUID
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	var claims AuthClaims
	_, _, err := parser.ParseUnverified(tokenStr, &claims)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	// 2. 从数据库查找用户，获取当前的 JWTSecret
	var user database.User
	if err := database.DB.Where("uuid = ?", claims.UserUUID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. 用当前 secret 重新验签（确保 token 是用最新 secret 签发的）
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(user.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &user, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// 密码管理
// ──────────────────────────────────────────────────────────────────────────────

// ChangePassword 修改用户密码（需旧密码验证）
// 成功后自动 RotateJWTSecret，使其他所有设备的 Token 失效
// 返回新签发的 JWT Token，让当前设备无缝续用
func (s *AuthService) ChangePassword(userUUID, oldPwd, newPwd string) (string, error) {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return "", fmt.Errorf("user not found")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return "", fmt.Errorf("invalid old password")
	}

	// 新密码长度校验
	if len(newPwd) < 6 {
		return "", fmt.Errorf("new password must be at least 6 characters")
	}

	// 新旧密码不能相同
	if oldPwd == newPwd {
		return "", fmt.Errorf("new password cannot be the same as current password")
	}

	// 生成新密码哈希
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash new password: %w", err)
	}

	// 生成新的 JWTSecret（使旧 Token 全部失效）
	newSecret, err := generateRandomHex(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// 原子更新：密码哈希 + JWTSecret
	if err := database.DB.Model(&user).Updates(map[string]any{
		"password_hash": string(hashedPwd),
		"jwt_secret":    newSecret,
	}).Error; err != nil {
		return "", err
	}

	// 重新加载用户以获取更新后的 JWTSecret
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return "", fmt.Errorf("failed to reload user: %w", err)
	}

	// 为当前设备签发新 Token（使其无缝续用）
	token, err := s.GenerateToken(&user)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}
	return token, nil
}

// RotateJWTSecret 重新生成用户的 JWTSecret，使所有旧 Token 失效
// 用于：改密后、手动注销所有设备、开启/关闭 TOTP
func (s *AuthService) RotateJWTSecret(userUUID string) error {
	newSecret, err := generateRandomHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	return database.DB.Model(&database.User{}).
		Where("uuid = ?", userUUID).
		Update("jwt_secret", newSecret).Error
}

// ──────────────────────────────────────────────────────────────────────────────
// TOTP 两步验证
// ──────────────────────────────────────────────────────────────────────────────

// EnableTOTP 为用户生成随机 TOTP 密钥，暂存到数据库（但 TOTPEnabled 仍为 false）。
// 返回 secret 字符串和 otpauth:// URI（前端渲染二维码用）。
// 调用方需通过 VerifyAndActivateTOTP 验证后才真正启用。
func (s *AuthService) EnableTOTP(userUUID string) (secret, otpauthURL string, err error) {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	if user.TOTPEnabled {
		return "", "", fmt.Errorf("totp already enabled")
	}

	// 使用 pquerna/otp 生成随机 TOTP 密钥
	// Issuer 固定为 "v2rayN-Go"，前端可自行拼接自定义的 otpauth URL
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "v2rayN-Go",
		AccountName: user.Username,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// 暂存密钥到数据库（但 TOTPEnabled 仍为 false，等验证通过后才启用）
	if err := database.DB.Model(&user).Update("totp_secret", key.Secret()).Error; err != nil {
		return "", "", fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// CheckTOTPSecret 校验自定义 TOTP 密钥是否合法（不写数据库，仅校验）。
// 返回 (valid bool, cleanedSecret string)：
//   - valid=true  → 密钥格式合法，cleanedSecret 为清洗后的大写无空格值
//   - valid=false → 密钥格式不合法，cleanedSecret 为原始输入值（供前端回滚用）
//
// 校验规则：
//  1. 数据清洗：大写化 + 去除所有空格
//  2. 仅允许 Base32 字符（A-Z, 2-7, = 填充符）
//  3. 长度 16-64 字符（排除填充符后）
func (s *AuthService) CheckTOTPSecret(secret string) (bool, string) {
	// 数据清洗：大写化 + 去除所有空格
	cleaned := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	// 空值 → 不合法
	if cleaned == "" {
		return false, secret
	}

	// Base32 格式校验（RFC 4648）
	if !base32Pattern.MatchString(cleaned) {
		return false, secret
	}

	// 长度校验（去掉填充符 = 后）
	coreLen := len(strings.TrimRight(cleaned, "="))
	if coreLen < 16 || coreLen > 64 {
		return false, secret
	}

	return true, cleaned
}

// VerifyAndActivateTOTP 验证用户输入的 TOTP 动态码，正确后正式启用 TOTP。
//
// secret 参数为可选的自定义密钥（用户在前端"自定义密钥"输入框中填写的值）：
//   - 非空 → 先用 CheckTOTPSecret 校验格式，通过后写入 DB 替换随机密钥，再验证动态码
//   - 空字符串 → 使用 DB 中已有的密钥（由 EnableTOTP 生成的随机密钥）验证
//
// 使用默认时间窗口配置（前后各 1 个周期 = ±30 秒），容忍手机时间漂移。
func (s *AuthService) VerifyAndActivateTOTP(userUUID, code, secret string) error {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// 如果用户提供了自定义密钥，先校验格式并写入 DB
	if secret != "" {
		valid, cleaned := s.CheckTOTPSecret(secret)
		if !valid {
			return NewValidation("invalid secret: must be Base32 format, 16-128 chars", nil)
		}
		// 将清洗后的自定义密钥写入 DB（替换之前 EnableTOTP 生成的随机密钥）
		if err := database.DB.Model(&user).Update("totp_secret", cleaned).Error; err != nil {
			return fmt.Errorf("failed to save custom secret: %w", err)
		}
		// 重新加载用户以获取最新的 TOTPSecret
		if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
			return fmt.Errorf("failed to reload user: %w", err)
		}
	}

	if user.TOTPSecret == "" {
		return fmt.Errorf("totp not initialized, call enable first")
	}

	if user.TOTPEnabled {
		return fmt.Errorf("totp already enabled")
	}

	// 验证动态码（默认允许 ±30 秒偏差）
	if !totp.Validate(code, user.TOTPSecret) {
		return fmt.Errorf("invalid totp code")
	}

	// 验证通过，正式启用 TOTP
	return database.DB.Model(&user).Update("totp_enabled", true).Error
}

// DisableTOTP 关闭用户的 TOTP 两步验证
// 需要当前 TOTP 验证码确认（比密码更安全，防止密码泄露后被绕过）
func (s *AuthService) DisableTOTP(userUUID, totpCode string) error {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// 校验 TOTP 验证码格式（6 位纯数字）
	if len(totpCode) != 6 || !isDigitsOnly(totpCode) {
		return NewValidation("invalid totp code: must be exactly 6 digits", nil)
	}

	// 验证 TOTP 验证码（必须在允许的时间窗口内）
	if !totp.Validate(totpCode, user.TOTPSecret) {
		return fmt.Errorf("invalid totp code")
	}

	// 清空 TOTP 密钥 + 关闭开关
	return database.DB.Model(&user).Updates(map[string]any{
		"totp_secret":  "",
		"totp_enabled": false,
	}).Error
}

// isDigitsOnly 检查字符串是否全部由数字组成
func isDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ──────────────────────────────────────────────────────────────────────────────
// 用户信息查询
// ──────────────────────────────────────────────────────────────────────────────

// GetUserByUUID 按 UUID 查询用户
func (s *AuthService) GetUserByUUID(uuid string) (*database.User, error) {
	var user database.User
	if err := database.DB.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────────────────────────────────────

// generateRandomHex 生成 n 字节的随机数据并返回 hex 编码字符串
func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
