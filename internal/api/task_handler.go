package api

import (
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
		tasks.GET("/:id", h.Get)
		tasks.DELETE("/:id", h.Cancel)
		tasks.GET("/:id/output", h.GetOutput)
		tasks.GET("/:id/logs", h.GetLogs)
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	ProfileID  string            `json:"profile_id" binding:"required"`
	SessionID  string            `json:"session_id,omitempty"` // 可选：使用已存在的 Session
	Prompt     string            `json:"prompt" binding:"required"`
	Input      *task.Input       `json:"input,omitempty"`
	Output     *task.OutputConfig `json:"output,omitempty"`
	WebhookURL string            `json:"webhook_url,omitempty"`
	Timeout    int               `json:"timeout,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Create godoc
// @Summary Create a task
// @Description Create a new async task for batch processing with a specified profile
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body CreateTaskRequest true "Task configuration"
// @Success 201 {object} Response{data=task.Task}
// @Failure 400 {object} Response
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	t, err := h.manager.CreateTask(&task.CreateTaskRequest{
		ProfileID:  req.ProfileID,
		SessionID:  req.SessionID,
		Prompt:     req.Prompt,
		Input:      req.Input,
		Output:     req.Output,
		WebhookURL: req.WebhookURL,
		Timeout:    req.Timeout,
		Metadata:   req.Metadata,
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

// List godoc
// @Summary List tasks
// @Description Get a list of all tasks with optional filtering and pagination
// @Tags Tasks
// @Produce json
// @Param status query string false "Filter by status" Enums(pending, queued, running, completed, failed, cancelled)
// @Param profile_id query string false "Filter by Profile ID"
// @Param limit query int false "Number of results to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} Response{data=ListTasksResponse}
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	filter := &task.ListFilter{
		OrderDesc: true,
	}

	// 状态过滤
	if status := c.Query("status"); status != "" {
		filter.Status = []task.Status{task.Status(status)}
	}

	// Profile 过滤
	if profileID := c.Query("profile_id"); profileID != "" {
		filter.ProfileID = profileID
	}

	// 分页
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

	// 获取总数（不带分页）
	countFilter := &task.ListFilter{
		Status:    filter.Status,
		ProfileID: filter.ProfileID,
	}
	total, _ := h.manager.ListTasks(countFilter)

	Success(c, ListTasksResponse{
		Tasks: tasks,
		Total: len(total),
	})
}

// Get godoc
// @Summary Get a task
// @Description Get detailed information about a specific task
// @Tags Tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=task.Task}
// @Failure 404 {object} Response
// @Router /tasks/{id} [get]
func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")

	t, err := h.manager.GetTask(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, t)
}

// Cancel godoc
// @Summary Cancel a task
// @Description Cancel a pending or running task
// @Tags Tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=task.Task}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.CancelTask(id); err != nil {
		HandleError(c, err)
		return
	}

	// 返回更新后的任务
	t, _ := h.manager.GetTask(id)
	Success(c, t)
}

// GetOutput godoc
// @Summary Get task output
// @Description Get the result/output of a completed task
// @Tags Tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=task.Result}
// @Failure 404 {object} Response
// @Router /tasks/{id}/output [get]
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

// GetLogs godoc
// @Summary Get task logs
// @Description Get execution logs for a task
// @Tags Tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=object{task_id=string,status=string,logs=string}}
// @Failure 404 {object} Response
// @Router /tasks/{id}/logs [get]
func (h *TaskHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")

	t, err := h.manager.GetTask(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	// TODO: 从 Session 获取详细日志
	logs := ""
	if t.Result != nil && t.Result.Logs != "" {
		logs = t.Result.Logs
	}

	Success(c, map[string]interface{}{
		"task_id": t.ID,
		"status":  t.Status,
		"logs":    logs,
	})
}
