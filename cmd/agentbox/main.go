package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tmalldedede/agentbox/internal/agent"
	_ "github.com/tmalldedede/agentbox/internal/agent/claude" // 注册 Claude Code 适配器
	_ "github.com/tmalldedede/agentbox/internal/agent/codex"  // 注册 Codex 适配器
	"github.com/tmalldedede/agentbox/internal/api"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/session"
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

	// 创建 HTTP 服务器
	server := api.NewServer(sessionMgr, registry)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Println()
	log.Println("API Endpoints:")
	log.Println("  GET  /health              - Health check")
	log.Println("  GET  /api/agents          - List agents")
	log.Println("  POST /api/sessions        - Create session")
	log.Println("  GET  /api/sessions        - List sessions")
	log.Println("  GET  /api/sessions/:id    - Get session")
	log.Println("  DELETE /api/sessions/:id  - Delete session")
	log.Println("  POST /api/sessions/:id/exec - Execute prompt")
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
