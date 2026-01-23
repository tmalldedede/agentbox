package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/task"
)

// DashboardHandler 态势感知大屏 API
type DashboardHandler struct {
	taskMgr      *task.Manager
	agentMgr     *agent.Manager
	sessionMgr   *session.Manager
	providerMgr  *provider.Manager
	containerMgr container.Manager
	historyMgr   *history.Manager
	startTime    time.Time
}

// NewDashboardHandler 创建 DashboardHandler
func NewDashboardHandler(
	taskMgr *task.Manager,
	agentMgr *agent.Manager,
	sessionMgr *session.Manager,
	providerMgr *provider.Manager,
	containerMgr container.Manager,
	historyMgr *history.Manager,
) *DashboardHandler {
	return &DashboardHandler{
		taskMgr:      taskMgr,
		agentMgr:     agentMgr,
		sessionMgr:   sessionMgr,
		providerMgr:  providerMgr,
		containerMgr: containerMgr,
		historyMgr:   historyMgr,
		startTime:    time.Now(),
	}
}

// RegisterRoutes 注册路由
func (h *DashboardHandler) RegisterRoutes(r *gin.RouterGroup) {
	dashboard := r.Group("/dashboard")
	{
		dashboard.GET("/stats", h.Stats)
	}
}

// DashboardStatsResponse 大屏聚合数据响应
type DashboardStatsResponse struct {
	Agents     DashboardAgentStats     `json:"agents"`
	Tasks      DashboardTaskStats      `json:"tasks"`
	Sessions   DashboardSessionStats   `json:"sessions"`
	Tokens     DashboardTokenStats     `json:"tokens"`
	Containers DashboardContainerStats `json:"containers"`
	Providers  []DashboardProviderInfo `json:"providers"`
	System     DashboardSystemInfo     `json:"system"`
	RecentTasks []DashboardRecentTask  `json:"recent_tasks"`
}

// DashboardAgentStats Agent 统计
type DashboardAgentStats struct {
	Total     int                      `json:"total"`
	Active    int                      `json:"active"`
	ByAdapter map[string]int           `json:"by_adapter"`
	Details   []DashboardAgentDetail   `json:"details"`
}

// DashboardAgentDetail 单个 Agent 详情
type DashboardAgentDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Adapter     string `json:"adapter"`
	Model       string `json:"model"`
	Status      string `json:"status"`
	Running     int    `json:"running"`
	Queued      int    `json:"queued"`
	Completed   int    `json:"completed"`
	Failed      int    `json:"failed"`
}

// DashboardTaskStats Task 统计
type DashboardTaskStats struct {
	Total       int            `json:"total"`
	Today       int            `json:"today"`
	ByStatus    map[string]int `json:"by_status"`
	AvgDuration float64        `json:"avg_duration_seconds"`
	SuccessRate float64        `json:"success_rate"`
}

// DashboardSessionStats Session 统计
type DashboardSessionStats struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Creating int `json:"creating"`
	Stopped  int `json:"stopped"`
	Error    int `json:"error"`
}

// DashboardTokenStats Token 使用统计
type DashboardTokenStats struct {
	TotalInput   int64 `json:"total_input"`
	TotalOutput  int64 `json:"total_output"`
	TotalTokens  int64 `json:"total_tokens"`
}

// DashboardContainerStats 容器统计
type DashboardContainerStats struct {
	Total   int   `json:"total"`
	Running int   `json:"running"`
	Stopped int   `json:"stopped"`
}

// DashboardProviderInfo Provider 状态信息
type DashboardProviderInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"` // "online" | "offline" | "degraded"
	IsConfigured bool   `json:"is_configured"`
	IsValid      bool   `json:"is_valid"`
	Category     string `json:"category"`
	Icon         string `json:"icon"`
	IconColor    string `json:"icon_color"`
}

// DashboardSystemInfo 系统信息
type DashboardSystemInfo struct {
	Uptime    string `json:"uptime"`
	StartedAt string `json:"started_at"`
}

// DashboardRecentTask 最近的任务
type DashboardRecentTask struct {
	ID        string  `json:"id"`
	AgentID   string  `json:"agent_id"`
	AgentName string  `json:"agent_name"`
	Adapter   string  `json:"adapter"`
	Prompt    string  `json:"prompt"`
	Status    string  `json:"status"`
	Duration  float64 `json:"duration_seconds"`
	CreatedAt string  `json:"created_at"`
}

