package models

import (
	"time"
	"github.com/google/uuid"
)

type Task struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	ScriptPath      string    `json:"script_path" gorm:"not null"`
	CondaEnv        string    `json:"conda_env" gorm:"not null"`
	CronExpr        string    `json:"cron_expr" gorm:"not null"`
	Enabled         bool      `json:"enabled" gorm:"default:true"`
	NotifyOnFailure bool      `json:"notify_on_failure" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (t *Task) BeforeCreate() error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}
