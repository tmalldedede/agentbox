package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// SkillHandler Skill API 处理器
type SkillHandler struct {
	manager      *skill.Manager
	store        *skill.SkillStore
	resolver     *skill.Resolver
	containerMgr container.Manager
}

// NewSkillHandler 创建 SkillHandler
func NewSkillHandler(manager *skill.Manager) *SkillHandler {
	return &SkillHandler{
		manager: manager,
		store:   skill.NewSkillStore(manager),
	}
}

// SetContainerManager 设置容器管理器（用于依赖检查）
func (h *SkillHandler) SetContainerManager(mgr container.Manager) {
	h.containerMgr = mgr
	h.resolver = skill.NewResolver(mgr)
}

// RegisterRoutes 注册路由
func (h *SkillHandler) RegisterRoutes(r *gin.RouterGroup) {
	skills := r.Group("/skills")
	{
		skills.GET("", h.List)
		skills.POST("", h.Create)
		skills.GET("/status", h.GetStatus)          // 完整状态报告（放在 :id 之前）
		skills.GET("/stats", h.GetStats)            // 统计信息
		skills.GET("/bins", h.ListRequiredBins)     // 列出所有需要的二进制
		skills.GET("/:id", h.Get)
		skills.PUT("/:id", h.Update)
		skills.DELETE("/:id", h.Delete)
		skills.POST("/:id/clone", h.Clone)
		skills.GET("/:id/export", h.Export)
		skills.POST("/:id/check-deps", h.CheckDeps) // 依赖检查
		skills.PUT("/:id/config", h.UpdateConfig)   // 更新 Skill 配置
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
// Query params:
//   - level: metadata | body | full (默认 body)
//   - category: coding | review | docs | security | testing | other
//   - source: extra | bundled | managed | workspace
//   - enabled: true | false
func (h *SkillHandler) List(c *gin.Context) {
	level := c.DefaultQuery("level", "body")
	category := c.Query("category")
	source := c.Query("source")
	enabledOnly := c.Query("enabled") == "true"

	loader := h.manager.GetLoader()

	// 根据加载级别返回不同的数据
	switch skill.LoadLevel(level) {
	case skill.LoadLevelMetadata:
		// 仅返回元数据（快速列表）
		var metas []*skill.SkillMetadata

		if category != "" {
			metas = loader.ListMetadataByCategory(skill.Category(category))
		} else if source != "" {
			metas = loader.ListMetadataBySource(skill.SkillSource(source))
		} else if enabledOnly {
			metas = loader.ListEnabledMetadata()
		} else {
			metas = loader.ListMetadata()
		}

		Success(c, metas)
		return

	case skill.LoadLevelFull:
		// 完整加载（包含所有引用文件）
		skills, err := loader.LoadSkillsByLevel(skill.LoadLevelFull)
		if err != nil {
			HandleError(c, apperr.Wrap(err, "failed to load skills"))
			return
		}
		Success(c, skills)
		return

	default:
		// 默认返回 body 级别
		var skills []*skill.Skill

		if category != "" {
			skills = h.manager.ListByCategory(skill.Category(category))
		} else if source != "" {
			skills = h.manager.ListBySource(skill.SkillSource(source))
		} else if enabledOnly {
			skills = h.manager.ListEnabled()
		} else {
			skills = h.manager.List()
		}

		Success(c, skills)
	}
}

// Get 获取单个 Skill
// GET /api/v1/skills/:id
// Query params:
//   - level: metadata | body | full (默认 body)
func (h *SkillHandler) Get(c *gin.Context) {
	id := c.Param("id")
	level := c.DefaultQuery("level", "body")

	loader := h.manager.GetLoader()

	switch skill.LoadLevel(level) {
	case skill.LoadLevelMetadata:
		meta, err := loader.GetMetadata(id)
		if err != nil {
			HandleError(c, err)
			return
		}
		Success(c, meta)

	case skill.LoadLevelFull:
		s, err := loader.LoadFull(id)
		if err != nil {
			HandleError(c, err)
			return
		}
		Success(c, s)

	default:
		s, err := loader.LoadBody(id)
		if err != nil {
			HandleError(c, err)
			return
		}
		Success(c, s)
	}
}

// Create 创建 Skill
// POST /api/v1/skills
func (h *SkillHandler) Create(c *gin.Context) {
	var req skill.CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	s, err := h.manager.Create(&req)
	if err != nil {
		HandleError(c, err)
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
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	s, err := h.manager.Update(id, &req)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, s)
}

// Delete 删除 Skill
// DELETE /api/v1/skills/:id
func (h *SkillHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
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
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	s, err := h.manager.Clone(id, req.NewID, req.NewName)
	if err != nil {
		HandleError(c, err)
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
		HandleError(c, err)
		return
	}

	// 返回 SKILL.md 格式
	c.Header("Content-Type", "text/markdown")
	c.Header("Content-Disposition", "attachment; filename=SKILL.md")
	c.String(http.StatusOK, s.ToSkillMD())
}

// CheckDeps 检查 Skill 依赖
// POST /api/v1/skills/:id/check-deps
// Body: { "container_id": "xxx" }
func (h *SkillHandler) CheckDeps(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		ContainerID string `json:"container_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许不传 body，此时返回快速检查结果
	}

	s, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 如果没有 resolver 或没有指定容器，返回快速检查结果
	if h.resolver == nil || req.ContainerID == "" {
		result := &skill.CheckResult{
			SkillID:   id,
			Satisfied: !s.Requirements.HasRequirements(),
		}
		if s.Requirements.HasRequirements() {
			result.Missing = &skill.MissingDeps{
				Bins:   s.Requirements.Bins,
				Env:    s.Requirements.Env,
				Config: s.Requirements.Config,
				Pip:    s.Requirements.Pip,
				Npm:    s.Requirements.Npm,
			}
		}
		Success(c, result)
		return
	}

	// 在容器内执行依赖检查
	result, err := h.resolver.Check(c.Request.Context(), s, req.ContainerID)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to check dependencies"))
		return
	}

	Success(c, result)
}

// GetStats 获取 Skill 统计信息
// GET /api/v1/skills/stats
func (h *SkillHandler) GetStats(c *gin.Context) {
	stats := h.manager.GetSkillStats()
	Success(c, stats)
}

// GetStatus 获取完整的 Skill 状态报告（借鉴 Clawdbot）
// GET /api/v1/skills/status
func (h *SkillHandler) GetStatus(c *gin.Context) {
	skills := h.manager.List()

	// 如果没有 resolver，创建一个临时的
	resolver := h.resolver
	if resolver == nil {
		resolver = skill.NewResolver(nil)
	}

	// 获取配置（TODO: 从持久化存储加载）
	configs := make(map[string]*skill.SkillConfig)

	report := resolver.BuildStatusReport(c.Request.Context(), skills, configs)
	Success(c, report)
}

// ListRequiredBins 列出所有 Skill 需要的二进制
// GET /api/v1/skills/bins
func (h *SkillHandler) ListRequiredBins(c *gin.Context) {
	skills := h.manager.List()

	bins := make(map[string]bool)
	for _, s := range skills {
		if s.Requirements == nil {
			continue
		}
		for _, bin := range s.Requirements.Bins {
			bins[bin] = true
		}
		for _, bin := range s.Requirements.AnyBins {
			bins[bin] = true
		}
		// 收集安装规范中的二进制
		for _, spec := range s.Install {
			for _, bin := range spec.Bins {
				bins[bin] = true
			}
		}
	}

	// 转换为排序数组
	result := make([]string, 0, len(bins))
	for bin := range bins {
		result = append(result, bin)
	}
	// 简单排序
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	Success(c, gin.H{"bins": result})
}

// UpdateConfig 更新 Skill 配置
// PUT /api/v1/skills/:id/config
func (h *SkillHandler) UpdateConfig(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Enabled *bool             `json:"enabled,omitempty"`
		APIKey  *string           `json:"api_key,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	// 验证 Skill 存在
	s, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	// TODO: 持久化配置到存储
	// 目前返回成功响应
	config := skill.SkillConfig{
		Enabled: true,
	}
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.APIKey != nil {
		config.APIKey = *req.APIKey
	}
	if req.Env != nil {
		config.Env = req.Env
	}

	Success(c, gin.H{
		"skill_id": s.ID,
		"config":   config,
	})
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
	var req skill.RepoSource
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	// 验证必填字段
	if req.ID == "" || req.Owner == "" || req.Repo == "" {
		HandleError(c, apperr.Validation("id, owner, repo are required"))
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
		HandleError(c, apperr.NotFound("source"))
		return
	}
	if source.Type == "official" {
		HandleError(c, apperr.Forbidden("cannot remove official source"))
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
		HandleError(c, apperr.Wrap(err, "failed to fetch remote skills"))
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
		HandleError(c, apperr.Wrap(err, "failed to fetch skills from source"))
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
		HandleError(c, apperr.Validation(err.Error()))
		return
	}

	installed, err := h.store.InstallSkill(c.Request.Context(), req.SourceID, req.SkillID)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to install skill"))
		return
	}

	Created(c, installed)
}

// UninstallSkill 卸载 Skill
// DELETE /api/v1/skill-store/uninstall/:id
func (h *SkillHandler) UninstallSkill(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.UninstallSkill(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"uninstalled": id})
}

// RefreshSource 刷新源缓存
// POST /api/v1/skill-store/refresh/:sourceId
func (h *SkillHandler) RefreshSource(c *gin.Context) {
	sourceID := c.Param("sourceId")

	if err := h.store.RefreshCache(c.Request.Context(), sourceID); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to refresh source cache"))
		return
	}

	Success(c, gin.H{"refreshed": sourceID})
}
