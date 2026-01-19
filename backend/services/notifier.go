package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"scriptguard/backend/models"
	"strings"
	"sync"
	"time"
)

type NotifierService struct {
	mu              sync.RWMutex
	dingTalkWebhook string
	wecomWebhook    string
	client          *http.Client
}

func NewNotifierService(dingTalkWebhook, wecomWebhook string) *NotifierService {
	return &NotifierService{
		dingTalkWebhook: dingTalkWebhook,
		wecomWebhook:    wecomWebhook,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// SetWebhooks 用于运行时热更新告警配置
func (s *NotifierService) SetWebhooks(dingTalkWebhook, wecomWebhook string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dingTalkWebhook = dingTalkWebhook
	s.wecomWebhook = wecomWebhook
}

func (s *NotifierService) getWebhooks() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dingTalkWebhook, s.wecomWebhook
}

// NotifyFailure 发送失败通知
func (s *NotifierService) NotifyFailure(task *models.Task, execution *models.Execution, err error) {
	message := fmt.Sprintf(
		"⚠️ 脚本执行失败\n\n任务名称: %s\n脚本路径: %s\n环境: %s\n错误信息: %s\n执行ID: %s",
		task.Name,
		task.ScriptPath,
		task.CondaEnv,
		err.Error(),
		execution.ID,
	)

	ding, wecom := s.getWebhooks()

	if ding != "" {
		if sendErr := s.sendDingTalk(message); sendErr != nil {
			log.Printf("发送钉钉告警失败: %v", sendErr)
		}
	}

	if wecom != "" {
		if sendErr := s.sendWeCom(message); sendErr != nil {
			log.Printf("发送企业微信告警失败: %v", sendErr)
		}
	}
}

// sendDingTalk 发送钉钉通知
func (s *NotifierService) sendDingTalk(message string) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	ding, _ := s.getWebhooks()
	return s.sendWebhook(ding, payload)
}

// sendWeCom 发送企业微信通知
func (s *NotifierService) sendWeCom(message string) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	_, wecom := s.getWebhooks()
	return s.sendWebhook(wecom, payload)
}

// sendWebhook 发送webhook请求（带超时和状态码检查）
func (s *NotifierService) sendWebhook(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("webhook返回非2xx: %s, body=%s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

// SG-013: SendTest 发送测试通知
func (s *NotifierService) SendTest(target string, webhook string) error {
	message := fmt.Sprintf("✅ ScriptGuard 测试通知\n时间: %s\n\n如果您收到此消息，说明告警配置正确！",
		time.Now().Format("2006-01-02 15:04:05"))

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	url := strings.TrimSpace(webhook)
	if url == "" {
		// 如果未传入 webhook，使用已保存的配置
		ding, wecom := s.getWebhooks()
		switch target {
		case "dingtalk":
			url = ding
		case "wecom":
			url = wecom
		default:
			return fmt.Errorf("未知通知类型: %s", target)
		}
	}

	if url == "" {
		return fmt.Errorf("Webhook 未配置")
	}

	return s.sendWebhook(url, payload)
}
