package batch

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/session"
)

// Manager orchestrates batch execution with worker pools.
type Manager struct {
	store      Store
	sessionMgr *session.Manager
	agentMgr   *agent.Manager

	// Redis queue (optional, nil if disabled)
	redisQueue *RedisQueue

	// Running batches
	running map[string]*runningBatch
	mu      sync.RWMutex

	// Event subscribers
	eventSubs map[string][]chan *BatchEvent
	eventMu   sync.RWMutex

	// Configuration
	maxBatches       int           // Maximum concurrent batches
	pollInterval     time.Duration // Task polling interval
	progressInterval time.Duration // Progress update interval

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// runningBatch holds runtime state for an active batch.
type runningBatch struct {
	batch     *Batch
	workers   []*worker
	taskQueue chan *BatchTask
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Progress tracking
	startTime     time.Time
	completedLock sync.Mutex
	recentTasks   []time.Time // Track recent completions for rate calculation

	// Completion flag to prevent duplicate completeBatch calls
	completing bool
}

// worker represents a single worker processing tasks.
type worker struct {
	id        string
	sessionID string
	status    string // idle, busy, error, stopped
	cancel    context.CancelFunc
	completed int
	lastError string
}

// ManagerConfig holds configuration for the batch manager.
type ManagerConfig struct {
	MaxBatches       int
	PollInterval     time.Duration
	ProgressInterval time.Duration
	RedisQueue       *RedisQueue // Optional Redis queue
}

// DefaultManagerConfig returns default configuration.
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxBatches:       10,
		PollInterval:     100 * time.Millisecond,
		ProgressInterval: 1 * time.Second,
	}
}

// NewManager creates a new batch manager.
func NewManager(store Store, sessionMgr *session.Manager, agentMgr *agent.Manager, cfg *ManagerConfig) *Manager {
	if cfg == nil {
		cfg = DefaultManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		store:            store,
		sessionMgr:       sessionMgr,
		agentMgr:         agentMgr,
		redisQueue:       cfg.RedisQueue,
		running:          make(map[string]*runningBatch),
		eventSubs:        make(map[string][]chan *BatchEvent),
		maxBatches:       cfg.MaxBatches,
		pollInterval:     cfg.PollInterval,
		progressInterval: cfg.ProgressInterval,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start Redis recovery loop if enabled
	if m.redisQueue != nil {
		m.redisQueue.StartRecoveryLoop(30*time.Second, m.getRunningBatchIDs)
	}

	// Recover interrupted batches on startup
	go m.recoverOnStartup()

	return m
}

// recoverOnStartup recovers batches that were running when the server stopped.
func (m *Manager) recoverOnStartup() {
	// Wait a bit for initialization
	time.Sleep(2 * time.Second)

	batches, err := m.store.ListRunningBatches()
	if err != nil {
		logger.Warn("Failed to list running batches for recovery", "error", err)
		return
	}

	if len(batches) == 0 {
		return
	}

	logger.Info("Recovering interrupted batches", "count", len(batches))

	for _, b := range batches {
		// Reset running tasks to pending
		count, err := m.store.ResetRunningTasks(b.ID)
		if err != nil {
			logger.Warn("Failed to reset running tasks", "batch_id", b.ID, "error", err)
			continue
		}

		if count > 0 {
			logger.Info("Reset running tasks for batch", "batch_id", b.ID, "count", count)
		}

		// Re-enqueue to Redis if enabled
		if m.redisQueue != nil {
			// Get pending tasks and re-enqueue
			tasks, _, err := m.store.ListTasks(b.ID, &ListTaskFilter{
				Status: BatchTaskPending,
				Limit:  100000,
			})
			if err != nil {
				logger.Warn("Failed to list pending tasks for recovery", "batch_id", b.ID, "error", err)
				continue
			}

			if len(tasks) > 0 {
				if err := m.redisQueue.Enqueue(context.Background(), b.ID, tasks); err != nil {
					logger.Warn("Failed to re-enqueue tasks to Redis", "batch_id", b.ID, "error", err)
				} else {
					logger.Info("Re-enqueued tasks to Redis", "batch_id", b.ID, "count", len(tasks))
				}
			}
		}

		// Update batch status to paused (user needs to manually resume)
		b.Status = BatchStatusPaused
		if err := m.store.UpdateBatch(b); err != nil {
			logger.Warn("Failed to update batch status", "batch_id", b.ID, "error", err)
		}

		logger.Info("Batch recovered and paused", "batch_id", b.ID)
	}
}

// Create creates a new batch with tasks.
func (m *Manager) Create(req *CreateBatchRequest) (*Batch, error) {
	// Validate agent exists
	if _, err := m.agentMgr.Get(req.AgentID); err != nil {
		return nil, fmt.Errorf("agent not found: %s", req.AgentID)
	}

	// Validate inputs
	if len(req.Inputs) == 0 {
		return nil, fmt.Errorf("inputs cannot be empty")
	}

	// Validate template
	if req.PromptTemplate == "" {
		return nil, fmt.Errorf("prompt_template cannot be empty")
	}
	if _, err := template.New("test").Parse(req.PromptTemplate); err != nil {
		return nil, fmt.Errorf("invalid prompt_template: %w", err)
	}

	// Set defaults
	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 300 // 5 minutes
	}
	maxRetries := req.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	// Create batch
	batchID := "batch-" + uuid.New().String()[:8]
	now := time.Now()

	batch := &Batch{
		ID:          batchID,
		UserID:      req.UserID,
		Name:        req.Name,
		AgentID:     req.AgentID,
		Template: BatchTemplate{
			PromptTemplate: req.PromptTemplate,
			Timeout:        timeout,
			MaxRetries:     maxRetries,
			RuntimeID:      req.RuntimeID,
		},
		Concurrency:  concurrency,
		Status:       BatchStatusPending,
		TotalTasks:   len(req.Inputs),
		Completed:    0,
		Failed:       0,
		CreatedAt:    now,
		Workers:      []WorkerInfo{},
		ErrorSummary: make(map[string]int),
	}

	if err := m.store.CreateBatch(batch); err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	// Create tasks
	tasks := make([]*BatchTask, len(req.Inputs))
	for i, input := range req.Inputs {
		tasks[i] = &BatchTask{
			ID:        fmt.Sprintf("%s-%d", batchID, i),
			BatchID:   batchID,
			Index:     i,
			Input:     input,
			Status:    BatchTaskPending,
			Attempts:  0,
			CreatedAt: now,
		}
	}

	if err := m.store.CreateTasks(tasks); err != nil {
		// Rollback batch
		m.store.DeleteBatch(batchID)
		return nil, fmt.Errorf("failed to create tasks: %w", err)
	}

	// Enqueue to Redis if enabled
	if m.redisQueue != nil {
		if err := m.redisQueue.Enqueue(context.Background(), batchID, tasks); err != nil {
			logger.Warn("Failed to enqueue tasks to Redis", "batch_id", batchID, "error", err)
			// Continue anyway - SQLite store is the source of truth
		}
	}

	logger.Info("Created batch", "batch_id", batchID, "tasks", len(tasks))

	// Auto-start if requested
	if req.AutoStart {
		if err := m.Start(batchID); err != nil {
			logger.Warn("Failed to auto-start batch", "batch_id", batchID, "error", err)
		}
	}

	return batch, nil
}

