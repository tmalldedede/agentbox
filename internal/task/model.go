package task

import (
	"errors"
	"time"
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

// 常见错误
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskExists   = errors.New("task already exists")
	ErrInvalidInput = errors.New("invalid input")
)

// Task 任务定义
type Task struct {
	// 基本信息
	ID          string `json:"id"`
	ProfileID   string `json:"profile_id"`
	ProfileName string `json:"profile_name,omitempty"` // 冗余，方便展示
	AgentType   string `json:"agent_type,omitempty"`   // claude-code / codex

	// 任务内容
	Prompt string `json:"prompt"`
	Input  *Input `json:"input,omitempty"`

	// 输出配置
	Output     *OutputConfig `json:"output,omitempty"`
	WebhookURL string        `json:"webhook_url,omitempty"`

	// 执行配置
	Timeout int `json:"timeout,omitempty"` // 秒，0 表示使用默认

	// 运行时状态
	Status       Status  `json:"status"`
	SessionID    string  `json:"session_id,omitempty"`    // 关联的 Session
	ErrorMessage string  `json:"error_message,omitempty"` // 失败原因
	Result       *Result `json:"result,omitempty"`        // 执行结果

	// 时间戳
	CreatedAt   time.Time  `json:"created_at"`
	QueuedAt    *time.Time `json:"queued_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// 元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Input 任务输入
type Input struct {
	Type   string `json:"type"` // git / files / text
	URL    string `json:"url,omitempty"`
	Branch string `json:"branch,omitempty"`
	Path   string `json:"path,omitempty"`
	Text   string `json:"text,omitempty"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Type   string   `json:"type"`             // files / text / json
	Format string   `json:"format,omitempty"` // 输出格式
	Files  []string `json:"files,omitempty"`  // 指定输出的文件
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
