package task

import (
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// 任务状态
type Status string

const (
	StatusPending   Status = "pending"   // 刚创建，等待入队
	StatusQueued    Status = "queued"    // 已入队，等待调度
	StatusRunning   Status = "running"   // 容器启动，Agent 执行中
	StatusCompleted Status = "completed" // 执行成功
	StatusFailed    Status = "failed"    // 执行失败
	StatusCancelled Status = "cancelled" // 用户取消
)

// 常见错误 - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrTaskNotFound = apperr.NotFound("task")
	ErrTaskExists   = apperr.AlreadyExists("task")
	ErrInvalidInput = apperr.Validation("invalid input")
)

// Task 任务定义
type Task struct {
	// 基本信息
	ID        string `json:"id"`
	UserID    string `json:"user_id,omitempty"`    // 归属用户
	AgentID   string `json:"agent_id"`             // 引用 Agent
	AgentName string `json:"agent_name,omitempty"` // 冗余，方便展示
	AgentType string `json:"agent_type,omitempty"` // claude-code / codex

	// 任务内容
	Prompt string `json:"prompt"`

	// 附件和输出文件
	Attachments []string     `json:"attachments,omitempty"`  // 输入文件 IDs
	OutputFiles []OutputFile `json:"output_files,omitempty"` // 产出文件

	// 多轮对话
	Turns     []Turn `json:"turns,omitempty"`  // 对话轮次记录
	TurnCount int    `json:"turn_count"`       // 轮次计数

	// 输出配置
	WebhookURL string `json:"webhook_url,omitempty"`

	// 执行配置
	Timeout int `json:"timeout,omitempty"` // 秒，0 表示使用默认

	// 运行时状态
	Status       Status  `json:"status"`
	SessionID    string  `json:"session_id,omitempty"`    // 关联的 Session
	ThreadID     string  `json:"thread_id,omitempty"`     // 多轮对话 Thread ID (Codex resume)
	ErrorMessage string  `json:"error_message,omitempty"` // 失败原因
	Result       *Result `json:"result,omitempty"`        // 执行结果（最后一轮）

	// 时间戳
	CreatedAt   time.Time  `json:"created_at"`
	QueuedAt    *time.Time `json:"queued_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// 元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Turn 对话轮次
type Turn struct {
	ID        string    `json:"id"`
	Prompt    string    `json:"prompt"`
	Result    *Result   `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Result 执行结果
type Result struct {
	Files   []OutputFile `json:"files,omitempty"`
	Text    string       `json:"text,omitempty"`
	Logs    string       `json:"logs,omitempty"`
	Usage   *Usage       `json:"usage,omitempty"`
	Summary string       `json:"summary,omitempty"`
}

// OutputFile 输出文件
type OutputFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type,omitempty"`
	URL      string `json:"url,omitempty"` // 下载地址
}

// Usage 资源使用统计
type Usage struct {
	DurationSeconds int   `json:"duration_seconds"`
	InputTokens     int64 `json:"input_tokens,omitempty"`
	OutputTokens    int64 `json:"output_tokens,omitempty"`
	TotalTokens     int64 `json:"total_tokens,omitempty"`
}

// IsTerminal 是否是终止状态
func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// CanCancel 是否可以取消
func (t *Task) CanCancel() bool {
	return t.Status == StatusPending || t.Status == StatusQueued || t.Status == StatusRunning
}

// Duration 获取执行时长
func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil {
		return 0
	}
	end := time.Now()
	if t.CompletedAt != nil {
		end = *t.CompletedAt
	}
	return end.Sub(*t.StartedAt)
}