// Start begins batch execution by creating workers.
func (m *Manager) Start(batchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, ok := m.running[batchID]; ok {
		return fmt.Errorf("batch %s is already running", batchID)
	}

	// Check max batches
	if len(m.running) >= m.maxBatches {
		return fmt.Errorf("maximum concurrent batches (%d) reached", m.maxBatches)
	}

	batch, err := m.store.GetBatch(batchID)
	if err != nil {
		return err
	}

	if batch.Status != BatchStatusPending && batch.Status != BatchStatusPaused {
		return fmt.Errorf("batch status is %s, cannot start", batch.Status)
	}

	// Create running batch context
	ctx, cancel := context.WithCancel(m.ctx)
	rb := &runningBatch{
		batch:     batch,
		workers:   make([]*worker, 0, batch.Concurrency),
		taskQueue: make(chan *BatchTask, batch.Concurrency*2),
		cancel:    cancel,
		startTime: time.Now(),
	}

	// Create workers
	for i := 0; i < batch.Concurrency; i++ {
		w, err := m.createWorker(ctx, batch, i)
		if err != nil {
			// Cleanup created workers
			cancel()
			for _, existingWorker := range rb.workers {
				m.stopWorker(existingWorker)
			}
			return fmt.Errorf("failed to create worker %d: %w", i, err)
		}
		rb.workers = append(rb.workers, w)
	}

	m.running[batchID] = rb

	// Update batch status
	now := time.Now()
	batch.Status = BatchStatusRunning
	batch.StartedAt = &now
	batch.Workers = m.buildWorkerInfo(rb.workers)
	if err := m.store.UpdateBatch(batch); err != nil {
		logger.Warn("Failed to update batch status", "error", err)
	}

	// Start workers
	for _, w := range rb.workers {
		rb.wg.Add(1)
		go m.runWorker(ctx, rb, w)
	}

	// Start task dispatcher
	go m.runTaskDispatcher(ctx, rb)

	// Start progress reporter
	go m.runProgressReporter(ctx, rb)

	// Broadcast start event
	m.broadcast(batchID, &BatchEvent{
		Type:      EventBatchStarted,
		BatchID:   batchID,
		Timestamp: time.Now(),
	})

	logger.Info("Started batch", "batch_id", batchID, "workers", batch.Concurrency)
	return nil
}

