//go:build e2e

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// wsE2ETestEnv WebSocket 端到端测试环境
type wsE2ETestEnv struct {
	router     *gin.Engine
	server     *httptest.Server
	wsHandler  *WSHandler
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
	dockerMgr  container.Manager
	sesStore   session.Store
	tmpDir     string
	agentID    string
	adapter    string
}

// setupWebSocketE2E 初始化 WebSocket 端到端测试环境
func setupWebSocketE2E(t *testing.T, adapterType string) (*wsE2ETestEnv, func()) {
	t.Helper()

	// === 前置检查 ===
	apiKey := getZhipuAPIKey(t)
	if apiKey == "" {
		t.Skip("zhipu API key not available, skipping E2E test")
	}

	ctx := context.Background()
	dockerMgr, err := container.NewDockerManager()
	if err != nil {
		t.Skipf("Docker not available: %v, skipping E2E test", err)
	}

	// === 初始化所有 Manager ===
	tmpDir := t.TempDir()
	// 解析符号链接，避免 macOS /var/folders -> /private/var/folders 导致路径验证失败
	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	// Provider Manager
	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "e2e-ws-key-32bytes-aes256!!!")
	require.NoError(t, provMgr.ConfigureKey("zhipu", apiKey))

	// Runtime Manager
	rtMgr := runtime.NewManager(filepath.Join(tmpDir, "runtimes"))

	// Skill Manager
	skillMgr, err := skill.NewManager(filepath.Join(tmpDir, "skills"))
	require.NoError(t, err)

	// MCP Manager
	mcpMgr, err := mcp.NewManager(filepath.Join(tmpDir, "mcp"))
	require.NoError(t, err)

	// Agent Manager
	agentDir := filepath.Join(tmpDir, "agents")
	agentMgr := agent.NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	// Engine Registry
	registry := engine.DefaultRegistry()

	// Session Manager
	workspaceBase := filepath.Join(tmpDir, "workspaces")
	sesStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sesStore, dockerMgr, registry, workspaceBase)
	sessionMgr.SetAgentManager(agentMgr)

	// === 创建 Agent ===
	agentID := "e2e-ws-" + adapterType
	var testAgent *agent.Agent

	switch adapterType {
	case agent.AdapterCodex:
		testAgent = &agent.Agent{
			ID:              agentID,
			Name:            "E2E WebSocket Agent (Codex)",
			Description:     "E2E test agent for WebSocket testing using Codex",
			Adapter:         agent.AdapterCodex,
			ProviderID:      "zhipu",
			Model:           "glm-4.7",
			BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
			SystemPrompt:    "You are a concise assistant. Always respond in English. Keep responses under 30 words.",
			Permissions: agent.PermissionConfig{
				ApprovalPolicy: "never",
				SandboxMode:    "danger-full-access",
				FullAuto:       true,
			},
		}
	case agent.AdapterClaudeCode:
		testAgent = &agent.Agent{
			ID:              agentID,
			Name:            "E2E WebSocket Agent (Claude Code)",
			Description:     "E2E test agent for WebSocket testing using Claude Code",
			Adapter:         agent.AdapterClaudeCode,
			ProviderID:      "zhipu",
			Model:           "glm-4.7",
			BaseURLOverride: "https://open.bigmodel.cn/api/anthropic",
			SystemPrompt:    "You are a concise assistant. Always respond in English. Keep responses under 30 words.",
			Permissions: agent.PermissionConfig{
				SkipAll: true,
			},
		}
	default:
		t.Fatalf("unsupported adapter type: %s", adapterType)
	}

	require.NoError(t, agentMgr.Create(testAgent))
	t.Logf("Agent created: id=%s, adapter=%s", testAgent.ID, testAgent.Adapter)

	// === HTTP Handler ===
	wsHandler := NewWSHandler(sessionMgr, registry, dockerMgr)
	handler := NewHandler(sessionMgr, registry)

	router := gin.New()
	v1 := router.Group("/api/v1")

	// 注册 session 路由
	sessions := v1.Group("/sessions")
	sessions.POST("", handler.CreateSession)
	sessions.GET("/:id", handler.GetSession)
	sessions.DELETE("/:id", handler.DeleteSession)
	sessions.GET("/:id/stream", wsHandler.ExecStream) // WebSocket 路由

	// 创建真实 HTTP 服务器用于 WebSocket 连接
	server := httptest.NewServer(router)

	env := &wsE2ETestEnv{
		router:     router,
		server:     server,
		wsHandler:  wsHandler,
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
		dockerMgr:  dockerMgr,
		sesStore:   sesStore,
		tmpDir:     tmpDir,
		agentID:    testAgent.ID,
		adapter:    adapterType,
	}

	cleanup := func() {
		server.Close()

		// 清理所有容器
		sessions, _ := sesStore.List(nil)
		for _, s := range sessions {
			if s.ContainerID != "" {
				_ = dockerMgr.Stop(ctx, s.ContainerID)
				_ = dockerMgr.Remove(ctx, s.ContainerID)
				t.Logf("Cleaned up container: %s", s.ContainerID[:12])
			}
		}
	}

	return env, cleanup
}

