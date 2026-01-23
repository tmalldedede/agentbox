package session

import (
	"encoding/json"
	"time"
)

// Session 会话
type Session struct {
	ID          string            `json:"id"`
	AgentID     string            `json:"agent_id"`      // 引用 Agent
	Agent       string            `json:"agent"`         // 引擎适配器名 (claude-code/codex/opencode)
	Status      Status            `json:"status"`
	Workspace   string            `json:"workspace"`
	ContainerID string            `json:"container_id,omitempty"`
	Config      Config            `json:"config"`
	Env         map[string]string `json:"env,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Status 会话状态
type Status string

const (
	StatusCreating Status = "creating"
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
	StatusError    Status = "error"
)

// Config 会话配置
type Config struct {
	CPULimit    float64 `json:"cpu_limit"`
	MemoryLimit int64   `json:"memory_limit"`
}

// CreateRequest 创建会话请求
type CreateRequest struct {
	AgentID   string            `json:"agent_id"`                     // 通过 AgentID 自动解析所有配置
	Agent     string            `json:"agent,omitempty"`              // 引擎适配器名（AgentID 为空时必填）
	Workspace string            `json:"workspace" binding:"required"`
	Env       map[string]string `json:"env,omitempty"`
	Config    *Config           `json:"config,omitempty"`
}

// Execution 执行记录
type Execution struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	Prompt    string          `json:"prompt"`
	Status    ExecutionStatus `json:"status"`
	Output    string          `json:"output,omitempty"`
	Error     string          `json:"error,omitempty"`
	ExitCode  int             `json:"exit_code"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionPending ExecutionStatus = "pending"
	ExecutionRunning ExecutionStatus = "running"
	ExecutionSuccess ExecutionStatus = "success"
	ExecutionFailed  ExecutionStatus = "failed"
)

// ExecRequest 执行请求
type ExecRequest struct {
	Prompt          string   `json:"prompt" binding:"required"`
	MaxTurns        int      `json:"max_turns,omitempty"`        // 最大对话轮数 (1-100, 默认 10)
	Timeout         int      `json:"timeout,omitempty"`          // 超时秒数 (10-3600, 默认 300)
	AllowedTools    []string `json:"allowed_tools,omitempty"`    // 允许的工具列表
	DisallowedTools []string `json:"disallowed_tools,omitempty"` // 禁用的工具列表
	IncludeEvents   bool     `json:"include_events,omitempty"`   // 是否返回完整事件列表
	ThreadID        string   `json:"thread_id,omitempty"`        // 多轮对话 Thread ID (resume)
}

// ExecResponse 执行响应
type ExecResponse struct {
	ExecutionID string       `json:"execution_id"`
	Message     string       `json:"message"`                // Agent 最终回复 (新增)
	Output      string       `json:"output"`                 // 原始输出 (兼容旧版)
	Events      []ExecEvent  `json:"events,omitempty"`       // 完整事件列表 (当 include_events=true)
	Usage       *TokenUsage  `json:"usage,omitempty"`        // Token 使用统计
	ExitCode    int          `json:"exit_code"`
	Error       string       `json:"error,omitempty"`
	ThreadID    string       `json:"thread_id,omitempty"`    // 多轮对话 Thread ID
}

// TokenUsage Token 使用统计
type TokenUsage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens,omitempty"`
	OutputTokens      int `json:"output_tokens"`
}

// ExecEvent 执行事件
type ExecEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw,omitempty"`
}

// StreamEvent SSE 流式事件
type StreamEvent struct {
	Type        string          `json:"type"`                   // 事件类型
	ExecutionID string          `json:"execution_id,omitempty"` // 执行 ID
	Data        json.RawMessage `json:"data,omitempty"`         // 事件数据
	Text        string          `json:"text,omitempty"`         // 文本内容 (agent_message)
	Error       string          `json:"error,omitempty"`        // 错误信息
}

// ListFilter 列表过滤器
type ListFilter struct {
	Agent  string `form:"agent"`
	Status Status `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}
