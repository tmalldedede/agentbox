package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProfileModel represents a profile in the database
type ProfileModel struct {
	BaseModel
	Name                  string `gorm:"size:255;not null" json:"name"`
	Description           string `gorm:"size:1024" json:"description"`
	Icon                  string `gorm:"size:255" json:"icon"`
	Tags                  string `gorm:"type:text" json:"tags"` // JSON array
	Adapter               string `gorm:"size:64;not null" json:"adapter"`
	Extends               string `gorm:"size:64" json:"extends"`
	CredentialID          string `gorm:"size:64" json:"credential_id"`
	ModelConfig           string `gorm:"type:text" json:"model_config"`           // JSON
	MCPServers            string `gorm:"type:text" json:"mcp_servers"`             // JSON array
	SkillIDs              string `gorm:"type:text" json:"skill_ids"`               // JSON array
	Permissions           string `gorm:"type:text" json:"permissions"`             // JSON
	Resources             string `gorm:"type:text" json:"resources"`               // JSON
	Features              string `gorm:"type:text" json:"features"`                // JSON
	SystemPrompt          string `gorm:"type:text" json:"system_prompt"`
	AppendSystemPrompt    string `gorm:"type:text" json:"append_system_prompt"`
	BaseInstructions      string `gorm:"type:text" json:"base_instructions"`
	DeveloperInstructions string `gorm:"type:text" json:"developer_instructions"`
	CustomAgents          string `gorm:"type:text" json:"custom_agents"`    // JSON
	ConfigOverrides       string `gorm:"type:text" json:"config_overrides"` // JSON
	OutputFormat          string `gorm:"size:64" json:"output_format"`
	OutputSchema          string `gorm:"type:text" json:"output_schema"`
	Debug                 string `gorm:"type:text" json:"debug"` // JSON
	IsBuiltIn             bool   `gorm:"default:false" json:"is_built_in"`
	IsPublic              bool   `gorm:"default:false" json:"is_public"`
	CreatedBy             string `gorm:"size:64" json:"created_by"`
}

func (ProfileModel) TableName() string {
	return "profiles"
}

// MCPServerModel represents an MCP server in the database
type MCPServerModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"size:1024" json:"description"`
	Command     string `gorm:"size:512;not null" json:"command"`
	Args        string `gorm:"type:text" json:"args"`     // JSON array
	Env         string `gorm:"type:text" json:"env"`      // JSON object
	Metadata    string `gorm:"type:text" json:"metadata"` // JSON object
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`
}

func (MCPServerModel) TableName() string {
	return "mcp_servers"
}

// SkillModel represents a skill in the database
type SkillModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Slug        string `gorm:"size:255;uniqueIndex" json:"slug"`
	Description string `gorm:"size:1024" json:"description"`
	Version     string `gorm:"size:64" json:"version"`
	Author      string `gorm:"size:255" json:"author"`
	Tags        string `gorm:"type:text" json:"tags"`       // JSON array
	Triggers    string `gorm:"type:text" json:"triggers"`   // JSON array
	Content     string `gorm:"type:text" json:"content"`    // SKILL.md content
	Files       string `gorm:"type:text" json:"files"`      // JSON array of files
	Scripts     string `gorm:"type:text" json:"scripts"`    // JSON array
	References  string `gorm:"type:text" json:"references"` // JSON array
	Metadata    string `gorm:"type:text" json:"metadata"`   // JSON object
	IsBuiltIn   bool   `gorm:"default:false" json:"is_built_in"`
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"`
}

func (SkillModel) TableName() string {
	return "skills"
}

// CredentialModel represents a credential in the database
type CredentialModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Type        string `gorm:"size:64;not null" json:"type"` // api_key, oauth, etc.
	Provider    string `gorm:"size:64" json:"provider"`      // anthropic, openai, etc.
	Description string `gorm:"size:1024" json:"description"`
	Value       string `gorm:"type:text" json:"value"`       // Encrypted
	Metadata    string `gorm:"type:text" json:"metadata"`    // JSON object
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

func (CredentialModel) TableName() string {
	return "credentials"
}

// SessionModel represents a session in the database
type SessionModel struct {
	BaseModel
	ProfileID   string     `gorm:"size:64;index" json:"profile_id"`
	Agent       string     `gorm:"size:64;not null" json:"agent"`
	Status      string     `gorm:"size:32;not null;index" json:"status"`
	ContainerID string     `gorm:"size:128" json:"container_id"`
	Workspace   string     `gorm:"size:512" json:"workspace"`
	Config      string     `gorm:"type:text" json:"config"` // JSON
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

func (SessionModel) TableName() string {
	return "sessions"
}

// TaskModel represents a task in the database
type TaskModel struct {
	BaseModel
	SessionID   string     `gorm:"size:64;index" json:"session_id"`
	ProfileID   string     `gorm:"size:64;index" json:"profile_id"`
	Status      string     `gorm:"size:32;not null;index" json:"status"`
	Prompt      string     `gorm:"type:text" json:"prompt"`
	Input       string     `gorm:"type:text" json:"input"`    // JSON
	Output      string     `gorm:"type:text" json:"output"`   // JSON
	Metadata    string     `gorm:"type:text" json:"metadata"` // JSON
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (TaskModel) TableName() string {
	return "tasks"
}

// ExecutionModel represents an execution record in the database
type ExecutionModel struct {
	BaseModel
	SessionID   string     `gorm:"size:64;index;not null" json:"session_id"`
	TaskID      string     `gorm:"size:64;index" json:"task_id"`
	Prompt      string     `gorm:"type:text" json:"prompt"`
	Status      string     `gorm:"size:32;not null" json:"status"`
	Output      string     `gorm:"type:text" json:"output"`
	Error       string     `gorm:"type:text" json:"error"`
	TokensIn    int        `json:"tokens_in"`
	TokensOut   int        `json:"tokens_out"`
	DurationMs  int64      `json:"duration_ms"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (ExecutionModel) TableName() string {
	return "executions"
}

// WebhookModel represents a webhook in the database
type WebhookModel struct {
	BaseModel
	Name      string `gorm:"size:255;not null" json:"name"`
	URL       string `gorm:"size:1024;not null" json:"url"`
	Secret    string `gorm:"size:255" json:"secret"`
	Events    string `gorm:"type:text" json:"events"` // JSON array
	IsEnabled bool   `gorm:"default:true" json:"is_enabled"`
	Metadata  string `gorm:"type:text" json:"metadata"` // JSON
}

func (WebhookModel) TableName() string {
	return "webhooks"
}

// ImageModel represents a Docker image in the database
type ImageModel struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Tag         string `gorm:"size:128;not null" json:"tag"`
	Agent       string `gorm:"size:64;index" json:"agent"`
	Digest      string `gorm:"size:128" json:"digest"`
	Size        int64  `json:"size"`
	Status      string `gorm:"size:32" json:"status"`
	Description string `gorm:"size:1024" json:"description"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

func (ImageModel) TableName() string {
	return "images"
}

// HistoryModel represents execution history
type HistoryModel struct {
	BaseModel
	SessionID string `gorm:"size:64;index" json:"session_id"`
	TaskID    string `gorm:"size:64;index" json:"task_id"`
	Action    string `gorm:"size:64;not null" json:"action"`
	Details   string `gorm:"type:text" json:"details"` // JSON
	UserID    string `gorm:"size:64" json:"user_id"`
}

func (HistoryModel) TableName() string {
	return "history"
}
