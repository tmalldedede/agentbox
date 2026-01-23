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
	Files     FilesConfig     `json:"files"`
}

// FilesConfig 文件上传配置
type FilesConfig struct {
	UploadDir      string        `json:"upload_dir"`       // 上传文件存储目录
	RetentionHours int           `json:"retention_hours"`  // 文件保留时长（小时），0 表示永不过期
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理扫描间隔
	MaxFileSize    int64         `json:"max_file_size"`    // 单文件最大大小 (bytes)
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
	GCInterval    time.Duration `json:"gc_interval"`     // GC 扫描间隔
	ContainerTTL  time.Duration `json:"container_ttl"`   // 容器最大存活时间
	IdleTimeout   time.Duration `json:"idle_timeout"`    // Stopped 状态后多久删除
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
			GCInterval:    60 * time.Second,
			ContainerTTL:  2 * time.Hour,
			IdleTimeout:   10 * time.Minute,
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
		Files: FilesConfig{
			UploadDir:       "", // 为空时使用 {WorkspaceBase}/uploads
			RetentionHours:  72, // 默认 3 天
			CleanupInterval: 30 * time.Minute,
			MaxFileSize:     100 * 1024 * 1024, // 100MB
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

	// GC 配置
	if v := os.Getenv("AGENTBOX_GC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Container.GCInterval = d
		}
	}
	if v := os.Getenv("AGENTBOX_CONTAINER_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Container.ContainerTTL = d
		}
	}
	if v := os.Getenv("AGENTBOX_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Container.IdleTimeout = d
		}
	}

	// 文件存储配置
	if v := os.Getenv("AGENTBOX_UPLOAD_DIR"); v != "" {
		cfg.Files.UploadDir = v
	}
	if v := os.Getenv("AGENTBOX_FILE_RETENTION_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil {
			cfg.Files.RetentionHours = h
		}
	}
	if v := os.Getenv("AGENTBOX_FILE_MAX_SIZE"); v != "" {
		if s, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Files.MaxFileSize = s
		}
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
