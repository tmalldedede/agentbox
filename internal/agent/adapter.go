package agent

import (
	"github.com/tmalldedede/agentbox/internal/container"
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

	// PrepareExec 准备执行命令
	PrepareExec(prompt string) []string

	// RequiredEnvVars 返回必需的环境变量
	RequiredEnvVars() []string
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
