package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/session"
)

// Server HTTP 服务器
type Server struct {
	engine  *gin.Engine
	handler *Handler
}

// NewServer 创建服务器
func NewServer(sessionMgr *session.Manager, registry *agent.Registry) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(loggerMiddleware())

	handler := NewHandler(sessionMgr, registry)

	s := &Server{
		engine:  engine,
		handler: handler,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.engine.GET("/health", s.handler.HealthCheck)

	// API v1
	v1 := s.engine.Group("/api/v1")
	{
		// Agent
		v1.GET("/agents", s.handler.ListAgents)

		// Sessions
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
		}
	}

	// 兼容简短路由
	s.engine.GET("/api/health", s.handler.HealthCheck)
	s.engine.GET("/api/agents", s.handler.ListAgents)
	s.engine.POST("/api/sessions", s.handler.CreateSession)
	s.engine.GET("/api/sessions", s.handler.ListSessions)
	s.engine.GET("/api/sessions/:id", s.handler.GetSession)
	s.engine.DELETE("/api/sessions/:id", s.handler.DeleteSession)
	s.engine.POST("/api/sessions/:id/exec", s.handler.ExecSession)
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

// Engine 获取 Gin 引擎 (用于测试)
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
