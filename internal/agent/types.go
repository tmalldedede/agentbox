// Package agent provides the Agent system for AgentBox.
// Agent is the unified concept combining Provider + Runtime + Skills/MCP + Prompt,
// replacing the old Profile + SmartAgent split architecture.
package agent

import (
	"encoding/json"
	"time"
)

// Agent represents a complete AI agent configuration.
// Agent = Provider + Runtime + Skills + MCP + Prompt + Permissions
type Agent struct {
	// Basic info
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`

	// Engine adapter: "claude-code" | "codex" | "opencode"
	Adapter string `json:"adapter"`

	// References
	ProviderID string `json:"provider_id"` // → Provider (API service + key)
	RuntimeID  string `json:"runtime_id"`  // → AgentRuntime (image + resources)

	// Model selection (from Provider's available models)
	Model string `json:"model"`

	// Override Provider's base_url (useful when same provider has different endpoints per adapter)
	BaseURLOverride string `json:"base_url_override,omitempty"`

	// Advanced model configuration
	ModelConfig ModelConfig `json:"model_config,omitempty"`

	// Capabilities
	SkillIDs     []string `json:"skill_ids,omitempty"`
	MCPServerIDs []string `json:"mcp_server_ids,omitempty"`

	// Behavior
	SystemPrompt       string `json:"system_prompt,omitempty"`
	AppendSystemPrompt string `json:"append_system_prompt,omitempty"`

	// Permissions
	Permissions PermissionConfig `json:"permissions"`

	// Default workspace path (relative to workspaceBase, or absolute)
	// If empty, auto-generated as "agent-{id}-{task-id}" per task
	Workspace string `json:"workspace,omitempty"`

	// Environment variables (injected into container)
	Env map[string]string `json:"env,omitempty"`

	// API access settings
	APIAccess  string `json:"api_access"`            // public, api_key, private
	RateLimit  int    `json:"rate_limit,omitempty"`   // requests per minute
	WebhookURL string `json:"webhook_url,omitempty"` // callback URL for async results

	// Output configuration
	OutputFormat string `json:"output_format,omitempty"` // text/json/stream-json
	OutputSchema string `json:"output_schema,omitempty"` // JSON Schema

	// Feature flags
	Features FeatureConfig `json:"features,omitempty"`

	// Config overrides (Codex specific: config.toml fields)
	ConfigOverrides map[string]string `json:"config_overrides,omitempty"`

	// Custom agents config (Claude Code specific)
	CustomAgents json.RawMessage `json:"custom_agents,omitempty"`

	// Status
	Status string `json:"status"` // active, inactive

	// Metadata
	IsBuiltIn bool      `json:"is_built_in"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
}

// ModelConfig defines advanced model settings
type ModelConfig struct {
	// Reasoning effort (Codex): low/medium/high
	ReasoningEffort string `json:"reasoning_effort,omitempty"`

	// Model tier (Claude Code)
	HaikuModel  string `json:"haiku_model,omitempty"`
	SonnetModel string `json:"sonnet_model,omitempty"`
	OpusModel   string `json:"opus_model,omitempty"`

	// API settings
	TimeoutMS       int  `json:"timeout_ms,omitempty"`
	MaxOutputTokens int  `json:"max_output_tokens,omitempty"`
	DisableTraffic  bool `json:"disable_traffic,omitempty"`

	// Codex specific
	WireAPI string `json:"wire_api,omitempty"` // chat/responses
}

// PermissionConfig defines permission settings
type PermissionConfig struct {
	// Claude Code specific
	Mode            string   `json:"mode,omitempty"`              // permission-mode
	AllowedTools    []string `json:"allowed_tools,omitempty"`     // --allowedTools
	DisallowedTools []string `json:"disallowed_tools,omitempty"`  // --disallowedTools
	Tools           []string `json:"tools,omitempty"`             // --tools
	SkipAll         bool     `json:"skip_all,omitempty"`          // --dangerously-skip-permissions

	// Codex specific
	SandboxMode    string `json:"sandbox_mode,omitempty"`    // read-only/workspace-write/danger-full-access
	ApprovalPolicy string `json:"approval_policy,omitempty"` // untrusted/on-failure/on-request/never
	FullAuto       bool   `json:"full_auto,omitempty"`       // --full-auto

	// Common
	AdditionalDirs []string `json:"additional_dirs,omitempty"` // --add-dir
}

// FeatureConfig defines feature flags
type FeatureConfig struct {
	WebSearch bool `json:"web_search,omitempty"`
}

// ResourceConfig defines resource budget limits (separate from Runtime hardware)
type ResourceConfig struct {
	MaxBudgetUSD float64       `json:"max_budget_usd,omitempty"`
	MaxTurns     int           `json:"max_turns,omitempty"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
}

// APIAccess constants
const (
	APIAccessPublic  = "public"
	APIAccessAPIKey  = "api_key"
	APIAccessPrivate = "private"
)

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Adapter constants
const (
	AdapterClaudeCode = "claude-code"
	AdapterCodex      = "codex"
	AdapterOpenCode   = "opencode"
)

// Validate validates the Agent
func (a *Agent) Validate() error {
	if a.ID == "" {
		return ErrAgentIDRequired
	}
	if a.Name == "" {
		return ErrAgentNameRequired
	}
	if a.Adapter == "" {
		return ErrAgentAdapterRequired
	}
	if a.Adapter != AdapterClaudeCode && a.Adapter != AdapterCodex && a.Adapter != AdapterOpenCode {
		return ErrAgentInvalidAdapter
	}
	if a.ProviderID == "" {
		return ErrAgentProviderRequired
	}
	return nil
}

// RunRequest represents a request to run an agent
type RunRequest struct {
	Prompt   string            `json:"prompt"`
	Input    *RunInput         `json:"input,omitempty"`
	Options  *RunOptions       `json:"options,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RunInput represents input data for an agent run
type RunInput struct {
	Files []string `json:"files,omitempty"`
	Text  string   `json:"text,omitempty"`
}

// RunOptions represents options for an agent run
type RunOptions struct {
	MaxTurns int  `json:"max_turns,omitempty"`
	Timeout  int  `json:"timeout,omitempty"`
	Stream   bool `json:"stream,omitempty"`
	Async    bool `json:"async,omitempty"`
}

// RunResult represents the result of an agent run
type RunResult struct {
	RunID     string     `json:"run_id"`
	AgentID   string     `json:"agent_id"`
	AgentName string     `json:"agent_name"`
	Status    string     `json:"status"`
	Output    string     `json:"output,omitempty"`
	Error     string     `json:"error,omitempty"`
	Usage     *UsageInfo `json:"usage,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// UsageInfo represents resource usage information
type UsageInfo struct {
	InputTokens  int     `json:"input_tokens,omitempty"`
	OutputTokens int     `json:"output_tokens,omitempty"`
	DurationMs   int64   `json:"duration_ms,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}

// RunStatus constants
const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
)
