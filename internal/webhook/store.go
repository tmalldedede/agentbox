package webhook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/apperr"
)

// Webhook errors - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrWebhookNotFound = apperr.NotFound("webhook")
	ErrWebhookExists   = apperr.AlreadyExists("webhook")
)

// Store Webhook 存储接口
type Store interface {
	Create(w *Webhook) error
	Get(id string) (*Webhook, error)
	Update(w *Webhook) error
	Delete(id string) error
	List() ([]*Webhook, error)
	ListByEvent(event string) ([]*Webhook, error)
}

// FileStore 基于文件的存储
type FileStore struct {
	dataDir  string
	webhooks map[string]*Webhook
	mu       sync.RWMutex
}

// NewFileStore 创建文件存储
func NewFileStore(dataDir string) (*FileStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	store := &FileStore{
		dataDir:  dataDir,
		webhooks: make(map[string]*Webhook),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// load 从文件加载所有 Webhook
func (s *FileStore) load() error {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dataDir, file.Name()))
		if err != nil {
			continue
		}

		var w Webhook
		if err := json.Unmarshal(data, &w); err != nil {
			continue
		}

		s.webhooks[w.ID] = &w
	}

	return nil
}

// save 保存单个 Webhook 到文件
func (s *FileStore) save(w *Webhook) error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(s.dataDir, w.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// Create 创建 Webhook
func (s *FileStore) Create(w *Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w.ID == "" {
		w.ID = "wh-" + uuid.New().String()[:8]
	}

	if _, exists := s.webhooks[w.ID]; exists {
		return ErrWebhookExists
	}

	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now

	if err := s.save(w); err != nil {
		return err
	}

	s.webhooks[w.ID] = w
	return nil
}

// Get 获取 Webhook
func (s *FileStore) Get(id string) (*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w, ok := s.webhooks[id]
	if !ok {
		return nil, ErrWebhookNotFound
	}

	return w, nil
}

// Update 更新 Webhook
func (s *FileStore) Update(w *Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.webhooks[w.ID]; !exists {
		return ErrWebhookNotFound
	}

	w.UpdatedAt = time.Now()

	if err := s.save(w); err != nil {
		return err
	}

	s.webhooks[w.ID] = w
	return nil
}

// Delete 删除 Webhook
func (s *FileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.webhooks[id]; !exists {
		return ErrWebhookNotFound
	}

	filename := filepath.Join(s.dataDir, id+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(s.webhooks, id)
	return nil
}

// List 列出所有 Webhook
func (s *FileStore) List() ([]*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Webhook, 0, len(s.webhooks))
	for _, w := range s.webhooks {
		result = append(result, w)
	}

	return result, nil
}

// ListByEvent 列出订阅指定事件的 Webhook
func (s *FileStore) ListByEvent(event string) ([]*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Webhook, 0)
	for _, w := range s.webhooks {
		if !w.IsActive {
			continue
		}
		// 如果 Events 为空，表示订阅所有事件
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
