package claude

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

const (
	// AgentName Agent 名称
	AgentName = "claude-code"

	// DefaultImage 默认镜像
	DefaultImage = "agentbox/agent:latest"
)

// Adapter Claude Code 适配器
type Adapter struct {
	image string
}

// New 创建 Claude Code 适配器
func New() *Adapter {
	return &Adapter{
		image: DefaultImage,
	}
}

// NewWithImage 使用自定义镜像创建适配器
func NewWithImage(image string) *Adapter {
	return &Adapter{
		image: image,
	}
}

// Name 返回 Agent 名称
func (a *Adapter) Name() string {
	return AgentName
}

// DisplayName 返回显示名称
func (a *Adapter) DisplayName() string {
	return "Claude Code"
}

// Description 返回描述
func (a *Adapter) Description() string {
	return "Anthropic's official CLI for Claude - an agentic coding assistant"
}

// Image 返回 Docker 镜像
func (a *Adapter) Image() string {
	return a.image
}

// PrepareContainer 准备容器配置
func (a *Adapter) PrepareContainer(session *agent.SessionInfo) *container.CreateConfig {
	// 构建环境变量
	env := make(map[string]string)
	for k, v := range session.Env {
		env[k] = v
	}

	return &container.CreateConfig{
		Name:  fmt.Sprintf("agentbox-%s-%s", AgentName, session.ID),
		Image: a.image,
		Cmd:   []string{"sleep", "infinity"}, // 保持容器运行
		Env:   env,
		Mounts: []container.Mount{
			{
				Source:   session.Workspace,
				Target:   "/workspace",
				ReadOnly: false,
			},
		},
		Resources: container.ResourceConfig{
			CPULimit:    2.0,
			MemoryLimit: 4 * 1024 * 1024 * 1024, // 4GB
		},
		NetworkMode: "bridge", // Claude Code 需要网络访问 API
		Labels: map[string]string{
			"agentbox.managed":    "true",
			"agentbox.agent":      AgentName,
			"agentbox.session.id": session.ID,
		},
	}
}

// PrepareContainerWithProfile 使用 Profile 准备容器配置
func (a *Adapter) PrepareContainerWithProfile(session *agent.SessionInfo, p *profile.Profile) *container.CreateConfig {
	config := a.PrepareContainer(session)

	// 应用 Profile 资源限制
	if p.Resources.CPUs > 0 {
		config.Resources.CPULimit = p.Resources.CPUs
	}
	if p.Resources.MemoryMB > 0 {
		config.Resources.MemoryLimit = int64(p.Resources.MemoryMB) * 1024 * 1024
	}

	// 添加 Profile 标签
	config.Labels["agentbox.profile.id"] = p.ID
	config.Labels["agentbox.profile.name"] = p.Name

	// ===== 应用 Model 环境变量配置 =====
	a.applyModelEnvVars(config.Env, &p.Model)

	return config
}

// applyModelEnvVars 根据 ModelConfig 设置环境变量
// 这是 CC-Switch 模式的核心：通过环境变量控制 Claude Code 的 API 配置
func (a *Adapter) applyModelEnvVars(env map[string]string, model *profile.ModelConfig) {
	// ANTHROPIC_BASE_URL - 自定义 API 端点（代理/第三方兼容）
	if model.BaseURL != "" {
		env["ANTHROPIC_BASE_URL"] = model.BaseURL
	}

	// ANTHROPIC_AUTH_TOKEN - 第三方提供商的认证 Token（如智谱 GLM）
	// 优先使用 BearerToken，它在 Claude Code 中对应 ANTHROPIC_AUTH_TOKEN
	if model.BearerToken != "" {
		env["ANTHROPIC_AUTH_TOKEN"] = model.BearerToken
	}

	// ANTHROPIC_MODEL - 默认模型
	if model.Name != "" {
		env["ANTHROPIC_MODEL"] = model.Name
	}

	// 模型层级配置
	// ANTHROPIC_DEFAULT_HAIKU_MODEL - Haiku 层模型
	if model.HaikuModel != "" {
		env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = model.HaikuModel
	}

	// ANTHROPIC_DEFAULT_SONNET_MODEL - Sonnet 层模型
	if model.SonnetModel != "" {
		env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = model.SonnetModel
	}

	// ANTHROPIC_DEFAULT_OPUS_MODEL - Opus 层模型
	if model.OpusModel != "" {
		env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = model.OpusModel
	}

	// 高级配置
	// API_TIMEOUT_MS - 请求超时
	if model.TimeoutMS > 0 {
		env["API_TIMEOUT_MS"] = strconv.Itoa(model.TimeoutMS)
	}

	// CLAUDE_CODE_MAX_OUTPUT_TOKENS - 最大输出 Token
	if model.MaxOutputTokens > 0 {
		env["CLAUDE_CODE_MAX_OUTPUT_TOKENS"] = strconv.Itoa(model.MaxOutputTokens)
	}

	// CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC - 禁用非必要流量
	if model.DisableTraffic {
		env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"
	}
}

// PrepareExec 准备执行命令
func (a *Adapter) PrepareExec(req *agent.ExecOptions) []string {
	args := []string{
		"claude",
		"-p", req.Prompt,
		"--dangerously-skip-permissions",
	}

	// 添加最大轮数
	if req.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}

	// 添加允许的工具
	for _, tool := range req.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}

	// 添加禁用的工具
	for _, tool := range req.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}

	return args
}

