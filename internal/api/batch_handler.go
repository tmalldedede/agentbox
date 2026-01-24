package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tmalldedede/agentbox/internal/batch"
)

// BatchHandler handles batch API requests.
type BatchHandler struct {
	batchMgr *batch.Manager
}

// NewBatchHandler creates a new batch handler.
func NewBatchHandler(batchMgr *batch.Manager) *BatchHandler {
	return &BatchHandler{
		batchMgr: batchMgr,
	}
}

// RegisterRoutes registers batch routes.
func (h *BatchHandler) RegisterRoutes(r *gin.RouterGroup) {
	batches := r.Group("/batches")
	{
		batches.POST("", h.Create)
		batches.GET("", h.List)
		batches.GET("/:id", h.Get)
		batches.DELETE("/:id", h.Delete)
		batches.POST("/:id/start", h.Start)
		batches.POST("/:id/pause", h.Pause)
		batches.POST("/:id/resume", h.Resume)
		batches.POST("/:id/cancel", h.Cancel)
		batches.POST("/:id/retry", h.RetryFailed)
		batches.GET("/:id/tasks", h.ListTasks)
		batches.GET("/:id/tasks/:taskId", h.GetTask)
		batches.GET("/:id/stats", h.GetStats)
		batches.GET("/:id/events", h.StreamEvents)
		batches.GET("/:id/export", h.Export)

		// Dead letter queue operations
		batches.GET("/:id/dead", h.ListDeadTasks)
		batches.POST("/:id/dead/retry", h.RetryDeadTasks)
	}

	// Queue overview (global)
	r.GET("/queue/overview", h.GetQueueOverview)
	r.GET("/queue/pool", h.GetPoolStats)
}

// Create creates a new batch.
func (h *BatchHandler) Create(c *gin.Context) {
	var req batch.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	req.UserID = c.GetString("user_id")

	b, err := h.batchMgr.Create(&req)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Created(c, b)
}

// List returns all batches with optional filtering.
func (h *BatchHandler) List(c *gin.Context) {
	filter := &batch.ListBatchFilter{}

	// 非 admin 用户只能看自己的 batch
	if role := c.GetString("role"); role != "admin" {
		filter.UserID = c.GetString("user_id")
	}

	if status := c.Query("status"); status != "" {
		filter.Status = batch.BatchStatus(status)
	}
	if agentID := c.Query("agent_id"); agentID != "" {
		filter.AgentID = agentID
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	batches, total, err := h.batchMgr.List(filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	SuccessWithPagination(c, gin.H{"batches": batches}, total, filter.Limit, filter.Offset)
}

// checkBatchOwnership 检查 batch 归属权（非 admin 用户只能访问自己的 batch）
func (h *BatchHandler) checkBatchOwnership(c *gin.Context, batchID string) (*batch.Batch, bool) {
	b, err := h.batchMgr.Get(batchID)
	if err != nil {
		if err == batch.ErrBatchNotFound {
			Error(c, http.StatusNotFound, "batch not found")
			return nil, false
		}
		HandleError(c, err)
		return nil, false
	}

	// admin 可以访问所有 batch
	if c.GetString("role") == "admin" {
		return b, true
	}

	// 非 admin 只能访问自己的 batch
	if b.UserID != c.GetString("user_id") {
		Forbidden(c, "access denied: not your batch")
		return nil, false
	}

	return b, true
}

// Get returns a single batch.
func (h *BatchHandler) Get(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}
	Success(c, b)
}

// Delete deletes a batch.
func (h *BatchHandler) Delete(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.Delete(b.ID); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"deleted": true})
}

// Start starts a batch.
func (h *BatchHandler) Start(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.Start(b.ID); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	b, _ = h.batchMgr.Get(b.ID)
	Success(c, b)
}

// Pause pauses a running batch.
func (h *BatchHandler) Pause(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.Pause(b.ID); err != nil {
		if err == batch.ErrBatchNotRunning {
			Error(c, http.StatusBadRequest, "batch is not running")
			return
		}
		HandleError(c, err)
		return
	}

	b, _ = h.batchMgr.Get(b.ID)
	Success(c, b)
}

// Resume resumes a paused batch.
func (h *BatchHandler) Resume(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.Resume(b.ID); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	b, _ = h.batchMgr.Get(b.ID)
	Success(c, b)
}

// Cancel cancels a batch.
func (h *BatchHandler) Cancel(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.Cancel(b.ID); err != nil {
		HandleError(c, err)
		return
	}

	b, _ = h.batchMgr.Get(b.ID)
	Success(c, b)
}

