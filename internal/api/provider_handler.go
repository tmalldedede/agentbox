package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/provider"
)

// ProviderHandler Provider API 处理器
type ProviderHandler struct {
	manager *provider.Manager
}

// NewProviderHandler 创建 Provider 处理器
func NewProviderHandler(manager *provider.Manager) *ProviderHandler {
	return &ProviderHandler{
		manager: manager,
	}
}

// RegisterRoutes 注册路由
func (h *ProviderHandler) RegisterRoutes(r *gin.RouterGroup) {
	providers := r.Group("/providers")
	{
		providers.GET("", h.List)
		providers.GET("/:id", h.Get)
		providers.POST("", h.Create)
		providers.PUT("/:id", h.Update)
		providers.DELETE("/:id", h.Delete)
	}
}

// List godoc
// @Summary List all providers
// @Description Get a list of API providers with optional filtering by agent type or category
// @Tags Providers
// @Accept json
// @Produce json
// @Param agent query string false "Filter by agent type" Enums(claude-code, codex, opencode, all)
// @Param category query string false "Filter by category" Enums(official, cn_official, aggregator, third_party)
// @Success 200 {object} Response{data=[]provider.Provider}
// @Router /providers [get]
func (h *ProviderHandler) List(c *gin.Context) {
	agent := c.Query("agent")
	category := c.Query("category")

	var providers []*provider.Provider

	if agent != "" {
		providers = h.manager.ListByAgent(agent)
	} else if category != "" {
		providers = h.manager.ListByCategory(provider.ProviderCategory(category))
	} else {
		providers = h.manager.List()
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    providers,
	})
}

// Get godoc
// @Summary Get a provider
// @Description Get detailed information about a specific provider
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} Response{data=provider.Provider}
// @Failure 404 {object} Response
// @Router /providers/{id} [get]
func (h *ProviderHandler) Get(c *gin.Context) {
	id := c.Param("id")

	p, err := h.manager.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    p,
	})
}

// CreateProviderRequest represents the request body for creating a provider
// @Description Request body for creating a new provider
type CreateProviderRequest struct {
	ID            string            `json:"id" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description,omitempty"`
	Agent         string            `json:"agent" binding:"required"`
	Category      string            `json:"category,omitempty"`
	WebsiteURL    string            `json:"website_url,omitempty"`
	APIKeyURL     string            `json:"api_key_url,omitempty"`
	DocsURL       string            `json:"docs_url,omitempty"`
	BaseURL       string            `json:"base_url,omitempty"`
	EnvConfig     map[string]string `json:"env_config,omitempty"`
	DefaultModel  string            `json:"default_model,omitempty"`
	DefaultModels []string          `json:"default_models,omitempty"`
	Icon          string            `json:"icon,omitempty"`
	IconColor     string            `json:"icon_color,omitempty"`
}

// Create godoc
// @Summary Create a provider
// @Description Create a new custom API provider configuration
// @Tags Providers
// @Accept json
// @Produce json
// @Param request body CreateProviderRequest true "Provider configuration"
// @Success 201 {object} Response{data=provider.Provider}
// @Failure 400 {object} Response
// @Router /providers [post]
func (h *ProviderHandler) Create(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	p := &provider.Provider{
		ID:            req.ID,
		Name:          req.Name,
		Description:   req.Description,
		Agent:         req.Agent,
		Category:      provider.ProviderCategory(req.Category),
		WebsiteURL:    req.WebsiteURL,
		APIKeyURL:     req.APIKeyURL,
		DocsURL:       req.DocsURL,
		BaseURL:       req.BaseURL,
		EnvConfig:     req.EnvConfig,
		DefaultModel:  req.DefaultModel,
		DefaultModels: req.DefaultModels,
		Icon:          req.Icon,
		IconColor:     req.IconColor,
		IsBuiltIn:     false,
		RequiresAK:    true,
		IsEnabled:     true,
	}

	if err := h.manager.Create(p); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    p,
	})
}

// UpdateProviderRequest represents the request body for updating a provider
// @Description Request body for updating an existing provider
type UpdateProviderRequest struct {
	Name          string            `json:"name,omitempty"`
	Description   string            `json:"description,omitempty"`
	BaseURL       string            `json:"base_url,omitempty"`
	EnvConfig     map[string]string `json:"env_config,omitempty"`
	DefaultModel  string            `json:"default_model,omitempty"`
	DefaultModels []string          `json:"default_models,omitempty"`
}

// Update godoc
// @Summary Update a provider
// @Description Update an existing provider configuration (custom providers only)
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Param request body UpdateProviderRequest true "Fields to update"
// @Success 200 {object} Response{data=provider.Provider}
// @Failure 400 {object} Response
// @Router /providers/{id} [put]
func (h *ProviderHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	updates := &provider.Provider{
		Name:          req.Name,
		Description:   req.Description,
		BaseURL:       req.BaseURL,
		EnvConfig:     req.EnvConfig,
		DefaultModel:  req.DefaultModel,
		DefaultModels: req.DefaultModels,
	}

	if err := h.manager.Update(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 获取更新后的 Provider
	p, _ := h.manager.Get(id)

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "updated",
		Data:    p,
	})
}

// Delete godoc
// @Summary Delete a provider
// @Description Delete a custom provider (built-in providers cannot be deleted)
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /providers/{id} [delete]
func (h *ProviderHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "deleted",
	})
}
