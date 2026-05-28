package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitTestDB 初始化内存 SQLite 数据库用于测试
func InitTestDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to init test db: " + err.Error())
	}

	if err := DB.AutoMigrate(
		&Profile{},
		&NodeGroup{},
		&RoutingRule{},
		&StrategyGroup{},
		&AppSetting{},
	); err != nil {
		panic("failed to migrate test db: " + err.Error())
	}
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
		DB = nil
	}
}
