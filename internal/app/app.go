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
	"github.com/tmalldedede/agentbox/internal/channel"
	"github.com/tmalldedede/agentbox/internal/channel/feishu"
	"github.com/tmalldedede/agentbox/internal/config"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/coordinate"
	"github.com/tmalldedede/agentbox/internal/cron"
	"github.com/tmalldedede/agentbox/internal/database"
	"github.com/tmalldedede/agentbox/internal/engine"
	_ "github.com/tmalldedede/agentbox/internal/engine/claude"   // 注册 Claude Code 适配器
	_ "github.com/tmalldedede/agentbox/internal/engine/codex"    // 注册 Codex 适配器
	_ "github.com/tmalldedede/agentbox/internal/engine/opencode" // 注册 OpenCode 适配器
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/plugin"
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

	// 多通道支持 (Phase 2)
	Channel       *channel.Manager
	FeishuChannel *feishu.Channel

	// 定时任务 (Phase 1)
	Cron *cron.Manager

	// 插件系统 (Phase 1)
	Plugin *plugin.Manager

	// 跨会话协调 (Phase 2)
	Coordinate *coordinate.Manager

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

	// 16. 初始化 Plugin Manager (Phase 1)
	a.Plugin = plugin.NewManager()
	log.Info("plugin manager initialized")

	// 16.5. 初始化 Coordinate Manager (Phase 2)
	a.Coordinate = coordinate.NewManager(a.Task)
	log.Info("coordinate manager initialized")

	// 17. 初始化 Cron Manager (Phase 1)
	cronStore := cron.NewDBStore()
	a.Cron = cron.NewManager(cronStore, a.cronJobExecutor)
	log.Info("cron manager initialized")

	// 18. 初始化 Channel Manager (Phase 2)
	a.Channel = channel.NewManager()

	// 尝试加载飞书配置
	feishuStore := feishu.NewStore()
	if feishuCfg, err := feishuStore.GetEnabledConfig(); err == nil {
		a.FeishuChannel = feishu.New(feishuCfg)
		if err := a.Channel.Register(a.FeishuChannel); err != nil {
			log.Warn("register feishu channel failed", "error", err)
		} else {
			log.Info("feishu channel registered")
		}
	} else {
		log.Info("feishu channel not configured")
	}

	// 添加消息处理器（将消息转发到 Task API）
	a.Channel.AddHandler(a.channelMessageHandler)
	log.Info("channel manager initialized")

	return nil
}

// Start 启动后台服务
func (a *App) Start() {
	a.Task.Start()
	a.GC.Start()

	// 启动 Cron
	if a.Cron != nil {
		if err := a.Cron.Start(context.Background()); err != nil {
			log.Error("start cron failed", "error", err)
		}
	}

	// 启动 Channel
	if a.Channel != nil {
		if err := a.Channel.Start(context.Background()); err != nil {
			log.Error("start channel failed", "error", err)
		}
	}

	log.Info("app started")
}

