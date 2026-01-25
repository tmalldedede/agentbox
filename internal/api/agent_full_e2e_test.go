//go:build e2e

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// agentE2ETestEnv Agent 端到端测试环境
type agentE2ETestEnv struct {
	router       *gin.Engine
	handler      *AgentHandler
	agentMgr     *agent.Manager
	sessionMgr   *session.Manager
	providerMgr  *provider.Manager
	dockerMgr    container.Manager
	sesStore     session.Store
	tmpDir       string
	adapter      string
	apiKey       string
}

// setupAgentE2E 初始化 Agent 端到端测试环境
func setupAgentE2E(t *testing.T, adapterType string) (*agentE2ETestEnv, func()) {
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
	provMgr := provider.NewManager(providerDir, "e2e-agent-key-32bytes-aes256!!")
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

	// History Manager
	historyMgr := history.NewManager(nil)

	// === HTTP Handler ===
	handler := NewAgentHandler(agentMgr, sessionMgr, historyMgr)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	env := &agentE2ETestEnv{
		router:      router,
		handler:     handler,
		agentMgr:    agentMgr,
		sessionMgr:  sessionMgr,
		providerMgr: provMgr,
		dockerMgr:   dockerMgr,
		sesStore:    sesStore,
		tmpDir:      tmpDir,
		adapter:     adapterType,
		apiKey:      apiKey,
	}

	cleanup := func() {
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

// createAgentViaAPI 通过 HTTP API 创建 Agent
func createAgentViaAPI(t *testing.T, env *agentE2ETestEnv, ag *agent.Agent) string {
	t.Helper()

	body, _ := json.Marshal(ag)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "create agent failed: %s", w.Body.String())

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	agentData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	agentID, ok := agentData["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, agentID)

	return agentID
}

// runAgentViaAPI 通过 HTTP API 执行 Agent
func runAgentViaAPI(t *testing.T, env *agentE2ETestEnv, agentID string, req RunAgentReq, timeout time.Duration) map[string]interface{} {
	t.Helper()

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+agentID+"/run", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, httpReq)

	require.Equal(t, http.StatusOK, w.Code, "run agent failed: %s", w.Body.String())

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	runData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	return runData
}

// ============================================================
// 测试 1: Codex 引擎执行
// ============================================================

// TestAgentE2E_CodexExecution 端到端：创建 Codex Agent → 执行 prompt → 获取结果
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestAgentE2E_CodexExecution -timeout 300s ./internal/api/
func TestAgentE2E_CodexExecution(t *testing.T) {
	env, cleanup := setupAgentE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create Codex Agent ===")
	testAgent := &agent.Agent{
		ID:              "e2e-codex-exec",
		Name:            "E2E Codex Execution Agent",
		Description:     "E2E test agent for Codex execution",
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
	agentID := createAgentViaAPI(t, env, testAgent)
	t.Logf("Agent created: %s", agentID)

	t.Log("=== Step 2: Execute prompt ===")
	t.Log("Waiting for container startup and LLM response (may take 30-60s)...")
	runResult := runAgentViaAPI(t, env, agentID, RunAgentReq{
		Prompt: "What is 3*4? Reply with just the number.",
		Options: &agent.RunOptions{
			MaxTurns: 3,
			Timeout:  60,
		},
	}, 90*time.Second)

	t.Log("=== Step 3: Verify result ===")
	status, _ := runResult["status"].(string)
	output, _ := runResult["output"].(string)

	t.Logf("Status: %s, Output: %s", status, truncate(output, 200))

	assert.Equal(t, "completed", status, "run should complete successfully")
	assert.NotEmpty(t, output, "should have output")
	assert.Contains(t, output, "12", "3*4 should equal 12")

	t.Log("=== PASS [codex]: Codex execution works correctly ===")
}

// ============================================================
// 测试 2: Claude Code 引擎执行
// ============================================================

// TestAgentE2E_ClaudeCodeExecution 端到端：创建 Claude Code Agent → 执行 prompt → 获取结果
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestAgentE2E_ClaudeCodeExecution -timeout 300s ./internal/api/
func TestAgentE2E_ClaudeCodeExecution(t *testing.T) {
	env, cleanup := setupAgentE2E(t, agent.AdapterClaudeCode)
	defer cleanup()

	t.Log("=== Step 1: Create Claude Code Agent ===")
	testAgent := &agent.Agent{
		ID:              "e2e-claude-exec",
		Name:            "E2E Claude Code Execution Agent",
		Description:     "E2E test agent for Claude Code execution",
		Adapter:         agent.AdapterClaudeCode,
		ProviderID:      "zhipu",
		Model:           "glm-4.7",
		BaseURLOverride: "https://open.bigmodel.cn/api/anthropic",
		SystemPrompt:    "You are a concise assistant. Always respond in English. Keep responses under 30 words.",
		Permissions: agent.PermissionConfig{
			SkipAll: true,
		},
	}
	agentID := createAgentViaAPI(t, env, testAgent)
	t.Logf("Agent created: %s", agentID)

	t.Log("=== Step 2: Execute prompt ===")
	t.Log("Waiting for container startup and LLM response (may take 30-60s)...")
	runResult := runAgentViaAPI(t, env, agentID, RunAgentReq{
		Prompt: "What is 5*6? Reply with just the number.",
		Options: &agent.RunOptions{
			MaxTurns: 3,
			Timeout:  60,
		},
	}, 90*time.Second)

	t.Log("=== Step 3: Verify result ===")
	status, _ := runResult["status"].(string)
	output, _ := runResult["output"].(string)

	t.Logf("Status: %s, Output: %s", status, truncate(output, 200))

	assert.Equal(t, "completed", status, "run should complete successfully")
	assert.NotEmpty(t, output, "should have output")
	assert.Contains(t, output, "30", "5*6 should equal 30")

	t.Log("=== PASS [claude-code]: Claude Code execution works correctly ===")
}

// ============================================================
// 测试 3: Agent CRUD 操作
// ============================================================

// TestAgentE2E_CRUD 端到端：创建 → 读取 → 更新 → 删除 Agent
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestAgentE2E_CRUD -timeout 120s ./internal/api/
func TestAgentE2E_CRUD(t *testing.T) {
	env, cleanup := setupAgentE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create Agent ===")
	testAgent := &agent.Agent{
		ID:              "e2e-crud-agent",
		Name:            "E2E CRUD Test Agent",
		Description:     "Agent for testing CRUD operations",
		Adapter:         agent.AdapterCodex,
		ProviderID:      "zhipu",
		Model:           "glm-4.7",
		BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
		SystemPrompt:    "Test system prompt",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "read-only",
			FullAuto:       true,
		},
	}
	agentID := createAgentViaAPI(t, env, testAgent)
	assert.Equal(t, "e2e-crud-agent", agentID)
	t.Logf("Agent created: %s", agentID)

	t.Log("=== Step 2: Read Agent ===")
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	gw := httptest.NewRecorder()
	env.router.ServeHTTP(gw, getReq)

	require.Equal(t, http.StatusOK, gw.Code)
	var getResp Response
	require.NoError(t, json.Unmarshal(gw.Body.Bytes(), &getResp))
	agentData, _ := getResp.Data.(map[string]interface{})
	assert.Equal(t, "E2E CRUD Test Agent", agentData["name"])
	assert.Equal(t, "codex", agentData["adapter"])
	t.Logf("Agent read: name=%s, adapter=%s", agentData["name"], agentData["adapter"])

	t.Log("=== Step 3: Update Agent ===")
	// Update 使用 CreateAgentReq 验证，需要传入所有必需字段
	updateBody, _ := json.Marshal(map[string]interface{}{
		"id":               agentID,
		"name":             "E2E CRUD Test Agent (Updated)",
		"description":      "Updated description",
		"adapter":          "codex",
		"provider_id":      "zhipu",
		"model":            "glm-4.7",
		"base_url_override": "https://open.bigmodel.cn/api/coding/paas/v4",
		"system_prompt":    "Test system prompt",
	})
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+agentID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	uw := httptest.NewRecorder()
	env.router.ServeHTTP(uw, updateReq)

	require.Equal(t, http.StatusOK, uw.Code)
	var updateResp Response
	require.NoError(t, json.Unmarshal(uw.Body.Bytes(), &updateResp))
	updatedData, _ := updateResp.Data.(map[string]interface{})
	assert.Equal(t, "E2E CRUD Test Agent (Updated)", updatedData["name"])
	assert.Equal(t, "Updated description", updatedData["description"])
	t.Logf("Agent updated: name=%s", updatedData["name"])

	t.Log("=== Step 4: List Agents ===")
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	lw := httptest.NewRecorder()
	env.router.ServeHTTP(lw, listReq)

	require.Equal(t, http.StatusOK, lw.Code)
	var listResp Response
	require.NoError(t, json.Unmarshal(lw.Body.Bytes(), &listResp))
	agents, _ := listResp.Data.([]interface{})
	assert.GreaterOrEqual(t, len(agents), 1)
	t.Logf("Agent list count: %d", len(agents))

	t.Log("=== Step 5: Delete Agent ===")
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/"+agentID, nil)
	dw := httptest.NewRecorder()
	env.router.ServeHTTP(dw, deleteReq)

	require.Equal(t, http.StatusOK, dw.Code)
	t.Log("Agent deleted")

	t.Log("=== Step 6: Verify deletion ===")
	verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	vw := httptest.NewRecorder()
	env.router.ServeHTTP(vw, verifyReq)

	assert.Equal(t, http.StatusNotFound, vw.Code, "agent should be deleted")
	t.Log("Deletion verified")

	t.Log("=== PASS: Agent CRUD works correctly ===")
}

// ============================================================
// 测试 4: 多 Agent 并发执行
// ============================================================

// TestAgentE2E_ConcurrentExecution 端到端：创建多个 Agent → 并发执行 → 验证结果
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestAgentE2E_ConcurrentExecution -timeout 600s ./internal/api/
func TestAgentE2E_ConcurrentExecution(t *testing.T) {
	env, cleanup := setupAgentE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create 2 Agents ===")
	agent1 := &agent.Agent{
		ID:              "e2e-concurrent-1",
		Name:            "E2E Concurrent Agent 1",
		Adapter:         agent.AdapterCodex,
		ProviderID:      "zhipu",
		Model:           "glm-4.7",
		BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
		SystemPrompt:    "You are a calculator. Only respond with numbers.",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "danger-full-access",
			FullAuto:       true,
		},
	}
	agent2 := &agent.Agent{
		ID:              "e2e-concurrent-2",
		Name:            "E2E Concurrent Agent 2",
		Adapter:         agent.AdapterCodex,
		ProviderID:      "zhipu",
		Model:           "glm-4.7",
		BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
		SystemPrompt:    "You are a calculator. Only respond with numbers.",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "danger-full-access",
			FullAuto:       true,
		},
	}

	agentID1 := createAgentViaAPI(t, env, agent1)
	agentID2 := createAgentViaAPI(t, env, agent2)
	t.Logf("Agents created: %s, %s", agentID1, agentID2)

	t.Log("=== Step 2: Execute both agents concurrently ===")
	type execResult struct {
		agentID string
		output  string
		status  string
		err     error
	}

	results := make(chan execResult, 2)

	// 启动两个 goroutine 并发执行
	go func() {
		t.Log("Starting agent 1...")
		result := runAgentViaAPI(t, env, agentID1, RunAgentReq{
			Prompt: "What is 10+10? Reply with just the number.",
			Options: &agent.RunOptions{
				MaxTurns: 2,
				Timeout:  60,
			},
		}, 120*time.Second)
		output, _ := result["output"].(string)
		status, _ := result["status"].(string)
		results <- execResult{
			agentID: agentID1,
			output:  output,
			status:  status,
		}
	}()

	go func() {
		t.Log("Starting agent 2...")
		result := runAgentViaAPI(t, env, agentID2, RunAgentReq{
			Prompt: "What is 20+20? Reply with just the number.",
			Options: &agent.RunOptions{
				MaxTurns: 2,
				Timeout:  60,
			},
		}, 120*time.Second)
		output, _ := result["output"].(string)
		status, _ := result["status"].(string)
		results <- execResult{
			agentID: agentID2,
			output:  output,
			status:  status,
		}
	}()

	t.Log("=== Step 3: Wait for results ===")
	var res1, res2 execResult
	for i := 0; i < 2; i++ {
		res := <-results
		t.Logf("Agent %s completed: status=%s, output=%s", res.agentID, res.status, truncate(res.output, 100))
		if res.agentID == agentID1 {
			res1 = res
		} else {
			res2 = res
		}
	}

	t.Log("=== Step 4: Verify results ===")
	assert.Equal(t, "completed", res1.status)
	assert.Equal(t, "completed", res2.status)
	assert.Contains(t, res1.output, "20", "10+10 should equal 20")
	assert.Contains(t, res2.output, "40", "20+20 should equal 40")

	t.Log("=== PASS: Concurrent execution works correctly ===")
}

