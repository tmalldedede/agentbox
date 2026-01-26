package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/channel/dingtalk"
)

// DingtalkHandler 钉钉配置处理器
type DingtalkHandler struct {
	store *dingtalk.Store
}

// NewDingtalkHandler 创建钉钉处理器
func NewDingtalkHandler() *DingtalkHandler {
	return &DingtalkHandler{
		store: dingtalk.NewStore(),
	}
}

// RegisterRoutes 注册路由
func (h *DingtalkHandler) RegisterRoutes(rg *gin.RouterGroup) {
	dingtalkGroup := rg.Group("/dingtalk")
	{
		dingtalkGroup.POST("/config", h.SaveConfig)
		dingtalkGroup.GET("/config", h.GetConfig)
		dingtalkGroup.DELETE("/config/:id", h.DeleteConfig)
		dingtalkGroup.GET("/configs", h.ListConfigs)
	}
}

// DingtalkConfigRequest 钉钉配置请求
type DingtalkConfigRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name" binding:"required"`
	AppKey         string `json:"app_key" binding:"required"`
	AppSecret      string `json:"app_secret"`
	AgentID        int64  `json:"agent_id" binding:"required"`
	RobotCode      string `json:"robot_code"`
	DefaultAgentID string `json:"default_agent_id"`
}

// SaveConfig 保存钉钉配置
func (h *DingtalkHandler) SaveConfig(c *gin.Context) {
	var req DingtalkConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cfg := &dingtalk.Config{
		Name:           req.Name,
		AppKey:         req.AppKey,
		AppSecret:      req.AppSecret,
		AgentID:        req.AgentID,
		RobotCode:      req.RobotCode,
		DefaultAgentID: req.DefaultAgentID,
	}

	id := req.ID
	if id != "" {
		existing, _, err := h.store.Get(id)
		if err == nil && existing != nil {
			if cfg.AppSecret == "" {
				cfg.AppSecret = existing.AppSecret
			}
		}
	} else {
		if cfg.AppSecret == "" {
			Error(c, http.StatusBadRequest, "app_secret is required for new configuration")
			return
		}
		id = uuid.New().String()
	}

	if err := h.store.Save(cfg, id, true); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{
		"id":      id,
		"message": "config saved, restart server to apply",
	})
}

// GetConfig 获取当前启用的钉钉配置
func (h *DingtalkHandler) GetConfig(c *gin.Context) {
	cfg, err := h.store.GetEnabledConfig()
	if err != nil {
		Error(c, http.StatusNotFound, "no dingtalk config found")
		return
	}

	Success(c, gin.H{
		"name":             cfg.Name,
		"app_key":          cfg.AppKey,
		"app_secret":       maskSecret(cfg.AppSecret),
		"agent_id":         cfg.AgentID,
		"robot_code":       cfg.RobotCode,
		"default_agent_id": cfg.DefaultAgentID,
	})
}

// DeleteConfig 删除钉钉配置
func (h *DingtalkHandler) DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.Delete(id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListConfigs 列出所有钉钉配置
func (h *DingtalkHandler) ListConfigs(c *gin.Context) {
	configs, err := h.store.List()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 隐藏敏感信息
	for _, cfg := range configs {
		if secret, ok := cfg["app_secret"].(string); ok {
			cfg["app_secret"] = maskSecret(secret)
		}
	}

	Success(c, configs)
}
