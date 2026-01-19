package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

// Manager 会话管理器
type Manager struct {
	store         Store
	containerMgr  container.Manager
	agentRegistry *agent.Registry
	profileMgr    *profile.Manager
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

// SetProfileManager 设置 Profile 管理器（可选依赖）
func (m *Manager) SetProfileManager(mgr *profile.Manager) {
	m.profileMgr = mgr
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
		ProfileID: req.ProfileID,
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
		_ = m.store.Update(session)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	session.ContainerID = ctr.ID

	// 启动容器
	if err := m.containerMgr.Start(ctx, ctr.ID); err != nil {
		session.Status = StatusError
		_ = m.store.Update(session)
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// 写入配置文件（如果适配器需要）
	if err := m.writeConfigFiles(ctx, adapter, ctr.ID, req); err != nil {
		// 配置文件写入失败不中断创建，但记录警告
		fmt.Printf("Warning: failed to write config files: %v\n", err)
	}

	session.Status = StatusRunning
	if err := m.store.Update(session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// writeConfigFiles 写入配置文件到容器
func (m *Manager) writeConfigFiles(ctx context.Context, adapter agent.Adapter, containerID string, req *CreateRequest) error {
	// 检查适配器是否实现 ConfigFilesProvider 接口
	provider, ok := adapter.(agent.ConfigFilesProvider)
	if !ok {
		return nil // 适配器不需要配置文件
	}

	// 获取 Profile
	var p *profile.Profile
	if req.ProfileID != "" && m.profileMgr != nil {
		var err error
		p, err = m.profileMgr.Get(req.ProfileID)
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}
	} else {
		// 创建空 Profile
		p = &profile.Profile{}
	}

	// 从环境变量提取 API Key
	// 常见的 API Key 环境变量名
	apiKey := ""
	for _, key := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "API_KEY"} {
		if v, ok := req.Env[key]; ok && v != "" {
			apiKey = v
			break
		}
	}

	// 获取配置文件
	configFiles := provider.GetConfigFiles(p, apiKey)
	if len(configFiles) == 0 {
		return nil
	}

	// 通过 exec 命令写入每个配置文件
	for path, content := range configFiles {
		// 创建目录
		dir := filepath.Dir(path)
		mkdirCmd := []string{"mkdir", "-p", dir}
		if _, err := m.containerMgr.Exec(ctx, containerID, mkdirCmd); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// 使用 cat 和 heredoc 写入文件内容
		// 转义内容中的特殊字符
		escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
		writeCmd := []string{"sh", "-c", fmt.Sprintf("cat > '%s' << 'AGENTBOX_EOF'\n%s\nAGENTBOX_EOF", path, escapedContent)}
		if _, err := m.containerMgr.Exec(ctx, containerID, writeCmd); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

// Get 获取会话
func (m *Manager) Get(ctx context.Context, id string) (*Session, error) {
	return m.store.Get(id)
}

// List 列出会话
func (m *Manager) List(ctx context.Context, filter *ListFilter) ([]*Session, error) {
	return m.store.List(filter)
}

// ListWithCount 列出会话并返回总数 (用于分页)
func (m *Manager) ListWithCount(ctx context.Context, filter *ListFilter) ([]*Session, int, error) {
	sessions, err := m.store.List(filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := m.store.Count(filter)
	if err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

// Delete 删除会话
func (m *Manager) Delete(ctx context.Context, id string) error {
	session, err := m.store.Get(id)
	if err != nil {
		return err
	}

	// 删除容器
	if session.ContainerID != "" {
		// 忽略停止错误，继续删除
		_ = m.containerMgr.Stop(ctx, session.ContainerID)
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

// Reconnect 重连会话
// 如果会话已停止，尝试重新启动容器
// 如果会话正在运行，直接返回会话信息
func (m *Manager) Reconnect(ctx context.Context, id string) (*Session, error) {
	session, err := m.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// 如果会话已在运行，直接返回
	if session.Status == StatusRunning {
		// 验证容器是否真的在运行
		if session.ContainerID != "" {
			ctr, err := m.containerMgr.Inspect(ctx, session.ContainerID)
			if err == nil && ctr.Status == container.StatusRunning {
				return session, nil
			}
		}
	}

	// 尝试重新启动容器
	if session.ContainerID != "" {
		// 先检查容器状态
		ctr, err := m.containerMgr.Inspect(ctx, session.ContainerID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container: %w", err)
		}

		// 如果容器已停止，重新启动
		if ctr.Status != container.StatusRunning {
			if err := m.containerMgr.Start(ctx, session.ContainerID); err != nil {
				return nil, fmt.Errorf("failed to restart container: %w", err)
			}
		}

		session.Status = StatusRunning
		if err := m.store.Update(session); err != nil {
			return nil, fmt.Errorf("failed to update session: %w", err)
		}
	} else {
		return nil, fmt.Errorf("session has no associated container")
	}

	return session, nil
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

	// 准备执行选项
	execOpts := &agent.ExecOptions{
		Prompt:          req.Prompt,
		MaxTurns:        req.MaxTurns,
		Timeout:         req.Timeout,
		AllowedTools:    req.AllowedTools,
		DisallowedTools: req.DisallowedTools,
	}

	// 设置默认值
	if execOpts.MaxTurns <= 0 {
		execOpts.MaxTurns = 10
	}
	if execOpts.Timeout <= 0 {
		execOpts.Timeout = 300 // 默认 5 分钟
	}

	// 准备执行命令
	cmd := adapter.PrepareExec(execOpts)

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

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(execOpts.Timeout)*time.Second)
	defer cancel()

	// 在容器中执行
	result, err := m.containerMgr.Exec(execCtx, session.ContainerID, cmd)
	if err != nil {
		execution.Status = ExecutionFailed
		if execCtx.Err() == context.DeadlineExceeded {
			execution.Error = fmt.Sprintf("execution timeout after %d seconds", execOpts.Timeout)
		} else {
			execution.Error = err.Error()
		}
		now := time.Now()
		execution.EndedAt = &now
		_ = m.store.UpdateExecution(execution)
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
	_ = m.store.UpdateExecution(execution)

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

// GetExecution 获取单个执行记录
func (m *Manager) GetExecution(ctx context.Context, sessionID, execID string) (*Execution, error) {
	exec, err := m.store.GetExecution(execID)
	if err != nil {
		return nil, err
	}
	// 验证执行记录属于该会话
	if exec.SessionID != sessionID {
		return nil, fmt.Errorf("execution %s does not belong to session %s", execID, sessionID)
	}
	return exec, nil
}

// GetWorkspace 获取会话工作空间路径
func (m *Manager) GetWorkspace(sessionID string) (string, error) {
	session, err := m.store.Get(sessionID)
	if err != nil {
		return "", err
	}
	return session.Workspace, nil
}

// GetLogs 获取会话容器日志
func (m *Manager) GetLogs(ctx context.Context, id string) (string, error) {
	session, err := m.store.Get(id)
	if err != nil {
		return "", err
	}

	if session.ContainerID == "" {
		return "", fmt.Errorf("session has no container")
	}

	reader, err := m.containerMgr.Logs(ctx, session.ContainerID)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	// 读取日志内容
	buf := make([]byte, 64*1024) // 64KB 限制
	n, _ := reader.Read(buf)
	return string(buf[:n]), nil
}
