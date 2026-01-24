package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MCPServerModel represents an MCP server in the database
type MCPServerModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"size:1024" json:"description"`
	Command     string `gorm:"size:512;not null" json:"command"`
	Args        string `gorm:"type:text" json:"args"`     // JSON array
	Env         string `gorm:"type:text" json:"env"`      // JSON object
	Metadata    string `gorm:"type:text" json:"metadata"` // JSON object
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`
}

func (MCPServerModel) TableName() string {
	return "mcp_servers"
}

// SkillModel represents a skill in the database
type SkillModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Slug        string `gorm:"size:255;uniqueIndex" json:"slug"`
	Description string `gorm:"size:1024" json:"description"`
	Version     string `gorm:"size:64" json:"version"`
	Author      string `gorm:"size:255" json:"author"`
	Tags        string `gorm:"type:text" json:"tags"`       // JSON array
	Triggers    string `gorm:"type:text" json:"triggers"`   // JSON array
	Content     string `gorm:"type:text" json:"content"`    // SKILL.md content
	Files       string `gorm:"type:text" json:"files"`      // JSON array of files
	Scripts     string `gorm:"type:text" json:"scripts"`    // JSON array
	References  string `gorm:"type:text" json:"references"` // JSON array
	Metadata    string `gorm:"type:text" json:"metadata"`   // JSON object
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`
}

func (SkillModel) TableName() string {
	return "skills"
}

// SessionModel represents a session in the database
type SessionModel struct {
	BaseModel
	AgentID     string     `gorm:"size:64;index" json:"agent_id"`
	Agent       string     `gorm:"size:64;not null" json:"agent"`
	Status      string     `gorm:"size:32;not null;index" json:"status"`
	ContainerID string     `gorm:"size:128" json:"container_id"`
	Workspace   string     `gorm:"size:512" json:"workspace"`
	Config      string     `gorm:"type:text" json:"config"` // JSON
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

func (SessionModel) TableName() string {
	return "sessions"
}

// TaskModel represents a task in the database
type TaskModel struct {
	BaseModel
	AgentID     string `gorm:"size:64;index;not null" json:"agent_id"`
	AgentName   string `gorm:"size:255" json:"agent_name"`
	AgentType   string `gorm:"size:64" json:"agent_type"`
	Prompt      string `gorm:"type:text;not null" json:"prompt"`

	// JSON fields
	AttachmentsJSON string `gorm:"type:text" json:"attachments_json"` // []string
	OutputFilesJSON string `gorm:"type:text" json:"output_files_json"` // []OutputFile
	TurnsJSON       string `gorm:"type:text" json:"turns_json"`        // []Turn
	TurnCount       int    `gorm:"default:0" json:"turn_count"`

	// Config
	WebhookURL string `gorm:"size:1024" json:"webhook_url"`
	Timeout    int    `gorm:"default:0" json:"timeout"`

	// Runtime state
	Status       string `gorm:"size:32;not null;index;default:'pending'" json:"status"`
	SessionID    string `gorm:"size:64;index" json:"session_id"`
	ThreadID     string `gorm:"size:128" json:"thread_id"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`
	ResultJSON   string `gorm:"type:text" json:"result_json"` // *Result
	MetadataJSON string `gorm:"type:text" json:"metadata_json"` // map[string]string

	// Timestamps
	QueuedAt    *time.Time `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

