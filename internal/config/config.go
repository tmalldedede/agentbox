package config

import (
	"os"
	"strconv"
	"time"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `json:"server"`
	Container ContainerConfig `json:"container"`
	Storage   StorageConfig   `json:"storage"`
	Database  DatabaseConfig  `json:"database"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ContainerConfig 容器默认配置
type ContainerConfig struct {
	CPULimit      float64       `json:"cpu_limit"`       // CPU 核心数
	MemoryLimit   int64         `json:"memory_limit"`    // 内存限制 (bytes)
	DiskLimit     int64         `json:"disk_limit"`      // 磁盘限制 (bytes)
	Timeout       time.Duration `json:"timeout"`         // 执行超时
	NetworkMode   string        `json:"network_mode"`    // 网络模式
	WorkspaceBase string        `json:"workspace_base"`  // 工作空间基础目录
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type string `json:"type"` // sqlite, postgres
	DSN  string `json:"dsn"`  // 数据源
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `json:"driver"`    // sqlite, postgres
	DSN      string `json:"dsn"`       // 数据源名称
	LogLevel string `json:"log_level"` // silent, error, warn, info
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 18080,
		},
		Container: ContainerConfig{
			CPULimit:      2.0,
			MemoryLimit:   4 * 1024 * 1024 * 1024, // 4GB
			DiskLimit:     10 * 1024 * 1024 * 1024, // 10GB
			Timeout:       1 * time.Hour,
			NetworkMode:   "bridge",
			WorkspaceBase: "/tmp/agentbox/workspaces",
		},
		Storage: StorageConfig{
			Type: "sqlite",
			DSN:  "agentbox.db",
		},
		Database: DatabaseConfig{
			Driver:   "sqlite",
			DSN:      "", // Will use default path ~/.agentbox/data/agentbox.db
			LogLevel: "warn",
		},
	}
}

// Load 从环境变量加载配置
func Load() *Config {
	cfg := Default()

	// 从环境变量覆盖
	if host := os.Getenv("AGENTBOX_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("AGENTBOX_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if workspace := os.Getenv("AGENTBOX_WORKSPACE_BASE"); workspace != "" {
		cfg.Container.WorkspaceBase = workspace
	}

	// 数据库配置
	if driver := os.Getenv("AGENTBOX_DB_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	if dsn := os.Getenv("AGENTBOX_DB_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if logLevel := os.Getenv("AGENTBOX_DB_LOG_LEVEL"); logLevel != "" {
		cfg.Database.LogLevel = logLevel
	}

	return cfg
}
