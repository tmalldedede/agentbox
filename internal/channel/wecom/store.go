// Package wecom 企业微信配置存储
package wecom

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// Store 企业微信配置存储
type Store struct {
	db *gorm.DB
}

// NewStore 创建存储
func NewStore() *Store {
	db := database.GetDB()

	// 自动迁移
	if err := db.AutoMigrate(&WecomConfigModel{}); err != nil {
		log.Error("auto migrate wecom_configs failed", "error", err)
	}

	return &Store{db: db}
}

// WecomConfigModel 企业微信配置模型
type WecomConfigModel struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	Name           string    `gorm:"size:255" json:"name"`
	CorpID         string    `gorm:"size:128" json:"corp_id"`
	AgentID        int       `gorm:"" json:"agent_id"`
	Secret         string    `gorm:"size:255" json:"secret"`
	Token          string    `gorm:"size:255" json:"token"`
	EncodingAESKey string    `gorm:"size:255" json:"encoding_aes_key"`
	DefaultAgentID string    `gorm:"size:64" json:"default_agent_id"`
	Enabled        bool      `gorm:"default:false" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 返回表名
func (WecomConfigModel) TableName() string {
	return "wecom_configs"
}

// Save 保存配置
func (s *Store) Save(cfg *Config, id string, enabled bool) error {
	model := &WecomConfigModel{
		ID:             id,
		Name:           cfg.Name,
		CorpID:         cfg.CorpID,
		AgentID:        cfg.AgentID,
		Secret:         cfg.Secret,
		Token:          cfg.Token,
		EncodingAESKey: cfg.EncodingAESKey,
		DefaultAgentID: cfg.DefaultAgentID,
		Enabled:        enabled,
		UpdatedAt:      time.Now(),
	}

	return s.db.Save(model).Error
}

// Get 获取配置
func (s *Store) Get(id string) (*Config, bool, error) {
	var model WecomConfigModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &Config{
		Name:           model.Name,
		CorpID:         model.CorpID,
		AgentID:        model.AgentID,
		Secret:         model.Secret,
		Token:          model.Token,
		EncodingAESKey: model.EncodingAESKey,
		DefaultAgentID: model.DefaultAgentID,
	}, model.Enabled, nil
}

// GetEnabledConfig 获取启用的配置
func (s *Store) GetEnabledConfig() (*Config, error) {
	var model WecomConfigModel
	if err := s.db.First(&model, "enabled = ?", true).Error; err != nil {
		return nil, err
	}

	return &Config{
		Name:           model.Name,
		CorpID:         model.CorpID,
		AgentID:        model.AgentID,
		Secret:         model.Secret,
		Token:          model.Token,
		EncodingAESKey: model.EncodingAESKey,
		DefaultAgentID: model.DefaultAgentID,
	}, nil
}

// List 列出所有配置
func (s *Store) List() ([]map[string]interface{}, error) {
	var models []WecomConfigModel
	if err := s.db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		result[i] = map[string]interface{}{
			"id":               m.ID,
			"name":             m.Name,
			"corp_id":          m.CorpID,
			"agent_id":         m.AgentID,
			"default_agent_id": m.DefaultAgentID,
			"enabled":          m.Enabled,
			"created_at":       m.CreatedAt,
			"updated_at":       m.UpdatedAt,
		}
	}
	return result, nil
}

// Delete 删除配置
func (s *Store) Delete(id string) error {
	return s.db.Delete(&WecomConfigModel{}, "id = ?", id).Error
}

// SetEnabled 设置启用状态
func (s *Store) SetEnabled(id string, enabled bool) error {
	// 如果启用，先禁用其他配置
	if enabled {
		if err := s.db.Model(&WecomConfigModel{}).Where("id != ?", id).Update("enabled", false).Error; err != nil {
			return err
		}
	}
	return s.db.Model(&WecomConfigModel{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ConfigToJSON 配置转 JSON
func ConfigToJSON(cfg *Config) string {
	b, _ := json.Marshal(cfg)
	return string(b)
}

// JSONToConfig JSON 转配置
func JSONToConfig(data string) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
