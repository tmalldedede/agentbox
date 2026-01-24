package batch

import (
	"encoding/json"

	"github.com/tmalldedede/agentbox/internal/database"
)

// GormStore implements Store using GORM repositories.
type GormStore struct {
	batchRepo *database.BatchRepository
	taskRepo  *database.BatchTaskRepository
}

// NewGormStore creates a new GORM-based store.
func NewGormStore() *GormStore {
	return &GormStore{
		batchRepo: database.NewBatchRepository(),
		taskRepo:  database.NewBatchTaskRepository(),
	}
}

// CreateBatch creates a new batch.
func (s *GormStore) CreateBatch(batch *Batch) error {
	model := s.batchToModel(batch)
	if err := s.batchRepo.Create(model); err != nil {
		return err
	}
	batch.ID = model.ID
	return nil
}

// GetBatch retrieves a batch by ID.
func (s *GormStore) GetBatch(id string) (*Batch, error) {
	model, err := s.batchRepo.Get(id)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, ErrBatchNotFound
		}
		return nil, err
	}
	return s.modelToBatch(model), nil
}

// UpdateBatch updates an existing batch.
func (s *GormStore) UpdateBatch(batch *Batch) error {
	model := s.batchToModel(batch)
	return s.batchRepo.Update(model)
}

// DeleteBatch deletes a batch and its tasks.
func (s *GormStore) DeleteBatch(id string) error {
	return s.batchRepo.Delete(id)
}

