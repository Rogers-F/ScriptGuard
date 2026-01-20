package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"scriptguard/backend/database"
	"scriptguard/backend/models"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// 日志处理常量
const (
	maxLogLineBytes  = 1024 * 1024           // 单行日志最大长度（超过截断，但继续 drain）
	logQueueSize     = 2000                  // 单次执行日志写入队列容量（满则丢弃并统计）
	logBatchSize     = 200                   // 批量写入条数
	logFlushInterval = 200 * time.Millisecond // 批量写入间隔
)

type ExecutorService struct {
	logChan chan LogMessage
	limiter *ConcurrencyLimiter
	timeout time.Duration // 0 表示不限制
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
		limiter: NewConcurrencyLimiter(5), // 默认最大并发 5
		timeout: 0,                        // 默认不限制超时
	}
}

// SetMaxConcurrency 设置最大并发数
func (s *ExecutorService) SetMaxConcurrency(max int) {
	s.limiter.SetMax(max)
}

// SetTimeout 设置执行超时时间
func (s *ExecutorService) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// TryExecute 尝试执行（非阻塞，达到上限时返回 false）
func (s *ExecutorService) TryExecute() bool {
	return s.limiter.TryAcquire()
}

// ReleaseExecution 释放执行权限
func (s *ExecutorService) ReleaseExecution() {
	s.limiter.Release()
}

// ExecuteScript 执行Python脚本
func (s *ExecutorService) ExecuteScript(task *models.Task) (*models.Execution, error) {
	execution := &models.Execution{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		StartTime: NowBeijing(), // SG-023: 使用北京时间
		Status:    models.StatusRunning,
	}

	// SG-004: 支持超时控制
	var cmd *exec.Cmd
	var cancel context.CancelFunc
	if s.timeout > 0 {
		ctx, c := context.WithTimeout(context.Background(), s.timeout)
		cancel = c
		cmd = exec.CommandContext(ctx, "conda", "run", "-n", task.CondaEnv, "python", task.ScriptPath)
	} else {
		cmd = exec.Command("conda", "run", "-n", task.CondaEnv, "python", task.ScriptPath)
		cancel = func() {}
	}
	defer cancel()

	// SG-029: 设置 UTF-8 编码，修复 Windows conda 中文输出乱码问题
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// SG-005: 检查 Pipe 错误
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		execution.Status = models.StatusFailed
		execution.ErrorMessage = "创建 stdout pipe 失败: " + err.Error()
		now := NowBeijing() // SG-023: 使用北京时间
		execution.EndTime = &now
		return execution, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close() // SG-022: 关闭已创建的 stdout pipe
		execution.Status = models.StatusFailed
		execution.ErrorMessage = "创建 stderr pipe 失败: " + err.Error()
		now := NowBeijing() // SG-023: 使用北京时间
		execution.EndTime = &now
		return execution, err
	}

	if err := cmd.Start(); err != nil {
		// SG-022: 关闭已创建的 pipe，避免句柄泄漏
		stdout.Close()
		stderr.Close()
		execution.Status = models.StatusFailed
		execution.ErrorMessage = err.Error()
		now := NowBeijing() // SG-023: 使用北京时间
		execution.EndTime = &now
		return execution, err
	}

	// 单次执行的日志写入队列（避免高输出阻塞 pipe 读取）
	logQueue := make(chan *models.Log, logQueueSize)
	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)

		ticker := time.NewTicker(logFlushInterval)
		defer ticker.Stop()

		batch := make([]*models.Log, 0, logBatchSize)

		flush := func() {
			if len(batch) == 0 {
				return
			}
			if err := database.GetDB().CreateInBatches(batch, logBatchSize).Error; err != nil {
				log.Printf("批量保存日志失败(execution_id=%s): %v", execution.ID, err)
			}
			batch = batch[:0]
		}

		for {
			select {
			case entry, ok := <-logQueue:
				if !ok {
					flush()
					return
				}
				batch = append(batch, entry)
				if len(batch) >= logBatchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()

	// 使用 WaitGroup 确保 stdout/stderr 读取完成
	var wg sync.WaitGroup
	wg.Add(2)

	droppedStdout := 0
	droppedStderr := 0

	// 实时读取 stdout（单行最多 1MB，超出部分截断但继续 drain）
	go func() {
		defer wg.Done()
		reader := bufio.NewReaderSize(stdout, 64*1024)

		for {
			line, truncated, readAny, readErr := readLineWithLimit(reader, maxLogLineBytes)
			if readAny {
				content := line
				if truncated {
					content += " ...(已截断)"
				}

				ts := NowBeijing() // SG-023: 使用北京时间

				logMsg := LogMessage{
					ExecutionID: execution.ID,
					TaskID:      task.ID,
					Timestamp:   ts,
					Level:       string(models.LogLevelStdout),
					Content:     content,
				}
				s.logChan <- logMsg

				entry := &models.Log{
					ExecutionID: execution.ID,
					TaskID:      task.ID,
					Timestamp:   ts,
					Level:       models.LogLevelStdout,
					Content:     content,
				}

				select {
				case logQueue <- entry:
				default:
					droppedStdout++
				}
			}

			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				log.Printf("读取 stdout 错误(execution_id=%s): %v", execution.ID, readErr)
				break
			}
		}
	}()

	// 实时读取 stderr（单行最多 1MB，超出部分截断但继续 drain）
	go func() {
		defer wg.Done()
		reader := bufio.NewReaderSize(stderr, 64*1024)

		for {
			line, truncated, readAny, readErr := readLineWithLimit(reader, maxLogLineBytes)
			if readAny {
				content := line
				if truncated {
					content += " ...(已截断)"
				}

				ts := NowBeijing() // SG-023: 使用北京时间

				logMsg := LogMessage{
					ExecutionID: execution.ID,
					TaskID:      task.ID,
					Timestamp:   ts,
					Level:       string(models.LogLevelStderr),
					Content:     content,
				}
				s.logChan <- logMsg

				entry := &models.Log{
					ExecutionID: execution.ID,
					TaskID:      task.ID,
					Timestamp:   ts,
					Level:       models.LogLevelStderr,
					Content:     content,
				}

				select {
				case logQueue <- entry:
				default:
					droppedStderr++
				}
			}

			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				log.Printf("读取 stderr 错误(execution_id=%s): %v", execution.ID, readErr)
				break
			}
		}
	}()

	// 等待 stdout/stderr 读取完成（确保 pipe 被 drain）
	wg.Wait()

	// 刷新并结束批量写入
	close(logQueue)
	<-writerDone

	// 写入丢弃汇总（避免静默丢日志）
	if droppedStdout > 0 || droppedStderr > 0 {
		summary := fmt.Sprintf(
			"日志过多已丢弃：stdout=%d 行，stderr=%d 行（为避免阻塞执行，系统已自动丢弃部分日志）",
			droppedStdout, droppedStderr,
		)

		summaryLog := &models.Log{
			ExecutionID: execution.ID,
			TaskID:      task.ID,
			Timestamp:   NowBeijing(),
			Level:       models.LogLevelWarning,
			Content:     summary,
		}
		if err := database.GetDB().Create(summaryLog).Error; err != nil {
			log.Printf("保存日志丢弃汇总失败(execution_id=%s): %v", execution.ID, err)
		}
	}

	// 等待执行完成
	err = cmd.Wait()
	now := NowBeijing() // SG-023: 使用北京时间
	execution.EndTime = &now
	execution.DurationMs = now.Sub(execution.StartTime).Milliseconds()

	if err != nil {
		execution.Status = models.StatusFailed
		if cmd.ProcessState != nil {
			execution.ExitCode = cmd.ProcessState.ExitCode()
		}
		// 检查是否超时
		if s.timeout > 0 && execution.DurationMs >= s.timeout.Milliseconds() {
			execution.ErrorMessage = "执行超时: " + err.Error()
		} else {
			execution.ErrorMessage = err.Error()
		}
	} else {
		execution.Status = models.StatusSuccess
		execution.ExitCode = 0
	}

	return execution, err
}

