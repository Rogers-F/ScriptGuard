package services

import (
	"log"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"sync"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron     *cron.Cron
	tasks    map[string][]cron.EntryID // 支持一个任务多个时间点
	executor *ExecutorService
	notifier *NotifierService
	mu       sync.RWMutex
}

func NewSchedulerService(executor *ExecutorService, notifier *NotifierService) *SchedulerService {
	// SG-007/SG-023: 使用北京时间（复用公共时区定义）
	return &SchedulerService{
		cron:     cron.New(cron.WithSeconds(), cron.WithLocation(BeijingLocation)),
		tasks:    make(map[string][]cron.EntryID),
		executor: executor,
		notifier: notifier,
	}
}

// Start 启动调度器
func (s *SchedulerService) Start() {
	s.cron.Start()
}

// Stop 停止调度器
func (s *SchedulerService) Stop() {
	s.cron.Stop()
}

// AddTask 添加任务（支持多个时间点）
func (s *SchedulerService) AddTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !task.Enabled {
		return nil
	}

	// 归一化 cron 表达式（兼容旧数据）
	task.NormalizeCron()

	entryIDs := make([]cron.EntryID, 0, len(task.CronExprs))
	for _, expr := range task.CronExprs {
		entryID, err := s.cron.AddFunc(expr, func() {
			s.executeTask(task)
		})
		if err != nil {
			// 回滚已添加的 entry
			for _, id := range entryIDs {
				s.cron.Remove(id)
			}
			return err
		}
		entryIDs = append(entryIDs, entryID)
	}

	s.tasks[task.ID] = entryIDs
	return nil
}

// RemoveTask 移除任务（删除所有关联的时间点）
func (s *SchedulerService) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryIDs, exists := s.tasks[taskID]; exists {
		for _, id := range entryIDs {
			s.cron.Remove(id)
		}
		delete(s.tasks, taskID)
	}
}

// UpdateTask 更新任务
func (s *SchedulerService) UpdateTask(task *models.Task) error {
	s.RemoveTask(task.ID)
	return s.AddTask(task)
}

// executeTask 执行任务
func (s *SchedulerService) executeTask(task *models.Task) {
	// SG-003: 并发限制，达到上限时跳过本次触发
	if !s.executor.TryExecute() {
		log.Printf("并发达到上限，跳过本次触发(task_id=%s, task_name=%s)", task.ID, task.Name)
		// 写一条警告日志到数据库
		s.executor.SaveInfoLog("", task.ID, "并发达到上限，跳过本次定时触发")
		return
	}
	defer s.executor.ReleaseExecution()

	execution, err := s.executor.ExecuteScript(task)

	// 定时执行也写入执行历史，检查写库错误
	if dbErr := database.GetDB().Create(execution).Error; dbErr != nil {
		log.Printf("写入执行历史失败(task_id=%s, execution_id=%s): %v", task.ID, execution.ID, dbErr)
	}

	if err != nil && task.NotifyOnFailure {
		s.notifier.NotifyFailure(task, execution, err)
	}
}
