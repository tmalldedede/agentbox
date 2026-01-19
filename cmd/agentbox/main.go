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
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tmalldedede/agentbox/internal/agent"
	_ "github.com/tmalldedede/agentbox/internal/agent/claude"   // 注册 Claude Code 适配器
	_ "github.com/tmalldedede/agentbox/internal/agent/codex"    // 注册 Codex 适配器
	_ "github.com/tmalldedede/agentbox/internal/agent/opencode" // 注册 OpenCode 适配器
	"github.com/tmalldedede/agentbox/internal/api"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"github.com/tmalldedede/agentbox/internal/task"
	"github.com/tmalldedede/agentbox/internal/webhook"
)

const (
	version = "0.1.0"
	banner  = `
    _                    _   ____
   / \   __ _  ___ _ __ | |_| __ )  _____  __
  / _ \ / _' |/ _ \ '_ \| __|  _ \ / _ \ \/ /
 / ___ \ (_| |  __/ | | | |_| |_) | (_) >  <
/_/   \_\__, |\___|_| |_|\__|____/ \___/_/\_\
        |___/
`
)

func main() {
	fmt.Print(banner)
	fmt.Printf("AgentBox v%s\n", version)
	fmt.Println("Open-source AI Agent containerized runtime platform")
	fmt.Println()

	// 加载配置
	cfg := config.Load()

	// 初始化容器管理器
	containerMgr, err := container.NewDockerManager()
	if err != nil {
		log.Fatalf("Failed to initialize Docker manager: %v", err)
	}
	defer containerMgr.Close()

	// 测试 Docker 连接
	ctx := context.Background()
	if err := containerMgr.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to Docker: %v", err)
	}
	log.Println("Docker connection OK")

	// 初始化存储
	store := session.NewMemoryStore()

	// 获取 Agent 注册表
	registry := agent.DefaultRegistry()
	log.Printf("Registered agents: %v", registry.Names())

	// 初始化会话管理器
	sessionMgr := session.NewManager(store, containerMgr, registry, cfg.Container.WorkspaceBase)

	// 初始化 Profile 管理器
	profileDataDir := filepath.Join(cfg.Container.WorkspaceBase, "profiles")
	profileMgr, err := profile.NewManager(profileDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize Profile manager: %v", err)
	}
	log.Printf("Loaded %d profiles", len(profileMgr.List()))

	// 设置 Profile Manager 到 Session Manager（用于 Codex 配置文件生成）
	sessionMgr.SetProfileManager(profileMgr)

	// 初始化 Provider 管理器
	providerDataDir := filepath.Join(cfg.Container.WorkspaceBase, "providers")
	providerMgr := provider.NewManager(providerDataDir)
	log.Printf("Loaded %d providers (%d built-in)", len(providerMgr.List()), len(provider.GetBuiltinProviders()))

	// 初始化 MCP Server 管理器
	mcpDataDir := filepath.Join(cfg.Container.WorkspaceBase, "mcp-servers")
	mcpMgr, err := mcp.NewManager(mcpDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize MCP manager: %v", err)
	}
	log.Printf("Loaded %d MCP servers", len(mcpMgr.List()))

	// 初始化 Skill 管理器
	skillDataDir := filepath.Join(cfg.Container.WorkspaceBase, "skills")
	skillMgr, err := skill.NewManager(skillDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize Skill manager: %v", err)
	}
	log.Printf("Loaded %d skills", len(skillMgr.List()))

	// 初始化 Credential 管理器
	credentialDataDir := filepath.Join(cfg.Container.WorkspaceBase, "credentials")
	// 使用固定的加密密钥（生产环境应从配置或环境变量读取）
	encryptionKey := os.Getenv("AGENTBOX_ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "agentbox-default-encryption-key-32b" // 默认密钥，仅用于开发
	}
	credentialMgr, err := credential.NewManager(credentialDataDir, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize Credential manager: %v", err)
	}
	log.Printf("Loaded %d credentials", len(credentialMgr.List()))

	// 初始化 Task Store (SQLite)
	taskDBPath := filepath.Join(cfg.Container.WorkspaceBase, "agentbox.db")
	taskStore, err := task.NewSQLiteStore(taskDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize Task store: %v", err)
	}
	defer taskStore.Close()
	log.Printf("Task database initialized: %s", taskDBPath)

	// 初始化 Task Manager
	taskMgr := task.NewManager(taskStore, profileMgr, sessionMgr, nil)
	taskMgr.Start()
	defer taskMgr.Stop()

	// 初始化 Webhook Manager
	webhookDataDir := filepath.Join(cfg.Container.WorkspaceBase, "webhooks")
	webhookMgr, err := webhook.NewManager(webhookDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize Webhook manager: %v", err)
	}
	webhooks, _ := webhookMgr.List()
	log.Printf("Loaded %d webhooks", len(webhooks))

	// 创建 HTTP 服务器
	server := api.NewServer(sessionMgr, registry, containerMgr, profileMgr, providerMgr, mcpMgr, skillMgr, credentialMgr, taskMgr, webhookMgr)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Println()
	log.Println("Public API (对外服务，参考 Manus API):")
	log.Println("  GET    /api/v1/health                 - Health check")
	log.Println("  GET    /api/v1/agents                 - List agents")
	log.Println("  *      /api/v1/profiles/*             - Profile management (CRUD)")
	log.Println("  *      /api/v1/providers/*            - Provider management (CRUD)")
	log.Println("  *      /api/v1/sessions/*             - Session management (CRUD)")
	log.Println("  *      /api/v1/tasks/*                - Task management (CRUD)")
	log.Println("  *      /api/v1/files/*                - File upload (CRUD)")
	log.Println("  *      /api/v1/webhooks/*             - Webhook management (CRUD)")
	log.Println()
	log.Println("Admin API (平台管理):")
	log.Println("  *      /api/v1/admin/mcp-servers/*    - MCP server management")
	log.Println("  *      /api/v1/admin/skills/*         - Skill management")
	log.Println("  *      /api/v1/admin/credentials/*    - Credential management")
	log.Println("  *      /api/v1/admin/images/*         - Image management")
	log.Println("  *      /api/v1/admin/system/*         - System management")
	log.Println()

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := server.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
