// Package profile provides the Profile system for AgentBox.
// Profile is the core differentiating feature that allows users to save and reuse
// pre-configured Agent settings (adapter + model + MCP + permissions).
package profile

import (
	"encoding/json"
	"time"
)

// Profile is a user-configurable Agent runtime template
type Profile struct {
	// Basic info
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Adapter selection: "claude-code" | "codex"
	Adapter string `json:"adapter"`

	// Inheritance: extends another Profile
	Extends string `json:"extends,omitempty"`

	// Model configuration
	Model ModelConfig `json:"model"`

	// MCP servers
	MCPServers []MCPServerConfig `json:"mcp_servers,omitempty"`

	// Permission configuration
	Permissions PermissionConfig `json:"permissions"`

	// Resource limits
	Resources ResourceConfig `json:"resources"`

	// System prompts
	SystemPrompt          string `json:"system_prompt,omitempty"`
	AppendSystemPrompt    string `json:"append_system_prompt,omitempty"`
	BaseInstructions      string `json:"base_instructions,omitempty"`
	DeveloperInstructions string `json:"developer_instructions,omitempty"`

	// Feature flags
	Features FeatureConfig `json:"features"`

	// Custom agents (Claude Code specific)
	CustomAgents json.RawMessage `json:"custom_agents,omitempty"`

	// Config overrides (Codex specific)
	ConfigOverrides map[string]string `json:"config_overrides,omitempty"`

	// Output configuration
	OutputFormat string `json:"output_format,omitempty"` // text/json/stream-json
	OutputSchema string `json:"output_schema,omitempty"` // JSON Schema file path

	// Debug settings
	Debug DebugConfig `json:"debug,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
	IsBuiltIn bool      `json:"is_built_in"`
	IsPublic  bool      `json:"is_public"`
}

// ModelConfig defines model settings
type ModelConfig struct {
	// Basic configuration
	Name            string `json:"name"`                       // e.g., "sonnet", "opus", "o3"
	Provider        string `json:"provider,omitempty"`         // e.g., "anthropic", "openai", "glm"
	BaseURL         string `json:"base_url,omitempty"`         // Custom API endpoint (proxy/compatible API)
	ReasoningEffort string `json:"reasoning_effort,omitempty"` // low/medium/high (Codex)

	// Model tier configuration (Claude Code)
	HaikuModel  string `json:"haiku_model,omitempty"`  // ANTHROPIC_DEFAULT_HAIKU_MODEL
	SonnetModel string `json:"sonnet_model,omitempty"` // ANTHROPIC_DEFAULT_SONNET_MODEL
	OpusModel   string `json:"opus_model,omitempty"`   // ANTHROPIC_DEFAULT_OPUS_MODEL

	// Advanced configuration
	TimeoutMS       int  `json:"timeout_ms,omitempty"`        // API_TIMEOUT_MS
	MaxOutputTokens int  `json:"max_output_tokens,omitempty"` // CLAUDE_CODE_MAX_OUTPUT_TOKENS
	DisableTraffic  bool `json:"disable_traffic,omitempty"`   // CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC

	// Codex specific - third-party provider support
	WireAPI     string `json:"wire_api,omitempty"`     // chat/responses (Codex config.toml)
	EnvKey      string `json:"env_key,omitempty"`      // Environment variable name for API key (e.g., "ZHIPU_API_KEY")
	BearerToken string `json:"bearer_token,omitempty"` // experimental_bearer_token for Codex
}

// MCPServerConfig defines an MCP server
type MCPServerConfig struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
}

// PermissionConfig defines permission settings
type PermissionConfig struct {
	// Claude Code specific
	Mode            string   `json:"mode,omitempty"`             // permission-mode
	AllowedTools    []string `json:"allowed_tools,omitempty"`    // --allowedTools
	DisallowedTools []string `json:"disallowed_tools,omitempty"` // --disallowedTools
	Tools           []string `json:"tools,omitempty"`            // --tools
	SkipAll         bool     `json:"skip_all,omitempty"`         // --dangerously-skip-permissions

	// Codex specific
	SandboxMode    string `json:"sandbox_mode,omitempty"`    // read-only/workspace-write/danger-full-access
	ApprovalPolicy string `json:"approval_policy,omitempty"` // untrusted/on-failure/on-request/never
	FullAuto       bool   `json:"full_auto,omitempty"`       // --full-auto

	// Common
	AdditionalDirs []string `json:"additional_dirs,omitempty"` // --add-dir
}

// ResourceConfig defines resource limits
type ResourceConfig struct {
	MaxBudgetUSD float64       `json:"max_budget_usd,omitempty"`
	MaxTurns     int           `json:"max_turns,omitempty"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`

	// Container resources
	CPUs     float64 `json:"cpus,omitempty"`
	MemoryMB int     `json:"memory_mb,omitempty"`
	DiskGB   int     `json:"disk_gb,omitempty"`
}

// FeatureConfig defines feature flags
type FeatureConfig struct {
	WebSearch bool `json:"web_search,omitempty"`
}

// DebugConfig defines debug settings
type DebugConfig struct {
	Verbose bool `json:"verbose,omitempty"`
}

// PermissionMode constants for Claude Code
const (
	PermissionModeAcceptEdits       = "acceptEdits"
	PermissionModeBypassPermissions = "bypassPermissions"
	PermissionModeDefault           = "default"
	PermissionModeDelegate          = "delegate"
	PermissionModeDontAsk           = "dontAsk"
	PermissionModePlan              = "plan"
)

// SandboxMode constants for Codex
const (
	SandboxModeReadOnly         = "read-only"
	SandboxModeWorkspaceWrite   = "workspace-write"
	SandboxModeDangerFullAccess = "danger-full-access"
)

// ApprovalPolicy constants for Codex
const (
	ApprovalPolicyUntrusted = "untrusted"
	ApprovalPolicyOnFailure = "on-failure"
	ApprovalPolicyOnRequest = "on-request"
	ApprovalPolicyNever     = "never"
)

// AdapterType constants
const (
	AdapterClaudeCode = "claude-code"
	AdapterCodex      = "codex"
	AdapterOpenCode   = "opencode"
)

// Validate validates the Profile
func (p *Profile) Validate() error {
	if p.ID == "" {
		return ErrProfileIDRequired
	}
	if p.Name == "" {
		return ErrProfileNameRequired
	}
	if p.Adapter == "" {
		return ErrProfileAdapterRequired
	}
	if p.Adapter != AdapterClaudeCode && p.Adapter != AdapterCodex && p.Adapter != AdapterOpenCode {
		return ErrProfileInvalidAdapter
	}
	return nil
}

// Clone creates a deep copy of the Profile with a new ID
func (p *Profile) Clone(newID, newName string) *Profile {
	clone := *p
	clone.ID = newID
	clone.Name = newName
	clone.IsBuiltIn = false
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()
	clone.Extends = ""

	// Deep copy slices and maps
	if p.Tags != nil {
		clone.Tags = make([]string, len(p.Tags))
		copy(clone.Tags, p.Tags)
	}
	if p.MCPServers != nil {
		clone.MCPServers = make([]MCPServerConfig, len(p.MCPServers))
		copy(clone.MCPServers, p.MCPServers)
	}
	if p.Permissions.AllowedTools != nil {
		clone.Permissions.AllowedTools = make([]string, len(p.Permissions.AllowedTools))
		copy(clone.Permissions.AllowedTools, p.Permissions.AllowedTools)
	}
	if p.Permissions.DisallowedTools != nil {
		clone.Permissions.DisallowedTools = make([]string, len(p.Permissions.DisallowedTools))
		copy(clone.Permissions.DisallowedTools, p.Permissions.DisallowedTools)
	}
	if p.Permissions.AdditionalDirs != nil {
		clone.Permissions.AdditionalDirs = make([]string, len(p.Permissions.AdditionalDirs))
		copy(clone.Permissions.AdditionalDirs, p.Permissions.AdditionalDirs)
	}
	if p.ConfigOverrides != nil {
		clone.ConfigOverrides = make(map[string]string)
		for k, v := range p.ConfigOverrides {
			clone.ConfigOverrides[k] = v
		}
	}

	return &clone
}

// Merge merges another Profile into this one (for inheritance)
// The current Profile's values take precedence over the parent's
func (p *Profile) Merge(parent *Profile) {
	// Model configuration
	if p.Model.Name == "" {
		p.Model.Name = parent.Model.Name
	}
	if p.Model.Provider == "" {
		p.Model.Provider = parent.Model.Provider
	}
	if p.Model.BaseURL == "" {
		p.Model.BaseURL = parent.Model.BaseURL
	}
	if p.Model.ReasoningEffort == "" {
		p.Model.ReasoningEffort = parent.Model.ReasoningEffort
	}
	if p.Model.HaikuModel == "" {
		p.Model.HaikuModel = parent.Model.HaikuModel
	}
	if p.Model.SonnetModel == "" {
		p.Model.SonnetModel = parent.Model.SonnetModel
	}
	if p.Model.OpusModel == "" {
		p.Model.OpusModel = parent.Model.OpusModel
	}
	if p.Model.TimeoutMS == 0 {
		p.Model.TimeoutMS = parent.Model.TimeoutMS
	}
	if p.Model.MaxOutputTokens == 0 {
		p.Model.MaxOutputTokens = parent.Model.MaxOutputTokens
	}
	if !p.Model.DisableTraffic {
		p.Model.DisableTraffic = parent.Model.DisableTraffic
	}
	if p.Model.WireAPI == "" {
		p.Model.WireAPI = parent.Model.WireAPI
	}
	if p.Model.EnvKey == "" {
		p.Model.EnvKey = parent.Model.EnvKey
	}
	if p.Model.BearerToken == "" {
		p.Model.BearerToken = parent.Model.BearerToken
	}
	if p.SystemPrompt == "" {
		p.SystemPrompt = parent.SystemPrompt
	}
	if p.AppendSystemPrompt == "" {
		p.AppendSystemPrompt = parent.AppendSystemPrompt
	}
	if p.BaseInstructions == "" {
		p.BaseInstructions = parent.BaseInstructions
	}
	if p.DeveloperInstructions == "" {
		p.DeveloperInstructions = parent.DeveloperInstructions
	}
	if p.OutputFormat == "" {
		p.OutputFormat = parent.OutputFormat
	}
	if p.OutputSchema == "" {
		p.OutputSchema = parent.OutputSchema
	}

	// Merge MCP servers (append parent's if not overridden)
	if len(p.MCPServers) == 0 {
		p.MCPServers = parent.MCPServers
	}

	// Merge permissions
	if p.Permissions.Mode == "" {
		p.Permissions.Mode = parent.Permissions.Mode
	}
	if p.Permissions.SandboxMode == "" {
		p.Permissions.SandboxMode = parent.Permissions.SandboxMode
	}
	if p.Permissions.ApprovalPolicy == "" {
		p.Permissions.ApprovalPolicy = parent.Permissions.ApprovalPolicy
	}
	if len(p.Permissions.AllowedTools) == 0 {
		p.Permissions.AllowedTools = parent.Permissions.AllowedTools
	}
	if len(p.Permissions.DisallowedTools) == 0 {
		p.Permissions.DisallowedTools = parent.Permissions.DisallowedTools
	}
	if len(p.Permissions.AdditionalDirs) == 0 {
		p.Permissions.AdditionalDirs = parent.Permissions.AdditionalDirs
	}

	// Merge resources (use parent's if zero)
	if p.Resources.MaxBudgetUSD == 0 {
		p.Resources.MaxBudgetUSD = parent.Resources.MaxBudgetUSD
	}
	if p.Resources.MaxTurns == 0 {
		p.Resources.MaxTurns = parent.Resources.MaxTurns
	}
	if p.Resources.MaxTokens == 0 {
		p.Resources.MaxTokens = parent.Resources.MaxTokens
	}
	if p.Resources.Timeout == 0 {
		p.Resources.Timeout = parent.Resources.Timeout
	}
	if p.Resources.CPUs == 0 {
		p.Resources.CPUs = parent.Resources.CPUs
	}
	if p.Resources.MemoryMB == 0 {
		p.Resources.MemoryMB = parent.Resources.MemoryMB
	}
	if p.Resources.DiskGB == 0 {
		p.Resources.DiskGB = parent.Resources.DiskGB
	}

	// Merge features
	if !p.Features.WebSearch {
		p.Features.WebSearch = parent.Features.WebSearch
	}

	// Merge config overrides
	if p.ConfigOverrides == nil && parent.ConfigOverrides != nil {
		p.ConfigOverrides = make(map[string]string)
		for k, v := range parent.ConfigOverrides {
			p.ConfigOverrides[k] = v
		}
	}
}
