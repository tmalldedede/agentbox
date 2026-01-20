package opencode

import (
	"fmt"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

const (
	// AgentName Agent 名称
	AgentName = "opencode"

	// DefaultImage 默认镜像
	DefaultImage = "agentbox/agent:latest"
)

// Adapter OpenCode 适配器
type Adapter struct {
	image string
}

// New 创建 OpenCode 适配器
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
	return "OpenCode"
}

// Description 返回描述
func (a *Adapter) Description() string {
	return "Open-source AI coding agent with multi-model support"
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
		NetworkMode: "bridge", // OpenCode 需要网络访问 API
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

// PrepareExec 准备执行命令
// OpenCode CLI 使用 `opencode run [message..]` 格式
func (a *Adapter) PrepareExec(req *agent.ExecOptions) []string {
	args := []string{
		"opencode",
		"run",           // 子命令
		req.Prompt,      // message 作为位置参数
		"--format", "json", // JSON 输出便于解析
	}

	return args
}

// PrepareExecWithProfile 使用 Profile 准备执行命令
// OpenCode CLI 使用 `opencode run [message..] [options]` 格式
// 支持的选项参考: opencode run --help
func (a *Adapter) PrepareExecWithProfile(req *agent.ExecOptions, p *profile.Profile) []string {
	// 基础命令: opencode run "prompt"
	args := []string{"opencode", "run", req.Prompt}

	// ===== 输出格式 =====
	// --format: default (formatted) 或 json (raw JSON events)
	if p.OutputFormat == "json" || p.OutputFormat == "" {
		args = append(args, "--format", "json")
	} else {
		args = append(args, "--format", "default")
	}

	// ===== 模型配置 =====
	// --model, -m: provider/model 格式 (如 anthropic/claude-sonnet-4-20250514)
	if p.Model.Name != "" {
		// 如果指定了 provider，使用 provider/model 格式
		if p.Model.Provider != "" {
			args = append(args, "--model", p.Model.Provider+"/"+p.Model.Name)
		} else {
			args = append(args, "--model", p.Model.Name)
		}
	}

	// ===== Agent 配置 =====
	// --agent: 指定使用的 OpenCode agent（可通过 ConfigOverrides 传递）
	if agent, ok := p.ConfigOverrides["agent"]; ok && agent != "" {
		args = append(args, "--agent", agent)
	}

	// ===== 调试模式 =====
	if p.Debug.Verbose {
		args = append(args, "--print-logs")
	}

	return args
}

// RequiredEnvVars 返回必需的环境变量
// OpenCode 支持多个 LLM 提供商，至少需要其中一个 API Key
func (a *Adapter) RequiredEnvVars() []string {
	// 返回所有可能的 API Key，用户只需配置其中一个
	return []string{
		"ANTHROPIC_API_KEY",  // Claude
		"OPENAI_API_KEY",     // OpenAI
		"GEMINI_API_KEY",     // Google Gemini
		"GROQ_API_KEY",       // Groq
		"GITHUB_TOKEN",       // GitHub Copilot
	}
}

// ValidateProfile 验证 Profile 是否与此适配器兼容
func (a *Adapter) ValidateProfile(p *profile.Profile) error {
	if p.Adapter != profile.AdapterOpenCode {
		return fmt.Errorf("profile adapter %q is not compatible with OpenCode adapter", p.Adapter)
	}

	// 验证输出格式
	validOutputFormats := map[string]bool{
		"":     true,
		"text": true,
		"json": true,
	}
	if !validOutputFormats[p.OutputFormat] {
		return fmt.Errorf("invalid output format %q for OpenCode, must be 'text' or 'json'", p.OutputFormat)
	}

	// Claude Code 专有字段不应该在 OpenCode Profile 中设置
	if p.Permissions.Mode != "" {
		return fmt.Errorf("permission_mode is a Claude Code-specific option, not valid for OpenCode")
	}
	if p.Permissions.SkipAll {
		return fmt.Errorf("skip_all is a Claude Code-specific option, not valid for OpenCode")
	}
	if len(p.MCPServers) > 0 {
		return fmt.Errorf("mcp_servers is a Claude Code-specific option, not valid for OpenCode")
	}
	if len(p.CustomAgents) > 0 {
		return fmt.Errorf("custom_agents is a Claude Code-specific option, not valid for OpenCode")
	}

	// Codex 专有字段不应该在 OpenCode Profile 中设置
	if p.Permissions.SandboxMode != "" {
		return fmt.Errorf("sandbox_mode is a Codex-specific option, not valid for OpenCode")
	}
	if p.Permissions.ApprovalPolicy != "" {
		return fmt.Errorf("approval_policy is a Codex-specific option, not valid for OpenCode")
	}
	if p.Permissions.FullAuto {
		return fmt.Errorf("full_auto is a Codex-specific option, not valid for OpenCode")
	}

	return nil
}

// SupportedFeatures 返回此适配器支持的功能列表
func (a *Adapter) SupportedFeatures() []string {
	return []string{
		"output_format",   // text/json
		"verbose",         // debug mode
		"multi_model",     // 支持多个 LLM 提供商
		"session_persist", // 会话持久化
		"lsp_integration", // LSP 集成
	}
}

// init 自动注册到默认注册表
func init() {
	agent.Register(New())
}
