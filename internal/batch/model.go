// Package batch provides batch task processing with worker pool pattern.
package batch

import (
	"time"
)

// BatchStatus represents the current state of a batch.
type BatchStatus string

const (
	BatchStatusPending   BatchStatus = "pending"
	BatchStatusRunning   BatchStatus = "running"
	BatchStatusPaused    BatchStatus = "paused"
	BatchStatusCompleted BatchStatus = "completed"
	BatchStatusFailed    BatchStatus = "failed"
	BatchStatusCancelled BatchStatus = "cancelled"
)

// BatchTaskStatus represents the current state of a batch task.
type BatchTaskStatus string

const (
	BatchTaskPending   BatchTaskStatus = "pending"
	BatchTaskRunning   BatchTaskStatus = "running"
	BatchTaskCompleted BatchTaskStatus = "completed"
	BatchTaskFailed    BatchTaskStatus = "failed"
	BatchTaskDead      BatchTaskStatus = "dead" // 死信队列
)

// Batch represents a batch of homogeneous tasks sharing the same configuration.
type Batch struct {
	ID     string `json:"id"`                // batch-{uuid}
	UserID string `json:"user_id,omitempty"` // Owner user ID
	Name   string `json:"name"`              // User-defined name

	// Agent configuration
	AgentID string `json:"agent_id"` // Agent to use for execution

	// Template configuration
	Template BatchTemplate `json:"template"`

	// Concurrency configuration
	Concurrency int `json:"concurrency"` // Number of workers

	// Progress tracking
	Status     BatchStatus `json:"status"`
	TotalTasks int         `json:"total_tasks"`
	Completed  int         `json:"completed"`
	Failed     int         `json:"failed"`

	// Computed fields (not stored)
	ProgressPercent float64 `json:"progress_percent,omitempty"`
	EstimatedETA    string  `json:"estimated_eta,omitempty"`
	TasksPerSec     float64 `json:"tasks_per_sec,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Workers info
	Workers []WorkerInfo `json:"workers,omitempty"`

	// Error aggregation
	ErrorSummary map[string]int `json:"error_summary,omitempty"` // error_type -> count
}

// BatchTemplate defines the configuration template for batch tasks.
type BatchTemplate struct {
	PromptTemplate string `json:"prompt_template"` // Supports {{.field}} variables
	Timeout        int    `json:"timeout"`         // Per-task timeout in seconds
	MaxRetries     int    `json:"max_retries"`     // Number of retry attempts
	RuntimeID      string `json:"runtime_id"`      // Optional runtime configuration
}

// WorkerInfo contains runtime information about a worker.
type WorkerInfo struct {
	ID          string `json:"id"`                     // worker-{index}
	SessionID   string `json:"session_id"`             // Associated session ID
	ContainerID string `json:"container_id,omitempty"` // Container ID if running
	Status      string `json:"status"`                 // idle/busy/error/stopped
	CurrentTask string `json:"current_task,omitempty"` // Currently processing task ID
	Completed   int    `json:"completed"`              // Tasks completed by this worker
	LastError   string `json:"last_error,omitempty"`   // Last error if any
}

// BatchTask represents a single task within a batch.
type BatchTask struct {
	ID      string `json:"id"`       // {batch_id}-{index}
	BatchID string `json:"batch_id"` // Parent batch ID
	Index   int    `json:"index"`    // Index within the batch

	// Input data
	Input map[string]interface{} `json:"input"` // Template variables

	// Rendered prompt (computed at runtime)
	Prompt string `json:"prompt,omitempty"`

	// Execution state
	Status   BatchTaskStatus `json:"status"`
	WorkerID string          `json:"worker_id,omitempty"` // Worker that executed/is executing

	// Result
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`

	// Retry tracking
	Attempts int `json:"attempts"`

	// Claim tracking (for checkpoint/resume)
	ClaimedAt *time.Time `json:"claimed_at,omitempty"` // When task was claimed by worker
	ClaimedBy string     `json:"claimed_by,omitempty"` // Worker ID that claimed this task

	// Dead letter tracking
	DeadAt     *time.Time `json:"dead_at,omitempty"`     // When task was moved to dead letter queue
	DeadReason string     `json:"dead_reason,omitempty"` // Reason for dead letter

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	DurationMs int64      `json:"duration_ms,omitempty"` // Execution duration
}

// CreateBatchRequest is the request to create a new batch.
type CreateBatchRequest struct {
	Name           string                   `json:"name"`            // Batch name
	AgentID        string                   `json:"agent_id"`        // Agent to use
	PromptTemplate string                   `json:"prompt_template"` // e.g., "Analyze: {{.data}}"
	Inputs         []map[string]interface{} `json:"inputs"`          // List of input maps
	Concurrency    int                      `json:"concurrency"`     // Number of workers
	Timeout        int                      `json:"timeout"`         // Per-task timeout (seconds)
	MaxRetries     int                      `json:"max_retries"`     // Retry count
	RuntimeID      string                   `json:"runtime_id"`      // Optional runtime
	AutoStart      bool                     `json:"auto_start"`      // Start immediately after creation
	UserID         string                   `json:"-"`               // Injected by middleware
}

// UpdateBatchRequest is the request to update batch settings.
type UpdateBatchRequest struct {
	Name        *string `json:"name,omitempty"`
	Concurrency *int    `json:"concurrency,omitempty"` // Can adjust while paused
}

