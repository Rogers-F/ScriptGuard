package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"scriptguard/backend/services"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/wailsapp/wails/v3/pkg/application"
	"gorm.io/gorm"
)

// cronParser 用于校验 Cron 表达式
var cronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// validateCronExpr 校验单个 Cron 表达式是否合法
func validateCronExpr(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("Cron表达式不能为空")
	}
	if _, err := cronParser.Parse(expr); err != nil {
		return fmt.Errorf("Cron表达式非法: %w", err)
	}
	return nil
}

// validateCronExprs 校验多个 Cron 表达式（支持最多60个时间点）
func validateCronExprs(exprs []string) error {
	if len(exprs) == 0 {
		return fmt.Errorf("Cron表达式不能为空")
	}
	if len(exprs) > 60 {
		return fmt.Errorf("最多支持 60 个时间点")
	}
	seen := make(map[string]struct{}, len(exprs))
	for _, expr := range exprs {
		expr = strings.TrimSpace(expr)
		if err := validateCronExpr(expr); err != nil {
			return err
		}
		if _, ok := seen[expr]; ok {
			return fmt.Errorf("存在重复的时间点: %s", expr)
		}
		seen[expr] = struct{}{}
	}
	return nil
}

type App struct {
	ctx       context.Context
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

	// SG-002: 使用用户数据目录
	dbPath, err := database.GetDefaultDBPath()
	if err != nil {
		return fmt.Errorf("获取数据库路径失败: %w", err)
	}

	// 初始化数据库
	if err := database.InitDB(dbPath); err != nil {
		return err
	}

	// 初始化服务
	a.conda = services.NewCondaService()
	a.executor = services.NewExecutorService()
	a.notifier = services.NewNotifierService("", "")

	// 启动时从配置表加载 webhook
	if err := a.reloadNotifierConfig(); err != nil {
		return err
	}

	// SG-019: 启动时加载并发配置
	if err := a.reloadExecutorConfig(); err != nil {
		return err
	}

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
	// SG-079: 增加 nil 保护
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	if a.cleanup != nil {
		a.cleanup.Stop()
	}
	return database.CloseDB()
}

// reloadNotifierConfig 从配置表加载告警配置
func (a *App) reloadNotifierConfig() error {
	config, err := a.GetAllConfig()
	if err != nil {
		return err
	}
	a.notifier.SetWebhooks(
		config[models.ConfigKeyDingTalkWebhook],
		config[models.ConfigKeyWeComWebhook],
	)
	return nil
}

// SG-019: reloadExecutorConfig 从配置表加载执行器配置
func (a *App) reloadExecutorConfig() error {
	config, err := a.GetAllConfig()
	if err != nil {
		return err
	}

	// 加载最大并发数
	if val, ok := config[models.ConfigKeyMaxConcurrency]; ok && val != "" {
		if maxConcurrency, err := strconv.Atoi(val); err == nil && maxConcurrency > 0 {
			a.executor.SetMaxConcurrency(maxConcurrency)
			log.Printf("已加载最大并发配置: %d", maxConcurrency)
		}
	}

	// 加载执行超时（单位：秒；0 表示不限制）
	if val, ok := config[models.ConfigKeyExecutionTimeoutSeconds]; ok {
		val = strings.TrimSpace(val)
		if val != "" {
			seconds, err := strconv.Atoi(val)
			if err != nil {
				log.Printf("加载执行超时配置失败(key=%s, value=%q): %v，使用默认值",
					models.ConfigKeyExecutionTimeoutSeconds, val, err)
			} else if seconds == 0 {
				a.executor.SetTimeout(0)
				log.Printf("已加载执行超时配置: 0（不限制）")
			} else if seconds < 60 || seconds > 86400 {
				log.Printf("加载执行超时配置失败(key=%s, value=%d): 超出允许范围(0 或 60~86400 秒)，使用默认值",
					models.ConfigKeyExecutionTimeoutSeconds, seconds)
			} else {
				a.executor.SetTimeout(time.Duration(seconds) * time.Second)
				log.Printf("已加载执行超时配置: %d 秒", seconds)
			}
		}
	}

	return nil
}