// createWorker creates a new worker with its session.
func (m *Manager) createWorker(ctx context.Context, batch *Batch, index int) (*worker, error) {
	workerID := fmt.Sprintf("worker-%d", index)

	// Verify agent exists
	if _, err := m.agentMgr.GetFullConfig(batch.AgentID); err != nil {
		return nil, fmt.Errorf("failed to get agent config: %w", err)
	}

	// Create session for worker
	// Note: RuntimeID is resolved via AgentID configuration
	sessionReq := &session.CreateRequest{
		AgentID: batch.AgentID,
	}

	sess, err := m.sessionMgr.Create(ctx, sessionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Start session
	if err := m.sessionMgr.Start(ctx, sess.ID); err != nil {
		m.sessionMgr.Delete(ctx, sess.ID)
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	workerCtx, workerCancel := context.WithCancel(ctx)

	w := &worker{
		id:        workerID,
		sessionID: sess.ID,
		status:    "idle",
		cancel:    workerCancel,
	}

	// Broadcast worker started
	m.broadcast(batch.ID, &BatchEvent{
		Type:      EventWorkerStarted,
		BatchID:   batch.ID,
		Timestamp: time.Now(),
		Data: WorkerEventData{
			WorkerID:  workerID,
			SessionID: sess.ID,
		},
	})

	logger.Info("Created worker", "worker_id", workerID, "session_id", sess.ID, "batch_id", batch.ID)

	// Keep context reference (not used directly but good for future)
	_ = workerCtx

	return w, nil
}

// stopWorker stops a worker and cleans up its session.
func (m *Manager) stopWorker(w *worker) {
	w.cancel()
	w.status = "stopped"

	// Stop and delete session
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.sessionMgr.Stop(ctx, w.sessionID); err != nil {
		logger.Warn("Failed to stop session", "session_id", w.sessionID, "error", err)
	}

	if err := m.sessionMgr.Delete(ctx, w.sessionID); err != nil {
		logger.Warn("Failed to delete session", "session_id", w.sessionID, "error", err)
	}

	logger.Info("Stopped worker", "worker_id", w.id, "session_id", w.sessionID)
}

// runWorker is the main worker loop.
func (m *Manager) runWorker(ctx context.Context, rb *runningBatch, w *worker) {
	defer rb.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-rb.taskQueue:
			if !ok {
				return
			}
			m.executeTask(ctx, rb, w, task)
		}
	}
}

// executeTask executes a single task.
func (m *Manager) executeTask(ctx context.Context, rb *runningBatch, w *worker, task *BatchTask) {
	startTime := time.Now()
	w.status = "busy"

	// Update task state
	task.Status = BatchTaskRunning
	task.WorkerID = w.id
	task.StartedAt = &startTime
	m.store.UpdateTask(task)

	// Broadcast task started
	m.broadcast(rb.batch.ID, &BatchEvent{
		Type:      EventTaskStarted,
		BatchID:   rb.batch.ID,
		Timestamp: time.Now(),
		Data: TaskEventData{
			TaskID:    task.ID,
			TaskIndex: task.Index,
			WorkerID:  w.id,
		},
	})

	// Render prompt
	prompt, err := m.renderPrompt(rb.batch.Template.PromptTemplate, task.Input)
	if err != nil {
		m.handleTaskError(rb, w, task, startTime, fmt.Errorf("template error: %w", err))
		return
	}
	task.Prompt = prompt

	// Execute via session
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(rb.batch.Template.Timeout)*time.Second)
	defer cancel()

	result, err := m.sessionMgr.Exec(execCtx, w.sessionID, &session.ExecRequest{
		Prompt: prompt,
	})

	duration := time.Since(startTime).Milliseconds()
	task.DurationMs = duration

	if err != nil {
		m.handleTaskError(rb, w, task, startTime, err)
		return
	}

	// Success
	task.Status = BatchTaskCompleted
	task.Result = result.Message
	if task.Result == "" {
		task.Result = result.Output // Fallback to raw output
	}
	if err := m.store.UpdateTask(task); err != nil {
		logger.Warn("Failed to update task", "task_id", task.ID, "error", err)
	}

	// Mark complete in Redis
	if m.redisQueue != nil {
		if err := m.redisQueue.Complete(context.Background(), rb.batch.ID, task.ID); err != nil {
			logger.Warn("Failed to complete task in Redis", "task_id", task.ID, "error", err)
		}
	}

	// Update counters
	m.incrementCompleted(rb)
	w.completed++
	w.status = "idle"

	// Track completion time for rate calculation
	rb.completedLock.Lock()
	rb.recentTasks = append(rb.recentTasks, time.Now())
	// Keep only last 100 for rate calculation
	if len(rb.recentTasks) > 100 {
		rb.recentTasks = rb.recentTasks[len(rb.recentTasks)-100:]
	}
	rb.completedLock.Unlock()

	// Broadcast task completed
	m.broadcast(rb.batch.ID, &BatchEvent{
		Type:      EventTaskCompleted,
		BatchID:   rb.batch.ID,
		Timestamp: time.Now(),
		Data: TaskEventData{
			TaskID:     task.ID,
			TaskIndex:  task.Index,
			WorkerID:   w.id,
			DurationMs: duration,
		},
	})

	// Check if batch is complete
	m.checkBatchComplete(rb)
}

