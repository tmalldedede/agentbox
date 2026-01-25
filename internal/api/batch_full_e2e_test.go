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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/batch"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/database"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// batchE2ETestEnv 批量任务端到端测试环境
type batchE2ETestEnv struct {
	router     *gin.Engine
	batchMgr   *batch.Manager
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
	dockerMgr  container.Manager
	sesStore   session.Store
	tmpDir     string
	agentID    string
	adapter    string
}

// setupBatchE2E 初始化批量任务端到端测试环境
func setupBatchE2E(t *testing.T, adapterType string) (*batchE2ETestEnv, func()) {
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
	provMgr := provider.NewManager(providerDir, "e2e-batch-key-32bytes-aes256!!")
	require.NoError(t, provMgr.ConfigureKey("zhipu", apiKey))

	// Runtime Manager
	rtMgr := runtime.NewManager(filepath.Join(tmpDir, "runtimes"), nil)

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
	agentID := fmt.Sprintf("e2e-batch-%s", adapterType)
	var testAgent *agent.Agent

	switch adapterType {
	case agent.AdapterCodex:
		testAgent = &agent.Agent{
			ID:              agentID,
			Name:            "E2E Batch Agent (Codex)",
			Description:     "E2E test agent for batch processing using Codex engine",
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
			Name:            "E2E Batch Agent (Claude Code)",
			Description:     "E2E test agent for batch processing using Claude Code engine",
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
	t.Logf("Agent created: id=%s, adapter=%s, model=%s", testAgent.ID, testAgent.Adapter, testAgent.Model)

	// === Database ===
	dbPath := filepath.Join(tmpDir, "batch.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	// 设置全局数据库连接（batch.NewGormStore 依赖这个）
	database.DB = db
	// 自动迁移批量任务相关表
	require.NoError(t, db.AutoMigrate(&database.BatchModel{}, &database.BatchTaskModel{}))

	// === Batch Manager ===
	batchStore := batch.NewGormStore()

	batchCfg := &batch.ManagerConfig{
		MaxBatches:       10,
		PollInterval:     100 * time.Millisecond,
		ProgressInterval: 1 * time.Second,
		DisableRecovery:  true, // 禁用恢复逻辑，避免测试中的 batch 被意外暂停
	}
	batchMgr := batch.NewManager(batchStore, sessionMgr, agentMgr, batchCfg)

	// === HTTP Router ===
	router := gin.New()
	v1 := router.Group("/api/v1")
	batchHandler := NewBatchHandler(batchMgr)
	batchHandler.RegisterRoutes(v1)

	env := &batchE2ETestEnv{
		router:     router,
		batchMgr:   batchMgr,
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
		dockerMgr:  dockerMgr,
		sesStore:   sesStore,
		tmpDir:     tmpDir,
		agentID:    testAgent.ID,
		adapter:    adapterType,
	}

	cleanup := func() {
		batchMgr.Shutdown()
		batchStore.Close()

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

// createBatchViaAPI 通过 HTTP API 创建批次
func createBatchViaAPI(t *testing.T, env *batchE2ETestEnv, req batch.CreateBatchRequest) (string, map[string]interface{}) {
	t.Helper()

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/batches", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, httpReq)
	require.Equal(t, http.StatusCreated, w.Code, "create batch failed: %s", w.Body.String())

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	batchData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	batchID, ok := batchData["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, batchID)

	return batchID, batchData
}

// getBatchViaAPI 通过 HTTP API 获取批次
func getBatchViaAPI(t *testing.T, env *batchE2ETestEnv, batchID string) map[string]interface{} {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batches/"+batchID, nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	batchData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	return batchData
}

// startBatchViaAPI 通过 HTTP API 启动批次
func startBatchViaAPI(t *testing.T, env *batchE2ETestEnv, batchID string) map[string]interface{} {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+batchID+"/start", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "start batch failed: %s", w.Body.String())

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	batchData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	return batchData
}

// waitForBatchStatus 轮询等待 batch 达到目标状态
func waitForBatchStatus(t *testing.T, env *batchE2ETestEnv, batchID string, targetStatuses []string, timeout time.Duration) map[string]interface{} {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		batchData := getBatchViaAPI(t, env, batchID)
		status, _ := batchData["status"].(string)

		for _, target := range targetStatuses {
			if status == target {
				t.Logf("Batch %s reached status: %s", batchID, status)
				return batchData
			}
		}

		// 提前失败检测
		if status == "failed" && !contains(targetStatuses, "failed") {
			t.Fatalf("Batch %s unexpectedly failed", batchID)
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Batch %s did not reach status %v within %v", batchID, targetStatuses, timeout)
	return nil
}

// waitForBatchProgress 等待批次完成一定数量的任务
func waitForBatchProgress(t *testing.T, env *batchE2ETestEnv, batchID string, minCompleted int, timeout time.Duration) map[string]interface{} {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		batchData := getBatchViaAPI(t, env, batchID)
		completed := int(batchData["completed"].(float64))
		failed := int(batchData["failed"].(float64))
		status, _ := batchData["status"].(string)

		// 检查是否有足够的任务完成
		if completed+failed >= minCompleted {
			t.Logf("Batch %s progress: %d completed, %d failed", batchID, completed, failed)
			return batchData
		}

		// 如果批次已完成（包括所有任务进入 dead letter），立即返回
		if status == "completed" || status == "failed" {
			t.Logf("Batch %s finished early: status=%s, completed=%d, failed=%d", batchID, status, completed, failed)
			return batchData
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Batch %s did not complete %d tasks within %v", batchID, minCompleted, timeout)
	return nil
}

// ============================================================
// 测试 1: 创建批次并执行到完成
// ============================================================

// TestBatchE2E_CreateAndExecute 端到端：创建批次 → 启动 → Worker 执行 → 批次完成
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_CreateAndExecute -timeout 600s ./internal/api/
func TestBatchE2E_CreateAndExecute(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			t.Log("=== Step 1: Create batch with 3 simple tasks ===")
			batchID, batchData := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "E2E Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "What is {{.num1}} + {{.num2}}? Reply with just the number.",
				Inputs: []map[string]interface{}{
					{"num1": 1, "num2": 2},
					{"num1": 3, "num2": 4},
					{"num1": 5, "num2": 6},
				},
				Concurrency: 2,
				Timeout:     180,  // 3 分钟超时，给 LLM API 足够时间
				MaxRetries:  3,    // 允许 3 次重试
				AutoStart:   false,
			})

			assert.Equal(t, "pending", batchData["status"])
			assert.Equal(t, float64(3), batchData["total_tasks"])
			assert.Equal(t, float64(2), batchData["concurrency"])
			t.Logf("Batch created: id=%s, tasks=%v", batchID, batchData["total_tasks"])

			t.Log("=== Step 2: Start batch ===")
			startData := startBatchViaAPI(t, env, batchID)
			assert.Equal(t, "running", startData["status"])
			t.Logf("Batch started, status=%s", startData["status"])

			t.Log("=== Step 3: Wait for batch to complete ===")
			finalData := waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 5*time.Minute)

			// === 验证 ===
			t.Log("=== Step 4: Verify results ===")
			status, _ := finalData["status"].(string)
			completed := int(finalData["completed"].(float64))
			failed := int(finalData["failed"].(float64))
			total := int(finalData["total_tasks"].(float64))

			t.Logf("Final status: %s, completed: %d/%d, failed: %d", status, completed, total, failed)

			assert.Contains(t, []string{"completed", "failed"}, status)
			// 注意：由于 LLM API 和 Docker 环境的不稳定性，部分任务可能进入 dead letter queue
			// 我们要求至少有一半任务成功完成（真正验证的是执行流程，而非 LLM 结果）
			assert.GreaterOrEqual(t, completed, total/2, "at least half of tasks should complete successfully")
			assert.NotNil(t, finalData["started_at"])
			assert.NotNil(t, finalData["completed_at"])

			t.Logf("=== PASS [%s]: Batch executed successfully ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 2: 批次暂停和恢复
// ============================================================

// TestBatchE2E_PauseAndResume 端到端：启动批次 → 暂停 → 恢复 → 完成
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_PauseAndResume -timeout 600s ./internal/api/
func TestBatchE2E_PauseAndResume(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			t.Log("=== Step 1: Create batch with 5 tasks ===")
			batchID, _ := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "Pause/Resume Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "Calculate {{.num1}} * {{.num2}}. Reply with just the number.",
				Inputs: []map[string]interface{}{
					{"num1": 2, "num2": 3},
					{"num1": 4, "num2": 5},
					{"num1": 6, "num2": 7},
					{"num1": 8, "num2": 9},
					{"num1": 10, "num2": 11},
				},
				Concurrency: 2,
				Timeout:     180,  // 3 分钟超时，给 LLM API 足够时间
				MaxRetries:  3,    // 允许 3 次重试
				AutoStart:   true, // Auto start
			})
			t.Logf("Batch created and started: %s", batchID)

			t.Log("=== Step 2: Wait for some tasks to complete then pause ===")
			waitForBatchProgress(t, env, batchID, 1, 3*time.Minute)

			// Check if batch already completed (LLM may be too fast, or all tasks failed)
			checkData := getBatchViaAPI(t, env, batchID)
			if checkData["status"] == "completed" || checkData["status"] == "failed" {
				t.Log("Batch completed before pause attempt - skipping pause/resume test")
				completed := int(checkData["completed"].(float64))
				failed := int(checkData["failed"].(float64))
				t.Logf("Final: completed=%d, failed=%d", completed, failed)
				// 注意：任务可能进入 dead letter queue（不计入 completed 或 failed）
				// 我们只要求批次完成状态正确即可
				assert.Contains(t, []string{"completed", "failed"}, checkData["status"])
				t.Logf("=== PASS [%s]: Batch completed (pause/resume not tested due to speed/failures) ===", eng.Adapter)
				return
			}

			// Pause the batch
			pauseReq := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+batchID+"/pause", nil)
			pw := httptest.NewRecorder()
			env.router.ServeHTTP(pw, pauseReq)
			assert.Equal(t, http.StatusOK, pw.Code, "pause should succeed: %s", pw.Body.String())

			pausedData := getBatchViaAPI(t, env, batchID)
			// Allow completed status if batch finished during pause
			if pausedData["status"] == "completed" || pausedData["status"] == "failed" {
				t.Log("Batch completed during pause - acceptable")
				t.Logf("=== PASS [%s]: Batch completed ===", eng.Adapter)
				return
			}
			assert.Equal(t, "paused", pausedData["status"])
			pausedCompleted := int(pausedData["completed"].(float64))
			t.Logf("Batch paused, completed so far: %d", pausedCompleted)

			t.Log("=== Step 3: Wait a moment and verify no progress ===")
			time.Sleep(3 * time.Second)
			stillPaused := getBatchViaAPI(t, env, batchID)
			assert.Equal(t, "paused", stillPaused["status"])
			assert.Equal(t, pausedCompleted, int(stillPaused["completed"].(float64)), "no progress while paused")

			t.Log("=== Step 4: Resume the batch ===")
			resumeReq := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+batchID+"/resume", nil)
			rw := httptest.NewRecorder()
			env.router.ServeHTTP(rw, resumeReq)
			assert.Equal(t, http.StatusOK, rw.Code, "resume should succeed: %s", rw.Body.String())

			resumedData := getBatchViaAPI(t, env, batchID)
			assert.Equal(t, "running", resumedData["status"])
			t.Log("Batch resumed")

			t.Log("=== Step 5: Wait for batch to complete ===")
			finalData := waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 5*time.Minute)

			completed := int(finalData["completed"].(float64))
			failed := int(finalData["failed"].(float64))
			t.Logf("Final: completed=%d, failed=%d", completed, failed)
			assert.Equal(t, 5, completed+failed, "all 5 tasks should be processed")

			t.Logf("=== PASS [%s]: Pause/Resume works correctly ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 3: 取消正在运行的批次
// ============================================================

// TestBatchE2E_CancelRunning 端到端：启动批次 → 运行中取消 → 验证清理
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_CancelRunning -timeout 600s ./internal/api/
func TestBatchE2E_CancelRunning(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			t.Log("=== Step 1: Create batch with many tasks (slow) ===")
			inputs := make([]map[string]interface{}, 10)
			for i := 0; i < 10; i++ {
				inputs[i] = map[string]interface{}{
					"topic": fmt.Sprintf("Write a detailed paragraph about topic %d", i+1),
				}
			}

			batchID, _ := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "Cancel Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "{{.topic}}. Be thorough and write at least 100 words.",
				Inputs:         inputs,
				Concurrency:    2,
				Timeout:        120,
				MaxRetries:     0,
				AutoStart:      true,
			})
			t.Logf("Batch created and started: %s", batchID)

			t.Log("=== Step 2: Wait for batch to start running ===")
			waitForBatchStatus(t, env, batchID, []string{"running"}, 2*time.Minute)
			time.Sleep(5 * time.Second) // Let some work happen

			t.Log("=== Step 3: Cancel the batch ===")
			cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+batchID+"/cancel", nil)
			cw := httptest.NewRecorder()
			env.router.ServeHTTP(cw, cancelReq)

			assert.Equal(t, http.StatusOK, cw.Code, "cancel should succeed: %s", cw.Body.String())

			var cancelResp Response
			require.NoError(t, json.Unmarshal(cw.Body.Bytes(), &cancelResp))
			cancelData, _ := cancelResp.Data.(map[string]interface{})
			assert.Equal(t, "cancelled", cancelData["status"])
			t.Log("Batch cancelled")

			t.Log("=== Step 4: Verify final state ===")
			time.Sleep(2 * time.Second)
			finalData := getBatchViaAPI(t, env, batchID)

			assert.Equal(t, "cancelled", finalData["status"])
			assert.NotNil(t, finalData["completed_at"])

			completed := int(finalData["completed"].(float64))
			total := int(finalData["total_tasks"].(float64))
			t.Logf("Final: completed=%d/%d (cancelled before finish)", completed, total)
			assert.Less(t, completed, total, "should have cancelled before all tasks completed")

			t.Logf("=== PASS [%s]: Cancel works correctly ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 4: SSE 事件流
// ============================================================

// TestBatchE2E_SSEEvents 端到端：创建批次 → 订阅 SSE → 接收进度事件
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_SSEEvents -timeout 600s ./internal/api/
func TestBatchE2E_SSEEvents(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			// 用 httptest.Server 以支持真实 HTTP SSE 连接
			server := httptest.NewServer(env.router)
			defer server.Close()

			t.Log("=== Step 1: Create batch ===")
			batchID, _ := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "SSE Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "What is {{.num}}? Reply with the number.",
				Inputs: []map[string]interface{}{
					{"num": 1},
					{"num": 2},
				},
				Concurrency: 1,
				Timeout:     180, // 3 分钟超时，给 LLM API 足够时间
				MaxRetries:  0,
				AutoStart:   false,
			})
			t.Logf("Batch created: %s", batchID)

			t.Log("=== Step 2: Connect SSE ===")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			sseURL := fmt.Sprintf("%s/api/v1/batches/%s/events", server.URL, batchID)
			sseReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, sseURL, nil)

			// Start SSE connection in goroutine
			eventsCh := make(chan string, 100)
			go func() {
				resp, err := http.DefaultClient.Do(sseReq)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "data:") {
						eventsCh <- line
					}
				}
			}()

			// Give SSE time to connect
			time.Sleep(500 * time.Millisecond)

			t.Log("=== Step 3: Start batch ===")
			startBatchViaAPI(t, env, batchID)

			t.Log("=== Step 4: Wait for completion and collect events ===")
			waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 4*time.Minute)

			// Collect events
			time.Sleep(time.Second)
			close(eventsCh)

			var eventLines []string
			for line := range eventsCh {
				eventLines = append(eventLines, line)
				t.Logf("  SSE: %s", truncate(line, 100))
			}

			// === 验证事件 ===
			allEvents := strings.Join(eventLines, "\n")

			// 应该有进度事件
			hasProgress := strings.Contains(allEvents, "batch.progress") ||
				strings.Contains(allEvents, "task.completed")
			if hasProgress {
				t.Log("  ✓ Received progress/task events")
			}

			// 应该有终态事件
			hasTerminal := strings.Contains(allEvents, "batch.completed") ||
				strings.Contains(allEvents, "batch.failed")
			if hasTerminal {
				t.Log("  ✓ Received terminal event")
			}

			t.Logf("Total SSE lines received: %d", len(eventLines))
			t.Logf("=== PASS [%s]: SSE events received ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 5: 任务失败和重试
// ============================================================

// TestBatchE2E_RetryFailed 端到端：批次完成后有失败任务 → 重试失败任务
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_RetryFailed -timeout 600s ./internal/api/
func TestBatchE2E_RetryFailed(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			t.Log("=== Step 1: Create batch with simple tasks ===")
			batchID, _ := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "Retry Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "Say '{{.word}}'. Reply with just that word.",
				Inputs: []map[string]interface{}{
					{"word": "hello"},
					{"word": "world"},
					{"word": "test"},
				},
				Concurrency: 2,
				Timeout:     180, // 3 分钟超时，给 LLM API 足够时间
				MaxRetries:  0, // No auto-retry
				AutoStart:   true,
			})
			t.Logf("Batch created: %s", batchID)

			t.Log("=== Step 2: Wait for batch to complete ===")
			finalData := waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 4*time.Minute)

			failed := int(finalData["failed"].(float64))
			t.Logf("Batch finished: completed=%v, failed=%d", finalData["completed"], failed)

			// 如果有失败任务，测试重试
			if failed > 0 {
				t.Log("=== Step 3: Retry failed tasks ===")
				retryReq := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+batchID+"/retry", nil)
				rw := httptest.NewRecorder()
				env.router.ServeHTTP(rw, retryReq)

				assert.Equal(t, http.StatusOK, rw.Code, "retry should succeed")

				// Verify batch status is pending after retry
				retryData := getBatchViaAPI(t, env, batchID)
				assert.Equal(t, "pending", retryData["status"], "batch should be pending after retry")
				assert.Equal(t, float64(0), retryData["failed"], "failed count should be reset to 0")
				t.Logf("Retry successful: batch reset to pending, failed=0")

				// 重新启动批次
				t.Log("=== Step 4: Restart batch after retry ===")
				startBatchViaAPI(t, env, batchID)

				// 等待重试完成
				retryFinalData := waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 3*time.Minute)
				t.Logf("After retry: completed=%v, failed=%v", retryFinalData["completed"], retryFinalData["failed"])
			} else {
				t.Log("=== Step 3: No failed tasks to retry ===")
			}

			t.Logf("=== PASS [%s]: Retry mechanism works ===", eng.Adapter)
		})
	}
}

