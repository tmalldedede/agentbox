package session

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/profile"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("session")
}

// Manager 会话管理器
type Manager struct {
	store         Store
	containerMgr  container.Manager
	agentRegistry *agent.Registry
	profileMgr    *profile.Manager
	credentialMgr *credential.Manager
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

// SetCredentialManager 设置 Credential 管理器（可选依赖）
func (m *Manager) SetCredentialManager(mgr *credential.Manager) {
	m.credentialMgr = mgr
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

	// 准备环境变量（可能需要从 Credential 注入 API Key）
	envVars := make(map[string]string)
	for k, v := range req.Env {
		envVars[k] = v
	}

	// 如果请求中没有 API Key，尝试从 Profile 关联的 Credential 获取
	hasAPIKey := false
	for _, key := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "API_KEY"} {
		if _, ok := envVars[key]; ok {
			hasAPIKey = true
			break
		}
	}
	if !hasAPIKey && req.ProfileID != "" && m.profileMgr != nil && m.credentialMgr != nil {
		if p, err := m.profileMgr.Get(req.ProfileID); err == nil && p.CredentialID != "" {
			if apiKey, err := m.credentialMgr.GetDecrypted(p.CredentialID); err == nil {
				envVars["OPENAI_API_KEY"] = apiKey
			}
		}
	}

	// 准备容器配置
	containerConfig := adapter.PrepareContainer(&agent.SessionInfo{
		ID:        sessionID,
		Workspace: workspace,
		Env:       envVars,
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
		log.Warn("failed to write config files", "session_id", sessionID, "error", err)
	} else {
		log.Debug("config files written", "session_id", sessionID)
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
	cfgProvider, ok := adapter.(agent.ConfigFilesProvider)
	if !ok {
		log.Debug("adapter does not implement ConfigFilesProvider", "adapter", adapter.Name())
		return nil // 适配器不需要配置文件
	}
	log.Debug("adapter implements ConfigFilesProvider", "adapter", adapter.Name(), "profile_id", req.ProfileID)

	// 获取 Profile
	var p *profile.Profile
	if req.ProfileID != "" && m.profileMgr != nil {
		var err error
		p, err = m.profileMgr.Get(req.ProfileID)
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}
		// 创建副本以避免修改原始 Profile
		pCopy := *p
		p = &pCopy
	} else {
		// 创建空 Profile
		p = &profile.Profile{}
	}

	// 如果 Profile 没有 Model 配置，尝试从环境变量提取
	if p.Model.BaseURL == "" {
		// Codex 使用 OPENAI_BASE_URL
		if baseURL, ok := req.Env["OPENAI_BASE_URL"]; ok && baseURL != "" {
			p.Model.BaseURL = baseURL
		}
	}
	if p.Model.Provider == "" {
		// 尝试从 MODEL_PROVIDER 环境变量获取
		if provider, ok := req.Env["MODEL_PROVIDER"]; ok && provider != "" {
			p.Model.Provider = provider
		} else if p.Model.BaseURL != "" {
			// 从 BaseURL 推断 Provider 名称
			p.Model.Provider = inferProviderFromBaseURL(p.Model.BaseURL)
		}
	}
	if p.Model.Name == "" {
		// 尝试从 MODEL 或 CODEX_MODEL 环境变量获取
		for _, key := range []string{"MODEL", "CODEX_MODEL", "OPENAI_MODEL"} {
			if model, ok := req.Env[key]; ok && model != "" {
				p.Model.Name = model
				break
			}
		}
	}
	if p.Model.WireAPI == "" {
		// 尝试从 WIRE_API 环境变量获取
		if wireAPI, ok := req.Env["WIRE_API"]; ok && wireAPI != "" {
			p.Model.WireAPI = wireAPI
		}
	}

	// 获取 API Key（优先级：环境变量 > Profile Credential）
	apiKey := ""
	// 1. 先尝试从环境变量获取
	for _, key := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "API_KEY"} {
		if v, ok := req.Env[key]; ok && v != "" {
			apiKey = v
			break
		}
	}
	// 2. 如果没有，尝试从 Profile 关联的 Credential 获取
	if apiKey == "" && p.CredentialID != "" && m.credentialMgr != nil {
		if decrypted, err := m.credentialMgr.GetDecrypted(p.CredentialID); err == nil {
			apiKey = decrypted
		}
	}

	// 获取配置文件
	log.Debug("profile model config",
		"model", p.Model.Name,
		"provider", p.Model.Provider,
		"base_url", p.Model.BaseURL,
		"api_key_present", apiKey != "",
	)
	configFiles := cfgProvider.GetConfigFiles(p, apiKey)
	log.Debug("got config files", "count", len(configFiles))
	if len(configFiles) == 0 {
		log.Debug("no config files to write")
		return nil
	}

	// 通过 exec 命令写入每个配置文件
	for path, content := range configFiles {
		log.Debug("writing config file", "path", path, "content_len", len(content))

		// 处理 ~ 路径：使用 shell 展开
		// 注意：~ 只在 shell 中展开，所以必须用 sh -c
		expandedPath := path
		if strings.HasPrefix(path, "~/") {
			// 使用 $HOME 替代 ~，在 sh -c 中会正确展开
			expandedPath = "$HOME" + path[1:]
		}

		// 获取目录路径
		dir := filepath.Dir(expandedPath)

		// 使用单个 shell 命令完成创建目录和写入文件
		// 转义内容中的特殊字符（单引号）
		escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
		writeCmd := []string{
			"sh", "-c",
			fmt.Sprintf("mkdir -p %s && cat > %s << 'AGENTBOX_EOF'\n%s\nAGENTBOX_EOF", dir, expandedPath, escapedContent),
		}
		log.Debug("exec write command", "dir", dir, "path", expandedPath)
		result, err := m.containerMgr.Exec(ctx, containerID, writeCmd)
		if err != nil {
			log.Error("failed to write config file", "path", path, "error", err)
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
		log.Debug("config file written", "path", path, "exit_code", result.ExitCode)
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

	// 删除容器（忽略容器不存在的错误）
	if session.ContainerID != "" {
		_ = m.containerMgr.Stop(ctx, session.ContainerID)
		_ = m.containerMgr.Remove(ctx, session.ContainerID)
		// 忽略错误，容器可能已经被删除
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

	// 检查容器是否存在
	_, err = m.containerMgr.Inspect(ctx, session.ContainerID)
	if err != nil {
		// 容器不存在，更新 session 状态
		session.Status = StatusError
		_ = m.store.Update(session)
		return nil, fmt.Errorf("container no longer exists, session marked as error: %w", err)
	}

	// 获取 Agent 适配器
	adapter, err := m.agentRegistry.Get(session.Agent)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %s", session.Agent)
	}

	// 准备执行选项
	execOpts := &agent.ExecOptions{
		Prompt:           req.Prompt,
		MaxTurns:         req.MaxTurns,
		Timeout:          req.Timeout,
		AllowedTools:     req.AllowedTools,
		DisallowedTools:  req.DisallowedTools,
		IncludeEvents:    req.IncludeEvents,
		WorkingDirectory: session.Workspace,
	}

	// 设置默认值
	if execOpts.MaxTurns <= 0 {
		execOpts.MaxTurns = 10
	}
	if execOpts.Timeout <= 0 {
		execOpts.Timeout = 300 // 默认 5 分钟
	}

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

	// 检查 adapter 是否实现了 DirectExecutor 接口
	if directExec, ok := adapter.(agent.DirectExecutor); ok {
		// 使用 Go SDK 直接执行
		return m.execDirect(execCtx, directExec, execOpts, execution)
	}

	// 回退到 CLI 执行方式
	return m.execViaCLI(execCtx, adapter, execOpts, session.ContainerID, execution)
}

// execDirect 使用 Go SDK 直接执行 (Codex)
func (m *Manager) execDirect(ctx context.Context, executor agent.DirectExecutor, opts *agent.ExecOptions, execution *Execution) (*ExecResponse, error) {
	result, err := executor.Execute(ctx, opts)
	if err != nil {
		execution.Status = ExecutionFailed
		if ctx.Err() == context.DeadlineExceeded {
			execution.Error = fmt.Sprintf("execution timeout after %d seconds", opts.Timeout)
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
	execution.Output = result.Message
	execution.ExitCode = result.ExitCode
	if result.ExitCode == 0 && result.Error == "" {
		execution.Status = ExecutionSuccess
	} else {
		execution.Status = ExecutionFailed
		execution.Error = result.Error
	}
	_ = m.store.UpdateExecution(execution)

	// 构建响应
	resp := &ExecResponse{
		ExecutionID: execution.ID,
		Message:     result.Message,
		Output:      result.Message, // 兼容旧版
		ExitCode:    result.ExitCode,
		Error:       result.Error,
	}

	// 添加 token 使用统计
	if result.Usage != nil {
		resp.Usage = &TokenUsage{
			InputTokens:       result.Usage.InputTokens,
			CachedInputTokens: result.Usage.CachedInputTokens,
			OutputTokens:      result.Usage.OutputTokens,
		}
	}

	// 添加事件列表
	if opts.IncludeEvents && len(result.Events) > 0 {
		resp.Events = make([]ExecEvent, len(result.Events))
		for i, e := range result.Events {
			resp.Events[i] = ExecEvent{
				Type: e.Type,
				Raw:  e.Raw,
			}
		}
	}

	return resp, nil
}

// execViaCLI 通过 CLI 在容器中执行 (Claude Code, OpenCode, Codex)
func (m *Manager) execViaCLI(ctx context.Context, adapter agent.Adapter, opts *agent.ExecOptions, containerID string, execution *Execution) (*ExecResponse, error) {
	// 准备执行命令
	cmd := adapter.PrepareExec(opts)

	// 在容器中执行
	result, err := m.containerMgr.Exec(ctx, containerID, cmd)
	if err != nil {
		execution.Status = ExecutionFailed
		if ctx.Err() == context.DeadlineExceeded {
			execution.Error = fmt.Sprintf("execution timeout after %d seconds", opts.Timeout)
		} else {
			execution.Error = err.Error()
		}
		now := time.Now()
		execution.EndedAt = &now
		_ = m.store.UpdateExecution(execution)
		return nil, fmt.Errorf("failed to execute: %w", err)
	}

	// 检查 adapter 是否实现了 JSONOutputParser 接口
	if parser, ok := adapter.(agent.JSONOutputParser); ok {
		// 使用 JSON 解析器解析输出 (如 Codex --json 模式)
		return m.execViaCLIWithJSONParser(parser, opts, result, execution)
	}

	// 更新执行记录 (普通文本输出模式)
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
		ExecutionID: execution.ID,
		Message:     result.Stdout, // CLI 模式下，output 就是 message
		Output:      result.Stdout,
		ExitCode:    result.ExitCode,
		Error:       result.Stderr,
	}, nil
}

// execViaCLIWithJSONParser 使用 JSON 解析器处理 CLI 输出 (如 Codex --json 模式)
func (m *Manager) execViaCLIWithJSONParser(parser agent.JSONOutputParser, opts *agent.ExecOptions, result *container.ExecResult, execution *Execution) (*ExecResponse, error) {
	// 解析 JSONL 输出
	parsed, err := parser.ParseJSONLOutput(result.Stdout, opts.IncludeEvents)
	if err != nil {
		// 解析失败，回退到普通文本模式
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
			ExecutionID: execution.ID,
			Message:     result.Stdout,
			Output:      result.Stdout,
			ExitCode:    result.ExitCode,
			Error:       fmt.Sprintf("JSON parse failed: %v; stderr: %s", err, result.Stderr),
		}, nil
	}

	// 更新执行记录
	now := time.Now()
	execution.EndedAt = &now
	execution.Output = parsed.Message
	execution.ExitCode = parsed.ExitCode
	if parsed.ExitCode == 0 && parsed.Error == "" {
		execution.Status = ExecutionSuccess
	} else {
		execution.Status = ExecutionFailed
		execution.Error = parsed.Error
	}
	_ = m.store.UpdateExecution(execution)

	// 构建响应
	resp := &ExecResponse{
		ExecutionID: execution.ID,
		Message:     parsed.Message,
		Output:      parsed.Message, // 兼容旧版
		ExitCode:    parsed.ExitCode,
		Error:       parsed.Error,
	}

	// 添加 token 使用统计
	if parsed.Usage != nil {
		resp.Usage = &TokenUsage{
			InputTokens:       parsed.Usage.InputTokens,
			CachedInputTokens: parsed.Usage.CachedInputTokens,
			OutputTokens:      parsed.Usage.OutputTokens,
		}
	}

	// 添加事件列表
	if opts.IncludeEvents && len(parsed.Events) > 0 {
		resp.Events = make([]ExecEvent, len(parsed.Events))
		for i, e := range parsed.Events {
			resp.Events[i] = ExecEvent{
				Type: e.Type,
				Raw:  e.Raw,
			}
		}
	}

	return resp, nil
}

// ExecStream 流式执行命令，返回事件通道 (目前仅支持 Codex)
func (m *Manager) ExecStream(ctx context.Context, id string, req *ExecRequest) (<-chan *StreamEvent, string, error) {
	session, err := m.store.Get(id)
	if err != nil {
		return nil, "", err
	}

	if session.Status != StatusRunning {
		return nil, "", fmt.Errorf("session is not running: %s", session.Status)
	}

	// 检查容器是否存在
	_, err = m.containerMgr.Inspect(ctx, session.ContainerID)
	if err != nil {
		session.Status = StatusError
		_ = m.store.Update(session)
		return nil, "", fmt.Errorf("container no longer exists: %w", err)
	}

	// 获取 Agent 适配器
	adapter, err := m.agentRegistry.Get(session.Agent)
	if err != nil {
		return nil, "", fmt.Errorf("agent not found: %s", session.Agent)
	}

	// 目前只有 Codex 支持流式输出 (--json 模式)
	if session.Agent != "codex" {
		return nil, "", fmt.Errorf("streaming exec only supported for codex agent, got: %s", session.Agent)
	}

	// 准备执行选项
	execOpts := &agent.ExecOptions{
		Prompt:           req.Prompt,
		MaxTurns:         req.MaxTurns,
		Timeout:          req.Timeout,
		IncludeEvents:    true,
		WorkingDirectory: session.Workspace,
	}

	if execOpts.MaxTurns <= 0 {
		execOpts.MaxTurns = 10
	}
	if execOpts.Timeout <= 0 {
		execOpts.Timeout = 300
	}

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
		return nil, "", fmt.Errorf("failed to create execution: %w", err)
	}

	// 准备执行命令
	cmd := adapter.PrepareExec(execOpts)

	// 启动流式执行
	stream, err := m.containerMgr.ExecStream(ctx, session.ContainerID, cmd)
	if err != nil {
		execution.Status = ExecutionFailed
		execution.Error = err.Error()
		now := time.Now()
		execution.EndedAt = &now
		_ = m.store.UpdateExecution(execution)
		return nil, "", fmt.Errorf("failed to start exec stream: %w", err)
	}

	// 创建事件通道
	eventCh := make(chan *StreamEvent, 100)

	// 启动 goroutine 读取输出并解析
	go m.processExecStream(ctx, stream, execution, eventCh)

	return eventCh, execID, nil
}

// processExecStream 处理流式执行输出
func (m *Manager) processExecStream(ctx context.Context, stream *container.ExecStream, execution *Execution, eventCh chan<- *StreamEvent) {
	defer close(eventCh)
	defer stream.Reader.Close()

	// 发送开始事件
	eventCh <- &StreamEvent{
		Type:        "execution.started",
		ExecutionID: execution.ID,
	}

	var outputBuilder strings.Builder
	var lastMessage string
	scanner := bufio.NewScanner(stream.Reader)
	// 增大缓冲区以处理长行
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			eventCh <- &StreamEvent{
				Type:        "execution.cancelled",
				ExecutionID: execution.ID,
				Error:       "context cancelled",
			}
			return
		default:
		}

		line := scanner.Text()
		outputBuilder.WriteString(line)
		outputBuilder.WriteString("\n")

		// 找到 JSON 对象的开始位置 (Codex 输出可能有长度前缀)
		jsonStart := strings.Index(line, "{")
		if jsonStart < 0 {
			continue
		}
		jsonLine := line[jsonStart:]

		// 解析 Codex 事件
		var rawEvent map[string]json.RawMessage
		if err := json.Unmarshal([]byte(jsonLine), &rawEvent); err != nil {
			continue
		}

		// 获取事件类型
		var eventType string
		if typeData, ok := rawEvent["type"]; ok {
			json.Unmarshal(typeData, &eventType)
		}

		// 构建流式事件
		streamEvent := &StreamEvent{
			Type:        eventType,
			ExecutionID: execution.ID,
			Data:        json.RawMessage(jsonLine),
		}

		// 提取 agent_message 文本
		if eventType == "item.completed" {
			if itemData, ok := rawEvent["item"]; ok {
				var item struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				if err := json.Unmarshal(itemData, &item); err == nil {
					if item.Type == "agent_message" && item.Text != "" {
						streamEvent.Text = item.Text
						lastMessage = item.Text
					}
				}
			}
		}

		// 处理错误
		if eventType == "error" || eventType == "turn.failed" {
			if msgData, ok := rawEvent["message"]; ok {
				var msg string
				json.Unmarshal(msgData, &msg)
				streamEvent.Error = msg
			}
			if errData, ok := rawEvent["error"]; ok {
				var errObj struct {
					Message string `json:"message"`
				}
				if err := json.Unmarshal(errData, &errObj); err == nil {
					streamEvent.Error = errObj.Message
				}
			}
		}

		eventCh <- streamEvent
	}

	// 更新执行记录
	now := time.Now()
	execution.EndedAt = &now
	execution.Output = lastMessage
	execution.Status = ExecutionSuccess
	if err := scanner.Err(); err != nil {
		execution.Status = ExecutionFailed
		execution.Error = err.Error()
	}
	_ = m.store.UpdateExecution(execution)

	// 发送完成事件
	eventCh <- &StreamEvent{
		Type:        "execution.completed",
		ExecutionID: execution.ID,
		Text:        lastMessage,
	}
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

