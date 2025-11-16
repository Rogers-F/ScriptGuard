package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"scriptguard/backend/models"
	"strings"
	"syscall"
)

type CondaService struct {
	scanAttempts int
	lastScanErr  error
}

func NewCondaService() *CondaService {
	return &CondaService{
		scanAttempts: 0,
	}
}

// ScanEnvironments 扫描所有conda环境
func (s *CondaService) ScanEnvironments() ([]models.Environment, error) {
	// 如果已经失败3次，直接返回上次的错误，避免无限重试
	if s.scanAttempts >= 3 && s.lastScanErr != nil {
		return []models.Environment{}, s.lastScanErr
	}

	s.scanAttempts++

	cmd := exec.Command("conda", "env", "list")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		s.lastScanErr = err
		// 达到3次重试后，返回友好的错误信息
		if s.scanAttempts >= 3 {
			return []models.Environment{}, fmt.Errorf("无法找到 Conda 环境（已尝试 %d 次）。请确保已安装 Anaconda/Miniconda 并添加到系统 PATH，或手动配置环境路径", s.scanAttempts)
		}
		return nil, err
	}

	// 成功扫描后重置计数器
	s.scanAttempts = 0
	s.lastScanErr = nil

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

// ValidateEnvironment 验证环境有效性
func (s *CondaService) ValidateEnvironment(envName string) bool {
	cmd := exec.Command("cmd", "/C", "conda activate "+envName+" && python --version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
