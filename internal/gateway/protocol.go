package gateway

import "time"

// MessageType 定义消息类型
type MessageType string

const (
	// 客户端 -> 服务端
	MsgTypeAuth        MessageType = "auth"           // 认证
	MsgTypeSubscribe   MessageType = "subscribe"      // 订阅任务/会话事件
	MsgTypeUnsubscribe MessageType = "unsubscribe"    // 取消订阅
	MsgTypePing        MessageType = "ping"           // 心跳
	MsgTypeTaskAction  MessageType = "task.action"    // 任务操作（取消、追加等）
	MsgTypeSessionExec MessageType = "session.exec"   // 会话执行命令

	// 服务端 -> 客户端
	MsgTypeAuthResult    MessageType = "auth.result"      // 认证结果
	MsgTypeSubscribed    MessageType = "subscribed"       // 订阅确认
	MsgTypeUnsubscribed  MessageType = "unsubscribed"     // 取消订阅确认
	MsgTypePong          MessageType = "pong"             // 心跳响应
	MsgTypeEvent         MessageType = "event"            // 事件推送
	MsgTypeError         MessageType = "error"            // 错误消息
	MsgTypeTaskResult    MessageType = "task.result"      // 任务操作结果
	MsgTypeSessionOutput MessageType = "session.output"   // 会话输出
)

// Message 通用消息结构
type Message struct {
	ID        string      `json:"id,omitempty"`        // 消息ID（用于请求/响应匹配）
	Type      MessageType `json:"type"`                // 消息类型
	Payload   interface{} `json:"payload,omitempty"`   // 消息负载
	Timestamp int64       `json:"timestamp"`           // 时间戳
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, payload interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewMessageWithID 创建带ID的消息
func NewMessageWithID(id string, msgType MessageType, payload interface{}) *Message {
	return &Message{
		ID:        id,
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// AuthPayload 认证请求
type AuthPayload struct {
	Token    string `json:"token"`              // API Key 或 JWT
	DeviceID string `json:"device_id,omitempty"` // 设备标识（用于多端同步）
}

// AuthResult 认证结果
type AuthResult struct {
	Success  bool   `json:"success"`
	UserID   string `json:"user_id,omitempty"`
	Message  string `json:"message,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

// SubscribePayload 订阅请求
type SubscribePayload struct {
	Channel string   `json:"channel"` // task, session, system
	Topics  []string `json:"topics"`  // 具体的任务ID或会话ID列表，空表示订阅所有
}

// SubscribeResult 订阅结果
type SubscribeResult struct {
	Channel string   `json:"channel"`
	Topics  []string `json:"topics"`
	Success bool     `json:"success"`
}

// EventPayload 事件推送
type EventPayload struct {
	Channel   string      `json:"channel"`    // task, session, system
	Topic     string      `json:"topic"`      // 具体的任务ID或会话ID
	EventType string      `json:"event_type"` // 具体事件类型
	Data      interface{} `json:"data"`       // 事件数据
}

// TaskActionPayload 任务操作请求
type TaskActionPayload struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"` // cancel, append_turn
	Data   string `json:"data,omitempty"` // append_turn 时的用户输入
}

// TaskActionResult 任务操作结果
type TaskActionResult struct {
	TaskID  string `json:"task_id"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SessionExecPayload 会话执行请求
type SessionExecPayload struct {
	SessionID string `json:"session_id"`
	Command   string `json:"command"`
}

// SessionOutputPayload 会话输出
type SessionOutputPayload struct {
	SessionID string `json:"session_id"`
	ExecID    string `json:"exec_id"`
	Content   string `json:"content"`
	Type      string `json:"type"` // message, error, done
}

// ErrorPayload 错误消息
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Channel 定义
const (
	ChannelTask    = "task"
	ChannelSession = "session"
	ChannelSystem  = "system"
)
