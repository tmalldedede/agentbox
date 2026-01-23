package database

import (
	"github.com/tmalldedede/agentbox/internal/logger"
)

var migLog = logger.Module("database.migrations")

// AutoMigrate runs database migrations
func AutoMigrate() error {
	migLog.Info("running database migrations...")

	models := []interface{}{
		&MCPServerModel{},
		&SkillModel{},
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

// SeedBuiltInData seeds built-in MCP servers and skills
func SeedBuiltInData() error {
	migLog.Info("seeding built-in data...")

	// Check if already seeded (use MCP servers as indicator)
	var count int64
	DB.Model(&MCPServerModel{}).Where("is_built_in = ?", true).Count(&count)
	if count > 0 {
		migLog.Debug("built-in data already exists, skipping seed")
		return nil
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
