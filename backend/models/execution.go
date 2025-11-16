package models

import (
	"time"
	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	StatusRunning ExecutionStatus = "running"
	StatusSuccess ExecutionStatus = "success"
	StatusFailed  ExecutionStatus = "failed"
)

type Execution struct {
	ID           string          `json:"id" gorm:"primaryKey"`
	TaskID       string          `json:"task_id" gorm:"not null"`
	Status       ExecutionStatus `json:"status" gorm:"not null"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      *time.Time      `json:"end_time"`
	DurationMs   int64           `json:"duration_ms"`
	ExitCode     int             `json:"exit_code"`
	ErrorMessage string          `json:"error_message"`
}

func (e *Execution) BeforeCreate() error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	e.StartTime = time.Now()
	e.Status = StatusRunning
	return nil
}
