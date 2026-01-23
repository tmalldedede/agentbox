package api

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/session"
)

// SystemHandler 系统维护 API 处理器
type SystemHandler struct {
	containerMgr container.Manager
	sessionMgr   *session.Manager
	gc           *container.GarbageCollector
	startTime    time.Time
}

// NewSystemHandler 创建 SystemHandler
func NewSystemHandler(containerMgr container.Manager, sessionMgr *session.Manager, gc *container.GarbageCollector) *SystemHandler {
	return &SystemHandler{
		containerMgr: containerMgr,
		sessionMgr:   sessionMgr,
		gc:           gc,
		startTime:    time.Now(),
	}
}

// RegisterRoutes 注册路由
func (h *SystemHandler) RegisterRoutes(r *gin.RouterGroup) {
	system := r.Group("/system")
	{
		system.GET("/health", h.Health)
		system.GET("/stats", h.Stats)
		system.POST("/cleanup/containers", h.CleanupContainers)
		system.POST("/cleanup/images", h.CleanupImages)
		system.GET("/gc/stats", h.GCStats)
		system.POST("/gc/trigger", h.TriggerGC)
		system.POST("/gc/preview", h.PreviewGC)
		system.PUT("/gc/config", h.UpdateGCConfig)
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Uptime    string            `json:"uptime"`
	Docker    DockerHealth      `json:"docker"`
	Resources ResourcesHealth   `json:"resources"`
	Checks    map[string]string `json:"checks"`
}

// DockerHealth Docker 健康状态
type DockerHealth struct {
	Status      string `json:"status"`
	Version     string `json:"version,omitempty"`
	Containers  int    `json:"containers"`
	Images      int    `json:"images"`
	ErrorMsg    string `json:"error,omitempty"`
}

// ResourcesHealth 资源健康状态
type ResourcesHealth struct {
	MemoryUsageMB  uint64 `json:"memory_usage_mb"`
	NumGoroutines  int    `json:"num_goroutines"`
	NumCPU         int    `json:"num_cpu"`
}

// Health 健康检查
// GET /api/v1/system/health
func (h *SystemHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	resp := HealthResponse{
		Status: "healthy",
		Uptime: time.Since(h.startTime).Round(time.Second).String(),
		Checks: make(map[string]string),
	}

	// Docker 健康检查
	dockerHealth := DockerHealth{
		Status: "healthy",
	}

	if err := h.containerMgr.Ping(ctx); err != nil {
		dockerHealth.Status = "unhealthy"
		dockerHealth.ErrorMsg = err.Error()
		resp.Status = "degraded"
		resp.Checks["docker"] = "failed: " + err.Error()
	} else {
		resp.Checks["docker"] = "ok"

		// 获取容器和镜像统计
		containers, err := h.containerMgr.ListContainers(ctx)
		if err == nil {
			dockerHealth.Containers = len(containers)
		}

		images, err := h.containerMgr.ListImages(ctx)
		if err == nil {
			dockerHealth.Images = len(images)
		}
	}
	resp.Docker = dockerHealth

	// 资源统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	resp.Resources = ResourcesHealth{
		MemoryUsageMB: memStats.Alloc / 1024 / 1024,
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
	}
	resp.Checks["memory"] = "ok"
	resp.Checks["goroutines"] = "ok"

	Success(c, resp)
}

// StatsResponse 系统统计响应
type StatsResponse struct {
	Sessions   SessionStats   `json:"sessions"`
	Containers ContainerStats `json:"containers"`
	Images     ImageStats     `json:"images"`
	System     SystemStats    `json:"system"`
}

// SessionStats 会话统计
type SessionStats struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Stopped  int `json:"stopped"`
	Error    int `json:"error"`
	Creating int `json:"creating"`
}

// ContainerStats 容器统计
type ContainerStats struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Other   int `json:"other"`
}

// ImageStats 镜像统计
type ImageStats struct {
	Total       int    `json:"total"`
	AgentImages int    `json:"agent_images"`
	TotalSize   int64  `json:"total_size"`
	InUse       int    `json:"in_use"`
}

// SystemStats 系统统计
type SystemStats struct {
	Uptime        string `json:"uptime"`
	MemoryUsageMB uint64 `json:"memory_usage_mb"`
	GoVersion     string `json:"go_version"`
	NumCPU        int    `json:"num_cpu"`
	NumGoroutines int    `json:"num_goroutines"`
}