// PrepareExecWithProfile 使用 Profile 准备执行命令
// 这是 Profile 系统的核心方法，根据 Profile 配置生成完整的 claude CLI 命令
func (a *Adapter) PrepareExecWithProfile(req *agent.ExecOptions, p *profile.Profile) []string {
	args := []string{"claude", "-p", req.Prompt}

	// ===== 模型配置 =====
	if p.Model.Name != "" {
		args = append(args, "--model", p.Model.Name)
	}

	// ===== 权限模式 =====
	if p.Permissions.SkipAll {
		args = append(args, "--dangerously-skip-permissions")
	} else if p.Permissions.Mode != "" {
		args = append(args, "--permission-mode", p.Permissions.Mode)
	}

	// ===== 工具配置 =====
	// 允许的工具
	for _, tool := range p.Permissions.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}

	// 禁止的工具
	for _, tool := range p.Permissions.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}

	// 指定工具
	for _, tool := range p.Permissions.Tools {
		args = append(args, "--tools", tool)
	}

	// ===== 目录访问 =====
	for _, dir := range p.Permissions.AdditionalDirs {
		args = append(args, "--add-dir", dir)
	}

	// ===== 系统提示词 =====
	if p.SystemPrompt != "" {
		args = append(args, "--system-prompt", p.SystemPrompt)
	}
	if p.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", p.AppendSystemPrompt)
	}

	// ===== MCP 服务器 =====
	if len(p.MCPServers) > 0 {
		// 将 MCP 配置转为 JSON
		mcpConfig := make(map[string]interface{})
		mcpConfig["mcpServers"] = make(map[string]interface{})

		for _, server := range p.MCPServers {
			serverConfig := map[string]interface{}{
				"command": server.Command,
			}
			if len(server.Args) > 0 {
				serverConfig["args"] = server.Args
			}
			if len(server.Env) > 0 {
				serverConfig["env"] = server.Env
			}
			mcpConfig["mcpServers"].(map[string]interface{})[server.Name] = serverConfig
		}

		if configJSON, err := json.Marshal(mcpConfig); err == nil {
			args = append(args, "--mcp-config", string(configJSON))
		}
	}

	// ===== 资源限制 =====
	if p.Resources.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", p.Resources.MaxTurns))
	}
	if p.Resources.MaxBudgetUSD > 0 {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", p.Resources.MaxBudgetUSD))
	}

	// ===== 自定义 Agent =====
	if len(p.CustomAgents) > 0 {
		args = append(args, "--agents", string(p.CustomAgents))
	}

	// ===== 输出格式 =====
	if p.OutputFormat != "" {
		args = append(args, "--output-format", p.OutputFormat)
	}

	// ===== 调试 =====
	if p.Debug.Verbose {
		args = append(args, "--verbose")
	}

	// ===== 请求级别覆盖 =====
	// 请求级别的选项可以覆盖 Profile 设置
	if req.MaxTurns > 0 {
		// 已经在 Profile 中设置，但请求级别可以覆盖
		// 这里需要移除之前添加的 --max-turns 并添加新值
		args = removeArg(args, "--max-turns")
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}

	// 请求级别的工具配置追加到 Profile 配置
	for _, tool := range req.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}
	for _, tool := range req.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}

	return args
}

// removeArg 从参数列表中移除指定参数及其值
func removeArg(args []string, argName string) []string {
	result := make([]string, 0, len(args))
	skip := false
	for _, arg := range args {
		if skip {
			skip = false
			continue
		}
		if arg == argName {
			skip = true
			continue
		}
		result = append(result, arg)
	}
	return result
}

// RequiredEnvVars 返回必需的环境变量
// 注意：可以使用 ANTHROPIC_API_KEY 或 ANTHROPIC_AUTH_TOKEN
// 第三方提供商（如智谱 GLM）通常使用 ANTHROPIC_AUTH_TOKEN
func (a *Adapter) RequiredEnvVars() []string {
	return []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"}
}

// ValidateProfile 验证 Profile 是否与此适配器兼容
func (a *Adapter) ValidateProfile(p *profile.Profile) error {
	if p.Adapter != profile.AdapterClaudeCode {
		return fmt.Errorf("profile adapter %q is not compatible with Claude Code adapter", p.Adapter)
	}

	// 验证权限模式
	validModes := map[string]bool{
		"":                                     true,
		profile.PermissionModeAcceptEdits:      true,
		profile.PermissionModeBypassPermissions: true,
		profile.PermissionModeDefault:          true,
		profile.PermissionModeDelegate:         true,
		profile.PermissionModeDontAsk:          true,
		profile.PermissionModePlan:             true,
	}
	if !validModes[p.Permissions.Mode] {
		return fmt.Errorf("invalid permission mode %q for Claude Code", p.Permissions.Mode)
	}

	// Codex 专有字段不应该在 Claude Code Profile 中设置
	if p.Permissions.SandboxMode != "" {
		return fmt.Errorf("sandbox_mode is a Codex-specific option, not valid for Claude Code")
	}
	if p.Permissions.ApprovalPolicy != "" {
		return fmt.Errorf("approval_policy is a Codex-specific option, not valid for Claude Code")
	}
	if p.Permissions.FullAuto {
		return fmt.Errorf("full_auto is a Codex-specific option, not valid for Claude Code")
	}

	return nil
}

// SupportedFeatures 返回此适配器支持的功能列表
func (a *Adapter) SupportedFeatures() []string {
	return []string{
		"model",
		"permission_mode",
		"allowed_tools",
		"disallowed_tools",
		"tools",
		"additional_dirs",
		"system_prompt",
		"append_system_prompt",
		"mcp_servers",
		"max_turns",
		"max_budget_usd",
		"custom_agents",
		"output_format",
		"verbose",
		// 第三方提供商支持
		"custom_provider",
		"base_url",
		"auth_token",
	}
}

// init 自动注册到默认注册表
func init() {
	agent.Register(New())
}