// Close 关闭所有资源
func (a *App) Close() error {
	log.Info("closing app...")

	// 按初始化的逆序关闭
	if a.Channel != nil {
		a.Channel.Stop()
	}

	if a.Cron != nil {
		a.Cron.Stop()
	}

	if a.Plugin != nil {
		a.Plugin.Shutdown()
	}

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

// cronJobExecutor Cron 任务执行器
func (a *App) cronJobExecutor(ctx context.Context, job *cron.Job) error {
	log.Info("executing cron job", "id", job.ID, "name", job.Name, "agent_id", job.AgentID)

	// 创建 Task
	taskReq := &task.CreateTaskRequest{
		AgentID: job.AgentID,
		Prompt:  job.Prompt,
	}

	t, err := a.Task.CreateTask(taskReq)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	log.Info("cron job task created", "cron_id", job.ID, "task_id", t.ID)
	return nil
}

// channelMessageHandler 通道消息处理器
func (a *App) channelMessageHandler(ctx context.Context, msg *channel.Message) error {
	log.Info("received channel message",
		"channel", msg.ChannelType,
		"chat_id", msg.ChannelID,
		"sender", msg.SenderID,
		"content", msg.Content,
	)

	// 获取 Agent ID（优先使用通道配置的默认 Agent）
	agentID := a.getAgentForChannel(msg.ChannelType, msg.ChannelID)
	if agentID == "" {
		log.Warn("no agent configured for channel, ignoring message",
			"channel", msg.ChannelType,
			"chat_id", msg.ChannelID,
		)
		return nil
	}

	// 保存消息日志（仅飞书）
	if msg.ChannelType == "feishu" {
		a.saveFeishuMessageLog(msg, "")
	}

	// 创建 Task
	taskReq := &task.CreateTaskRequest{
		AgentID: agentID,
		Prompt:  msg.Content,
		Metadata: map[string]string{
			"channel_type": msg.ChannelType,
			"channel_id":   msg.ChannelID,
			"message_id":   msg.ID,
			"sender_id":    msg.SenderID,
		},
	}

	t, err := a.Task.CreateTask(taskReq)
	if err != nil {
		log.Error("create task from channel message failed", "error", err)
		// 发送错误提示到通道
		a.sendChannelReply(msg.ChannelType, msg.ChannelID, msg.ID, "❌ 任务创建失败，请稍后重试")
		return fmt.Errorf("create task: %w", err)
	}

	log.Info("channel message task created", "message_id", msg.ID, "task_id", t.ID)

	// 更新消息日志关联 Task ID
	if msg.ChannelType == "feishu" {
		feishuStore := feishu.NewStore()
		feishuStore.UpdateMessageTaskID(msg.ID, t.ID)
	}

	// 异步等待 Task 完成并回复
	go a.waitAndReplyChannel(msg.ChannelType, msg.ChannelID, msg.ID, t.ID)

	return nil
}

// getAgentForChannel 获取通道对应的 Agent ID
func (a *App) getAgentForChannel(channelType, channelID string) string {
	// 飞书：使用配置的 default_agent_id
	if channelType == "feishu" && a.FeishuChannel != nil {
		cfg := a.FeishuChannel.GetConfig()
		if cfg != nil && cfg.DefaultAgentID != "" {
			return cfg.DefaultAgentID
		}
	}

	// 后备：使用第一个可用的 Agent
	agents := a.Agent.List()
	if len(agents) > 0 {
		return agents[0].ID
	}

	return ""
}

// saveFeishuMessageLog 保存飞书消息日志
func (a *App) saveFeishuMessageLog(msg *channel.Message, taskID string) {
	store := feishu.NewStore()
	msgLog := &feishu.MessageLog{
		ID:          msg.ID,
		ChatID:      msg.ChannelID,
		SenderID:    msg.SenderID,
		SenderName:  msg.SenderName,
		Content:     msg.Content,
		MessageType: msg.Metadata["message_type"],
		ReplyID:     msg.ReplyTo,
		TaskID:      taskID,
		ReceivedAt:  msg.ReceivedAt,
	}
	if err := store.SaveMessageLog(msgLog); err != nil {
		log.Warn("save feishu message log failed", "error", err)
	}
}

// waitAndReplyChannel 等待 Task 完成并回复通道
func (a *App) waitAndReplyChannel(channelType, channelID, replyTo, taskID string) {
	// 订阅任务事件
	eventCh := a.Task.SubscribeEvents(taskID)
	defer a.Task.UnsubscribeEvents(taskID, eventCh)

	// 设置超时（10 分钟）
	timeout := time.After(10 * time.Minute)

	var result string
	var completed bool

	for !completed {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// 通道关闭
				completed = true
				break
			}

			switch event.Type {
			case "task.completed":
				// 获取任务结果
				t, err := a.Task.GetTask(taskID)
				if err != nil {
					log.Error("get completed task failed", "task_id", taskID, "error", err)
					result = "❌ 获取结果失败"
				} else if t.Result != nil && t.Result.Text != "" {
					result = t.Result.Text
				} else {
					result = "✅ 任务已完成"
				}
				completed = true

			case "task.failed":
				// 获取错误信息
				t, err := a.Task.GetTask(taskID)
				if err != nil {
					result = "❌ 任务执行失败"
				} else if t.ErrorMessage != "" {
					result = fmt.Sprintf("❌ 任务失败: %s", t.ErrorMessage)
				} else {
					result = "❌ 任务执行失败"
				}
				completed = true

			case "task.cancelled":
				result = "⚠️ 任务已取消"
				completed = true
			}

		case <-timeout:
			log.Warn("task timeout waiting for completion", "task_id", taskID)
			result = "⏱️ 任务执行超时，请稍后查看结果"
			completed = true
		}
	}

	// 发送回复
	if result != "" {
		a.sendChannelReply(channelType, channelID, replyTo, result)
	}
}

// sendChannelReply 发送通道回复
func (a *App) sendChannelReply(channelType, channelID, replyTo, content string) {
	if a.Channel == nil {
		return
	}

	// 限制回复长度（飞书单条消息限制 4000 字符）
	const maxLength = 3800
	if len(content) > maxLength {
		content = content[:maxLength] + "\n\n... (内容过长已截断)"
	}

	resp, err := a.Channel.Send(context.Background(), channelType, &channel.SendRequest{
		ChannelID: channelID,
		Content:   content,
		ReplyTo:   replyTo,
	})
	if err != nil {
		log.Error("send channel reply failed",
			"channel", channelType,
			"chat_id", channelID,
			"error", err,
		)
		return
	}

	log.Info("channel reply sent",
		"channel", channelType,
		"chat_id", channelID,
		"message_id", resp.MessageID,
	)
}
