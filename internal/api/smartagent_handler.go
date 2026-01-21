package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/smartagent"
)

// SmartAgentHandler handles smart agent API requests
type SmartAgentHandler struct {
	manager    *smartagent.Manager
	sessionMgr *session.Manager
	profileMgr *profile.Manager
	historyMgr *history.Manager
}

// NewSmartAgentHandler creates a new smart agent handler
func NewSmartAgentHandler(manager *smartagent.Manager, sessionMgr *session.Manager, profileMgr *profile.Manager, historyMgr *history.Manager) *SmartAgentHandler {
	return &SmartAgentHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		profileMgr: profileMgr,
		historyMgr: historyMgr,
	}
}

// RegisterRoutes registers smart agent routes
func (h *SmartAgentHandler) RegisterRoutes(r *gin.RouterGroup) {
	agents := r.Group("/agents")
	{
		agents.GET("", h.List)
		agents.POST("", h.Create)
		agents.GET("/:id", h.Get)
		agents.PUT("/:id", h.Update)
		agents.DELETE("/:id", h.Delete)
		agents.POST("/:id/run", h.Run)
		// agents.GET("/:id/runs", h.ListRuns)  // TODO: implement
	}
}

// CreateAgentRequest represents a request to create an agent
type CreateAgentRequest struct {
	ID           string            `json:"id" binding:"required"`
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Icon         string            `json:"icon"`
	ProfileID    string            `json:"profile_id" binding:"required"`
	SystemPrompt string            `json:"system_prompt"`
	Env          map[string]string `json:"env"`
	APIAccess    string            `json:"api_access"`
	RateLimit    int               `json:"rate_limit"`
	WebhookURL   string            `json:"webhook_url"`
}

// List godoc
// @Summary List all agents
// @Description Get a list of all smart agents
// @Tags Agents
// @Produce json
// @Success 200 {object} Response{data=[]smartagent.Agent}
// @Router /agents [get]
func (h *SmartAgentHandler) List(c *gin.Context) {
	agents := h.manager.List()
	Success(c, agents)
}

// Get godoc
// @Summary Get an agent
// @Description Get detailed information about a specific agent
// @Tags Agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} Response{data=smartagent.Agent}
// @Failure 404 {object} Response
// @Router /agents/{id} [get]
func (h *SmartAgentHandler) Get(c *gin.Context) {
	id := c.Param("id")

	agent, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, agent)
}

// Create godoc
// @Summary Create an agent
// @Description Create a new smart agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param agent body CreateAgentRequest true "Agent configuration"
// @Success 201 {object} Response{data=smartagent.Agent}
// @Failure 400 {object} Response
// @Router /agents [post]
func (h *SmartAgentHandler) Create(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	agent := &smartagent.Agent{
		ID:           req.ID,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		ProfileID:    req.ProfileID,
		SystemPrompt: req.SystemPrompt,
		Env:          req.Env,
		APIAccess:    req.APIAccess,
		RateLimit:    req.RateLimit,
		WebhookURL:   req.WebhookURL,
		Status:       smartagent.StatusActive,
	}

	if err := h.manager.Create(agent); err != nil {
		HandleError(c, err)
		return
	}

	Created(c, agent)
}

// Update godoc
// @Summary Update an agent
// @Description Update an existing agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param agent body CreateAgentRequest true "Agent configuration"
// @Success 200 {object} Response{data=smartagent.Agent}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id} [put]
func (h *SmartAgentHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	agent := &smartagent.Agent{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		ProfileID:    req.ProfileID,
		SystemPrompt: req.SystemPrompt,
		Env:          req.Env,
		APIAccess:    req.APIAccess,
		RateLimit:    req.RateLimit,
		WebhookURL:   req.WebhookURL,
	}

	if err := h.manager.Update(agent); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, agent)
}

// Delete godoc
// @Summary Delete an agent
// @Description Delete an existing agent
// @Tags Agents
// @Param id path string true "Agent ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id} [delete]
func (h *SmartAgentHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, map[string]interface{}{"id": id, "deleted": true})
}

