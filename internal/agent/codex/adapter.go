package codex

import (
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
		NetworkMode: "bridge", // Codex 需要网络访问 API
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

	sb.WriteString("\n")

	// ===== Provider 配置 =====
	// 如果有自定义 Provider，生成 [model_providers.xxx] 配置块
	if p.Model.Provider != "" && p.Model.BaseURL != "" {
		providerName := strings.ToLower(p.Model.Provider)

		sb.WriteString(fmt.Sprintf("[model_providers.%s]\n", providerName))
		sb.WriteString(fmt.Sprintf("name = \"%s\"\n", p.Model.Provider))
		sb.WriteString(fmt.Sprintf("base_url = \"%s\"\n", p.Model.BaseURL))

		// wire_api - API 协议类型 (chat/responses)
		wireAPI := p.Model.WireAPI
		if wireAPI == "" {
			// 第三方提供商（非 OpenAI）默认使用 chat API
			if strings.ToLower(p.Model.Provider) != "openai" {
				wireAPI = "chat"
			} else {
				wireAPI = "responses"
			}
		}
		sb.WriteString(fmt.Sprintf("wire_api = \"%s\"\n", wireAPI))

		// 认证配置 - 支持三种方式:
		// 1. env_key: 使用指定的环境变量名获取 API Key
		// 2. bearer_token: 直接使用 token（experimental_bearer_token）
		// 3. requires_openai_auth: 使用 OPENAI_API_KEY（默认）
		if p.Model.EnvKey != "" {
			sb.WriteString(fmt.Sprintf("env_key = \"%s\"\n", p.Model.EnvKey))
			sb.WriteString("requires_openai_auth = false\n")
		} else if p.Model.BearerToken != "" {
			sb.WriteString(fmt.Sprintf("experimental_bearer_token = \"%s\"\n", p.Model.BearerToken))
			sb.WriteString("requires_openai_auth = false\n")
		} else {
			sb.WriteString("requires_openai_auth = true\n")
		}

		// 重试配置
		sb.WriteString("request_max_retries = 4\n")
		sb.WriteString("stream_max_retries = 10\n")
		sb.WriteString("stream_idle_timeout_ms = 300000\n")

		sb.WriteString("\n")

		// ===== Profile 配置 =====
		// 生成 [profiles.xxx] 配置块，方便使用 codex -p xxx 切换
		sb.WriteString(fmt.Sprintf("[profiles.%s]\n", providerName))
		sb.WriteString(fmt.Sprintf("model = \"%s\"\n", p.Model.Name))
		sb.WriteString(fmt.Sprintf("model_provider = \"%s\"\n", providerName))
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
// 注意: Codex 容器以 agent 用户运行，配置文件在 /home/agent/.codex/
func (a *Adapter) GetConfigFiles(p *profile.Profile, apiKey string) map[string]string {
	files := make(map[string]string)

	// ~/.codex/config.toml (agent 用户)
	configTOML := a.GenerateConfigTOML(p)
	if configTOML != "" {
		files["/home/agent/.codex/config.toml"] = configTOML
	}

	// ~/.codex/auth.json (agent 用户)
	if apiKey != "" {
		files["/home/agent/.codex/auth.json"] = a.GenerateAuthJSON(apiKey)
	}

	return files
}

// PrepareExec 准备执行命令
func (a *Adapter) PrepareExec(req *agent.ExecOptions) []string {
	// 使用 exec 子命令执行非交互式任务
	// --full-auto 等同于 -a on-request + -s workspace-write
	// --skip-git-repo-check 跳过 git 仓库检查（容器内可能没有 .git 目录）
	args := []string{
		"codex",
		"exec",
		req.Prompt,
		"--full-auto",
		"--skip-git-repo-check",
	}

	// Codex 暂不支持 max_turns 和 tools 参数
	// 未来可以扩展

	return args
}

// PrepareExecWithProfile 使用 Profile 准备执行命令
// 根据 Profile 配置生成完整的 codex CLI 命令
func (a *Adapter) PrepareExecWithProfile(req *agent.ExecOptions, p *profile.Profile) []string {
	// Codex 使用 exec 子命令执行非交互式任务
	// --skip-git-repo-check 跳过 git 仓库检查（容器内可能没有 .git 目录）
	args := []string{"codex", "exec", req.Prompt, "--skip-git-repo-check"}

	// ===== 模型配置 =====
	if p.Model.Name != "" {
		args = append(args, "--model", p.Model.Name)
	}

	// ===== 推理强度 (Codex 特有) =====
	if p.Model.ReasoningEffort != "" {
		args = append(args, "--reasoning-effort", p.Model.ReasoningEffort)
	}

	// ===== 沙箱模式 =====
	if p.Permissions.FullAuto {
		// --full-auto 是 -a on-request + -s workspace-write 的快捷方式
		args = append(args, "--full-auto")
	} else {
		// 审批策略
		if p.Permissions.ApprovalPolicy != "" {
			args = append(args, "--ask-for-approval", p.Permissions.ApprovalPolicy)
		}
		// 沙箱模式
		if p.Permissions.SandboxMode != "" {
			args = append(args, "--sandbox", p.Permissions.SandboxMode)
		}
	}

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
		"custom_provider",
		"env_key",
		"bearer_token",
		"wire_api",
	}
}

// init 自动注册到默认注册表
func init() {
	agent.Register(New())
}
