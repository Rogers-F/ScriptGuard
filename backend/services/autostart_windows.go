//go:build windows

package services

import (
	"errors"
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// HKCU\Software\Microsoft\Windows\CurrentVersion\Run
// 作用：用户登录后自启动（无需管理员权限）
const windowsRunKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// IsAutoStartEnabled 判断"当前可执行文件"是否已配置为开机自启动
func IsAutoStartEnabled(appName, exePath string, args []string) (bool, error) {
	if strings.TrimSpace(appName) == "" {
		return false, fmt.Errorf("appName 不能为空")
	}
	if strings.TrimSpace(exePath) == "" {
		return false, fmt.Errorf("exePath 不能为空")
	}

	runKey, err := registry.OpenKey(registry.CURRENT_USER, windowsRunKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("打开注册表 Run 键失败: %w", err)
	}
	defer runKey.Close()

	existing, _, err := runKey.GetStringValue(appName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("读取注册表 Run 值失败: %w", err)
	}

	expected := windowsRunCommand(exePath, args)
	return strings.TrimSpace(existing) == expected, nil
}

// SetAutoStartEnabled 设置 Windows 登录自启动
func SetAutoStartEnabled(appName, exePath string, args []string, enabled bool) error {
	if strings.TrimSpace(appName) == "" {
		return fmt.Errorf("appName 不能为空")
	}
	if strings.TrimSpace(exePath) == "" {
		return fmt.Errorf("exePath 不能为空")
	}

	if enabled {
		runKey, _, err := registry.CreateKey(registry.CURRENT_USER, windowsRunKeyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("打开/创建注册表 Run 键失败: %w", err)
		}
		defer runKey.Close()

		cmd := windowsRunCommand(exePath, args)
		if err := runKey.SetStringValue(appName, cmd); err != nil {
			return fmt.Errorf("写入注册表 Run 值失败: %w", err)
		}
		return nil
	}

	runKey, err := registry.OpenKey(registry.CURRENT_USER, windowsRunKeyPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("打开注册表 Run 键失败: %w", err)
	}
	defer runKey.Close()

	if err := runKey.DeleteValue(appName); err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("删除注册表 Run 值失败: %w", err)
	}
	return nil
}

// windowsRunCommand 生成写入注册表的命令行
func windowsRunCommand(exePath string, args []string) string {
	parts := make([]string, 0, 1+len(args))
	parts = append(parts, syscall.EscapeArg(exePath))
	for _, a := range args {
		parts = append(parts, syscall.EscapeArg(a))
	}
	return strings.Join(parts, " ")
}
