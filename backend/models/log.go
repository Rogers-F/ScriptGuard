package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelStdout  LogLevel = "stdout"
	LogLevelStderr  LogLevel = "stderr"
	LogLevelSuccess LogLevel = "success"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type Log struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ExecutionID string    `json:"execution_id" gorm:"not null;index"`
	TaskID      string    `json:"task_id" gorm:"not null;index"`
	Timestamp   time.Time `json:"timestamp" gorm:"index"`
	Level       LogLevel  `json:"level" gorm:"not null"`
	Content     string    `json:"content" gorm:"type:text"`
}

func (l *Log) BeforeCreate(_ *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	return nil
}
