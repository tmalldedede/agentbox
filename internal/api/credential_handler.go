package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/credential"
)

// CredentialHandler Credential API 处理器
type CredentialHandler struct {
	manager *credential.Manager
}

// NewCredentialHandler 创建 CredentialHandler
func NewCredentialHandler(manager *credential.Manager) *CredentialHandler {
	return &CredentialHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *CredentialHandler) RegisterRoutes(r *gin.RouterGroup) {
	credentials := r.Group("/credentials")
	{
		credentials.GET("", h.List)
		credentials.POST("", h.Create)
		credentials.GET("/:id", h.Get)
		credentials.PUT("/:id", h.Update)
		credentials.DELETE("/:id", h.Delete)
		credentials.POST("/:id/verify", h.Verify)
	}
}

// List 列出所有凭证
// GET /api/v1/credentials
func (h *CredentialHandler) List(c *gin.Context) {
	scope := c.Query("scope")
	provider := c.Query("provider")

	var credentials []*credential.Credential

	if scope != "" {
		credentials = h.manager.ListByScope(credential.Scope(scope))
	} else if provider != "" {
		credentials = h.manager.ListByProvider(credential.Provider(provider))
	} else {
		credentials = h.manager.List()
	}

	Success(c, credentials)
}

// Get 获取单个凭证
// GET /api/v1/credentials/:id
func (h *CredentialHandler) Get(c *gin.Context) {
	id := c.Param("id")

	cred, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, cred)
}

// Create 创建凭证
// POST /api/v1/credentials
func (h *CredentialHandler) Create(c *gin.Context) {
	var req credential.CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	cred, err := h.manager.Create(&req)
	if err != nil {
		HandleError(c, err)
		return
	}

	Created(c, cred)
}

// Update 更新凭证
// PUT /api/v1/credentials/:id
func (h *CredentialHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req credential.UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	cred, err := h.manager.Update(id, &req)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, cred)
}

// Delete 删除凭证
// DELETE /api/v1/credentials/:id
func (h *CredentialHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"deleted": id})
}

// Verify 验证凭证有效性
// POST /api/v1/credentials/:id/verify
func (h *CredentialHandler) Verify(c *gin.Context) {
	id := c.Param("id")

	valid, err := h.manager.Verify(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{
		"valid":   valid,
		"message": map[bool]string{true: "Credential is valid", false: "Credential is invalid or expired"}[valid],
	})
}