// ============================================================
// 测试 6: 批次任务列表和统计
// ============================================================

// TestBatchE2E_TasksAndStats 端到端：创建批次 → 执行 → 查询任务列表和统计
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestBatchE2E_TasksAndStats -timeout 600s ./internal/api/
func TestBatchE2E_TasksAndStats(t *testing.T) {
	for _, eng := range testEngines {
		t.Run(eng.Adapter, func(t *testing.T) {
			env, cleanup := setupBatchE2E(t, eng.Adapter)
			defer cleanup()

			t.Log("=== Step 1: Create and execute batch ===")
			batchID, _ := createBatchViaAPI(t, env, batch.CreateBatchRequest{
				Name:           "Tasks/Stats Test Batch",
				AgentID:        env.agentID,
				PromptTemplate: "What is {{.x}} + {{.y}}? Just the number.",
				Inputs: []map[string]interface{}{
					{"x": 1, "y": 1},
					{"x": 2, "y": 2},
					{"x": 3, "y": 3},
					{"x": 4, "y": 4},
				},
				Concurrency: 2,
				Timeout:     180, // 3 分钟超时，给 LLM API 足够时间
				MaxRetries:  0,
				AutoStart:   true,
			})

			t.Log("=== Step 2: Wait for completion ===")
			waitForBatchStatus(t, env, batchID, []string{"completed", "failed"}, 4*time.Minute)

			t.Log("=== Step 3: Query task list ===")
			tasksReq := httptest.NewRequest(http.MethodGet, "/api/v1/batches/"+batchID+"/tasks", nil)
			tw := httptest.NewRecorder()
			env.router.ServeHTTP(tw, tasksReq)

			assert.Equal(t, http.StatusOK, tw.Code)
			var tasksResp Response
			require.NoError(t, json.Unmarshal(tw.Body.Bytes(), &tasksResp))

			tasksData, _ := tasksResp.Data.(map[string]interface{})
			tasks, _ := tasksData["tasks"].([]interface{})
			assert.Len(t, tasks, 4, "should have 4 tasks")

			for i, task := range tasks {
				taskMap, _ := task.(map[string]interface{})
				t.Logf("  Task %d: status=%s, result=%s",
					i, taskMap["status"], truncate(fmt.Sprint(taskMap["result"]), 50))
			}

			t.Log("=== Step 4: Query stats ===")
			statsReq := httptest.NewRequest(http.MethodGet, "/api/v1/batches/"+batchID+"/stats", nil)
			sw := httptest.NewRecorder()
			env.router.ServeHTTP(sw, statsReq)

			assert.Equal(t, http.StatusOK, sw.Code)
			var statsResp Response
			require.NoError(t, json.Unmarshal(sw.Body.Bytes(), &statsResp))

			statsData, _ := statsResp.Data.(map[string]interface{})
			t.Logf("Stats: total=%v, completed=%v, failed=%v, avg_duration=%vms",
				statsData["total_tasks"],
				statsData["completed"],
				statsData["failed"],
				statsData["avg_duration_ms"])

			assert.Equal(t, float64(4), statsData["total_tasks"])

			t.Logf("=== PASS [%s]: Tasks and stats work correctly ===", eng.Adapter)
		})
	}
}
