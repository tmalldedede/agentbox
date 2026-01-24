package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return db
}

func TestGormStore_Create(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:        "task-001",
		AgentID:   "agent-1",
		AgentName: "Test Agent",
		Prompt:    "Hello world",
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	err = store.Create(task)
	require.NoError(t, err)

	// Verify task was created
	got, err := store.Get("task-001")
	require.NoError(t, err)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, task.AgentID, got.AgentID)
	assert.Equal(t, task.Prompt, got.Prompt)
	assert.Equal(t, task.Status, got.Status)
}

func TestGormStore_CreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:        "task-dup",
		AgentID:   "agent-1",
		Prompt:    "Hello",
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	err = store.Create(task)
	require.NoError(t, err)

	// Create duplicate
	err = store.Create(task)
	assert.ErrorIs(t, err, ErrTaskExists)
}

func TestGormStore_Get(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Get non-existent task
	_, err = store.Get("non-existent")
	assert.ErrorIs(t, err, ErrTaskNotFound)

	// Create and get
	task := &Task{
		ID:        "task-get",
		AgentID:   "agent-1",
		Prompt:    "Test prompt",
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		Metadata:  map[string]string{"key": "value"},
	}
	require.NoError(t, store.Create(task))

	got, err := store.Get("task-get")
	require.NoError(t, err)
	assert.Equal(t, "task-get", got.ID)
	assert.Equal(t, StatusQueued, got.Status)
	assert.Equal(t, "value", got.Metadata["key"])
}

func TestGormStore_Update(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:        "task-update",
		AgentID:   "agent-1",
		Prompt:    "Original prompt",
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(task))

	// Update task
	now := time.Now()
	task.Status = StatusRunning
	task.StartedAt = &now
	task.ErrorMessage = "test error"
	err = store.Update(task)
	require.NoError(t, err)

	// Verify update
	got, err := store.Get("task-update")
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, got.Status)
	assert.NotNil(t, got.StartedAt)
	assert.Equal(t, "test error", got.ErrorMessage)
}

