package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Manager Webhook 管理器
type Manager struct {
	store  Store
	client *http.Client
}

// NewManager 创建 Webhook 管理器
func NewManager(dataDir string) (*Manager, error) {
	store, err := NewFileStore(dataDir)
	if err != nil {
		return nil, err
	}

	return &Manager{
		store: store,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
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
		log.Printf("[Webhook] Failed to list webhooks for event %s: %v", event, err)
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
		log.Printf("[Webhook] Failed to marshal payload: %v", err)
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
		log.Printf("[Webhook] Failed to create request for %s: %v", w.ID, err)
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
		log.Printf("[Webhook] Failed to send to %s (%s): %v", w.ID, w.URL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[Webhook] Sent to %s successfully", w.ID)
	} else {
		log.Printf("[Webhook] Failed to send to %s, status: %d", w.ID, resp.StatusCode)
	}
}

// sign 使用 HMAC-SHA256 签名
func (m *Manager) sign(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
