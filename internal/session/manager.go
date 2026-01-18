package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
)

// Manager 会话管理器
type Manager struct {
	store         Store
	containerMgr  container.Manager
	agentRegistry *agent.Registry
	workspaceBase string
}

// NewManager 创建会话管理器
func NewManager(store Store, containerMgr container.Manager, registry *agent.Registry, workspaceBase string) *Manager {
	return &Manager{
		store:         store,
		containerMgr:  containerMgr,
		agentRegistry: registry,
		workspaceBase: workspaceBase,
	}
}

// Create 创建会话
func (m *Manager) Create(ctx context.Context, req *CreateRequest) (*Session, error) {
	// 获取 Agent 适配器
	adapter, err := m.agentRegistry.Get(req.Agent)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %s", req.Agent)
	}

	// 生成会话 ID
	sessionID := uuid.New().String()[:8]

	// 解析工作空间路径
	workspace := req.Workspace
	if !filepath.IsAbs(workspace) {
		workspace = filepath.Join(m.workspaceBase, workspace)
	}

	// 确保工作空间存在
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// 创建会话
	session := &Session{
		ID:        sessionID,
		Agent:     req.Agent,
		Status:    StatusCreating,
		Workspace: workspace,
		Env:       req.Env,
		Config: Config{
			CPULimit:    2.0,
			MemoryLimit: 4 * 1024 * 1024 * 1024,
		},
	}

	if req.Config != nil {
		if req.Config.CPULimit > 0 {
			session.Config.CPULimit = req.Config.CPULimit
		}
		if req.Config.MemoryLimit > 0 {
			session.Config.MemoryLimit = req.Config.MemoryLimit
		}
	}

	// 保存会话
	if err := m.store.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 准备容器配置
	containerConfig := adapter.PrepareContainer(&agent.SessionInfo{
		ID:        sessionID,
		Workspace: workspace,
		Env:       req.Env,
	})

	// 应用资源限制
	containerConfig.Resources.CPULimit = session.Config.CPULimit
	containerConfig.Resources.MemoryLimit = session.Config.MemoryLimit

	// 创建容器
	ctr, err := m.containerMgr.Create(ctx, containerConfig)
	if err != nil {
		session.Status = StatusError
		m.store.Update(session)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	session.ContainerID = ctr.ID

	// 启动容器
	if err := m.containerMgr.Start(ctx, ctr.ID); err != nil {
		session.Status = StatusError
		m.store.Update(session)
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	session.Status = StatusRunning
	if err := m.store.Update(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// Get 获取会话
func (m *Manager) Get(ctx context.Context, id string) (*Session, error) {
	return m.store.Get(id)
}

// List 列出会话
func (m *Manager) List(ctx context.Context, filter *ListFilter) ([]*Session, error) {
	return m.store.List(filter)
}

// Delete 删除会话
func (m *Manager) Delete(ctx context.Context, id string) error {
	session, err := m.store.Get(id)
	if err != nil {
		return err
	}

	// 删除容器
	if session.ContainerID != "" {
		if err := m.containerMgr.Stop(ctx, session.ContainerID); err != nil {
			// 忽略停止错误，继续删除
		}
		if err := m.containerMgr.Remove(ctx, session.ContainerID); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	// 删除会话记录
	return m.store.Delete(id)
}

// Stop 停止会话
func (m *Manager) Stop(ctx context.Context, id string) error {
	session, err := m.store.Get(id)
	if err != nil {
		return err
	}

	if session.ContainerID != "" {
		if err := m.containerMgr.Stop(ctx, session.ContainerID); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}
	}

	session.Status = StatusStopped
	return m.store.Update(session)
}

// Start 启动已停止的会话
func (m *Manager) Start(ctx context.Context, id string) error {
	session, err := m.store.Get(id)
	if err != nil {
		return err
	}

	if session.ContainerID != "" {
		if err := m.containerMgr.Start(ctx, session.ContainerID); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}
	}

	session.Status = StatusRunning
	return m.store.Update(session)
}

// Exec 在会话中执行命令
func (m *Manager) Exec(ctx context.Context, id string, req *ExecRequest) (*ExecResponse, error) {
	session, err := m.store.Get(id)
	if err != nil {
		return nil, err
	}

	if session.Status != StatusRunning {
		return nil, fmt.Errorf("session is not running: %s", session.Status)
	}

	// 获取 Agent 适配器
	adapter, err := m.agentRegistry.Get(session.Agent)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %s", session.Agent)
	}

	// 准备执行命令
	cmd := adapter.PrepareExec(req.Prompt)

	// 创建执行记录
	execID := uuid.New().String()[:8]
	execution := &Execution{
		ID:        execID,
		SessionID: id,
		Prompt:    req.Prompt,
		Status:    ExecutionRunning,
		StartedAt: time.Now(),
	}
	if err := m.store.CreateExecution(execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// 在容器中执行
	result, err := m.containerMgr.Exec(ctx, session.ContainerID, cmd)
	if err != nil {
		execution.Status = ExecutionFailed
		execution.Error = err.Error()
		now := time.Now()
		execution.EndedAt = &now
		m.store.UpdateExecution(execution)
		return nil, fmt.Errorf("failed to execute: %w", err)
	}

	// 更新执行记录
	now := time.Now()
	execution.EndedAt = &now
	execution.Output = result.Stdout
	execution.ExitCode = result.ExitCode
	if result.ExitCode == 0 {
		execution.Status = ExecutionSuccess
	} else {
		execution.Status = ExecutionFailed
		execution.Error = result.Stderr
	}
	m.store.UpdateExecution(execution)

	return &ExecResponse{
		ExecutionID: execID,
		Output:      result.Stdout,
		ExitCode:    result.ExitCode,
		Error:       result.Stderr,
	}, nil
}

// GetExecutions 获取会话的执行历史
func (m *Manager) GetExecutions(ctx context.Context, sessionID string) ([]*Execution, error) {
	return m.store.ListExecutions(sessionID)
}
