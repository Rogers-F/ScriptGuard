package services

import (
	"log"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"strconv"

	"github.com/robfig/cron/v3"
)

// CleanupService 清理服务
type CleanupService struct {
	cron *cron.Cron
}

// NewCleanupService 创建清理服务
func NewCleanupService() *CleanupService {
	// SG-007/SG-023: 使用北京时间（复用公共时区定义）
	return &CleanupService{
		cron: cron.New(cron.WithSeconds(), cron.WithLocation(BeijingLocation)),
	}
}

// Start 启动清理服务
func (s *CleanupService) Start() {
	// SG-012: 检查 AddFunc 错误
	_, err := s.cron.AddFunc("0 0 2 * * *", func() {
		s.cleanupOldLogs()
	})
	if err != nil {
		log.Printf("添加清理任务失败: %v", err)
		return
	}
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
		// SG-012: 记录错误而非静默返回
		log.Printf("读取日志保留天数配置失败: %v，使用默认值 30 天", err)
		config.Value = "30"
	}

	retentionDays, err := strconv.Atoi(config.Value)
	if err != nil {
		retentionDays = 30 // 默认30天
	}

	// 计算截止日期（SG-023: 使用北京时间）
	cutoffDate := NowBeijing().AddDate(0, 0, -retentionDays)

	// 删除旧日志
	result := db.Where("timestamp < ?", cutoffDate).Delete(&models.Log{})
	if result.Error != nil {
		log.Printf("清理旧日志失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("已清理 %d 条过期日志", result.RowsAffected)
	}

	// 删除旧的执行记录（与日志使用同一保留天数；仅删除已结束的记录）
	execResult := db.Where("end_time IS NOT NULL AND end_time < ?", cutoffDate).Delete(&models.Execution{})
	if execResult.Error != nil {
		log.Printf("清理旧执行记录失败: %v", execResult.Error)
	} else if execResult.RowsAffected > 0 {
		log.Printf("已清理 %d 条过期执行记录", execResult.RowsAffected)
	}
}

// CleanupDatabase 手动触发数据库清理
func (s *CleanupService) CleanupDatabase() error {
	db := database.GetDB()

	// VACUUM SQLite数据库（压缩和优化）
	return db.Exec("VACUUM").Error
}
