package api

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/task"
)

// TaskHandler Task API 处理器
type TaskHandler struct {
	manager *task.Manager
}

// NewTaskHandler 创建 Task 处理器
func NewTaskHandler(manager *task.Manager) *TaskHandler {
	return &TaskHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.Create)
		tasks.GET("", h.List)
		tasks.GET("/stats", h.Stats)
		tasks.POST("/cleanup", h.Cleanup)
		tasks.GET("/:id", h.Get)
		tasks.DELETE("/:id", h.Delete)
		tasks.POST("/:id/cancel", h.Cancel)
		tasks.POST("/:id/retry", h.Retry)
		tasks.GET("/:id/events", h.StreamEvents)
		tasks.GET("/:id/output", h.GetOutput)
	}
}

// CreateTaskAPIRequest 创建任务 API 请求（简化版，对齐 Manus）
type CreateTaskAPIRequest struct {
	AgentID     string            `json:"agent_id,omitempty"`    // 首次创建时必填
	Prompt      string            `json:"prompt" binding:"required"`
	TaskID      string            `json:"task_id,omitempty"`     // 多轮时传入
	Attachments []string          `json:"attachments,omitempty"` // file IDs
	WebhookURL  string            `json:"webhook_url,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Create 创建任务或追加多轮
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	t, err := h.manager.CreateTask(&task.CreateTaskRequest{
		AgentID:     req.AgentID,
		Prompt:      req.Prompt,
		TaskID:      req.TaskID,
		Attachments: req.Attachments,
		WebhookURL:  req.WebhookURL,
		Timeout:     req.Timeout,
		Metadata:    req.Metadata,
	})
	if err != nil {
		HandleError(c, err)
		return
	}

	Created(c, t)
}

// ListTasksResponse 任务列表响应
type ListTasksResponse struct {
	Tasks []*task.Task `json:"tasks"`
	Total int          `json:"total"`
}

// List 列出任务
// GET /api/v1/tasks
func (h *TaskHandler) List(c *gin.Context) {
	filter := &task.ListFilter{
		OrderDesc: true,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = []task.Status{task.Status(status)}
	}

	if agentID := c.Query("agent_id"); agentID != "" {
		filter.AgentID = agentID
	}

	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	} else {
		filter.Limit = 20
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	tasks, err := h.manager.ListTasks(filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 使用 Count 高效获取总数
	countFilter := &task.ListFilter{
		Status:  filter.Status,
		AgentID: filter.AgentID,
		Search:  filter.Search,
	}
	total, _ := h.manager.CountTasks(countFilter)

	Success(c, ListTasksResponse{
		Tasks: tasks,
		Total: total,
	})
}

// Get 获取任务详情
// GET /api/v1/tasks/:id
func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")

	t, err := h.manager.GetTask(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, t)
}

// Cancel 取消任务
// POST /api/v1/tasks/:id/cancel
func (h *TaskHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.CancelTask(id); err != nil {
		HandleError(c, err)
		return
	}

	t, _ := h.manager.GetTask(id)
	Success(c, t)
}

// Delete 删除任务
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.DeleteTask(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"deleted": true})
}

// Retry 重试任务
// POST /api/v1/tasks/:id/retry
func (h *TaskHandler) Retry(c *gin.Context) {
	id := c.Param("id")

	t, err := h.manager.RetryTask(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Created(c, t)
}

// Stats 获取任务统计
// GET /api/v1/tasks/stats
func (h *TaskHandler) Stats(c *gin.Context) {
	stats, err := h.manager.GetStats()
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, stats)
}

// CleanupRequest 清理请求
type CleanupRequest struct {
	BeforeDays int           `json:"before_days"` // 清理多少天前的任务
	Statuses   []task.Status `json:"statuses"`    // 要清理的状态，默认 completed/failed/cancelled
}

// Cleanup 清理旧任务
// POST /api/v1/tasks/cleanup
func (h *TaskHandler) Cleanup(c *gin.Context) {
	var req CleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body，使用默认值
		req.BeforeDays = 7
	}

	count, err := h.manager.CleanupTasks(req.BeforeDays, req.Statuses)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"deleted": count})
}

// StreamEvents SSE 实时事件流
// GET /api/v1/tasks/:id/events
func (h *TaskHandler) StreamEvents(c *gin.Context) {
	taskID := c.Param("id")

	// 验证 task 存在
	t, err := h.manager.GetTask(taskID)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 如果任务已经完成，直接返回最终事件
	if t.Status.IsTerminal() {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		eventType := "task.completed"
		if t.Status == task.StatusFailed {
			eventType = "task.failed"
		} else if t.Status == task.StatusCancelled {
			eventType = "task.cancelled"
		}

		data, _ := json.Marshal(map[string]interface{}{
			"task_id": t.ID,
			"status":  t.Status,
			"result":  t.Result,
		})
		c.Writer.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data)))
		c.Writer.Flush()
		return
	}

	// 设置 SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 订阅事件
	eventCh := h.manager.SubscribeEvents(taskID)
	defer h.manager.UnsubscribeEvents(taskID, eventCh)

	// 发送初始状态
	initData, _ := json.Marshal(map[string]interface{}{
		"task_id":    t.ID,
		"status":    t.Status,
		"turn_count": t.TurnCount,
	})
	c.Writer.WriteString(fmt.Sprintf("event: task.status\ndata: %s\n\n", string(initData)))
	c.Writer.Flush()

	// 转发事件
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case event, ok := <-eventCh:
			if !ok {
				return
			}

			data, _ := json.Marshal(event.Data)
			c.Writer.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(data)))
			c.Writer.Flush()

			// 终态事件后关闭连接
			if event.Type == "task.completed" || event.Type == "task.failed" || event.Type == "task.cancelled" {
				return
			}
		}
	}
}

// GetOutput 获取任务输出
// GET /api/v1/tasks/:id/output
func (h *TaskHandler) GetOutput(c *gin.Context) {
	id := c.Param("id")

	t, err := h.manager.GetTask(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	if t.Result == nil {
		Success(c, map[string]interface{}{
			"status":  t.Status,
			"message": "task has no output yet",
		})
		return
	}

	Success(c, t.Result)
}