// handleTaskError handles task failure and retry logic.
func (m *Manager) handleTaskError(rb *runningBatch, w *worker, task *BatchTask, startTime time.Time, err error) {
	task.DurationMs = time.Since(startTime).Milliseconds()
	task.Error = err.Error()
	task.Attempts++

	w.lastError = err.Error()

	// Check if should retry
	if task.Attempts < rb.batch.Template.MaxRetries {
		// Requeue for retry
		task.Status = BatchTaskPending
		task.WorkerID = ""
		task.StartedAt = nil
		if err := m.store.RequeueTask(task); err != nil {
			logger.Warn("Failed to requeue task", "task_id", task.ID, "error", err)
		}

		// Requeue in Redis
		if m.redisQueue != nil {
			if err := m.redisQueue.Requeue(context.Background(), rb.batch.ID, task.ID, task.Attempts); err != nil {
				logger.Warn("Failed to requeue task in Redis", "task_id", task.ID, "error", err)
			}
		}

		logger.Info("Requeuing task for retry", "task_id", task.ID, "attempt", task.Attempts, "max_retries", rb.batch.Template.MaxRetries)
	} else {
		// Max retries exceeded - move to dead letter queue
		task.Status = BatchTaskDead
		reason := fmt.Sprintf("max_retries_exceeded: %s", task.Error)

		if err := m.store.MarkTaskDead(task, reason); err != nil {
			logger.Warn("Failed to mark task as dead", "task_id", task.ID, "error", err)
		}

		// Move to dead in Redis
		if m.redisQueue != nil {
			if err := m.redisQueue.MoveToDead(context.Background(), rb.batch.ID, task.ID, task.Attempts, task.Error); err != nil {
				logger.Warn("Failed to move task to dead in Redis", "task_id", task.ID, "error", err)
			}
		}

		m.incrementDead(rb, task.Error)

		// Broadcast task dead
		m.broadcast(rb.batch.ID, &BatchEvent{
			Type:      EventTaskFailed,
			BatchID:   rb.batch.ID,
			Timestamp: time.Now(),
			Data: TaskEventData{
				TaskID:     task.ID,
				TaskIndex:  task.Index,
				WorkerID:   w.id,
				DurationMs: task.DurationMs,
				Error:      "DEAD: " + task.Error,
			},
		})

		logger.Warn("Task moved to dead letter queue", "task_id", task.ID, "attempts", task.Attempts, "error", task.Error)

		// Check if batch is complete
		m.checkBatchComplete(rb)
	}

	w.status = "idle"
}

// renderPrompt renders the template with input variables.
func (m *Manager) renderPrompt(templateStr string, input map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, input); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// incrementCompleted safely increments the completed counter.
func (m *Manager) incrementCompleted(rb *runningBatch) {
	rb.completedLock.Lock()
	rb.batch.Completed++
	rb.completedLock.Unlock()

	// Persist periodically (every 10 completions)
	if rb.batch.Completed%10 == 0 {
		if err := m.store.UpdateBatch(rb.batch); err != nil {
			logger.Warn("Failed to persist batch progress", "error", err)
		}
	}
}

// incrementFailed safely increments the failed counter.
func (m *Manager) incrementFailed(rb *runningBatch, errMsg string) {
	rb.completedLock.Lock()
	rb.batch.Failed++
	// Aggregate error types
	if rb.batch.ErrorSummary == nil {
		rb.batch.ErrorSummary = make(map[string]int)
	}
	// Truncate error for grouping
	key := errMsg
	if len(key) > 50 {
		key = key[:50] + "..."
	}
	rb.batch.ErrorSummary[key]++
	rb.completedLock.Unlock()
}