// Stats 系统统计
// GET /api/v1/system/stats
func (h *SystemHandler) Stats(c *gin.Context) {
	ctx := c.Request.Context()

	resp := StatsResponse{}

	// 会话统计
	sessions, err := h.sessionMgr.List(ctx, nil)
	if err == nil {
		resp.Sessions.Total = len(sessions)
		for _, s := range sessions {
			switch s.Status {
			case session.StatusRunning:
				resp.Sessions.Running++
			case session.StatusStopped:
				resp.Sessions.Stopped++
			case session.StatusError:
				resp.Sessions.Error++
			case session.StatusCreating:
				resp.Sessions.Creating++
			}
		}
	}

	// 容器统计
	containers, err := h.containerMgr.ListContainers(ctx)
	if err == nil {
		resp.Containers.Total = len(containers)
		for _, c := range containers {
			switch c.Status {
			case container.StatusRunning:
				resp.Containers.Running++
			case container.StatusExited:
				resp.Containers.Stopped++
			default:
				resp.Containers.Other++
			}
		}
	}

	// 镜像统计
	images, err := h.containerMgr.ListImages(ctx)
	if err == nil {
		resp.Images.Total = len(images)
		for _, img := range images {
			resp.Images.TotalSize += img.Size
			if img.IsAgentImage {
				resp.Images.AgentImages++
			}
			if img.InUse {
				resp.Images.InUse++
			}
		}
	}

	// 系统统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	resp.System = SystemStats{
		Uptime:        time.Since(h.startTime).Round(time.Second).String(),
		MemoryUsageMB: memStats.Alloc / 1024 / 1024,
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
	}

	Success(c, resp)
}

// CleanupContainersResponse 清理容器响应
type CleanupContainersResponse struct {
	Removed []string `json:"removed"`
	Errors  []string `json:"errors,omitempty"`
}

// CleanupContainers 清理孤立容器
// POST /api/v1/system/cleanup/containers
func (h *SystemHandler) CleanupContainers(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取所有会话的容器 ID
	sessions, err := h.sessionMgr.List(ctx, nil)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to list sessions"))
		return
	}

	sessionContainerIDs := make(map[string]bool)
	for _, s := range sessions {
		if s.ContainerID != "" {
			sessionContainerIDs[s.ContainerID] = true
		}
	}

	// 获取所有 AgentBox 管理的容器
	containers, err := h.containerMgr.ListContainers(ctx)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to list containers"))
		return
	}

	resp := CleanupContainersResponse{
		Removed: make([]string, 0),
		Errors:  make([]string, 0),
	}

	// 删除孤立容器（有 agentbox 标签但没有对应会话的容器）
	for _, ctr := range containers {
		if !sessionContainerIDs[ctr.ID] {
			// 先停止再删除
			_ = h.containerMgr.Stop(ctx, ctr.ID)
			if err := h.containerMgr.Remove(ctx, ctr.ID); err != nil {
				resp.Errors = append(resp.Errors, ctr.ID[:12]+": "+err.Error())
			} else {
				resp.Removed = append(resp.Removed, ctr.ID[:12])
			}
		}
	}

	Success(c, resp)
}

// CleanupImagesRequest 清理镜像请求
type CleanupImagesRequest struct {
	UnusedOnly bool `json:"unused_only"`
}

// CleanupImagesResponse 清理镜像响应
type CleanupImagesResponse struct {
	Removed     []string `json:"removed"`
	SpaceFreed  int64    `json:"space_freed"`
	Errors      []string `json:"errors,omitempty"`
}

// CleanupImages 清理未使用的镜像
// POST /api/v1/system/cleanup/images
func (h *SystemHandler) CleanupImages(c *gin.Context) {
	ctx := c.Request.Context()

	var req CleanupImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认只清理未使用的镜像
		req.UnusedOnly = true
	}

	images, err := h.containerMgr.ListImages(ctx)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to list images"))
		return
	}

	resp := CleanupImagesResponse{
		Removed: make([]string, 0),
		Errors:  make([]string, 0),
	}

	for _, img := range images {
		// 跳过正在使用的镜像
		if img.InUse {
			continue
		}

		// 跳过 Agent 镜像（除非明确要求删除所有）
		if img.IsAgentImage && req.UnusedOnly {
			continue
		}

		// 获取镜像名称
		imageName := img.ID
		if len(img.Tags) > 0 {
			imageName = img.Tags[0]
		}

		if err := h.containerMgr.RemoveImage(ctx, img.ID); err != nil {
			resp.Errors = append(resp.Errors, imageName+": "+err.Error())
		} else {
			resp.Removed = append(resp.Removed, imageName)
			resp.SpaceFreed += img.Size
		}
	}

	Success(c, resp)
}

