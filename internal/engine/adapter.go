package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tmalldedede/agentbox/internal/container"
)

// AgentConfig 适配器统一配置
// 由 Session Manager 从 Agent 配置转换而来，传递给适配器使用
type AgentConfig struct {
	// 基本信息
	ID      string `json:"id"`
	Name    string `json:"name"`
	Adapter string `json:"adapter"` // claude-code / codex / opencode

	// 模型配置
	Model ModelConfig `json:"model"`

	// 权限配置
	Permissions PermissionConfig `json:"permissions"`

	// 资源限制
	Resources ResourceConfig `json:"resources"`

	// 提示词
	SystemPrompt       string `json:"system_prompt,omitempty"`
	AppendSystemPrompt string `json:"append_system_prompt,omitempty"`

	// MCP 服务器（已解析）
	MCPServers []MCPServerConfig `json:"mcp_servers,omitempty"`

	// 自定义 Agent 配置
	CustomAgents json.RawMessage `json:"custom_agents,omitempty"`

	// Codex 特有
	BaseInstructions      string            `json:"base_instructions,omitempty"`
	DeveloperInstructions string            `json:"developer_instructions,omitempty"`
	ConfigOverrides       map[string]string `json:"config_overrides,omitempty"`
	OutputSchema          string            `json:"output_schema,omitempty"`

	// 功能开关
	Features FeaturesConfig `json:"features,omitempty"`

	// 输出格式
	OutputFormat string `json:"output_format,omitempty"`

	// 调试
	Debug DebugConfig `json:"debug,omitempty"`
}

// FeaturesConfig 功能开关配置
type FeaturesConfig struct {
	WebSearch bool `json:"web_search,omitempty"` // Codex: --search
}

// ModelConfig 模型配置
type ModelConfig struct {
	Name            string `json:"name"`
	Provider        string `json:"provider,omitempty"`
	BaseURL         string `json:"base_url,omitempty"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`

	// Model tier (Claude Code)
	HaikuModel  string `json:"haiku_model,omitempty"`
	SonnetModel string `json:"sonnet_model,omitempty"`
	OpusModel   string `json:"opus_model,omitempty"`

	// Advanced
	TimeoutMS       int  `json:"timeout_ms,omitempty"`
	MaxOutputTokens int  `json:"max_output_tokens,omitempty"`
	DisableTraffic  bool `json:"disable_traffic,omitempty"`

	// Codex specific
	WireAPI     string `json:"wire_api,omitempty"`
	EnvKey      string `json:"env_key,omitempty"`
	BearerToken string `json:"bearer_token,omitempty"`
}

// PermissionConfig 权限配置
type PermissionConfig struct {
	// Claude Code
	Mode            string   `json:"mode,omitempty"`
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`
	Tools           []string `json:"tools,omitempty"`
	SkipAll         bool     `json:"skip_all,omitempty"`

	// Codex
	SandboxMode    string `json:"sandbox_mode,omitempty"`
	ApprovalPolicy string `json:"approval_policy,omitempty"`
	FullAuto       bool   `json:"full_auto,omitempty"`

	// Common
	AdditionalDirs []string `json:"additional_dirs,omitempty"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	MaxBudgetUSD float64       `json:"max_budget_usd,omitempty"`
	MaxTurns     int           `json:"max_turns,omitempty"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`

	// Container resources
	CPUs     float64 `json:"cpus,omitempty"`
	MemoryMB int     `json:"memory_mb,omitempty"`
}

// MCPServerConfig MCP 服务器配置
type MCPServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// DebugConfig 调试配置
type DebugConfig struct {
	Verbose bool `json:"verbose,omitempty"`
}

// Permission mode constants
const (
	PermissionModeAcceptEdits       = "acceptEdits"
	PermissionModeBypassPermissions = "bypassPermissions"
	PermissionModeDefault           = "default"
	PermissionModeDelegate          = "delegate"
	PermissionModeDontAsk           = "dontAsk"
	PermissionModePlan              = "plan"
)

// Sandbox mode constants (Codex)
const (
	SandboxModeReadOnly         = "read-only"
	SandboxModeWorkspaceWrite   = "workspace-write"
	SandboxModeDangerFullAccess = "danger-full-access"
)

// Approval policy constants (Codex)
const (
	ApprovalPolicyUntrusted = "untrusted"
	ApprovalPolicyOnFailure = "on-failure"
	ApprovalPolicyOnRequest = "on-request"
	ApprovalPolicyNever     = "never"
)

// Adapter name constants
const (
	AdapterClaudeCode = "claude-code"
	AdapterCodex      = "codex"
	AdapterOpenCode   = "opencode"
)