// GetLogChannel 获取日志通道
func (s *ExecutorService) GetLogChannel() <-chan LogMessage {
	return s.logChan
}

// saveLog 保存日志到数据库
func (s *ExecutorService) saveLog(logMsg *LogMessage) {
	logEntry := &models.Log{
		ExecutionID: logMsg.ExecutionID,
		TaskID:      logMsg.TaskID,
		Timestamp:   logMsg.Timestamp,
		Level:       models.LogLevel(logMsg.Level),
		Content:     logMsg.Content,
	}
	if err := database.GetDB().Create(logEntry).Error; err != nil {
		log.Printf("保存日志失败: %v", err)
	}
}

// SaveInfoLog 保存信息日志
func (s *ExecutorService) SaveInfoLog(executionID, taskID, content string) {
	logMsg := &LogMessage{
		ExecutionID: executionID,
		TaskID:      taskID,
		Timestamp:   NowBeijing(), // SG-023: 使用北京时间
		Level:       string(models.LogLevelInfo),
		Content:     content,
	}
	s.logChan <- *logMsg
	s.saveLog(logMsg)
}

// readLineWithLimit 读取一行，超过 maxBytes 截断但继续 drain 直到换行
func readLineWithLimit(r *bufio.Reader, maxBytes int) (line string, truncated bool, readAny bool, err error) {
	buf := make([]byte, 0, 1024)
	truncated = false
	readAny = false

	for {
		fragment, e := r.ReadSlice('\n')
		if len(fragment) > 0 {
			readAny = true
		}

		// fragment 可能是"整行的一部分"，即使 ErrBufferFull 也需要继续读取直到遇到 '\n' 或 EOF
		if len(fragment) > 0 {
			if len(buf) < maxBytes {
				remain := maxBytes - len(buf)
				if len(fragment) > remain {
					buf = append(buf, fragment[:remain]...)
					truncated = true
				} else {
					buf = append(buf, fragment...)
				}
			} else {
				truncated = true
			}
		}

		// bufio.ErrBufferFull：还没读到 '\n'，继续读下一段
		if errors.Is(e, bufio.ErrBufferFull) {
			continue
		}

		// 统一去掉行尾换行符，避免日志落库携带 '\n'
		line = strings.TrimRight(string(buf), "\r\n")
		err = e
		return
	}
}