// incrementDead safely increments the dead counter (dead letter queue).
func (m *Manager) incrementDead(rb *runningBatch, errMsg string) {
	rb.completedLock.Lock()
	// Update error summary with DEAD prefix
	if rb.batch.ErrorSummary == nil {
		rb.batch.ErrorSummary = make(map[string]int)
	}
	key := "DEAD: " + errMsg
	if len(key) > 50 {
		key = key[:50] + "..."
	}
	rb.batch.ErrorSummary[key]++
	rb.completedLock.Unlock()

	// Update dead counter in store atomically
	if err := m.store.UpdateCounters(rb.batch.ID, 0, 0, 1); err != nil {
		logger.Warn("Failed to update dead counter", "batch_id", rb.batch.ID, "error", err)
	}
}

// checkBatchComplete checks if all tasks are done (including dead letter tasks).
func (m *Manager) checkBatchComplete(rb *runningBatch) {
	// Get stats to include dead count
	stats, err := m.store.GetTaskStats(rb.batch.ID)
	if err != nil {
		logger.Warn("Failed to get task stats for completion check", "batch_id", rb.batch.ID, "error", err)
		return
	}

	rb.completedLock.Lock()
	total := rb.batch.TotalTasks
	// Include completed, failed, and dead tasks
	done := stats.Completed + stats.Failed + stats.Dead
	alreadyCompleting := rb.completing
	if done >= total && !alreadyCompleting {
		rb.completing = true
	}
	rb.completedLock.Unlock()

	if done >= total && !alreadyCompleting {
		// Use goroutine to avoid deadlock: completeBatch waits for workers,
		// but this function is called from within a worker's executeTask
		go m.completeBatch(rb)
	}
}

// completeBatch marks batch as complete and cleans up.
func (m *Manager) completeBatch(rb *runningBatch) {
	m.mu.Lock()
	delete(m.running, rb.batch.ID)
	m.mu.Unlock()

	// Cancel context to stop workers
	rb.cancel()

	// Wait for workers to finish
	rb.wg.Wait()

	// Stop all workers
	for _, w := range rb.workers {
		m.stopWorker(w)
	}

	// Update batch status
	now := time.Now()
	if rb.batch.Failed > 0 && rb.batch.Completed == 0 {
		rb.batch.Status = BatchStatusFailed
	} else {
		rb.batch.Status = BatchStatusCompleted
	}
	rb.batch.CompletedAt = &now
	rb.batch.Workers = nil // Clear workers info
	if err := m.store.UpdateBatch(rb.batch); err != nil {
		logger.Warn("Failed to update batch status", "error", err)
	}

	// Broadcast completion
	eventType := EventBatchCompleted
	if rb.batch.Status == BatchStatusFailed {
		eventType = EventBatchFailed
	}
	m.broadcast(rb.batch.ID, &BatchEvent{
		Type:      eventType,
		BatchID:   rb.batch.ID,
		Timestamp: time.Now(),
	})

	logger.Info("Batch completed", "batch_id", rb.batch.ID, "succeeded", rb.batch.Completed, "failed", rb.batch.Failed)
}

