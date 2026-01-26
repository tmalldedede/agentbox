package cron

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// ErrNotFound 未找到错误
var ErrNotFound = errors.New("cron job not found")

// DBStore 数据库存储
type DBStore struct {
	db *gorm.DB
}

// CronJobModel 数据库模型
type CronJobModel struct {
	ID         string     `gorm:"primaryKey;size:36"`
	Name       string     `gorm:"size:255;not null"`
	Schedule   string     `gorm:"size:100;not null"`
	Enabled    bool       `gorm:"default:true"`
	AgentID    string     `gorm:"size:36;not null"`
	Prompt     string     `gorm:"type:text"`
	Metadata   string     `gorm:"type:text"` // JSON
	LastRun    *time.Time
	NextRun    *time.Time
	LastStatus string     `gorm:"size:20"`
	LastError  string     `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName 表名
func (CronJobModel) TableName() string {
	return "cron_jobs"
}

// NewDBStore 创建数据库存储
func NewDBStore() *DBStore {
	db := database.GetDB()

	// 自动迁移
	if err := db.AutoMigrate(&CronJobModel{}); err != nil {
		log.Error("auto migrate cron_jobs failed", "error", err)
	}

	return &DBStore{db: db}
}

// Create 创建任务
func (s *DBStore) Create(job *Job) error {
	model := s.toModel(job)
	return s.db.Create(model).Error
}

// Get 获取任务
func (s *DBStore) Get(id string) (*Job, error) {
	var model CronJobModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.toJob(&model), nil
}

// Update 更新任务
func (s *DBStore) Update(job *Job) error {
	model := s.toModel(job)
	return s.db.Save(model).Error
}

// Delete 删除任务
func (s *DBStore) Delete(id string) error {
	return s.db.Delete(&CronJobModel{}, "id = ?", id).Error
}

// List 列出所有任务
func (s *DBStore) List() ([]*Job, error) {
	var models []CronJobModel
	if err := s.db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	jobs := make([]*Job, len(models))
	for i, m := range models {
		jobs[i] = s.toJob(&m)
	}
	return jobs, nil
}

// ListEnabled 列出启用的任务
func (s *DBStore) ListEnabled() ([]*Job, error) {
	var models []CronJobModel
	if err := s.db.Where("enabled = ?", true).Find(&models).Error; err != nil {
		return nil, err
	}

	jobs := make([]*Job, len(models))
	for i, m := range models {
		jobs[i] = s.toJob(&m)
	}
	return jobs, nil
}

// toModel 转换为数据库模型
func (s *DBStore) toModel(job *Job) *CronJobModel {
	metadata := ""
	if job.Metadata != nil {
		if b, err := json.Marshal(job.Metadata); err == nil {
			metadata = string(b)
		}
	}

	return &CronJobModel{
		ID:         job.ID,
		Name:       job.Name,
		Schedule:   job.Schedule,
		Enabled:    job.Enabled,
		AgentID:    job.AgentID,
		Prompt:     job.Prompt,
		Metadata:   metadata,
		LastRun:    job.LastRun,
		NextRun:    job.NextRun,
		LastStatus: job.LastStatus,
		LastError:  job.LastError,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
	}
}

// toJob 转换为业务对象
func (s *DBStore) toJob(model *CronJobModel) *Job {
	var metadata map[string]string
	if model.Metadata != "" {
		json.Unmarshal([]byte(model.Metadata), &metadata)
	}

	return &Job{
		ID:         model.ID,
		Name:       model.Name,
		Schedule:   model.Schedule,
		Enabled:    model.Enabled,
		AgentID:    model.AgentID,
		Prompt:     model.Prompt,
		Metadata:   metadata,
		LastRun:    model.LastRun,
		NextRun:    model.NextRun,
		LastStatus: model.LastStatus,
		LastError:  model.LastError,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}
