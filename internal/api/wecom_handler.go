package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/channel/wecom"
)

// WecomHandler 企业微信配置处理器
type WecomHandler struct {
	store *wecom.Store
}

// NewWecomHandler 创建企业微信处理器
func NewWecomHandler() *WecomHandler {
	return &WecomHandler{
		store: wecom.NewStore(),
	}
}

// RegisterRoutes 注册路由
func (h *WecomHandler) RegisterRoutes(rg *gin.RouterGroup) {
	wecomGroup := rg.Group("/wecom")
	{
		wecomGroup.POST("/config", h.SaveConfig)
		wecomGroup.GET("/config", h.GetConfig)
		wecomGroup.DELETE("/config/:id", h.DeleteConfig)
		wecomGroup.GET("/configs", h.ListConfigs)
	}
}

// WecomConfigRequest 企业微信配置请求
type WecomConfigRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name" binding:"required"`
	CorpID         string `json:"corp_id" binding:"required"`
	AgentID        int    `json:"agent_id" binding:"required"`
	Secret         string `json:"secret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	DefaultAgentID string `json:"default_agent_id"`
}

// SaveConfig 保存企业微信配置
func (h *WecomHandler) SaveConfig(c *gin.Context) {
	var req WecomConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cfg := &wecom.Config{
		Name:           req.Name,
		CorpID:         req.CorpID,
		AgentID:        req.AgentID,
		Secret:         req.Secret,
		Token:          req.Token,
		EncodingAESKey: req.EncodingAESKey,
		DefaultAgentID: req.DefaultAgentID,
	}

	id := req.ID
	if id != "" {
		existing, _, err := h.store.Get(id)
		if err == nil && existing != nil {
			if cfg.Secret == "" {
				cfg.Secret = existing.Secret
			}
			if cfg.Token == "" {
				cfg.Token = existing.Token
			}
			if cfg.EncodingAESKey == "" {
				cfg.EncodingAESKey = existing.EncodingAESKey
			}
		}
	} else {
		if cfg.Secret == "" {
			Error(c, http.StatusBadRequest, "secret is required for new configuration")
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

// GetConfig 获取当前启用的企业微信配置
func (h *WecomHandler) GetConfig(c *gin.Context) {
	cfg, err := h.store.GetEnabledConfig()
	if err != nil {
		Error(c, http.StatusNotFound, "no wecom config found")
		return
	}

	Success(c, gin.H{
		"name":             cfg.Name,
		"corp_id":          cfg.CorpID,
		"agent_id":         cfg.AgentID,
		"secret":           maskSecret(cfg.Secret),
		"token":            maskSecret(cfg.Token),
		"encoding_aes_key": maskSecret(cfg.EncodingAESKey),
		"default_agent_id": cfg.DefaultAgentID,
	})
}

// DeleteConfig 删除企业微信配置
func (h *WecomHandler) DeleteConfig(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.Delete(id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListConfigs 列出所有企业微信配置
func (h *WecomHandler) ListConfigs(c *gin.Context) {
	configs, err := h.store.List()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 隐藏敏感信息
	for _, cfg := range configs {
		if secret, ok := cfg["secret"].(string); ok {
			cfg["secret"] = maskSecret(secret)
		}
		if token, ok := cfg["token"].(string); ok {
			cfg["token"] = maskSecret(token)
		}
		if aesKey, ok := cfg["encoding_aes_key"].(string); ok {
			cfg["encoding_aes_key"] = maskSecret(aesKey)
		}
	}

	Success(c, configs)
}
