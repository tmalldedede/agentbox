package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/channel/feishu"
)

// FeishuHandler 飞书配置处理器
type FeishuHandler struct {
	store *feishu.Store
}

// NewFeishuHandler 创建飞书处理器
func NewFeishuHandler() *FeishuHandler {
	return &FeishuHandler{
		store: feishu.NewStore(),
	}
}

// RegisterRoutes 注册路由
func (h *FeishuHandler) RegisterRoutes(rg *gin.RouterGroup) {
	feishuGroup := rg.Group("/feishu")
	{
		feishuGroup.POST("/config", h.SaveConfig)
		feishuGroup.GET("/config", h.GetConfig)
		feishuGroup.DELETE("/config/:id", h.DeleteConfig)
		feishuGroup.GET("/configs", h.ListConfigs)
	}
}

// FeishuConfigRequest 飞书配置请求
type FeishuConfigRequest struct {
	ID                string `json:"id"`                        // 编辑时提供 ID
	Name              string `json:"name" binding:"required"`
	AppID             string `json:"app_id" binding:"required"`
	AppSecret         string `json:"app_secret"`                // 编辑时可为空，保持原有值
	VerificationToken string `json:"verification_token"`
	EncryptKey        string `json:"encrypt_key"`
	BotName           string `json:"bot_name"`
	DefaultAgentID    string `json:"default_agent_id"`
}

// SaveConfig 保存飞书配置
// @Summary 保存飞书配置
// @Tags Feishu
// @Accept json
// @Produce json
// @Param request body FeishuConfigRequest true "飞书配置"
// @Success 200 {object} Response
// @Router /api/v1/admin/feishu/config [post]
func (h *FeishuHandler) SaveConfig(c *gin.Context) {
	var req FeishuConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cfg := &feishu.Config{
		Name:              req.Name,
		AppID:             req.AppID,
		AppSecret:         req.AppSecret,
		VerificationToken: req.VerificationToken,
		EncryptKey:        req.EncryptKey,
		BotName:           req.BotName,
		DefaultAgentID:    req.DefaultAgentID,
	}

	// 如果提供了 ID，是更新操作
	id := req.ID
	if id != "" {
		// 获取现有配置，保持未更新的敏感字段
		existing, err := h.store.GetConfig(id)
		if err == nil {
			if cfg.AppSecret == "" {
				cfg.AppSecret = existing.AppSecret
			}
			if cfg.VerificationToken == "" {
				cfg.VerificationToken = existing.VerificationToken
			}
			if cfg.EncryptKey == "" {
				cfg.EncryptKey = existing.EncryptKey
			}
		}
	} else {
		// 新建时必须提供 AppSecret
		if cfg.AppSecret == "" {
			Error(c, http.StatusBadRequest, "app_secret is required for new configuration")
			return
		}
		id = uuid.New().String()
	}

	if err := h.store.SaveConfig(id, cfg); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"id":      id,
		"message": "config saved, restart server to apply",
	})
}

// GetConfig 获取当前启用的飞书配置
// @Summary 获取飞书配置
// @Tags Feishu
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/admin/feishu/config [get]
func (h *FeishuHandler) GetConfig(c *gin.Context) {
	cfg, err := h.store.GetEnabledConfig()
	if err != nil {
		Error(c, http.StatusNotFound, "no feishu config found")
		return
	}

	// 隐藏敏感信息
	Success(c, gin.H{
		"name":               cfg.Name,
		"app_id":             cfg.AppID,
		"app_secret":         maskSecret(cfg.AppSecret),
		"verification_token": maskSecret(cfg.VerificationToken),
		"encrypt_key":        maskSecret(cfg.EncryptKey),
		"bot_name":           cfg.BotName,
		"default_agent_id":   cfg.DefaultAgentID,
	})
}

// DeleteConfig 删除飞书配置
// @Summary 删除飞书配置
// @Tags Feishu
// @Param id path string true "配置 ID"
// @Success 204
// @Router /api/v1/admin/feishu/config/{id} [delete]
func (h *FeishuHandler) DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteConfig(id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListConfigs 列出所有飞书配置
// @Summary 列出所有飞书配置
// @Tags Feishu
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/admin/feishu/configs [get]
func (h *FeishuHandler) ListConfigs(c *gin.Context) {
	configs, err := h.store.ListConfigs()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换并隐藏敏感信息
	result := make([]gin.H, len(configs))
	for i, cfg := range configs {
		result[i] = gin.H{
			"id":                 cfg.ID,
			"name":               cfg.Name,
			"app_id":             cfg.AppID,
			"app_secret":         maskSecret(cfg.AppSecret),
			"verification_token": maskSecret(cfg.VerificationToken),
			"encrypt_key":        maskSecret(cfg.EncryptKey),
			"bot_name":           cfg.BotName,
			"default_agent_id":   cfg.DefaultAgentID,
			"enabled":            cfg.Enabled,
			"created_at":         cfg.CreatedAt,
			"updated_at":         cfg.UpdatedAt,
		}
	}

	Success(c, result)
}

// maskSecret 隐藏敏感信息
func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