func TestGormStore_UpdateNonExistent(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:     "non-existent",
		Status: StatusRunning,
	}
	err = store.Update(task)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestGormStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:        "task-delete",
		AgentID:   "agent-1",
		Prompt:    "To be deleted",
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(task))

	// Delete
	err = store.Delete("task-delete")
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get("task-delete")
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestGormStore_DeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	err = store.Delete("non-existent")
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestGormStore_List(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Create multiple tasks
	for i := 0; i < 5; i++ {
		task := &Task{
			ID:        "task-list-" + string(rune('a'+i)),
			AgentID:   "agent-1",
			Prompt:    "Test prompt",
			Status:    StatusPending,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, store.Create(task))
	}

	// List all
	tasks, err := store.List(nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 5)

	// List with limit
	tasks, err = store.List(&ListFilter{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	// List with offset
	tasks, err = store.List(&ListFilter{Limit: 2, Offset: 3})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestGormStore_ListWithFilter(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Create tasks with different statuses
	tasks := []struct {
		id      string
		agentID string
		status  Status
		prompt  string
	}{
		{"task-1", "agent-1", StatusPending, "pending task"},
		{"task-2", "agent-1", StatusQueued, "queued task"},
		{"task-3", "agent-2", StatusRunning, "running task"},
		{"task-4", "agent-2", StatusCompleted, "completed task"},
	}

	for _, tc := range tasks {
		task := &Task{
			ID:        tc.id,
			AgentID:   tc.agentID,
			Status:    tc.status,
			Prompt:    tc.prompt,
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.Create(task))
	}

	// Filter by status
	result, err := store.List(&ListFilter{Status: []Status{StatusPending, StatusQueued}})
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Filter by agent
	result, err = store.List(&ListFilter{AgentID: "agent-2"})
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Filter by search
	result, err = store.List(&ListFilter{Search: "queued"})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "task-2", result[0].ID)
}

func TestGormStore_Count(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Empty count
	count, err := store.Count(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create tasks
	for i := 0; i < 3; i++ {
		task := &Task{
			ID:        "task-count-" + string(rune('a'+i)),
			AgentID:   "agent-1",
			Prompt:    "Test",
			Status:    StatusPending,
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.Create(task))
	}

	count, err = store.Count(nil)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Count with filter
	count, err = store.Count(&ListFilter{Status: []Status{StatusCompleted}})
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestGormStore_Stats(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Create tasks with different statuses
	now := time.Now()
	startedAt := now.Add(-5 * time.Minute)
	completedAt := now

	tasks := []*Task{
		{ID: "t1", AgentID: "agent-1", AgentName: "Agent 1", Status: StatusPending, Prompt: "test", CreatedAt: now},
		{ID: "t2", AgentID: "agent-1", AgentName: "Agent 1", Status: StatusQueued, Prompt: "test", CreatedAt: now},
		{ID: "t3", AgentID: "agent-2", AgentName: "Agent 2", Status: StatusCompleted, Prompt: "test", CreatedAt: now, StartedAt: &startedAt, CompletedAt: &completedAt},
	}

	for _, task := range tasks {
		require.NoError(t, store.Create(task))
	}

	stats, err := store.Stats()
	require.NoError(t, err)

	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.ByStatus[StatusPending])
	assert.Equal(t, 1, stats.ByStatus[StatusQueued])
	assert.Equal(t, 1, stats.ByStatus[StatusCompleted])
	assert.Contains(t, stats.ByAgent, "Agent 1")
	assert.Contains(t, stats.ByAgent, "Agent 2")
	assert.Greater(t, stats.AvgDuration, 0.0)
}

func TestGormStore_Cleanup(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now()

	// Create old and new tasks
	tasks := []*Task{
		{ID: "old-1", AgentID: "agent-1", Status: StatusCompleted, Prompt: "test", CreatedAt: oldTime},
		{ID: "old-2", AgentID: "agent-1", Status: StatusFailed, Prompt: "test", CreatedAt: oldTime},
		{ID: "new-1", AgentID: "agent-1", Status: StatusCompleted, Prompt: "test", CreatedAt: newTime},
	}

	for _, task := range tasks {
		require.NoError(t, store.Create(task))
	}

	// Cleanup old completed tasks
	cutoff := time.Now().Add(-24 * time.Hour)
	deleted, err := store.Cleanup(cutoff, []Status{StatusCompleted})
	require.NoError(t, err)
	assert.Equal(t, 1, deleted) // Only old-1 should be deleted

	// Verify remaining
	count, err := store.Count(nil)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestGormStore_ClaimQueued(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Create queued tasks
	for i := 0; i < 5; i++ {
		task := &Task{
			ID:        "claim-" + string(rune('a'+i)),
			AgentID:   "agent-1",
			Status:    StatusQueued,
			Prompt:    "test",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, store.Create(task))
	}

	// Claim 2 tasks
	claimed, err := store.ClaimQueued(2)
	require.NoError(t, err)
	assert.Len(t, claimed, 2)

	// Verify claimed tasks are now running
	for _, task := range claimed {
		assert.Equal(t, StatusRunning, task.Status)
		assert.NotNil(t, task.StartedAt)
	}

	// Verify only 3 queued tasks remain
	remaining, err := store.List(&ListFilter{Status: []Status{StatusQueued}})
	require.NoError(t, err)
	assert.Len(t, remaining, 3)

	// Claim with limit 0 returns empty
	claimed, err = store.ClaimQueued(0)
	require.NoError(t, err)
	assert.Len(t, claimed, 0)
}

func TestGormStore_JSONFields(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	task := &Task{
		ID:          "json-test",
		AgentID:     "agent-1",
		Prompt:      "test",
		Status:      StatusCompleted,
		CreatedAt:   time.Now(),
		Attachments: []string{"file1.txt", "file2.txt"},
		OutputFiles: []OutputFile{
			{Name: "output.txt", Path: "/tmp/output.txt", Size: 100},
		},
		Turns: []Turn{
			{ID: "turn-1", Prompt: "first turn"},
			{ID: "turn-2", Prompt: "second turn"},
		},
		TurnCount: 2,
		Result: &Result{
			Text: "task completed",
			Usage: &Usage{
				DurationSeconds: 10,
				InputTokens:     100,
				OutputTokens:    200,
			},
		},
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	require.NoError(t, store.Create(task))

	// Get and verify all JSON fields are preserved
	got, err := store.Get("json-test")
	require.NoError(t, err)

	assert.Equal(t, []string{"file1.txt", "file2.txt"}, got.Attachments)
	assert.Len(t, got.OutputFiles, 1)
	assert.Equal(t, "output.txt", got.OutputFiles[0].Name)
	assert.Len(t, got.Turns, 2)
	assert.Equal(t, 2, got.TurnCount)
	assert.NotNil(t, got.Result)
	assert.Equal(t, "task completed", got.Result.Text)
	assert.Equal(t, int64(100), got.Result.Usage.InputTokens)
	assert.Equal(t, "value1", got.Metadata["key1"])
}

func TestGormStore_Close(t *testing.T) {
	db := setupTestDB(t)
	store, err := NewGormStore(db)
	require.NoError(t, err)

	// Close should not error (GORM connection managed externally)
	err = store.Close()
	assert.NoError(t, err)
}
