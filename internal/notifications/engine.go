// Notifications - Email, Telegram, Webhook notifications
package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

// Notification types
const (
	NotifyEmail    = "email"
	NotifyTelegram = "telegram"
	NotifyWebhook  = "webhook"
	NotifySlack    = "slack"
)

// Notification levels
const (
	LevelInfo    = "info"
	LevelWarning = "warning"
	LevelError   = "error"
	LevelSuccess = "success"
)

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Enabled    bool     `json:"enabled"`
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`

	// Email settings
	SMTPServer string `json:"smtp_server"`
	SMTPPort   int    `json:"smtp_port"`
	SMTPUser   string `json:"smtp_user"`
	SMTPPass   string `json:"-"`
	FromEmail  string `json:"from_email"`
	UseTLS     bool   `json:"use_tls"`

	// Telegram settings
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`

	// Webhook settings
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"-"`
	WebhookMethod string `json:"webhook_method"`

	// Common
	MinLevel string   `json:"min_level"` // minimum level to send
	Events   []string `json:"events"`    // backup_start, backup_success, backup_failed, etc.
}

// Notification represents a notification message
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Level     string                 `json:"level"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NotificationEngine handles all notifications
type NotificationEngine struct {
	configs []*NotificationConfig
	sent    int
	failed  int
}

// NewNotificationEngine creates a new notification engine
func NewNotificationEngine() *NotificationEngine {
	return &NotificationEngine{
		configs: make([]*NotificationConfig, 0),
	}
}

// AddConfig adds a notification configuration
func (e *NotificationEngine) AddConfig(config *NotificationConfig) error {
	if err := e.validateConfig(config); err != nil {
		return err
	}
	e.configs = append(e.configs, config)
	return nil
}

// RemoveConfig removes a notification configuration
func (e *NotificationEngine) RemoveConfig(name string) {
	for i, config := range e.configs {
		if config.Name == name {
			e.configs = append(e.configs[:i], e.configs[i+1:]...)
			return
		}
	}
}

// Send sends a notification to all configured channels
func (e *NotificationEngine) Send(notification *Notification) {
	notification.Timestamp = time.Now()

	for _, config := range e.configs {
		if !config.Enabled {
			continue
		}

		// Check if event should be sent
		if !e.shouldSendEvent(config, notification) {
			continue
		}

		// Check minimum level
		if !e.checkMinLevel(config, notification.Level) {
			continue
		}

		// Send based on type
		switch config.Type {
		case NotifyEmail:
			e.sendEmail(config, notification)
		case NotifyTelegram:
			e.sendTelegram(config, notification)
		case NotifyWebhook:
			e.sendWebhook(config, notification)
		case NotifySlack:
			e.sendSlack(config, notification)
		}
	}
}

// SendBackupStarted sends backup started notification
func (e *NotificationEngine) SendBackupStarted(jobName, jobType string) {
	notification := &Notification{
		ID:      fmt.Sprintf("backup_start_%d", time.Now().Unix()),
		Type:    "backup",
		Level:   LevelInfo,
		Title:   "🚀 Резервне копіювання розпочато",
		Message: fmt.Sprintf("Завдання '%s' (%s) розпочало виконання", jobName, jobType),
		Data: map[string]interface{}{
			"job_name": jobName,
			"job_type": jobType,
			"event":    "backup_start",
		},
	}
	e.Send(notification)
}

// SendBackupSuccess sends backup success notification
func (e *NotificationEngine) SendBackupSuccess(jobName string, duration time.Duration, filesProcessed int, bytesWritten int64) {
	notification := &Notification{
		ID:    fmt.Sprintf("backup_success_%d", time.Now().Unix()),
		Type:  "backup",
		Level: LevelSuccess,
		Title: "✅ Резервне копіювання успішно завершено",
		Message: fmt.Sprintf("Завдання '%s' успішно завершено за %s. Оброблено файлів: %d, Розмір: %s",
			jobName, duration.Round(time.Second), filesProcessed, formatBytes(bytesWritten)),
		Data: map[string]interface{}{
			"job_name":        jobName,
			"duration":        duration.String(),
			"files_processed": filesProcessed,
			"bytes_written":   bytesWritten,
			"event":           "backup_success",
		},
	}
	e.Send(notification)
}

// SendBackupFailed sends backup failed notification
func (e *NotificationEngine) SendBackupFailed(jobName string, err error) {
	notification := &Notification{
		ID:      fmt.Sprintf("backup_failed_%d", time.Now().Unix()),
		Type:    "backup",
		Level:   LevelError,
		Title:   "❌ Помилка резервного копіювання",
		Message: fmt.Sprintf("Завдання '%s' не виконано: %v", jobName, err),
		Data: map[string]interface{}{
			"job_name": jobName,
			"error":    err.Error(),
			"event":    "backup_failed",
		},
	}
	e.Send(notification)
}

