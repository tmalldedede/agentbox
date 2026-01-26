package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/auth"
	"github.com/tmalldedede/agentbox/internal/gateway"
	"github.com/tmalldedede/agentbox/internal/task"
)

// GatewayHandler Gateway 处理器
type GatewayHandler struct {
	gateway     *gateway.Gateway
	taskBridge  *gateway.TaskEventBridge
	sysPublisher *gateway.SystemEventPublisher
	authManager *auth.Manager
}

// NewGatewayHandler 创建 Gateway 处理器
func NewGatewayHandler(authManager *auth.Manager, taskManager *task.Manager) *GatewayHandler {
	gw := gateway.NewGateway()

	// 设置认证函数
	gw.SetAuthFunc(func(token string) (string, error) {
		user, err := authManager.ValidateToken(token)
		if err != nil {
			return "", err
		}
		if user == nil {
			return "", errors.New("invalid token")
		}
		return user.ID, nil
	})

	// 创建任务事件桥接器
	taskBridge := gateway.NewTaskEventBridge(gw, &taskManagerAdapter{tm: taskManager})

	// 创建系统事件发布器
	sysPublisher := gateway.NewSystemEventPublisher(gw)

	// 启动 Gateway
	gw.Start()

	return &GatewayHandler{
		gateway:      gw,
		taskBridge:   taskBridge,
		sysPublisher: sysPublisher,
		authManager:  authManager,
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *GatewayHandler) HandleWebSocket(c *gin.Context) {
	h.gateway.HandleConnection(c.Writer, c.Request)
}

// GetStats 获取 Gateway 统计信息
func (h *GatewayHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connected_clients": h.gateway.GetClientCount(),
	})
}

// RegisterRoutes 注册路由
func (h *GatewayHandler) RegisterRoutes(router *gin.RouterGroup) {
	gw := router.Group("/gateway")
	{
		gw.GET("/ws", h.HandleWebSocket)
		gw.GET("/stats", h.GetStats)
	}
}

// RegisterPublicRoutes 注册公开路由（无需认证，WebSocket 通过消息认证）
func (h *GatewayHandler) RegisterPublicRoutes(router *gin.RouterGroup) {
	router.GET("/gateway/ws", h.HandleWebSocket)
}

// StartTaskForwarding 开始转发任务事件
func (h *GatewayHandler) StartTaskForwarding(taskID string) {
	h.taskBridge.StartForwarding(taskID)
}

// StopTaskForwarding 停止转发任务事件
func (h *GatewayHandler) StopTaskForwarding(taskID string) {
	h.taskBridge.StopForwarding(taskID)
}

// PublishTaskCreated 发布任务创建事件
func (h *GatewayHandler) PublishTaskCreated(taskID string, data interface{}) {
	h.sysPublisher.PublishTaskCreated(taskID, data)
}

// Stop 停止 Gateway
func (h *GatewayHandler) Stop() {
	h.gateway.Stop()
}

// Gateway 获取 Gateway 实例
func (h *GatewayHandler) Gateway() *gateway.Gateway {
	return h.gateway
}

// taskManagerAdapter 适配 task.Manager 到 gateway.TaskManager 接口
type taskManagerAdapter struct {
	tm *task.Manager
}

func (a *taskManagerAdapter) SubscribeEvents(taskID string) <-chan *gateway.TaskEvent {
	ch := a.tm.SubscribeEvents(taskID)
	// 转换 channel 类型
	out := make(chan *gateway.TaskEvent, 100)
	go func() {
		defer close(out)
		for event := range ch {
			out <- &gateway.TaskEvent{
				Type: event.Type,
				Data: event.Data,
			}
		}
	}()
	return out
}

func (a *taskManagerAdapter) UnsubscribeEvents(taskID string, ch <-chan *gateway.TaskEvent) {
	// 注意：这里无法直接取消订阅，因为我们包装了 channel
	// 实际实现需要维护映射关系
}

func (a *taskManagerAdapter) CancelTask(taskID string) error {
	return a.tm.CancelTask(taskID)
}

func (a *taskManagerAdapter) CreateTask(req *gateway.CreateTaskRequest) (*gateway.Task, error) {
	t, err := a.tm.CreateTask(&task.CreateTaskRequest{
		TaskID: req.TaskID,
		Prompt: req.Prompt,
	})
	if err != nil {
		return nil, err
	}
	return &gateway.Task{ID: t.ID}, nil
}
