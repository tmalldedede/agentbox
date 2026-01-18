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

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	Success(c, gin.H{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// ListAgents 列出支持的 Agent
func (h *Handler) ListAgents(c *gin.Context) {
	agents := h.agentRegistry.List()
	Success(c, agents)
}

// CreateSession 创建会话
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

// ListSessions 列出会话
func (h *Handler) ListSessions(c *gin.Context) {
	var filter session.ListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		BadRequest(c, err.Error())
		return
	}

	sessions, err := h.sessionMgr.List(c.Request.Context(), &filter)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, sessions)
}

// GetSession 获取会话详情
func (h *Handler) GetSession(c *gin.Context) {
	id := c.Param("id")

	sess, err := h.sessionMgr.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, sess)
}

// DeleteSession 删除会话
func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Delete(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"deleted": id})
}

// StartSession 启动会话
func (h *Handler) StartSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Start(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	sess, _ := h.sessionMgr.Get(c.Request.Context(), id)
	Success(c, sess)
}

// StopSession 停止会话
func (h *Handler) StopSession(c *gin.Context) {
	id := c.Param("id")

	if err := h.sessionMgr.Stop(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	sess, _ := h.sessionMgr.Get(c.Request.Context(), id)
	Success(c, sess)
}

// ExecSession 在会话中执行命令
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

// GetExecutions 获取执行历史
func (h *Handler) GetExecutions(c *gin.Context) {
	id := c.Param("id")

	executions, err := h.sessionMgr.GetExecutions(c.Request.Context(), id)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, executions)
}
