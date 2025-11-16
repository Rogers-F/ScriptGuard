package services

import (
	"bufio"
	"fmt"
	"os/exec"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"syscall"
	"time"
)

type ExecutorService struct {
	logChan chan LogMessage
}

type LogMessage struct {
	ExecutionID string
	TaskID      string
	Timestamp   time.Time
	Level       string
	Content     string
}

func NewExecutorService() *ExecutorService {
	return &ExecutorService{
		logChan: make(chan LogMessage, 1000),
	}
}

// ExecuteScript 执行Python脚本
func (s *ExecutorService) ExecuteScript(task *models.Task) (*models.Execution, error) {
	execution := &models.Execution{
		TaskID: task.ID,
	}
	execution.BeforeCreate()

	// 构建命令
	cmdStr := fmt.Sprintf("conda activate %s && python %s", task.CondaEnv, task.ScriptPath)
	cmd := exec.Command("cmd", "/C", cmdStr)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// 捕获输出
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		execution.Status = models.StatusFailed
		execution.ErrorMessage = err.Error()
		now := time.Now()
		execution.EndTime = &now
		return execution, err
	}

	// 实时读取stdout并持久化
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logMsg := LogMessage{
				ExecutionID: execution.ID,
				TaskID:      task.ID,
				Timestamp:   time.Now(),
				Level:       "stdout",
				Content:     scanner.Text(),
			}
			s.logChan <- logMsg
			s.saveLog(&logMsg)
		}
	}()

	// 实时读取stderr并持久化
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logMsg := LogMessage{
				ExecutionID: execution.ID,
				TaskID:      task.ID,
				Timestamp:   time.Now(),
				Level:       "stderr",
				Content:     scanner.Text(),
			}
			s.logChan <- logMsg
			s.saveLog(&logMsg)
		}
	}()

	// 等待执行完成
	err := cmd.Wait()
	now := time.Now()
	execution.EndTime = &now
	execution.DurationMs = now.Sub(execution.StartTime).Milliseconds()

	if err != nil {
		execution.Status = models.StatusFailed
		execution.ExitCode = cmd.ProcessState.ExitCode()
		execution.ErrorMessage = err.Error()
	} else {
		execution.Status = models.StatusSuccess
		execution.ExitCode = 0
	}

	return execution, nil
}

// GetLogChannel 获取日志通道
func (s *ExecutorService) GetLogChannel() <-chan LogMessage {
	return s.logChan
}

// saveLog 保存日志到数据库
func (s *ExecutorService) saveLog(logMsg *LogMessage) {
	log := &models.Log{
		ExecutionID: logMsg.ExecutionID,
		TaskID:      logMsg.TaskID,
		Timestamp:   logMsg.Timestamp,
		Level:       models.LogLevel(logMsg.Level),
		Content:     logMsg.Content,
	}
	database.GetDB().Create(log)
}

// SaveInfoLog 保存信息日志
func (s *ExecutorService) SaveInfoLog(executionID, taskID, content string) {
	logMsg := &LogMessage{
		ExecutionID: executionID,
		TaskID:      taskID,
		Timestamp:   time.Now(),
		Level:       string(models.LogLevelInfo),
		Content:     content,
	}
	s.logChan <- *logMsg
	s.saveLog(logMsg)
}
