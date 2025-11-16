package database

import (
	"scriptguard/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&models.Task{},
		&models.Execution{},
		&models.Log{},
		&models.Config{},
	)
	if err != nil {
		return err
	}

	// 初始化默认配置
	initDefaultConfig()
	return nil
}

// initDefaultConfig 初始化默认配置
func initDefaultConfig() {
	defaults := map[string]string{
		models.ConfigKeyDingTalkWebhook:  "",
		models.ConfigKeyWeComWebhook:     "",
		models.ConfigKeyLogRetentionDays: "30",
		models.ConfigKeyMaxConcurrency:   "5",
	}

	for key, value := range defaults {
		var config models.Config
		if err := DB.Where("key = ?", key).First(&config).Error; err != nil {
			// 配置不存在，创建默认配置
			DB.Create(&models.Config{
				Key:   key,
				Value: value,
			})
		}
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
