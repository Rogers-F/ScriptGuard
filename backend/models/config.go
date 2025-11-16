package models

import (
	"time"
)

type Config struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 配置键常量
const (
	ConfigKeyDingTalkWebhook  = "dingtalk_webhook"
	ConfigKeyWeComWebhook     = "wecom_webhook"
	ConfigKeyLogRetentionDays = "log_retention_days"
	ConfigKeyMaxConcurrency   = "max_concurrency"
)