// ConfigFilesProvider 配置文件提供者接口
// 某些 Agent (如 Codex) 需要在容器中写入配置文件
type ConfigFilesProvider interface {
	// GetConfigFiles 返回需要写入容器的配置文件
	// 返回: map[容器内路径]文件内容
	// apiKey: 用于认证的 API Key
	GetConfigFiles(cfg *AgentConfig, apiKey string) map[string]string
}

// Adapter Agent 适配器接口
// 每种 Agent (Claude Code, Codex 等) 需要实现此接口
type Adapter interface {
	// Name 返回 Agent 名称
	Name() string

	// DisplayName 返回显示名称
	DisplayName() string

	// Description 返回描述
	Description() string

	// Image 返回 Docker 镜像
	Image() string

	// PrepareContainer 准备容器配置（无 AgentConfig）
	PrepareContainer(session *SessionInfo) *container.CreateConfig

	// PrepareContainerWithConfig 使用 AgentConfig 准备容器配置
	PrepareContainerWithConfig(session *SessionInfo, cfg *AgentConfig) *container.CreateConfig

	// PrepareExec 准备执行命令（无 AgentConfig）
	PrepareExec(req *ExecOptions) []string

	// PrepareExecWithConfig 使用 AgentConfig 准备执行命令
	PrepareExecWithConfig(req *ExecOptions, cfg *AgentConfig) []string

	// RequiredEnvVars 返回必需的环境变量
	RequiredEnvVars() []string

	// ValidateConfig 验证 AgentConfig 是否与此适配器兼容
	ValidateConfig(cfg *AgentConfig) error

	// SupportedFeatures 返回此适配器支持的功能列表
	SupportedFeatures() []string
}

// ExecOptions 执行选项
type ExecOptions struct {
	Prompt          string
	MaxTurns        int
	Timeout         int
	AllowedTools    []string
	DisallowedTools []string
	IncludeEvents   bool // 是否在结果中包含完整事件列表

	// 多轮对话：ThreadID 用于 resume 上一轮的对话上下文
	// Codex: codex exec resume <ThreadID> "prompt"
	ThreadID string

	// AgentConfig (override individual options if set)
	Config *AgentConfig

	// 工作目录 (由 session manager 设置)
	WorkingDirectory string
}

// ExecResult 执行结果
type ExecResult struct {
	Message  string       `json:"message"`          // Agent 最终回复
	Events   []ExecEvent  `json:"events,omitempty"` // 完整事件列表 (当 IncludeEvents=true)
	ExitCode int          `json:"exit_code"`        // 退出码
	Usage    *TokenUsage  `json:"usage,omitempty"`  // Token 使用统计
	Error    string       `json:"error,omitempty"`  // 错误信息
	ThreadID string       `json:"thread_id,omitempty"` // 多轮对话 Thread ID (从 thread.started 事件中提取)
}

// TokenUsage Token 使用统计
type TokenUsage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens,omitempty"`
	OutputTokens      int `json:"output_tokens"`
}

// ExecEvent 执行事件
type ExecEvent struct {
	Type string          `json:"type"` // 事件类型
	Item json.RawMessage `json:"item,omitempty"`
	Raw  json.RawMessage `json:"raw,omitempty"` // 原始 JSON
}

// DirectExecutor 直接执行接口
// 某些 Agent (如 Codex) 可以实现此接口，使用 Go SDK 直接执行而不是调用 CLI
type DirectExecutor interface {
	// Execute 直接执行 prompt，返回结构化结果
	Execute(ctx context.Context, opts *ExecOptions) (*ExecResult, error)
}

// JSONOutputParser JSONL 输出解析接口
// 支持 --json 输出的 CLI (如 Codex) 应实现此接口
type JSONOutputParser interface {
	// ParseJSONLOutput 解析 JSONL 格式的 CLI 输出
	// output: CLI 的 stdout 内容 (JSONL 格式，每行一个 JSON 对象)
	// includeEvents: 是否在结果中包含完整事件列表
	// 返回: 结构化的执行结果
	ParseJSONLOutput(output string, includeEvents bool) (*ExecResult, error)
}

// SessionInfo 会话信息，传递给适配器
type SessionInfo struct {
	ID        string            // 会话 ID
	Workspace string            // 工作空间路径
	Env       map[string]string // 环境变量
}

// Info Agent 信息
type Info struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	RequiredEnv []string `json:"required_env"`
}

// GetInfo 从适配器获取信息
func GetInfo(a Adapter) Info {
	return Info{
		Name:        a.Name(),
		DisplayName: a.DisplayName(),
		Description: a.Description(),
		Image:       a.Image(),
		RequiredEnv: a.RequiredEnvVars(),
	}
}
