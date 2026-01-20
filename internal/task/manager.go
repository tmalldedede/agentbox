package task

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/session"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("task")
}

// Manager 任务管理器
type Manager struct {
	store      Store
	profileMgr *profile.Manager
	sessionMgr *session.Manager

	// 调度配置
	maxConcurrent int           // 最大并发任务数
	pollInterval  time.Duration // 轮询间隔

	// 运行时状态
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   map[string]context.CancelFunc // taskID -> cancel func
	runningMu sync.Mutex
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	MaxConcurrent int
	PollInterval  time.Duration
}

// DefaultManagerConfig 默认配置
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxConcurrent: 5,
		PollInterval:  time.Second * 2,
	}
}

// NewManager 创建任务管理器
func NewManager(store Store, profileMgr *profile.Manager, sessionMgr *session.Manager, cfg *ManagerConfig) *Manager {
	if cfg == nil {
		cfg = DefaultManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		store:         store,
		profileMgr:    profileMgr,
		sessionMgr:    sessionMgr,
		maxConcurrent: cfg.MaxConcurrent,
		pollInterval:  cfg.PollInterval,
		ctx:           ctx,
		cancel:        cancel,
		running:       make(map[string]context.CancelFunc),
	}
}

// Start 启动调度器
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.scheduler()
	log.Info("task manager started", "max_concurrent", m.maxConcurrent, "poll_interval", m.pollInterval)
}

// Stop 停止调度器
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	log.Info("task manager stopped")
}

// scheduler 调度循环
func (m *Manager) scheduler() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.scheduleNext()
		}
	}
}

// scheduleNext 调度下一个任务
func (m *Manager) scheduleNext() {
	m.runningMu.Lock()
	currentRunning := len(m.running)
	m.runningMu.Unlock()

	if currentRunning >= m.maxConcurrent {
		return
	}

	// 获取等待中的任务
	tasks, err := m.store.List(&ListFilter{
		Status:    []Status{StatusQueued},
		Limit:     m.maxConcurrent - currentRunning,
		OrderBy:   "created_at",
		OrderDesc: false, // FIFO
	})
	if err != nil {
		log.Error("failed to list queued tasks", "error", err)
		return
	}

	for _, task := range tasks {
		m.runningMu.Lock()
		if len(m.running) >= m.maxConcurrent {
			m.runningMu.Unlock()
			break
		}
		m.runningMu.Unlock()

		m.wg.Add(1)
		go m.executeTask(task)
	}
}

// CreateTask 创建任务
func (m *Manager) CreateTask(req *CreateTaskRequest) (*Task, error) {
	// 验证 Profile
	p, err := m.profileMgr.Get(req.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", req.ProfileID)
	}

	// 检查 Profile 是否公开（如果需要）
	if !p.IsPublic {
		return nil, fmt.Errorf("profile is not public: %s", req.ProfileID)
	}

	// 创建任务
	now := time.Now()
	task := &Task{
		ID:          "task-" + uuid.New().String()[:8],
		ProfileID:   req.ProfileID,
		ProfileName: p.Name,
		AgentType:   p.Adapter,
		SessionID:   req.SessionID, // 使用指定的 Session（如果提供）
		Prompt:      req.Prompt,
		Input:       req.Input,
		Output:      req.Output,
		WebhookURL:  req.WebhookURL,
		Timeout:     req.Timeout,
		Status:      StatusPending,
		Metadata:    req.Metadata,
		CreatedAt:   now,
	}

	// 保存到数据库
	if err := m.store.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 立即入队
	task.Status = StatusQueued
	queuedAt := time.Now()
	task.QueuedAt = &queuedAt
	if err := m.store.Update(task); err != nil {
		log.Error("failed to queue task", "task_id", task.ID, "error", err)
	}

	log.Info("task created", "task_id", task.ID, "profile_id", task.ProfileID)
	return task, nil
}

// GetTask 获取任务
func (m *Manager) GetTask(id string) (*Task, error) {
	return m.store.Get(id)
}

// ListTasks 列出任务
func (m *Manager) ListTasks(filter *ListFilter) ([]*Task, error) {
	return m.store.List(filter)
}

// CancelTask 取消任务
func (m *Manager) CancelTask(id string) error {
	task, err := m.store.Get(id)
	if err != nil {
		return err
	}

	if !task.CanCancel() {
		return fmt.Errorf("task cannot be cancelled in status: %s", task.Status)
	}

	// 如果正在运行，取消执行
	m.runningMu.Lock()
	if cancel, ok := m.running[id]; ok {
		cancel()
		delete(m.running, id)
	}
	m.runningMu.Unlock()

	// 更新状态
	task.Status = StatusCancelled
	now := time.Now()
	task.CompletedAt = &now

	return m.store.Update(task)
}