// SendRestoreStarted sends restore started notification
func (e *NotificationEngine) SendRestoreStarted(jobName, restoreType string) {
	notification := &Notification{
		ID:      fmt.Sprintf("restore_start_%d", time.Now().Unix()),
		Type:    "restore",
		Level:   LevelInfo,
		Title:   "♻️ Відновлення розпочато",
		Message: fmt.Sprintf("Відновлення '%s' (%s) розпочато", jobName, restoreType),
		Data: map[string]interface{}{
			"job_name":     jobName,
			"restore_type": restoreType,
			"event":        "restore_start",
		},
	}
	e.Send(notification)
}

// SendRestoreSuccess sends restore success notification
func (e *NotificationEngine) SendRestoreSuccess(jobName string, filesRestored int) {
	notification := &Notification{
		ID:      fmt.Sprintf("restore_success_%d", time.Now().Unix()),
		Type:    "restore",
		Level:   LevelSuccess,
		Title:   "✅ Відновлення успішно завершено",
		Message: fmt.Sprintf("Відновлено %d файлів для '%s'", filesRestored, jobName),
		Data: map[string]interface{}{
			"job_name":       jobName,
			"files_restored": filesRestored,
			"event":          "restore_success",
		},
	}
	e.Send(notification)
}

// sendEmail sends notification via email
func (e *NotificationEngine) sendEmail(config *NotificationConfig, notification *Notification) {
	// Create email message
	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(notification.Level), notification.Title)

	body := fmt.Sprintf(`
<html>
<head>
<style>
	body { font-family: Arial, sans-serif; background: #f5f5f5; padding: 20px; }
	.container { background: white; padding: 20px; border-radius: 8px; max-width: 600px; margin: 0 auto; }
	.header { padding: 15px; border-radius: 4px; margin-bottom: 20px; }
	.info { background: #e3f2fd; color: #1565c0; }
	.success { background: #e8f5e9; color: #2e7d32; }
	.warning { background: #fff3e0; color: #ef6c00; }
	.error { background: #ffebee; color: #c62828; }
	.content { color: #333; line-height: 1.6; }
	.footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #eee; color: #999; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
	<div class="header %s">
		<h2>%s</h2>
	</div>
	<div class="content">
		<p>%s</p>
		<p><strong>Час:</strong> %s</p>
		<p><strong>Рівень:</strong> %s</p>
	</div>
	<div class="footer">
		NovaBackup Enterprise v7.0<br>
		Це автоматичне сповіщення від системи резервного копіювання
	</div>
</div>
</body>
</html>
`,
		getEmailClass(notification.Level),
		notification.Title,
		notification.Message,
		notification.Timestamp.Format("02.01.2006 15:04:05"),
		notification.Level,
	)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		config.FromEmail,
		strings.Join(config.Recipients, ", "),
		subject,
		body,
	))

	// Send email
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPass, config.SMTPServer)
	addr := fmt.Sprintf("%s:%d", config.SMTPServer, config.SMTPPort)

	err := smtp.SendMail(addr, auth, config.FromEmail, config.Recipients, msg)
	if err != nil {
		fmt.Printf("❌ Помилка відправки email: %v\n", err)
		e.failed++
	} else {
		fmt.Printf("✅ Email сповіщення відправлено\n")
		e.sent++
	}
}

// sendTelegram sends notification via Telegram
func (e *NotificationEngine) sendTelegram(config *NotificationConfig, notification *Notification) {
	// Format message with emoji based on level
	emoji := getEmoji(notification.Level)

	message := fmt.Sprintf("%s *%s*\n\n%s\n\n_Час: %s_",
		emoji,
		notification.Title,
		notification.Message,
		notification.Timestamp.Format("02.01.2006 15:04:05"),
	)

	// Telegram API URL
	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		config.TelegramBotToken,
	)

	// Prepare request
	data := url.Values{}
	data.Set("chat_id", config.TelegramChatID)
	data.Set("text", message)
	data.Set("parse_mode", "Markdown")

	// Send request
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		fmt.Printf("❌ Помилка відправки Telegram: %v\n", err)
		e.failed++
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ Telegram сповіщення відправлено\n")
		e.sent++
	} else {
		fmt.Printf("❌ Telegram повернув статус: %d\n", resp.StatusCode)
		e.failed++
	}
}

