package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/session"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	sessionMgr    *session.Manager
	agentRegistry *engine.Registry
	containerMgr  container.Manager
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(sessionMgr *session.Manager, registry *engine.Registry, containerMgr container.Manager) *WSHandler {
	return &WSHandler{
		sessionMgr:    sessionMgr,
		agentRegistry: registry,
		containerMgr:  containerMgr,
	}
}

// StreamMessage WebSocket 消息
type StreamMessage struct {
	Type      string `json:"type"`                 // message, error, done, ping
	Content   string `json:"content,omitempty"`    // 消息内容
	Timestamp int64  `json:"timestamp"`            // 时间戳
	ExecID    string `json:"execution_id,omitempty"` // 执行 ID
}

// StreamExecRequest 流式执行请求
type StreamExecRequest struct {
	Prompt          string   `json:"prompt"`
	MaxTurns        int      `json:"max_turns,omitempty"`
	Timeout         int      `json:"timeout,omitempty"`
	AllowedTools    []string `json:"allowed_tools,omitempty"`
	DisallowedTools []string `json:"disallowed_tools,omitempty"`
}

// ExecStream WebSocket 流式执行
func (h *WSHandler) ExecStream(c *gin.Context) {
	sessionID := c.Param("id")

	// 升级到 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取会话
	sess, err := h.sessionMgr.Get(c.Request.Context(), sessionID)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("session not found: %s", sessionID))
		return
	}

	if sess.Status != session.StatusRunning {
		h.sendError(conn, fmt.Sprintf("session is not running: %s", sess.Status))
		return
	}

	// 读取执行请求
	_, message, err := conn.ReadMessage()
	if err != nil {
		h.sendError(conn, "failed to read request")
		return
	}

	var req StreamExecRequest
	if err := json.Unmarshal(message, &req); err != nil {
		h.sendError(conn, "invalid request format")
		return
	}

	if req.Prompt == "" {
		h.sendError(conn, "prompt is required")
		return
	}

	// 获取 Agent 适配器
	adapter, err := h.agentRegistry.Get(sess.Agent)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("agent not found: %s", sess.Agent))
		return
	}

	// 准备执行选项
	execOpts := &engine.ExecOptions{
		Prompt:          req.Prompt,
		MaxTurns:        req.MaxTurns,
		Timeout:         req.Timeout,
		AllowedTools:    req.AllowedTools,
		DisallowedTools: req.DisallowedTools,
	}

	// 设置默认值
	if execOpts.MaxTurns <= 0 {
		execOpts.MaxTurns = 10
	}
	if execOpts.Timeout <= 0 {
		execOpts.Timeout = 300
	}

	// 准备执行命令
	cmd := adapter.PrepareExec(execOpts)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(execOpts.Timeout)*time.Second)
	defer cancel()

	// 流式执行
	stream, err := h.containerMgr.ExecStream(ctx, sess.ContainerID, cmd)
	if err != nil {
		h.sendError(conn, fmt.Sprintf("failed to execute: %v", err))
		return
	}
	defer stream.Reader.Close()

	// 发送执行开始消息
	h.sendMessage(conn, &StreamMessage{
		Type:      "start",
		ExecID:    stream.ExecID,
		Timestamp: time.Now().UnixMilli(),
	})

	// 启动心跳
	go h.heartbeat(ctx, conn)

	// 读取并发送输出
	scanner := bufio.NewScanner(stream.Reader)
	scanner.Buffer(make([]byte, 64*1024), 64*1024) // 64KB buffer

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			h.sendMessage(conn, &StreamMessage{
				Type:      "error",
				Content:   "execution timeout",
				Timestamp: time.Now().UnixMilli(),
			})
			return
		default:
			line := scanner.Text()
			h.sendMessage(conn, &StreamMessage{
				Type:      "message",
				Content:   line,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		h.sendMessage(conn, &StreamMessage{
			Type:      "error",
			Content:   fmt.Sprintf("read error: %v", err),
			Timestamp: time.Now().UnixMilli(),
		})
	}

	// 发送完成消息
	h.sendMessage(conn, &StreamMessage{
		Type:      "done",
		ExecID:    stream.ExecID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// sendMessage 发送消息
func (h *WSHandler) sendMessage(conn *websocket.Conn, msg *StreamMessage) {
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendError 发送错误
func (h *WSHandler) sendError(conn *websocket.Conn, errMsg string) {
	h.sendMessage(conn, &StreamMessage{
		Type:      "error",
		Content:   errMsg,
		Timestamp: time.Now().UnixMilli(),
	})
}

// heartbeat 心跳
func (h *WSHandler) heartbeat(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.sendMessage(conn, &StreamMessage{
				Type:      "ping",
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}
