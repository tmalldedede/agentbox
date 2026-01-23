package api

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/session"
)

// Handler API 处理器
type Handler struct {
	sessionMgr    *session.Manager
	agentRegistry *engine.Registry
}

// NewHandler 创建处理器
func NewHandler(sessionMgr *session.Manager, registry *engine.Registry) *Handler {
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
// @Success 200 {object} Response{data=[]engine.Info}
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
		return
	}

	Success(c, result)
}

// ExecSessionStream godoc
// @Summary Execute prompt in session with SSE streaming (Codex only)
// @Description Send a prompt to the agent and receive real-time events via SSE
// @Tags Sessions
// @Accept json
// @Produce text/event-stream
// @Param id path string true "Session ID"
// @Param request body session.ExecRequest true "Prompt to execute"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /sessions/{id}/exec/stream [post]
func (h *Handler) ExecSessionStream(c *gin.Context) {
	id := c.Param("id")

	var req session.ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 启动流式执行
	eventCh, execID, err := h.sessionMgr.ExecStream(c.Request.Context(), id, &req)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Execution-ID", execID)

	w := c.Writer

	// 发送连接成功消息
	fmt.Fprintf(w, "event: connected\ndata: {\"execution_id\":\"%s\"}\n\n", execID)
	w.Flush()

	// 从通道读取事件并发送
	for event := range eventCh {
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		// 序列化事件
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		// 发送 SSE 事件
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
		w.Flush()
	}

	// 发送结束事件
	fmt.Fprintf(w, "event: done\ndata: {\"execution_id\":\"%s\"}\n\n", execID)
	w.Flush()
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
		HandleError(c, err)
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
		HandleError(c, err)
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
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"logs": logs})
}

// StreamSessionLogs godoc
// @Summary Stream session logs (SSE)
// @Description Get real-time container logs via Server-Sent Events
// @Tags Sessions
// @Produce text/event-stream
// @Param id path string true "Session ID"
// @Success 200 {string} string "SSE stream"
// @Failure 500 {object} Response
// @Router /sessions/{id}/logs/stream [get]
func (h *Handler) StreamSessionLogs(c *gin.Context) {
	id := c.Param("id")

	// 获取日志流
	reader, err := h.sessionMgr.StreamLogs(c.Request.Context(), id)
	if err != nil {
		HandleError(c, err)
		return
	}
	defer reader.Close()

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取底层响应写入器
	w := c.Writer

	// 立即发送连接成功消息，确保客户端收到响应头
	fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\"}\n\n", id)
	w.Flush()

	// 创建缓冲区读取日志
	buf := make([]byte, 4096)

	// 持续读取并发送日志
	for {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				// Docker 日志有 8 字节头部，需要跳过
				data := buf[:n]
				if len(data) > 8 {
					// 跳过 Docker multiplexed stream header
					data = stripDockerLogHeader(data)
				}

				// 发送 SSE 事件
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.Flush()
			}
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
					w.Flush()
				}
				return
			}
		}
	}
}

// stripDockerLogHeader 移除 Docker 日志的多路复用头部
func stripDockerLogHeader(data []byte) []byte {
	// Docker multiplexed stream format:
	// [8]byte header + payload
	// header[0]: stream type (1=stdout, 2=stderr)
	// header[1-3]: reserved
	// header[4-7]: payload size (big endian)
	result := make([]byte, 0, len(data))
	for len(data) >= 8 {
		// 读取 payload 大小
		size := int(data[4])<<24 | int(data[5])<<16 | int(data[6])<<8 | int(data[7])
		data = data[8:] // 跳过头部

		if size > len(data) {
			size = len(data)
		}
		result = append(result, data[:size]...)
		data = data[size:]
	}
	// 如果还有剩余数据（不完整的帧），也加入
	if len(data) > 0 {
		result = append(result, data...)
	}
	return result
}
