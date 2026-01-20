package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// SkillHandler Skill API 处理器
type SkillHandler struct {
	manager *skill.Manager
	store   *skill.SkillStore
}

// NewSkillHandler 创建 SkillHandler
func NewSkillHandler(manager *skill.Manager) *SkillHandler {
	return &SkillHandler{
		manager: manager,
		store:   skill.NewSkillStore(manager),
	}
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

	// Skill Store API
	store := r.Group("/skill-store")
	{
		store.GET("/sources", h.ListSources)
		store.POST("/sources", h.AddSource)
		store.DELETE("/sources/:id", h.RemoveSource)
		store.GET("/skills", h.ListRemoteSkills)
		store.GET("/skills/:sourceId", h.ListSourceSkills)
		store.POST("/install", h.InstallSkill)
		store.DELETE("/uninstall/:id", h.UninstallSkill)
		store.POST("/refresh/:sourceId", h.RefreshSource)
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

// ============== Skill Store API ==============

// ListSources 列出所有 Skill 源
// GET /api/v1/skill-store/sources
func (h *SkillHandler) ListSources(c *gin.Context) {
	sources := h.store.ListSources()
	Success(c, sources)
}

// AddSource 添加 Skill 源
// POST /api/v1/skill-store/sources
func (h *SkillHandler) AddSource(c *gin.Context) {
	var req skill.SkillSource
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证必填字段
	if req.ID == "" || req.Owner == "" || req.Repo == "" {
		Error(c, http.StatusBadRequest, "id, owner, repo are required")
		return
	}

	h.store.AddSource(&req)
	Created(c, req)
}

// RemoveSource 移除 Skill 源
// DELETE /api/v1/skill-store/sources/:id
func (h *SkillHandler) RemoveSource(c *gin.Context) {
	id := c.Param("id")

	// 不允许删除官方源
	source, ok := h.store.GetSource(id)
	if !ok {
		Error(c, http.StatusNotFound, "source not found")
		return
	}
	if source.Type == "official" {
		Error(c, http.StatusForbidden, "cannot remove official source")
		return
	}

	h.store.RemoveSource(id)
	Success(c, gin.H{"deleted": id})
}

// ListRemoteSkills 列出所有远程 Skills
// GET /api/v1/skill-store/skills
func (h *SkillHandler) ListRemoteSkills(c *gin.Context) {
	skills, err := h.store.FetchAllSkills(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, skills)
}

// ListSourceSkills 列出指定源的 Skills
// GET /api/v1/skill-store/skills/:sourceId
func (h *SkillHandler) ListSourceSkills(c *gin.Context) {
	sourceID := c.Param("sourceId")

	skills, err := h.store.FetchSkills(c.Request.Context(), sourceID)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	Success(c, skills)
}

// InstallSkill 安装远程 Skill
// POST /api/v1/skill-store/install
func (h *SkillHandler) InstallSkill(c *gin.Context) {
	var req struct {
		SourceID string `json:"source_id" binding:"required"`
		SkillID  string `json:"skill_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	installed, err := h.store.InstallSkill(c.Request.Context(), req.SourceID, req.SkillID)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Created(c, installed)
}

// UninstallSkill 卸载 Skill
// DELETE /api/v1/skill-store/uninstall/:id
func (h *SkillHandler) UninstallSkill(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.UninstallSkill(id); err != nil {
		if err == skill.ErrSkillNotFound {
			Error(c, http.StatusNotFound, err.Error())
			return
		}
		if err == skill.ErrSkillIsBuiltIn {
			Error(c, http.StatusForbidden, err.Error())
			return
		}
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{"uninstalled": id})
}

// RefreshSource 刷新源缓存
// POST /api/v1/skill-store/refresh/:sourceId
func (h *SkillHandler) RefreshSource(c *gin.Context) {
	sourceID := c.Param("sourceId")

	if err := h.store.RefreshCache(c.Request.Context(), sourceID); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{"refreshed": sourceID})
}
