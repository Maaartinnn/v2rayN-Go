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
	models := []interface{}{
		&Profile{},
		&NodeGroup{},
		&RoutingRule{},
		&StrategyGroup{},
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
	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
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
		&StrategyGroup{},
		&AppSetting{},
	); err != nil {
		return err
	}

	// 清理历史残留的软删除数据
	purgeDeleted()

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
