// Package channel 通道会话持久化存储
package channel

import (
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// SessionFilter 会话过滤器
type SessionFilter struct {
	ChannelType string
	Status      string
	AgentID     string
	Limit       int
	Offset      int
}

// SessionStore 会话存储接口
type SessionStore interface {
	Create(session *ChannelSessionData) error
	GetByID(id string) (*ChannelSessionData, error)
	GetByKey(channelType, chatID, userID string, isGroup bool) (*ChannelSessionData, error)
	Update(session *ChannelSessionData) error
	UpdateStatus(id, status string) error
	IncrementMessageCount(id string) error
	List(filter *SessionFilter) ([]*ChannelSessionData, int64, error)
	Delete(id string) error
	DeleteByTaskID(taskID string) error
	GetStats() (*SessionStats, error)
	GetStatsByChannel(channelType string) (*SessionStats, error)
}

// ChannelSessionData 会话数据（用于存储层）
type ChannelSessionData struct {
	ID            string
	ChannelType   string
	ChatID        string
	UserID        string
	UserName      string
	IsGroup       bool
	TaskID        string
	AgentID       string
	AgentName     string
	Status        string // active, completed, expired
	MessageCount  int
	LastMessageAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SessionStats 会话统计
type SessionStats struct {
	TotalSessions   int64                  `json:"total_sessions"`
	ActiveSessions  int64                  `json:"active_sessions"`
	TotalMessages   int64                  `json:"total_messages"`
	MessagesToday   int64                  `json:"messages_today"`
	ByChannel       map[string]ChannelStat `json:"by_channel"`
}

// ChannelStat 单通道统计
type ChannelStat struct {
	Sessions int64 `json:"sessions"`
	Messages int64 `json:"messages"`
}

// GormSessionStore GORM 实现的会话存储
type GormSessionStore struct {
	db *gorm.DB
}

// NewGormSessionStore 创建 GORM 会话存储
func NewGormSessionStore() *GormSessionStore {
	return &GormSessionStore{
		db: database.GetDB(),
	}
}

// Create 创建会话
func (s *GormSessionStore) Create(session *ChannelSessionData) error {
	now := time.Now()
	model := &database.ChannelSessionModel{
		ChannelType:   session.ChannelType,
		ChatID:        session.ChatID,
		UserID:        session.UserID,
		UserName:      session.UserName,
		IsGroup:       session.IsGroup,
		TaskID:        session.TaskID,
		AgentID:       session.AgentID,
		AgentName:     session.AgentName,
		Status:        "active",
		MessageCount:  0,
		LastMessageAt: &now,
	}
	model.ID = session.ID
	model.CreatedAt = now
	model.UpdatedAt = now

	if err := s.db.Create(model).Error; err != nil {
		return err
	}

	session.CreatedAt = model.CreatedAt
	session.UpdatedAt = model.UpdatedAt
	session.Status = model.Status
	return nil
}

// GetByID 根据 ID 获取会话
func (s *GormSessionStore) GetByID(id string) (*ChannelSessionData, error) {
	var model database.ChannelSessionModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return s.modelToData(&model), nil
}

// GetByKey 根据会话键获取活跃会话
func (s *GormSessionStore) GetByKey(channelType, chatID, userID string, isGroup bool) (*ChannelSessionData, error) {
	var model database.ChannelSessionModel
	query := s.db.Where("channel_type = ? AND chat_id = ? AND status = ?", channelType, chatID, "active")

	if isGroup {
		query = query.Where("user_id = ? AND is_group = ?", userID, true)
	} else {
		query = query.Where("is_group = ?", false)
	}

	if err := query.Order("created_at DESC").First(&model).Error; err != nil {
		return nil, err
	}
	return s.modelToData(&model), nil
}

// Update 更新会话
func (s *GormSessionStore) Update(session *ChannelSessionData) error {
	return s.db.Model(&database.ChannelSessionModel{}).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"status":          session.Status,
			"message_count":   session.MessageCount,
			"last_message_at": session.LastMessageAt,
			"updated_at":      time.Now(),
		}).Error
}

