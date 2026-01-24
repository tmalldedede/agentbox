package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("webhook")
}

// Manager Webhook 管理器
type Manager struct {
	store  Store
	client *http.Client
}

// NewManager 创建 Webhook 管理器（使用数据库存储）
func NewManager() *Manager {
	return &Manager{
		store: NewDBStore(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewManagerWithStore 创建 Webhook 管理器（自定义存储，用于测试）
func NewManagerWithStore(store Store) *Manager {
	return &Manager{
		store: store,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Create 创建 Webhook
func (m *Manager) Create(req *CreateWebhookRequest) (*Webhook, error) {
	w := &Webhook{
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   req.Events,
		IsActive: true,
	}

	if err := m.store.Create(w); err != nil {
		return nil, err
	}

	return w, nil
}

// Get 获取 Webhook
func (m *Manager) Get(id string) (*Webhook, error) {
	return m.store.Get(id)
}

// Update 更新 Webhook
func (m *Manager) Update(id string, req *UpdateWebhookRequest) (*Webhook, error) {
	w, err := m.store.Get(id)
	if err != nil {
		return nil, err
	}

	if req.URL != "" {
		w.URL = req.URL
	}
	if req.Secret != "" {
		w.Secret = req.Secret
	}
	if req.Events != nil {
		w.Events = req.Events
	}
	if req.IsActive != nil {
		w.IsActive = *req.IsActive
	}

	if err := m.store.Update(w); err != nil {
		return nil, err
	}

	return w, nil
}

// Delete 删除 Webhook
func (m *Manager) Delete(id string) error {
	return m.store.Delete(id)
}

// List 列出所有 Webhook
func (m *Manager) List() ([]*Webhook, error) {
	return m.store.List()
}

// Send 发送 Webhook 通知
func (m *Manager) Send(event string, data interface{}) {
	webhooks, err := m.store.ListByEvent(event)
	if err != nil {
		log.Error("failed to list webhooks", "event", event, "error", err)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	payload := &WebhookPayload{
		ID:        uuid.New().String(),
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal payload", "error", err)
		return
	}

	for _, w := range webhooks {
		go m.sendToWebhook(w, payloadBytes)
	}
}

// sendToWebhook 发送到单个 Webhook
func (m *Manager) sendToWebhook(w *Webhook, payload []byte) {
	req, err := http.NewRequest("POST", w.URL, bytes.NewReader(payload))
	if err != nil {
		log.Error("failed to create request", "webhook_id", w.ID, "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-ID", w.ID)

	// 如果配置了 Secret，添加签名
	if w.Secret != "" {
		signature := m.sign(payload, w.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		log.Error("failed to send webhook", "webhook_id", w.ID, "url", w.URL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Debug("webhook sent", "webhook_id", w.ID)
	} else {
		log.Warn("webhook returned non-2xx status", "webhook_id", w.ID, "status", resp.StatusCode)
	}
}

// sign 使用 HMAC-SHA256 签名
func (m *Manager) sign(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
