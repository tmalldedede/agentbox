package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/profile"
)

// ProfileHandler Profile API 处理器
type ProfileHandler struct {
	manager *profile.Manager
}

// NewProfileHandler 创建 Profile 处理器
func NewProfileHandler(manager *profile.Manager) *ProfileHandler {
	return &ProfileHandler{manager: manager}
}

// RegisterRoutes 注册 Profile 路由 (/api/v1/profiles)
func (h *ProfileHandler) RegisterRoutes(r *gin.RouterGroup) {
	profiles := r.Group("/profiles")
	{
		profiles.GET("", h.List)
		profiles.POST("", h.Create)
		profiles.GET("/:id", h.Get)
		profiles.PUT("/:id", h.Update)
		profiles.DELETE("/:id", h.Delete)
		profiles.POST("/:id/clone", h.Clone)
		profiles.GET("/:id/resolved", h.GetResolved)
	}
}

// List godoc
// @Summary List all profiles
// @Description Get a list of all available runtime configuration profiles
// @Tags Profiles
// @Produce json
// @Success 200 {object} Response{data=[]profile.Profile}
// @Router /profiles [get]
func (h *ProfileHandler) List(c *gin.Context) {
	profiles := h.manager.List()
	Success(c, profiles)
}

// Get godoc
// @Summary Get a profile
// @Description Get detailed information about a specific profile
// @Tags Profiles
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} Response{data=profile.Profile}
// @Failure 404 {object} Response
// @Router /profiles/{id} [get]
func (h *ProfileHandler) Get(c *gin.Context) {
	id := c.Param("id")

	p, err := h.manager.Get(id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, p)
}

// GetResolved godoc
// @Summary Get resolved profile
// @Description Get a profile with inheritance chain resolved (merges parent profile settings)
// @Tags Profiles
// @Produce json
// @Param id path string true "Profile ID"
// @Success 200 {object} Response{data=profile.Profile}
// @Failure 404 {object} Response
// @Router /profiles/{id}/resolved [get]
func (h *ProfileHandler) GetResolved(c *gin.Context) {
	id := c.Param("id")

	p, err := h.manager.GetResolved(id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, p)
}

// CreateProfileRequest 创建 Profile 请求
type CreateProfileRequest struct {
	ID                    string                    `json:"id" binding:"required"`
	Name                  string                    `json:"name" binding:"required"`
	Description           string                    `json:"description"`
	Icon                  string                    `json:"icon"`
	Tags                  []string                  `json:"tags"`
	IsPublic              bool                      `json:"is_public"`
	Adapter               string                    `json:"adapter" binding:"required"`
	Extends               string                    `json:"extends"`
	CredentialID          string                    `json:"credential_id"`
	Model                 profile.ModelConfig       `json:"model"`
	MCPServers            []profile.MCPServerConfig `json:"mcp_servers"`
	Permissions           profile.PermissionConfig  `json:"permissions"`
	Resources             profile.ResourceConfig    `json:"resources"`
	SystemPrompt          string                    `json:"system_prompt"`
	AppendSystemPrompt    string                    `json:"append_system_prompt"`
	BaseInstructions      string                    `json:"base_instructions"`
	DeveloperInstructions string                    `json:"developer_instructions"`
	Features              profile.FeatureConfig     `json:"features"`
	ConfigOverrides       map[string]string         `json:"config_overrides"`
	OutputFormat          string                    `json:"output_format"`
	OutputSchema          string                    `json:"output_schema"`
	Debug                 profile.DebugConfig       `json:"debug"`
}

// Create godoc
// @Summary Create a profile
// @Description Create a new runtime configuration profile for agents
// @Tags Profiles
// @Accept json
// @Produce json
// @Param profile body CreateProfileRequest true "Profile configuration"
// @Success 201 {object} Response{data=profile.Profile}
// @Failure 400 {object} Response
// @Router /profiles [post]
func (h *ProfileHandler) Create(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	p := &profile.Profile{
		ID:                    req.ID,
		Name:                  req.Name,
		Description:           req.Description,
		Icon:                  req.Icon,
		Tags:                  req.Tags,
		IsPublic:              req.IsPublic,
		Adapter:               req.Adapter,
		Extends:               req.Extends,
		CredentialID:          req.CredentialID,
		Model:                 req.Model,
		MCPServers:            req.MCPServers,
		Permissions:           req.Permissions,
		Resources:             req.Resources,
		SystemPrompt:          req.SystemPrompt,
		AppendSystemPrompt:    req.AppendSystemPrompt,
		BaseInstructions:      req.BaseInstructions,
		DeveloperInstructions: req.DeveloperInstructions,
		Features:              req.Features,
		ConfigOverrides:       req.ConfigOverrides,
		OutputFormat:          req.OutputFormat,
		OutputSchema:          req.OutputSchema,
		Debug:                 req.Debug,
	}

	if err := h.manager.Create(p); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Created(c, p)
}

// Update godoc
// @Summary Update a profile
// @Description Update an existing profile configuration
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID"
// @Param profile body CreateProfileRequest true "Profile configuration"
// @Success 200 {object} Response{data=profile.Profile}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /profiles/{id} [put]
func (h *ProfileHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	p := &profile.Profile{
		ID:                    id,
		Name:                  req.Name,
		Description:           req.Description,
		Icon:                  req.Icon,
		Tags:                  req.Tags,
		IsPublic:              req.IsPublic,
		Adapter:               req.Adapter,
		Extends:               req.Extends,
		CredentialID:          req.CredentialID,
		Model:                 req.Model,
		MCPServers:            req.MCPServers,
		Permissions:           req.Permissions,
		Resources:             req.Resources,
		SystemPrompt:          req.SystemPrompt,
		AppendSystemPrompt:    req.AppendSystemPrompt,
		BaseInstructions:      req.BaseInstructions,
		DeveloperInstructions: req.DeveloperInstructions,
		Features:              req.Features,
		ConfigOverrides:       req.ConfigOverrides,
		OutputFormat:          req.OutputFormat,
		OutputSchema:          req.OutputSchema,
		Debug:                 req.Debug,
	}

	if err := h.manager.Update(p); err != nil {
		if err == profile.ErrProfileNotFound {
			NotFound(c, err.Error())
			return
		}
		BadRequest(c, err.Error())
		return
	}

	Success(c, p)
}

// Delete godoc
// @Summary Delete a profile
// @Description Delete an existing profile (built-in profiles cannot be deleted)
// @Tags Profiles
// @Param id path string true "Profile ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /profiles/{id} [delete]
func (h *ProfileHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		if err == profile.ErrProfileNotFound {
			NotFound(c, err.Error())
			return
		}
		BadRequest(c, err.Error())
		return
	}

	Success(c, map[string]interface{}{"id": id, "deleted": true})
}

// CloneRequest 克隆请求
type CloneRequest struct {
	NewID   string `json:"new_id" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

// Clone godoc
// @Summary Clone a profile
// @Description Create a copy of an existing profile with a new ID and name
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path string true "Source Profile ID"
// @Param clone body CloneRequest true "Clone configuration"
// @Success 201 {object} Response{data=profile.Profile}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /profiles/{id}/clone [post]
func (h *ProfileHandler) Clone(c *gin.Context) {
	id := c.Param("id")

	var req CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	p, err := h.manager.Clone(id, req.NewID, req.NewName)
	if err != nil {
		if err == profile.ErrProfileNotFound {
			NotFound(c, err.Error())
			return
		}
		BadRequest(c, err.Error())
		return
	}

	Created(c, p)
}
