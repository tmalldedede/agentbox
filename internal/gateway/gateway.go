package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 生产环境应该验证 Origin
	},
}

// AuthFunc 认证函数类型
type AuthFunc func(token string) (userID string, err error)

// TaskActionFunc 任务操作函数类型
type TaskActionFunc func(taskID, action, data string) error

// SessionExecFunc 会话执行函数类型
type SessionExecFunc func(sessionID, command string) (execID string, err error)

// Gateway WebSocket 网关
type Gateway struct {
	clients    map[string]*Client // clientID -> Client
	users      map[string]map[string]*Client // userID -> deviceID -> Client
	mu         sync.RWMutex

	// 事件订阅管理
	// channel -> topic -> clients
	subscriptions map[string]map[string]map[string]*Client
	subMu         sync.RWMutex

	// 回调函数
	authFunc       AuthFunc
	taskActionFunc TaskActionFunc
	sessionExecFunc SessionExecFunc

	// 控制
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
}

// NewGateway 创建新的网关
func NewGateway() *Gateway {
	return &Gateway{
		clients:       make(map[string]*Client),
		users:         make(map[string]map[string]*Client),
		subscriptions: make(map[string]map[string]map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		done:          make(chan struct{}),
	}
}

// SetAuthFunc 设置认证函数
func (g *Gateway) SetAuthFunc(f AuthFunc) {
	g.authFunc = f
}

// SetTaskActionFunc 设置任务操作函数
func (g *Gateway) SetTaskActionFunc(f TaskActionFunc) {
	g.taskActionFunc = f
}

// SetSessionExecFunc 设置会话执行函数
func (g *Gateway) SetSessionExecFunc(f SessionExecFunc) {
	g.sessionExecFunc = f
}

// Start 启动网关
func (g *Gateway) Start() {
	go g.run()
}

// Stop 停止网关
func (g *Gateway) Stop() {
	close(g.done)
}

// run 网关主循环
func (g *Gateway) run() {
	for {
		select {
		case client := <-g.register:
			g.mu.Lock()
			g.clients[client.ID] = client
			g.mu.Unlock()
			log.Printf("[Gateway] Client connected: %s", client.ID)

		case client := <-g.unregister:
			g.mu.Lock()
			if _, ok := g.clients[client.ID]; ok {
				delete(g.clients, client.ID)

				// 从用户设备映射中移除
				if client.UserID != "" {
					if devices, ok := g.users[client.UserID]; ok {
						delete(devices, client.DeviceID)
						if len(devices) == 0 {
							delete(g.users, client.UserID)
						}
					}
				}
			}
			g.mu.Unlock()

			// 从订阅中移除
			g.removeClientFromSubscriptions(client)
			log.Printf("[Gateway] Client disconnected: %s", client.ID)

		case <-g.done:
			// 关闭所有客户端
			g.mu.Lock()
			for _, client := range g.clients {
				client.Close()
			}
			g.mu.Unlock()
			return
		}
	}
}

// Unregister 注销客户端
func (g *Gateway) Unregister(client *Client) {
	select {
	case g.unregister <- client:
	case <-g.done:
	}
}

// removeClientFromSubscriptions 从所有订阅中移除客户端
func (g *Gateway) removeClientFromSubscriptions(client *Client) {
	g.subMu.Lock()
	defer g.subMu.Unlock()

	for channel, topics := range g.subscriptions {
		for topic, clients := range topics {
			delete(clients, client.ID)
			if len(clients) == 0 {
				delete(g.subscriptions[channel], topic)
			}
		}
		if len(g.subscriptions[channel]) == 0 {
			delete(g.subscriptions, channel)
		}
	}
}

// HandleConnection HTTP 升级为 WebSocket
func (g *Gateway) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Gateway] Upgrade error: %v", err)
		return
	}

	clientID := uuid.New().String()
	client := NewClient(clientID, conn, g)

	g.register <- client

	go client.WritePump()
	go client.ReadPump()
}

