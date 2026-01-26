// Package task provides task management functionality
package task

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LaneQueue manages serial execution queues per backend (Provider+Adapter combination)
// This prevents concurrent requests to the same backend that could cause race conditions or rate limiting
type LaneQueue struct {
	lanes      map[string]*lane    // laneKey -> lane
	mu         sync.RWMutex
	maxPerLane int                 // Max concurrent jobs per lane (typically 1 for serial execution)
	queueSize  int                 // Max pending jobs per lane
}

// lane represents a single execution lane
type lane struct {
	key       string
	jobs      chan *laneJob
	active    int
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// laneJob represents a job to be executed in a lane
type laneJob struct {
	fn       func()
	done     chan struct{}
	enqueued time.Time
}

// LaneConfig configuration for LaneQueue
type LaneConfig struct {
	MaxPerLane int // Max concurrent jobs per lane (default: 1)
	QueueSize  int // Max pending jobs per lane (default: 100)
}

// DefaultLaneConfig returns default configuration
func DefaultLaneConfig() *LaneConfig {
	return &LaneConfig{
		MaxPerLane: 1,   // Serial execution by default
		QueueSize:  100, // Up to 100 pending jobs per lane
	}
}

// NewLaneQueue creates a new lane queue
func NewLaneQueue(cfg *LaneConfig) *LaneQueue {
	if cfg == nil {
		cfg = DefaultLaneConfig()
	}
	if cfg.MaxPerLane <= 0 {
		cfg.MaxPerLane = 1
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	return &LaneQueue{
		lanes:      make(map[string]*lane),
		maxPerLane: cfg.MaxPerLane,
		queueSize:  cfg.QueueSize,
	}
}

// GetLaneKey generates a lane key from provider ID and adapter name
func GetLaneKey(providerID, adapterName string) string {
	return providerID + ":" + adapterName
}

// Enqueue adds a job to the appropriate lane and returns a channel that closes when done
// The job will be executed serially with other jobs on the same lane
func (lq *LaneQueue) Enqueue(laneKey string, fn func()) <-chan struct{} {
	done := make(chan struct{})

	lq.mu.Lock()
	l, ok := lq.lanes[laneKey]
	if !ok {
		l = lq.createLane(laneKey)
		lq.lanes[laneKey] = l
	}
	lq.mu.Unlock()

	job := &laneJob{
		fn:       fn,
		done:     done,
		enqueued: time.Now(),
	}

	select {
	case l.jobs <- job:
		// Job enqueued successfully
	default:
		// Queue is full, execute immediately (with warning)
		log.Warn("lane queue full, executing immediately", "lane", laneKey)
		go func() {
			fn()
			close(done)
		}()
	}

	return done
}

// EnqueueWithContext adds a job with context support
func (lq *LaneQueue) EnqueueWithContext(ctx context.Context, laneKey string, fn func()) <-chan struct{} {
	done := make(chan struct{})

	lq.mu.Lock()
	l, ok := lq.lanes[laneKey]
	if !ok {
		l = lq.createLane(laneKey)
		lq.lanes[laneKey] = l
	}
	lq.mu.Unlock()

	job := &laneJob{
		fn:       fn,
		done:     done,
		enqueued: time.Now(),
	}

	select {
	case <-ctx.Done():
		close(done)
		return done
	case l.jobs <- job:
		// Job enqueued successfully
	default:
		// Queue is full
		log.Warn("lane queue full, executing immediately", "lane", laneKey)
		go func() {
			fn()
			close(done)
		}()
	}

	return done
}

// createLane creates a new lane and starts its worker goroutine
func (lq *LaneQueue) createLane(key string) *lane {
	ctx, cancel := context.WithCancel(context.Background())
	l := &lane{
		key:    key,
		jobs:   make(chan *laneJob, lq.queueSize),
		ctx:    ctx,
		cancel: cancel,
	}

	// Start worker goroutines
	for i := 0; i < lq.maxPerLane; i++ {
		l.wg.Add(1)
		go lq.worker(l)
	}

	log.Debug("lane created", "key", key, "workers", lq.maxPerLane)
	return l
}

// worker processes jobs in a lane
func (lq *LaneQueue) worker(l *lane) {
	defer l.wg.Done()

	for {
		select {
		case <-l.ctx.Done():
			return
		case job, ok := <-l.jobs:
			if !ok {
				return
			}
			lq.executeJob(l, job)
		}
	}
}

// executeJob executes a single job with metrics
func (lq *LaneQueue) executeJob(l *lane, job *laneJob) {
	l.mu.Lock()
	l.active++
	l.mu.Unlock()

	startTime := time.Now()
	waitTime := startTime.Sub(job.enqueued)

	defer func() {
		l.mu.Lock()
		l.active--
		l.mu.Unlock()

		duration := time.Since(startTime)
		log.Debug("lane job completed",
			"lane", l.key,
			"wait_time", waitTime,
			"exec_time", duration,
		)

		close(job.done)
	}()

	// Execute the job
	job.fn()
}

// Stats returns queue statistics
type LaneStats struct {
	TotalLanes   int            `json:"total_lanes"`
	LaneDetails  map[string]int `json:"lane_details"` // laneKey -> pending count
}

func (lq *LaneQueue) Stats() *LaneStats {
	lq.mu.RLock()
	defer lq.mu.RUnlock()

	stats := &LaneStats{
		TotalLanes:  len(lq.lanes),
		LaneDetails: make(map[string]int),
	}

	for key, l := range lq.lanes {
		stats.LaneDetails[key] = len(l.jobs)
	}

	return stats
}

// PendingCount returns the number of pending jobs for a lane
func (lq *LaneQueue) PendingCount(laneKey string) int {
	lq.mu.RLock()
	l, ok := lq.lanes[laneKey]
	lq.mu.RUnlock()

	if !ok {
		return 0
	}
	return len(l.jobs)
}

// ActiveCount returns the number of active jobs for a lane
func (lq *LaneQueue) ActiveCount(laneKey string) int {
	lq.mu.RLock()
	l, ok := lq.lanes[laneKey]
	lq.mu.RUnlock()

	if !ok {
		return 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active
}

// Stop stops all lanes and waits for pending jobs to complete
func (lq *LaneQueue) Stop() {
	lq.mu.Lock()
	lanes := make([]*lane, 0, len(lq.lanes))
	for _, l := range lq.lanes {
		lanes = append(lanes, l)
	}
	lq.mu.Unlock()

	// Cancel all lanes
	for _, l := range lanes {
		l.cancel()
	}

	// Wait for all workers to finish
	for _, l := range lanes {
		l.wg.Wait()
	}

	log.Info("lane queue stopped", "total_lanes", len(lanes))
}

// StopLane stops a specific lane
func (lq *LaneQueue) StopLane(laneKey string) {
	lq.mu.Lock()
	l, ok := lq.lanes[laneKey]
	if ok {
		delete(lq.lanes, laneKey)
	}
	lq.mu.Unlock()

	if ok {
		l.cancel()
		l.wg.Wait()
		log.Debug("lane stopped", "key", laneKey)
	}
}

// DrainLane drains pending jobs from a lane without executing them
func (lq *LaneQueue) DrainLane(laneKey string) int {
	lq.mu.RLock()
	l, ok := lq.lanes[laneKey]
	lq.mu.RUnlock()

	if !ok {
		return 0
	}

	count := 0
	for {
		select {
		case job := <-l.jobs:
			close(job.done)
			count++
		default:
			return count
		}
	}
}

// Metrics for monitoring
type LaneMetrics struct {
	Lane           string        `json:"lane"`
	PendingJobs    int           `json:"pending_jobs"`
	ActiveJobs     int           `json:"active_jobs"`
	TotalProcessed int64         `json:"total_processed"`
	AvgWaitTime    time.Duration `json:"avg_wait_time"`
	AvgExecTime    time.Duration `json:"avg_exec_time"`
}

// GetMetrics returns detailed metrics for a lane (placeholder for future implementation)
func (lq *LaneQueue) GetMetrics(laneKey string) *LaneMetrics {
	lq.mu.RLock()
	l, ok := lq.lanes[laneKey]
	lq.mu.RUnlock()

	if !ok {
		return nil
	}

	l.mu.Lock()
	active := l.active
	l.mu.Unlock()

	return &LaneMetrics{
		Lane:        laneKey,
		PendingJobs: len(l.jobs),
		ActiveJobs:  active,
	}
}

// module logger (uses the existing log from manager.go)
func init() {
	// log is already initialized in manager.go
	if log == nil {
		log = slog.Default().With("module", "task.lane")
	}
}
