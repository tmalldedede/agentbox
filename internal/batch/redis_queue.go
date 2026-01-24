package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/logger"
)

// RedisQueue implements task queue operations using Redis.
type RedisQueue struct {
	client       *redis.Client
	claimTimeout time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewRedisQueue creates a new Redis-based queue.
func NewRedisQueue(cfg config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	queueCtx, cancel := context.WithCancel(context.Background())

	return &RedisQueue{
		client:       client,
		claimTimeout: cfg.ClaimTimeout,
		ctx:          queueCtx,
		cancel:       cancel,
	}, nil
}

// Redis key helpers
func keyPending(batchID string) string     { return fmt.Sprintf("batch:%s:pending", batchID) }
func keyProcessing(batchID string) string  { return fmt.Sprintf("batch:%s:processing", batchID) }
func keyDead(batchID string) string        { return fmt.Sprintf("batch:%s:dead", batchID) }
func keyClaims(batchID string) string      { return fmt.Sprintf("batch:%s:claims", batchID) }
func keyTaskData(batchID string) string    { return fmt.Sprintf("batch:%s:tasks", batchID) }

// TaskQueueItem represents a task in the queue.
type TaskQueueItem struct {
	TaskID   string `json:"task_id"`
	BatchID  string `json:"batch_id"`
	Index    int    `json:"index"`
	Attempts int    `json:"attempts"`
}

// Enqueue adds tasks to the pending queue.
func (q *RedisQueue) Enqueue(ctx context.Context, batchID string, tasks []*BatchTask) error {
	if len(tasks) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()

	for _, task := range tasks {
		item := TaskQueueItem{
			TaskID:   task.ID,
			BatchID:  batchID,
			Index:    task.Index,
			Attempts: task.Attempts,
		}
		data, _ := json.Marshal(item)

		// Add to pending list (RPUSH for FIFO order)
		pipe.RPush(ctx, keyPending(batchID), string(data))

		// Store task data for quick lookup
		taskData, _ := json.Marshal(task)
		pipe.HSet(ctx, keyTaskData(batchID), task.ID, string(taskData))
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Claim atomically claims up to `limit` tasks from pending queue.
// Uses Lua script for atomicity.
var claimScript = redis.NewScript(`
local pending_key = KEYS[1]
local processing_key = KEYS[2]
local claims_key = KEYS[3]
local limit = tonumber(ARGV[1])
local worker_id = ARGV[2]
local now = tonumber(ARGV[3])

local claimed = {}
for i = 1, limit do
    local task_data = redis.call('LPOP', pending_key)
    if not task_data then
        break
    end
    -- Add to processing set with current timestamp as score
    redis.call('ZADD', processing_key, now, task_data)
    -- Record claim
    local task = cjson.decode(task_data)
    redis.call('HSET', claims_key, task.task_id, worker_id)
    table.insert(claimed, task_data)
end

return claimed
`)

// Claim claims pending tasks for a worker.
func (q *RedisQueue) Claim(ctx context.Context, batchID string, workerID string, limit int) ([]*TaskQueueItem, error) {
	now := time.Now().Unix()

	result, err := claimScript.Run(ctx, q.client,
		[]string{keyPending(batchID), keyProcessing(batchID), keyClaims(batchID)},
		limit, workerID, now,
	).Result()

	if err != nil && err != redis.Nil {
		return nil, err
	}

	items, ok := result.([]interface{})
	if !ok || len(items) == 0 {
		return nil, nil
	}

	tasks := make([]*TaskQueueItem, 0, len(items))
	for _, item := range items {
		data, ok := item.(string)
		if !ok {
			continue
		}
		var task TaskQueueItem
		if err := json.Unmarshal([]byte(data), &task); err != nil {
			continue
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// Complete removes a task from processing (success).
func (q *RedisQueue) Complete(ctx context.Context, batchID, taskID string) error {
	pipe := q.client.Pipeline()

	// Find and remove from processing set
	// We need to find the item by task_id
	items, err := q.client.ZRange(ctx, keyProcessing(batchID), 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		var task TaskQueueItem
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			continue
		}
		if task.TaskID == taskID {
			pipe.ZRem(ctx, keyProcessing(batchID), item)
			break
		}
	}

	// Remove claim record
	pipe.HDel(ctx, keyClaims(batchID), taskID)
	// Remove task data
	pipe.HDel(ctx, keyTaskData(batchID), taskID)

	_, err = pipe.Exec(ctx)
	return err
}

// Requeue moves a task from processing back to pending for retry.
func (q *RedisQueue) Requeue(ctx context.Context, batchID, taskID string, attempts int) error {
	// Find in processing
	items, err := q.client.ZRange(ctx, keyProcessing(batchID), 0, -1).Result()
	if err != nil {
		return err
	}

	var foundItem string
	var task TaskQueueItem
	for _, item := range items {
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			continue
		}
		if task.TaskID == taskID {
			foundItem = item
			break
		}
	}

	if foundItem == "" {
		return nil // Not in processing, maybe already removed
	}

	pipe := q.client.Pipeline()

	// Remove from processing
	pipe.ZRem(ctx, keyProcessing(batchID), foundItem)

	// Update attempts and add back to pending
	task.Attempts = attempts
	newData, _ := json.Marshal(task)
	pipe.RPush(ctx, keyPending(batchID), string(newData))

	// Remove claim
	pipe.HDel(ctx, keyClaims(batchID), taskID)

	_, err = pipe.Exec(ctx)
	return err
}

// MoveToDead moves a task to the dead letter queue.
func (q *RedisQueue) MoveToDead(ctx context.Context, batchID, taskID string, attempts int, errorMsg string) error {
	// Find in processing
	items, err := q.client.ZRange(ctx, keyProcessing(batchID), 0, -1).Result()
	if err != nil {
		return err
	}

	var foundItem string
	var task TaskQueueItem
	for _, item := range items {
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			continue
		}
		if task.TaskID == taskID {
			foundItem = item
			break
		}
	}

	if foundItem == "" {
		// Create new item for dead queue
		task = TaskQueueItem{
			TaskID:   taskID,
			BatchID:  batchID,
			Attempts: attempts,
		}
	} else {
		task.Attempts = attempts
	}

	pipe := q.client.Pipeline()

	// Remove from processing if found
	if foundItem != "" {
		pipe.ZRem(ctx, keyProcessing(batchID), foundItem)
	}

	// Add to dead letter queue
	deadData, _ := json.Marshal(task)
	pipe.RPush(ctx, keyDead(batchID), string(deadData))

	// Remove claim
	pipe.HDel(ctx, keyClaims(batchID), taskID)

	_, err = pipe.Exec(ctx)
	return err
}

// ListDead returns tasks in the dead letter queue.
func (q *RedisQueue) ListDead(ctx context.Context, batchID string, limit int) ([]*TaskQueueItem, error) {
	items, err := q.client.LRange(ctx, keyDead(batchID), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	tasks := make([]*TaskQueueItem, 0, len(items))
	for _, item := range items {
		var task TaskQueueItem
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			continue
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// RetryDead moves tasks from dead letter queue back to pending.
func (q *RedisQueue) RetryDead(ctx context.Context, batchID string, taskIDs []string) (int, error) {
	if len(taskIDs) == 0 {
		// Retry all
		items, err := q.client.LRange(ctx, keyDead(batchID), 0, -1).Result()
		if err != nil {
			return 0, err
		}

		pipe := q.client.Pipeline()
		count := 0

		for _, item := range items {
			var task TaskQueueItem
			if err := json.Unmarshal([]byte(item), &task); err != nil {
				continue
			}
			// Reset attempts
			task.Attempts = 0
			newData, _ := json.Marshal(task)
			pipe.RPush(ctx, keyPending(batchID), string(newData))
			count++
		}

		// Clear dead queue
		pipe.Del(ctx, keyDead(batchID))

		_, err = pipe.Exec(ctx)
		return count, err
	}

	// Retry specific tasks
	taskIDSet := make(map[string]bool)
	for _, id := range taskIDs {
		taskIDSet[id] = true
	}

	items, err := q.client.LRange(ctx, keyDead(batchID), 0, -1).Result()
	if err != nil {
		return 0, err
	}

	pipe := q.client.Pipeline()
	count := 0
	remaining := make([]string, 0)

	for _, item := range items {
		var task TaskQueueItem
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			remaining = append(remaining, item)
			continue
		}

		if taskIDSet[task.TaskID] {
			// Move to pending
			task.Attempts = 0
			newData, _ := json.Marshal(task)
			pipe.RPush(ctx, keyPending(batchID), string(newData))
			count++
		} else {
			remaining = append(remaining, item)
		}
	}

	// Rebuild dead queue with remaining items
	pipe.Del(ctx, keyDead(batchID))
	if len(remaining) > 0 {
		args := make([]interface{}, len(remaining))
		for i, item := range remaining {
			args[i] = item
		}
		pipe.RPush(ctx, keyDead(batchID), args...)
	}

	_, err = pipe.Exec(ctx)
	return count, err
}

// RecoverTimedOut finds tasks stuck in processing and requeues them.
func (q *RedisQueue) RecoverTimedOut(ctx context.Context, batchID string) (int, error) {
	cutoff := time.Now().Add(-q.claimTimeout).Unix()

	// Find timed out tasks
	items, err := q.client.ZRangeByScore(ctx, keyProcessing(batchID), &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(cutoff, 10),
	}).Result()

	if err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	pipe := q.client.Pipeline()
	count := 0

	for _, item := range items {
		var task TaskQueueItem
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			continue
		}

		// Remove from processing
		pipe.ZRem(ctx, keyProcessing(batchID), item)

		// Add back to pending (keep attempts count)
		pipe.RPush(ctx, keyPending(batchID), item)

		// Remove claim
		pipe.HDel(ctx, keyClaims(batchID), task.TaskID)

		count++
	}

	_, err = pipe.Exec(ctx)
	if err == nil && count > 0 {
		logger.Info("Recovered timed out tasks", "batch_id", batchID, "count", count)
	}

	return count, err
}

// QueueStats returns queue statistics.
type QueueStats struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Dead       int64 `json:"dead"`
}

// Stats returns current queue statistics for a batch.
func (q *RedisQueue) Stats(ctx context.Context, batchID string) (*QueueStats, error) {
	pipe := q.client.Pipeline()

	pendingCmd := pipe.LLen(ctx, keyPending(batchID))
	processingCmd := pipe.ZCard(ctx, keyProcessing(batchID))
	deadCmd := pipe.LLen(ctx, keyDead(batchID))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &QueueStats{
		Pending:    pendingCmd.Val(),
		Processing: processingCmd.Val(),
		Dead:       deadCmd.Val(),
	}, nil
}

// GlobalStats returns aggregate queue stats across all batches.
type GlobalQueueStats struct {
	TotalPending    int64              `json:"total_pending"`
	TotalProcessing int64              `json:"total_processing"`
	TotalDead       int64              `json:"total_dead"`
	ByBatch         map[string]*QueueStats `json:"by_batch,omitempty"`
}

// GlobalStats returns queue stats for all active batches.
func (q *RedisQueue) GlobalStats(ctx context.Context, batchIDs []string) (*GlobalQueueStats, error) {
	stats := &GlobalQueueStats{
		ByBatch: make(map[string]*QueueStats),
	}

	for _, batchID := range batchIDs {
		batchStats, err := q.Stats(ctx, batchID)
		if err != nil {
			continue
		}
		stats.ByBatch[batchID] = batchStats
		stats.TotalPending += batchStats.Pending
		stats.TotalProcessing += batchStats.Processing
		stats.TotalDead += batchStats.Dead
	}

	return stats, nil
}

// Cleanup removes all Redis keys for a batch.
func (q *RedisQueue) Cleanup(ctx context.Context, batchID string) error {
	keys := []string{
		keyPending(batchID),
		keyProcessing(batchID),
		keyDead(batchID),
		keyClaims(batchID),
		keyTaskData(batchID),
	}

	return q.client.Del(ctx, keys...).Err()
}

// StartRecoveryLoop starts a background goroutine to recover timed-out tasks.
func (q *RedisQueue) StartRecoveryLoop(interval time.Duration, getBatchIDs func() []string) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-q.ctx.Done():
				return
			case <-ticker.C:
				batchIDs := getBatchIDs()
				for _, batchID := range batchIDs {
					if _, err := q.RecoverTimedOut(context.Background(), batchID); err != nil {
						logger.Warn("Failed to recover timed out tasks", "batch_id", batchID, "error", err)
					}
				}
			}
		}
	}()
}

// Close closes the Redis connection.
func (q *RedisQueue) Close() error {
	q.cancel()
	return q.client.Close()
}

// Ping checks Redis connectivity.
func (q *RedisQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}
