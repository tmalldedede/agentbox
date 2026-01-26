// Package cron 提供定时任务调度
package cron

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("cron")
}

// JobFunc 任务执行函数
type JobFunc func(ctx context.Context, job *Job) error

// Job 定时任务
type Job struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Schedule    string            `json:"schedule"`     // cron 表达式
	Enabled     bool              `json:"enabled"`
	AgentID     string            `json:"agent_id"`     // 关联的 Agent
	Prompt      string            `json:"prompt"`       // 执行的 prompt
	Metadata    map[string]string `json:"metadata"`     // 额外数据
	LastRun     *time.Time        `json:"last_run"`     // 上次执行时间
	NextRun     *time.Time        `json:"next_run"`     // 下次执行时间
	LastStatus  string            `json:"last_status"`  // 上次状态: success, failed
	LastError   string            `json:"last_error"`   // 上次错误信息
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	entryID     cron.EntryID      // cron 库的 entry ID
}

// CreateJobRequest 创建任务请求
type CreateJobRequest struct {
	Name     string            `json:"name" binding:"required"`
	Schedule string            `json:"schedule" binding:"required"` // cron 表达式
	AgentID  string            `json:"agent_id" binding:"required"`
	Prompt   string            `json:"prompt" binding:"required"`
	Enabled  *bool             `json:"enabled"`
	Metadata map[string]string `json:"metadata"`
}

// UpdateJobRequest 更新任务请求
type UpdateJobRequest struct {
	Name     string            `json:"name"`
	Schedule string            `json:"schedule"`
	AgentID  string            `json:"agent_id"`
	Prompt   string            `json:"prompt"`
	Enabled  *bool             `json:"enabled"`
	Metadata map[string]string `json:"metadata"`
}

// Store 任务存储接口
type Store interface {
	Create(job *Job) error
	Get(id string) (*Job, error)
	Update(job *Job) error
	Delete(id string) error
	List() ([]*Job, error)
	ListEnabled() ([]*Job, error)
}

// Manager Cron 管理器
type Manager struct {
	store    Store
	cron     *cron.Cron
	executor JobFunc

	jobs   map[string]*Job // id -> job
	mu     sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager 创建 Cron 管理器
func NewManager(store Store, executor JobFunc) *Manager {
	return &Manager{
		store:    store,
		executor: executor,
		jobs:     make(map[string]*Job),
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
	}
}

// Start 启动调度器
func (m *Manager) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	// 加载已有任务
	jobs, err := m.store.ListEnabled()
	if err != nil {
		return fmt.Errorf("load jobs: %w", err)
	}

	for _, job := range jobs {
		if err := m.scheduleJob(job); err != nil {
			log.Error("schedule job failed", "id", job.ID, "name", job.Name, "error", err)
		}
	}

	m.cron.Start()
	log.Info("cron manager started", "jobs", len(jobs))
	return nil
}

// Stop 停止调度器
func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}

	stopCtx := m.cron.Stop()
	<-stopCtx.Done()

	log.Info("cron manager stopped")
	return nil
}

// Create 创建任务
func (m *Manager) Create(req *CreateJobRequest) (*Job, error) {
	// 验证 cron 表达式
	if _, err := cron.ParseStandard(req.Schedule); err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	job := &Job{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Schedule:  req.Schedule,
		Enabled:   enabled,
		AgentID:   req.AgentID,
		Prompt:    req.Prompt,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.store.Create(job); err != nil {
		return nil, err
	}

	if job.Enabled {
		if err := m.scheduleJob(job); err != nil {
			log.Error("schedule job failed", "id", job.ID, "error", err)
		}
	}

	return job, nil
}

// Get 获取任务
func (m *Manager) Get(id string) (*Job, error) {
	return m.store.Get(id)
}

// Update 更新任务
func (m *Manager) Update(id string, req *UpdateJobRequest) (*Job, error) {
	job, err := m.store.Get(id)
	if err != nil {
		return nil, err
	}

	// 先取消调度
	m.unscheduleJob(job)

	// 更新字段
	if req.Name != "" {
		job.Name = req.Name
	}
	if req.Schedule != "" {
		if _, err := cron.ParseStandard(req.Schedule); err != nil {
			return nil, fmt.Errorf("invalid schedule: %w", err)
		}
		job.Schedule = req.Schedule
	}
	if req.AgentID != "" {
		job.AgentID = req.AgentID
	}
	if req.Prompt != "" {
		job.Prompt = req.Prompt
	}
	if req.Enabled != nil {
		job.Enabled = *req.Enabled
	}
	if req.Metadata != nil {
		job.Metadata = req.Metadata
	}
	job.UpdatedAt = time.Now()

	if err := m.store.Update(job); err != nil {
		return nil, err
	}

	// 重新调度
	if job.Enabled {
		if err := m.scheduleJob(job); err != nil {
			log.Error("schedule job failed", "id", job.ID, "error", err)
		}
	}

	return job, nil
}

// Delete 删除任务
func (m *Manager) Delete(id string) error {
	job, err := m.store.Get(id)
	if err != nil {
		return err
	}

	m.unscheduleJob(job)
	return m.store.Delete(id)
}

// List 列出所有任务
func (m *Manager) List() ([]*Job, error) {
	return m.store.List()
}

// TriggerNow 立即执行任务
func (m *Manager) TriggerNow(id string) error {
	job, err := m.store.Get(id)
	if err != nil {
		return err
	}

	go m.runJob(job)
	return nil
}

// scheduleJob 调度任务
func (m *Manager) scheduleJob(job *Job) error {
	entryID, err := m.cron.AddFunc(job.Schedule, func() {
		m.runJob(job)
	})
	if err != nil {
		return err
	}

	m.mu.Lock()
	job.entryID = entryID
	m.jobs[job.ID] = job

	// 计算下次执行时间
	entry := m.cron.Entry(entryID)
	next := entry.Next
	job.NextRun = &next
	m.mu.Unlock()

	log.Info("job scheduled", "id", job.ID, "name", job.Name, "schedule", job.Schedule, "next", next)
	return nil
}

// unscheduleJob 取消任务调度
func (m *Manager) unscheduleJob(job *Job) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if j, ok := m.jobs[job.ID]; ok {
		m.cron.Remove(j.entryID)
		delete(m.jobs, job.ID)
		log.Info("job unscheduled", "id", job.ID, "name", job.Name)
	}
}

// runJob 执行任务
func (m *Manager) runJob(job *Job) {
	log.Info("running job", "id", job.ID, "name", job.Name)

	now := time.Now()
	job.LastRun = &now

	err := m.executor(m.ctx, job)
	if err != nil {
		job.LastStatus = "failed"
		job.LastError = err.Error()
		log.Error("job failed", "id", job.ID, "name", job.Name, "error", err)
	} else {
		job.LastStatus = "success"
		job.LastError = ""
		log.Info("job completed", "id", job.ID, "name", job.Name)
	}

	// 更新下次执行时间
	m.mu.RLock()
	if j, ok := m.jobs[job.ID]; ok {
		entry := m.cron.Entry(j.entryID)
		next := entry.Next
		job.NextRun = &next
	}
	m.mu.RUnlock()

	// 保存状态
	job.UpdatedAt = time.Now()
	if err := m.store.Update(job); err != nil {
		log.Error("update job failed", "id", job.ID, "error", err)
	}
}