func (a *App) loadTasks() {
	var tasks []models.Task
	// SG-010: 检查 DB 错误
	if err := database.GetDB().Where("enabled = ?", true).Find(&tasks).Error; err != nil {
		log.Printf("加载任务失败: %v", err)
		return
	}

	for i := range tasks {
		if err := a.scheduler.AddTask(&tasks[i]); err != nil {
			log.Printf("添加任务到调度器失败(task_id=%s): %v", tasks[i].ID, err)
		}
	}
}

// GetEnvironments 获取所有conda环境
func (a *App) GetEnvironments() ([]models.Environment, error) {
	return a.conda.ScanEnvironments()
}

// GetTasks 获取所有任务
func (a *App) GetTasks() ([]models.Task, error) {
	var tasks []models.Task
	if err := database.GetDB().Find(&tasks).Error; err != nil {
		return nil, err
	}
	// 归一化 cron 表达式，保证前端拿到稳定的 cron_exprs
	for i := range tasks {
		tasks[i].NormalizeCron()
	}
	return tasks, nil
}

// CreateTask 创建任务
// SG-024: 先写库成功，再加调度（避免 Commit 失败时调度已触发执行）
func (a *App) CreateTask(task models.Task) error {
	// 归一化并校验 Cron 表达式
	task.NormalizeCron()
	if err := validateCronExprs(task.CronExprs); err != nil {
		return err
	}

	db := database.GetDB()

	// 先写库
	if err := db.Create(&task).Error; err != nil {
		return err
	}

	// DB 写入成功后，再加调度
	if err := a.scheduler.AddTask(&task); err != nil {
		// 调度失败，补偿：禁用任务（保守策略，避免删除数据）
		task.Enabled = false
		_ = db.Save(&task)
		return fmt.Errorf("任务已创建但调度失败，已自动禁用: %w", err)
	}

	return nil
}

// UpdateTask 更新任务
// SG-024: 先停调度 → 写库 → 再恢复调度（与 DeleteTask 策略对齐）
func (a *App) UpdateTask(task models.Task) error {
	// 归一化并校验 Cron 表达式
	task.NormalizeCron()
	if err := validateCronExprs(task.CronExprs); err != nil {
		return err
	}

	db := database.GetDB()

	// 读取旧任务，用于失败时恢复调度
	var old models.Task
	if err := db.First(&old, "id = ?", task.ID).Error; err != nil {
		return err
	}

	// 1. 先停止旧调度（避免更新期间触发执行）
	a.scheduler.RemoveTask(task.ID)

	// 2. 写库
	if err := db.Save(&task).Error; err != nil {
		// 写库失败，恢复旧调度
		_ = a.scheduler.AddTask(&old)
		return err
	}

	// 3. 添加新调度
	if err := a.scheduler.AddTask(&task); err != nil {
		// 调度失败，恢复旧数据和旧调度
		_ = db.Save(&old)
		_ = a.scheduler.AddTask(&old)
		return fmt.Errorf("任务已更新但调度失败: %w", err)
	}

	return nil
}

// DeleteTask 删除任务
func (a *App) DeleteTask(taskID string) error {
	// SG-009: 先读取任务，删除失败时可恢复调度
	var task models.Task
	if err := database.GetDB().First(&task, "id = ?", taskID).Error; err != nil {
		return err
	}

	a.scheduler.RemoveTask(taskID)
	if err := database.GetDB().Delete(&models.Task{}, "id = ?", taskID).Error; err != nil {
		// 删除失败，恢复调度
		_ = a.scheduler.AddTask(&task)
		return err
	}
	return nil
}

// ExecuteTaskNow 立即执行任务
func (a *App) ExecuteTaskNow(taskID string) (*models.Execution, error) {
	// SG-020: 立即执行也需要检查并发限制
	if !a.executor.TryExecute() {
		return nil, fmt.Errorf("当前并发任务数已达上限，请稍后重试")
	}
	defer a.executor.ReleaseExecution()

	var task models.Task
	if err := database.GetDB().First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}

	execution, err := a.executor.ExecuteScript(&task)

	// 无论成功失败都记录执行历史，并检查写库错误
	if dbErr := database.GetDB().Create(execution).Error; dbErr != nil {
		if err == nil {
			// 脚本执行成功，但历史写入失败
			err = fmt.Errorf("脚本执行成功，但写入执行历史失败: %w", dbErr)
		} else {
			// 脚本失败 + 写库也失败
			err = fmt.Errorf("%w; 写入执行历史失败: %v", err, dbErr)
		}
	}

	return execution, err
}