// ANSI color codes
const (
	ansiReset      = "\x1b[0m"
	ansiBold       = "\x1b[1m"
	ansiDim        = "\x1b[2m"
	ansiRed        = "\x1b[31m"
	ansiGreen      = "\x1b[32m"
	ansiYellow     = "\x1b[33m"
	ansiBlue       = "\x1b[34m"
	ansiMagenta    = "\x1b[35m"
	ansiCyan       = "\x1b[36m"
	ansiGray       = "\x1b[90m"
	ansiBrightWhite = "\x1b[97m"
)

// GetLogs 获取会话的执行日志（带 ANSI 颜色）
func (m *Manager) GetLogs(ctx context.Context, id string) (string, error) {
	// 验证会话存在
	_, err := m.store.Get(id)
	if err != nil {
		return "", err
	}

	// 从执行记录聚合日志
	executions, err := m.store.ListExecutions(id)
	if err != nil {
		return "", fmt.Errorf("failed to get executions: %w", err)
	}

	if len(executions) == 0 {
		return "", nil
	}

	// 按开始时间排序
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartedAt.Before(executions[j].StartedAt)
	})

	var logs strings.Builder
	for i, exec := range executions {
		if i > 0 {
			logs.WriteString("\n")
		}

		// 状态颜色
		statusColor := ansiGray
		switch exec.Status {
		case ExecutionSuccess:
			statusColor = ansiGreen
		case ExecutionFailed:
			statusColor = ansiRed
		case ExecutionRunning:
			statusColor = ansiYellow
		}

		// 执行头部信息
		logs.WriteString(fmt.Sprintf("%s%s=== Execution %s %s%s[%s]%s %s===%s\n",
			ansiBold, ansiCyan, exec.ID, ansiReset,
			statusColor, exec.Status, ansiReset,
			ansiBold+ansiCyan, ansiReset))
		logs.WriteString(fmt.Sprintf("%sStarted: %s%s\n", ansiGray, exec.StartedAt.Format(time.RFC3339), ansiReset))
		if exec.EndedAt != nil {
			logs.WriteString(fmt.Sprintf("%sEnded: %s%s\n", ansiGray, exec.EndedAt.Format(time.RFC3339), ansiReset))
		}

		// Prompt
		logs.WriteString(fmt.Sprintf("\n%s❯%s %s%s%s\n", ansiGreen, ansiReset, ansiBrightWhite, exec.Prompt, ansiReset))

		// 输出内容
		if exec.Output != "" {
			logs.WriteString(fmt.Sprintf("\n%s--- Output ---%s\n", ansiDim, ansiReset))
			logs.WriteString(exec.Output)
			if !strings.HasSuffix(exec.Output, "\n") {
				logs.WriteString("\n")
			}
		}

		// 错误内容
		if exec.Error != "" {
			logs.WriteString(fmt.Sprintf("\n%s--- Error ---%s\n", ansiRed, ansiReset))
			logs.WriteString(fmt.Sprintf("%s%s%s\n", ansiRed, exec.Error, ansiReset))
		}

		// 退出码
		if exec.ExitCode != 0 {
			logs.WriteString(fmt.Sprintf("\n%sExit Code: %d%s\n", ansiRed, exec.ExitCode, ansiReset))
		}
	}

	return logs.String(), nil
}

