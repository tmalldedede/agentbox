package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/session"
)

// Handler API 处理器
type Handler struct {
	sessionMgr    *session.Manager
	agentRegistry *agent.Registry
}

// NewHandler 创建处理器
func NewHandler(sessionMgr *session.Manager, registry *agent.Registry) *Handler {
	return &Handler{
		sessionMgr:    sessionMgr,
		agentRegistry: registry,
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Check if the API server is running
// @Tags Health
// @Produce json
// @Success 200 {object} Response{data=object{status=string,version=string}}
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	Success(c, gin.H{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// ListAgents godoc
// @Summary List agents
// @Description Get a list of all supported agent types (Claude Code, Codex, OpenCode)
// @Tags Agents
// @Produce json
// @Success 200 {object} Response{data=[]agent.Info}
// @Router /agents [get]
func (h *Handler) ListAgents(c *gin.Context) {
	agents := h.agentRegistry.List()
	Success(c, agents)
}

// CreateSession godoc
// @Summary Create a session
// @Description Create a new container session for running an agent
// @Tags Sessions
// @Accept json
// @Produce json
// @Param request body session.CreateRequest true "Session configuration"
// @Success 201 {object} Response{data=session.Session}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /sessions [post]
func (h *Handler) CreateSession(c *gin.Context) {
	var req session.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	sess, err := h.sessionMgr.Create(c.Request.Context(), &req)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Created(c, sess)
}

// ListSessions godoc
// @Summary List sessions
// @Description Get a list of all container sessions with optional filtering
// @Tags Sessions
// @Produce json
// @Param status query string false "Filter by status" Enums(creating, running, stopped, error)
// @Param agent query string false "Filter by agent type"
// @Param limit query int false "Number of results to return"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} Response{data=[]session.Session}
// @Router /sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	var filter session.ListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 使用分页响应
	sessions, total, err := h.sessionMgr.ListWithCount(c.Request.Context(), &filter)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	// 如果有分页参数，返回分页信息
	if filter.Limit > 0 {
		SuccessWithPagination(c, sessions, total, filter.Limit, filter.Offset)
		return
	}

	Success(c, sessions)
}

// GetSession godoc
// @Summary Get a session
// @Description Get detailed information about a specific session
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=session.Session}
// @Failure 404 {object} Response
// @Router /sessions/{id} [get]
func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")

	sess, err := h.sessionMgr.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, sess)
}

// DeleteSession godoc
// @Summary Delete a session
// @Description Stop and delete a session, removing its container
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /sessions/{id} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Delete(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"deleted": id})
}

// StartSession godoc
// @Summary Start a session
// @Description Start a stopped session's container
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=session.Session}
// @Failure 500 {object} Response
// @Router /sessions/{id}/start [post]
func (h *Handler) StartSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Start(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	sess, _ := h.sessionMgr.Get(c.Request.Context(), id)
	Success(c, sess)
}

// StopSession godoc
// @Summary Stop a session
// @Description Stop a running session's container
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=session.Session}
// @Failure 500 {object} Response
// @Router /sessions/{id}/stop [post]
func (h *Handler) StopSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Stop(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	sess, _ := h.sessionMgr.Get(c.Request.Context(), id)
	Success(c, sess)
}

// ReconnectSession godoc
// @Summary Reconnect to a session
// @Description Attempt to reconnect to an existing container session
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=session.Session}
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /sessions/{id}/reconnect [post]
func (h *Handler) ReconnectSession(c *gin.Context) {
	id := c.Param("id")

	sess, err := h.sessionMgr.Reconnect(c.Request.Context(), id)
	if err != nil {
		// 区分不同错误类型
		if err.Error() == "session not found: "+id {
			NotFound(c, err.Error())
			return
		}
		InternalError(c, err.Error())
		return
	}

	Success(c, sess)
}

// ExecSession godoc
// @Summary Execute prompt in session
// @Description Send a prompt to the agent running in the session and get the response
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body session.ExecRequest true "Prompt to execute"
// @Success 200 {object} Response{data=session.ExecResponse}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /sessions/{id}/exec [post]
func (h *Handler) ExecSession(c *gin.Context) {
	id := c.Param("id")

	var req session.ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	result, err := h.sessionMgr.Exec(c.Request.Context(), id, &req)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, result)
}

// GetExecutions godoc
// @Summary List executions
// @Description Get the execution history for a session
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=[]session.Execution}
// @Failure 500 {object} Response
// @Router /sessions/{id}/executions [get]
func (h *Handler) GetExecutions(c *gin.Context) {
	id := c.Param("id")

	executions, err := h.sessionMgr.GetExecutions(c.Request.Context(), id)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, executions)
}

// GetExecution godoc
// @Summary Get an execution
// @Description Get details of a specific execution
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Param execId path string true "Execution ID"
// @Success 200 {object} Response{data=session.Execution}
// @Failure 404 {object} Response
// @Router /sessions/{id}/executions/{execId} [get]
func (h *Handler) GetExecution(c *gin.Context) {
	sessionID := c.Param("id")
	execID := c.Param("execId")

	execution, err := h.sessionMgr.GetExecution(c.Request.Context(), sessionID, execID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, execution)
}

// GetSessionLogs godoc
// @Summary Get session logs
// @Description Get container logs for a session
// @Tags Sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=object{logs=string}}
// @Failure 500 {object} Response
// @Router /sessions/{id}/logs [get]
func (h *Handler) GetSessionLogs(c *gin.Context) {
	id := c.Param("id")

	logs, err := h.sessionMgr.GetLogs(c.Request.Context(), id)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"logs": logs})
}