// Stats 获取大屏聚合统计数据
// GET /api/v1/admin/dashboard/stats
func (h *DashboardHandler) Stats(c *gin.Context) {
	ctx := c.Request.Context()

	resp := DashboardStatsResponse{
		Providers:   make([]DashboardProviderInfo, 0),
		RecentTasks: make([]DashboardRecentTask, 0),
	}

	// ==================== Agent 统计 ====================
	agents := h.agentMgr.List()
	resp.Agents.Total = len(agents)
	resp.Agents.ByAdapter = make(map[string]int)
	resp.Agents.Details = make([]DashboardAgentDetail, 0)

	for _, a := range agents {
		if a.Status == "active" {
			resp.Agents.Active++
		}
		resp.Agents.ByAdapter[a.Adapter]++

		detail := DashboardAgentDetail{
			ID:      a.ID,
			Name:    a.Name,
			Adapter: a.Adapter,
			Model:   a.Model,
			Status:  a.Status,
		}

		// 查询该 Agent 的任务统计
		taskList, terr := h.taskMgr.ListTasks(&task.ListFilter{
			AgentID: a.ID,
		})
		if terr == nil {
			for _, t := range taskList {
				switch t.Status {
				case task.StatusRunning:
					detail.Running++
				case task.StatusQueued, task.StatusPending:
					detail.Queued++
				case task.StatusCompleted:
					detail.Completed++
				case task.StatusFailed:
					detail.Failed++
				}
			}
		}

		resp.Agents.Details = append(resp.Agents.Details, detail)
	}

	// ==================== Task 统计 ====================
	taskStats, err := h.taskMgr.GetStats()
	if err == nil {
		resp.Tasks.Total = taskStats.Total
		resp.Tasks.AvgDuration = taskStats.AvgDuration
		resp.Tasks.ByStatus = make(map[string]int)
		for status, count := range taskStats.ByStatus {
			resp.Tasks.ByStatus[string(status)] = count
		}

		completed := taskStats.ByStatus[task.StatusCompleted]
		failed := taskStats.ByStatus[task.StatusFailed]
		total := completed + failed
		if total > 0 {
			resp.Tasks.SuccessRate = float64(completed) / float64(total) * 100
		}
	}

	// 统计今日任务数
	todayStart := time.Now().Truncate(24 * time.Hour)
	allTasks, err := h.taskMgr.ListTasks(&task.ListFilter{})
	if err == nil {
		for _, t := range allTasks {
			if t.CreatedAt.After(todayStart) {
				resp.Tasks.Today++
			}
		}
	}

	// ==================== Session 统计 ====================
	sessions, err := h.sessionMgr.List(ctx, nil)
	if err == nil {
		resp.Sessions.Total = len(sessions)
		for _, s := range sessions {
			switch s.Status {
			case session.StatusRunning:
				resp.Sessions.Running++
			case session.StatusCreating:
				resp.Sessions.Creating++
			case session.StatusStopped:
				resp.Sessions.Stopped++
			case session.StatusError:
				resp.Sessions.Error++
			}
		}
	}

	// ==================== Token 统计 ====================
	historyStats, err := h.historyMgr.GetStats(nil)
	if err == nil {
		resp.Tokens.TotalInput = int64(historyStats.TotalInputTokens)
		resp.Tokens.TotalOutput = int64(historyStats.TotalOutputTokens)
		resp.Tokens.TotalTokens = int64(historyStats.TotalInputTokens + historyStats.TotalOutputTokens)
	}

	// ==================== Container 统计 ====================
	containers, err := h.containerMgr.ListContainers(ctx)
	if err == nil {
		resp.Containers.Total = len(containers)
		for _, ctr := range containers {
			switch ctr.Status {
			case container.StatusRunning:
				resp.Containers.Running++
			default:
				resp.Containers.Stopped++
			}
		}
	}

	// ==================== Provider 状态 ====================
	providers := h.providerMgr.List()
	for _, p := range providers {
		status := "offline"
		if p.IsConfigured && p.IsValid {
			status = "online"
		} else if p.IsConfigured {
			status = "degraded"
		}
		resp.Providers = append(resp.Providers, DashboardProviderInfo{
			ID:           p.ID,
			Name:         p.Name,
			Status:       status,
			IsConfigured: p.IsConfigured,
			IsValid:      p.IsValid,
			Category:     string(p.Category),
			Icon:         p.Icon,
			IconColor:    p.IconColor,
		})
	}

	// ==================== 系统信息 ====================
	resp.System.Uptime = time.Since(h.startTime).Round(time.Second).String()
	resp.System.StartedAt = h.startTime.Format(time.RFC3339)

	// ==================== 最近任务 ====================
	recentTasks, err := h.taskMgr.ListTasks(&task.ListFilter{
		Limit:     20,
		OrderBy:   "created_at",
		OrderDesc: true,
	})
	if err == nil {
		for _, t := range recentTasks {
			prompt := t.Prompt
			if len(prompt) > 80 {
				prompt = prompt[:80] + "..."
			}

			var duration float64
			if t.StartedAt != nil && t.CompletedAt != nil {
				duration = t.CompletedAt.Sub(*t.StartedAt).Seconds()
			}

			resp.RecentTasks = append(resp.RecentTasks, DashboardRecentTask{
				ID:        t.ID,
				AgentID:   t.AgentID,
				AgentName: t.AgentName,
				Adapter:   t.AgentType,
				Prompt:    prompt,
				Status:    string(t.Status),
				Duration:  duration,
				CreatedAt: t.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	Success(c, resp)
}
