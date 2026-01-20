package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

const (
	// AgentName Agent 名称
	AgentName = "codex"

	// DefaultImage 默认镜像
	DefaultImage = "agentbox/agent:latest"
)

// Adapter Codex 适配器
type Adapter struct {
	image string
}

// New 创建 Codex 适配器
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
	return "OpenAI Codex"
}

// Description 返回描述
func (a *Adapter) Description() string {
	return "OpenAI Codex CLI - lightweight coding agent"
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
		NetworkMode: "bridge",      // Codex 需要网络访问 API
		Privileged:  true,          // Codex landlock 需要特权模式
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

	return config
}

// GenerateConfigTOML 生成 Codex 的 config.toml 内容
// 参考 Codex 源码: codex-rs/core/src/config.rs
func (a *Adapter) GenerateConfigTOML(p *profile.Profile) string {
	var sb strings.Builder

	// ===== 基础配置 =====
	// model_provider - 模型提供商
	if p.Model.Provider != "" {
		sb.WriteString(fmt.Sprintf("model_provider = \"%s\"\n", p.Model.Provider))
	}

	// model - 默认模型
	if p.Model.Name != "" {
		sb.WriteString(fmt.Sprintf("model = \"%s\"\n", p.Model.Name))
	}

	// model_reasoning_effort - 推理强度 (low/medium/high)
	if p.Model.ReasoningEffort != "" {
		sb.WriteString(fmt.Sprintf("model_reasoning_effort = \"%s\"\n", p.Model.ReasoningEffort))
	}

	// disable_response_storage - 禁用响应存储 (隐私)
	sb.WriteString("disable_response_storage = true\n")

	// sandbox_mode - 沙箱模式
	// Docker 容器环境不支持 landlock，必须使用 danger-full-access 禁用沙箱
	// 参考: codex-rs/core/src/sandboxing/mod.rs - SandboxType::None
	sb.WriteString("sandbox_mode = \"danger-full-access\"\n")

	sb.WriteString("\n")

	// ===== Provider 配置 =====
	// 如果有自定义 Provider，生成 [model_providers.xxx] 配置块
	if p.Model.Provider != "" && p.Model.BaseURL != "" {
		providerName := strings.ToLower(p.Model.Provider)

		sb.WriteString(fmt.Sprintf("[model_providers.%s]\n", providerName))
		sb.WriteString(fmt.Sprintf("name = \"%s\"\n", p.Model.Provider))
		sb.WriteString(fmt.Sprintf("base_url = \"%s\"\n", p.Model.BaseURL))

		// wire_api - API 协议类型 (chat/responses)
		// - "responses": OpenAI 官方新版 Responses API
		// - "chat": OpenAI 兼容的 Chat Completions API（第三方 Provider 使用）
		wireAPI := p.Model.WireAPI
		if wireAPI == "" {
			// 只有官方 OpenAI 使用 responses API，其他第三方都用 chat API
			if providerName == "openai" {
				wireAPI = "responses"
			} else {
				wireAPI = "chat"
			}
		}
		sb.WriteString(fmt.Sprintf("wire_api = \"%s\"\n", wireAPI))

		// requires_openai_auth - 是否需要 OpenAI OAuth 认证
		// - true: 需要 OpenAI OAuth 登录，凭证存储在 auth.json
		// - false: API Key 从 env_key 环境变量获取（第三方 Provider 使用）
		if providerName == "openai" {
			sb.WriteString("requires_openai_auth = true\n")
		} else {
			sb.WriteString("requires_openai_auth = false\n")
			sb.WriteString("env_key = \"OPENAI_API_KEY\"\n")
		}

		// 重试配置
		sb.WriteString("request_max_retries = 4\n")
		sb.WriteString("stream_max_retries = 10\n")
		sb.WriteString("stream_idle_timeout_ms = 300000\n")
	}

	return sb.String()
}

// GenerateAuthJSON 生成 Codex 的 auth.json 内容
// 参考 Codex 源码: codex-rs/core/src/config.rs
func (a *Adapter) GenerateAuthJSON(apiKey string) string {
	// auth.json 格式: {"OPENAI_API_KEY": "sk-xxx"}
	return fmt.Sprintf(`{"OPENAI_API_KEY": "%s"}`, apiKey)
}

// GetConfigFiles 返回需要挂载到容器的配置文件
// 返回: map[容器内路径]文件内容
// 注意: 使用 ~ 前缀表示需要在运行时展开为用户 home 目录
func (a *Adapter) GetConfigFiles(p *profile.Profile, apiKey string) map[string]string {
	files := make(map[string]string)

	// ~/.codex/config.toml
	configTOML := a.GenerateConfigTOML(p)
	if configTOML != "" {
		files["~/.codex/config.toml"] = configTOML
	}

	// ~/.codex/auth.json
	// 只有官方 OpenAI 需要 auth.json（requires_openai_auth = true）
	// 第三方 Provider 通过环境变量 OPENAI_API_KEY 传入（requires_openai_auth = false）
	providerName := strings.ToLower(p.Model.Provider)
	if apiKey != "" && providerName == "openai" {
		files["~/.codex/auth.json"] = a.GenerateAuthJSON(apiKey)
	}

	return files
}

