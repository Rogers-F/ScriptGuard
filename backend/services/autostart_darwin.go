//go:build darwin

package services

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsAutoStartEnabled 判断 LaunchAgents plist 是否存在
func IsAutoStartEnabled(appName, exePath string, args []string) (bool, error) {
	plistPath, err := launchAgentPlistPath(appName)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(plistPath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("检查 LaunchAgents 文件失败: %w", err)
}

// SetAutoStartEnabled 设置 Mac 登录自启动
func SetAutoStartEnabled(appName, exePath string, args []string, enabled bool) error {
	plistPath, err := launchAgentPlistPath(appName)
	if err != nil {
		return err
	}

	if !enabled {
		if err := os.Remove(plistPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("删除 LaunchAgents 文件失败: %w", err)
		}
		return nil
	}

	if strings.TrimSpace(exePath) == "" {
		return fmt.Errorf("exePath 不能为空")
	}

	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建 LaunchAgents 目录失败: %w", err)
	}

	label := launchAgentLabel(appName)
	plist, err := buildLaunchAgentPlist(label, exePath, args)
	if err != nil {
		return err
	}

	// 原子写入
	tmpFile, err := os.CreateTemp(dir, filepath.Base(plistPath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("创建临时 plist 文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.Write([]byte(plist)); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("写入临时 plist 文件失败: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("关闭临时 plist 文件失败: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return fmt.Errorf("设置 plist 文件权限失败: %w", err)
	}
	if err := os.Rename(tmpPath, plistPath); err != nil {
		return fmt.Errorf("替换 LaunchAgents plist 失败: %w", err)
	}
	return nil
}

func launchAgentPlistPath(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	label := launchAgentLabel(appName)
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func launchAgentLabel(appName string) string {
	name := strings.ToLower(strings.TrimSpace(appName))
	name = strings.ReplaceAll(name, " ", "")
	if name == "" {
		name = "scriptguard"
	}
	return "com.scriptguard." + name
}

func buildLaunchAgentPlist(label, exePath string, args []string) (string, error) {
	if strings.TrimSpace(label) == "" {
		return "", fmt.Errorf("LaunchAgent Label 不能为空")
	}

	programArgs := make([]string, 0, 1+len(args))
	programArgs = append(programArgs, exePath)
	programArgs = append(programArgs, args...)

	escLabel, err := escapeXMLText(label)
	if err != nil {
		return "", fmt.Errorf("转义 LaunchAgent Label 失败: %w", err)
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")
	b.WriteString("  <key>Label</key>\n")
	b.WriteString("  <string>" + escLabel + "</string>\n")
	b.WriteString("  <key>ProgramArguments</key>\n")
	b.WriteString("  <array>\n")
	for _, arg := range programArgs {
		escArg, err := escapeXMLText(arg)
		if err != nil {
			return "", fmt.Errorf("转义 ProgramArguments 失败: %w", err)
		}
		b.WriteString("    <string>" + escArg + "</string>\n")
	}
	b.WriteString("  </array>\n")
	b.WriteString("  <key>RunAtLoad</key>\n")
	b.WriteString("  <true/>\n")
	b.WriteString("</dict>\n")
	b.WriteString("</plist>\n")
	return b.String(), nil
}

func escapeXMLText(s string) (string, error) {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
