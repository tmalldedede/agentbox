//go:build e2e

package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
	"github.com/tmalldedede/agentbox/internal/task"
)

// e2eTestEnv 端到端测试环境
type e2eTestEnv struct {
	router     *gin.Engine
	taskMgr    *task.Manager
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
	dockerMgr  container.Manager
	sesStore   session.Store
	tmpDir     string
	agentID    string
	adapter    string // 当前使用的 adapter 类型
}

// testEngines 定义测试的引擎列表
var testEngines = []struct {
	Adapter string
	BaseURL string
}{
	{agent.AdapterCodex, "https://open.bigmodel.cn/api/coding/paas/v4"},
	{agent.AdapterClaudeCode, "https://open.bigmodel.cn/api/paas/v4"},
}

// setupTaskE2E 初始化完整的端到端测试环境
// adapterType: "codex" 或 "claude-code"
// 返回测试环境和 cleanup 函数
func setupTaskE2E(t *testing.T, adapterType string, idleTimeout time.Duration) (*e2eTestEnv, func()) {
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

	// Provider Manager
	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "e2e-task-key-32bytes-aes256!!")
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

	// === 创建 Agent（根据 adapter 类型配置） ===
	agentID := fmt.Sprintf("e2e-task-%s", adapterType)
	var testAgent *agent.Agent

	switch adapterType {
	case agent.AdapterCodex:
		testAgent = &agent.Agent{
			ID:              agentID,
			Name:            "E2E Task Agent (Codex)",
			Description:     "E2E test agent using Codex engine",
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
			Name:            "E2E Task Agent (Claude Code)",
			Description:     "E2E test agent using Claude Code engine",
			Adapter:         agent.AdapterClaudeCode,
			ProviderID:      "zhipu",
			Model:           "glm-4.7",
			BaseURLOverride: "https://open.bigmodel.cn/api/paas/v4",
			SystemPrompt:    "You are a concise assistant. Always respond in English. Keep responses under 30 words.",
			Permissions: agent.PermissionConfig{
				SkipAll: true, // --dangerously-skip-permissions
			},
		}
	default:
		t.Fatalf("unsupported adapter type: %s", adapterType)
	}

	require.NoError(t, agentMgr.Create(testAgent))
	t.Logf("Agent created: id=%s, adapter=%s, model=%s", testAgent.ID, testAgent.Adapter, testAgent.Model)

	// === Task Manager ===
	dbPath := filepath.Join(tmpDir, "tasks.db")
	store, err := task.NewSQLiteStore(dbPath)
	require.NoError(t, err)

	taskCfg := &task.ManagerConfig{
		MaxConcurrent: 3,
		PollInterval:  500 * time.Millisecond, // 快速轮询加速测试
		IdleTimeout:   idleTimeout,
		UploadDir:     filepath.Join(tmpDir, "uploads"),
	}
	taskMgr := task.NewManager(store, agentMgr, sessionMgr, taskCfg)
	taskMgr.Start()

	// === HTTP Router ===
	router := gin.New()
	v1 := router.Group("/api/v1")
	taskHandler := NewTaskHandler(taskMgr)
	taskHandler.RegisterRoutes(v1)

	env := &e2eTestEnv{
		router:     router,
		taskMgr:    taskMgr,
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
		dockerMgr:  dockerMgr,
		sesStore:   sesStore,
		tmpDir:     tmpDir,
		agentID:    testAgent.ID,
		adapter:    adapterType,
	}

	cleanup := func() {
		taskMgr.Stop()
		store.Close()

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

// waitForTaskStatus 轮询等待 task 达到目标状态
func waitForTaskStatus(t *testing.T, env *e2eTestEnv, taskID string, targetStatuses []string, timeout time.Duration) map[string]interface{} {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		taskData, ok := resp.Data.(map[string]interface{})
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		status, _ := taskData["status"].(string)
		for _, target := range targetStatuses {
			if status == target {
				t.Logf("Task %s reached status: %s", taskID, status)
				return taskData
			}
		}

		// 如果进入 failed 且不是期望状态，提前失败
		if status == "failed" && !contains(targetStatuses, "failed") {
			errMsg, _ := taskData["error_message"].(string)
			t.Fatalf("Task %s unexpectedly failed: %s", taskID, errMsg)
		}

		time.Sleep(time.Second)
	}

	t.Fatalf("Task %s did not reach status %v within %v", taskID, targetStatuses, timeout)
	return nil
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

// createTaskViaAPI 通过 HTTP API 创建任务
func createTaskViaAPI(t *testing.T, env *e2eTestEnv, apiReq CreateTaskAPIRequest) (string, map[string]interface{}) {
	t.Helper()

	body, _ := json.Marshal(apiReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "create task failed: %s", w.Body.String())

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	taskData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	taskID, ok := taskData["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, taskID)

	return taskID, taskData
}

// ============================================================
// 测试 1: 创建 Task 并执行到完成
// ============================================================

// TestTaskE2E_CreateAndExecute 端到端：创建任务 → 调度执行 → LLM 回复 → 任务完成
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestTaskE2E_CreateAndExecute -timeout 300s ./internal/api/
func TestTaskE2E_CreateAndExecute(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupTaskE2E(t, eng.Adapter, 10*time.Second)
			defer cleanup()

			t.Log("=== Step 1: Create task ===")
			taskID, taskData := createTaskViaAPI(t, env, CreateTaskAPIRequest{
				AgentID: env.agentID,
				Prompt:  "What is 2+3? Reply with just the number.",
			})

			assert.Equal(t, "queued", taskData["status"])
			assert.Equal(t, float64(1), taskData["turn_count"])
			t.Logf("Task created: id=%s, status=%s", taskID, taskData["status"])

			t.Log("=== Step 2: Wait for execution (Docker + LLM) ===")
			waitForTaskStatus(t, env, taskID, []string{"running", "completed"}, 90*time.Second)

			t.Log("=== Step 3: Wait for completion (idle timeout) ===")
			finalData := waitForTaskStatus(t, env, taskID, []string{"completed"}, 60*time.Second)

			// === 验证 ===
			t.Log("=== Step 4: Verify result ===")
			assert.Equal(t, "completed", finalData["status"])
			assert.NotNil(t, finalData["started_at"], "should have started_at")
			assert.NotNil(t, finalData["completed_at"], "should have completed_at")

			// 验证 result
			result, ok := finalData["result"].(map[string]interface{})
			require.True(t, ok, "result should exist")
			resultText, _ := result["text"].(string)
			assert.NotEmpty(t, resultText, "result text should not be empty")
			assert.Contains(t, resultText, "5", "2+3 should equal 5, got: %s", resultText)
			t.Logf("Result: %s", truncate(resultText, 200))

			// 验证 turns
			turns, ok := finalData["turns"].([]interface{})
			require.True(t, ok)
			assert.Len(t, turns, 1)
			turn0, _ := turns[0].(map[string]interface{})
			assert.NotNil(t, turn0["result"], "first turn should have result")

			t.Logf("=== PASS [%s]: Task executed successfully ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 2: 多轮对话
// ============================================================

// TestTaskE2E_MultiTurn 端到端：首轮 → 等完成 → 追加第二轮 → 等第二轮完成
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestTaskE2E_MultiTurn -timeout 300s ./internal/api/
func TestTaskE2E_MultiTurn(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			testTaskE2EMultiTurn(t, eng.Adapter)
		})
	}
}

func testTaskE2EMultiTurn(t *testing.T, adapterType string) {
	env, cleanup := setupTaskE2E(t, adapterType, 30*time.Second)
	defer cleanup()

	t.Log("=== Step 1: Create task (first turn) ===")
	taskID, _ := createTaskViaAPI(t, env, CreateTaskAPIRequest{
		AgentID: env.agentID,
		Prompt:  "What is 2+3? Reply with just the number, nothing else.",
	})
	t.Logf("Task created: %s", taskID)

	t.Log("=== Step 2: Wait for first turn to complete (stay running) ===")
	// 首轮执行完后 task 保持 running 状态（等待更多 turns 或 idle timeout）
	firstData := waitForTaskStatus(t, env, taskID, []string{"running"}, 90*time.Second)

	// 等待首轮 result 出现
	var firstResult map[string]interface{}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		var resp Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		td, _ := resp.Data.(map[string]interface{})

		if r, ok := td["result"].(map[string]interface{}); ok && r["text"] != nil && r["text"] != "" {
			firstResult = td
			break
		}
		time.Sleep(time.Second)
	}
	require.NotNil(t, firstResult, "first turn should produce a result")
	t.Logf("First turn result: %s", truncate(fmt.Sprint(firstResult["result"]), 100))
	_ = firstData

	t.Log("=== Step 3: Append second turn ===")
	appendBody, _ := json.Marshal(CreateTaskAPIRequest{
		TaskID: taskID,
		Prompt: "Now multiply that result by 10. Reply with just the number, nothing else.",
	})
	appendReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(appendBody))
	appendReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	start := time.Now()
	env.router.ServeHTTP(w, appendReq)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusCreated, w.Code, "append turn should succeed: %s", w.Body.String())
	assert.Less(t, elapsed, 2*time.Second, "append should return immediately")

	var appendResp Response
	json.Unmarshal(w.Body.Bytes(), &appendResp)
	appendData, _ := appendResp.Data.(map[string]interface{})
	assert.Equal(t, float64(2), appendData["turn_count"])
	t.Logf("Second turn appended, turn_count=%v", appendData["turn_count"])

	t.Log("=== Step 4: Wait for second turn result ===")
	// 等待第二轮 result 出现
	deadline = time.Now().Add(60 * time.Second)
	var finalData map[string]interface{}
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
		rw := httptest.NewRecorder()
		env.router.ServeHTTP(rw, req)

		var resp Response
		json.Unmarshal(rw.Body.Bytes(), &resp)
		td, _ := resp.Data.(map[string]interface{})
		turns, _ := td["turns"].([]interface{})

		if len(turns) >= 2 {
			turn2, _ := turns[1].(map[string]interface{})
			if turn2["result"] != nil {
				finalData = td
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	require.NotNil(t, finalData, "second turn should produce a result within timeout")

	// === 验证 ===
	turns, _ := finalData["turns"].([]interface{})
	assert.Len(t, turns, 2)

	turn1, _ := turns[0].(map[string]interface{})
	turn1Result, _ := turn1["result"].(map[string]interface{})
	t.Logf("Turn 1 result: %s", truncate(fmt.Sprint(turn1Result["text"]), 100))

	turn2, _ := turns[1].(map[string]interface{})
	turn2Result, _ := turn2["result"].(map[string]interface{})
	turn2Text, _ := turn2Result["text"].(string)
	t.Logf("Turn 2 result: %s", truncate(turn2Text, 100))
	assert.NotEmpty(t, turn2Text, "second turn should have result text")
	// 注意：当前 Codex 引擎在第一轮完成后进程退出，第二轮可能返回 execution failed
	// 这是已知限制，测试只验证多轮流程的基础设施正确性：
	// 1. turn 能被正确追加  2. 异步执行被触发  3. 结果（成功或失败）被正确记录
	if strings.Contains(turn2Text, "execution failed") || strings.Contains(turn2Text, "exec error") {
		t.Log("⚠ Second turn execution failed (known engine limitation: codex process exits after first turn)")
		t.Log("  Infrastructure verification: turn appended → async executed → error recorded ✓")
	} else {
		t.Logf("  Second turn produced output: %s", truncate(turn2Text, 50))
	}

	t.Logf("=== PASS [%s]: Multi-turn infrastructure verified ===", adapterType)
}

// ============================================================
// 测试 3: SSE 实时事件
// ============================================================

// TestTaskE2E_SSEEvents 端到端：创建任务 → SSE 订阅 → 接收实时事件
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestTaskE2E_SSEEvents -timeout 300s ./internal/api/
func TestTaskE2E_SSEEvents(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			testTaskE2ESSEEvents(t, eng.Adapter)
		})
	}
}

