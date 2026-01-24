package settings

import (
	"encoding/json"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Manager 配置管理器
type Manager struct {
	db       *gorm.DB
	settings *Settings
	mu       sync.RWMutex
}

// NewManager 创建配置管理器
func NewManager(db *gorm.DB) (*Manager, error) {
	m := &Manager{
		db:       db,
		settings: Default(),
	}

	// 确保表存在
	if err := db.AutoMigrate(&SettingItem{}); err != nil {
		return nil, err
	}

	// 加载配置
	if err := m.load(); err != nil {
		return nil, err
	}

	return m, nil
}

// Get 获取当前配置
func (m *Manager) Get() *Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	copy := *m.settings
	return &copy
}

// GetAgent 获取 Agent 配置
func (m *Manager) GetAgent() AgentSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Agent
}

// GetTask 获取 Task 配置
func (m *Manager) GetTask() TaskSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Task
}

// GetBatch 获取 Batch 配置
func (m *Manager) GetBatch() BatchSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Batch
}

// GetStorage 获取 Storage 配置
func (m *Manager) GetStorage() StorageSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Storage
}

// GetNotify 获取 Notify 配置
func (m *Manager) GetNotify() NotifySettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.Notify
}

// Update 更新配置
func (m *Manager) Update(settings *Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 保存到数据库
	if err := m.save(settings); err != nil {
		return err
	}

	// 更新内存
	m.settings = settings
	return nil
}

// UpdateAgent 更新 Agent 配置
func (m *Manager) UpdateAgent(agent AgentSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.Agent = agent
	return m.saveKey("agent", agent)
}

// UpdateTask 更新 Task 配置
func (m *Manager) UpdateTask(task TaskSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.Task = task
	return m.saveKey("task", task)
}

// UpdateBatch 更新 Batch 配置
func (m *Manager) UpdateBatch(batch BatchSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.Batch = batch
	return m.saveKey("batch", batch)
}

// UpdateStorage 更新 Storage 配置
func (m *Manager) UpdateStorage(storage StorageSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.Storage = storage
	return m.saveKey("storage", storage)
}

// UpdateNotify 更新 Notify 配置
func (m *Manager) UpdateNotify(notify NotifySettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings.Notify = notify
	return m.saveKey("notify", notify)
}

// Reset 重置为默认配置
func (m *Manager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 删除所有配置
	if err := m.db.Where("1 = 1").Delete(&SettingItem{}).Error; err != nil {
		return err
	}

	// 重置内存
	m.settings = Default()
	return nil
}

// load 从数据库加载配置
func (m *Manager) load() error {
	var items []SettingItem
	if err := m.db.Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		switch item.Key {
		case "agent":
			if err := json.Unmarshal([]byte(item.Value), &m.settings.Agent); err != nil {
				continue
			}
		case "task":
			if err := json.Unmarshal([]byte(item.Value), &m.settings.Task); err != nil {
				continue
			}
		case "batch":
			if err := json.Unmarshal([]byte(item.Value), &m.settings.Batch); err != nil {
				continue
			}
		case "storage":
			if err := json.Unmarshal([]byte(item.Value), &m.settings.Storage); err != nil {
				continue
			}
		case "notify":
			if err := json.Unmarshal([]byte(item.Value), &m.settings.Notify); err != nil {
				continue
			}
		}
	}

	return nil
}

// save 保存整个配置到数据库
func (m *Manager) save(settings *Settings) error {
	items := []struct {
		key   string
		value interface{}
	}{
		{"agent", settings.Agent},
		{"task", settings.Task},
		{"batch", settings.Batch},
		{"storage", settings.Storage},
		{"notify", settings.Notify},
	}

	for _, item := range items {
		if err := m.saveKey(item.key, item.value); err != nil {
			return err
		}
	}

	return nil
}

// saveKey 保存单个配置项
func (m *Manager) saveKey(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	item := SettingItem{
		Key:       key,
		Value:     string(data),
		UpdatedAt: time.Now(),
	}

	return m.db.Save(&item).Error
}