// UpdateStatus 更新会话状态
func (s *GormSessionStore) UpdateStatus(id, status string) error {
	return s.db.Model(&database.ChannelSessionModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// IncrementMessageCount 增加消息计数
func (s *GormSessionStore) IncrementMessageCount(id string) error {
	now := time.Now()
	return s.db.Model(&database.ChannelSessionModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"message_count":   gorm.Expr("message_count + 1"),
			"last_message_at": now,
			"updated_at":      now,
		}).Error
}

// List 列出会话
func (s *GormSessionStore) List(filter *SessionFilter) ([]*ChannelSessionData, int64, error) {
	query := s.db.Model(&database.ChannelSessionModel{})

	if filter != nil {
		if filter.ChannelType != "" {
			query = query.Where("channel_type = ?", filter.ChannelType)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.AgentID != "" {
			query = query.Where("agent_id = ?", filter.AgentID)
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

	var models []database.ChannelSessionModel
	if err := query.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	sessions := make([]*ChannelSessionData, len(models))
	for i, model := range models {
		sessions[i] = s.modelToData(&model)
	}
	return sessions, total, nil
}

// Delete 删除会话
func (s *GormSessionStore) Delete(id string) error {
	return s.db.Delete(&database.ChannelSessionModel{}, "id = ?", id).Error
}

// DeleteByTaskID 根据 TaskID 删除会话
func (s *GormSessionStore) DeleteByTaskID(taskID string) error {
	return s.db.Model(&database.ChannelSessionModel{}).
		Where("task_id = ?", taskID).
		Update("status", "completed").Error
}

// GetStats 获取全局统计
func (s *GormSessionStore) GetStats() (*SessionStats, error) {
	stats := &SessionStats{
		ByChannel: make(map[string]ChannelStat),
	}

	// 总会话数
	s.db.Model(&database.ChannelSessionModel{}).Count(&stats.TotalSessions)

	// 活跃会话数
	s.db.Model(&database.ChannelSessionModel{}).Where("status = ?", "active").Count(&stats.ActiveSessions)

	// 总消息数
	s.db.Model(&database.ChannelMessageModel{}).Count(&stats.TotalMessages)

	// 今日消息数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&database.ChannelMessageModel{}).Where("created_at >= ?", today).Count(&stats.MessagesToday)

	// 按通道统计
	type channelStat struct {
		ChannelType string
		Count       int64
	}

	var sessionsByChannel []channelStat
	s.db.Model(&database.ChannelSessionModel{}).
		Select("channel_type, count(*) as count").
		Group("channel_type").
		Scan(&sessionsByChannel)

	var msgsByChannel []channelStat
	s.db.Model(&database.ChannelMessageModel{}).
		Select("channel_type, count(*) as count").
		Group("channel_type").
		Scan(&msgsByChannel)

	msgMap := make(map[string]int64)
	for _, m := range msgsByChannel {
		msgMap[m.ChannelType] = m.Count
	}

	for _, sc := range sessionsByChannel {
		stats.ByChannel[sc.ChannelType] = ChannelStat{
			Sessions: sc.Count,
			Messages: msgMap[sc.ChannelType],
		}
	}

	return stats, nil
}

// GetStatsByChannel 获取指定通道的统计
func (s *GormSessionStore) GetStatsByChannel(channelType string) (*SessionStats, error) {
	stats := &SessionStats{
		ByChannel: make(map[string]ChannelStat),
	}

	// 通道会话数
	s.db.Model(&database.ChannelSessionModel{}).Where("channel_type = ?", channelType).Count(&stats.TotalSessions)

	// 活跃会话数
	s.db.Model(&database.ChannelSessionModel{}).Where("channel_type = ? AND status = ?", channelType, "active").Count(&stats.ActiveSessions)

	// 通道消息数
	s.db.Model(&database.ChannelMessageModel{}).Where("channel_type = ?", channelType).Count(&stats.TotalMessages)

	// 今日消息数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&database.ChannelMessageModel{}).Where("channel_type = ? AND created_at >= ?", channelType, today).Count(&stats.MessagesToday)

	stats.ByChannel[channelType] = ChannelStat{
		Sessions: stats.TotalSessions,
		Messages: stats.TotalMessages,
	}

	return stats, nil
}

// modelToData 将模型转换为数据
func (s *GormSessionStore) modelToData(model *database.ChannelSessionModel) *ChannelSessionData {
	return &ChannelSessionData{
		ID:            model.ID,
		ChannelType:   model.ChannelType,
		ChatID:        model.ChatID,
		UserID:        model.UserID,
		UserName:      model.UserName,
		IsGroup:       model.IsGroup,
		TaskID:        model.TaskID,
		AgentID:       model.AgentID,
		AgentName:     model.AgentName,
		Status:        model.Status,
		MessageCount:  model.MessageCount,
		LastMessageAt: model.LastMessageAt,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