// ListBatches returns batches matching the filter.
func (s *GormStore) ListBatches(filter *ListBatchFilter) ([]*Batch, int, error) {
	var status, agentID, userID string
	var limit, offset int

	if filter != nil {
		status = string(filter.Status)
		agentID = filter.AgentID
		userID = filter.UserID
		limit = filter.Limit
		offset = filter.Offset
	}

	models, total, err := s.batchRepo.List(status, agentID, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	batches := make([]*Batch, len(models))
	for i, m := range models {
		batches[i] = s.modelToBatch(&m)
	}

	return batches, int(total), nil
}

// CreateTasks bulk creates tasks.
func (s *GormStore) CreateTasks(tasks []*BatchTask) error {
	if len(tasks) == 0 {
		return nil
	}

	models := make([]database.BatchTaskModel, len(tasks))
	for i, t := range tasks {
		models[i] = *s.taskToModel(t)
	}

	return s.taskRepo.CreateBulk(models)
}

// GetTask retrieves a single task.
func (s *GormStore) GetTask(batchID, taskID string) (*BatchTask, error) {
	model, err := s.taskRepo.GetByBatchAndID(batchID, taskID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return s.modelToTask(model), nil
}

// UpdateTask updates a task.
func (s *GormStore) UpdateTask(task *BatchTask) error {
	model := s.taskToModel(task)
	return s.taskRepo.Update(model)
}

// ListTasks returns tasks for a batch.
func (s *GormStore) ListTasks(batchID string, filter *ListTaskFilter) ([]*BatchTask, int, error) {
	var status, workerID string
	var limit, offset int

	if filter != nil {
		status = string(filter.Status)
		workerID = filter.WorkerID
		limit = filter.Limit
		offset = filter.Offset
	}

	models, total, err := s.taskRepo.List(batchID, status, workerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	tasks := make([]*BatchTask, len(models))
	for i, m := range models {
		tasks[i] = s.modelToTask(&m)
	}

	return tasks, int(total), nil
}

// DeleteTasks deletes all tasks for a batch.
func (s *GormStore) DeleteTasks(batchID string) error {
	return s.taskRepo.DeleteByBatch(batchID)
}

// ClaimPendingTasks is a legacy method - with Redis queue this is handled differently.
// Kept for SQLite-only fallback mode.
func (s *GormStore) ClaimPendingTasks(batchID string, limit int) ([]*BatchTask, error) {
	// This is only used when Redis is disabled
	// For now, return empty - Redis queue handles claiming
	return nil, nil
}

// RequeueTask puts a task back to pending.
func (s *GormStore) RequeueTask(task *BatchTask) error {
	return s.taskRepo.UpdateStatus(task.ID, string(BatchTaskPending), map[string]interface{}{
		"worker_id":  "",
		"started_at": nil,
		"claimed_at": nil,
		"claimed_by": "",
	})
}

// MarkTaskDead moves a task to dead letter status.
func (s *GormStore) MarkTaskDead(task *BatchTask, reason string) error {
	return s.taskRepo.MarkDead(task.ID, reason)
}

// ListDeadTasks returns dead letter tasks.
func (s *GormStore) ListDeadTasks(batchID string, limit int) ([]*BatchTask, error) {
	models, err := s.taskRepo.ListDead(batchID, limit)
	if err != nil {
		return nil, err
	}

	tasks := make([]*BatchTask, len(models))
	for i, m := range models {
		tasks[i] = s.modelToTask(&m)
	}
	return tasks, nil
}

// RetryDeadTasks retries dead tasks.
func (s *GormStore) RetryDeadTasks(batchID string, taskIDs []string) (int, error) {
	count, err := s.taskRepo.RetryDead(batchID, taskIDs)
	return int(count), err
}

// ResetRunningTasks resets running tasks to pending.
func (s *GormStore) ResetRunningTasks(batchID string) (int, error) {
	count, err := s.taskRepo.ResetRunning(batchID)
	return int(count), err
}

// ListRunningBatches returns batches with running status.
func (s *GormStore) ListRunningBatches() ([]*Batch, error) {
	models, err := s.batchRepo.ListByStatus(string(BatchStatusRunning))
	if err != nil {
		return nil, err
	}

	batches := make([]*Batch, len(models))
	for i, m := range models {
		batches[i] = s.modelToBatch(&m)
	}
	return batches, nil
}

// GetTaskStats returns statistics for a batch.
func (s *GormStore) GetTaskStats(batchID string) (*BatchStats, error) {
	statsMap, err := s.taskRepo.GetStats(batchID)
	if err != nil {
		return nil, err
	}

	stats := &BatchStats{
		ByWorker:   make(map[string]int),
		ErrorTypes: make(map[string]int),
	}

	for status, count := range statsMap {
		stats.TotalTasks += int(count)
		switch BatchTaskStatus(status) {
		case BatchTaskPending:
			stats.Pending = int(count)
		case BatchTaskRunning:
			stats.Running = int(count)
		case BatchTaskCompleted:
			stats.Completed = int(count)
		case BatchTaskFailed:
			stats.Failed = int(count)
		case BatchTaskDead:
			stats.Dead = int(count)
		}
	}

	return stats, nil
}

// Close closes the store (no-op for GORM, connection managed globally).
func (s *GormStore) Close() error {
	return nil
}

// Conversion helpers

func (s *GormStore) batchToModel(b *Batch) *database.BatchModel {
	templateJSON, _ := json.Marshal(b.Template)
	workersJSON, _ := json.Marshal(b.Workers)
	errorSummaryJSON, _ := json.Marshal(b.ErrorSummary)

	return &database.BatchModel{
		BaseModel: database.BaseModel{
			ID:        b.ID,
			CreatedAt: b.CreatedAt,
		},
		UserID:           b.UserID,
		Name:             b.Name,
		AgentID:          b.AgentID,
		TemplateJSON:     string(templateJSON),
		Concurrency:      b.Concurrency,
		Status:           string(b.Status),
		TotalTasks:       b.TotalTasks,
		Completed:        b.Completed,
		Failed:           b.Failed,
		Dead:             0, // Calculated from tasks
		WorkersJSON:      string(workersJSON),
		ErrorSummaryJSON: string(errorSummaryJSON),
		StartedAt:        b.StartedAt,
		CompletedAt:      b.CompletedAt,
	}
}

func (s *GormStore) modelToBatch(m *database.BatchModel) *Batch {
	b := &Batch{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		AgentID:     m.AgentID,
		Concurrency: m.Concurrency,
		Status:      BatchStatus(m.Status),
		TotalTasks:  m.TotalTasks,
		Completed:   m.Completed,
		Failed:      m.Failed,
		CreatedAt:   m.CreatedAt,
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,
	}

	json.Unmarshal([]byte(m.TemplateJSON), &b.Template)
	json.Unmarshal([]byte(m.WorkersJSON), &b.Workers)
	json.Unmarshal([]byte(m.ErrorSummaryJSON), &b.ErrorSummary)

	// Calculate progress
	if b.TotalTasks > 0 {
		b.ProgressPercent = float64(b.Completed+b.Failed) / float64(b.TotalTasks) * 100
	}

	return b
}

func (s *GormStore) taskToModel(t *BatchTask) *database.BatchTaskModel {
	inputJSON, _ := json.Marshal(t.Input)

	return &database.BatchTaskModel{
		BaseModel: database.BaseModel{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
		},
		BatchID:    t.BatchID,
		TaskIndex:  t.Index,
		InputJSON:  string(inputJSON),
		Prompt:     t.Prompt,
		Status:     string(t.Status),
		WorkerID:   t.WorkerID,
		Result:     t.Result,
		Error:      t.Error,
		Attempts:   t.Attempts,
		ClaimedAt:  t.ClaimedAt,
		ClaimedBy:  t.ClaimedBy,
		DeadAt:     t.DeadAt,
		DeadReason: t.DeadReason,
		StartedAt:  t.StartedAt,
		DurationMs: t.DurationMs,
	}
}

func (s *GormStore) modelToTask(m *database.BatchTaskModel) *BatchTask {
	t := &BatchTask{
		ID:         m.ID,
		BatchID:    m.BatchID,
		Index:      m.TaskIndex,
		Prompt:     m.Prompt,
		Status:     BatchTaskStatus(m.Status),
		WorkerID:   m.WorkerID,
		Result:     m.Result,
		Error:      m.Error,
		Attempts:   m.Attempts,
		ClaimedAt:  m.ClaimedAt,
		ClaimedBy:  m.ClaimedBy,
		DeadAt:     m.DeadAt,
		DeadReason: m.DeadReason,
		CreatedAt:  m.CreatedAt,
		StartedAt:  m.StartedAt,
		DurationMs: m.DurationMs,
	}

	json.Unmarshal([]byte(m.InputJSON), &t.Input)

	return t
}

// UpdateCounters atomically updates batch counters
func (s *GormStore) UpdateCounters(batchID string, completed, failed, dead int) error {
	return s.batchRepo.UpdateCounters(batchID, completed, failed, dead)
}

// GetBatchByTask returns the batch for a given task
func (s *GormStore) GetBatchByTask(taskID string) (*Batch, error) {
	task, err := s.taskRepo.Get(taskID)
	if err != nil {
		return nil, err
	}
	return s.GetBatch(task.BatchID)
}