// GetExecutions 获取执行历史
func (a *App) GetExecutions(taskID string, limit int) ([]models.Execution, error) {
	// SG-018: 服务端上限保护
	const maxLimit = 5000
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}

	var executions []models.Execution
	query := database.GetDB().Order("start_time DESC")

	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}

	query = query.Limit(limit)

	err := query.Find(&executions).Error
	return executions, err
}

// GetLogs 获取日志
func (a *App) GetLogs(executionID string, taskID string, limit int) ([]models.Log, error) {
	// SG-018: 服务端上限保护
	const maxLimit = 5000
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}

	var logs []models.Log

	query := database.GetDB()
	if executionID != "" {
		query = query.Where("execution_id = ?", executionID)
	} else if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}

	// 按时间倒序取最新 N 条，再在内存中反转成正序
	if err := query.Order("timestamp DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	// 反转切片：让返回结果按时间 ASC
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
	return logs, nil
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
	value = strings.TrimSpace(value)

	// 执行超时（秒）参数校验：0 表示不限制；允许范围 0 或 60~86400
	if key == models.ConfigKeyExecutionTimeoutSeconds {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s 必须为整数秒: %w", models.ConfigKeyExecutionTimeoutSeconds, err)
		}
		if seconds != 0 && (seconds < 60 || seconds > 86400) {
			return fmt.Errorf("%s 超出允许范围：0 或 60~86400（单位：秒）", models.ConfigKeyExecutionTimeoutSeconds)
		}
	}

	var config models.Config
	err := database.GetDB().Where("key = ?", key).First(&config).Error

	if err != nil {
		// SG-008: 仅 ErrRecordNotFound 时创建，其他错误返回
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config = models.Config{
				Key:   key,
				Value: value,
			}
			if err := database.GetDB().Create(&config).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// 更新现有配置
		config.Value = value
		if err := database.GetDB().Save(&config).Error; err != nil {
			return err
		}
	}

	// 热更新告警配置
	if key == models.ConfigKeyDingTalkWebhook || key == models.ConfigKeyWeComWebhook {
		if err := a.reloadNotifierConfig(); err != nil {
			return err
		}
	}

	// SG-019: 热更新并发配置
	if key == models.ConfigKeyMaxConcurrency {
		if maxConcurrency, err := strconv.Atoi(value); err == nil && maxConcurrency > 0 {
			a.executor.SetMaxConcurrency(maxConcurrency)
			log.Printf("已热更新最大并发配置: %d", maxConcurrency)
		}
	}

	// 热更新执行超时配置
	if key == models.ConfigKeyExecutionTimeoutSeconds {
		seconds, _ := strconv.Atoi(value) // 已校验
		a.executor.SetTimeout(time.Duration(seconds) * time.Second)
		log.Printf("已热更新执行超时配置: %d 秒", seconds)
	}

	return nil
}

// SelectScriptFile 打开文件选择对话框选择Python脚本
func (a *App) SelectScriptFile() (string, error) {
	dialog := application.OpenFileDialog()
	dialog.SetTitle("选择Python脚本")
	dialog.AddFilter("Python脚本", "*.py")
	dialog.AddFilter("所有文件", "*.*")

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return path, nil
}

// SG-013: TestNotification 测试通知
func (a *App) TestNotification(target string, webhook string) error {
	return a.notifier.SendTest(target, webhook)
}

// ==================== 开机自启动 API ====================

const autoStartAppName = "ScriptGuard"

// GetAutoStartEnabled 获取当前"系统层面"的自启动状态
func (a *App) GetAutoStartEnabled() bool {
	exePath, err := os.Executable()
	if err == nil {
		// 开机自启动时使用 --autostart 参数
		enabled, err := services.IsAutoStartEnabled(autoStartAppName, exePath, []string{"--autostart"})
		if err == nil {
			return enabled
		}
		log.Printf("读取开机自启动状态失败: %v，回退到配置表", err)
	} else {
		log.Printf("获取应用可执行文件路径失败: %v，回退到配置表", err)
	}

	cfg, err := a.GetAllConfig()
	if err != nil {
		log.Printf("读取配置表失败: %v", err)
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg[models.ConfigKeyAutoStartEnabled]), "true")
}