// ListBatchFilter defines filtering options for listing batches.
type ListBatchFilter struct {
	UserID  string      `json:"user_id,omitempty"`
	Status  BatchStatus `json:"status,omitempty"`
	AgentID string      `json:"agent_id,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	Offset  int         `json:"offset,omitempty"`
}

// ListTaskFilter defines filtering options for listing batch tasks.
type ListTaskFilter struct {
	Status   BatchTaskStatus `json:"status,omitempty"`
	WorkerID string          `json:"worker_id,omitempty"`
	Limit    int             `json:"limit,omitempty"`
	Offset   int             `json:"offset,omitempty"`
}

// BatchStats contains statistics for a batch.
type BatchStats struct {
	TotalTasks  int            `json:"total_tasks"`
	Pending     int            `json:"pending"`
	Running     int            `json:"running"`
	Completed   int            `json:"completed"`
	Failed      int            `json:"failed"`
	Dead        int            `json:"dead"`                   // Dead letter count
	ByWorker    map[string]int `json:"by_worker,omitempty"`    // worker_id -> completed count
	AvgDuration float64        `json:"avg_duration_ms"`        // Average task duration
	ErrorTypes  map[string]int `json:"error_types,omitempty"`  // error type -> count
}

// BatchEvent represents an event during batch execution.
type BatchEvent struct {
	Type      string      `json:"type"`
	BatchID   string      `json:"batch_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Event types
const (
	EventBatchStarted   = "batch.started"
	EventBatchProgress  = "batch.progress"
	EventBatchPaused    = "batch.paused"
	EventBatchResumed   = "batch.resumed"
	EventBatchCompleted = "batch.completed"
	EventBatchFailed    = "batch.failed"
	EventBatchCancelled = "batch.cancelled"
	EventWorkerStarted  = "worker.started"
	EventWorkerStopped  = "worker.stopped"
	EventWorkerError    = "worker.error"
	EventTaskStarted    = "task.started"
	EventTaskCompleted  = "task.completed"
	EventTaskFailed     = "task.failed"
)

// ProgressData is the payload for batch.progress events.
type ProgressData struct {
	Completed   int     `json:"completed"`
	Failed      int     `json:"failed"`
	Total       int     `json:"total"`
	Percent     float64 `json:"percent"`
	ETA         string  `json:"eta"`
	TasksPerSec float64 `json:"tasks_per_sec"`
}

// WorkerEventData is the payload for worker events.
type WorkerEventData struct {
	WorkerID  string `json:"worker_id"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TaskEventData is the payload for task events.
type TaskEventData struct {
	TaskID     string `json:"task_id"`
	TaskIndex  int    `json:"task_index"`
	WorkerID   string `json:"worker_id"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ImportBatchRequest is the request to import tasks from file.
type ImportBatchRequest struct {
	Name           string `json:"name"`            // Batch name
	AgentID        string `json:"agent_id"`        // Agent to use
	PromptTemplate string `json:"prompt_template"` // Prompt template
	Concurrency    int    `json:"concurrency"`     // Number of workers
	Timeout        int    `json:"timeout"`         // Per-task timeout
	MaxRetries     int    `json:"max_retries"`     // Retry count
	RuntimeID      string `json:"runtime_id"`      // Optional runtime
	AutoStart      bool   `json:"auto_start"`      // Start immediately
	// File is uploaded via multipart form
}

// ImportBatchResponse is the response after importing a batch.
type ImportBatchResponse struct {
	Batch       *Batch `json:"batch"`
	TasksLoaded int    `json:"tasks_loaded"`
	Errors      int    `json:"errors"`
	Message     string `json:"message,omitempty"`
}

// RetryDeadRequest is the request to retry dead letter tasks.
type RetryDeadRequest struct {
	TaskIDs []string `json:"task_ids,omitempty"` // Empty = retry all
}

// RetryDeadResponse is the response after retrying dead tasks.
type RetryDeadResponse struct {
	Retried int    `json:"retried"`
	Message string `json:"message,omitempty"`
}

// DeadLetterInfo contains details about a dead letter task.
type DeadLetterInfo struct {
	Task      *BatchTask `json:"task"`
	Attempts  int        `json:"attempts"`
	LastError string     `json:"last_error"`
	DeadAt    time.Time  `json:"dead_at"`
}

// QueueOverview provides a summary of all queues.
type QueueOverview struct {
	TotalPending   int          `json:"total_pending"`
	TotalRunning   int          `json:"total_running"`
	TotalCompleted int          `json:"total_completed"`
	TotalFailed    int          `json:"total_failed"`
	TotalDead      int          `json:"total_dead"`
	Batches        []BatchQueue `json:"batches"`
	RedisEnabled   bool         `json:"redis_enabled"`
}

// BatchQueue contains queue stats for a single batch.
type BatchQueue struct {
	BatchID    string      `json:"batch_id"`
	BatchName  string      `json:"batch_name"`
	Status     BatchStatus `json:"status"`
	Pending    int         `json:"pending"`
	Running    int         `json:"running"`
	Completed  int         `json:"completed"`
	Failed     int         `json:"failed"`
	Dead       int         `json:"dead"`
	Total      int         `json:"total"`

	// Redis queue stats (if Redis enabled)
	RedisPending    int64 `json:"redis_pending,omitempty"`
	RedisProcessing int64 `json:"redis_processing,omitempty"`
	RedisDead       int64 `json:"redis_dead,omitempty"`
}
