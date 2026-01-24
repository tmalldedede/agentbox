package webhook

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/database"
)

// DBStore 基于数据库的 Webhook 存储
type DBStore struct {
	repo *database.WebhookRepository
}

// NewDBStore 创建数据库存储
func NewDBStore() *DBStore {
	return &DBStore{
		repo: database.NewWebhookRepository(),
	}
}

// Create 创建 Webhook
func (s *DBStore) Create(w *Webhook) error {
	if w.ID == "" {
		w.ID = "wh-" + uuid.New().String()[:8]
	}

	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now

	model := s.toModel(w)
	if err := s.repo.Create(model); err != nil {
		return err
	}

	return nil
}

// Get 获取 Webhook
func (s *DBStore) Get(id string) (*Webhook, error) {
	model, err := s.repo.Get(id)
	if err != nil {
		return nil, ErrWebhookNotFound
	}
	return s.fromModel(model), nil
}

// Update 更新 Webhook
func (s *DBStore) Update(w *Webhook) error {
	// 检查是否存在
	_, err := s.repo.Get(w.ID)
	if err != nil {
		return ErrWebhookNotFound
	}

	w.UpdatedAt = time.Now()
	model := s.toModel(w)
	return s.repo.Update(model)
}

// Delete 删除 Webhook
func (s *DBStore) Delete(id string) error {
	_, err := s.repo.Get(id)
	if err != nil {
		return ErrWebhookNotFound
	}
	return s.repo.Delete(id)
}

// List 列出所有 Webhook
func (s *DBStore) List() ([]*Webhook, error) {
	models, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	result := make([]*Webhook, 0, len(models))
	for i := range models {
		result = append(result, s.fromModel(&models[i]))
	}
	return result, nil
}

// ListByEvent 列出订阅指定事件的活跃 Webhook
func (s *DBStore) ListByEvent(event string) ([]*Webhook, error) {
	models, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}

	result := make([]*Webhook, 0)
	for i := range models {
		w := s.fromModel(&models[i])
		// Events 为空表示订阅所有事件
		if len(w.Events) == 0 {
			result = append(result, w)
			continue
		}
		for _, e := range w.Events {
			if e == event {
				result = append(result, w)
				break
			}
		}
	}
	return result, nil
}

// toModel 将领域模型转为数据库模型
func (s *DBStore) toModel(w *Webhook) *database.WebhookModel {
	eventsJSON := ""
	if len(w.Events) > 0 {
		b, _ := json.Marshal(w.Events)
		eventsJSON = string(b)
	}

	// 从 URL 生成默认 Name
	name := w.URL
	if len(name) > 255 {
		name = name[:255]
	}

	return &database.WebhookModel{
		BaseModel: database.BaseModel{
			ID:        w.ID,
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
		},
		Name:      name,
		URL:       w.URL,
		Secret:    w.Secret,
		Events:    eventsJSON,
		IsEnabled: w.IsActive,
	}
}

// fromModel 将数据库模型转为领域模型
func (s *DBStore) fromModel(m *database.WebhookModel) *Webhook {
	var events []string
	if m.Events != "" {
		_ = json.Unmarshal([]byte(m.Events), &events)
	}

	return &Webhook{
		ID:        m.ID,
		URL:       m.URL,
		Secret:    m.Secret,
		Events:    events,
		IsActive:  m.IsEnabled,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