// runTaskDispatcher dispatches pending tasks to the queue.
func (m *Manager) runTaskDispatcher(ctx context.Context, rb *runningBatch) {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	workerID := fmt.Sprintf("batch-%s", rb.batch.ID)

	for {
		select {
		case <-ctx.Done():
			close(rb.taskQueue)
			return
		case <-ticker.C:
			var tasks []*BatchTask

			// Try Redis first if enabled
			if m.redisQueue != nil {
				claimedItems, err := m.redisQueue.Claim(ctx, rb.batch.ID, workerID, rb.batch.Concurrency)
				if err != nil {
					logger.Warn("Failed to claim tasks from Redis", "batch_id", rb.batch.ID, "error", err)
				} else if len(claimedItems) > 0 {
					// Fetch task details from store
					for _, item := range claimedItems {
						task, err := m.store.GetTask(rb.batch.ID, item.TaskID)
						if err != nil {
							logger.Warn("Failed to get task from store", "task_id", item.TaskID, "error", err)
							continue
						}
						tasks = append(tasks, task)
					}
				}
			}

			// Fallback to SQLite polling if Redis not enabled or returned no tasks
			if len(tasks) == 0 {
				var err error
				tasks, err = m.store.ClaimPendingTasks(rb.batch.ID, rb.batch.Concurrency)
				if err != nil {
					logger.Warn("Failed to claim tasks from store", "batch_id", rb.batch.ID, "error", err)
					continue
				}
			}

			// Dispatch to queue
			for _, task := range tasks {
				select {
				case rb.taskQueue <- task:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// runProgressReporter periodically broadcasts progress updates.
func (m *Manager) runProgressReporter(ctx context.Context, rb *runningBatch) {
	ticker := time.NewTicker(m.progressInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			progress := m.calculateProgress(rb)
			m.broadcast(rb.batch.ID, &BatchEvent{
				Type:      EventBatchProgress,
				BatchID:   rb.batch.ID,
				Timestamp: time.Now(),
				Data:      progress,
			})
		}
	}
}

// calculateProgress calculates current progress and ETA.
func (m *Manager) calculateProgress(rb *runningBatch) *ProgressData {
	rb.completedLock.Lock()
	completed := rb.batch.Completed
	failed := rb.batch.Failed
	total := rb.batch.TotalTasks
	recentTasks := make([]time.Time, len(rb.recentTasks))
	copy(recentTasks, rb.recentTasks)
	rb.completedLock.Unlock()

	done := completed + failed
	percent := float64(done) / float64(total) * 100

	// Calculate rate from recent completions
	var tasksPerSec float64
	if len(recentTasks) >= 2 {
		duration := recentTasks[len(recentTasks)-1].Sub(recentTasks[0]).Seconds()
		if duration > 0 {
			tasksPerSec = float64(len(recentTasks)) / duration
		}
	}

	// Calculate ETA
	var eta string
	remaining := total - done
	if tasksPerSec > 0 && remaining > 0 {
		seconds := float64(remaining) / tasksPerSec
		eta = formatDuration(time.Duration(seconds) * time.Second)
	} else {
		eta = "calculating..."
	}

	return &ProgressData{
		Completed:   completed,
		Failed:      failed,
		Total:       total,
		Percent:     percent,
		ETA:         eta,
		TasksPerSec: tasksPerSec,
	}
}

// formatDuration formats duration as human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// buildWorkerInfo builds worker info slice for batch.
func (m *Manager) buildWorkerInfo(workers []*worker) []WorkerInfo {
	info := make([]WorkerInfo, len(workers))
	for i, w := range workers {
		info[i] = WorkerInfo{
			ID:        w.id,
			SessionID: w.sessionID,
			Status:    w.status,
			Completed: w.completed,
			LastError: w.lastError,
		}
	}
	return info
}

// Pause pauses a running batch.
func (m *Manager) Pause(batchID string) error {
	m.mu.Lock()
	rb, ok := m.running[batchID]
	if !ok {
		m.mu.Unlock()
		return ErrBatchNotRunning
	}
	delete(m.running, batchID)
	m.mu.Unlock()

	// Cancel context to stop workers (but don't delete sessions)
	rb.cancel()
	rb.wg.Wait()

	// Stop workers but keep sessions for resume
	for _, w := range rb.workers {
		w.status = "stopped"
		// Note: Not stopping sessions here to allow resume
	}

	// Update status
	rb.batch.Status = BatchStatusPaused
	rb.batch.Workers = m.buildWorkerInfo(rb.workers)
	if err := m.store.UpdateBatch(rb.batch); err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	m.broadcast(batchID, &BatchEvent{
		Type:      EventBatchPaused,
		BatchID:   batchID,
		Timestamp: time.Now(),
	})

	logger.Info("Paused batch", "batch_id", batchID)
	return nil
}

// Resume resumes a paused batch.
func (m *Manager) Resume(batchID string) error {
	return m.Start(batchID)
}

// Cancel cancels a batch and cleans up.
func (m *Manager) Cancel(batchID string) error {
	m.mu.Lock()
	rb, ok := m.running[batchID]
	if ok {
		delete(m.running, batchID)
		m.mu.Unlock()

		// Cancel and cleanup
		rb.cancel()
		rb.wg.Wait()

		for _, w := range rb.workers {
			m.stopWorker(w)
		}
	} else {
		m.mu.Unlock()
	}

	// Update status
	batch, err := m.store.GetBatch(batchID)
	if err != nil {
		return err
	}

	now := time.Now()
	batch.Status = BatchStatusCancelled
	batch.CompletedAt = &now
	batch.Workers = nil
	if err := m.store.UpdateBatch(batch); err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	m.broadcast(batchID, &BatchEvent{
		Type:      EventBatchCancelled,
		BatchID:   batchID,
		Timestamp: time.Now(),
	})

	logger.Info("Cancelled batch", "batch_id", batchID)
	return nil
}

// getRunningBatchIDs returns the IDs of all currently running batches.
// Used by Redis queue recovery loop to skip tasks from active batches.
func (m *Manager) getRunningBatchIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.running))
	for id := range m.running {
		ids = append(ids, id)
	}
	return ids
}

// Get retrieves a batch by ID.
func (m *Manager) Get(batchID string) (*Batch, error) {
	batch, err := m.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}

	// Add runtime info for running batches
	m.mu.RLock()
	if rb, ok := m.running[batchID]; ok {
		batch.Workers = m.buildWorkerInfo(rb.workers)
		progress := m.calculateProgress(rb)
		batch.ProgressPercent = progress.Percent
		batch.EstimatedETA = progress.ETA
		batch.TasksPerSec = progress.TasksPerSec
	}
	m.mu.RUnlock()

	return batch, nil
}

