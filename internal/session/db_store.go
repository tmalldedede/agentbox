package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// DBStore 会话存储（数据库实现）
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

// Create 创建会话
func (s *DBStore) Create(session *Session) error {
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	session.UpdatedAt = time.Now()
	model, err := s.toModel(session)
	if err != nil {
		return err
	}
	return s.db.Create(model).Error
}

// Get 获取会话
func (s *DBStore) Get(id string) (*Session, error) {
	var model database.SessionModel
	if err := s.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("session")
		}
		return nil, err
	}
	return s.fromModel(&model)
}

// List 列出会话
func (s *DBStore) List(filter *ListFilter) ([]*Session, error) {
	query := s.db.Model(&database.SessionModel{})

	if filter != nil {
		if filter.Agent != "" {
			query = query.Where("agent = ?", filter.Agent)
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

	query = query.Order("created_at DESC")

	var models []database.SessionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*Session, 0, len(models))
	for i := range models {
		sess, err := s.fromModel(&models[i])
		if err != nil {
			continue
		}
		result = append(result, sess)
	}
	return result, nil
}

// Count 统计会话数量
func (s *DBStore) Count(filter *ListFilter) (int, error) {
	query := s.db.Model(&database.SessionModel{})

	if filter != nil {
		if filter.Agent != "" {
			query = query.Where("agent = ?", filter.Agent)
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

// Update 更新会话
func (s *DBStore) Update(session *Session) error {
	session.UpdatedAt = time.Now()
	model, err := s.toModel(session)
	if err != nil {
		return err
	}
	return s.db.Save(model).Error
}

// Delete 删除会话
func (s *DBStore) Delete(id string) error {
	return s.db.Delete(&database.SessionModel{}, "id = ?", id).Error
}

// CreateExecution 创建执行记录
func (s *DBStore) CreateExecution(exec *Execution) error {
	if exec.StartedAt.IsZero() {
		exec.StartedAt = time.Now()
	}
	model := s.execToModel(exec)
	return s.db.Create(model).Error
}

// GetExecution 获取执行记录
func (s *DBStore) GetExecution(id string) (*Execution, error) {
	var model database.ExecutionModel
	if err := s.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("execution")
		}
		return nil, err
	}
	return s.execFromModel(&model), nil
}

// ListExecutions 列出会话的执行记录
func (s *DBStore) ListExecutions(sessionID string) ([]*Execution, error) {
	var models []database.ExecutionModel
	if err := s.db.Where("session_id = ?", sessionID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*Execution, 0, len(models))
	for i := range models {
		result = append(result, s.execFromModel(&models[i]))
	}
	return result, nil
}

// UpdateExecution 更新执行记录
func (s *DBStore) UpdateExecution(exec *Execution) error {
	model := s.execToModel(exec)
	return s.db.Save(model).Error
}

func (s *DBStore) toModel(session *Session) (*database.SessionModel, error) {
	configJSON, err := json.Marshal(session.Config)
	if err != nil {
		return nil, err
	}

	model := &database.SessionModel{
		BaseModel: database.BaseModel{
			ID:        session.ID,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		},
		AgentID:     session.AgentID,
		Agent:       session.Agent,
		Status:      string(session.Status),
		ContainerID: session.ContainerID,
		Workspace:   session.Workspace,
		Config:      string(configJSON),
	}

	return model, nil
}

func (s *DBStore) fromModel(model *database.SessionModel) (*Session, error) {
	var cfg Config
	if model.Config != "" {
		_ = json.Unmarshal([]byte(model.Config), &cfg)
	}

	return &Session{
		ID:          model.ID,
		AgentID:     model.AgentID,
		Agent:       model.Agent,
		Status:      Status(model.Status),
		Workspace:   model.Workspace,
		ContainerID: model.ContainerID,
		Config:      cfg,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (s *DBStore) execToModel(exec *Execution) *database.ExecutionModel {
	startedAt := exec.StartedAt
	var endedAt *time.Time
	if exec.EndedAt != nil {
		endedAt = exec.EndedAt
	}

	duration := int64(0)
	if endedAt != nil {
		duration = endedAt.Sub(startedAt).Milliseconds()
	}

	return &database.ExecutionModel{
		BaseModel: database.BaseModel{
			ID:        exec.ID,
			CreatedAt: exec.StartedAt,
			UpdatedAt: time.Now(),
		},
		SessionID:   exec.SessionID,
		Prompt:      exec.Prompt,
		Status:      string(exec.Status),
		Output:      exec.Output,
		Error:       exec.Error,
		ExitCode:    exec.ExitCode,
		DurationMs:  duration,
		StartedAt:   &startedAt,
		CompletedAt: endedAt,
	}
}

func (s *DBStore) execFromModel(model *database.ExecutionModel) *Execution {
	exec := &Execution{
		ID:        model.ID,
		SessionID: model.SessionID,
		Prompt:    model.Prompt,
		Status:    ExecutionStatus(model.Status),
		Output:    model.Output,
		Error:     model.Error,
		ExitCode:  model.ExitCode,
	}
	if model.StartedAt != nil {
		exec.StartedAt = *model.StartedAt
	}
	if model.CompletedAt != nil {
		exec.EndedAt = model.CompletedAt
	}
	return exec
}
