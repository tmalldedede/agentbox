package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// GormStore GORM 存储实现
type GormStore struct {
	db *gorm.DB
}

// NewGormStore 创建 GORM 存储
func NewGormStore(db *gorm.DB) (*GormStore, error) {
	// AutoMigrate
	if err := db.AutoMigrate(&database.TaskModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate task table: %w", err)
	}
	return &GormStore{db: db}, nil
}

// Create 创建任务
func (s *GormStore) Create(task *Task) error {
	model := taskToModel(task)
	result := s.db.Create(model)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") ||
			strings.Contains(result.Error.Error(), "duplicate key") {
			return ErrTaskExists
		}
		return result.Error
	}
	return nil
}

// Get 获取任务
func (s *GormStore) Get(id string) (*Task, error) {
	var model database.TaskModel
	result := s.db.First(&model, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, result.Error
	}
	return modelToTask(&model), nil
}

// Update 更新任务
func (s *GormStore) Update(task *Task) error {
	model := taskToModel(task)
	result := s.db.Model(&database.TaskModel{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"agent_id":         model.AgentID,
		"agent_name":       model.AgentName,
		"agent_type":       model.AgentType,
		"prompt":           model.Prompt,
		"attachments_json": model.AttachmentsJSON,
		"output_files_json": model.OutputFilesJSON,
		"turns_json":       model.TurnsJSON,
		"turn_count":       model.TurnCount,
		"webhook_url":      model.WebhookURL,
		"timeout":          model.Timeout,
		"status":           model.Status,
		"session_id":       model.SessionID,
		"thread_id":        model.ThreadID,
		"error_message":    model.ErrorMessage,
		"result_json":      model.ResultJSON,
		"metadata_json":    model.MetadataJSON,
		"queued_at":        model.QueuedAt,
		"started_at":       model.StartedAt,
		"completed_at":     model.CompletedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// Delete 删除任务
func (s *GormStore) Delete(id string) error {
	result := s.db.Delete(&database.TaskModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// List 列出任务
func (s *GormStore) List(filter *ListFilter) ([]*Task, error) {
	var models []database.TaskModel
	query := s.buildQuery(filter)

	// 排序
	orderBy := "created_at"
	orderDir := "DESC"
	if filter != nil {
		if filter.OrderBy != "" {
			orderBy = filter.OrderBy
		}
		if !filter.OrderDesc {
			orderDir = "ASC"
		}
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// 分页
	if filter != nil && filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(models))
	for i, model := range models {
		tasks[i] = modelToTask(&model)
	}
	return tasks, nil
}

// Count 统计任务数量
func (s *GormStore) Count(filter *ListFilter) (int, error) {
	var count int64
	query := s.buildQuery(filter)
	if err := query.Model(&database.TaskModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// Stats 获取任务统计
func (s *GormStore) Stats() (*TaskStats, error) {
	stats := &TaskStats{
		ByStatus: make(map[Status]int),
		ByAgent:  make(map[string]int),
	}

	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := s.db.Model(&database.TaskModel{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range statusCounts {
		stats.ByStatus[Status(sc.Status)] = sc.Count
		stats.Total += sc.Count
	}

	// 按 Agent 统计（Top 10）
	var agentCounts []struct {
		AgentName string
		Count     int
	}
	if err := s.db.Model(&database.TaskModel{}).
		Select("COALESCE(agent_name, agent_id) as agent_name, count(*) as count").
		Group("agent_id").
		Order("count DESC").
		Limit(10).
		Scan(&agentCounts).Error; err != nil {
		return nil, err
	}
	for _, ac := range agentCounts {
		stats.ByAgent[ac.AgentName] = ac.Count
	}

	// 平均执行时长（使用 Go 计算，避免 SQL 方言差异）
	var completedTasks []struct {
		StartedAt   time.Time
		CompletedAt time.Time
	}
	if err := s.db.Model(&database.TaskModel{}).
		Select("started_at, completed_at").
		Where("status = ? AND started_at IS NOT NULL AND completed_at IS NOT NULL", "completed").
		Limit(1000). // 限制数量避免内存问题
		Scan(&completedTasks).Error; err == nil && len(completedTasks) > 0 {
		var totalDuration float64
		for _, t := range completedTasks {
			totalDuration += t.CompletedAt.Sub(t.StartedAt).Seconds()
		}
		stats.AvgDuration = totalDuration / float64(len(completedTasks))
	}

	return stats, nil
}

// Cleanup 清理旧任务
func (s *GormStore) Cleanup(before time.Time, statuses []Status) (int, error) {
	query := s.db.Where("created_at < ?", before)

	if len(statuses) > 0 {
		statusStrs := make([]string, len(statuses))
		for i, st := range statuses {
			statusStrs[i] = string(st)
		}
		query = query.Where("status IN ?", statusStrs)
	}

	result := query.Delete(&database.TaskModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

// ClaimQueued 原子领取等待中的任务
func (s *GormStore) ClaimQueued(limit int) ([]*Task, error) {
	if limit <= 0 {
		return []*Task{}, nil
	}

	var claimed []*Task
	now := time.Now()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 查询待处理任务
		var models []database.TaskModel
		if err := tx.Where("status = ?", string(StatusQueued)).
			Order("created_at ASC").
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		// 逐个更新状态
		for _, model := range models {
			result := tx.Model(&database.TaskModel{}).
				Where("id = ? AND status = ?", model.ID, string(StatusQueued)).
				Updates(map[string]interface{}{
					"status":     string(StatusRunning),
					"started_at": now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				task := modelToTask(&model)
				task.Status = StatusRunning
				task.StartedAt = &now
				claimed = append(claimed, task)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// Close 关闭存储
func (s *GormStore) Close() error {
	// GORM 由外部管理连接，这里不关闭
	return nil
}

// buildQuery 构建查询条件
func (s *GormStore) buildQuery(filter *ListFilter) *gorm.DB {
	query := s.db.Model(&database.TaskModel{})

	if filter == nil {
		return query
	}

	// 用户过滤
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	// 状态过滤
	if len(filter.Status) > 0 {
		statusStrs := make([]string, len(filter.Status))
		for i, st := range filter.Status {
			statusStrs[i] = string(st)
		}
		query = query.Where("status IN ?", statusStrs)
	}

	// Agent 过滤
	if filter.AgentID != "" {
		query = query.Where("agent_id = ?", filter.AgentID)
	}

	// 搜索 prompt 关键字
	if filter.Search != "" {
		query = query.Where("prompt LIKE ?", "%"+filter.Search+"%")
	}

	return query
}

// taskToModel 将 Task 转换为 TaskModel
func taskToModel(task *Task) *database.TaskModel {
	attachmentsJSON, _ := json.Marshal(task.Attachments)
	outputFilesJSON, _ := json.Marshal(task.OutputFiles)
	turnsJSON, _ := json.Marshal(task.Turns)
	resultJSON, _ := json.Marshal(task.Result)
	metadataJSON, _ := json.Marshal(task.Metadata)

	return &database.TaskModel{
		BaseModel: database.BaseModel{
			ID:        task.ID,
			CreatedAt: task.CreatedAt,
		},
		UserID:          task.UserID,
		AgentID:         task.AgentID,
		AgentName:       task.AgentName,
		AgentType:       task.AgentType,
		Prompt:          task.Prompt,
		AttachmentsJSON: string(attachmentsJSON),
		OutputFilesJSON: string(outputFilesJSON),
		TurnsJSON:       string(turnsJSON),
		TurnCount:       task.TurnCount,
		WebhookURL:      task.WebhookURL,
		Timeout:         task.Timeout,
		Status:          string(task.Status),
		SessionID:       task.SessionID,
		ThreadID:        task.ThreadID,
		ErrorMessage:    task.ErrorMessage,
		ResultJSON:      string(resultJSON),
		MetadataJSON:    string(metadataJSON),
		QueuedAt:        task.QueuedAt,
		StartedAt:       task.StartedAt,
		CompletedAt:     task.CompletedAt,
	}
}

// modelToTask 将 TaskModel 转换为 Task
func modelToTask(model *database.TaskModel) *Task {
	task := &Task{
		ID:           model.ID,
		UserID:       model.UserID,
		AgentID:      model.AgentID,
		AgentName:    model.AgentName,
		AgentType:    model.AgentType,
		Prompt:       model.Prompt,
		TurnCount:    model.TurnCount,
		WebhookURL:   model.WebhookURL,
		Timeout:      model.Timeout,
		Status:       Status(model.Status),
		SessionID:    model.SessionID,
		ThreadID:     model.ThreadID,
		ErrorMessage: model.ErrorMessage,
		CreatedAt:    model.CreatedAt,
		QueuedAt:     model.QueuedAt,
		StartedAt:    model.StartedAt,
		CompletedAt:  model.CompletedAt,
	}

	// 解析 JSON 字段
	if model.AttachmentsJSON != "" && model.AttachmentsJSON != "null" {
		json.Unmarshal([]byte(model.AttachmentsJSON), &task.Attachments)
	}
	if model.OutputFilesJSON != "" && model.OutputFilesJSON != "null" {
		json.Unmarshal([]byte(model.OutputFilesJSON), &task.OutputFiles)
	}
	if model.TurnsJSON != "" && model.TurnsJSON != "null" {
		json.Unmarshal([]byte(model.TurnsJSON), &task.Turns)
	}
	if model.ResultJSON != "" && model.ResultJSON != "null" {
		json.Unmarshal([]byte(model.ResultJSON), &task.Result)
	}
	if model.MetadataJSON != "" && model.MetadataJSON != "null" {
		json.Unmarshal([]byte(model.MetadataJSON), &task.Metadata)
	}

	return task
}
