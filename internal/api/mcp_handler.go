package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/mcp"
)

// MCPHandler MCP Server API 处理器
type MCPHandler struct {
	manager *mcp.Manager
}

// NewMCPHandler 创建 MCPHandler
func NewMCPHandler(manager *mcp.Manager) *MCPHandler {
	return &MCPHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *MCPHandler) RegisterRoutes(r *gin.RouterGroup) {
	servers := r.Group("/mcp-servers")
	{
		servers.GET("", h.List)
		servers.GET("/stats", h.Stats)
		servers.POST("", h.Create)
		servers.GET("/:id", h.Get)
		servers.PUT("/:id", h.Update)
		servers.DELETE("/:id", h.Delete)
		servers.POST("/:id/clone", h.Clone)
		servers.POST("/:id/test", h.Test)
	}
}

// List 列出所有 MCP Servers
// GET /api/v1/mcp-servers
func (h *MCPHandler) List(c *gin.Context) {
	category := c.Query("category")
	enabledOnly := c.Query("enabled") == "true"

	var servers []*mcp.Server

	if category != "" {
		servers = h.manager.ListByCategory(mcp.Category(category))
	} else if enabledOnly {
		servers = h.manager.ListEnabled()
	} else {
		servers = h.manager.List()
	}

	Success(c, servers)
}

// Get 获取单个 MCP Server
// GET /api/v1/mcp-servers/:id
func (h *MCPHandler) Get(c *gin.Context) {
	id := c.Param("id")

	server, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, server)
}

// Create 创建 MCP Server
// POST /api/v1/mcp-servers
func (h *MCPHandler) Create(c *gin.Context) {
	var req mcp.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	server, err := h.manager.Create(&req)
	if err != nil {
		HandleError(c, err)
		return
	}

	Created(c, server)
}

// Update 更新 MCP Server
// PUT /api/v1/mcp-servers/:id
func (h *MCPHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req mcp.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	server, err := h.manager.Update(id, &req)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, server)
}

// Delete 删除 MCP Server
// DELETE /api/v1/mcp-servers/:id
func (h *MCPHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"deleted": id})
}

// Clone 克隆 MCP Server
// POST /api/v1/mcp-servers/:id/clone
func (h *MCPHandler) Clone(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewID   string `json:"new_id" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	server, err := h.manager.Clone(id, req.NewID, req.NewName)
	if err != nil {
		HandleError(c, err)
		return
	}

	Created(c, server)
}

// Stats 获取 MCP Server 统计信息
// GET /api/v1/mcp-servers/stats
func (h *MCPHandler) Stats(c *gin.Context) {
	stats := h.manager.Stats()
	Success(c, stats)
}

// Test 测试 MCP Server 连接
// POST /api/v1/mcp-servers/:id/test
func (h *MCPHandler) Test(c *gin.Context) {
	id := c.Param("id")

	result, err := h.manager.Test(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, result)
}
