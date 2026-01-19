package agent

import (
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/profile"
)

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

	// Profile-based options (override individual options if set)
	Profile *profile.Profile
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