// SetAutoStartEnabled 设置开机自启动
func (a *App) SetAutoStartEnabled(enabled bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取应用可执行文件路径失败: %w", err)
	}

	// 开机自启动时使用 --autostart 参数，实现静默启动到托盘
	args := []string{"--autostart"}

	// 记录变更前系统状态，用于失败回滚
	oldEnabled, err := services.IsAutoStartEnabled(autoStartAppName, exePath, args)
	if err != nil {
		return fmt.Errorf("读取当前开机自启动状态失败: %w", err)
	}

	if err := services.SetAutoStartEnabled(autoStartAppName, exePath, args, enabled); err != nil {
		return fmt.Errorf("设置开机自启动失败: %w", err)
	}

	// 设置后校验
	newEnabled, err := services.IsAutoStartEnabled(autoStartAppName, exePath, args)
	if err != nil {
		rollbackErr := services.SetAutoStartEnabled(autoStartAppName, exePath, args, oldEnabled)
		if rollbackErr != nil {
			return fmt.Errorf("设置后校验失败: %v；且回滚失败: %v", err, rollbackErr)
		}
		return fmt.Errorf("设置后校验失败，已回滚: %w", err)
	}
	if newEnabled != enabled {
		rollbackErr := services.SetAutoStartEnabled(autoStartAppName, exePath, args, oldEnabled)
		if rollbackErr != nil {
			return fmt.Errorf("开机自启动状态校验失败(期望=%v 实际=%v)；且回滚失败: %v", enabled, newEnabled, rollbackErr)
		}
		return fmt.Errorf("开机自启动状态校验失败(期望=%v 实际=%v)，已回滚", enabled, newEnabled)
	}

	// 写入配置表
	value := "false"
	if enabled {
		value = "true"
	}
	if err := a.UpdateConfig(models.ConfigKeyAutoStartEnabled, value); err != nil {
		rollbackErr := services.SetAutoStartEnabled(autoStartAppName, exePath, args, oldEnabled)
		if rollbackErr != nil {
			return fmt.Errorf("写入配置失败: %v；且回滚开机自启动失败: %v", err, rollbackErr)
		}
		return fmt.Errorf("写入配置失败，已回滚开机自启动: %w", err)
	}

	return nil
}

// ExportDebugLogs 导出调试日志到文件
func (a *App) ExportDebugLogs(frontendLogs string) (string, error) {
	// 打开保存文件对话框
	dialog := application.SaveFileDialog()

	// Wails v3 alpha.41: SaveFileDialogStruct 没有 SetTitle，需要通过 SetOptions 设置
	dialog.SetOptions(&application.SaveFileDialogOptions{
		Title:                "导出调试日志",
		CanCreateDirectories: true,
		Filename:             "ScriptGuard_debug_" + time.Now().Format("20060102_150405") + ".txt",
	})
	dialog.AddFilter("文本文件", "*.txt")

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // 用户取消
	}

	// 构建日志内容
	var content strings.Builder

	// 写入头部信息
	content.WriteString("========================================\n")
	content.WriteString("ScriptGuard 调试日志导出\n")
	content.WriteString("========================================\n")
	content.WriteString(fmt.Sprintf("导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("应用版本: %s\n", "1.2.0"))
	content.WriteString("\n")

	// 写入系统信息
	content.WriteString("--- 系统信息 ---\n")
	content.WriteString(fmt.Sprintf("操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	content.WriteString(fmt.Sprintf("Go版本: %s\n", runtime.Version()))
	content.WriteString("\n")

	// 写入前端控制台日志（如果有）
	if frontendLogs != "" {
		content.WriteString("--- 前端控制台日志 ---\n")
		content.WriteString(frontendLogs)
		content.WriteString("\n\n")
	}

	// 从数据库获取最近的任务执行日志（最近500条）
	var logs []models.Log
	database.GetDB().Order("timestamp DESC").Limit(500).Find(&logs)

	content.WriteString("--- 任务执行日志（最近500条）---\n")
	if len(logs) == 0 {
		content.WriteString("暂无日志\n")
	} else {
		for i := len(logs) - 1; i >= 0; i-- { // 倒序输出，最早的在前
			logEntry := logs[i]
			content.WriteString(fmt.Sprintf("[%s] [%s] %s\n",
				logEntry.Timestamp.Format("2006-01-02 15:04:05"),
				strings.ToUpper(string(logEntry.Level)),
				logEntry.Content,
			))
		}
	}

	// 写入文件
	err = os.WriteFile(path, []byte(content.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return path, nil
}
