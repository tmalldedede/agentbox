package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/cron"
)

// CronHandler Cron 任务处理器
type CronHandler struct {
	manager *cron.Manager
}

// NewCronHandler 创建 Cron 处理器
func NewCronHandler(manager *cron.Manager) *CronHandler {
	return &CronHandler{manager: manager}
}

// RegisterRoutes 注册路由
func (h *CronHandler) RegisterRoutes(rg *gin.RouterGroup) {
	crons := rg.Group("/crons")
	{
		crons.POST("", h.Create)
		crons.GET("", h.List)
		crons.GET("/:id", h.Get)
		crons.PUT("/:id", h.Update)
		crons.DELETE("/:id", h.Delete)
		crons.POST("/:id/trigger", h.Trigger)
	}
}

// Create 创建定时任务
// @Summary 创建定时任务
// @Tags Cron
// @Accept json
// @Produce json
// @Param request body cron.CreateJobRequest true "任务配置"
// @Success 201 {object} Response{data=cron.Job}
// @Router /api/v1/admin/crons [post]
func (h *CronHandler) Create(c *gin.Context) {
	var req cron.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.manager.Create(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Created(c, job)
}

// Get 获取定时任务
// @Summary 获取定时任务
// @Tags Cron
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} Response{data=cron.Job}
// @Router /api/v1/admin/crons/{id} [get]
func (h *CronHandler) Get(c *gin.Context) {
	id := c.Param("id")

	job, err := h.manager.Get(id)
	if err != nil {
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	Success(c, job)
}

// Update 更新定时任务
// @Summary 更新定时任务
// @Tags Cron
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param request body cron.UpdateJobRequest true "更新配置"
// @Success 200 {object} Response{data=cron.Job}
// @Router /api/v1/admin/crons/{id} [put]
func (h *CronHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req cron.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.manager.Update(id, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, job)
}

// Delete 删除定时任务
// @Summary 删除定时任务
// @Tags Cron
// @Param id path string true "任务 ID"
// @Success 204
// @Router /api/v1/admin/crons/{id} [delete]
func (h *CronHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// List 列出所有定时任务
// @Summary 列出所有定时任务
// @Tags Cron
// @Produce json
// @Success 200 {object} Response{data=[]cron.Job}
// @Router /api/v1/admin/crons [get]
func (h *CronHandler) List(c *gin.Context) {
	jobs, err := h.manager.List()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, jobs)
}

// Trigger 立即触发任务
// @Summary 立即触发任务
// @Tags Cron
// @Param id path string true "任务 ID"
// @Success 200 {object} Response
// @Router /api/v1/admin/crons/{id}/trigger [post]
func (h *CronHandler) Trigger(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.TriggerNow(id); err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	Success(c, gin.H{"message": "triggered"})
}