// List lists batches with filtering.
func (m *Manager) List(filter *ListBatchFilter) ([]*Batch, int, error) {
	return m.store.ListBatches(filter)
}

// Delete deletes a batch.
func (m *Manager) Delete(batchID string) error {
	// Cancel if running
	m.Cancel(batchID)
	return m.store.DeleteBatch(batchID)
}

// GetTask retrieves a single task.
func (m *Manager) GetTask(batchID, taskID string) (*BatchTask, error) {
	return m.store.GetTask(batchID, taskID)
}

// ListTasks lists tasks for a batch.
func (m *Manager) ListTasks(batchID string, filter *ListTaskFilter) ([]*BatchTask, int, error) {
	return m.store.ListTasks(batchID, filter)
}

// GetStats returns statistics for a batch.
func (m *Manager) GetStats(batchID string) (*BatchStats, error) {
	return m.store.GetTaskStats(batchID)
}

// RetryFailed requeues all failed tasks.
func (m *Manager) RetryFailed(batchID string) error {
	batch, err := m.store.GetBatch(batchID)
	if err != nil {
		return err
	}

	if batch.Status == BatchStatusRunning {
		return ErrBatchRunning
	}

	// Get all failed tasks
	tasks, _, err := m.store.ListTasks(batchID, &ListTaskFilter{
		Status: BatchTaskFailed,
		Limit:  100000, // Get all
	})
	if err != nil {
		return err
	}

	// Reset tasks to pending
	for _, task := range tasks {
		task.Status = BatchTaskPending
		task.Error = ""
		task.WorkerID = ""
		task.StartedAt = nil
		task.DurationMs = 0
		// Keep attempts count for tracking
		if err := m.store.UpdateTask(task); err != nil {
			logger.Warn("Failed to reset task", "task_id", task.ID, "error", err)
		}
	}

	// Update batch counters
	batch.Failed = 0
	batch.Status = BatchStatusPending
	batch.ErrorSummary = make(map[string]int)
	if err := m.store.UpdateBatch(batch); err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	logger.Info("Reset failed tasks for retry", "count", len(tasks), "batch_id", batchID)
	return nil
}

// Subscribe returns a channel for batch events.
func (m *Manager) Subscribe(batchID string) chan *BatchEvent {
	ch := make(chan *BatchEvent, 100)

	m.eventMu.Lock()
	m.eventSubs[batchID] = append(m.eventSubs[batchID], ch)
	m.eventMu.Unlock()

	return ch
}

