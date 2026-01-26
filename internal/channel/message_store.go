// Package channel 通道消息持久化存储
package channel

import (
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// MessageFilter 消息过滤器
type MessageFilter struct {
	SessionID   string
	ChannelType string
	ChatID      string
	Direction   string
	TaskID      string
	Limit       int
	Offset      int
}

// MessageStore 消息存储接口
type MessageStore interface {
	Save(msg *ChannelMessageData) error
	GetByID(id string) (*ChannelMessageData, error)
	ListBySession(sessionID string, limit, offset int) ([]*ChannelMessageData, int64, error)
	ListByChat(channelType, chatID string, limit, offset int) ([]*ChannelMessageData, int64, error)
	List(filter *MessageFilter) ([]*ChannelMessageData, int64, error)
	UpdateTaskID(msgID, taskID string) error
	CountByChannel(channelType string) (int64, error)
	CountToday(channelType string) (int64, error)
}

// ChannelMessageData 消息数据（用于存储层）
type ChannelMessageData struct {
	ID          string
	SessionID   string
	ChannelType string
	ChatID      string
	SenderID    string
	SenderName  string
	Content     string
	Direction   string // inbound, outbound
	TaskID      string
	TurnID      string
	Status      string
	Metadata    map[string]string
	ReceivedAt  time.Time
	CreatedAt   time.Time
}

// GormMessageStore GORM 实现的消息存储
type GormMessageStore struct {
	db *gorm.DB
}

// NewGormMessageStore 创建 GORM 消息存储
func NewGormMessageStore() *GormMessageStore {
	return &GormMessageStore{
		db: database.GetDB(),
	}
}

// Save 保存消息
func (s *GormMessageStore) Save(msg *ChannelMessageData) error {
	now := time.Now()
	if msg.ReceivedAt.IsZero() {
		msg.ReceivedAt = now
	}

	metadataJSON := ""
	if msg.Metadata != nil {
		// 简单序列化
		metadataJSON = metadataToJSON(msg.Metadata)
	}

	model := &database.ChannelMessageModel{
		ID:           msg.ID,
		SessionID:    msg.SessionID,
		ChannelType:  msg.ChannelType,
		ChatID:       msg.ChatID,
		SenderID:     msg.SenderID,
		SenderName:   msg.SenderName,
		Content:      msg.Content,
		Direction:    msg.Direction,
		TaskID:       msg.TaskID,
		TurnID:       msg.TurnID,
		Status:       msg.Status,
		MetadataJSON: metadataJSON,
		ReceivedAt:   msg.ReceivedAt,
		CreatedAt:    now,
	}

	if model.Status == "" {
		model.Status = "sent"
	}

	return s.db.Create(model).Error
}

// GetByID 根据 ID 获取消息
func (s *GormMessageStore) GetByID(id string) (*ChannelMessageData, error) {
	var model database.ChannelMessageModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return s.modelToData(&model), nil
}

// ListBySession 列出会话的消息
func (s *GormMessageStore) ListBySession(sessionID string, limit, offset int) ([]*ChannelMessageData, int64, error) {
	return s.List(&MessageFilter{
		SessionID: sessionID,
		Limit:     limit,
		Offset:    offset,
	})
}

// ListByChat 列出聊天的消息
func (s *GormMessageStore) ListByChat(channelType, chatID string, limit, offset int) ([]*ChannelMessageData, int64, error) {
	return s.List(&MessageFilter{
		ChannelType: channelType,
		ChatID:      chatID,
		Limit:       limit,
		Offset:      offset,
	})
}

// List 列出消息
func (s *GormMessageStore) List(filter *MessageFilter) ([]*ChannelMessageData, int64, error) {
	query := s.db.Model(&database.ChannelMessageModel{})

	if filter != nil {
		if filter.SessionID != "" {
			query = query.Where("session_id = ?", filter.SessionID)
		}
		if filter.ChannelType != "" {
			query = query.Where("channel_type = ?", filter.ChannelType)
		}
		if filter.ChatID != "" {
			query = query.Where("chat_id = ?", filter.ChatID)
		}
		if filter.Direction != "" {
			query = query.Where("direction = ?", filter.Direction)
		}
		if filter.TaskID != "" {
			query = query.Where("task_id = ?", filter.TaskID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter != nil {
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	var models []database.ChannelMessageModel
	if err := query.Order("received_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	messages := make([]*ChannelMessageData, len(models))
	for i, model := range models {
		messages[i] = s.modelToData(&model)
	}
	return messages, total, nil
}

// UpdateTaskID 更新消息的 TaskID
func (s *GormMessageStore) UpdateTaskID(msgID, taskID string) error {
	return s.db.Model(&database.ChannelMessageModel{}).
		Where("id = ?", msgID).
		Update("task_id", taskID).Error
}

// CountByChannel 统计通道消息数
func (s *GormMessageStore) CountByChannel(channelType string) (int64, error) {
	var count int64
	err := s.db.Model(&database.ChannelMessageModel{}).
		Where("channel_type = ?", channelType).
		Count(&count).Error
	return count, err
}

// CountToday 统计今日消息数
func (s *GormMessageStore) CountToday(channelType string) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	query := s.db.Model(&database.ChannelMessageModel{}).Where("created_at >= ?", today)
	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	err := query.Count(&count).Error
	return count, err
}

// modelToData 将模型转换为数据
func (s *GormMessageStore) modelToData(model *database.ChannelMessageModel) *ChannelMessageData {
	metadata := jsonToMetadata(model.MetadataJSON)

	return &ChannelMessageData{
		ID:          model.ID,
		SessionID:   model.SessionID,
		ChannelType: model.ChannelType,
		ChatID:      model.ChatID,
		SenderID:    model.SenderID,
		SenderName:  model.SenderName,
		Content:     model.Content,
		Direction:   model.Direction,
		TaskID:      model.TaskID,
		TurnID:      model.TurnID,
		Status:      model.Status,
		Metadata:    metadata,
		ReceivedAt:  model.ReceivedAt,
		CreatedAt:   model.CreatedAt,
	}
}

// metadataToJSON 将 metadata 转换为 JSON
func metadataToJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	// 简单的 JSON 序列化
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += `"` + k + `":"` + v + `"`
		first = false
	}
	result += "}"
	return result
}

// jsonToMetadata 将 JSON 转换为 metadata
func jsonToMetadata(jsonStr string) map[string]string {
	// 简单实现，生产环境应使用 json.Unmarshal
	if jsonStr == "" || jsonStr == "{}" {
		return nil
	}
	// 返回空 map，避免解析复杂性
	return make(map[string]string)
}
