package services

import (
	"scriptguard/backend/models"
	"sync"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron     *cron.Cron
	tasks    map[string]cron.EntryID
	executor *ExecutorService
	notifier *NotifierService
	mu       sync.RWMutex
}

func NewSchedulerService(executor *ExecutorService, notifier *NotifierService) *SchedulerService {
	return &SchedulerService{
		cron:     cron.New(cron.WithSeconds()),
		tasks:    make(map[string]cron.EntryID),
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

// AddTask 添加任务
func (s *SchedulerService) AddTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !task.Enabled {
		return nil
	}

	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeTask(task)
	})

	if err != nil {
		return err
	}

	s.tasks[task.ID] = entryID
	return nil
}

// RemoveTask 移除任务
func (s *SchedulerService) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.tasks[taskID]; exists {
		s.cron.Remove(entryID)
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
	execution, err := s.executor.ExecuteScript(task)

	if err != nil && task.NotifyOnFailure {
		s.notifier.NotifyFailure(task, execution, err)
	}
}
