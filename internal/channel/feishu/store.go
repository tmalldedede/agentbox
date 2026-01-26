package feishu

import (
	"encoding/json"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// ConfigModel 飞书配置数据库模型
type ConfigModel struct {
	ID                string    `gorm:"primaryKey;size:36"`
	Name              string    `gorm:"size:100"`
	AppID             string    `gorm:"size:100;not null"`
	AppSecret         string    `gorm:"size:200;not null"` // 加密存储
	VerificationToken string    `gorm:"size:100"`
	EncryptKey        string    `gorm:"size:200"`
	BotName           string    `gorm:"size:100"`
	DefaultAgentID    string    `gorm:"size:36"`
	Enabled           bool      `gorm:"default:false"`
	Metadata          string    `gorm:"type:text"` // JSON
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName 表名
func (ConfigModel) TableName() string {
	return "feishu_configs"
}

// Store 飞书配置存储
type Store struct {
	db *gorm.DB
}

// NewStore 创建存储
func NewStore() *Store {
	db := database.GetDB()

	// 自动迁移
	if err := db.AutoMigrate(&ConfigModel{}); err != nil {
		log.Error("auto migrate feishu_configs failed", "error", err)
	}

	return &Store{db: db}
}

// SaveConfig 保存配置
func (s *Store) SaveConfig(id string, cfg *Config) error {
	model := &ConfigModel{
		ID:                id,
		Name:              cfg.Name,
		AppID:             cfg.AppID,
		AppSecret:         cfg.AppSecret,
		VerificationToken: cfg.VerificationToken,
		EncryptKey:        cfg.EncryptKey,
		BotName:           cfg.BotName,
		DefaultAgentID:    cfg.DefaultAgentID,
		Enabled:           true,
		UpdatedAt:         time.Now(),
	}

	return s.db.Save(model).Error
}

// GetConfig 获取配置
func (s *Store) GetConfig(id string) (*Config, error) {
	var model ConfigModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &Config{
		Name:              model.Name,
		AppID:             model.AppID,
		AppSecret:         model.AppSecret,
		VerificationToken: model.VerificationToken,
		EncryptKey:        model.EncryptKey,
		BotName:           model.BotName,
		DefaultAgentID:    model.DefaultAgentID,
	}, nil
}

// GetEnabledConfig 获取启用的配置
func (s *Store) GetEnabledConfig() (*Config, error) {
	var model ConfigModel
	if err := s.db.Where("enabled = ?", true).First(&model).Error; err != nil {
		return nil, err
	}

	return &Config{
		Name:              model.Name,
		AppID:             model.AppID,
		AppSecret:         model.AppSecret,
		VerificationToken: model.VerificationToken,
		EncryptKey:        model.EncryptKey,
		BotName:           model.BotName,
		DefaultAgentID:    model.DefaultAgentID,
	}, nil
}

// DeleteConfig 删除配置
func (s *Store) DeleteConfig(id string) error {
	return s.db.Delete(&ConfigModel{}, "id = ?", id).Error
}

// ListConfigs 列出所有配置
func (s *Store) ListConfigs() ([]*ConfigModel, error) {
	var models []*ConfigModel
	if err := s.db.Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

// MessageLog 消息日志
type MessageLog struct {
	ID          string    `gorm:"primaryKey;size:100"`
	ChatID      string    `gorm:"size:100;index"`
	SenderID    string    `gorm:"size:100"`
	SenderName  string    `gorm:"size:100"`
	Content     string    `gorm:"type:text"`
	MessageType string    `gorm:"size:20"`
	ReplyID     string    `gorm:"size:100"`
	TaskID      string    `gorm:"size:36;index"` // 关联的 Task ID
	Metadata    string    `gorm:"type:text"`     // JSON
	ReceivedAt  time.Time `gorm:"index"`
	CreatedAt   time.Time
}

// TableName 表名
func (MessageLog) TableName() string {
	return "feishu_message_logs"
}

// SaveMessageLog 保存消息日志
func (s *Store) SaveMessageLog(msg *MessageLog) error {
	return s.db.Create(msg).Error
}

// GetMessageLogs 获取消息日志
func (s *Store) GetMessageLogs(chatID string, limit int) ([]*MessageLog, error) {
	var logs []*MessageLog
	if err := s.db.Where("chat_id = ?", chatID).Order("received_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// UpdateMessageTaskID 更新消息关联的 Task ID
func (s *Store) UpdateMessageTaskID(msgID, taskID string) error {
	return s.db.Model(&MessageLog{}).Where("id = ?", msgID).Update("task_id", taskID).Error
}

// metadataToJSON 元数据转 JSON
func metadataToJSON(m map[string]string) string {
	if m == nil {
		return ""
	}
	b, _ := json.Marshal(m)
	return string(b)
}
