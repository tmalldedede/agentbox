package gateway

import (
	"context"
	"log"
	"sync"
)

// TaskEventBridge 任务事件桥接器
// 将 Task Manager 的事件转发到 Gateway
type TaskEventBridge struct {
	gateway     *Gateway
	taskManager TaskManager
	cancelFuncs map[string]context.CancelFunc
	mu          sync.RWMutex
}

// TaskManager 任务管理器接口
type TaskManager interface {
	// SubscribeEvents 订阅任务事件
	SubscribeEvents(taskID string) <-chan *TaskEvent
	// UnsubscribeEvents 取消订阅任务事件
	UnsubscribeEvents(taskID string, ch <-chan *TaskEvent)
	// CancelTask 取消任务
	CancelTask(taskID string) error
	// CreateTask 创建任务（如果带 TaskID 则追加轮次）
	CreateTask(req *CreateTaskRequest) (*Task, error)
}

// CreateTaskRequest 创建任务请求（简化版）
type CreateTaskRequest struct {
	TaskID string // 如果指定则追加轮次
	Prompt string
}

// Task 任务（简化版）
type Task struct {
	ID string
}

// TaskEvent 任务事件（与 task.TaskEvent 对应）
type TaskEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// NewTaskEventBridge 创建任务事件桥接器
func NewTaskEventBridge(gw *Gateway, tm TaskManager) *TaskEventBridge {
	bridge := &TaskEventBridge{
		gateway:     gw,
		taskManager: tm,
		cancelFuncs: make(map[string]context.CancelFunc),
	}

	// 设置任务操作回调
	gw.SetTaskActionFunc(bridge.handleTaskAction)

	return bridge
}

// handleTaskAction 处理任务操作
func (b *TaskEventBridge) handleTaskAction(taskID, action, data string) error {
	switch action {
	case "cancel":
		return b.taskManager.CancelTask(taskID)
	case "append_turn":
		_, err := b.taskManager.CreateTask(&CreateTaskRequest{
			TaskID: taskID,
			Prompt: data,
		})
		return err
	default:
		return nil
	}
}

// StartForwarding 开始转发任务事件
func (b *TaskEventBridge) StartForwarding(taskID string) {
	eventCh := b.taskManager.SubscribeEvents(taskID)

	ctx, cancel := context.WithCancel(context.Background())

	b.mu.Lock()
	b.cancelFuncs[taskID] = cancel
	b.mu.Unlock()

	go func() {
		defer func() {
			b.taskManager.UnsubscribeEvents(taskID, eventCh)
			b.mu.Lock()
			delete(b.cancelFuncs, taskID)
			b.mu.Unlock()
		}()

		for {
			select {
			case event, ok := <-eventCh:
				if !ok {
					return
				}
				// 广播到 Gateway
				b.gateway.BroadcastEvent(ChannelTask, taskID, event.Type, event.Data)

				// 如果任务完成或失败，停止转发
				if event.Type == "task.completed" || event.Type == "task.failed" || event.Type == "task.cancelled" {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopForwarding 停止转发任务事件
func (b *TaskEventBridge) StopForwarding(taskID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if cancel, ok := b.cancelFuncs[taskID]; ok {
		cancel()
		delete(b.cancelFuncs, taskID)
	}
}

// SessionEventBridge 会话事件桥接器
type SessionEventBridge struct {
	gateway        *Gateway
	sessionManager SessionManager
}

// SessionManager 会话管理器接口
type SessionManager interface {
	// Execute 执行命令
	Execute(sessionID, command string) (execID string, outputCh <-chan string, err error)
}

// NewSessionEventBridge 创建会话事件桥接器
func NewSessionEventBridge(gw *Gateway, sm SessionManager) *SessionEventBridge {
	bridge := &SessionEventBridge{
		gateway:        gw,
		sessionManager: sm,
	}

	// 设置会话执行回调
	gw.SetSessionExecFunc(bridge.handleSessionExec)

	return bridge
}

// handleSessionExec 处理会话执行
func (b *SessionEventBridge) handleSessionExec(sessionID, command string) (string, error) {
	execID, outputCh, err := b.sessionManager.Execute(sessionID, command)
	if err != nil {
		return "", err
	}

	// 启动输出转发
	go func() {
		for output := range outputCh {
			b.gateway.BroadcastEvent(ChannelSession, sessionID, "output", SessionOutputPayload{
				SessionID: sessionID,
				ExecID:    execID,
				Content:   output,
				Type:      "message",
			})
		}
		// 发送完成事件
		b.gateway.BroadcastEvent(ChannelSession, sessionID, "done", SessionOutputPayload{
			SessionID: sessionID,
			ExecID:    execID,
			Type:      "done",
		})
	}()

	return execID, nil
}

// SystemEventPublisher 系统事件发布器
type SystemEventPublisher struct {
	gateway *Gateway
}

// NewSystemEventPublisher 创建系统事件发布器
func NewSystemEventPublisher(gw *Gateway) *SystemEventPublisher {
	return &SystemEventPublisher{gateway: gw}
}

// PublishTaskCreated 发布任务创建事件
func (p *SystemEventPublisher) PublishTaskCreated(taskID string, data interface{}) {
	p.gateway.BroadcastEvent(ChannelSystem, "tasks", "task.created", map[string]interface{}{
		"task_id": taskID,
		"data":    data,
	})
	log.Printf("[Gateway] Published task.created event for %s", taskID)
}

// PublishAgentCreated 发布 Agent 创建事件
func (p *SystemEventPublisher) PublishAgentCreated(agentID string, data interface{}) {
	p.gateway.BroadcastEvent(ChannelSystem, "agents", "agent.created", map[string]interface{}{
		"agent_id": agentID,
		"data":     data,
	})
}

// PublishSessionCreated 发布会话创建事件
func (p *SystemEventPublisher) PublishSessionCreated(sessionID string, data interface{}) {
	p.gateway.BroadcastEvent(ChannelSystem, "sessions", "session.created", map[string]interface{}{
		"session_id": sessionID,
		"data":       data,
	})
}

// PublishBatchCreated 发布批处理创建事件
func (p *SystemEventPublisher) PublishBatchCreated(batchID string, data interface{}) {
	p.gateway.BroadcastEvent(ChannelSystem, "batches", "batch.created", map[string]interface{}{
		"batch_id": batchID,
		"data":     data,
	})
}

// PublishAlert 发布系统告警
func (p *SystemEventPublisher) PublishAlert(level, message string, data interface{}) {
	p.gateway.BroadcastEvent(ChannelSystem, "alerts", "alert", map[string]interface{}{
		"level":   level,
		"message": message,
		"data":    data,
	})
}