// ExecutionModel represents an execution record in the database
type ExecutionModel struct {
	BaseModel
	SessionID   string     `gorm:"size:64;index;not null" json:"session_id"`
	TaskID      string     `gorm:"size:64;index" json:"task_id"`
	Prompt      string     `gorm:"type:text" json:"prompt"`
	Status      string     `gorm:"size:32;not null" json:"status"`
	Output      string     `gorm:"type:text" json:"output"`
	Error       string     `gorm:"type:text" json:"error"`
	ExitCode    int        `json:"exit_code"`
	TokensIn    int        `json:"tokens_in"`
	TokensOut   int        `json:"tokens_out"`
	DurationMs  int64      `json:"duration_ms"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (ExecutionModel) TableName() string {
	return "executions"
}

// WebhookModel represents a webhook in the database
type WebhookModel struct {
	BaseModel
	Name      string `gorm:"size:255;not null" json:"name"`
	URL       string `gorm:"size:1024;not null" json:"url"`
	Secret    string `gorm:"size:255" json:"secret"`
	Events    string `gorm:"type:text" json:"events"` // JSON array
	IsEnabled bool   `gorm:"default:true" json:"is_enabled"`
	Metadata  string `gorm:"type:text" json:"metadata"` // JSON
}

func (WebhookModel) TableName() string {
	return "webhooks"
}

// ImageModel represents a Docker image in the database
type ImageModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Tag         string `gorm:"size:128;not null" json:"tag"`
	Agent       string `gorm:"size:64;index" json:"agent"`
	Digest      string `gorm:"size:128" json:"digest"`
	Size        int64  `json:"size"`
	Status      string `gorm:"size:32" json:"status"`
	Description string `gorm:"size:1024" json:"description"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

func (ImageModel) TableName() string {
	return "images"
}

// HistoryModel represents execution history
type HistoryModel struct {
	BaseModel
	SessionID string `gorm:"size:64;index" json:"session_id"`
	TaskID    string `gorm:"size:64;index" json:"task_id"`
	Action    string `gorm:"size:64" json:"action"`
	Details   string `gorm:"type:text" json:"details"` // JSON
	UserID    string `gorm:"size:64" json:"user_id"`

	SourceType string     `gorm:"size:32;index" json:"source_type"`
	SourceID   string     `gorm:"size:64;index" json:"source_id"`
	SourceName string     `gorm:"size:255" json:"source_name"`
	Engine     string     `gorm:"size:64;index" json:"engine"`
	Status     string     `gorm:"size:32;index" json:"status"`
	Prompt     string     `gorm:"type:text" json:"prompt"`
	Output     string     `gorm:"type:text" json:"output"`
	Error      string     `gorm:"type:text" json:"error"`
	ExitCode   int        `json:"exit_code"`
	Usage      string     `gorm:"type:text" json:"usage"`    // JSON
	Metadata   string     `gorm:"type:text" json:"metadata"` // JSON
	StartedAt  time.Time  `gorm:"index" json:"started_at"`
	EndedAt    *time.Time `json:"ended_at"`
}

func (HistoryModel) TableName() string {
	return "history"
}

// BatchModel represents a batch job in the database
type BatchModel struct {
	BaseModel
	Name             string     `gorm:"size:255;not null" json:"name"`
	AgentID          string     `gorm:"size:64;index;not null" json:"agent_id"`
	TemplateJSON     string     `gorm:"type:text" json:"template_json"`       // JSON
	Concurrency      int        `gorm:"default:5" json:"concurrency"`
	Status           string     `gorm:"size:32;not null;index" json:"status"`
	TotalTasks       int        `json:"total_tasks"`
	Completed        int        `json:"completed"`
	Failed           int        `json:"failed"`
	Dead             int        `json:"dead"`
	WorkersJSON      string     `gorm:"type:text" json:"workers_json"`        // JSON
	ErrorSummaryJSON string     `gorm:"type:text" json:"error_summary_json"`  // JSON
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

func (BatchModel) TableName() string {
	return "batches"
}

// BatchTaskModel represents a task within a batch
type BatchTaskModel struct {
	BaseModel
	BatchID    string     `gorm:"size:64;index;not null" json:"batch_id"`
	TaskIndex  int        `gorm:"index" json:"task_index"`
	InputJSON  string     `gorm:"type:text" json:"input_json"`      // JSON
	Prompt     string     `gorm:"type:text" json:"prompt"`
	Status     string     `gorm:"size:32;not null;index" json:"status"`
	WorkerID   string     `gorm:"size:64" json:"worker_id"`
	Result     string     `gorm:"type:text" json:"result"`
	Error      string     `gorm:"type:text" json:"error"`
	Attempts   int        `gorm:"default:0" json:"attempts"`
	ClaimedAt  *time.Time `json:"claimed_at"`
	ClaimedBy  string     `gorm:"size:64" json:"claimed_by"`
	DeadAt     *time.Time `json:"dead_at"`
	DeadReason string     `gorm:"size:512" json:"dead_reason"`
	StartedAt  *time.Time `json:"started_at"`
	DurationMs int64      `json:"duration_ms"`
}

func (BatchTaskModel) TableName() string {
	return "batch_tasks"
}

// FileModel represents an uploaded file in the database
type FileModel struct {
	BaseModel
	Name      string     `gorm:"size:512;not null" json:"name"`
	Size      int64      `gorm:"default:0" json:"size"`
	MimeType  string     `gorm:"size:128;not null;default:'application/octet-stream'" json:"mime_type"`
	Path      string     `gorm:"size:1024;not null" json:"path"`
	TaskID    string     `gorm:"size:64;index" json:"task_id"`
	Purpose   string     `gorm:"size:32;not null;default:'general'" json:"purpose"`
	Status    string     `gorm:"size:32;not null;index;default:'active'" json:"status"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"`
}

func (FileModel) TableName() string {
	return "files"
}