// sendWebhook sends notification via webhook
func (e *NotificationEngine) sendWebhook(config *NotificationConfig, notification *Notification) {
	// Prepare payload
	payload := map[string]interface{}{
		"id":        notification.ID,
		"type":      notification.Type,
		"level":     notification.Level,
		"title":     notification.Title,
		"message":   notification.Message,
		"timestamp": notification.Timestamp,
		"data":      notification.Data,
		"source":    "NovaBackup Enterprise",
	}

	jsonData, _ := json.Marshal(payload)

	// Create request
	req, err := http.NewRequest(config.WebhookMethod, config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Помилка створення webhook запиту: %v\n", err)
		e.failed++
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Add secret header if configured
	if config.WebhookSecret != "" {
		req.Header.Set("X-NovaBackup-Secret", config.WebhookSecret)
	}

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Помилка відправки webhook: %v\n", err)
		e.failed++
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ Webhook сповіщення відправлено\n")
		e.sent++
	} else {
		fmt.Printf("❌ Webhook повернув статус: %d\n", resp.StatusCode)
		e.failed++
	}
}

// sendSlack sends notification to Slack
func (e *NotificationEngine) sendSlack(config *NotificationConfig, notification *Notification) {
	// Slack color based on level
	color := "#36a64f" // green
	switch notification.Level {
	case LevelWarning:
		color = "#ff9800" // orange
	case LevelError:
		color = "#ff0000" // red
	case LevelInfo:
		color = "#2196f3" // blue
	}

	// Slack payload
	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": notification.Title,
				"text":  notification.Message,
				"fields": []map[string]interface{}{
					{
						"title": "Час",
						"value": notification.Timestamp.Format("02.01.2006 15:04:05"),
						"short": true,
					},
					{
						"title": "Рівень",
						"value": notification.Level,
						"short": true,
					},
				},
				"footer": "NovaBackup Enterprise v7.0",
				"ts":     notification.Timestamp.Unix(),
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	// Send to Slack
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Помилка відправки Slack: %v\n", err)
		e.failed++
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ Slack сповіщення відправлено\n")
		e.sent++
	} else {
		fmt.Printf("❌ Slack повернув статус: %d\n", resp.StatusCode)
		e.failed++
	}
}

// Helper functions

func (e *NotificationEngine) validateConfig(config *NotificationConfig) error {
	if config.Name == "" {
		return fmt.Errorf("назва обов'язкова")
	}

	switch config.Type {
	case NotifyEmail:
		if config.SMTPServer == "" || config.FromEmail == "" || len(config.Recipients) == 0 {
			return fmt.Errorf("невірні налаштування email")
		}
	case NotifyTelegram:
		if config.TelegramBotToken == "" || config.TelegramChatID == "" {
			return fmt.Errorf("невірні налаштування Telegram")
		}
	case NotifyWebhook:
		if config.WebhookURL == "" {
			return fmt.Errorf("невірні налаштування webhook")
		}
	}

	return nil
}

func (e *NotificationEngine) shouldSendEvent(config *NotificationConfig, notification *Notification) bool {
	if len(config.Events) == 0 {
		return true // Send all events
	}

	event, exists := notification.Data["event"]
	if !exists {
		return false
	}

	eventStr, ok := event.(string)
	if !ok {
		return false
	}

	for _, e := range config.Events {
		if e == eventStr {
			return true
		}
	}

	return false
}

func (e *NotificationEngine) checkMinLevel(config *NotificationConfig, level string) bool {
	levelOrder := map[string]int{
		LevelInfo:    0,
		LevelWarning: 1,
		LevelError:   2,
		LevelSuccess: 3,
	}

	configLevel, exists := levelOrder[config.MinLevel]
	if !exists {
		return true // Default to all levels
	}

	notificationLevel, exists := levelOrder[level]
	if !exists {
		return false
	}

	return notificationLevel >= configLevel
}

func getEmailClass(level string) string {
	classes := map[string]string{
		LevelInfo:    "info",
		LevelWarning: "warning",
		LevelError:   "error",
		LevelSuccess: "success",
	}
	if class, exists := classes[level]; exists {
		return class
	}
	return "info"
}

func getEmoji(level string) string {
	emojis := map[string]string{
		LevelInfo:    "ℹ️",
		LevelWarning: "⚠️",
		LevelError:   "❌",
		LevelSuccess: "✅",
	}
	if emoji, exists := emojis[level]; exists {
		return emoji
	}
	return "ℹ️"
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetStats returns notification statistics
func (e *NotificationEngine) GetStats() map[string]int {
	return map[string]int{
		"sent":   e.sent,
		"failed": e.failed,
		"total":  e.sent + e.failed,
	}
}
