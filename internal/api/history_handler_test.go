package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/history"
)

func setupHistoryTestRouter(t *testing.T) (*gin.Engine, *history.Manager) {
	store := history.NewMemoryStore()
	mgr := history.NewManager(store)
	handler := NewHistoryHandler(mgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, mgr
}

func TestHistoryList(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create some history entries
	_ = mgr.Record(&history.Entry{
		ID:         "entry-1",
		SourceType: "task",
		SourceID:   "task-1",
		SourceName: "Test Task 1",
		Engine:     "codex",
		Prompt:     "Test prompt 1",
		Status:     "completed",
		StartedAt:  time.Now().Add(-time.Hour),
		EndedAt:    ptrTime(time.Now()),
	})
	_ = mgr.Record(&history.Entry{
		ID:         "entry-2",
		SourceType: "task",
		SourceID:   "task-2",
		SourceName: "Test Task 2",
		Engine:     "claude-code",
		Prompt:     "Test prompt 2",
		Status:     "completed",
		StartedAt:  time.Now().Add(-30 * time.Minute),
		EndedAt:    ptrTime(time.Now()),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "entries")
	assert.Contains(t, data, "total")

	entries, ok := data["entries"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(entries))
}

func TestHistoryListWithFilters(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create entries with different attributes
	_ = mgr.Record(&history.Entry{
		ID:         "entry-1",
		SourceType: "task",
		SourceID:   "task-1",
		SourceName: "Task 1",
		Engine:     "codex",
		Status:     "completed",
	})
	_ = mgr.Record(&history.Entry{
		ID:         "entry-2",
		SourceType: "session",
		SourceID:   "session-1",
		SourceName: "Session 1",
		Engine:     "claude-code",
		Status:     "failed",
	})

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"filter_by_source_type", "?source_type=task", 1},
		{"filter_by_source_id", "?source_id=task-1", 1},
		{"filter_by_engine", "?engine=codex", 1},
		{"filter_by_status", "?status=completed", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/history"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			data, ok := resp.Data.(map[string]interface{})
			require.True(t, ok)

			entries, ok := data["entries"].([]interface{})
			require.True(t, ok)
			assert.Equal(t, tt.expected, len(entries))
		})
	}
}

func TestHistoryListPagination(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create 10 entries
	for i := 0; i < 10; i++ {
		_ = mgr.Record(&history.Entry{
			ID:         "entry-" + string(rune('a'+i)),
			SourceType: "task",
			SourceID:   "task-" + string(rune('a'+i)),
			Status:     "completed",
		})
	}

	// Test limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/history?limit=5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	entries, ok := data["entries"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 5, len(entries))
	assert.Equal(t, float64(10), data["total"])
}

func TestHistoryGet(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create an entry
	entry := &history.Entry{
		ID:         "test-entry",
		SourceType: "task",
		SourceID:   "task-1",
		SourceName: "Test Task",
		Engine:     "codex",
		Prompt:     "Test prompt",
		Output:     "Test output",
		Status:     "completed",
		Usage: &history.UsageInfo{
			InputTokens:  100,
			OutputTokens: 200,
		},
		StartedAt: time.Now().Add(-time.Hour),
		EndedAt:   ptrTime(time.Now()),
	}
	_ = mgr.Record(entry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/test-entry", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-entry", data["id"])
	assert.Equal(t, "Test prompt", data["prompt"])
	assert.Equal(t, "Test output", data["output"])
}

func TestHistoryGetNotFound(t *testing.T) {
	router, _ := setupHistoryTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryDelete(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create an entry
	_ = mgr.Record(&history.Entry{
		ID:         "delete-me",
		SourceType: "task",
		SourceID:   "task-1",
		Status:     "completed",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/history/delete-me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, data["deleted"])

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/history/delete-me", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryStats(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create entries with various statuses and token counts
	now := time.Now()
	_ = mgr.Record(&history.Entry{
		ID:         "entry-1",
		SourceType: "task",
		Status:     "completed",
		Usage: &history.UsageInfo{
			InputTokens:  100,
			OutputTokens: 200,
		},
		StartedAt: now.Add(-time.Hour),
		EndedAt:   ptrTime(now.Add(-30 * time.Minute)),
	})
	_ = mgr.Record(&history.Entry{
		ID:         "entry-2",
		SourceType: "task",
		Status:     "completed",
		Usage: &history.UsageInfo{
			InputTokens:  150,
			OutputTokens: 250,
		},
		StartedAt: now.Add(-2 * time.Hour),
		EndedAt:   ptrTime(now.Add(-time.Hour)),
	})
	_ = mgr.Record(&history.Entry{
		ID:         "entry-3",
		SourceType: "task",
		Status:     "failed",
		StartedAt:  now.Add(-3 * time.Hour),
		EndedAt:    ptrTime(now.Add(-2 * time.Hour)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	stats, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, stats, "total_executions")
	assert.Contains(t, stats, "total_input_tokens")
	assert.Contains(t, stats, "total_output_tokens")
}

func TestHistoryStatsWithFilters(t *testing.T) {
	router, mgr := setupHistoryTestRouter(t)

	// Create entries with different source types
	_ = mgr.Record(&history.Entry{
		ID:         "task-entry",
		SourceType: "task",
		Status:     "completed",
		Usage: &history.UsageInfo{
			InputTokens:  100,
			OutputTokens: 200,
		},
	})
	_ = mgr.Record(&history.Entry{
		ID:         "session-entry",
		SourceType: "session",
		Status:     "completed",
		Usage: &history.UsageInfo{
			InputTokens:  50,
			OutputTokens: 100,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/stats?source_type=task", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	stats, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	// Should only count task entries
	assert.Equal(t, float64(1), stats["total_executions"])
}

// Helper function
func ptrTime(t time.Time) *time.Time {
	return &t
}
