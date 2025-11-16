package backend

import (
	"context"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"scriptguard/backend/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	ctx       context.Context
	app       *application.App
	conda     *services.CondaService
	executor  *services.ExecutorService
	scheduler *services.SchedulerService
	notifier  *services.NotifierService
	cleanup   *services.CleanupService
}

func NewApp() *App {
	return &App{}
}

func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	a.app = options.App

	// 初始化数据库
	database.InitDB("./database/scriptguard.db")

	// 初始化服务
	a.conda = services.NewCondaService()
	a.executor = services.NewExecutorService()
	a.notifier = services.NewNotifierService("", "")
	a.scheduler = services.NewSchedulerService(a.executor, a.notifier)
	a.cleanup = services.NewCleanupService()

	// 启动调度器和清理服务
	a.scheduler.Start()
	a.cleanup.Start()

	// 加载已有任务
	a.loadTasks()

	// 启动日志流转发（将executor的logChan转发到前端Event）
	go a.startLogStreaming()

	return nil
}

func (a *App) ServiceShutdown() error {
	a.scheduler.Stop()
	a.cleanup.Stop()
	return nil
}

func (a *App) loadTasks() {
	var tasks []models.Task
	database.GetDB().Where("enabled = ?", true).Find(&tasks)

	for _, task := range tasks {
		a.scheduler.AddTask(&task)
	}
}

// GetEnvironments 获取所有conda环境
func (a *App) GetEnvironments() ([]models.Environment, error) {
	return a.conda.ScanEnvironments()
}

// GetTasks 获取所有任务
func (a *App) GetTasks() ([]models.Task, error) {
	var tasks []models.Task
	err := database.GetDB().Find(&tasks).Error
	return tasks, err
}

// CreateTask 创建任务
func (a *App) CreateTask(task models.Task) error {
	if err := database.GetDB().Create(&task).Error; err != nil {
		return err
	}
	return a.scheduler.AddTask(&task)
}

// UpdateTask 更新任务
func (a *App) UpdateTask(task models.Task) error {
	if err := database.GetDB().Save(&task).Error; err != nil {
		return err
	}
	return a.scheduler.UpdateTask(&task)
}

// DeleteTask 删除任务
func (a *App) DeleteTask(taskID string) error {
	a.scheduler.RemoveTask(taskID)
	return database.GetDB().Delete(&models.Task{}, "id = ?", taskID).Error
}

// ExecuteTaskNow 立即执行任务
func (a *App) ExecuteTaskNow(taskID string) (*models.Execution, error) {
	var task models.Task
	if err := database.GetDB().First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}

	execution, err := a.executor.ExecuteScript(&task)
	if err == nil {
		database.GetDB().Create(execution)
	}
	return execution, err
}

// GetExecutions 获取执行历史
func (a *App) GetExecutions(taskID string, limit int) ([]models.Execution, error) {
	var executions []models.Execution
	query := database.GetDB().Order("start_time DESC")

	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&executions).Error
	return executions, err
}

// GetLogs 获取日志
func (a *App) GetLogs(executionID string, taskID string, limit int) ([]models.Log, error) {
	var logs []models.Log
	query := database.GetDB().Order("timestamp ASC")

	if executionID != "" {
		query = query.Where("execution_id = ?", executionID)
	} else if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// startLogStreaming 启动日志流转发
func (a *App) startLogStreaming() {
	logChan := a.executor.GetLogChannel()
	for logMsg := range logChan {
		// 将日志消息转发到前端（通过Wails Events）
		// 注意：Wails 3的Events API可能需要调整
		// application.EmitEvent(a.ctx, "log:stream", logMsg)
		// 暂时注释，因为需要application实例的引用
		_ = logMsg
	}
}

// GetConfig 获取配置
func (a *App) GetConfig(key string) (string, error) {
	var config models.Config
	err := database.GetDB().Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// GetAllConfig 获取所有配置
func (a *App) GetAllConfig() (map[string]string, error) {
	var configs []models.Config
	err := database.GetDB().Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}
	return result, nil
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(key, value string) error {
	var config models.Config
	err := database.GetDB().Where("key = ?", key).First(&config).Error

	if err != nil {
		// 配置不存在，创建新配置
		config = models.Config{
			Key:   key,
			Value: value,
		}
		return database.GetDB().Create(&config).Error
	}

	// 更新现有配置
	config.Value = value
	return database.GetDB().Save(&config).Error
}

// SelectScriptFile 打开文件选择对话框选择Python脚本
func (a *App) SelectScriptFile() (string, error) {
	if a.app == nil {
		return "", nil
	}

	dialog := a.app.Dialog.OpenFile()
	dialog.SetTitle("选择Python脚本")
	dialog.AddFilter("Python脚本", "*.py")
	dialog.AddFilter("所有文件", "*.*")

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return path, nil
}