// RetryFailed requeues all failed tasks.
func (h *BatchHandler) RetryFailed(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	if err := h.batchMgr.RetryFailed(b.ID); err != nil {
		if err == batch.ErrBatchRunning {
			Error(c, http.StatusBadRequest, "cannot retry while batch is running")
			return
		}
		HandleError(c, err)
		return
	}

	b, _ = h.batchMgr.Get(b.ID)
	Success(c, b)
}

// ListTasks returns tasks for a batch.
func (h *BatchHandler) ListTasks(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}
	batchID := b.ID

	filter := &batch.ListTaskFilter{}
	if status := c.Query("status"); status != "" {
		filter.Status = batch.BatchTaskStatus(status)
	}
	if workerID := c.Query("worker_id"); workerID != "" {
		filter.WorkerID = workerID
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	tasks, total, err := h.batchMgr.ListTasks(batchID, filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	SuccessWithPagination(c, gin.H{"tasks": tasks}, total, filter.Limit, filter.Offset)
}

// GetTask returns a single task.
func (h *BatchHandler) GetTask(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	task, err := h.batchMgr.GetTask(b.ID, c.Param("taskId"))
	if err != nil {
		if err == batch.ErrTaskNotFound {
			Error(c, http.StatusNotFound, "task not found")
			return
		}
		HandleError(c, err)
		return
	}

	Success(c, task)
}

// GetStats returns statistics for a batch.
func (h *BatchHandler) GetStats(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	stats, err := h.batchMgr.GetStats(b.ID)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, stats)
}

// StreamEvents streams batch events via SSE.
func (h *BatchHandler) StreamEvents(c *gin.Context) {
	// Verify batch exists and has permission
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}
	batchID := b.ID

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Subscribe to events
	eventCh := h.batchMgr.Subscribe(batchID)
	defer h.batchMgr.Unsubscribe(batchID, eventCh)

	// Stream events
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return false
			}

			data, _ := json.Marshal(event)
			c.SSEvent(event.Type, string(data))
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// Export exports batch results as CSV or JSON.
func (h *BatchHandler) Export(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	format := c.DefaultQuery("format", "json")

	// Get all completed and failed tasks
	tasks, _, err := h.batchMgr.ListTasks(b.ID, &batch.ListTaskFilter{
		Limit: 100000, // Get all
	})
	if err != nil {
		HandleError(c, err)
		return
	}

	switch format {
	case "csv":
		h.exportCSV(c, b.ID, tasks)
	default:
		h.exportJSON(c, b.ID, tasks)
	}
}

func (h *BatchHandler) exportJSON(c *gin.Context, batchID string, tasks []*batch.BatchTask) {
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-results.json\"", batchID))
	c.JSON(http.StatusOK, gin.H{
		"batch_id": batchID,
		"tasks":    tasks,
	})
}

func (h *BatchHandler) exportCSV(c *gin.Context, batchID string, tasks []*batch.BatchTask) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-results.csv\"", batchID))

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	// Write header
	w.Write([]string{"index", "status", "input", "result", "error", "duration_ms", "attempts"})

	// Write rows
	for _, task := range tasks {
		inputJSON, _ := json.Marshal(task.Input)
		w.Write([]string{
			strconv.Itoa(task.Index),
			string(task.Status),
			string(inputJSON),
			task.Result,
			task.Error,
			strconv.FormatInt(task.DurationMs, 10),
			strconv.Itoa(task.Attempts),
		})
	}
}

// ListDeadTasks returns dead letter tasks for a batch.
func (h *BatchHandler) ListDeadTasks(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	tasks, err := h.batchMgr.ListDeadTasks(b.ID, limit)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{
		"batch_id": b.ID,
		"tasks":    tasks,
		"count":    len(tasks),
	})
}

// RetryDeadTasks retries dead letter tasks.
func (h *BatchHandler) RetryDeadTasks(c *gin.Context) {
	b, ok := h.checkBatchOwnership(c, c.Param("id"))
	if !ok {
		return
	}

	var req batch.RetryDeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body to retry all
		req.TaskIDs = nil
	}

	count, err := h.batchMgr.RetryDeadTasks(b.ID, req.TaskIDs)
	if err != nil {
		if err == batch.ErrBatchRunning {
			Error(c, http.StatusBadRequest, "cannot retry while batch is running")
			return
		}
		HandleError(c, err)
		return
	}

	Success(c, gin.H{
		"batch_id":      b.ID,
		"retried_count": count,
	})
}

// GetQueueOverview returns global queue statistics.
func (h *BatchHandler) GetQueueOverview(c *gin.Context) {
	overview, err := h.batchMgr.GetQueueOverview()
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, overview)
}

// GetPoolStats returns worker pool statistics.
func (h *BatchHandler) GetPoolStats(c *gin.Context) {
	stats := h.batchMgr.GetPoolStats()
	Success(c, stats)
}