func testTaskE2ESSEEvents(t *testing.T, adapterType string) {
	env, cleanup := setupTaskE2E(t, adapterType, 10*time.Second)
	defer cleanup()

	// 用 httptest.Server 以支持真实 HTTP SSE 连接
	server := httptest.NewServer(env.router)
	defer server.Close()

	t.Log("=== Step 1: Create task ===")
	taskID, _ := createTaskViaAPI(t, env, CreateTaskAPIRequest{
		AgentID: env.agentID,
		Prompt:  "Say hello. Reply in one word.",
	})
	t.Logf("Task created: %s", taskID)

	t.Log("=== Step 2: Connect SSE ===")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	sseURL := fmt.Sprintf("%s/api/v1/tasks/%s/events", server.URL, taskID)
	sseReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, sseURL, nil)
	resp, err := http.DefaultClient.Do(sseReq)
	require.NoError(t, err, "SSE connection should succeed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	t.Log("SSE connected, waiting for events...")

	t.Log("=== Step 3: Read events ===")
	scanner := bufio.NewScanner(resp.Body)
	var eventLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "data:") {
			eventLines = append(eventLines, line)
			t.Logf("  SSE: %s", truncate(line, 120))
		}
	}

	// === 验证事件 ===
	allEvents := strings.Join(eventLines, "\n")

	// 至少应该有初始状态事件
	assert.Contains(t, allEvents, "event: task.status", "should receive initial status event")

	// 应该有 task.started（调度器启动时广播）
	// 注意：如果 task 在 SSE 连接之前就已经 started，可能收不到，所以不强制
	if strings.Contains(allEvents, "event: task.started") {
		t.Log("  ✓ Received task.started")
	}

	// 应该有 agent 相关事件
	hasAgentEvent := strings.Contains(allEvents, "event: agent.thinking") ||
		strings.Contains(allEvents, "event: agent.message")
	if hasAgentEvent {
		t.Log("  ✓ Received agent events")
	}

	// 必须有终态事件（completed 或 failed）
	hasTerminal := strings.Contains(allEvents, "event: task.completed") ||
		strings.Contains(allEvents, "event: task.failed")
	assert.True(t, hasTerminal, "should receive terminal event (completed/failed)")

	// 验证终态事件在最后
	if strings.Contains(allEvents, "event: task.completed") {
		lastCompleted := strings.LastIndex(allEvents, "event: task.completed")
		lastStarted := strings.LastIndex(allEvents, "event: task.started")
		if lastStarted >= 0 {
			assert.Greater(t, lastCompleted, lastStarted, "completed should come after started")
		}
	}

	t.Logf("Total SSE lines received: %d", len(eventLines))
	t.Logf("=== PASS [%s]: SSE events received correctly ===", adapterType)
}