// StreamLogs 获取容器实时日志流
func (m *Manager) StreamLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	// 获取会话
	sess, err := m.store.Get(id)
	if err != nil {
		return nil, err
	}

	// 检查容器是否存在
	if sess.ContainerID == "" {
		return nil, fmt.Errorf("session has no container")
	}

	// 获取容器日志流
	return m.containerMgr.Logs(ctx, sess.ContainerID)
}

// inferProviderFromBaseURL 从 BaseURL 推断 Provider 名称
func inferProviderFromBaseURL(baseURL string) string {
	// 常见的 Provider URL 模式
	providerPatterns := map[string][]string{
		"openai":    {"api.openai.com"},
		"azure":     {"azure.com", "openai.azure.com"},
		"deepseek":  {"api.deepseek.com"},
		"zhipu":     {"open.bigmodel.cn", "bigmodel.cn"},
		"qwen":      {"dashscope.aliyuncs.com"},
		"kimi":      {"api.moonshot.cn", "moonshot.cn"},
		"minimax":   {"api.minimax.chat", "api.minimaxi.com"},
		"baichuan":  {"api.baichuan-ai.com"},
		"openrouter": {"openrouter.ai"},
		"together":  {"api.together.xyz"},
		"groq":      {"api.groq.com"},
		"fireworks": {"api.fireworks.ai"},
	}

	baseURLLower := strings.ToLower(baseURL)
	for provider, patterns := range providerPatterns {
		for _, pattern := range patterns {
			if strings.Contains(baseURLLower, pattern) {
				return provider
			}
		}
	}

	// 默认返回 "custom"
	return "custom"
}
