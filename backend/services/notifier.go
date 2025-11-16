package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scriptguard/backend/models"
)

type NotifierService struct {
	dingTalkWebhook string
	wecomWebhook    string
}

func NewNotifierService(dingTalkWebhook, wecomWebhook string) *NotifierService {
	return &NotifierService{
		dingTalkWebhook: dingTalkWebhook,
		wecomWebhook:    wecomWebhook,
	}
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

	if s.dingTalkWebhook != "" {
		s.sendDingTalk(message)
	}

	if s.wecomWebhook != "" {
		s.sendWeCom(message)
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

	return s.sendWebhook(s.dingTalkWebhook, payload)
}

// sendWeCom 发送企业微信通知
func (s *NotifierService) sendWeCom(message string) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	return s.sendWebhook(s.wecomWebhook, payload)
}

// sendWebhook 发送webhook请求
func (s *NotifierService) sendWebhook(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