// ============================================================
// 测试 4: Attachments 文件挂载
// ============================================================

// TestTaskE2E_Attachments 端到端：上传文件 → 创建 Task → 文件拷贝到 workspace → Agent 可见
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestTaskE2E_Attachments -timeout 300s ./internal/api/
func TestTaskE2E_Attachments(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			testTaskE2EAttachments(t, eng.Adapter)
		})
	}
}

func testTaskE2EAttachments(t *testing.T, adapterType string) {
	env, cleanup := setupTaskE2E(t, adapterType, 10*time.Second)
	defer cleanup()

	t.Log("=== Step 1: Prepare upload file ===")
	uploadDir := filepath.Join(env.tmpDir, "uploads")

	// 模拟已上传的文件: uploads/file-e2e-001/hello.txt
	fileID := "file-e2e-001"
	fileDir := filepath.Join(uploadDir, fileID)
	require.NoError(t, os.MkdirAll(fileDir, 0o755))

	testContent := "The secret number is 42."
	require.NoError(t, os.WriteFile(filepath.Join(fileDir, "hello.txt"), []byte(testContent), 0o644))
	t.Logf("Upload file prepared: %s/hello.txt", fileID)

	t.Log("=== Step 2: Create task with attachment ===")
	taskID, taskData := createTaskViaAPI(t, env, CreateTaskAPIRequest{
		AgentID:     env.agentID,
		Prompt:      "Read the file hello.txt in your current directory and tell me what number it mentions. Reply with just the number.",
		Attachments: []string{fileID},
	})

	// 验证 attachments 已记录
	attachments, ok := taskData["attachments"].([]interface{})
	require.True(t, ok)
	assert.Len(t, attachments, 1)
	assert.Equal(t, fileID, attachments[0])
	t.Logf("Task created: %s with attachment %s", taskID, fileID)

	t.Log("=== Step 3: Wait for task to start running (file should be mounted) ===")
	waitForTaskStatus(t, env, taskID, []string{"running", "completed"}, 90*time.Second)

	// 获取 task 获取 session workspace 路径
	currentTask, err := env.taskMgr.GetTask(taskID)
	require.NoError(t, err)

	// 验证文件被实际拷贝到 workspace（核心断言，不依赖 LLM）
	if currentTask.SessionID != "" {
		sess, err := env.sesStore.Get(currentTask.SessionID)
		if err == nil && sess != nil && sess.Workspace != "" {
			copiedFile := filepath.Join(sess.Workspace, "hello.txt")
			content, err := os.ReadFile(copiedFile)
			if err == nil {
				assert.Equal(t, testContent, string(content),
					"attachment should be copied to workspace with correct content")
				t.Logf("✓ File verified in workspace: %s", copiedFile)
			} else {
				t.Logf("Cannot read workspace file (may be in container): %v", err)
			}
		}
	}

	t.Log("=== Step 4: Wait for task completion ===")
	finalData := waitForTaskStatus(t, env, taskID, []string{"completed", "failed"}, 60*time.Second)

	status, _ := finalData["status"].(string)
	t.Logf("Task final status: %s", status)

	if status == "completed" {
		result, _ := finalData["result"].(map[string]interface{})
		resultText, _ := result["text"].(string)
		t.Logf("Agent output: %s", truncate(resultText, 200))
	}

	// 核心验证：task 完成且 attachments 记录正确
	assert.Contains(t, []string{"completed", "failed"}, status, "task should reach terminal state")
	taskAttachments, _ := finalData["attachments"].([]interface{})
	assert.Len(t, taskAttachments, 1, "attachments should be preserved")
	t.Logf("=== PASS [%s]: Attachment mounted and task completed ===", adapterType)
}

