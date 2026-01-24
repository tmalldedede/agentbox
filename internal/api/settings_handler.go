package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/settings"
)

// SettingsHandler Settings API 处理器
type SettingsHandler struct {
	manager *settings.Manager
}

// NewSettingsHandler 创建处理器
func NewSettingsHandler(manager *settings.Manager) *SettingsHandler {
	return &SettingsHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *SettingsHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/settings")
	{
		g.GET("", h.GetAll)
		g.PUT("", h.UpdateAll)
		g.POST("/reset", h.Reset)

		// 分类配置
		g.GET("/agent", h.GetAgent)
		g.PUT("/agent", h.UpdateAgent)

		g.GET("/task", h.GetTask)
		g.PUT("/task", h.UpdateTask)

		g.GET("/batch", h.GetBatch)
		g.PUT("/batch", h.UpdateBatch)

		g.GET("/storage", h.GetStorage)
		g.PUT("/storage", h.UpdateStorage)

		g.GET("/notify", h.GetNotify)
		g.PUT("/notify", h.UpdateNotify)
	}
}

// GetAll 获取所有配置
// @Summary Get all settings
// @Tags Settings
// @Produce json
// @Success 200 {object} Response{data=settings.Settings}
// @Router /admin/settings [get]
func (h *SettingsHandler) GetAll(c *gin.Context) {
	Success(c, h.manager.Get())
}

// UpdateAll 更新所有配置
// @Summary Update all settings
// @Tags Settings
// @Accept json
// @Produce json
// @Param body body settings.Settings true "Settings"
// @Success 200 {object} Response{data=settings.Settings}
// @Router /admin/settings [put]
func (h *SettingsHandler) UpdateAll(c *gin.Context) {
	var req settings.Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.Update(&req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update settings: "+err.Error())
		return
	}

	Success(c, h.manager.Get())
}

// Reset 重置为默认配置
// @Summary Reset settings to defaults
// @Tags Settings
// @Produce json
// @Success 200 {object} Response{data=settings.Settings}
// @Router /admin/settings/reset [post]
func (h *SettingsHandler) Reset(c *gin.Context) {
	if err := h.manager.Reset(); err != nil {
		Error(c, http.StatusInternalServerError, "failed to reset settings: "+err.Error())
		return
	}

	Success(c, h.manager.Get())
}

// GetAgent 获取 Agent 配置
func (h *SettingsHandler) GetAgent(c *gin.Context) {
	Success(c, h.manager.GetAgent())
}

// UpdateAgent 更新 Agent 配置
func (h *SettingsHandler) UpdateAgent(c *gin.Context) {
	var req settings.AgentSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.UpdateAgent(req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update agent settings: "+err.Error())
		return
	}

	Success(c, h.manager.GetAgent())
}

// GetTask 获取 Task 配置
func (h *SettingsHandler) GetTask(c *gin.Context) {
	Success(c, h.manager.GetTask())
}

// UpdateTask 更新 Task 配置
func (h *SettingsHandler) UpdateTask(c *gin.Context) {
	var req settings.TaskSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.UpdateTask(req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update task settings: "+err.Error())
		return
	}

	Success(c, h.manager.GetTask())
}

// GetBatch 获取 Batch 配置
func (h *SettingsHandler) GetBatch(c *gin.Context) {
	Success(c, h.manager.GetBatch())
}

// UpdateBatch 更新 Batch 配置
func (h *SettingsHandler) UpdateBatch(c *gin.Context) {
	var req settings.BatchSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.UpdateBatch(req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update batch settings: "+err.Error())
		return
	}

	Success(c, h.manager.GetBatch())
}

// GetStorage 获取 Storage 配置
func (h *SettingsHandler) GetStorage(c *gin.Context) {
	Success(c, h.manager.GetStorage())
}

// UpdateStorage 更新 Storage 配置
func (h *SettingsHandler) UpdateStorage(c *gin.Context) {
	var req settings.StorageSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.UpdateStorage(req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update storage settings: "+err.Error())
		return
	}

	Success(c, h.manager.GetStorage())
}

// GetNotify 获取 Notify 配置
func (h *SettingsHandler) GetNotify(c *gin.Context) {
	Success(c, h.manager.GetNotify())
}

// UpdateNotify 更新 Notify 配置
func (h *SettingsHandler) UpdateNotify(c *gin.Context) {
	var req settings.NotifySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := h.manager.UpdateNotify(req); err != nil {
		Error(c, http.StatusInternalServerError, "failed to update notify settings: "+err.Error())
		return
	}

	Success(c, h.manager.GetNotify())
}
