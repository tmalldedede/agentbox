package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Driver   string // "sqlite" or "postgres"
	DSN      string // Data Source Name
	LogLevel string // "silent", "error", "warn", "info"
}

// DB is the global database instance
var DB *gorm.DB

var log = logger.Module("database")

// Initialize initializes the database connection
func Initialize(cfg Config) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite", "":
		// Default to SQLite
		if cfg.DSN == "" {
			// Default path
			homeDir, _ := os.UserHomeDir()
			dataDir := filepath.Join(homeDir, ".agentbox", "data")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return fmt.Errorf("failed to create data directory: %w", err)
			}
			cfg.DSN = filepath.Join(dataDir, "agentbox.db")
		}
		dialector = sqlite.Open(cfg.DSN)
		log.Info("using SQLite database", "path", cfg.DSN)

	case "postgres":
		if cfg.DSN == "" {
			return fmt.Errorf("PostgreSQL DSN is required")
		}
		dialector = postgres.Open(cfg.DSN)
		log.Info("using PostgreSQL database")

	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Configure GORM logger
	var gormLogLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		gormLogLevel = gormlogger.Silent
	case "error":
		gormLogLevel = gormlogger.Error
	case "warn":
		gormLogLevel = gormlogger.Warn
	case "info":
		gormLogLevel = gormlogger.Info
	default:
		gormLogLevel = gormlogger.Warn
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	var err error
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for PostgreSQL
	if cfg.Driver == "postgres" {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// Run migrations
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("database initialized successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