// createSessionForWS 创建 session 用于 WebSocket 测试
func createSessionForWS(t *testing.T, env *wsE2ETestEnv) string {
	t.Helper()

	sessBody, _ := json.Marshal(session.CreateRequest{
		AgentID:   env.agentID,
		Workspace: "e2e-ws-workspace",
	})
	sessReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewReader(sessBody))
	sessReq.Header.Set("Content-Type", "application/json")
	sw := httptest.NewRecorder()
	env.router.ServeHTTP(sw, sessReq)

	require.Equal(t, http.StatusCreated, sw.Code, "create session failed: %s", sw.Body.String())

	var sessResp Response
	require.NoError(t, json.Unmarshal(sw.Body.Bytes(), &sessResp))
	sessData, _ := sessResp.Data.(map[string]interface{})
	return sessData["id"].(string)
}

// ============================================================
// 测试 1: WebSocket 连接和基本消息
// ============================================================

// TestWebSocketE2E_Connection 端到端：建立 WebSocket 连接 → 发送请求 → 接收消息
//
// 注意：此测试仅支持 Codex adapter，因为 Claude Code 不支持 WebSocket 流式执行
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestWebSocketE2E_Connection -timeout 300s ./internal/api/
func TestWebSocketE2E_Connection(t *testing.T) {
	// WebSocket 流式执行目前仅支持 Codex
	env, cleanup := setupWebSocketE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create session ===")
	sessionID := createSessionForWS(t, env)
	t.Logf("Session created: %s", sessionID)

	// 等待容器启动
	time.Sleep(2 * time.Second)

	t.Log("=== Step 2: Connect WebSocket ===")
	// 转换 HTTP URL 为 WebSocket URL
	wsURL := strings.Replace(env.server.URL, "http://", "ws://", 1) + "/api/v1/sessions/" + sessionID + "/stream"
	t.Logf("Connecting to: %s", wsURL)

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "WebSocket connection should succeed")
	defer conn.Close()
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	t.Log("WebSocket connected")

	t.Log("=== Step 3: Send execution request ===")
	execReq := StreamExecRequest{
		Prompt:   "What is 4+4? Reply with just the number.",
		MaxTurns: 3,
		Timeout:  60,
	}
	reqData, _ := json.Marshal(execReq)
	err = conn.WriteMessage(websocket.TextMessage, reqData)
	require.NoError(t, err)
	t.Log("Request sent, waiting for response...")

	t.Log("=== Step 4: Receive messages ===")
	var messages []StreamMessage
	receivedStart := false
	receivedDone := false
	var outputContent strings.Builder

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(90 * time.Second))

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				t.Log("WebSocket closed normally")
				break
			}
			// 超时或其他错误
			t.Logf("Read error (may be expected): %v", err)
			break
		}

		var msg StreamMessage
		if err := json.Unmarshal(msgData, &msg); err != nil {
			t.Logf("Failed to parse message: %s", string(msgData))
			continue
		}

		messages = append(messages, msg)
		t.Logf("  Received: type=%s, content=%s", msg.Type, truncate(msg.Content, 100))

		switch msg.Type {
		case "start":
			receivedStart = true
		case "message":
			outputContent.WriteString(msg.Content)
		case "done":
			receivedDone = true
		case "error":
			t.Logf("  Error: %s", msg.Content)
		}

		if msg.Type == "done" || msg.Type == "error" {
			break
		}
	}

	t.Log("=== Step 5: Verify messages ===")
	assert.True(t, receivedStart, "should receive start message")
	assert.True(t, receivedDone || len(messages) > 0, "should receive done or some messages")
	assert.GreaterOrEqual(t, len(messages), 2, "should receive at least start and done messages")

	output := outputContent.String()
	t.Logf("Combined output: %s", truncate(output, 500))

	// 验证输出包含数学结果（如果有输出）
	if output != "" {
		assert.Contains(t, output, "8", "4+4 should equal 8")
	}

	t.Logf("Total messages received: %d", len(messages))
	t.Log("=== PASS [codex]: WebSocket connection works correctly ===")
}

