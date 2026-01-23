package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/runtime"
)

// RuntimeHandler Runtime API handler
type RuntimeHandler struct {
	manager *runtime.Manager
}

// NewRuntimeHandler creates a new runtime handler
func NewRuntimeHandler(manager *runtime.Manager) *RuntimeHandler {
	return &RuntimeHandler{manager: manager}
}

// RegisterRoutes registers runtime API routes
func (h *RuntimeHandler) RegisterRoutes(r *gin.RouterGroup) {
	runtimes := r.Group("/runtimes")
	{
		runtimes.GET("", h.List)
		runtimes.GET("/:id", h.Get)
		runtimes.POST("", h.Create)
		runtimes.PUT("/:id", h.Update)
		runtimes.DELETE("/:id", h.Delete)
		runtimes.POST("/:id/set-default", h.SetDefault)
	}
}

func (h *RuntimeHandler) List(c *gin.Context) {
	Success(c, h.manager.List())
}

func (h *RuntimeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	r, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, apperr.NotFound("runtime"))
		return
	}
	Success(c, r)
}

// CreateRuntimeRequest represents the request body for creating a runtime
type CreateRuntimeRequest struct {
	ID          string  `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	Image       string  `json:"image" binding:"required"`
	CPUs        float64 `json:"cpus,omitempty"`
	MemoryMB    int     `json:"memory_mb,omitempty"`
	Network     string  `json:"network,omitempty"`
	Privileged  bool    `json:"privileged,omitempty"`
}

func (h *RuntimeHandler) Create(c *gin.Context) {
	var req CreateRuntimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	r := &runtime.AgentRuntime{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		CPUs:        req.CPUs,
		MemoryMB:    req.MemoryMB,
		Network:     req.Network,
		Privileged:  req.Privileged,
	}

	if err := h.manager.Create(r); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to create runtime"))
		return
	}

	Created(c, r)
}

// UpdateRuntimeRequest represents the request body for updating a runtime
type UpdateRuntimeRequest struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Image       string  `json:"image,omitempty"`
	CPUs        float64 `json:"cpus,omitempty"`
	MemoryMB    int     `json:"memory_mb,omitempty"`
	Network     string  `json:"network,omitempty"`
	Privileged  *bool   `json:"privileged,omitempty"`
}

func (h *RuntimeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRuntimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	updates := &runtime.AgentRuntime{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		CPUs:        req.CPUs,
		MemoryMB:    req.MemoryMB,
		Network:     req.Network,
	}

	if err := h.manager.Update(id, updates); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to update runtime"))
		return
	}

	// Privileged 是 bool 类型，需要显式设置时单独处理
	if req.Privileged != nil {
		if err := h.manager.SetPrivileged(id, *req.Privileged); err != nil {
			HandleError(c, apperr.Wrap(err, "failed to update privileged"))
			return
		}
	}

	r, _ := h.manager.Get(id)
	Success(c, r)
}

func (h *RuntimeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.Delete(id); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to delete runtime"))
		return
	}
	Success(c, gin.H{"deleted": id})
}

func (h *RuntimeHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.SetDefault(id); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to set default runtime"))
		return
	}
	r, _ := h.manager.Get(id)
	Success(c, r)
}
