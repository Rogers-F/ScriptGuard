package services

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"regexp"
	"scriptguard/backend/models"
	"strings"
	"syscall"
)

type CondaService struct{}

func NewCondaService() *CondaService {
	return &CondaService{}
}

// ScanEnvironments 扫描所有conda环境
func (s *CondaService) ScanEnvironments() ([]models.Environment, error) {
	cmd := exec.Command("conda", "env", "list")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
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

// ValidateEnvironment 验证环境有效性
func (s *CondaService) ValidateEnvironment(envName string) bool {
	cmd := exec.Command("cmd", "/C", "conda activate "+envName+" && python --version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