// HandleMessage 处理客户端消息
func (g *Gateway) HandleMessage(client *Client, msg *Message) {
	switch msg.Type {
	case MsgTypeAuth:
		g.handleAuth(client, msg)
	case MsgTypeSubscribe:
		g.handleSubscribe(client, msg)
	case MsgTypeUnsubscribe:
		g.handleUnsubscribe(client, msg)
	case MsgTypePing:
		g.handlePing(client, msg)
	case MsgTypeTaskAction:
		g.handleTaskAction(client, msg)
	case MsgTypeSessionExec:
		g.handleSessionExec(client, msg)
	default:
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "UNKNOWN_TYPE",
			Message: "Unknown message type: " + string(msg.Type),
		}))
	}
}

// handleAuth 处理认证
func (g *Gateway) handleAuth(client *Client, msg *Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload AuthPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeAuthResult, AuthResult{
			Success: false,
			Message: "Invalid payload",
		}))
		return
	}

	if g.authFunc == nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeAuthResult, AuthResult{
			Success: false,
			Message: "Auth not configured",
		}))
		return
	}

	userID, err := g.authFunc(payload.Token)
	if err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeAuthResult, AuthResult{
			Success: false,
			Message: err.Error(),
		}))
		return
	}

	client.UserID = userID
	client.DeviceID = payload.DeviceID
	if client.DeviceID == "" {
		client.DeviceID = "default"
	}

	// 更新用户设备映射
	g.mu.Lock()
	if g.users[userID] == nil {
		g.users[userID] = make(map[string]*Client)
	}
	// 如果同一设备已存在连接，关闭旧连接
	if old, ok := g.users[userID][client.DeviceID]; ok && old.ID != client.ID {
		old.Close()
	}
	g.users[userID][client.DeviceID] = client
	g.mu.Unlock()

	client.SendMessage(NewMessageWithID(msg.ID, MsgTypeAuthResult, AuthResult{
		Success:  true,
		UserID:   userID,
		DeviceID: client.DeviceID,
	}))
}

// handleSubscribe 处理订阅
func (g *Gateway) handleSubscribe(client *Client, msg *Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SubscribePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "INVALID_PAYLOAD",
			Message: "Invalid subscribe payload",
		}))
		return
	}

	// 验证频道
	if payload.Channel != ChannelTask && payload.Channel != ChannelSession && payload.Channel != ChannelSystem {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "INVALID_CHANNEL",
			Message: "Invalid channel: " + payload.Channel,
		}))
		return
	}

	// 更新客户端订阅
	client.Subscribe(payload.Channel, payload.Topics)

	// 更新全局订阅映射
	g.subMu.Lock()
	if g.subscriptions[payload.Channel] == nil {
		g.subscriptions[payload.Channel] = make(map[string]map[string]*Client)
	}

	topics := payload.Topics
	if len(topics) == 0 {
		topics = []string{"*"}
	}

	for _, topic := range topics {
		if g.subscriptions[payload.Channel][topic] == nil {
			g.subscriptions[payload.Channel][topic] = make(map[string]*Client)
		}
		g.subscriptions[payload.Channel][topic][client.ID] = client
	}
	g.subMu.Unlock()

	client.SendMessage(NewMessageWithID(msg.ID, MsgTypeSubscribed, SubscribeResult{
		Channel: payload.Channel,
		Topics:  payload.Topics,
		Success: true,
	}))
}

// handleUnsubscribe 处理取消订阅
func (g *Gateway) handleUnsubscribe(client *Client, msg *Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SubscribePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "INVALID_PAYLOAD",
			Message: "Invalid unsubscribe payload",
		}))
		return
	}

	// 更新客户端订阅
	client.Unsubscribe(payload.Channel, payload.Topics)

	// 更新全局订阅映射
	g.subMu.Lock()
	if channelSubs, ok := g.subscriptions[payload.Channel]; ok {
		topics := payload.Topics
		if len(topics) == 0 {
			// 取消整个频道的订阅
			for topic := range channelSubs {
				delete(g.subscriptions[payload.Channel][topic], client.ID)
			}
		} else {
			for _, topic := range topics {
				if topicSubs, ok := channelSubs[topic]; ok {
					delete(topicSubs, client.ID)
				}
			}
		}
	}
	g.subMu.Unlock()

	client.SendMessage(NewMessageWithID(msg.ID, MsgTypeUnsubscribed, SubscribeResult{
		Channel: payload.Channel,
		Topics:  payload.Topics,
		Success: true,
	}))
}

