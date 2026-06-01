package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"v2rayn-go/config"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// purgeDeleted 启动时物理删除所有软删除残留记录，防止历史数据堆积
func purgeDeleted() {
	models := []any{
		&User{},
		&Profile{},
		&NodeGroup{},
		&RoutingRule{},
		&AppSetting{},
	}
	for _, m := range models {
		if err := DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(m).Error; err != nil {
			log.Printf("[WARN] purge deleted records for %T: %v", m, err)
		}
	}
}

var DB *gorm.DB

// Init 初始化 SQLite 数据库连接并自动迁移表结构
func Init(cfg *config.AppConfig) error {
	// 连接参数说明：
	// - busy_timeout(5000): 遇到锁冲突时排队等待 5 秒，而非立刻报 "database is locked"
	// - journal_mode(WAL): 启用 Write-Ahead Logging 模式
	//   WAL 模式下断电最多丢失最近一个事务，数据库结构不会损坏；
	//   同时读写并发性能远优于默认的 DELETE journal 模式
	// - synchronous(NORMAL): WAL 模式下的推荐级别，在安全性和性能之间取得最佳平衡
	//   FULL 级别在 WAL 下已无必要，NORMAL 足以保证数据库文件不损坏
	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.DBPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	// 自动迁移表结构
	if err := DB.AutoMigrate(
		&User{},
		&Profile{},
		&NodeGroup{},
		&RoutingRule{},
		&AppSetting{},
	); err != nil {
		return err
	}

	// 清理历史残留的软删除数据
	purgeDeleted()
	log.Println("[INFO] Purged soft-deleted records")

	// 启动时全表重排排序值
	RebalanceAll()
	log.Println("[INFO] Startup sort_order rebalance completed")

	// 如果分组为空，创建默认分组
	var count int64
	DB.Model(&NodeGroup{}).Count(&count)
	if count == 0 {
		DB.Create(&NodeGroup{
			UUID:      GenerateUUID(),
			Alias:     "",
			SortOrder: 10,
			Enabled:   true,
		})
		log.Println("Created default group")
	}

	return nil
}

// SeedDefaults 首次启动时注入默认数据：app_settings 默认配置 + 初始管理员账号
// 仅在对应表为空时执行，幂等安全，可重复调用
func SeedDefaults() error {
	// ── 1. app_settings 默认键值对 ──────────────────────────────────
	var settingCount int64
	DB.Model(&AppSetting{}).Count(&settingCount)
	if settingCount == 0 {
		defaults := []AppSetting{
			{Key: "force_https", Value: "false"},   // 是否强制 HTTPS
			{Key: "custom_base_path", Value: "/"},  // 自定义路由前缀
			{Key: "jwt_expire_hours", Value: "24"}, // JWT 过期时间（小时）
		}
		if err := DB.Create(&defaults).Error; err != nil {
			return fmt.Errorf("seed app_settings: %w", err)
		}
		log.Println("[INFO] Seeded default app_settings")
	}

	// ── 2. 初始管理员账号（仅 users 表为空时） ─────────────────────
	var userCount int64
	DB.Model(&User{}).Count(&userCount)
	if userCount > 0 {
		return nil
	}

	// 生成 16 字节随机密码 → Base64 编码为可打印明文
	pwdBytes := make([]byte, 16)
	if _, err := rand.Read(pwdBytes); err != nil {
		return fmt.Errorf("generate random password: %w", err)
	}
	// 使用 hex 编码（32 字符，纯字母数字，无特殊字符，方便终端复制）
	plainPassword := hex.EncodeToString(pwdBytes)

	// bcrypt 哈希（cost=10，平衡安全与性能）
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt hash password: %w", err)
	}

	// 生成 32 字节 JWT 签名密钥
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return fmt.Errorf("generate JWT secret: %w", err)
	}
	jwtSecret := hex.EncodeToString(secretBytes)

	// 创建 admin 用户（Role=1 代表超管，全局唯一）
	admin := User{
		UUID:         GenerateUUID(),
		Username:     "admin",
		PasswordHash: string(hashedPwd),
		JWTSecret:    jwtSecret,
		Role:         1,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	// 高亮打印初始密码（ANSI 粗体黄色，仅首次启动显示）
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "\033[1;33m══════════════════════════════════════════════\033[0m")
	fmt.Fprintf(os.Stderr, "\033[1;33m║  Initial admin password: %-20s║\033[0m\n", plainPassword)
	fmt.Fprintln(os.Stderr, "\033[1;33m║  Please change it after first login!         ║\033[0m")
	fmt.Fprintln(os.Stderr, "\033[1;33m══════════════════════════════════════════════\033[0m")
	fmt.Fprintln(os.Stderr, "")

	return nil
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
