package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/channel"
	"github.com/tmalldedede/agentbox/internal/channel/dingtalk"
	"github.com/tmalldedede/agentbox/internal/channel/feishu"
	"github.com/tmalldedede/agentbox/internal/channel/wecom"
)

// ChannelHandler 通道处理器
type ChannelHandler struct {
	manager         *channel.Manager
	feishuChannel   *feishu.Channel
	wecomChannel    *wecom.Channel
	dingtalkChannel *dingtalk.Channel
	sessionStore    channel.SessionStore
	messageStore    channel.MessageStore
}

// NewChannelHandler 创建通道处理器
func NewChannelHandler(
	manager *channel.Manager,
	feishuChannel *feishu.Channel,
	wecomChannel *wecom.Channel,
	dingtalkChannel *dingtalk.Channel,
) *ChannelHandler {
	return &ChannelHandler{
		manager:         manager,
		feishuChannel:   feishuChannel,
		wecomChannel:    wecomChannel,
		dingtalkChannel: dingtalkChannel,
		sessionStore:    channel.NewGormSessionStore(),
		messageStore:    channel.NewGormMessageStore(),
	}
}

// RegisterRoutes 注册路由（Admin 路由）
func (h *ChannelHandler) RegisterRoutes(rg *gin.RouterGroup) {
	channels := rg.Group("/channels")
	{
		channels.GET("", h.List)
		channels.POST("/send", h.Send)
	}

	// 会话管理
	sessions := rg.Group("/channel-sessions")
	{
		sessions.GET("", h.ListSessions)
		sessions.GET("/:id", h.GetSession)
		sessions.GET("/:id/messages", h.GetSessionMessages)
		sessions.POST("/:id/end", h.EndSession)
	}

	// 消息列表和统计
	rg.GET("/channel-messages", h.ListMessages)
	rg.GET("/channel-stats", h.GetStats)
}

// RegisterWebhookRoutes 注册 Webhook 路由（公开路由，无需认证）
func (h *ChannelHandler) RegisterWebhookRoutes(rg *gin.RouterGroup) {
	webhooks := rg.Group("/webhooks")
	{
		// 飞书事件回调
		webhooks.POST("/feishu", h.FeishuWebhook)
		// 企业微信事件回调
		webhooks.Any("/wecom", h.WecomWebhook)
		// 钉钉事件回调
		webhooks.POST("/dingtalk", h.DingtalkWebhook)
	}
}

// List 列出所有通道
// @Summary 列出所有通道
// @Tags Channel
// @Produce json
// @Success 200 {object} Response{data=[]string}
// @Router /api/v1/admin/channels [get]
func (h *ChannelHandler) List(c *gin.Context) {
	channels := h.manager.ListChannels()
	Success(c, channels)
}

// SendRequest 发送消息请求
type SendRequest struct {
	ChannelType string `json:"channel_type" binding:"required"` // feishu, telegram, slack
	ChannelID   string `json:"channel_id" binding:"required"`   // 目标通道 ID
	Content     string `json:"content" binding:"required"`      // 消息内容
	ReplyTo     string `json:"reply_to"`                        // 回复消息 ID
}

// Send 发送消息
// @Summary 发送消息到通道
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body SendRequest true "发送请求"
// @Success 200 {object} Response{data=channel.SendResponse}
// @Router /api/v1/admin/channels/send [post]
func (h *ChannelHandler) Send(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.manager.Send(c.Request.Context(), req.ChannelType, &channel.SendRequest{
		ChannelID: req.ChannelID,
		Content:   req.Content,
		ReplyTo:   req.ReplyTo,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, resp)
}

// FeishuWebhook 飞书事件回调
// @Summary 飞书事件回调
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/webhooks/feishu [post]
func (h *ChannelHandler) FeishuWebhook(c *gin.Context) {
	if h.feishuChannel == nil {
		Error(c, http.StatusServiceUnavailable, "feishu channel not configured")
		return
	}

	h.feishuChannel.HandleWebhook(c.Writer, c.Request)
}

// WecomWebhook 企业微信事件回调
// @Summary 企业微信事件回调
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/webhooks/wecom [post]
func (h *ChannelHandler) WecomWebhook(c *gin.Context) {
	if h.wecomChannel == nil {
		Error(c, http.StatusServiceUnavailable, "wecom channel not configured")
		return
	}

	h.wecomChannel.HandleWebhook(c.Writer, c.Request)
}

// DingtalkWebhook 钉钉事件回调
// @Summary 钉钉事件回调
// @Tags Channel
// @Accept json
// @Produce json
// @Router /api/v1/webhooks/dingtalk [post]
func (h *ChannelHandler) DingtalkWebhook(c *gin.Context) {
	if h.dingtalkChannel == nil {
		Error(c, http.StatusServiceUnavailable, "dingtalk channel not configured")
		return
	}

	h.dingtalkChannel.HandleWebhook(c.Writer, c.Request)
}

// ListSessions 列出通道会话
// @Summary 列出通道会话
// @Tags Channel
// @Produce json
// @Param channel_type query string false "通道类型"
// @Param status query string false "状态"
// @Param agent_id query string false "Agent ID"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-sessions [get]
func (h *ChannelHandler) ListSessions(c *gin.Context) {
	filter := &channel.SessionFilter{
		ChannelType: c.Query("channel_type"),
		Status:      c.Query("status"),
		AgentID:     c.Query("agent_id"),
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	sessions, total, err := h.sessionStore.List(filter)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"sessions": sessions,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// GetSession 获取会话详情
// @Summary 获取会话详情
// @Tags Channel
// @Produce json
// @Param id path string true "会话 ID"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-sessions/{id} [get]
func (h *ChannelHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	session, err := h.sessionStore.GetByID(id)
	if err != nil {
		Error(c, http.StatusNotFound, "session not found")
		return
	}
	Success(c, session)
}

// GetSessionMessages 获取会话消息
// @Summary 获取会话消息
// @Tags Channel
// @Produce json
// @Param id path string true "会话 ID"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-sessions/{id}/messages [get]
func (h *ChannelHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("id")

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	messages, total, err := h.messageStore.ListBySession(sessionID, limit, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"messages": messages,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// EndSession 结束会话
// @Summary 结束会话
// @Tags Channel
// @Produce json
// @Param id path string true "会话 ID"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-sessions/{id}/end [post]
func (h *ChannelHandler) EndSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.sessionStore.UpdateStatus(id, "completed"); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, gin.H{"message": "session ended"})
}

// ListMessages 列出所有消息
// @Summary 列出所有消息
// @Tags Channel
// @Produce json
// @Param channel_type query string false "通道类型"
// @Param direction query string false "方向"
// @Param limit query int false "限制数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-messages [get]
func (h *ChannelHandler) ListMessages(c *gin.Context) {
	filter := &channel.MessageFilter{
		ChannelType: c.Query("channel_type"),
		Direction:   c.Query("direction"),
		TaskID:      c.Query("task_id"),
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	messages, total, err := h.messageStore.List(filter)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"messages": messages,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// GetStats 获取通道统计
// @Summary 获取通道统计
// @Tags Channel
// @Produce json
// @Param channel_type query string false "通道类型"
// @Success 200 {object} Response
// @Router /api/v1/admin/channel-stats [get]
func (h *ChannelHandler) GetStats(c *gin.Context) {
	channelType := c.Query("channel_type")

	var stats *channel.SessionStats
	var err error

	if channelType != "" {
		stats, err = h.sessionStore.GetStatsByChannel(channelType)
	} else {
		stats, err = h.sessionStore.GetStats()
	}

	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, stats)
}
