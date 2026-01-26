package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler Gateway 的 Gin 处理器封装
type Handler struct {
	gateway   *Gateway
	authFunc  AuthFunc
}

// NewHandler 创建新的 Handler
func NewHandler(gw *Gateway, authFunc AuthFunc) *Handler {
	gw.SetAuthFunc(authFunc)
	return &Handler{
		gateway:  gw,
		authFunc: authFunc,
	}
}

// HandleWebSocket 处理 WebSocket 连接
// GET /api/v1/gateway/ws
func (h *Handler) HandleWebSocket(c *gin.Context) {
	h.gateway.HandleConnection(c.Writer, c.Request)
}

// GetStats 获取 Gateway 统计信息
// GET /api/v1/gateway/stats
func (h *Handler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connected_clients": h.gateway.GetClientCount(),
		"subscriptions":     h.getSubscriptionStats(),
	})
}

// getSubscriptionStats 获取订阅统计
func (h *Handler) getSubscriptionStats() map[string]int {
	h.gateway.subMu.RLock()
	defer h.gateway.subMu.RUnlock()

	stats := make(map[string]int)
	for channel, topics := range h.gateway.subscriptions {
		count := 0
		for _, clients := range topics {
			count += len(clients)
		}
		stats[channel] = count
	}
	return stats
}

// RegisterRoutes 注册路由到 Gin Router
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	gw := router.Group("/gateway")
	{
		gw.GET("/ws", h.HandleWebSocket)
		gw.GET("/stats", h.GetStats)
	}
}

// RegisterPublicRoutes 注册公开路由（无需认证）
func (h *Handler) RegisterPublicRoutes(router *gin.RouterGroup) {
	// WebSocket 连接通过消息认证，不需要 HTTP 层认证
	router.GET("/gateway/ws", h.HandleWebSocket)
}
