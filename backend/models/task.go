package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CronExprList 用于存储多个 Cron 表达式的 JSON 数组
type CronExprList []string

func (l CronExprList) MarshalJSON() ([]byte, error) {
	if l == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(l))
}

func (l CronExprList) Value() (driver.Value, error) {
	data, err := json.Marshal([]string(l))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (l *CronExprList) Scan(value any) error {
	if value == nil {
		*l = CronExprList{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("CronExprList.Scan: unsupported type %T", value)
	}
	if len(raw) == 0 {
		*l = CronExprList{}
		return nil
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err != nil {
		return err
	}
	*l = CronExprList(items)
	return nil
}

type Task struct {
	ID              string       `json:"id" gorm:"primaryKey"`
	Name            string       `json:"name" gorm:"not null"`
	ScriptPath      string       `json:"script_path" gorm:"not null"`
	CondaEnv        string       `json:"conda_env" gorm:"not null"`
	CronExpr        string       `json:"cron_expr" gorm:"not null"`       // 兼容字段：第一条 cron 表达式
	CronExprs       CronExprList `json:"cron_exprs" gorm:"type:TEXT"`     // 多时间点：JSON 数组
	Enabled         bool         `json:"enabled" gorm:"default:true"`
	NotifyOnFailure bool         `json:"notify_on_failure" gorm:"default:true"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// NormalizeCron 归一化 cron 表达式，确保 CronExprs 和 CronExpr 一致
func (t *Task) NormalizeCron() {
	// 如果 CronExprs 为空，从 CronExpr 补齐
	if len(t.CronExprs) == 0 {
		expr := strings.TrimSpace(t.CronExpr)
		if expr != "" {
			t.CronExprs = CronExprList{expr}
		}
	}
	// 去除空白
	for i := range t.CronExprs {
		t.CronExprs[i] = strings.TrimSpace(t.CronExprs[i])
	}
	// 确保 CronExpr 等于第一条
	if len(t.CronExprs) > 0 {
		t.CronExpr = t.CronExprs[0]
	}
}

func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}