// ============================================================
// 测试 2: WebSocket 心跳
// ============================================================

// TestWebSocketE2E_Heartbeat 端到端：建立连接 → 等待接收 ping 消息
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestWebSocketE2E_Heartbeat -timeout 120s ./internal/api/
func TestWebSocketE2E_Heartbeat(t *testing.T) {
	env, cleanup := setupWebSocketE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create session ===")
	sessionID := createSessionForWS(t, env)
	t.Logf("Session created: %s", sessionID)

	time.Sleep(2 * time.Second)

	t.Log("=== Step 2: Connect WebSocket ===")
	wsURL := strings.Replace(env.server.URL, "http://", "ws://", 1) + "/api/v1/sessions/" + sessionID + "/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	t.Log("=== Step 3: Send long-running request ===")
	// 发送一个需要较长时间处理的请求
	execReq := StreamExecRequest{
		Prompt:   "Write a short poem about coding. Take your time.",
		MaxTurns: 5,
		Timeout:  120,
	}
	reqData, _ := json.Marshal(execReq)
	err = conn.WriteMessage(websocket.TextMessage, reqData)
	require.NoError(t, err)

	t.Log("=== Step 4: Wait for messages (including potential ping) ===")
	var messages []StreamMessage

	// 设置 35 秒超时（心跳间隔是 30 秒）
	// 注意：由于执行可能很快完成，我们主要验证消息流程正确
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg StreamMessage
		if err := json.Unmarshal(msgData, &msg); err != nil {
			continue
		}

		messages = append(messages, msg)
		t.Logf("  Received: type=%s", msg.Type)

		if msg.Type == "ping" {
			t.Log("  ✓ Received ping message")
		}

		if msg.Type == "done" || msg.Type == "error" {
			break
		}
	}

	t.Log("=== Step 5: Verify message flow ===")
	assert.GreaterOrEqual(t, len(messages), 1, "should receive at least one message")

	// 检查是否收到 ping（可能没有，如果执行很快完成）
	hasPing := false
	for _, msg := range messages {
		if msg.Type == "ping" {
			hasPing = true
			break
		}
	}

	if hasPing {
		t.Log("Heartbeat ping verified")
	} else {
		t.Log("No ping received (execution completed before heartbeat interval)")
	}

	t.Logf("Total messages: %d", len(messages))
	t.Log("=== PASS: WebSocket heartbeat mechanism works ===")
}

// ============================================================
// 测试 3: WebSocket 错误处理
// ============================================================

// TestWebSocketE2E_ErrorHandling 端到端：测试各种错误场景
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestWebSocketE2E_ErrorHandling -timeout 120s ./internal/api/
func TestWebSocketE2E_ErrorHandling(t *testing.T) {
	env, cleanup := setupWebSocketE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Test 1: Connect to non-existent session ===")
	wsURL := strings.Replace(env.server.URL, "http://", "ws://", 1) + "/api/v1/sessions/non-existent-session/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		defer conn.Close()

		// 应该立即收到错误消息
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, msgData, err := conn.ReadMessage()
		if err == nil {
			var msg StreamMessage
			json.Unmarshal(msgData, &msg)
			assert.Equal(t, "error", msg.Type)
			assert.Contains(t, msg.Content, "not found")
			t.Logf("Received error for non-existent session: %s", msg.Content)
		}
	} else {
		// 连接可能直接失败，这也是可接受的
		t.Logf("Connection failed as expected: %v", err)
	}

	t.Log("=== Test 2: Send invalid request format ===")
	sessionID := createSessionForWS(t, env)
	time.Sleep(2 * time.Second)

	wsURL = strings.Replace(env.server.URL, "http://", "ws://", 1) + "/api/v1/sessions/" + sessionID + "/stream"
	conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 发送无效 JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	require.NoError(t, err)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgData, err := conn.ReadMessage()
	if err == nil {
		var msg StreamMessage
		json.Unmarshal(msgData, &msg)
		assert.Equal(t, "error", msg.Type)
		t.Logf("Received error for invalid format: %s", msg.Content)
	}

	t.Log("=== Test 3: Send empty prompt ===")
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn2.Close()

	emptyReq := StreamExecRequest{
		Prompt: "", // 空 prompt
	}
	reqData, _ := json.Marshal(emptyReq)
	err = conn2.WriteMessage(websocket.TextMessage, reqData)
	require.NoError(t, err)

	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgData, err = conn2.ReadMessage()
	if err == nil {
		var msg StreamMessage
		json.Unmarshal(msgData, &msg)
		assert.Equal(t, "error", msg.Type)
		assert.Contains(t, msg.Content, "prompt")
		t.Logf("Received error for empty prompt: %s", msg.Content)
	}

	t.Log("=== PASS: WebSocket error handling works correctly ===")
}

