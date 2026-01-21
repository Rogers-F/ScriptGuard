package database

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"scriptguard/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// GetDefaultDBPath 获取默认数据库路径（用户数据目录）
func GetDefaultDBPath() (string, error) {
	// 使用用户配置目录（Windows: %AppData%, Linux: ~/.config, macOS: ~/Library/Application Support）
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "ScriptGuard", "database", "scriptguard.db"), nil
}

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// SG-006: SQLite 并发优化
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	// 单连接避免 SQLite 并发写入冲突
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// 设置 busy_timeout 避免 "database is locked"
	if err := DB.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		return err
	}
	// 启用 WAL 模式提升并发读写性能
	if err := DB.Exec("PRAGMA journal_mode = WAL;").Error; err != nil {
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
	if err := initDefaultConfig(); err != nil {
		log.Printf("初始化默认配置失败: %v", err)
	}
	return nil
}

// initDefaultConfig 初始化默认配置
func initDefaultConfig() error {
	defaults := map[string]string{
		models.ConfigKeyDingTalkWebhook:         "",
		models.ConfigKeyWeComWebhook:            "",
		models.ConfigKeyLogRetentionDays:        "30",
		models.ConfigKeyMaxConcurrency:          "5",
		models.ConfigKeyExecutionTimeoutSeconds: "3600",
		models.ConfigKeyAutoStartEnabled:        "false",
	}

	for key, value := range defaults {
		var config models.Config
		err := DB.Where("key = ?", key).First(&config).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 仅"确实不存在"才创建
				if createErr := DB.Create(&models.Config{
					Key:   key,
					Value: value,
				}).Error; createErr != nil {
					return createErr
				}
			} else {
				// 其他 DB 错误，返回而非静默忽略
				return err
			}
		}
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
