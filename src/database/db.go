package database

import (
	"log"
	"v2rayn-go/config"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// purgeDeleted 启动时物理删除所有软删除残留记录，防止历史数据堆积
func purgeDeleted() {
	models := []any{
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

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