// ============================================================
// 测试 5: Agent 配置继承验证
// ============================================================

// TestAgentE2E_ConfigInheritance 端到端：验证 Agent 从 Provider 继承配置
//
// 运行方式:
//
//	go test -v -tags=e2e -run TestAgentE2E_ConfigInheritance -timeout 120s ./internal/api/
func TestAgentE2E_ConfigInheritance(t *testing.T) {
	env, cleanup := setupAgentE2E(t, agent.AdapterCodex)
	defer cleanup()

	t.Log("=== Step 1: Create Agent with Provider reference ===")
	testAgent := &agent.Agent{
		ID:          "e2e-config-inherit",
		Name:        "E2E Config Inheritance Agent",
		Description: "Agent for testing config inheritance",
		Adapter:     agent.AdapterCodex,
		ProviderID:  "zhipu", // 引用已配置的 provider
		Model:       "glm-4.7",
		// 不设置 BaseURLOverride，应该从 Provider 继承
		SystemPrompt: "You are a test assistant.",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "read-only",
			FullAuto:       true,
		},
	}
	agentID := createAgentViaAPI(t, env, testAgent)
	t.Logf("Agent created: %s", agentID)

	t.Log("=== Step 2: Get full config ===")
	fullConfig, err := env.agentMgr.GetFullConfig(agentID)
	require.NoError(t, err)

	t.Log("=== Step 3: Verify config inheritance ===")
	assert.Equal(t, "zhipu", fullConfig.Provider.ID, "should inherit provider ID")
	assert.NotEmpty(t, fullConfig.Provider.BaseURL, "should have provider base URL")
	assert.True(t, fullConfig.Provider.IsConfigured, "provider should be configured")
	t.Logf("Provider: id=%s, base_url=%s, configured=%v",
		fullConfig.Provider.ID, fullConfig.Provider.BaseURL, fullConfig.Provider.IsConfigured)

	// 验证环境变量可以获取
	envVars, err := env.agentMgr.GetProviderEnvVars(agentID)
	require.NoError(t, err)
	assert.NotEmpty(t, envVars["OPENAI_API_KEY"], "should have API key in env vars")
	t.Log("Provider env vars verified")

	t.Log("=== PASS: Config inheritance works correctly ===")
}
