package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"scriptguard/backend/models"
	"strings"
	"syscall"
	"time"
)

// SG-028: 删除未使用的 scanAttempts 和 lastScanErr 字段
type CondaService struct{}

func NewCondaService() *CondaService {
	return &CondaService{}
}

// SG-027: conda 命令超时时间
const condaTimeout = 30 * time.Second

// ScanEnvironments 扫描所有conda环境
func (s *CondaService) ScanEnvironments() ([]models.Environment, error) {
	// SG-027: 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), condaTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "conda", "env", "list")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		// 检查是否超时
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("执行 conda env list 超时（%v）", condaTimeout)
		}
		// 将 stderr 合入错误信息，提升可诊断性
		stderr := strings.TrimSpace(errOut.String())
		if stderr != "" {
			return nil, fmt.Errorf("执行 conda env list 失败: %w; stderr=%s", err, stderr)
		}
		return nil, fmt.Errorf("执行 conda env list 失败: %w。请确保已安装 Anaconda/Miniconda 并添加到系统 PATH", err)
	}

	envs := []models.Environment{}
	lines := strings.Split(out.String(), "\n")
	re := regexp.MustCompile(`^(\S+)\s+\*?\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			envName := matches[1]
			envPath := matches[2]
			pythonPath := filepath.Join(envPath, "python.exe")

			env := models.Environment{
				Name:       envName,
				Path:       envPath,
				PythonPath: pythonPath,
				IsValid:    s.ValidateEnvironment(envName),
			}
			envs = append(envs, env)
		}
	}

	return envs, nil
}

// ValidateEnvironment 验证环境有效性（使用参数化执行，避免注入）
func (s *CondaService) ValidateEnvironment(envName string) bool {
	// SG-027: 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), condaTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "conda", "run", "-n", envName, "python", "--version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
