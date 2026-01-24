package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// DBStore 历史记录存储（数据库实现）
type DBStore struct {
	db *gorm.DB
}

// NewDBStore 创建数据库存储
func NewDBStore(db *gorm.DB) (*DBStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return &DBStore{db: db}, nil
}

// Create 创建执行记录
func (s *DBStore) Create(entry *Entry) error {
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	model := s.toModel(entry)
	return s.db.Create(model).Error
}

// Get 获取执行记录
func (s *DBStore) Get(id string) (*Entry, error) {
	var model database.HistoryModel
	if err := s.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("history entry")
		}
		return nil, err
	}
	return s.fromModel(&model), nil
}

// List 列出执行记录
func (s *DBStore) List(filter *ListFilter) ([]*Entry, error) {
	query := s.db.Model(&database.HistoryModel{})

	if filter != nil {
		if filter.SourceType != "" {
			query = query.Where("source_type = ?", filter.SourceType)
		}
		if filter.SourceID != "" {
			query = query.Where("source_id = ?", filter.SourceID)
		}
		if filter.Engine != "" {
			query = query.Where("engine = ?", filter.Engine)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	query = query.Order("started_at DESC")

	var models []database.HistoryModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*Entry, 0, len(models))
	for i := range models {
		result = append(result, s.fromModel(&models[i]))
	}
	return result, nil
}

// Count 统计执行记录数量
func (s *DBStore) Count(filter *ListFilter) (int, error) {
	query := s.db.Model(&database.HistoryModel{})

	if filter != nil {
		if filter.SourceType != "" {
			query = query.Where("source_type = ?", filter.SourceType)
		}
		if filter.SourceID != "" {
			query = query.Where("source_id = ?", filter.SourceID)
		}
		if filter.Engine != "" {
			query = query.Where("engine = ?", filter.Engine)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// Update 更新执行记录
func (s *DBStore) Update(entry *Entry) error {
	model := s.toModel(entry)
	model.UpdatedAt = time.Now()
	return s.db.Save(model).Error
}

// Delete 删除执行记录
func (s *DBStore) Delete(id string) error {
	return s.db.Delete(&database.HistoryModel{}, "id = ?", id).Error
}

// GetStats 获取统计信息
func (s *DBStore) GetStats(filter *ListFilter) (*Stats, error) {
	entries, err := s.List(filter)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		BySource: make(map[string]int),
		ByEngine: make(map[string]int),
	}

	for _, entry := range entries {
		stats.TotalExecutions++
		switch entry.Status {
		case StatusCompleted:
			stats.CompletedCount++
		case StatusFailed:
			stats.FailedCount++
		}

		if entry.Usage != nil {
			stats.TotalInputTokens += entry.Usage.InputTokens
			stats.TotalOutputTokens += entry.Usage.OutputTokens
		}

		stats.BySource[string(entry.SourceType)]++
		if entry.Engine != "" {
			stats.ByEngine[entry.Engine]++
		}
	}

	return stats, nil
}

func (s *DBStore) toModel(entry *Entry) *database.HistoryModel {
	usageJSON, _ := json.Marshal(entry.Usage)
	metadataJSON, _ := json.Marshal(entry.Metadata)

	return &database.HistoryModel{
		BaseModel: database.BaseModel{
			ID:        entry.ID,
			CreatedAt: entry.StartedAt,
			UpdatedAt: time.Now(),
		},
		SourceType: string(entry.SourceType),
		SourceID:   entry.SourceID,
		SourceName: entry.SourceName,
		Engine:     entry.Engine,
		Status:     string(entry.Status),
		Prompt:     entry.Prompt,
		Output:     entry.Output,
		Error:      entry.Error,
		ExitCode:   entry.ExitCode,
		Usage:      string(usageJSON),
		Metadata:   string(metadataJSON),
		StartedAt:  entry.StartedAt,
		EndedAt:    entry.EndedAt,
	}
}

func (s *DBStore) fromModel(model *database.HistoryModel) *Entry {
	var usage *UsageInfo
	if model.Usage != "" {
		var parsed UsageInfo
		if err := json.Unmarshal([]byte(model.Usage), &parsed); err == nil {
			usage = &parsed
		}
	}

	var metadata map[string]string
	if model.Metadata != "" {
		_ = json.Unmarshal([]byte(model.Metadata), &metadata)
	}

	return &Entry{
		ID:         model.ID,
		SourceType: SourceType(model.SourceType),
		SourceID:   model.SourceID,
		SourceName: model.SourceName,
		Engine:     model.Engine,
		Prompt:     model.Prompt,
		Status:     Status(model.Status),
		Output:     model.Output,
		Error:      model.Error,
		ExitCode:   model.ExitCode,
		Usage:      usage,
		Metadata:   metadata,
		StartedAt:  model.StartedAt,
		EndedAt:    model.EndedAt,
	}
}