// RunAgentRequest represents a request to run an agent
type RunAgentRequest struct {
	Prompt    string                  `json:"prompt" binding:"required"`
	Workspace string                  `json:"workspace"` // Optional workspace path
	Input     *smartagent.RunInput    `json:"input"`
	Options   *smartagent.RunOptions  `json:"options"`
	Metadata  map[string]string       `json:"metadata"`
}

// Run godoc
// @Summary Run an agent
// @Description Execute an agent with the given prompt
// @Tags Agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param request body RunAgentRequest true "Run request"
// @Success 200 {object} Response{data=smartagent.RunResult}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id}/run [post]
func (h *SmartAgentHandler) Run(c *gin.Context) {
	id := c.Param("id")

	// Get agent
	agent, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Check agent status
	if agent.Status != smartagent.StatusActive {
		HandleError(c, smartagent.ErrAgentInactive)
		return
	}

	var req RunAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Get profile to determine the engine (adapter)
	prof, err := h.profileMgr.Get(agent.ProfileID)
	if err != nil {
		HandleError(c, smartagent.ErrProfileNotFound)
		return
	}

	// Generate workspace name if not provided
	workspace := req.Workspace
	if workspace == "" {
		workspace = fmt.Sprintf("agent-%s-%s", agent.ID, uuid.New().String()[:8])
	}

	// Merge agent env with request metadata
	env := make(map[string]string)
	for k, v := range agent.Env {
		env[k] = v
	}

	// Create session request
	sessionReq := &session.CreateRequest{
		Agent:     prof.Adapter, // Engine name from profile
		ProfileID: agent.ProfileID,
		Workspace: workspace,
		Env:       env,
	}

	// Create session
	sess, err := h.sessionMgr.Create(c.Request.Context(), sessionReq)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Build the prompt with system prompt if present
	prompt := req.Prompt
	if agent.SystemPrompt != "" {
		prompt = agent.SystemPrompt + "\n\n" + req.Prompt
	}

	// Execute prompt
	execReq := &session.ExecRequest{
		Prompt: prompt,
	}
	if req.Options != nil {
		execReq.MaxTurns = req.Options.MaxTurns
		execReq.Timeout = req.Options.Timeout
	}

	// Record execution to history (before execution)
	execID := uuid.New().String()
	if h.historyMgr != nil {
		_, _ = h.historyMgr.RecordAgent(
			execID,
			agent.ID,
			agent.Name,
			agent.ProfileID,
			prof.Name,
			prof.Adapter,
			req.Prompt, // Original prompt without system prompt
		)
	}

	startedAt := time.Now()
	result, err := h.sessionMgr.Exec(c.Request.Context(), sess.ID, execReq)
	if err != nil {
		// Record failure
		if h.historyMgr != nil {
			_ = h.historyMgr.Fail(execID, err.Error(), 1)
		}
		HandleError(c, err)
		return
	}

	// Build response
	runResult := &smartagent.RunResult{
		RunID:     execID,
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Status:    smartagent.RunStatusCompleted,
		Output:    result.Message,
		StartedAt: startedAt,
	}

	if result.Error != "" {
		runResult.Status = smartagent.RunStatusFailed
		runResult.Error = result.Error
		// Record failure
		if h.historyMgr != nil {
			_ = h.historyMgr.Fail(execID, result.Error, result.ExitCode)
		}
	} else {
		// Record success
		if h.historyMgr != nil {
			var usage *history.UsageInfo
			if result.Usage != nil {
				usage = &history.UsageInfo{
					InputTokens:  result.Usage.InputTokens,
					OutputTokens: result.Usage.OutputTokens,
				}
			}
			_ = h.historyMgr.Complete(execID, result.Message, usage)
		}
	}

	if result.Usage != nil {
		runResult.Usage = &smartagent.UsageInfo{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		}
	}

	endedAt := time.Now()
	runResult.EndedAt = &endedAt

	Success(c, runResult)
}
