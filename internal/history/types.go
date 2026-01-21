package history

import (
	"time"
)

// SourceType 执行来源类型
type SourceType string

const (
	SourceSession SourceType = "session" // Session 执行
	SourceAgent   SourceType = "agent"   // SmartAgent 执行
)

// Status 执行状态
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Entry 统一执行历史记录
type Entry struct {
	ID          string            `json:"id"`
	SourceType  SourceType        `json:"source_type"`  // "session" or "agent"
	SourceID    string            `json:"source_id"`    // session_id or agent_id
	SourceName  string            `json:"source_name"`  // Display name
	ProfileID   string            `json:"profile_id,omitempty"`
	ProfileName string            `json:"profile_name,omitempty"`
	Engine      string            `json:"engine,omitempty"` // claude-code, codex, etc.
	Prompt      string            `json:"prompt"`
	Status      Status            `json:"status"`
	Output      string            `json:"output,omitempty"`
	Error       string            `json:"error,omitempty"`
	ExitCode    int               `json:"exit_code"`
	Usage       *UsageInfo        `json:"usage,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	StartedAt   time.Time         `json:"started_at"`
	EndedAt     *time.Time        `json:"ended_at,omitempty"`
}

// UsageInfo Token 使用统计
type UsageInfo struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens,omitempty"`
	OutputTokens      int `json:"output_tokens"`
}

// Duration 返回执行时长
func (e *Entry) Duration() time.Duration {
	if e.EndedAt == nil {
		return time.Since(e.StartedAt)
	}
	return e.EndedAt.Sub(e.StartedAt)
}

// ListFilter 列表过滤器
type ListFilter struct {
	SourceType SourceType `form:"source_type"` // 按来源类型过滤
	SourceID   string     `form:"source_id"`   // 按来源ID过滤
	ProfileID  string     `form:"profile_id"`  // 按 Profile 过滤
	Engine     string     `form:"engine"`      // 按引擎过滤
	Status     Status     `form:"status"`      // 按状态过滤
	Limit      int        `form:"limit"`       // 分页大小
	Offset     int        `form:"offset"`      // 分页偏移
}

// Stats 执行统计
type Stats struct {
	TotalExecutions   int            `json:"total_executions"`
	CompletedCount    int            `json:"completed_count"`
	FailedCount       int            `json:"failed_count"`
	TotalInputTokens  int            `json:"total_input_tokens"`
	TotalOutputTokens int            `json:"total_output_tokens"`
	BySource          map[string]int `json:"by_source"`
	ByEngine          map[string]int `json:"by_engine"`
}
