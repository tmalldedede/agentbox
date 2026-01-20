// @title AgentBox API
// @version 1.0
// @description AgentBox - Open-source AI Agent containerized runtime platform
// @description
// @description ## Overview
// @description AgentBox provides a unified API for managing AI coding agents (Claude Code, Codex, OpenCode) in containerized environments.
// @description
// @description ## Authentication
// @description Currently no authentication required. API keys are managed via Credentials API.
// @description
// @description ## API Structure
// @description - **Public API** (`/api/v1/*`): Core functionality for external integrations
// @description - **Admin API** (`/api/v1/admin/*`): Platform management operations

// @contact.name AgentBox Team
// @contact.url https://github.com/user/agentbox
// @contact.email support@agentbox.dev

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:18080
// @BasePath /api/v1

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name Agents
// @tag.description Agent types (Claude Code, Codex, OpenCode)

// @tag.name Profiles
// @tag.description Runtime configuration templates for agents

// @tag.name Providers
// @tag.description API provider presets (Anthropic, OpenAI, DeepSeek, etc.)

// @tag.name Sessions
// @tag.description Container sessions for running agents

// @tag.name Tasks
// @tag.description Async task queue for batch processing

// @tag.name Webhooks
// @tag.description Webhook notifications for task events

// @tag.name MCP Servers
// @tag.description Model Context Protocol server management (Admin)

// @tag.name Skills
// @tag.description Reusable prompt templates (Admin)

// @tag.name Credentials
// @tag.description API key and token management (Admin)

// @tag.name System
// @tag.description System health and resource management (Admin)

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tmalldedede/agentbox/internal/api"
	"github.com/tmalldedede/agentbox/internal/app"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/logger"
)

const (
	version = "0.1.0"
	banner  = `
    _                    _   ____
   / \   __ _  ___ _ __ | |_| __ )  _____  __
  / _ \ / _' |/ _ \ '_ \| __|  _ \ / _ \ \/ /
 / ___ \ (_| |  __/ | | | |_| |_) | (_) >  <
/_/   \_\__, |\___|_| |_|\___|____/ \___/_/\_\
        |___/
`
)

func main() {
	// 初始化日志
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(&logger.Config{
		Level:  logLevel,
		Format: "text",
	})
	log := logger.Module("main")

	// 打印 Banner
	fmt.Print(banner)
	fmt.Printf("AgentBox v%s\n", version)
	fmt.Println("Open-source AI Agent containerized runtime platform")
	fmt.Println()

	// 加载配置
	cfg := config.Load()

	// 创建应用程序
	application, err := app.New(cfg)
	if err != nil {
		log.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	// 启动后台服务（Task Manager 等）
	application.Start()

	// 创建 HTTP 服务器
	server := api.NewServer(&api.Deps{
		Session:    application.Session,
		Registry:   application.AgentRegistry,
		Container:  application.Container,
		Profile:    application.Profile,
		Provider:   application.Provider,
		MCP:        application.MCP,
		Skill:      application.Skill,
		Credential: application.Credential,
		Task:       application.Task,
		Webhook:    application.Webhook,
	})

	// 打印 API 路由信息
	addr := application.ServerAddr()
	log.Info("starting server", "addr", addr)
	printRoutes()

	// 在后台启动服务器
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Run(addr); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", "error", err)
	case sig := <-quit:
		log.Info("received shutdown signal", "signal", sig)
	}

	// 优雅关闭
	log.Info("shutting down server...")

	// 给服务器 30 秒时间完成当前请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	// 关闭应用程序（Task Manager, Container Manager 等）
	if err := application.Close(); err != nil {
		log.Error("application close error", "error", err)
	}

	log.Info("server stopped")
}

// printRoutes 打印 API 路由信息
func printRoutes() {
	fmt.Println()
	fmt.Println("Public API (对外服务，参考 Manus API):")
	fmt.Println("  GET    /api/v1/health                 - Health check")
	fmt.Println("  GET    /api/v1/agents                 - List agents")
	fmt.Println("  *      /api/v1/profiles/*             - Profile management (CRUD)")
	fmt.Println("  *      /api/v1/providers/*            - Provider management (CRUD)")
	fmt.Println("  *      /api/v1/sessions/*             - Session management (CRUD)")
	fmt.Println("  *      /api/v1/tasks/*                - Task management (CRUD)")
	fmt.Println("  *      /api/v1/files/*                - File upload (CRUD)")
	fmt.Println("  *      /api/v1/webhooks/*             - Webhook management (CRUD)")
	fmt.Println()
	fmt.Println("Admin API (平台管理):")
	fmt.Println("  *      /api/v1/admin/mcp-servers/*    - MCP server management")
	fmt.Println("  *      /api/v1/admin/skills/*         - Skill management")
	fmt.Println("  *      /api/v1/admin/credentials/*    - Credential management")
	fmt.Println("  *      /api/v1/admin/images/*         - Image management")
	fmt.Println("  *      /api/v1/admin/system/*         - System management")
	fmt.Println()
}
