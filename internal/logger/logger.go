// Package logger 提供统一的日志接口
//
// 使用方式:
//
//	import "github.com/tmalldedede/agentbox/internal/logger"
//
//	// 模块级日志
//	log := logger.With("module", "session")
//	log.Info("session created", "id", sessionID)
//	log.Error("failed to start", "error", err)
//
//	// 带请求上下文
//	log := logger.With("module", "api", "request_id", requestID)
//	log.Debug("handling request", "path", path)
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

// Level 日志级别
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config 日志配置
type Config struct {
	Level  string // debug, info, warn, error
	Format string // text, json
	Output io.Writer
}

// Init 初始化全局日志器
func Init(cfg *Config) {
	once.Do(func() {
		if cfg == nil {
			cfg = &Config{
				Level:  "info",
				Format: "text",
				Output: os.Stderr,
			}
		}

		level := parseLevel(cfg.Level)
		opts := &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// 简化时间格式
				if a.Key == slog.TimeKey {
					return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format("15:04:05.000"))}
				}
				return a
			},
		}

		output := cfg.Output
		if output == nil {
			output = os.Stderr
		}

		var handler slog.Handler
		if cfg.Format == "json" {
			handler = slog.NewJSONHandler(output, opts)
		} else {
			handler = slog.NewTextHandler(output, opts)
		}

		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger)
	})
}

// parseLevel 解析日志级别
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Default 获取默认日志器
func Default() *slog.Logger {
	if defaultLogger == nil {
		Init(nil)
	}
	return defaultLogger
}

// With 创建带有固定属性的子日志器
// 常用于模块级日志:
//
//	log := logger.With("module", "session")
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// WithContext 从上下文获取日志器
func WithContext(ctx context.Context) *slog.Logger {
	// 未来可以从 context 中提取 trace_id 等
	return Default()
}

// 便捷方法 - 直接使用默认日志器

func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// Module 创建模块日志器的便捷方法
func Module(name string) *slog.Logger {
	return With("module", name)
}