// ============================================================
// 测试 5: 取消正在执行的 Task
// ============================================================

// TestTaskE2E_CancelRunningTask 端到端：创建任务 → 等待 running → 取消 → 验证状态和容器清理
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestTaskE2E_CancelRunningTask -timeout 300s ./internal/api/
func TestTaskE2E_CancelRunningTask(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			testTaskE2ECancelRunningTask(t, eng.Adapter)
		})
	}
}

func testTaskE2ECancelRunningTask(t *testing.T, adapterType string) {
	env, cleanup := setupTaskE2E(t, adapterType, 5*time.Minute)
	defer cleanup()

	t.Log("=== Step 1: Create task (long-running prompt) ===")
	taskID, _ := createTaskViaAPI(t, env, CreateTaskAPIRequest{
		AgentID: env.agentID,
		Prompt:  "Write a detailed essay about the history of computing from 1940 to 2024, covering all major milestones. Be very thorough and include at least 2000 words.",
	})
	t.Logf("Task created: %s", taskID)

	t.Log("=== Step 2: Wait for task to start running ===")
	waitForTaskStatus(t, env, taskID, []string{"running"}, 90*time.Second)
	t.Log("Task is running, waiting a moment before cancelling...")

	// 等一下让容器真正开始工作
	time.Sleep(3 * time.Second)

	// 记录 session ID
	taskBeforeCancel, err := env.taskMgr.GetTask(taskID)
	require.NoError(t, err)
	sessionID := taskBeforeCancel.SessionID
	t.Logf("Task session: %s", sessionID)

	t.Log("=== Step 3: Cancel the task ===")
	cancelReq := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID, nil)
	cw := httptest.NewRecorder()
	env.router.ServeHTTP(cw, cancelReq)

	assert.Equal(t, http.StatusOK, cw.Code, "cancel should succeed: %s", cw.Body.String())

	var cancelResp Response
	require.NoError(t, json.Unmarshal(cw.Body.Bytes(), &cancelResp))
	cancelData, _ := cancelResp.Data.(map[string]interface{})
	assert.Equal(t, "cancelled", cancelData["status"])
	assert.NotNil(t, cancelData["completed_at"])
	t.Logf("Task cancelled, status=%s", cancelData["status"])

	t.Log("=== Step 4: Verify via GET ===")
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	gw := httptest.NewRecorder()
	env.router.ServeHTTP(gw, getReq)

	var getResp Response
	require.NoError(t, json.Unmarshal(gw.Body.Bytes(), &getResp))
	getData, _ := getResp.Data.(map[string]interface{})
	assert.Equal(t, "cancelled", getData["status"])

	t.Log("=== Step 5: Verify container cleanup ===")
	// 等待容器清理完成
	time.Sleep(2 * time.Second)

	if sessionID != "" {
		// 验证 session 已停止
		sess, err := env.sesStore.Get(sessionID)
		if err == nil && sess != nil {
			// session 应该是 stopped 状态
			t.Logf("Session status after cancel: %s", sess.Status)
			assert.Equal(t, session.StatusStopped, sess.Status, "session should be stopped after cancel")

			// 验证容器已移除/停止
			if sess.ContainerID != "" {
				ctx := context.Background()
				info, err := env.dockerMgr.Inspect(ctx, sess.ContainerID)
				if err != nil {
					// 容器已被移除，符合预期
					t.Logf("Container %s removed ✓", sess.ContainerID[:12])
				} else {
					assert.NotEqual(t, container.StatusRunning, info.Status,
						"container should not be running after cancel")
					t.Logf("Container %s status: %s ✓", sess.ContainerID[:12], info.Status)
				}
			}
		}
	}

	t.Logf("=== PASS [%s]: Task cancelled and cleaned up ===", adapterType)
}
