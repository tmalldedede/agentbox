// Package dingtalk 钉钉配置存储
package dingtalk

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// Store 钉钉配置存储
type Store struct {
	db *gorm.DB
}

// NewStore 创建存储
func NewStore() *Store {
	db := database.GetDB()

	// 自动迁移
	if err := db.AutoMigrate(&DingtalkConfigModel{}); err != nil {
		log.Error("auto migrate dingtalk_configs failed", "error", err)
	}

	return &Store{db: db}
}

// DingtalkConfigModel 钉钉配置模型
type DingtalkConfigModel struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	Name           string    `gorm:"size:255" json:"name"`
	AppKey         string    `gorm:"size:128" json:"app_key"`
	AppSecret      string    `gorm:"size:255" json:"app_secret"`
	AgentID        int64     `gorm:"" json:"agent_id"`
	RobotCode      string    `gorm:"size:128" json:"robot_code"`
	DefaultAgentID string    `gorm:"size:64" json:"default_agent_id"`
	Enabled        bool      `gorm:"default:false" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 返回表名
func (DingtalkConfigModel) TableName() string {
	return "dingtalk_configs"
}

// Save 保存配置
func (s *Store) Save(cfg *Config, id string, enabled bool) error {
	model := &DingtalkConfigModel{
		ID:             id,
		Name:           cfg.Name,
		AppKey:         cfg.AppKey,
		AppSecret:      cfg.AppSecret,
		AgentID:        cfg.AgentID,
		RobotCode:      cfg.RobotCode,
		DefaultAgentID: cfg.DefaultAgentID,
		Enabled:        enabled,
		UpdatedAt:      time.Now(),
	}

	return s.db.Save(model).Error
}

// Get 获取配置
func (s *Store) Get(id string) (*Config, bool, error) {
	var model DingtalkConfigModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &Config{
		Name:           model.Name,
		AppKey:         model.AppKey,
		AppSecret:      model.AppSecret,
		AgentID:        model.AgentID,
		RobotCode:      model.RobotCode,
		DefaultAgentID: model.DefaultAgentID,
	}, model.Enabled, nil
}

// GetEnabledConfig 获取启用的配置
func (s *Store) GetEnabledConfig() (*Config, error) {
	var model DingtalkConfigModel
	if err := s.db.First(&model, "enabled = ?", true).Error; err != nil {
		return nil, err
	}

	return &Config{
		Name:           model.Name,
		AppKey:         model.AppKey,
		AppSecret:      model.AppSecret,
		AgentID:        model.AgentID,
		RobotCode:      model.RobotCode,
		DefaultAgentID: model.DefaultAgentID,
	}, nil
}

// List 列出所有配置
func (s *Store) List() ([]map[string]interface{}, error) {
	var models []DingtalkConfigModel
	if err := s.db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		result[i] = map[string]interface{}{
			"id":               m.ID,
			"name":             m.Name,
			"app_key":          m.AppKey,
			"agent_id":         m.AgentID,
			"robot_code":       m.RobotCode,
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
	return s.db.Delete(&DingtalkConfigModel{}, "id = ?", id).Error
}

// SetEnabled 设置启用状态
func (s *Store) SetEnabled(id string, enabled bool) error {
	// 如果启用，先禁用其他配置
	if enabled {
		if err := s.db.Model(&DingtalkConfigModel{}).Where("id != ?", id).Update("enabled", false).Error; err != nil {
			return err
		}
	}
	return s.db.Model(&DingtalkConfigModel{}).Where("id = ?", id).Update("enabled", enabled).Error
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
