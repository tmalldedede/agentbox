package session

import (
	"time"
)

// Session 会话
type Session struct {
	ID          string            `json:"id"`
	Agent       string            `json:"agent"`
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
	Agent     string            `json:"agent" binding:"required"`
	Workspace string            `json:"workspace" binding:"required"`
	Env       map[string]string `json:"env"`
	Config    *Config           `json:"config"`
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
	Prompt string `json:"prompt" binding:"required"`
}

// ExecResponse 执行响应
type ExecResponse struct {
	ExecutionID string `json:"execution_id"`
	Output      string `json:"output"`
	ExitCode    int    `json:"exit_code"`
	Error       string `json:"error,omitempty"`
}

// ListFilter 列表过滤器
type ListFilter struct {
	Agent  string `form:"agent"`
	Status Status `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}
