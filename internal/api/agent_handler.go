package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/session"
)

// AgentHandler handles agent API requests
type AgentHandler struct {
	manager    *agent.Manager
	sessionMgr *session.Manager
	historyMgr *history.Manager
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(manager *agent.Manager, sessionMgr *session.Manager, historyMgr *history.Manager) *AgentHandler {
	return &AgentHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		historyMgr: historyMgr,
	}
}

// RegisterRoutes registers agent routes
func (h *AgentHandler) RegisterRoutes(r *gin.RouterGroup) {
	agents := r.Group("/agents")
	{
		agents.GET("", h.List)
		agents.POST("", h.Create)
		agents.GET("/:id", h.Get)
		agents.PUT("/:id", h.Update)
		agents.DELETE("/:id", h.Delete)
		agents.POST("/:id/run", h.Run)
	}
}

// CreateAgentReq represents a request to create an agent
type CreateAgentReq struct {
	ID                 string            `json:"id" binding:"required"`
	Name               string            `json:"name" binding:"required"`
	Description        string            `json:"description"`
	Icon               string            `json:"icon"`
	Adapter            string            `json:"adapter" binding:"required"`
	ProviderID         string            `json:"provider_id" binding:"required"`
	RuntimeID          string            `json:"runtime_id"`
	Model              string            `json:"model"`
	BaseURLOverride    string            `json:"base_url_override"`
	ModelConfig        agent.ModelConfig  `json:"model_config"`
	SkillIDs           []string          `json:"skill_ids"`
	MCPServerIDs       []string          `json:"mcp_server_ids"`
	SystemPrompt       string            `json:"system_prompt"`
	AppendSystemPrompt string            `json:"append_system_prompt"`
	Permissions        agent.PermissionConfig `json:"permissions"`
	Env                map[string]string `json:"env"`
	APIAccess          string            `json:"api_access"`
	RateLimit          int               `json:"rate_limit"`
	WebhookURL         string            `json:"webhook_url"`
	OutputFormat       string            `json:"output_format"`
	Features           agent.FeatureConfig `json:"features"`
	ConfigOverrides    map[string]string `json:"config_overrides"`
}

// ListPublic 公开 API - 列出可用 Agent（只返回 active）
// GET /api/v1/agents
func (h *AgentHandler) ListPublic(c *gin.Context) {
	allAgents := h.manager.List()
	var activeAgents []*agent.Agent
	for _, ag := range allAgents {
		if ag.Status == agent.StatusActive {
			activeAgents = append(activeAgents, ag)
		}
	}
	Success(c, activeAgents)
}

// GetPublic 公开 API - 获取 Agent 详情
// GET /api/v1/agents/:id
func (h *AgentHandler) GetPublic(c *gin.Context) {
	id := c.Param("id")

	ag, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, ag)
}

// List godoc
// @Summary List all agents
// @Description Get a list of all agents (admin)
// @Tags Agents
// @Produce json
// @Success 200 {object} Response{data=[]agent.Agent}
// @Router /admin/agents [get]
func (h *AgentHandler) List(c *gin.Context) {
	agents := h.manager.List()
	Success(c, agents)
}

// Get godoc
// @Summary Get an agent
// @Description Get detailed information about a specific agent (admin)
// @Tags Agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} Response{data=agent.Agent}
// @Failure 404 {object} Response
// @Router /admin/agents/{id} [get]
func (h *AgentHandler) Get(c *gin.Context) {
	id := c.Param("id")

	ag, err := h.manager.Get(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	Success(c, ag)
}

// Create godoc
// @Summary Create an agent
// @Description Create a new agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param agent body CreateAgentReq true "Agent configuration"
// @Success 201 {object} Response{data=agent.Agent}
// @Failure 400 {object} Response
// @Router /agents [post]
func (h *AgentHandler) Create(c *gin.Context) {
	var req CreateAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ag := &agent.Agent{
		ID:                 req.ID,
		Name:               req.Name,
		Description:        req.Description,
		Icon:               req.Icon,
		Adapter:            req.Adapter,
		ProviderID:         req.ProviderID,
		RuntimeID:          req.RuntimeID,
		Model:              req.Model,
		BaseURLOverride:    req.BaseURLOverride,
		ModelConfig:        req.ModelConfig,
		SkillIDs:           req.SkillIDs,
		MCPServerIDs:       req.MCPServerIDs,
		SystemPrompt:       req.SystemPrompt,
		AppendSystemPrompt: req.AppendSystemPrompt,
		Permissions:        req.Permissions,
		Env:                req.Env,
		APIAccess:          req.APIAccess,
		RateLimit:          req.RateLimit,
		WebhookURL:         req.WebhookURL,
		OutputFormat:       req.OutputFormat,
		Features:           req.Features,
		ConfigOverrides:    req.ConfigOverrides,
	}

	if err := h.manager.Create(ag); err != nil {
		HandleError(c, err)
		return
	}

	Created(c, ag)
}

// Update godoc
// @Summary Update an agent
// @Description Update an existing agent
// @Tags Agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param agent body CreateAgentReq true "Agent configuration"
// @Success 200 {object} Response{data=agent.Agent}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id} [put]
func (h *AgentHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req CreateAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	ag := &agent.Agent{
		ID:                 id,
		Name:               req.Name,
		Description:        req.Description,
		Icon:               req.Icon,
		Adapter:            req.Adapter,
		ProviderID:         req.ProviderID,
		RuntimeID:          req.RuntimeID,
		Model:              req.Model,
		BaseURLOverride:    req.BaseURLOverride,
		ModelConfig:        req.ModelConfig,
		SkillIDs:           req.SkillIDs,
		MCPServerIDs:       req.MCPServerIDs,
		SystemPrompt:       req.SystemPrompt,
		AppendSystemPrompt: req.AppendSystemPrompt,
		Permissions:        req.Permissions,
		Env:                req.Env,
		APIAccess:          req.APIAccess,
		RateLimit:          req.RateLimit,
		WebhookURL:         req.WebhookURL,
		OutputFormat:       req.OutputFormat,
		Features:           req.Features,
		ConfigOverrides:    req.ConfigOverrides,
	}

	if err := h.manager.Update(ag); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, ag)
}