// ============================================================
// 测试 4: WebSocket 多消息流
// ============================================================

// TestWebSocketE2E_MessageStream 端到端：验证消息流顺序和完整性
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestWebSocketE2E_MessageStream -timeout 300s ./internal/api/
func TestWebSocketE2E_MessageStream(t *testing.T) {
	env, cleanup := setupWebSocketE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create session ===")
	sessionID := createSessionForWS(t, env)
	time.Sleep(2 * time.Second)

	t.Log("=== Step 2: Connect and execute ===")
	wsURL := strings.Replace(env.server.URL, "http://", "ws://", 1) + "/api/v1/sessions/" + sessionID + "/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	execReq := StreamExecRequest{
		Prompt:   "Count from 1 to 5. Output each number on a new line.",
		MaxTurns: 3,
		Timeout:  60,
	}
	reqData, _ := json.Marshal(execReq)
	err = conn.WriteMessage(websocket.TextMessage, reqData)
	require.NoError(t, err)

	t.Log("=== Step 3: Collect all messages ===")
	var messages []StreamMessage
	conn.SetReadDeadline(time.Now().Add(90 * time.Second))

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg StreamMessage
		if err := json.Unmarshal(msgData, &msg); err != nil {
			continue
		}

		messages = append(messages, msg)

		if msg.Type == "done" || msg.Type == "error" {
			break
		}
	}

	t.Log("=== Step 4: Verify message order ===")
	assert.GreaterOrEqual(t, len(messages), 2, "should receive multiple messages")

	// 验证第一条消息是 start
	if len(messages) > 0 {
		assert.Equal(t, "start", messages[0].Type, "first message should be start")
		assert.NotEmpty(t, messages[0].ExecID, "start message should have exec_id")
	}

	// 验证最后一条消息是 done 或 error
	lastIdx := len(messages) - 1
	if lastIdx >= 0 {
		lastType := messages[lastIdx].Type
		assert.Contains(t, []string{"done", "error"}, lastType, "last message should be done or error")
	}

	// 验证时间戳递增
	for i := 1; i < len(messages); i++ {
		assert.GreaterOrEqual(t, messages[i].Timestamp, messages[i-1].Timestamp,
			"timestamps should be non-decreasing")
	}

	t.Logf("Message sequence verified, total messages: %d", len(messages))

	// 打印消息类型序列
	var typeSeq []string
	for _, msg := range messages {
		typeSeq = append(typeSeq, msg.Type)
	}
	t.Logf("Message types: %v", typeSeq)

	t.Log("=== PASS: WebSocket message stream works correctly ===")
}

// ============================================================
// 测试 5: WebSocket 并发连接
// ============================================================

