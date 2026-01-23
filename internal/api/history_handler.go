package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/history"
)

// HistoryHandler handles execution history API requests
type HistoryHandler struct {
	manager *history.Manager
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(manager *history.Manager) *HistoryHandler {
	return &HistoryHandler{manager: manager}
}

// RegisterRoutes registers history routes
func (h *HistoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	hist := r.Group("/history")
	{
		hist.GET("", h.List)
		hist.GET("/stats", h.Stats)
		hist.GET("/:id", h.Get)
		hist.DELETE("/:id", h.Delete)
	}
}

// List godoc
// @Summary List execution history
// @Description Get a list of all execution history entries
// @Tags History
// @Produce json
// @Param source_type query string false "Filter by source type (session/agent)"
// @Param source_id query string false "Filter by source ID"
// @Param agent_id query string false "Filter by agent ID"
// @Param engine query string false "Filter by engine"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results (default 50)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} Response{data=[]history.Entry}
// @Router /history [get]
func (h *HistoryHandler) List(c *gin.Context) {
	var filter history.ListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Default limit
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	entries, err := h.manager.List(&filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	count, _ := h.manager.Count(&filter)

	Success(c, gin.H{
		"entries": entries,
		"total":   count,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// Get godoc
// @Summary Get a history entry
// @Description Get detailed information about a specific execution
// @Tags History
// @Produce json
// @Param id path string true "Entry ID"
// @Success 200 {object} Response{data=history.Entry}
// @Failure 404 {object} Response
// @Router /history/{id} [get]
func (h *HistoryHandler) Get(c *gin.Context) {
	id := c.Param("id")

	entry, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, entry)
}

// Delete godoc
// @Summary Delete a history entry
// @Description Delete an execution history entry
// @Tags History
// @Param id path string true "Entry ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /history/{id} [delete]
func (h *HistoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, gin.H{"id": id, "deleted": true})
}

// Stats godoc
// @Summary Get execution statistics
// @Description Get aggregate statistics about executions
// @Tags History
// @Produce json
// @Param source_type query string false "Filter by source type"
// @Param agent_id query string false "Filter by agent ID"
// @Param engine query string false "Filter by engine"
// @Success 200 {object} Response{data=history.Stats}
// @Router /history/stats [get]
func (h *HistoryHandler) Stats(c *gin.Context) {
	var filter history.ListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		BadRequest(c, err.Error())
		return
	}

	stats, err := h.manager.GetStats(&filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, stats)
}