// executeTask 执行任务
func (m *Manager) executeTask(task *Task) {
	defer m.wg.Done()

	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	// 注册到运行中
	m.runningMu.Lock()
	m.running[task.ID] = cancel
	m.runningMu.Unlock()

	defer func() {
		m.runningMu.Lock()
		delete(m.running, task.ID)
		m.runningMu.Unlock()
	}()

	// 更新状态为运行中
	task.Status = StatusRunning
	now := time.Now()
	task.StartedAt = &now
	if err := m.store.Update(task); err != nil {
		log.Error("failed to update task to running", "task_id", task.ID, "error", err)
		return
	}

	log.Info("executing task", "task_id", task.ID, "profile_id", task.ProfileID)

	// 执行任务
	err := m.doExecute(ctx, task)

	// 更新结果
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err != nil {
		task.Status = StatusFailed
		task.ErrorMessage = err.Error()
		log.Error("task failed", "task_id", task.ID, "error", err)
	} else {
		task.Status = StatusCompleted
		log.Info("task completed", "task_id", task.ID)
	}

	if err := m.store.Update(task); err != nil {
		log.Error("failed to update task result", "task_id", task.ID, "error", err)
	}

	// 发送 Webhook（如果配置了）
	if task.WebhookURL != "" {
		go m.sendWebhook(task)
	}
}

// doExecute 实际执行任务
func (m *Manager) doExecute(ctx context.Context, task *Task) error {
	// 获取解析后的 Profile
	p, err := m.profileMgr.GetResolved(task.ProfileID)
	if err != nil {
		return fmt.Errorf("failed to resolve profile: %w", err)
	}

	var sess *session.Session
	var needStopSession bool // 是否需要在任务完成后停止 session

	// 如果任务已指定 SessionID，使用现有 Session
	if task.SessionID != "" {
		sess, err = m.sessionMgr.Get(ctx, task.SessionID)
		if err != nil {
			return fmt.Errorf("failed to get session %s: %w", task.SessionID, err)
		}
		// 使用外部创建的 session，任务完成后不停止它
		needStopSession = false
		log.Info("using existing session", "session_id", sess.ID, "task_id", task.ID)
	} else {
		// 创建新 Session
		sess, err = m.sessionMgr.Create(ctx, &session.CreateRequest{
			Agent:     p.Adapter,
			ProfileID: task.ProfileID,
			Workspace: "", // 使用默认，由 session manager 分配
		})
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		task.SessionID = sess.ID
		needStopSession = true // 内部创建的 session，任务完成后需要停止

		// 启动新创建的 Session
		if err := m.sessionMgr.Start(ctx, sess.ID); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}
	}

	// 用于条件停止 session 的辅助函数
	stopSessionIfNeeded := func() {
		if needStopSession {
			m.sessionMgr.Stop(ctx, task.SessionID)
		}
	}

	// 执行任务 prompt
	timeout := task.Timeout
	if timeout == 0 {
		timeout = 1800 // 默认 30 分钟
	}

	execResp, err := m.sessionMgr.Exec(ctx, sess.ID, &session.ExecRequest{
		Prompt:  task.Prompt,
		Timeout: timeout,
	})
	if err != nil {
		stopSessionIfNeeded()
		return fmt.Errorf("failed to execute: %w", err)
	}

	// 等待执行完成
	execTimeout := time.Duration(timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	// 轮询检查执行状态
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			// 超时或取消
			stopSessionIfNeeded()
			if ctx.Err() != nil {
				return fmt.Errorf("task cancelled")
			}
			return fmt.Errorf("task timeout after %v", execTimeout)

		case <-ticker.C:
			// 检查执行状态
			exec, err := m.sessionMgr.GetExecution(ctx, sess.ID, execResp.ExecutionID)
			if err != nil {
				return fmt.Errorf("failed to get execution: %w", err)
			}

			switch exec.Status {
			case session.ExecutionSuccess:
				// 收集结果
				task.Result = m.collectResult(exec)
				// 停止 Session（如果是内部创建的）
				stopSessionIfNeeded()
				return nil
			case session.ExecutionFailed:
				stopSessionIfNeeded()
				return fmt.Errorf("execution failed: %s", exec.Error)
			}
			// ExecutionPending 或 ExecutionRunning 继续等待
		}
	}
}

// collectResult 收集执行结果
func (m *Manager) collectResult(exec *session.Execution) *Result {
	result := &Result{
		Summary: "Task completed",
		Text:    exec.Output,
	}

	// 计算耗时
	if exec.EndedAt != nil {
		result.Usage = &Usage{
			DurationSeconds: int(exec.EndedAt.Sub(exec.StartedAt).Seconds()),
		}
	}

	// TODO: 收集输出文件

	return result
}

// sendWebhook 发送 Webhook 通知
func (m *Manager) sendWebhook(task *Task) {
	// TODO: 实现 Webhook 发送
	log.Debug("sending webhook", "task_id", task.ID, "webhook_url", task.WebhookURL)
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	ProfileID  string            `json:"profile_id" binding:"required"`
	SessionID  string            `json:"session_id,omitempty"` // 可选：使用已存在的 Session
	Prompt     string            `json:"prompt" binding:"required"`
	Input      *Input            `json:"input,omitempty"`
	Output     *OutputConfig     `json:"output,omitempty"`
	WebhookURL string            `json:"webhook_url,omitempty"`
	Timeout    int               `json:"timeout,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
