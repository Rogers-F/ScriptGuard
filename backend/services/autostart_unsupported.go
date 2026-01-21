//go:build !windows && !darwin

package services

import "errors"

// IsAutoStartEnabled 不支持的平台返回 false
func IsAutoStartEnabled(appName, exePath string, args []string) (bool, error) {
	return false, errors.New("当前系统不支持开机自启动设置")
}

// SetAutoStartEnabled 不支持的平台返回错误
func SetAutoStartEnabled(appName, exePath string, args []string, enabled bool) error {
	return errors.New("当前系统不支持开机自启动设置")
}
