// Package smartagent provides the SmartAgent system for AgentBox.
// SmartAgent is the core product concept - an AI agent with specific capabilities
// that exposes an API endpoint for external invocation.
package smartagent

import (
	"time"
)

// Agent represents a smart agent that can be invoked via API.
// Agent = Profile + System Prompt + Environment Variables
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`

	// Profile reference - determines engine, model, MCP, skills, etc.
	ProfileID string `json:"profile_id"`

	// Agent-specific configuration
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Env          map[string]string `json:"env,omitempty"`

	// API settings
	APIAccess   string `json:"api_access"`            // public, api_key, private
	RateLimit   int    `json:"rate_limit,omitempty"`  // requests per minute
	WebhookURL  string `json:"webhook_url,omitempty"` // callback URL for async results

	// Status
	Status string `json:"status"` // active, inactive, error

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
}

// APIAccess constants
const (
	APIAccessPublic  = "public"   // No authentication required
	APIAccessAPIKey  = "api_key"  // Requires API key
	APIAccessPrivate = "private"  // Internal use only
)

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusError    = "error"
)

// Validate validates the Agent
func (a *Agent) Validate() error {
	if a.ID == "" {
		return ErrAgentIDRequired
	}
	if a.Name == "" {
		return ErrAgentNameRequired
	}
	if a.ProfileID == "" {
		return ErrAgentProfileRequired
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
	Files []string `json:"files,omitempty"` // File paths or URLs
	Text  string   `json:"text,omitempty"`  // Additional text context
}

// RunOptions represents options for an agent run
type RunOptions struct {
	MaxTurns int  `json:"max_turns,omitempty"`
	Timeout  int  `json:"timeout,omitempty"` // seconds
	Stream   bool `json:"stream,omitempty"`  // Enable SSE streaming
	Async    bool `json:"async,omitempty"`   // Run asynchronously (returns run ID)
}

// RunResult represents the result of an agent run
type RunResult struct {
	RunID     string     `json:"run_id"`
	AgentID   string     `json:"agent_id"`
	AgentName string     `json:"agent_name"`
	Status    string     `json:"status"` // running, completed, failed
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