// handlePing 处理心跳
func (g *Gateway) handlePing(client *Client, msg *Message) {
	client.SendMessage(NewMessageWithID(msg.ID, MsgTypePong, nil))
}

// handleTaskAction 处理任务操作
func (g *Gateway) handleTaskAction(client *Client, msg *Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload TaskActionPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "INVALID_PAYLOAD",
			Message: "Invalid task action payload",
		}))
		return
	}

	if g.taskActionFunc == nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeTaskResult, TaskActionResult{
			TaskID:  payload.TaskID,
			Action:  payload.Action,
			Success: false,
			Message: "Task action not configured",
		}))
		return
	}

	err := g.taskActionFunc(payload.TaskID, payload.Action, payload.Data)
	if err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeTaskResult, TaskActionResult{
			TaskID:  payload.TaskID,
			Action:  payload.Action,
			Success: false,
			Message: err.Error(),
		}))
		return
	}

	client.SendMessage(NewMessageWithID(msg.ID, MsgTypeTaskResult, TaskActionResult{
		TaskID:  payload.TaskID,
		Action:  payload.Action,
		Success: true,
	}))
}

// handleSessionExec 处理会话执行
func (g *Gateway) handleSessionExec(client *Client, msg *Message) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionExecPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "INVALID_PAYLOAD",
			Message: "Invalid session exec payload",
		}))
		return
	}

	if g.sessionExecFunc == nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "NOT_CONFIGURED",
			Message: "Session exec not configured",
		}))
		return
	}

	execID, err := g.sessionExecFunc(payload.SessionID, payload.Command)
	if err != nil {
		client.SendMessage(NewMessageWithID(msg.ID, MsgTypeError, ErrorPayload{
			Code:    "EXEC_FAILED",
			Message: err.Error(),
		}))
		return
	}

	client.SendMessage(NewMessageWithID(msg.ID, MsgTypeSessionOutput, SessionOutputPayload{
		SessionID: payload.SessionID,
		ExecID:    execID,
		Type:      "started",
	}))
}

// BroadcastEvent 广播事件到订阅者
func (g *Gateway) BroadcastEvent(channel, topic, eventType string, data interface{}) {
	event := NewMessage(MsgTypeEvent, EventPayload{
		Channel:   channel,
		Topic:     topic,
		EventType: eventType,
		Data:      data,
	})

	g.subMu.RLock()
	defer g.subMu.RUnlock()

	channelSubs, ok := g.subscriptions[channel]
	if !ok {
		return
	}

	// 发送给订阅了特定 topic 的客户端
	if topicClients, ok := channelSubs[topic]; ok {
		for _, client := range topicClients {
			client.SendMessage(event)
		}
	}

	// 发送给订阅了所有 topic 的客户端（通配符）
	if allClients, ok := channelSubs["*"]; ok {
		for _, client := range allClients {
			// 避免重复发送
			if topicClients, ok := channelSubs[topic]; ok {
				if _, exists := topicClients[client.ID]; exists {
					continue
				}
			}
			client.SendMessage(event)
		}
	}
}

// BroadcastToUser 向特定用户的所有设备广播消息
func (g *Gateway) BroadcastToUser(userID string, msg *Message) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if devices, ok := g.users[userID]; ok {
		for _, client := range devices {
			client.SendMessage(msg)
		}
	}
}

// SendToDevice 向特定用户的特定设备发送消息
func (g *Gateway) SendToDevice(userID, deviceID string, msg *Message) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if devices, ok := g.users[userID]; ok {
		if client, ok := devices[deviceID]; ok {
			client.SendMessage(msg)
		}
	}
}

// GetClientCount 获取当前连接的客户端数量
func (g *Gateway) GetClientCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.clients)
}

// GetUserDevices 获取用户的所有设备
func (g *Gateway) GetUserDevices(userID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var devices []string
	if devMap, ok := g.users[userID]; ok {
		for deviceID := range devMap {
			devices = append(devices, deviceID)
		}
	}
	return devices
}

// parseAuthHeader 从 HTTP 请求解析认证信息
func parseAuthHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// 也支持通过 query 参数传递 token
	return r.URL.Query().Get("token")
}