// Delete godoc
// @Summary Delete an agent
// @Description Delete an existing agent
// @Tags Agents
// @Param id path string true "Agent ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id} [delete]
func (h *AgentHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.manager.Delete(id); err != nil {
		HandleError(c, err)
		return
	}

	Success(c, map[string]interface{}{"id": id, "deleted": true})
}

// RunAgentReq represents a request to run an agent
type RunAgentReq struct {
	Prompt    string            `json:"prompt" binding:"required"`
	Workspace string            `json:"workspace"`
	Input     *agent.RunInput   `json:"input"`
	Options   *agent.RunOptions `json:"options"`
	Metadata  map[string]string `json:"metadata"`
}

// Run godoc
// @Summary Run an agent
// @Description Execute an agent with the given prompt
// @Tags Agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param request body RunAgentReq true "Run request"
// @Success 200 {object} Response{data=agent.RunResult}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /agents/{id}/run [post]
func (h *AgentHandler) Run(c *gin.Context) {
	id := c.Param("id")

	// Get full agent configuration (resolves all references)
	fullConfig, err := h.manager.GetFullConfig(id)
	if err != nil {
		HandleError(c, err)
		return
	}

	ag := fullConfig.Agent

	// Check agent status
	if ag.Status != agent.StatusActive {
		HandleError(c, agent.ErrAgentInactive)
		return
	}

	var req RunAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Generate workspace name if not provided
	workspace := req.Workspace
	if workspace == "" {
		workspace = fmt.Sprintf("agent-%s-%s", ag.ID, uuid.New().String()[:8])
	}

	// Build environment variables:
	// 1. Provider env vars (including decrypted API key)
	env := make(map[string]string)
	provEnv, err := h.manager.GetProviderEnvVars(id)
	if err == nil {
		for k, v := range provEnv {
			env[k] = v
		}
	}
	// 2. Agent's own env vars (override provider)
	for k, v := range ag.Env {
		env[k] = v
	}

	// Build session config from runtime
	var sessionConfig *session.Config
	if fullConfig.Runtime != nil {
		sessionConfig = &session.Config{
			CPULimit:    fullConfig.Runtime.CPUs,
			MemoryLimit: int64(fullConfig.Runtime.MemoryMB) * 1024 * 1024,
		}
	}

	// Create session request
	sessionReq := &session.CreateRequest{
		AgentID:   ag.ID,
		Agent:     ag.Adapter,
		Workspace: workspace,
		Env:       env,
		Config:    sessionConfig,
	}

	// Create session
	sess, err := h.sessionMgr.Create(c.Request.Context(), sessionReq)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Build the prompt with system prompt if present
	prompt := req.Prompt
	if ag.SystemPrompt != "" {
		prompt = ag.SystemPrompt + "\n\n" + req.Prompt
	}

	// Execute prompt
	execReq := &session.ExecRequest{
		Prompt: prompt,
	}
	if req.Options != nil {
		execReq.MaxTurns = req.Options.MaxTurns
		execReq.Timeout = req.Options.Timeout
	}

	// Record execution to history
	execID := uuid.New().String()
	if h.historyMgr != nil {
		_, _ = h.historyMgr.RecordAgent(
			execID,
			ag.ID,
			ag.Name,
			ag.Adapter,
			req.Prompt,
		)
	}

	startedAt := time.Now()
	result, err := h.sessionMgr.Exec(c.Request.Context(), sess.ID, execReq)
	if err != nil {
		if h.historyMgr != nil {
			_ = h.historyMgr.Fail(execID, err.Error(), 1)
		}
		HandleError(c, err)
		return
	}

	// Build response
	runResult := &agent.RunResult{
		RunID:     execID,
		AgentID:   ag.ID,
		AgentName: ag.Name,
		Status:    agent.RunStatusCompleted,
		Output:    result.Message,
		StartedAt: startedAt,
	}

	if result.Error != "" {
		runResult.Status = agent.RunStatusFailed
		runResult.Error = result.Error
		if h.historyMgr != nil {
			_ = h.historyMgr.Fail(execID, result.Error, result.ExitCode)
		}
	} else {
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
		runResult.Usage = &agent.UsageInfo{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		}
	}

	endedAt := time.Now()
	runResult.EndedAt = &endedAt

	Success(c, runResult)
}
