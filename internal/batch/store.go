package batch

import (
	"errors"
)

// Common errors
var (
	ErrBatchNotFound = errors.New("batch not found")
	ErrTaskNotFound  = errors.New("batch task not found")
	ErrBatchRunning  = errors.New("batch is running")
	ErrBatchNotRunning = errors.New("batch is not running")
)

// Store defines the storage interface for batches and batch tasks.
type Store interface {
	// Batch operations
	CreateBatch(batch *Batch) error
	GetBatch(id string) (*Batch, error)
	UpdateBatch(batch *Batch) error
	DeleteBatch(id string) error
	ListBatches(filter *ListBatchFilter) ([]*Batch, int, error)

	// BatchTask operations
	CreateTasks(tasks []*BatchTask) error                                // Bulk create
	GetTask(batchID, taskID string) (*BatchTask, error)
	UpdateTask(task *BatchTask) error
	ListTasks(batchID string, filter *ListTaskFilter) ([]*BatchTask, int, error)
	DeleteTasks(batchID string) error                                    // Delete all tasks for a batch

	// Queue operations (atomic) - Legacy, kept for SQLite fallback
	ClaimPendingTasks(batchID string, limit int) ([]*BatchTask, error)   // Atomically claim pending tasks
	RequeueTask(task *BatchTask) error                                   // Put task back to pending

	// Dead letter operations
	MarkTaskDead(task *BatchTask, reason string) error                   // Move task to dead letter
	ListDeadTasks(batchID string, limit int) ([]*BatchTask, error)       // List dead letter tasks
	RetryDeadTasks(batchID string, taskIDs []string) (int, error)        // Retry dead tasks

	// Recovery operations
	ResetRunningTasks(batchID string) (int, error)                       // Reset running tasks to pending (for recovery)
	ListRunningBatches() ([]*Batch, error)                               // Get batches that were running (for recovery)

	// Counter operations
	UpdateCounters(batchID string, completed, failed, dead int) error    // Atomically update batch counters

	// Statistics
	GetTaskStats(batchID string) (*BatchStats, error)

	// Lifecycle
	Close() error
}
