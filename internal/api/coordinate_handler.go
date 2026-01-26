package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/coordinate"
)

// CoordinateHandler 跨会话协调处理器
type CoordinateHandler struct {
	manager *coordinate.Manager
}

// NewCoordinateHandler 创建协调处理器
func NewCoordinateHandler(manager *coordinate.Manager) *CoordinateHandler {
	return &CoordinateHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *CoordinateHandler) RegisterRoutes(rg *gin.RouterGroup) {
	coord := rg.Group("/coordinate")
	{
		coord.GET("/sessions", h.ListSessions)
		coord.GET("/sessions/:id", h.GetSession)
		coord.GET("/sessions/:id/history", h.GetSessionHistory)
		coord.POST("/sessions/:id/send", h.SendMessage)
	}
}

// ListSessions 列出所有活跃会话
// @Summary 列出活跃会话
// @Description 返回所有 queued 和 running 状态的任务，供跨会话协调使用
// @Tags Coordinate
// @Produce json
// @Success 200 {object} Response{data=[]coordinate.SessionInfo}
// @Router /api/v1/coordinate/sessions [get]
func (h *CoordinateHandler) ListSessions(c *gin.Context) {
	sessions, err := h.manager.ListSessions(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, sessions)
}

// GetSession 获取单个会话信息
// @Summary 获取会话信息
// @Tags Coordinate
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=coordinate.SessionInfo}
// @Router /api/v1/coordinate/sessions/{id} [get]
func (h *CoordinateHandler) GetSession(c *gin.Context) {
	taskID := c.Param("id")

	session, err := h.manager.GetSession(c.Request.Context(), taskID)
	if err != nil {
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	Success(c, session)
}

// GetSessionHistory 获取会话历史
// @Summary 获取会话对话历史
// @Description 返回指定任务的完整对话记录（prompt + 多轮对话）
// @Tags Coordinate
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=[]coordinate.MessageRecord}
// @Router /api/v1/coordinate/sessions/{id}/history [get]
func (h *CoordinateHandler) GetSessionHistory(c *gin.Context) {
	taskID := c.Param("id")

	history, err := h.manager.GetSessionHistory(c.Request.Context(), taskID)
	if err != nil {
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	Success(c, history)
}

// SendMessageRequest 发送消息请求
type CoordinateSendRequest struct {
	Message string `json:"message" binding:"required"`
}

// SendMessage 向会话发送消息
// @Summary 向另一个会话发送消息
// @Description 向指定任务追加一轮对话（相当于用户发送新消息）
// @Tags Coordinate
// @Accept json
// @Produce json
// @Param id path string true "Target Task ID"
// @Param request body CoordinateSendRequest true "消息内容"
// @Success 200 {object} Response
// @Router /api/v1/coordinate/sessions/{id}/send [post]
func (h *CoordinateHandler) SendMessage(c *gin.Context) {
	taskID := c.Param("id")

	var req CoordinateSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.manager.SendMessage(c.Request.Context(), &coordinate.SendMessageRequest{
		TargetTaskID: taskID,
		Message:      req.Message,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{"message": "sent"})
}
