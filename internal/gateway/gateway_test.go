package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestGateway_Connection(t *testing.T) {
	gw := NewGateway()
	gw.Start()
	defer gw.Stop()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleConnection(w, r)
	}))
	defer server.Close()

	// 连接 WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证客户端已连接
	if gw.GetClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", gw.GetClientCount())
	}
}

func TestGateway_Ping(t *testing.T) {
	gw := NewGateway()
	gw.Start()
	defer gw.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// 发送 ping
	pingMsg := Message{
		ID:        "test-1",
		Type:      MsgTypePing,
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws.WriteJSON(pingMsg); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// 接收 pong
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read pong: %v", err)
	}

	var resp Message
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("Failed to unmarshal pong: %v", err)
	}

	if resp.Type != MsgTypePong {
		t.Errorf("Expected pong, got %s", resp.Type)
	}
	if resp.ID != "test-1" {
		t.Errorf("Expected ID test-1, got %s", resp.ID)
	}
}

func TestGateway_Auth(t *testing.T) {
	gw := NewGateway()
	gw.SetAuthFunc(func(token string) (string, error) {
		if token == "valid-token" {
			return "user-123", nil
		}
		return "", nil
	})
	gw.Start()
	defer gw.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// 发送认证
	authMsg := Message{
		ID:   "auth-1",
		Type: MsgTypeAuth,
		Payload: AuthPayload{
			Token:    "valid-token",
			DeviceID: "device-1",
		},
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws.WriteJSON(authMsg); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	// 接收认证结果
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read auth result: %v", err)
	}

	var resp Message
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("Failed to unmarshal auth result: %v", err)
	}

	if resp.Type != MsgTypeAuthResult {
		t.Errorf("Expected auth.result, got %s", resp.Type)
	}

	resultBytes, _ := json.Marshal(resp.Payload)
	var result AuthResult
	json.Unmarshal(resultBytes, &result)

	if !result.Success {
		t.Errorf("Expected auth success")
	}
	if result.UserID != "user-123" {
		t.Errorf("Expected user-123, got %s", result.UserID)
	}
	if result.DeviceID != "device-1" {
		t.Errorf("Expected device-1, got %s", result.DeviceID)
	}
}

func TestGateway_Subscribe(t *testing.T) {
	gw := NewGateway()
	gw.Start()
	defer gw.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// 订阅任务事件
	subMsg := Message{
		ID:   "sub-1",
		Type: MsgTypeSubscribe,
		Payload: SubscribePayload{
			Channel: ChannelTask,
			Topics:  []string{"task-123"},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws.WriteJSON(subMsg); err != nil {
		t.Fatalf("Failed to send subscribe: %v", err)
	}

	// 接收订阅确认
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read subscribe result: %v", err)
	}

	var resp Message
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("Failed to unmarshal subscribe result: %v", err)
	}

	if resp.Type != MsgTypeSubscribed {
		t.Errorf("Expected subscribed, got %s", resp.Type)
	}

	// 广播事件
	go func() {
		time.Sleep(100 * time.Millisecond)
		gw.BroadcastEvent(ChannelTask, "task-123", "task.started", map[string]string{"status": "running"})
	}()

	// 接收事件
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read event: %v", err)
	}

	var eventMsg Message
	if err := json.Unmarshal(msg, &eventMsg); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if eventMsg.Type != MsgTypeEvent {
		t.Errorf("Expected event, got %s", eventMsg.Type)
	}
}

func TestGateway_BroadcastToUser(t *testing.T) {
	gw := NewGateway()
	gw.SetAuthFunc(func(token string) (string, error) {
		if token == "valid-token" {
			return "user-123", nil
		}
		return "", nil
	})
	gw.Start()
	defer gw.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleConnection(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// 先认证
	authMsg := Message{
		ID:   "auth-1",
		Type: MsgTypeAuth,
		Payload: AuthPayload{
			Token:    "valid-token",
			DeviceID: "device-1",
		},
		Timestamp: time.Now().UnixMilli(),
	}
	if err := ws.WriteJSON(authMsg); err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	// 读取认证响应
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws.ReadMessage()

	// 广播到用户
	go func() {
		time.Sleep(100 * time.Millisecond)
		gw.BroadcastToUser("user-123", NewMessage(MsgTypeEvent, map[string]string{"test": "data"}))
	}()

	// 接收消息
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read broadcast: %v", err)
	}

	var eventMsg Message
	if err := json.Unmarshal(msg, &eventMsg); err != nil {
		t.Fatalf("Failed to unmarshal broadcast: %v", err)
	}

	if eventMsg.Type != MsgTypeEvent {
		t.Errorf("Expected event, got %s", eventMsg.Type)
	}
}
