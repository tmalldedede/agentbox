package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/channel"
	"github.com/tmalldedede/agentbox/internal/channel/feishu"
)

// ChannelHandler 通道处理器
type ChannelHandler struct {
	manager       *channel.Manager
	feishuChannel *feishu.Channel
}

// NewChannelHandler 创建通道处理器
func NewChannelHandler(manager *channel.Manager, feishuChannel *feishu.Channel) *ChannelHandler {
	return &ChannelHandler{
		manager:       manager,
		feishuChannel: feishuChannel,
	}
}

// RegisterRoutes 注册路由（Admin 路由）
func (h *ChannelHandler) RegisterRoutes(rg *gin.RouterGroup) {
	channels := rg.Group("/channels")
	{
		channels.GET("", h.List)
		channels.POST("/send", h.Send)
	}
}

// RegisterWebhookRoutes 注册 Webhook 路由（公开路由，无需认证）
func (h *ChannelHandler) RegisterWebhookRoutes(rg *gin.RouterGroup) {
	webhooks := rg.Group("/webhooks")
	{
		// 飞书事件回调
		webhooks.POST("/feishu", h.FeishuWebhook)
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
