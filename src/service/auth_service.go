package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"v2rayn-go/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务（登录、JWT、TOTP、密码管理）
type AuthService struct{}

// NewAuthService 创建 AuthService 实例
func NewAuthService() *AuthService {
	return &AuthService{}
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
	}

	return &user, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// JWT 签发与验证
// ──────────────────────────────────────────────────────────────────────────────

// GenerateToken 为指定用户签发 JWT Token
// 使用用户专属的 JWTSecret（HS256），过期时间从 app_settings 读取
func (s *AuthService) GenerateToken(user *database.User) (string, error) {
	// 从 app_settings 读取 JWT 过期时间（小时）
	expireHours := 24 // 默认 24 小时
	if val := getSettingValue("jwt_expire_hours"); val != "" {
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
func (s *AuthService) ChangePassword(userUUID, oldPwd, newPwd string) error {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// 新密码长度校验
	if len(newPwd) < 6 {
		return fmt.Errorf("new password must be at least 6 characters")
	}

	// 生成新密码哈希
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 生成新的 JWTSecret（使旧 Token 全部失效）
	newSecret, err := generateRandomHex(32)
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// 原子更新：密码哈希 + JWTSecret
	return database.DB.Model(&user).Updates(map[string]any{
		"password_hash": string(hashedPwd),
		"jwt_secret":    newSecret,
	}).Error
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

// EnableTOTP 为用户生成 TOTP 密钥，暂存到数据库（但 TOTPEnabled 仍为 false）
// 返回 secret 字符串和 otpauth:// URI，前端自行渲染二维码
// 调用方需通过 VerifyAndActivateTOTP 验证后才真正启用
func (s *AuthService) EnableTOTP(userUUID string) (secret, otpauthURL string, err error) {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	if user.TOTPEnabled {
		return "", "", fmt.Errorf("totp already enabled")
	}

	// 使用 pquerna/otp 生成 TOTP 密钥
	// 配置：发行者 = "v2rayN-Go"，账户名 = 用户名
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

// VerifyAndActivateTOTP 验证用户输入的 TOTP 动态码，正确后正式启用 TOTP
// 使用默认时间窗口配置（前后各 1 个周期 = ±30 秒），容忍手机时间漂移
func (s *AuthService) VerifyAndActivateTOTP(userUUID, code string) error {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
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
// 需要当前密码确认（安全起见）
func (s *AuthService) DisableTOTP(userUUID, password string) error {
	var user database.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}

	// 清空 TOTP 密钥 + 关闭开关
	return database.DB.Model(&user).Updates(map[string]any{
		"totp_secret":  "",
		"totp_enabled": false,
	}).Error
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

// getSettingValue 从 app_settings 表读取指定 key 的值（不存在则返回空字符串）
func getSettingValue(key string) string {
	var setting database.AppSetting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}
