package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"github.com/tmalldedede/agentbox/internal/smartagent"
	"github.com/tmalldedede/agentbox/internal/task"
	"github.com/tmalldedede/agentbox/internal/webhook"

	_ "github.com/tmalldedede/agentbox/docs" // swagger docs
)

// Server HTTP 服务器
type Server struct {
	httpServer         *http.Server
	engine             *gin.Engine
	handler            *Handler
	fileHandler        *FileHandler
	publicFileHandler  *PublicFileHandler
	wsHandler          *WSHandler
	profileHandler     *ProfileHandler
	providerHandler    *ProviderHandler
	mcpHandler         *MCPHandler
	skillHandler       *SkillHandler
	credentialHandler  *CredentialHandler
	imageHandler       *ImageHandler
	systemHandler      *SystemHandler
	taskHandler        *TaskHandler
	webhookHandler     *WebhookHandler
	smartAgentHandler  *SmartAgentHandler
	historyHandler     *HistoryHandler
}

// Deps 服务器依赖（从 App 容器注入）
type Deps struct {
	Session    *session.Manager
	Registry   *agent.Registry
	Container  container.Manager
	Profile    *profile.Manager
	Provider   *provider.Manager
	MCP        *mcp.Manager
	Skill      *skill.Manager
	Credential *credential.Manager
	Task       *task.Manager
	Webhook    *webhook.Manager
	SmartAgent *smartagent.Manager
	History    *history.Manager
}

// NewServer 创建服务器
func NewServer(deps *Deps) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(loggerMiddleware())

	handler := NewHandler(deps.Session, deps.Registry)
	fileHandler := NewFileHandler(deps.Session)
	publicFileHandler := NewPublicFileHandler()
	wsHandler := NewWSHandler(deps.Session, deps.Registry, deps.Container)
	profileHandler := NewProfileHandler(deps.Profile)
	providerHandler := NewProviderHandler(deps.Provider)
	mcpHandler := NewMCPHandler(deps.MCP)
	skillHandler := NewSkillHandler(deps.Skill)
	credentialHandler := NewCredentialHandler(deps.Credential)
	imageHandler := NewImageHandler(deps.Container)
	systemHandler := NewSystemHandler(deps.Container, deps.Session)
	taskHandler := NewTaskHandler(deps.Task)
	webhookHandler := NewWebhookHandler(deps.Webhook)
	smartAgentHandler := NewSmartAgentHandler(deps.SmartAgent, deps.Session, deps.Profile, deps.History)
	historyHandler := NewHistoryHandler(deps.History)

	s := &Server{
		engine:             engine,
		handler:            handler,
		fileHandler:        fileHandler,
		publicFileHandler:  publicFileHandler,
		wsHandler:          wsHandler,
		profileHandler:     profileHandler,
		providerHandler:    providerHandler,
		mcpHandler:         mcpHandler,
		skillHandler:       skillHandler,
		credentialHandler:  credentialHandler,
		imageHandler:       imageHandler,
		systemHandler:      systemHandler,
		taskHandler:        taskHandler,
		webhookHandler:     webhookHandler,
		smartAgentHandler:  smartAgentHandler,
		historyHandler:     historyHandler,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// ==================== Swagger API Docs ====================
	// GET /swagger/*any -> Swagger UI
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==================== Public API ====================
	// 参考 Manus API 设计，对外开放的 API
	v1 := s.engine.Group("/api/v1")
	{
		// Health
		v1.GET("/health", s.handler.HealthCheck)

		// Engines (只读) - 底层引擎适配器列表 (claude-code, codex, opencode)
		v1.GET("/engines", s.handler.ListAgents)

		// Agents (CRUD + Run) - 对外暴露的智能体 API
		s.smartAgentHandler.RegisterRoutes(v1)

		// Profiles (CRUD) - Agent 配置模板
		s.profileHandler.RegisterRoutes(v1)

		// Providers (只读 + 自定义) - API 提供商配置
		s.providerHandler.RegisterRoutes(v1)

		// Sessions (CRUD) - 工作空间/容器 (类似 Manus Projects)
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", s.handler.CreateSession)
			sessions.GET("", s.handler.ListSessions)
			sessions.GET("/:id", s.handler.GetSession)
			sessions.DELETE("/:id", s.handler.DeleteSession)
			sessions.POST("/:id/start", s.handler.StartSession)
			sessions.POST("/:id/stop", s.handler.StopSession)
			sessions.POST("/:id/exec", s.handler.ExecSession)
			sessions.POST("/:id/exec/stream", s.handler.ExecSessionStream) // SSE 流式执行 (Codex)
			sessions.GET("/:id/executions", s.handler.GetExecutions)
			sessions.GET("/:id/executions/:execId", s.handler.GetExecution)
			sessions.GET("/:id/logs", s.handler.GetSessionLogs)
			sessions.GET("/:id/logs/stream", s.handler.StreamSessionLogs) // SSE 实时日志流

			// Session 文件管理
			sessions.GET("/:id/files", s.fileHandler.ListFiles)
			sessions.GET("/:id/files/*path", s.fileHandler.DownloadFile)
			sessions.POST("/:id/files", s.fileHandler.UploadFile)
			sessions.DELETE("/:id/files/*path", s.fileHandler.DeleteFile)
			sessions.POST("/:id/directories", s.fileHandler.CreateDirectory)

			// WebSocket 流式执行
			sessions.GET("/:id/stream", s.wsHandler.ExecStream)
		}

		// Tasks (CRUD) - 异步任务队列
		s.taskHandler.RegisterRoutes(v1)

		// Files (CRUD) - 独立文件上传 (任务附件)
		s.publicFileHandler.RegisterRoutes(v1)

		// Webhooks (CRUD) - Webhook 管理
		s.webhookHandler.RegisterRoutes(v1)

		// History (只读) - 执行历史记录
		s.historyHandler.RegisterRoutes(v1)
	}

	// ==================== Admin API ====================
	// 平台管理接口，需要额外权限
	admin := s.engine.Group("/api/v1/admin")
	{
		// MCP Servers 管理
		s.mcpHandler.RegisterRoutes(admin)

		// Skills 管理
		s.skillHandler.RegisterRoutes(admin)

		// Credentials 管理
		s.credentialHandler.RegisterRoutes(admin)

		// Images 管理
		s.imageHandler.RegisterRoutes(admin)

		// System 管理
		s.systemHandler.RegisterRoutes(admin)
	}
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Engine 获取 Gin 引擎 (用于测试)
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
