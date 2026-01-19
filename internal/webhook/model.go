package webhook

import (
	"time"
)

// Webhook 表示一个 Webhook 配置
type Webhook struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret,omitempty"` // 用于签名验证
	Events    []string  `json:"events"`           // 订阅的事件类型
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event 类型
const (
	EventTaskCreated   = "task.created"
	EventTaskCompleted = "task.completed"
	EventTaskFailed    = "task.failed"
	EventSessionStart  = "session.started"
	EventSessionStop   = "session.stopped"
)

// AllEvents 所有支持的事件
var AllEvents = []string{
	EventTaskCreated,
	EventTaskCompleted,
	EventTaskFailed,
	EventSessionStart,
	EventSessionStop,
}

// WebhookPayload Webhook 推送的数据结构
type WebhookPayload struct {
	ID        string      `json:"id"`
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// CreateWebhookRequest 创建 Webhook 请求
type CreateWebhookRequest struct {
	URL    string   `json:"url" binding:"required"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events,omitempty"` // 为空则订阅所有事件
}

// UpdateWebhookRequest 更新 Webhook 请求
type UpdateWebhookRequest struct {
	URL      string   `json:"url,omitempty"`
	Secret   string   `json:"secret,omitempty"`
	Events   []string `json:"events,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
