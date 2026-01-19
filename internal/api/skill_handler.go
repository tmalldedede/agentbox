package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// SkillHandler Skill API 处理器
type SkillHandler struct {
	manager *skill.Manager
}

// NewSkillHandler 创建 SkillHandler
func NewSkillHandler(manager *skill.Manager) *SkillHandler {
	return &SkillHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *SkillHandler) RegisterRoutes(r *gin.RouterGroup) {
	skills := r.Group("/skills")
	{
		skills.GET("", h.List)
		skills.POST("", h.Create)
		skills.GET("/:id", h.Get)
		skills.PUT("/:id", h.Update)
		skills.DELETE("/:id", h.Delete)
		skills.POST("/:id/clone", h.Clone)
		skills.GET("/:id/export", h.Export)
	}
}

// List 列出所有 Skills
// GET /api/v1/skills
func (h *SkillHandler) List(c *gin.Context) {
	category := c.Query("category")
	enabledOnly := c.Query("enabled") == "true"

	var skills []*skill.Skill

	if category != "" {
		skills = h.manager.ListByCategory(skill.Category(category))
	} else if enabledOnly {
		skills = h.manager.ListEnabled()
	} else {
		skills = h.manager.List()
	}

	Success(c, skills)
}

// Get 获取单个 Skill
// GET /api/v1/skills/:id
func (h *SkillHandler) Get(c *gin.Context) {
	id := c.Param("id")

	s, err := h.manager.Get(id)
	if err != nil {
		if err == skill.ErrSkillNotFound {
			Error(c, http.StatusNotFound, err.Error())
			return
		}
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, s)
}

// Create 创建 Skill
// POST /api/v1/skills
func (h *SkillHandler) Create(c *gin.Context) {
	var req skill.CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	s, err := h.manager.Create(&req)
	if err != nil {
		if err == skill.ErrSkillAlreadyExists {
			Error(c, http.StatusConflict, err.Error())
			return
		}
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Created(c, s)
}

// Update 更新 Skill
// PUT /api/v1/skills/:id
func (h *SkillHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req skill.UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	s, err := h.manager.Update(id, &req)
	if err != nil {
		switch err {
		case skill.ErrSkillNotFound:
			Error(c, http.StatusNotFound, err.Error())
		case skill.ErrSkillIsBuiltIn:
			Error(c, http.StatusForbidden, err.Error())
		default:
			Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	Success(c, s)
}

// Delete 删除 Skill
// DELETE /api/v1/skills/:id
func (h *SkillHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		switch err {
		case skill.ErrSkillNotFound:
			Error(c, http.StatusNotFound, err.Error())
		case skill.ErrSkillIsBuiltIn:
			Error(c, http.StatusForbidden, err.Error())
		default:
			Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	Success(c, gin.H{"deleted": id})
}

// Clone 克隆 Skill
// POST /api/v1/skills/:id/clone
func (h *SkillHandler) Clone(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewID   string `json:"new_id" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	s, err := h.manager.Clone(id, req.NewID, req.NewName)
	if err != nil {
		switch err {
		case skill.ErrSkillNotFound:
			Error(c, http.StatusNotFound, err.Error())
		case skill.ErrSkillAlreadyExists:
			Error(c, http.StatusConflict, err.Error())
		default:
			Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	Created(c, s)
}

// Export 导出 Skill 为 SKILL.md 格式
// GET /api/v1/skills/:id/export
func (h *SkillHandler) Export(c *gin.Context) {
	id := c.Param("id")

	s, err := h.manager.Get(id)
	if err != nil {
		if err == skill.ErrSkillNotFound {
			Error(c, http.StatusNotFound, err.Error())
			return
		}
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回 SKILL.md 格式
	c.Header("Content-Type", "text/markdown")
	c.Header("Content-Disposition", "attachment; filename=SKILL.md")
	c.String(http.StatusOK, s.ToSkillMD())
}
