package database

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

// MCPServerRepository handles MCP server database operations
type MCPServerRepository struct {
	db *gorm.DB
}

// NewMCPServerRepository creates a new MCPServerRepository
func NewMCPServerRepository() *MCPServerRepository {
	return &MCPServerRepository{db: DB}
}

// Create creates a new MCP server
func (r *MCPServerRepository) Create(model *MCPServerModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves an MCP server by ID
func (r *MCPServerRepository) Get(id string) (*MCPServerModel, error) {
	var model MCPServerModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// GetByName retrieves an MCP server by name
func (r *MCPServerRepository) GetByName(name string) (*MCPServerModel, error) {
	var model MCPServerModel
	err := r.db.Where("name = ?", name).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all MCP servers
func (r *MCPServerRepository) List() ([]MCPServerModel, error) {
	var models []MCPServerModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// Update updates an MCP server
func (r *MCPServerRepository) Update(model *MCPServerModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes an MCP server
func (r *MCPServerRepository) Delete(id string) error {
	return r.db.Delete(&MCPServerModel{}, "id = ?", id).Error
}

// SkillRepository handles skill database operations
type SkillRepository struct {
	db *gorm.DB
}

// NewSkillRepository creates a new SkillRepository
func NewSkillRepository() *SkillRepository {
	return &SkillRepository{db: DB}
}

// Create creates a new skill
func (r *SkillRepository) Create(model *SkillModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a skill by ID
func (r *SkillRepository) Get(id string) (*SkillModel, error) {
	var model SkillModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// GetBySlug retrieves a skill by slug
func (r *SkillRepository) GetBySlug(slug string) (*SkillModel, error) {
	var model SkillModel
	err := r.db.Where("slug = ?", slug).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all skills
func (r *SkillRepository) List() ([]SkillModel, error) {
	var models []SkillModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// Update updates a skill
func (r *SkillRepository) Update(model *SkillModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a skill
func (r *SkillRepository) Delete(id string) error {
	return r.db.Delete(&SkillModel{}, "id = ?", id).Error
}

// SessionRepository handles session database operations
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{db: DB}
}

// Create creates a new session
func (r *SessionRepository) Create(model *SessionModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a session by ID
func (r *SessionRepository) Get(id string) (*SessionModel, error) {
	var model SessionModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves sessions with optional status filter
func (r *SessionRepository) List(status string, limit int) ([]SessionModel, error) {
	var models []SessionModel
	query := r.db.Model(&SessionModel{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&models).Error
	return models, err
}

// Update updates a session
func (r *SessionRepository) Update(model *SessionModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a session
func (r *SessionRepository) Delete(id string) error {
	return r.db.Delete(&SessionModel{}, "id = ?", id).Error
}

// UpdateStatus updates session status
func (r *SessionRepository) UpdateStatus(id, status string) error {
	return r.db.Model(&SessionModel{}).Where("id = ?", id).Update("status", status).Error
}

// WebhookRepository handles webhook database operations
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new WebhookRepository
func NewWebhookRepository() *WebhookRepository {
	return &WebhookRepository{db: DB}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(model *WebhookModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a webhook by ID
func (r *WebhookRepository) Get(id string) (*WebhookModel, error) {
	var model WebhookModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all webhooks
func (r *WebhookRepository) List() ([]WebhookModel, error) {
	var models []WebhookModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// ListEnabled retrieves all enabled webhooks
func (r *WebhookRepository) ListEnabled() ([]WebhookModel, error) {
	var models []WebhookModel
	err := r.db.Where("is_enabled = ?", true).Find(&models).Error
	return models, err
}

// Update updates a webhook
func (r *WebhookRepository) Update(model *WebhookModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a webhook
func (r *WebhookRepository) Delete(id string) error {
	return r.db.Delete(&WebhookModel{}, "id = ?", id).Error
}

// Helper functions for JSON serialization
func ToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func FromJSON(s string, v interface{}) error {
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}

// BatchRepository handles batch database operations
type BatchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a new BatchRepository
func NewBatchRepository() *BatchRepository {
	return &BatchRepository{db: DB}
}

// Create creates a new batch
func (r *BatchRepository) Create(model *BatchModel) error {
	if model.ID == "" {
		model.ID = "batch-" + uuid.New().String()[:8]
	}
	return r.db.Create(model).Error
}

// Get retrieves a batch by ID
func (r *BatchRepository) Get(id string) (*BatchModel, error) {
	var model BatchModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves batches with optional filters
func (r *BatchRepository) List(status string, agentID string, userID string, limit, offset int) ([]BatchModel, int64, error) {
	var models []BatchModel
	var total int64

	query := r.db.Model(&BatchModel{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit <= 0 {
		limit = 50
	}
	query = query.Limit(limit).Offset(offset)

	err := query.Order("created_at DESC").Find(&models).Error
	return models, total, err
}

// ListByStatus retrieves batches by status
func (r *BatchRepository) ListByStatus(status string) ([]BatchModel, error) {
	var models []BatchModel
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&models).Error
	return models, err
}

// Update updates a batch
func (r *BatchRepository) Update(model *BatchModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// UpdateCounters atomically updates completed/failed/dead counters
func (r *BatchRepository) UpdateCounters(id string, completed, failed, dead int) error {
	return r.db.Model(&BatchModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"completed":  gorm.Expr("completed + ?", completed),
		"failed":     gorm.Expr("failed + ?", failed),
		"dead":       gorm.Expr("dead + ?", dead),
		"updated_at": time.Now(),
	}).Error
}

// UpdateStatus updates batch status
func (r *BatchRepository) UpdateStatus(id, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.Model(&BatchModel{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a batch and its tasks
func (r *BatchRepository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete tasks first
		if err := tx.Where("batch_id = ?", id).Delete(&BatchTaskModel{}).Error; err != nil {
			return err
		}
		// Delete batch
		return tx.Delete(&BatchModel{}, "id = ?", id).Error
	})
}

// BatchTaskRepository handles batch task database operations
type BatchTaskRepository struct {
	db *gorm.DB
}

// NewBatchTaskRepository creates a new BatchTaskRepository
func NewBatchTaskRepository() *BatchTaskRepository {
	return &BatchTaskRepository{db: DB}
}

// CreateBulk creates multiple tasks in a transaction
func (r *BatchTaskRepository) CreateBulk(models []BatchTaskModel) error {
	if len(models) == 0 {
		return nil
	}
	return r.db.CreateInBatches(models, 100).Error
}

// Get retrieves a task by ID
func (r *BatchTaskRepository) Get(id string) (*BatchTaskModel, error) {
	var model BatchTaskModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// GetByBatchAndID retrieves a task by batch ID and task ID
func (r *BatchTaskRepository) GetByBatchAndID(batchID, taskID string) (*BatchTaskModel, error) {
	var model BatchTaskModel
	err := r.db.Where("batch_id = ? AND id = ?", batchID, taskID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves tasks for a batch with optional filters
func (r *BatchTaskRepository) List(batchID, status, workerID string, limit, offset int) ([]BatchTaskModel, int64, error) {
	var models []BatchTaskModel
	var total int64

	query := r.db.Model(&BatchTaskModel{}).Where("batch_id = ?", batchID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit <= 0 {
		limit = 100
	}
	query = query.Limit(limit).Offset(offset)

	err := query.Order("task_index ASC").Find(&models).Error
	return models, total, err
}

// Update updates a task
func (r *BatchTaskRepository) Update(model *BatchTaskModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// UpdateStatus updates task status and related fields
func (r *BatchTaskRepository) UpdateStatus(id, status string, updates map[string]interface{}) error {
	if updates == nil {
		updates = make(map[string]interface{})
	}
	updates["status"] = status
	updates["updated_at"] = time.Now()
	return r.db.Model(&BatchTaskModel{}).Where("id = ?", id).Updates(updates).Error
}

// MarkDead moves a task to dead letter status
func (r *BatchTaskRepository) MarkDead(id, reason string) error {
	now := time.Now()
	return r.db.Model(&BatchTaskModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "dead",
		"dead_at":     &now,
		"dead_reason": reason,
		"updated_at":  now,
	}).Error
}

// ListDead retrieves dead tasks for a batch
func (r *BatchTaskRepository) ListDead(batchID string, limit int) ([]BatchTaskModel, error) {
	var models []BatchTaskModel
	if limit <= 0 {
		limit = 100
	}
	err := r.db.Where("batch_id = ? AND status = ?", batchID, "dead").
		Order("dead_at DESC").Limit(limit).Find(&models).Error
	return models, err
}

// RetryDead resets dead tasks to pending
func (r *BatchTaskRepository) RetryDead(batchID string, taskIDs []string) (int64, error) {
	query := r.db.Model(&BatchTaskModel{}).Where("batch_id = ? AND status = ?", batchID, "dead")
	if len(taskIDs) > 0 {
		query = query.Where("id IN ?", taskIDs)
	}

	result := query.Updates(map[string]interface{}{
		"status":      "pending",
		"dead_at":     nil,
		"dead_reason": "",
		"error":       "",
		"attempts":    0,
		"worker_id":   "",
		"started_at":  nil,
		"updated_at":  time.Now(),
	})

	return result.RowsAffected, result.Error
}

// ResetRunning resets running tasks to pending (for recovery)
func (r *BatchTaskRepository) ResetRunning(batchID string) (int64, error) {
	result := r.db.Model(&BatchTaskModel{}).
		Where("batch_id = ? AND status = ?", batchID, "running").
		Updates(map[string]interface{}{
			"status":     "pending",
			"worker_id":  "",
			"started_at": nil,
			"claimed_at": nil,
			"claimed_by": "",
			"updated_at": time.Now(),
		})
	return result.RowsAffected, result.Error
}

// GetStats returns task statistics for a batch
func (r *BatchTaskRepository) GetStats(batchID string) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	err := r.db.Model(&BatchTaskModel{}).
		Select("status, COUNT(*) as count").
		Where("batch_id = ?", batchID).
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Status] = r.Count
	}
	return stats, nil
}

// DeleteByBatch deletes all tasks for a batch
func (r *BatchTaskRepository) DeleteByBatch(batchID string) error {
	return r.db.Where("batch_id = ?", batchID).Delete(&BatchTaskModel{}).Error
}

// ClaimPending claims pending tasks for processing (SQLite-only mode)
// Uses transaction to prevent race conditions between workers.
func (r *BatchTaskRepository) ClaimPending(batchID string, limit int) ([]BatchTaskModel, error) {
	var tasks []BatchTaskModel
	now := time.Now()

	// Use transaction for atomicity
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Find and lock pending tasks
		if err := tx.Where("batch_id = ? AND status = ?", batchID, "pending").
			Order("task_index ASC").
			Limit(limit).
			Find(&tasks).Error; err != nil {
			return err
		}

		if len(tasks) == 0 {
			return nil
		}

		// Update them to running atomically
		taskIDs := make([]string, len(tasks))
		for i, t := range tasks {
			taskIDs[i] = t.ID
		}

		if err := tx.Model(&BatchTaskModel{}).
			Where("id IN ? AND status = ?", taskIDs, "pending"). // 再次检查状态防止竞争
			Updates(map[string]interface{}{
				"status":     "running",
				"claimed_at": now,
				"started_at": now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	// Update in-memory objects
	for i := range tasks {
		tasks[i].Status = "running"
		tasks[i].ClaimedAt = &now
		tasks[i].StartedAt = &now
	}

	return tasks, nil
}