// GCStatsResponse GC 状态响应
type GCStatsResponse struct {
	Running           bool      `json:"running"`
	LastRunAt         time.Time `json:"last_run_at"`
	NextRunAt         time.Time `json:"next_run_at"`
	ContainersRemoved int       `json:"containers_removed"`
	TotalRuns         int       `json:"total_runs"`
	Errors            []string  `json:"errors"`
	Config            struct {
		IntervalSeconds    int `json:"interval_seconds"`
		ContainerTTLSeconds int `json:"container_ttl_seconds"`
		IdleTimeoutSeconds int `json:"idle_timeout_seconds"`
	} `json:"config"`
}

// GCStats 获取 GC 状态
// GET /api/v1/admin/system/gc/stats
func (h *SystemHandler) GCStats(c *gin.Context) {
	stats := h.gc.Stats()

	resp := GCStatsResponse{
		Running:           stats.Running,
		LastRunAt:         stats.LastRunAt,
		NextRunAt:         stats.NextRunAt,
		ContainersRemoved: stats.ContainersRemoved,
		TotalRuns:         stats.TotalRuns,
		Errors:            stats.Errors,
	}
	resp.Config.IntervalSeconds = int(stats.Config.Interval.Seconds())
	resp.Config.ContainerTTLSeconds = int(stats.Config.ContainerTTL.Seconds())
	resp.Config.IdleTimeoutSeconds = int(stats.Config.IdleTimeout.Seconds())

	Success(c, resp)
}

// TriggerGCResponse 手动触发 GC 响应
type TriggerGCResponse struct {
	Removed int `json:"removed"`
}

// TriggerGC 手动触发一次 GC
// POST /api/v1/admin/system/gc/trigger
func (h *SystemHandler) TriggerGC(c *gin.Context) {
	ctx := c.Request.Context()

	statsBefore := h.gc.Stats()

	if err := h.gc.RunOnce(ctx); err != nil {
		HandleError(c, apperr.Wrap(err, "gc trigger failed"))
		return
	}

	statsAfter := h.gc.Stats()
	removed := statsAfter.ContainersRemoved - statsBefore.ContainersRemoved

	Success(c, TriggerGCResponse{Removed: removed})
}

// PreviewGC 预览 GC 候选容器（不实际删除）
// POST /api/v1/admin/system/gc/preview
func (h *SystemHandler) PreviewGC(c *gin.Context) {
	ctx := c.Request.Context()

	candidates, err := h.gc.Preview(ctx)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "gc preview failed"))
		return
	}

	if candidates == nil {
		candidates = make([]container.GCCandidate, 0)
	}

	Success(c, candidates)
}

// UpdateGCConfigRequest 更新 GC 配置请求
type UpdateGCConfigRequest struct {
	IntervalSeconds    *int `json:"interval_seconds"`
	ContainerTTLSeconds *int `json:"container_ttl_seconds"`
	IdleTimeoutSeconds *int `json:"idle_timeout_seconds"`
}

// UpdateGCConfig 热更新 GC 配置
// PUT /api/v1/admin/system/gc/config
func (h *SystemHandler) UpdateGCConfig(c *gin.Context) {
	var req UpdateGCConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// 获取当前配置
	stats := h.gc.Stats()
	newConfig := stats.Config

	// 应用修改
	if req.IntervalSeconds != nil {
		if *req.IntervalSeconds < 10 {
			BadRequest(c, "interval_seconds must be >= 10")
			return
		}
		newConfig.Interval = time.Duration(*req.IntervalSeconds) * time.Second
	}
	if req.ContainerTTLSeconds != nil {
		if *req.ContainerTTLSeconds < 60 {
			BadRequest(c, "container_ttl_seconds must be >= 60")
			return
		}
		newConfig.ContainerTTL = time.Duration(*req.ContainerTTLSeconds) * time.Second
	}
	if req.IdleTimeoutSeconds != nil {
		if *req.IdleTimeoutSeconds < 30 {
			BadRequest(c, "idle_timeout_seconds must be >= 30")
			return
		}
		newConfig.IdleTimeout = time.Duration(*req.IdleTimeoutSeconds) * time.Second
	}

	h.gc.UpdateConfig(newConfig)

	// 返回更新后的状态
	updatedStats := h.gc.Stats()
	resp := GCStatsResponse{
		Running:           updatedStats.Running,
		LastRunAt:         updatedStats.LastRunAt,
		NextRunAt:         updatedStats.NextRunAt,
		ContainersRemoved: updatedStats.ContainersRemoved,
		TotalRuns:         updatedStats.TotalRuns,
		Errors:            updatedStats.Errors,
	}
	resp.Config.IntervalSeconds = int(updatedStats.Config.Interval.Seconds())
	resp.Config.ContainerTTLSeconds = int(updatedStats.Config.ContainerTTL.Seconds())
	resp.Config.IdleTimeoutSeconds = int(updatedStats.Config.IdleTimeout.Seconds())

	Success(c, resp)
}
