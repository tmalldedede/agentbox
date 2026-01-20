package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/container"
)

// ImageHandler 镜像 API 处理器
type ImageHandler struct {
	containerMgr container.Manager
}

// NewImageHandler 创建 ImageHandler
func NewImageHandler(containerMgr container.Manager) *ImageHandler {
	return &ImageHandler{containerMgr: containerMgr}
}

// RegisterRoutes 注册路由
func (h *ImageHandler) RegisterRoutes(r *gin.RouterGroup) {
	images := r.Group("/images")
	{
		images.GET("", h.List)
		images.POST("/pull", h.Pull)
		images.DELETE("/:id", h.Remove)
	}
}

// List 列出所有镜像
// GET /api/v1/images
func (h *ImageHandler) List(c *gin.Context) {
	agentOnly := c.Query("agent_only") == "true"

	images, err := h.containerMgr.ListImages(c.Request.Context())
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to list images"))
		return
	}

	// 如果只需要 Agent 镜像
	if agentOnly {
		filtered := make([]*container.Image, 0)
		for _, img := range images {
			if img.IsAgentImage {
				filtered = append(filtered, img)
			}
		}
		images = filtered
	}

	Success(c, images)
}

// PullRequest 拉取镜像请求
type PullRequest struct {
	Image string `json:"image" binding:"required"`
}

// Pull 拉取镜像
// POST /api/v1/images/pull
func (h *ImageHandler) Pull(c *gin.Context) {
	var req PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	if err := h.containerMgr.PullImage(c.Request.Context(), req.Image); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to pull image"))
		return
	}

	Success(c, gin.H{"message": "Image pulled successfully", "image": req.Image})
}

// Remove 删除镜像
// DELETE /api/v1/images/:id
func (h *ImageHandler) Remove(c *gin.Context) {
	id := c.Param("id")

	if err := h.containerMgr.RemoveImage(c.Request.Context(), id); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to remove image"))
		return
	}

	Success(c, gin.H{"deleted": id})
}
