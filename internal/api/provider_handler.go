package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
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
		providers.GET("/templates", h.ListTemplates)
		providers.GET("/stats", h.Stats)
		providers.POST("/verify-all", h.VerifyAll)
		providers.POST("/probe-models", h.ProbeModels)
		providers.GET("/:id", h.Get)
		providers.POST("", h.Create)
		providers.PUT("/:id", h.Update)
		providers.DELETE("/:id", h.Delete)

		// Key management
		providers.PUT("/:id/key", h.SetKey)
		providers.POST("/:id/verify", h.VerifyKey)
		providers.DELETE("/:id/key", h.DeleteKey)
		providers.GET("/:id/models", h.FetchModels)
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
	configured := c.Query("configured")

	var providers []*provider.Provider

	if configured == "true" {
		providers = h.manager.ListConfigured()
	} else if agent != "" {
		providers = h.manager.ListByAgent(agent)
	} else if category != "" {
		providers = h.manager.ListByCategory(provider.ProviderCategory(category))
	} else {
		providers = h.manager.List()
	}

	Success(c, providers)
}

// ListTemplates returns available provider templates for the "Add Provider" flow
func (h *ProviderHandler) ListTemplates(c *gin.Context) {
	templates := h.manager.ListTemplates()
	Success(c, templates)
}

// Stats returns provider statistics
func (h *ProviderHandler) Stats(c *gin.Context) {
	stats := h.manager.Stats()
	Success(c, stats)
}

// VerifyAll verifies API keys for all configured providers
func (h *ProviderHandler) VerifyAll(c *gin.Context) {
	providers := h.manager.ListConfigured()

	type VerifyResult struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Valid bool   `json:"valid"`
		Error string `json:"error,omitempty"`
	}

	var results []VerifyResult
	for _, p := range providers {
		if !p.IsConfigured {
			continue
		}
		valid, err := h.manager.ValidateKey(p.ID)
		result := VerifyResult{ID: p.ID, Name: p.Name, Valid: valid}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	Success(c, results)
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
		HandleError(c, apperr.NotFound("provider"))
		return
	}

	Success(c, p)
}

// CreateProviderRequest represents the request body for creating a provider
// @Description Request body for creating a new provider
type CreateProviderRequest struct {
	// Template-based creation (preferred)
	TemplateID string   `json:"template_id,omitempty"` // Create from template
	APIKey     string   `json:"api_key,omitempty"`     // API key to configure immediately
	Models     []string `json:"models,omitempty"`      // Override models list

	// Common fields
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	BaseURL string `json:"base_url,omitempty"`

	// Custom creation (when no template_id)
	Description   string            `json:"description,omitempty"`
	Agents        []string          `json:"agents,omitempty"`
	Category      string            `json:"category,omitempty"`
	WebsiteURL    string            `json:"website_url,omitempty"`
	APIKeyURL     string            `json:"api_key_url,omitempty"`
	DocsURL       string            `json:"docs_url,omitempty"`
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
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	// Template-based creation
	if req.TemplateID != "" {
		p, err := h.manager.CreateFromTemplate(req.TemplateID, req.ID, req.Name, req.BaseURL, req.APIKey, req.Models)
		if err != nil {
			HandleError(c, apperr.Wrap(err, "failed to create provider from template"))
			return
		}
		Created(c, p)
		return
	}

	// Custom creation (requires agents)
	if len(req.Agents) == 0 {
		HandleError(c, apperr.Validation("agents is required for custom provider creation"))
		return
	}

	p := &provider.Provider{
		ID:            req.ID,
		Name:          req.Name,
		Description:   req.Description,
		Agents:        req.Agents,
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
		HandleError(c, apperr.Wrap(err, "failed to create provider"))
		return
	}

	// Configure key if provided
	if req.APIKey != "" {
		h.manager.ConfigureKey(p.ID, req.APIKey)
		p, _ = h.manager.Get(p.ID)
	}

	Created(c, p)
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
		HandleError(c, apperr.Validation(err.Error()))
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
		HandleError(c, apperr.Wrap(err, "failed to update provider"))
		return
	}

	// 获取更新后的 Provider
	p, _ := h.manager.Get(id)

	Success(c, p)
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
		HandleError(c, apperr.Wrap(err, "failed to delete provider"))
		return
	}

	Success(c, gin.H{"deleted": id})
}

// --- API Key Management ---

// SetKeyRequest represents the request body for setting an API key
type SetKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// SetKey godoc
// @Summary Set API key for a provider
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Param request body SetKeyRequest true "API Key"
// @Success 200 {object} Response{data=provider.Provider}
// @Router /providers/{id}/key [put]
func (h *ProviderHandler) SetKey(c *gin.Context) {
	id := c.Param("id")

	var req SetKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	if err := h.manager.ConfigureKey(id, req.APIKey); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to set API key"))
		return
	}

	p, _ := h.manager.Get(id)
	Success(c, p)
}

// VerifyKey godoc
// @Summary Verify API key for a provider
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} Response
// @Router /providers/{id}/verify [post]
func (h *ProviderHandler) VerifyKey(c *gin.Context) {
	id := c.Param("id")

	valid, err := h.manager.ValidateKey(id)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to verify API key"))
		return
	}

	Success(c, gin.H{"valid": valid})
}

// DeleteKey godoc
// @Summary Delete API key for a provider
// @Tags Providers
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} Response
// @Router /providers/{id}/key [delete]
func (h *ProviderHandler) DeleteKey(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.DeleteKey(id); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to delete API key"))
		return
	}

	Success(c, gin.H{"deleted": id})
}

// FetchModels godoc
// @Summary Fetch available models for a configured provider
// @Tags Providers
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} Response{data=[]string}
// @Router /providers/{id}/models [get]
func (h *ProviderHandler) FetchModels(c *gin.Context) {
	id := c.Param("id")

	models, err := h.manager.FetchModels(id)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to fetch models"))
		return
	}

	Success(c, models)
}

// ProbeModelsRequest represents the request for probing models without a saved provider
type ProbeModelsRequest struct {
	BaseURL string   `json:"base_url"`
	APIKey  string   `json:"api_key" binding:"required"`
	Agents  []string `json:"agents"` // Used to determine protocol (anthropic vs openai)
}

// ProbeModels godoc
// @Summary Probe available models from a given API endpoint
// @Tags Providers
// @Accept json
// @Produce json
// @Param request body ProbeModelsRequest true "Probe configuration"
// @Success 200 {object} Response{data=[]string}
// @Router /providers/probe-models [post]
func (h *ProviderHandler) ProbeModels(c *gin.Context) {
	var req ProbeModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	models, err := h.manager.ProbeModels(req.BaseURL, req.APIKey, req.Agents)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to probe models"))
		return
	}

	Success(c, models)
}
