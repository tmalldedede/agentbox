package claude

import (
	"fmt"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
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

// PrepareExec 准备执行命令
func (a *Adapter) PrepareExec(prompt string) []string {
	return []string{
		"claude",
		"-p", prompt,
		"--dangerously-skip-permissions",
	}
}

// RequiredEnvVars 返回必需的环境变量
func (a *Adapter) RequiredEnvVars() []string {
	return []string{"ANTHROPIC_API_KEY"}
}

// init 自动注册到默认注册表
func init() {
	agent.Register(New())
}
