package agent

import (
	"context"
	"encoding/json"

	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

// ConfigFilesProvider 配置文件提供者接口
// 某些 Agent (如 Codex) 需要在容器中写入配置文件
type ConfigFilesProvider interface {
	// GetConfigFiles 返回需要写入容器的配置文件
	// 返回: map[容器内路径]文件内容
	// apiKey: 用于认证的 API Key
	GetConfigFiles(p *profile.Profile, apiKey string) map[string]string
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

	// PrepareContainer 准备容器配置
	PrepareContainer(session *SessionInfo) *container.CreateConfig

	// PrepareContainerWithProfile 使用 Profile 准备容器配置
	PrepareContainerWithProfile(session *SessionInfo, p *profile.Profile) *container.CreateConfig

	// PrepareExec 准备执行命令
	PrepareExec(req *ExecOptions) []string

	// PrepareExecWithProfile 使用 Profile 准备执行命令
	PrepareExecWithProfile(req *ExecOptions, p *profile.Profile) []string

	// RequiredEnvVars 返回必需的环境变量
	RequiredEnvVars() []string

	// ValidateProfile 验证 Profile 是否与此适配器兼容
	ValidateProfile(p *profile.Profile) error

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

	// Profile-based options (override individual options if set)
	Profile *profile.Profile

	// 工作目录 (由 session manager 设置)
	WorkingDirectory string
}

// ExecResult 执行结果
type ExecResult struct {
	Message  string       `json:"message"`            // Agent 最终回复
	Events   []ExecEvent  `json:"events,omitempty"`   // 完整事件列表 (当 IncludeEvents=true)
	ExitCode int          `json:"exit_code"`          // 退出码
	Usage    *TokenUsage  `json:"usage,omitempty"`    // Token 使用统计
	Error    string       `json:"error,omitempty"`    // 错误信息
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
