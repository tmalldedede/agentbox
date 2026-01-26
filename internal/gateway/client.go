package gateway

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	ID        string
	UserID    string
	DeviceID  string
	Conn      *websocket.Conn
	Gateway   *Gateway
	Send      chan []byte
	Subscriptions map[string]map[string]bool // channel -> topics
	mu        sync.RWMutex
	closed    bool
	closedMu  sync.RWMutex
}

// NewClient 创建新客户端
func NewClient(id string, conn *websocket.Conn, gw *Gateway) *Client {
	return &Client{
		ID:            id,
		Conn:          conn,
		Gateway:       gw,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]map[string]bool),
	}
}

// Subscribe 订阅频道的特定主题
func (c *Client) Subscribe(channel string, topics []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Subscriptions[channel] == nil {
		c.Subscriptions[channel] = make(map[string]bool)
	}

	if len(topics) == 0 {
		// 空主题表示订阅所有
		c.Subscriptions[channel]["*"] = true
	} else {
		for _, topic := range topics {
			c.Subscriptions[channel][topic] = true
		}
	}
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(channel string, topics []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Subscriptions[channel] == nil {
		return
	}

	if len(topics) == 0 {
		// 空主题表示取消订阅整个频道
		delete(c.Subscriptions, channel)
	} else {
		for _, topic := range topics {
			delete(c.Subscriptions[channel], topic)
		}
		if len(c.Subscriptions[channel]) == 0 {
			delete(c.Subscriptions, channel)
		}
	}
}

// IsSubscribed 检查是否订阅了特定频道和主题
func (c *Client) IsSubscribed(channel, topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics, ok := c.Subscriptions[channel]
	if !ok {
		return false
	}

	// 检查是否订阅了所有主题
	if topics["*"] {
		return true
	}

	return topics[topic]
}

// SendMessage 发送消息到客户端
func (c *Client) SendMessage(msg *Message) error {
	c.closedMu.RLock()
	if c.closed {
		c.closedMu.RUnlock()
		return nil
	}
	c.closedMu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		// 发送缓冲区满，关闭连接
		c.Close()
		return nil
	}
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.closedMu.Lock()
	if c.closed {
		c.closedMu.Unlock()
		return
	}
	c.closed = true
	c.closedMu.Unlock()

	c.Gateway.Unregister(c)
	close(c.Send)
	c.Conn.Close()
}

// ReadPump 读取消息循环
func (c *Client) ReadPump() {
	defer func() {
		c.Close()
	}()

	c.Conn.SetReadLimit(64 * 1024) // 64KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 记录异常关闭
			}
			break
		}

		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.SendMessage(NewMessage(MsgTypeError, ErrorPayload{
				Code:    "INVALID_JSON",
				Message: "Invalid JSON message",
			}))
			continue
		}

		c.Gateway.HandleMessage(c, &msg)
	}
}

// WritePump 写入消息循环
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送缓冲区中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
