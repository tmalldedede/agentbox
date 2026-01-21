package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tmalldedede/agentbox/internal/agent"
	_ "github.com/tmalldedede/agentbox/internal/agent/claude"   // 注册 Claude Code 适配器
	_ "github.com/tmalldedede/agentbox/internal/agent/codex"    // 注册 Codex 适配器
	_ "github.com/tmalldedede/agentbox/internal/agent/opencode" // 注册 OpenCode 适配器
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"github.com/tmalldedede/agentbox/internal/smartagent"
	"github.com/tmalldedede/agentbox/internal/task"
	"github.com/tmalldedede/agentbox/internal/webhook"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("app")
}

// App 应用程序容器，管理所有依赖
type App struct {
	// 配置
	Config *config.Config

	// 核心组件
	Container     container.Manager
	AgentRegistry *agent.Registry
	Session       *session.Manager
	Task          *task.Manager

	// 配置管理
	Profile    *profile.Manager
	Provider   *provider.Manager
	MCP        *mcp.Manager
	Skill      *skill.Manager
	Credential *credential.Manager
	Webhook    *webhook.Manager

	// 智能体管理
	SmartAgent *smartagent.Manager

	// 执行历史
	History *history.Manager

	// 内部状态
	taskStore *task.SQLiteStore
}

// New 创建应用程序实例
func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	if err := app.initialize(); err != nil {
		app.Close() // 清理已初始化的资源
		return nil, err
	}

	return app, nil
}

// initialize 初始化所有组件
func (a *App) initialize() error {
	var err error

	// 1. 初始化容器管理器
	a.Container, err = container.NewDockerManager()
	if err != nil {
		return fmt.Errorf("failed to initialize Docker manager: %w", err)
	}

	// 测试 Docker 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Container.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}
	log.Info("Docker connection OK")

	// 2. 获取 Agent 注册表
	a.AgentRegistry = agent.DefaultRegistry()
	log.Info("registered agents", "agents", a.AgentRegistry.Names())

	// 3. 初始化 Session 管理器
	sessionStore := session.NewMemoryStore()
	a.Session = session.NewManager(sessionStore, a.Container, a.AgentRegistry, a.Config.Container.WorkspaceBase)

	// 4. 初始化 Profile 管理器
	profileDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "profiles")
	a.Profile, err = profile.NewManager(profileDataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize Profile manager: %w", err)
	}
	log.Info("loaded profiles", "count", len(a.Profile.List()))

	// 设置 Profile Manager 到 Session Manager
	a.Session.SetProfileManager(a.Profile)

	// 5. 初始化 Provider 管理器
	providerDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "providers")
	a.Provider = provider.NewManager(providerDataDir)
	log.Info("loaded providers", "count", len(a.Provider.List()), "builtin", len(provider.GetBuiltinProviders()))

	// 6. 初始化 MCP Server 管理器
	mcpDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "mcp-servers")
	a.MCP, err = mcp.NewManager(mcpDataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP manager: %w", err)
	}
	log.Info("loaded MCP servers", "count", len(a.MCP.List()))

	// 7. 初始化 Skill 管理器
	skillDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "skills")
	a.Skill, err = skill.NewManager(skillDataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize Skill manager: %w", err)
	}
	log.Info("loaded skills", "count", len(a.Skill.List()))

	// 8. 初始化 Credential 管理器
	credentialDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "credentials")
	encryptionKey := os.Getenv("AGENTBOX_ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "agentbox-default-encryption-key-32b" // 默认密钥，仅用于开发
	}
	a.Credential, err = credential.NewManager(credentialDataDir, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to initialize Credential manager: %w", err)
	}
	log.Info("loaded credentials", "count", len(a.Credential.List()))

	// 设置 Credential Manager 到 Session Manager
	a.Session.SetCredentialManager(a.Credential)

	// 设置 Skill Manager 到 Session Manager
	a.Session.SetSkillManager(a.Skill)

	// 9. 初始化 Task Store (SQLite)
	taskDBPath := filepath.Join(a.Config.Container.WorkspaceBase, "agentbox.db")
	a.taskStore, err = task.NewSQLiteStore(taskDBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize Task store: %w", err)
	}
	log.Info("task database initialized", "path", taskDBPath)

	// 10. 初始化 Task Manager
	a.Task = task.NewManager(a.taskStore, a.Profile, a.Session, nil)

	// 11. 初始化 Webhook Manager
	webhookDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "webhooks")
	a.Webhook, err = webhook.NewManager(webhookDataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize Webhook manager: %w", err)
	}
	webhooks, _ := a.Webhook.List()
	log.Info("loaded webhooks", "count", len(webhooks))

	// 12. 初始化 SmartAgent Manager
	smartAgentDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "agents")
	a.SmartAgent = smartagent.NewManager(smartAgentDataDir, a.Profile)
	log.Info("loaded smart agents", "count", len(a.SmartAgent.List()))

	// 13. 初始化 History Manager
	historyStore := history.NewMemoryStore()
	a.History = history.NewManager(historyStore)
	log.Info("history manager initialized")

	return nil
}

// Start 启动后台服务
func (a *App) Start() {
	a.Task.Start()
	log.Info("app started")
}

// Close 关闭所有资源
func (a *App) Close() error {
	log.Info("closing app...")

	// 按初始化的逆序关闭
	if a.Task != nil {
		a.Task.Stop()
	}

	if a.taskStore != nil {
		a.taskStore.Close()
	}

	if a.Container != nil {
		a.Container.Close()
	}

	log.Info("app closed")
	return nil
}

// ServerAddr 返回服务器监听地址
func (a *App) ServerAddr() string {
	return fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port)
}