// Unsubscribe removes an event subscription.
func (m *Manager) Unsubscribe(batchID string, ch chan *BatchEvent) {
	m.eventMu.Lock()
	defer m.eventMu.Unlock()

	subs := m.eventSubs[batchID]
	for i, sub := range subs {
		if sub == ch {
			m.eventSubs[batchID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// broadcast sends an event to all subscribers.
func (m *Manager) broadcast(batchID string, event *BatchEvent) {
	m.eventMu.RLock()
	subs := m.eventSubs[batchID]
	m.eventMu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// Shutdown gracefully shuts down the manager.
func (m *Manager) Shutdown() {
	m.cancel()

	m.mu.Lock()
	running := make([]*runningBatch, 0, len(m.running))
	for _, rb := range m.running {
		running = append(running, rb)
	}
	m.running = make(map[string]*runningBatch)
	m.mu.Unlock()

	// Cancel all running batches
	for _, rb := range running {
		rb.cancel()
		rb.wg.Wait()
		for _, w := range rb.workers {
			m.stopWorker(w)
		}
	}

	// Close event channels
	m.eventMu.Lock()
	for _, subs := range m.eventSubs {
		for _, ch := range subs {
			close(ch)
		}
	}
	m.eventSubs = make(map[string][]chan *BatchEvent)
	m.eventMu.Unlock()

	m.store.Close()
	logger.Info("Batch manager shutdown complete")
}

// PoolStats returns global worker pool statistics.
type PoolStats struct {
	MaxBatches     int              `json:"max_batches"`
	RunningBatches int              `json:"running_batches"`
	TotalWorkers   int              `json:"total_workers"`
	BusyWorkers    int              `json:"busy_workers"`
	IdleWorkers    int              `json:"idle_workers"`
	Batches        []RunningBatchInfo `json:"batches"`
}

// RunningBatchInfo provides summary info for a running batch.
type RunningBatchInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Workers     int     `json:"workers"`
	Completed   int     `json:"completed"`
	Failed      int     `json:"failed"`
	Total       int     `json:"total"`
	Percent     float64 `json:"percent"`
	TasksPerSec float64 `json:"tasks_per_sec"`
}

// ListDeadTasks returns dead letter tasks for a batch.
func (m *Manager) ListDeadTasks(batchID string, limit int) ([]*BatchTask, error) {
	// Verify batch exists
	if _, err := m.store.GetBatch(batchID); err != nil {
		return nil, err
	}
	return m.store.ListDeadTasks(batchID, limit)
}

// RetryDeadTasks retries dead letter tasks.
func (m *Manager) RetryDeadTasks(batchID string, taskIDs []string) (int, error) {
	batch, err := m.store.GetBatch(batchID)
	if err != nil {
		return 0, err
	}

	if batch.Status == BatchStatusRunning {
		return 0, ErrBatchRunning
	}

	// Retry in store
	count, err := m.store.RetryDeadTasks(batchID, taskIDs)
	if err != nil {
		return 0, err
	}

	// Re-enqueue to Redis if enabled
	if m.redisQueue != nil && count > 0 {
		if _, err := m.redisQueue.RetryDead(context.Background(), batchID, taskIDs); err != nil {
			logger.Warn("Failed to retry dead tasks in Redis", "batch_id", batchID, "error", err)
		}
	}

	// Update batch status to pending if it was completed/failed
	if batch.Status == BatchStatusCompleted || batch.Status == BatchStatusFailed {
		batch.Status = BatchStatusPending
		if err := m.store.UpdateBatch(batch); err != nil {
			logger.Warn("Failed to update batch status", "batch_id", batchID, "error", err)
		}
	}

	return count, nil
}

// GetQueueOverview returns global queue statistics.
func (m *Manager) GetQueueOverview() (*QueueOverview, error) {
	overview := &QueueOverview{
		Batches: []BatchQueue{},
	}

	// Get all batches (limit to recent ones)
	batches, _, err := m.store.ListBatches(&ListBatchFilter{Limit: 100})
	if err != nil {
		return nil, err
	}

	for _, b := range batches {
		stats, err := m.store.GetTaskStats(b.ID)
		if err != nil {
			continue
		}

		bq := BatchQueue{
			BatchID:     b.ID,
			BatchName:   b.Name,
			Status:      b.Status,
			Pending:     stats.Pending,
			Running:     stats.Running,
			Completed:   stats.Completed,
			Failed:      stats.Failed,
			Dead:        stats.Dead,
			Total:       stats.TotalTasks,
		}

		// Add Redis stats if available
		if m.redisQueue != nil {
			redisStats, err := m.redisQueue.Stats(context.Background(), b.ID)
			if err == nil {
				bq.RedisPending = redisStats.Pending
				bq.RedisProcessing = redisStats.Processing
				bq.RedisDead = redisStats.Dead
			}
		}

		overview.Batches = append(overview.Batches, bq)
		overview.TotalPending += stats.Pending
		overview.TotalRunning += stats.Running
		overview.TotalCompleted += stats.Completed
		overview.TotalFailed += stats.Failed
		overview.TotalDead += stats.Dead
	}

	// Add Redis enabled flag
	overview.RedisEnabled = m.redisQueue != nil

	return overview, nil
}

// GetPoolStats returns current worker pool statistics.
func (m *Manager) GetPoolStats() *PoolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &PoolStats{
		MaxBatches:     m.maxBatches,
		RunningBatches: len(m.running),
		Batches:        make([]RunningBatchInfo, 0, len(m.running)),
	}

	for _, rb := range m.running {
		workerCount := len(rb.workers)
		stats.TotalWorkers += workerCount

		var busyCount, idleCount int
		for _, w := range rb.workers {
			switch w.status {
			case "busy":
				busyCount++
			case "idle":
				idleCount++
			}
		}
		stats.BusyWorkers += busyCount
		stats.IdleWorkers += idleCount

		// Calculate progress
		progress := m.calculateProgress(rb)

		stats.Batches = append(stats.Batches, RunningBatchInfo{
			ID:          rb.batch.ID,
			Name:        rb.batch.Name,
			Workers:     workerCount,
			Completed:   progress.Completed,
			Failed:      progress.Failed,
			Total:       progress.Total,
			Percent:     progress.Percent,
			TasksPerSec: progress.TasksPerSec,
		})
	}

	return stats
}
