package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/credential"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/profile"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"github.com/tmalldedede/agentbox/internal/task"
	"github.com/tmalldedede/agentbox/internal/webhook"

	_ "github.com/tmalldedede/agentbox/docs" // swagger docs
)

// Server HTTP 服务器
type Server struct {
	engine            *gin.Engine
	handler           *Handler
	fileHandler       *FileHandler
	publicFileHandler *PublicFileHandler
	wsHandler         *WSHandler
	profileHandler    *ProfileHandler
	providerHandler   *ProviderHandler
	mcpHandler        *MCPHandler
	skillHandler      *SkillHandler
	credentialHandler *CredentialHandler
	imageHandler      *ImageHandler
	systemHandler     *SystemHandler
	taskHandler       *TaskHandler
	webhookHandler    *WebhookHandler
}

// NewServer 创建服务器
func NewServer(sessionMgr *session.Manager, registry *agent.Registry, containerMgr container.Manager, profileMgr *profile.Manager, providerMgr *provider.Manager, mcpMgr *mcp.Manager, skillMgr *skill.Manager, credentialMgr *credential.Manager, taskMgr *task.Manager, webhookMgr *webhook.Manager) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(loggerMiddleware())

	handler := NewHandler(sessionMgr, registry)
	fileHandler := NewFileHandler(sessionMgr)
	publicFileHandler := NewPublicFileHandler()
	wsHandler := NewWSHandler(sessionMgr, registry, containerMgr)
	profileHandler := NewProfileHandler(profileMgr)
	providerHandler := NewProviderHandler(providerMgr)
	mcpHandler := NewMCPHandler(mcpMgr)
	skillHandler := NewSkillHandler(skillMgr)
	credentialHandler := NewCredentialHandler(credentialMgr)
	imageHandler := NewImageHandler(containerMgr)
	systemHandler := NewSystemHandler(containerMgr, sessionMgr)
	taskHandler := NewTaskHandler(taskMgr)
	webhookHandler := NewWebhookHandler(webhookMgr)

	s := &Server{
		engine:            engine,
		handler:           handler,
		fileHandler:       fileHandler,
		publicFileHandler: publicFileHandler,
		wsHandler:         wsHandler,
		profileHandler:    profileHandler,
		providerHandler:   providerHandler,
		mcpHandler:        mcpHandler,
		skillHandler:      skillHandler,
		credentialHandler: credentialHandler,
		imageHandler:      imageHandler,
		systemHandler:     systemHandler,
		taskHandler:       taskHandler,
		webhookHandler:    webhookHandler,
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

		// Agents (只读)
		v1.GET("/agents", s.handler.ListAgents)

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
			sessions.GET("/:id/executions", s.handler.GetExecutions)
			sessions.GET("/:id/executions/:execId", s.handler.GetExecution)
			sessions.GET("/:id/logs", s.handler.GetSessionLogs)

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
	return s.engine.Run(addr)
}

// Engine 获取 Gin 引擎 (用于测试)
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
