package database

import (
	"github.com/tmalldedede/agentbox/internal/logger"
)

var migLog = logger.Module("database.migrations")

// AutoMigrate runs database migrations
func AutoMigrate() error {
	migLog.Info("running database migrations...")

	models := []interface{}{
		&ProfileModel{},
		&MCPServerModel{},
		&SkillModel{},
		&CredentialModel{},
		&SessionModel{},
		&TaskModel{},
		&ExecutionModel{},
		&WebhookModel{},
		&ImageModel{},
		&HistoryModel{},
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return err
		}
	}

	migLog.Info("database migrations completed")
	return nil
}

// SeedBuiltInData seeds built-in profiles, MCP servers, and skills
func SeedBuiltInData() error {
	migLog.Info("seeding built-in data...")

	// Check if already seeded
	var count int64
	DB.Model(&ProfileModel{}).Where("is_built_in = ?", true).Count(&count)
	if count > 0 {
		migLog.Debug("built-in data already exists, skipping seed")
		return nil
	}

	// Seed built-in profiles
	builtInProfiles := []ProfileModel{
		{
			BaseModel:   BaseModel{ID: "claude-code-default"},
			Name:        "Claude Code Default",
			Description: "Default Claude Code profile with standard settings",
			Adapter:     "claude-code",
			IsBuiltIn:   true,
			IsPublic:    true,
		},
		{
			BaseModel:   BaseModel{ID: "claude-code-auto"},
			Name:        "Claude Code Auto",
			Description: "Claude Code with auto-accept edits mode",
			Adapter:     "claude-code",
			Permissions: `{"mode":"acceptEdits"}`,
			IsBuiltIn:   true,
			IsPublic:    true,
		},
		{
			BaseModel:   BaseModel{ID: "codex-default"},
			Name:        "Codex Default",
			Description: "Default Codex profile with standard settings",
			Adapter:     "codex",
			IsBuiltIn:   true,
			IsPublic:    true,
		},
		{
			BaseModel:   BaseModel{ID: "codex-full-auto"},
			Name:        "Codex Full Auto",
			Description: "Codex with full auto mode and danger full access",
			Adapter:     "codex",
			Permissions: `{"sandbox_mode":"danger-full-access","approval_policy":"never","full_auto":true}`,
			IsBuiltIn:   true,
			IsPublic:    true,
		},
		{
			BaseModel:   BaseModel{ID: "security-research"},
			Name:        "Security Research",
			Description: "Profile optimized for security research with cybersec tools",
			Adapter:     "claude-code",
			Permissions: `{"mode":"bypassPermissions"}`,
			IsBuiltIn:   true,
			IsPublic:    true,
		},
	}

	for _, p := range builtInProfiles {
		if err := DB.Create(&p).Error; err != nil {
			migLog.Warn("failed to create built-in profile", "id", p.ID, "error", err)
		}
	}

	// Seed built-in MCP servers
	builtInMCPServers := []MCPServerModel{
		{
			BaseModel:   BaseModel{ID: "filesystem"},
			Name:        "Filesystem",
			Description: "File system access MCP server",
			Command:     "npx",
			Args:        `["-y","@anthropic/mcp-server-filesystem","/workspace"]`,
			IsBuiltIn:   true,
			IsEnabled:   true,
		},
		{
			BaseModel:   BaseModel{ID: "github"},
			Name:        "GitHub",
			Description: "GitHub API MCP server",
			Command:     "npx",
			Args:        `["-y","@anthropic/mcp-server-github"]`,
			Env:         `{"GITHUB_TOKEN":""}`,
			IsBuiltIn:   true,
			IsEnabled:   false,
		},
		{
			BaseModel:   BaseModel{ID: "puppeteer"},
			Name:        "Puppeteer",
			Description: "Browser automation MCP server",
			Command:     "npx",
			Args:        `["-y","@anthropic/mcp-server-puppeteer"]`,
			IsBuiltIn:   true,
			IsEnabled:   false,
		},
	}

	for _, m := range builtInMCPServers {
		if err := DB.Create(&m).Error; err != nil {
			migLog.Warn("failed to create built-in MCP server", "id", m.ID, "error", err)
		}
	}

	migLog.Info("built-in data seeded successfully")
	return nil
}
