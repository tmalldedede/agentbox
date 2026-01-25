package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/auth"
	"github.com/tmalldedede/agentbox/internal/batch"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/database"
	"github.com/tmalldedede/agentbox/internal/engine"
	_ "github.com/tmalldedede/agentbox/internal/engine/claude"   // 注册 Claude Code 适配器
	_ "github.com/tmalldedede/agentbox/internal/engine/codex"    // 注册 Codex 适配器
	_ "github.com/tmalldedede/agentbox/internal/engine/opencode" // 注册 OpenCode 适配器
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/settings"
	"github.com/tmalldedede/agentbox/internal/skill"
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

	// 认证
	Auth *auth.Manager

	// 核心组件
	Container     container.Manager
	AgentRegistry *engine.Registry
	Session       *session.Manager
	Task          *task.Manager
	Batch         *batch.Manager
	GC            *container.GarbageCollector

	// 配置管理
	Provider *provider.Manager
	Runtime  *runtime.Manager
	MCP      *mcp.Manager
	Skill    *skill.Manager
	Webhook  *webhook.Manager

	// 智能体管理
	Agent *agent.Manager

	// 执行历史
	History *history.Manager

	// 业务配置
	Settings *settings.Manager

	// Redis 队列
	RedisQueue *batch.RedisQueue

	// 内部状态
	taskStore  *task.GormStore
	batchStore batch.Store
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

	// 1. 初始化容器管理器（Docker 不可用时降级为 NoopManager）
	a.Container, err = container.NewDockerManager()
	if err != nil {
		log.Warn("Docker manager initialization failed, running in degraded mode", "error", err)
		a.Container = container.NewNoopManager()
	} else {
		// 测试 Docker 连接
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Container.Ping(ctx); err != nil {
			log.Warn("Docker connection failed, running in degraded mode", "error", err)
			a.Container.Close()
			a.Container = container.NewNoopManager()
		} else {
			log.Info("Docker connection OK")
		}
	}

	// 1.5. 初始化认证管理器
	a.Auth = auth.NewManager(database.GetDB())
	log.Info("auth manager initialized")

	// 2. 获取 Agent 注册表
	a.AgentRegistry = engine.DefaultRegistry()
	log.Info("registered agents", "agents", a.AgentRegistry.Names())

	// 3. 初始化 Session 管理器
	var sessionStore session.Store
	if dbStore, err := session.NewDBStore(database.GetDB()); err != nil {
		log.Warn("failed to initialize DB session store, falling back to memory", "error", err)
		sessionStore = session.NewMemoryStore()
	} else {
		sessionStore = dbStore
	}
	a.Session = session.NewManager(sessionStore, a.Container, a.AgentRegistry, a.Config.Container.WorkspaceBase)

	// 3.5. 初始化 GC (依赖 Session Manager)
	a.GC = container.NewGarbageCollector(a.Container, a.Session, container.GCConfig{
		Interval:     a.Config.Container.GCInterval,
		ContainerTTL: a.Config.Container.ContainerTTL,
		IdleTimeout:  a.Config.Container.IdleTimeout,
	})

	// 获取加密密钥
	encryptionKey := os.Getenv("AGENTBOX_ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "agentbox-default-encryption-key-32b" // 默认密钥，仅用于开发
	}

	// 4. 初始化 Provider 管理器
	providerDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "providers")
	a.Provider = provider.NewManager(providerDataDir, encryptionKey)
	log.Info("loaded providers", "count", len(a.Provider.List()), "builtin", len(provider.GetBuiltinProviders()))

	// 5.5. 初始化 Runtime 管理器
	runtimeDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "runtimes")
	a.Runtime = runtime.NewManager(runtimeDataDir, a.Config)
	log.Info("loaded runtimes", "count", len(a.Runtime.List()))

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

	// 8. 设置 Skill Manager 到 Session Manager
	a.Session.SetSkillManager(a.Skill)

	// 9. 初始化 Agent Manager（Task Manager 依赖它）
	agentDataDir := filepath.Join(a.Config.Container.WorkspaceBase, "agents")
	a.Agent = agent.NewManager(agentDataDir, a.Provider, a.Runtime, a.Skill, a.MCP)
	log.Info("loaded agents", "count", len(a.Agent.List()))

	// 设置 Agent Manager 到 Session Manager
	a.Session.SetAgentManager(a.Agent)

	// 10. 解析文件上传目录
	if a.Config.Files.UploadDir == "" {
		a.Config.Files.UploadDir = filepath.Join(a.Config.Container.WorkspaceBase, "uploads")
	}
	os.MkdirAll(a.Config.Files.UploadDir, 0755)
	log.Info("file upload directory", "path", a.Config.Files.UploadDir)

	// 11. 初始化 Task Store (GORM)
	a.taskStore, err = task.NewGormStore(database.GetDB())
	if err != nil {
		return fmt.Errorf("failed to initialize Task store: %w", err)
	}
	log.Info("task store initialized (GORM)")

	// 12. 初始化 Task Manager
	a.Task = task.NewManager(a.taskStore, a.Agent, a.Session, &task.ManagerConfig{})

	// 12. 初始化 Webhook Manager（使用数据库存储）
	a.Webhook = webhook.NewManager()
	webhooks, _ := a.Webhook.List()
	log.Info("loaded webhooks", "count", len(webhooks))

	// 连接 Webhook 到 Task Manager
	a.Task.SetWebhookNotifier(a.Webhook)

	// 13. 初始化 History Manager
	var historyStore history.Store
	if dbHistStore, err := history.NewDBStore(database.GetDB()); err != nil {
		log.Warn("failed to initialize DB history store, falling back to memory", "error", err)
		historyStore = history.NewMemoryStore()
	} else {
		historyStore = dbHistStore
	}
	a.History = history.NewManager(historyStore)
	log.Info("history manager initialized")

	// 14. 初始化 Settings Manager
	a.Settings, err = settings.NewManager(database.GetDB())
	if err != nil {
		return fmt.Errorf("failed to initialize Settings manager: %w", err)
	}
	log.Info("settings manager initialized")

	// 15. 初始化 Batch Manager (使用 GORM + Redis)
	a.batchStore = batch.NewGormStore()
	log.Info("batch store initialized (GORM)")

	// 初始化 Redis 队列（可选）
	var redisQueue *batch.RedisQueue
	if a.Config.Redis.Enabled {
		redisQueue, err = batch.NewRedisQueue(a.Config.Redis)
		if err != nil {
			log.Warn("Redis queue initialization failed, running without Redis", "error", err)
		} else {
			a.RedisQueue = redisQueue
			log.Info("Redis queue initialized", "addr", a.Config.Redis.Addr)
		}
	}

	a.Batch = batch.NewManager(a.batchStore, a.Session, a.Agent, &batch.ManagerConfig{
		MaxBatches:       10,                      // 最多同时运行 10 个 batch
		PollInterval:     100 * time.Millisecond,  // 任务轮询间隔
		ProgressInterval: 1 * time.Second,         // 进度更新间隔
		RedisQueue:       redisQueue,
	})
	log.Info("batch manager initialized")

	return nil
}

// Start 启动后台服务
func (a *App) Start() {
	a.Task.Start()
	a.GC.Start()
	log.Info("app started")
}

// Close 关闭所有资源
func (a *App) Close() error {
	log.Info("closing app...")

	// 按初始化的逆序关闭
	if a.Batch != nil {
		a.Batch.Shutdown()
	}

	if a.RedisQueue != nil {
		a.RedisQueue.Close()
	}

	if a.batchStore != nil {
		a.batchStore.Close()
	}

	if a.GC != nil {
		a.GC.Stop()
	}

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