// PrepareExec 准备执行命令
func (a *Adapter) PrepareExec(req *agent.ExecOptions) []string {
	// 使用 exec 子命令执行非交互式任务
	// --dangerously-bypass-approvals-and-sandbox 绕过审批和沙箱（Docker 容器不支持 landlock）
	// --skip-git-repo-check 跳过 git 仓库检查（容器内可能没有 .git 目录）
	// --json 输出 JSONL 格式，便于解析
	args := []string{
		"codex",
		"exec",
		req.Prompt,
		"--dangerously-bypass-approvals-and-sandbox",
		"--skip-git-repo-check",
		"--json",
	}

	return args
}

// PrepareExecWithProfile 使用 Profile 准备执行命令
// 根据 Profile 配置生成完整的 codex CLI 命令
func (a *Adapter) PrepareExecWithProfile(req *agent.ExecOptions, p *profile.Profile) []string {
	// Codex 使用 exec 子命令执行非交互式任务
	// --dangerously-bypass-approvals-and-sandbox 绕过审批和沙箱（Docker 容器不支持 landlock）
	// --skip-git-repo-check 跳过 git 仓库检查（容器内可能没有 .git 目录）
	// --json 输出 JSONL 格式，便于解析
	args := []string{
		"codex", "exec", req.Prompt,
		"--dangerously-bypass-approvals-and-sandbox",
		"--skip-git-repo-check",
		"--json",
	}

	// ===== 模型配置 =====
	if p.Model.Name != "" {
		args = append(args, "--model", p.Model.Name)
	}

	// ===== 推理强度 (Codex 特有) =====
	if p.Model.ReasoningEffort != "" {
		args = append(args, "--reasoning-effort", p.Model.ReasoningEffort)
	}

	// 注意: 沙箱模式和审批策略已通过 --dangerously-bypass-approvals-and-sandbox 禁用
	// Docker 容器环境不支持 landlock，必须使用此参数

	// ===== 目录访问 =====
	for _, dir := range p.Permissions.AdditionalDirs {
		args = append(args, "--add-dir", dir)
	}

	// ===== 指令配置 =====
	if p.BaseInstructions != "" {
		args = append(args, "--base-instructions", p.BaseInstructions)
	}
	if p.DeveloperInstructions != "" {
		args = append(args, "--developer-instructions", p.DeveloperInstructions)
	}

	// ===== 功能开关 =====
	if p.Features.WebSearch {
		args = append(args, "--search")
	}

	// ===== 资源限制 =====
	if p.Resources.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", p.Resources.MaxTokens))
	}
	if p.Resources.Timeout > 0 {
		args = append(args, "--timeout", p.Resources.Timeout.String())
	}

	// ===== 配置覆盖 =====
	for key, value := range p.ConfigOverrides {
		args = append(args, "--config", fmt.Sprintf("%s=%s", key, value))
	}

	// ===== 输出配置 =====
	if p.OutputSchema != "" {
		args = append(args, "--output-schema", p.OutputSchema)
	}

	return args
}

// RequiredEnvVars 返回必需的环境变量
func (a *Adapter) RequiredEnvVars() []string {
	return []string{"OPENAI_API_KEY"}
}