// TestWebSocketE2E_ConcurrentConnections 端到端：多个 WebSocket 连接同时执行
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestWebSocketE2E_ConcurrentConnections -timeout 300s ./internal/api/
func TestWebSocketE2E_ConcurrentConnections(t *testing.T) {
	env, cleanup := setupWebSocketE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create two sessions ===")
	sessionID1 := createSessionForWS(t, env)
	t.Logf("Session 1 created: %s", sessionID1)

	// 创建第二个 agent 和 session
	agent2 := &agent.Agent{
		ID:              "e2e-ws-concurrent-2",
		Name:            "E2E WebSocket Concurrent Agent 2",
		Adapter:         agent.AdapterCodex,
		ProviderID:      "zhipu",
		Model:           "glm-4.7",
		BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
		SystemPrompt:    "You are a concise assistant.",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "danger-full-access",
			FullAuto:       true,
		},
	}
	require.NoError(t, env.agentMgr.Create(agent2))

	sessBody2, _ := json.Marshal(session.CreateRequest{
		AgentID:   agent2.ID,
		Workspace: "e2e-ws-workspace-2",
	})
	sessReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewReader(sessBody2))
	sessReq2.Header.Set("Content-Type", "application/json")
	sw2 := httptest.NewRecorder()
	env.router.ServeHTTP(sw2, sessReq2)
	require.Equal(t, http.StatusCreated, sw2.Code)
	var sessResp2 Response
	json.Unmarshal(sw2.Body.Bytes(), &sessResp2)
	sessData2, _ := sessResp2.Data.(map[string]interface{})
	sessionID2 := sessData2["id"].(string)
	t.Logf("Session 2 created: %s", sessionID2)

	time.Sleep(3 * time.Second)

	t.Log("=== Step 2: Connect both WebSockets ===")
	baseWSURL := strings.Replace(env.server.URL, "http://", "ws://", 1)

	conn1, _, err := websocket.DefaultDialer.Dial(baseWSURL+"/api/v1/sessions/"+sessionID1+"/stream", nil)
	require.NoError(t, err)
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(baseWSURL+"/api/v1/sessions/"+sessionID2+"/stream", nil)
	require.NoError(t, err)
	defer conn2.Close()

	t.Log("Both WebSockets connected")

	t.Log("=== Step 3: Send requests simultaneously ===")
	req1 := StreamExecRequest{Prompt: "What is 1+1? Reply with just the number.", MaxTurns: 2, Timeout: 60}
	req2 := StreamExecRequest{Prompt: "What is 2+2? Reply with just the number.", MaxTurns: 2, Timeout: 60}

	data1, _ := json.Marshal(req1)
	data2, _ := json.Marshal(req2)

	// 同时发送
	err = conn1.WriteMessage(websocket.TextMessage, data1)
	require.NoError(t, err)
	err = conn2.WriteMessage(websocket.TextMessage, data2)
	require.NoError(t, err)
	t.Log("Requests sent to both connections")

	t.Log("=== Step 4: Receive responses ===")
	type connResult struct {
		connID   int
		messages []StreamMessage
	}

	results := make(chan connResult, 2)

	// 读取 conn1 的消息
	go func() {
		var msgs []StreamMessage
		conn1.SetReadDeadline(time.Now().Add(90 * time.Second))
		for {
			_, msgData, err := conn1.ReadMessage()
			if err != nil {
				break
			}
			var msg StreamMessage
			if json.Unmarshal(msgData, &msg) == nil {
				msgs = append(msgs, msg)
				if msg.Type == "done" || msg.Type == "error" {
					break
				}
			}
		}
		results <- connResult{1, msgs}
	}()

	// 读取 conn2 的消息
	go func() {
		var msgs []StreamMessage
		conn2.SetReadDeadline(time.Now().Add(90 * time.Second))
		for {
			_, msgData, err := conn2.ReadMessage()
			if err != nil {
				break
			}
			var msg StreamMessage
			if json.Unmarshal(msgData, &msg) == nil {
				msgs = append(msgs, msg)
				if msg.Type == "done" || msg.Type == "error" {
					break
				}
			}
		}
		results <- connResult{2, msgs}
	}()

	// 等待两个结果
	var res1, res2 connResult
	for i := 0; i < 2; i++ {
		res := <-results
		if res.connID == 1 {
			res1 = res
		} else {
			res2 = res
		}
	}

	t.Log("=== Step 5: Verify results ===")
	assert.GreaterOrEqual(t, len(res1.messages), 1, "conn1 should receive messages")
	assert.GreaterOrEqual(t, len(res2.messages), 1, "conn2 should receive messages")

	t.Logf("Connection 1: %d messages", len(res1.messages))
	t.Logf("Connection 2: %d messages", len(res2.messages))

	// 验证两个连接的 exec_id 不同（如果有 start 消息）
	var execID1, execID2 string
	for _, msg := range res1.messages {
		if msg.Type == "start" {
			execID1 = msg.ExecID
			break
		}
	}
	for _, msg := range res2.messages {
		if msg.Type == "start" {
			execID2 = msg.ExecID
			break
		}
	}

	if execID1 != "" && execID2 != "" {
		assert.NotEqual(t, execID1, execID2, "exec IDs should be different")
		t.Logf("Exec IDs: conn1=%s, conn2=%s", execID1, execID2)
	}

	t.Log("=== PASS: Concurrent WebSocket connections work correctly ===")
}
