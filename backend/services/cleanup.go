package services

import (
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// CleanupService 清理服务
type CleanupService struct {
	cron *cron.Cron
}

// NewCleanupService 创建清理服务
func NewCleanupService() *CleanupService {
	return &CleanupService{
		cron: cron.New(cron.WithSeconds()),
	}
}

// Start 启动清理服务
func (s *CleanupService) Start() {
	// 每天凌晨2点执行清理任务
	s.cron.AddFunc("0 0 2 * * *", func() {
		s.cleanupOldLogs()
	})
	s.cron.Start()
}

// Stop 停止清理服务
func (s *CleanupService) Stop() {
	s.cron.Stop()
}

// cleanupOldLogs 清理旧日志
func (s *CleanupService) cleanupOldLogs() {
	// 获取日志保留天数配置
	var config models.Config
	db := database.GetDB()

	err := db.Where("key = ?", models.ConfigKeyLogRetentionDays).First(&config).Error
	if err != nil {
		return
	}

	retentionDays, err := strconv.Atoi(config.Value)
	if err != nil {
		retentionDays = 30 // 默认30天
	}

	// 计算截止日期
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// 删除旧日志
	result := db.Where("timestamp < ?", cutoffDate).Delete(&models.Log{})
	if result.Error == nil && result.RowsAffected > 0 {
		// 记录清理日志（可选）
		// log.Printf("Cleaned up %d old log entries", result.RowsAffected)
	}

	// 删除旧的执行记录（保留最近的记录，只删除日志）
	// 这里可以根据需求调整策略
}

// CleanupDatabase 手动触发数据库清理
func (s *CleanupService) CleanupDatabase() error {
	db := database.GetDB()

	// VACUUM SQLite数据库（压缩和优化）
	return db.Exec("VACUUM").Error
}
