package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramChannel sends alerts to Telegram
type TelegramChannel struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramChannel(botToken, chatID string) *TelegramChannel {
	return &TelegramChannel{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TelegramChannel) Name() string { return "telegram" }

func (c *TelegramChannel) Send(ctx context.Context, alert *Alert) error {
	emoji := map[AlertLevel]string{
		AlertInfo: "ℹ️", AlertWarning: "⚠️", AlertCritical: "🔴",
	}
	text := fmt.Sprintf("%s *%s*\n%s\n_%s / %s_", emoji[alert.Level], alert.Title, alert.Message, alert.ResourceType, alert.ResourceID)

	payload := map[string]string{"chat_id": c.chatID, "text": text, "parse_mode": "Markdown"}
	data, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// DingTalkChannel sends alerts to DingTalk (钉钉)
type DingTalkChannel struct {
	webhookURL string
	client     *http.Client
}

func NewDingTalkChannel(webhookURL string) *DingTalkChannel {
	return &DingTalkChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *DingTalkChannel) Name() string { return "dingtalk" }

func (c *DingTalkChannel) Send(ctx context.Context, alert *Alert) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[%s] %s\n%s", alert.Level, alert.Title, alert.Message),
		},
	}
	data, _ := json.Marshal(payload)
	resp, err := c.client.Post(c.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// EmailChannel sends alerts via email (SMTP)
type EmailChannel struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
}

func NewEmailChannel(smtpHost string, smtpPort int, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{smtpHost: smtpHost, smtpPort: smtpPort, username: username, password: password, from: from, to: to}
}

func (c *EmailChannel) Name() string { return "email" }

func (c *EmailChannel) Send(ctx context.Context, alert *Alert) error {
	// TODO: Implement SMTP email sending
	return nil
}