// ValidateProfile 验证 Profile 是否与此适配器兼容
func (a *Adapter) ValidateProfile(p *profile.Profile) error {
	if p.Adapter != profile.AdapterCodex {
		return fmt.Errorf("profile adapter %q is not compatible with Codex adapter", p.Adapter)
	}

	// 验证沙箱模式
	validSandboxModes := map[string]bool{
		"":                                  true,
		profile.SandboxModeReadOnly:         true,
		profile.SandboxModeWorkspaceWrite:   true,
		profile.SandboxModeDangerFullAccess: true,
	}
	if !validSandboxModes[p.Permissions.SandboxMode] {
		return fmt.Errorf("invalid sandbox mode %q for Codex", p.Permissions.SandboxMode)
	}

	// 验证审批策略
	validApprovalPolicies := map[string]bool{
		"":                             true,
		profile.ApprovalPolicyUntrusted: true,
		profile.ApprovalPolicyOnFailure: true,
		profile.ApprovalPolicyOnRequest: true,
		profile.ApprovalPolicyNever:     true,
	}
	if !validApprovalPolicies[p.Permissions.ApprovalPolicy] {
		return fmt.Errorf("invalid approval policy %q for Codex", p.Permissions.ApprovalPolicy)
	}

	// 验证推理强度
	validReasoningEfforts := map[string]bool{
		"":       true,
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if !validReasoningEfforts[p.Model.ReasoningEffort] {
		return fmt.Errorf("invalid reasoning effort %q for Codex", p.Model.ReasoningEffort)
	}

	// Claude Code 专有字段不应该在 Codex Profile 中设置
	if p.Permissions.Mode != "" {
		return fmt.Errorf("permission_mode is a Claude Code-specific option, not valid for Codex")
	}
	if p.Permissions.SkipAll {
		return fmt.Errorf("skip_all is a Claude Code-specific option, not valid for Codex")
	}
	if p.SystemPrompt != "" {
		return fmt.Errorf("system_prompt is a Claude Code-specific option, use base_instructions for Codex")
	}
	if len(p.CustomAgents) > 0 {
		return fmt.Errorf("custom_agents is a Claude Code-specific option, not valid for Codex")
	}

	return nil
}

// SupportedFeatures 返回此适配器支持的功能列表
func (a *Adapter) SupportedFeatures() []string {
	return []string{
		"model",
		"reasoning_effort",
		"sandbox_mode",
		"approval_policy",
		"full_auto",
		"additional_dirs",
		"base_instructions",
		"developer_instructions",
		"web_search",
		"max_tokens",
		"timeout",
		"config_overrides",
		"output_schema",
	}
}

// NOTE: DirectExecutor 接口已禁用
// codex.Run() 在宿主机进程中执行，而不是容器内，无法使用容器的环境变量和配置
// 改为使用 CLI 执行 + --json 输出格式，在 execViaCLI 中解析 JSONL 输出
//
// func (a *Adapter) Execute(ctx context.Context, opts *agent.ExecOptions) (*agent.ExecResult, error) {
//     ... 保留代码供将来本地执行场景使用 ...
// }

// CodexEvent Codex CLI --json 输出的事件结构
type CodexEvent struct {
	Type    string          `json:"type"`              // 事件类型: thread.started, turn.completed, item.completed, error
	Usage   *CodexUsage     `json:"usage,omitempty"`   // Token 使用统计 (turn.completed)
	Item    json.RawMessage `json:"item,omitempty"`    // 事件内容 (item.completed)
	Error   *CodexError     `json:"error,omitempty"`   // 错误信息 (error, turn.failed)
	Message string          `json:"message,omitempty"` // 错误消息 (error)
}

// CodexUsage Token 使用统计
type CodexUsage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens,omitempty"`
	OutputTokens      int `json:"output_tokens"`
}

// CodexError 错误信息
type CodexError struct {
	Message string `json:"message"`
}

// CodexItem item.completed 事件中的 item
type CodexItem struct {
	Type     string `json:"type"`                // item 类型: agent_message, command_execution, reasoning, etc.
	Text     string `json:"text,omitempty"`      // agent_message 的文本内容
	ExitCode *int   `json:"exit_code,omitempty"` // command_execution 的退出码
}

// ParseJSONLOutput 实现 JSONOutputParser 接口
// 解析 Codex CLI --json 输出的 JSONL 格式
func (a *Adapter) ParseJSONLOutput(output string, includeEvents bool) (*agent.ExecResult, error) {
	var (
		message  string
		events   []agent.ExecEvent
		usage    *agent.TokenUsage
		exitCode int
		execErr  string
	)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Codex --json 输出可能包含长度前缀，需要找到 JSON 对象的开始位置
		// 例如: "135{\"type\":\"item.completed\",...}"
		jsonStart := strings.Index(line, "{")
		if jsonStart < 0 {
			continue
		}
		line = line[jsonStart:]

		var event CodexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// 跳过无法解析的行
			continue
		}

		// 如果需要包含事件，记录原始 JSON
		if includeEvents {
			events = append(events, agent.ExecEvent{
				Type: event.Type,
				Raw:  json.RawMessage(line),
			})
		}

		// 处理不同事件类型
		switch event.Type {
		case "turn.completed":
			// 获取 token 使用统计
			if event.Usage != nil {
				usage = &agent.TokenUsage{
					InputTokens:       event.Usage.InputTokens,
					CachedInputTokens: event.Usage.CachedInputTokens,
					OutputTokens:      event.Usage.OutputTokens,
				}
			}

		case "turn.failed":
			if event.Error != nil {
				execErr = event.Error.Message
			}

		case "item.completed":
			// 解析 item 内容
			if len(event.Item) > 0 {
				var item CodexItem
				if err := json.Unmarshal(event.Item, &item); err == nil {
					switch item.Type {
					case "agent_message":
						// 提取 agent_message 作为最终结果
						message = item.Text
					case "command_execution":
						// 提取命令执行的退出码
						if item.ExitCode != nil {
							exitCode = *item.ExitCode
						}
					}
				}
			}

		case "error":
			execErr = event.Message
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan output: %w", err)
	}

	result := &agent.ExecResult{
		Message:  message,
		ExitCode: exitCode,
		Usage:    usage,
		Error:    execErr,
	}

	if includeEvents {
		result.Events = events
	}

	return result, nil
}

// init 自动注册到默认注册表
func init() {
	agent.Register(New())
}
