package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/webhook"
)

// WebhookHandler Webhook API 处理器
type WebhookHandler struct {
	manager *webhook.Manager
}

// NewWebhookHandler 创建 Webhook 处理器
func NewWebhookHandler(manager *webhook.Manager) *WebhookHandler {
	return &WebhookHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *WebhookHandler) RegisterRoutes(r *gin.RouterGroup) {
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("", h.Create)
		webhooks.GET("", h.List)
		webhooks.GET("/:id", h.Get)
		webhooks.PUT("/:id", h.Update)
		webhooks.DELETE("/:id", h.Delete)
	}
}

// Create 创建 Webhook
// @Summary 创建 Webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body webhook.CreateWebhookRequest true "Webhook 配置"
// @Success 201 {object} webhook.Webhook
// @Failure 400 {object} ErrorResponse
// @Router /webhooks [post]
func (h *WebhookHandler) Create(c *gin.Context) {
	var req webhook.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	w, err := h.manager.Create(&req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Created(c, w)
}

// List 列出所有 Webhook
// @Summary 列出所有 Webhook
// @Tags webhooks
// @Produce json
// @Success 200 {array} webhook.Webhook
// @Router /webhooks [get]
func (h *WebhookHandler) List(c *gin.Context) {
	webhooks, err := h.manager.List()
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, webhooks)
}

// Get 获取 Webhook
// @Summary 获取 Webhook 详情
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} webhook.Webhook
// @Failure 404 {object} ErrorResponse
// @Router /webhooks/{id} [get]
func (h *WebhookHandler) Get(c *gin.Context) {
	id := c.Param("id")

	w, err := h.manager.Get(id)
	if err != nil {
		if err == webhook.ErrWebhookNotFound {
			NotFound(c, "webhook not found")
			return
		}
		InternalError(c, err.Error())
		return
	}

	Success(c, w)
}

// Update 更新 Webhook
// @Summary 更新 Webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Param webhook body webhook.UpdateWebhookRequest true "Webhook 配置"
// @Success 200 {object} webhook.Webhook
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /webhooks/{id} [put]
func (h *WebhookHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req webhook.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	w, err := h.manager.Update(id, &req)
	if err != nil {
		if err == webhook.ErrWebhookNotFound {
			NotFound(c, "webhook not found")
			return
		}
		BadRequest(c, err.Error())
		return
	}

	Success(c, w)
}

// Delete 删除 Webhook
// @Summary 删除 Webhook
// @Tags webhooks
// @Param id path string true "Webhook ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /webhooks/{id} [delete]
func (h *WebhookHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		if err == webhook.ErrWebhookNotFound {
			NotFound(c, "webhook not found")
			return
		}
		BadRequest(c, err.Error())
		return
	}

	Success(c, map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}
