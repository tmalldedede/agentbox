package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/profile"
)

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
		fmt.Printf("Warning: failed to write config files: %v\n", err)
	} else {
		fmt.Printf("Config files written successfully for session %s\n", sessionID)
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
		fmt.Printf("[writeConfigFiles] Adapter does not implement ConfigFilesProvider\n")
		return nil // 适配器不需要配置文件
	}
	fmt.Printf("[writeConfigFiles] Adapter implements ConfigFilesProvider, profile_id=%s\n", req.ProfileID)

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
	fmt.Printf("[writeConfigFiles] Profile Model: name=%s, provider=%s, base_url=%s\n", p.Model.Name, p.Model.Provider, p.Model.BaseURL)
	fmt.Printf("[writeConfigFiles] API Key present: %v\n", apiKey != "")
	configFiles := cfgProvider.GetConfigFiles(p, apiKey)
	fmt.Printf("[writeConfigFiles] Got %d config files\n", len(configFiles))
	for path := range configFiles {
		fmt.Printf("[writeConfigFiles] - File: %s\n", path)
	}
	if len(configFiles) == 0 {
		fmt.Printf("[writeConfigFiles] No config files to write, returning early\n")
		return nil
	}

	// 通过 exec 命令写入每个配置文件
	for path, content := range configFiles {
		fmt.Printf("[writeConfigFiles] Writing file: %s (len=%d)\n", path, len(content))

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
		fmt.Printf("[writeConfigFiles] Exec command: mkdir -p %s && cat > %s ...\n", dir, expandedPath)
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
