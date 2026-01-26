package database

import (
	"github.com/tmalldedede/agentbox/internal/logger"
)

var migLog = logger.Module("database.migrations")

// AutoMigrate runs database migrations
func AutoMigrate() error {
	migLog.Info("running database migrations...")

	models := []interface{}{
		&UserModel{},
		&APIKeyModel{},
		&AuthProfileModel{},
		&MCPServerModel{},
		&SkillModel{},
		&SessionModel{},
		&TaskModel{},
		&ExecutionModel{},
		&WebhookModel{},
		&ImageModel{},
		&HistoryModel{},
		&BatchModel{},
		&BatchTaskModel{},
		&FileModel{},
		&ChannelSessionModel{},
		&ChannelMessageModel{},
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return err
		}
	}

	migLog.Info("database migrations completed")
	return nil
}

// SeedBuiltInData seeds built-in data (currently empty, user manages all MCP servers)
func SeedBuiltInData() error {
	migLog.Info("seeding built-in data...")
	// No built-in MCP servers - user manages all configurations
	migLog.Info("built-in data seeded successfully")
	return nil
}
