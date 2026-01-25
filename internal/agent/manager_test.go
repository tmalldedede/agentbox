package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// setupTestManager 创建一个完整的测试 Manager（包含 provider/runtime/skill/mcp 依赖）
func setupTestManager(t *testing.T) (*Manager, string) {
	t.Helper()

	tmpDir := t.TempDir()

	// 1. 创建 Provider Manager（包含 zhipu 等内置 providers）
	providerDir := filepath.Join(tmpDir, "providers")
	os.MkdirAll(providerDir, 0755)
	provMgr := provider.NewManager(providerDir, "test-encryption-key-1234")

	// 2. 创建 Runtime Manager（包含 default runtime）
	runtimeDir := filepath.Join(tmpDir, "runtimes")
	os.MkdirAll(runtimeDir, 0755)
	rtMgr := runtime.NewManager(runtimeDir, nil)

	// 3. 创建 Skill Manager
	skillDir := filepath.Join(tmpDir, "skills")
	skillMgr, err := skill.NewManager(skillDir)
	if err != nil {
		t.Fatalf("failed to create skill manager: %v", err)
	}

	// 4. 创建 MCP Manager
	mcpDir := filepath.Join(tmpDir, "mcp")
	mcpMgr, err2 := mcp.NewManager(mcpDir)
	if err2 != nil {
		t.Fatalf("failed to create mcp manager: %v", err2)
	}

	// 5. 创建 Agent Manager
	agentDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentDir, 0755)
	mgr := NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	return mgr, tmpDir
}

// TestCreateAgent_CodexZhipu 完整流程测试：用 codex adapter + zhipu provider 新建 Agent
func TestCreateAgent_CodexZhipu(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// 验证 zhipu provider 已加载（内置）
	zhipu, err := mgr.providerMgr.Get("zhipu")
	if err != nil {
		t.Fatalf("zhipu provider should be built-in, got error: %v", err)
	}
	t.Logf("zhipu provider: name=%s, agents=%v, base_url=%s", zhipu.Name, zhipu.Agents, zhipu.BaseURL)

	// 验证 default runtime 存在
	defaultRT := mgr.runtimeMgr.GetDefault()
	if defaultRT == nil {
		t.Fatal("default runtime should exist")
	}
	t.Logf("default runtime: id=%s, image=%s, cpus=%.1f, memory=%dMB",
		defaultRT.ID, defaultRT.Image, defaultRT.CPUs, defaultRT.MemoryMB)

	// === 创建 Agent ===
	agent := &Agent{
		ID:          "test-codex-zhipu",
		Name:        "Test Codex Agent",
		Description: "A test agent using Codex with Zhipu provider",
		Icon:        "🤖",
		Adapter:     AdapterCodex,
		ProviderID:  "zhipu",
		Model:       "glm-4-plus",
		SystemPrompt: "You are a helpful coding assistant.",
		Permissions: PermissionConfig{
			ApprovalPolicy: "on-failure",
			SandboxMode:    "workspace-write",
		},
	}

	err = mgr.Create(agent)
	if err != nil {
		t.Fatalf("Create agent failed: %v", err)
	}

	// 验证默认值被正确设置
	if agent.Status != StatusActive {
		t.Errorf("expected status=%s, got %s", StatusActive, agent.Status)
	}
	if agent.APIAccess != APIAccessPrivate {
		t.Errorf("expected api_access=%s, got %s", APIAccessPrivate, agent.APIAccess)
	}
	if agent.RuntimeID != "default" {
		t.Errorf("expected runtime_id=default, got %s", agent.RuntimeID)
	}
	if agent.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}

	t.Logf("Agent created: id=%s, adapter=%s, provider=%s, runtime=%s, status=%s",
		agent.ID, agent.Adapter, agent.ProviderID, agent.RuntimeID, agent.Status)

	// === 获取 Agent ===
	got, err := mgr.Get("test-codex-zhipu")
	if err != nil {
		t.Fatalf("Get agent failed: %v", err)
	}
	if got.Name != "Test Codex Agent" {
		t.Errorf("expected name='Test Codex Agent', got '%s'", got.Name)
	}
	if got.Model != "glm-4-plus" {
		t.Errorf("expected model='glm-4-plus', got '%s'", got.Model)
	}

	// === GetFullConfig 解析所有引用 ===
	fullCfg, err := mgr.GetFullConfig("test-codex-zhipu")
	if err != nil {
		t.Fatalf("GetFullConfig failed: %v", err)
	}
	if fullCfg.Provider == nil {
		t.Fatal("resolved provider should not be nil")
	}
	if fullCfg.Provider.ID != "zhipu" {
		t.Errorf("resolved provider id=%s, expected zhipu", fullCfg.Provider.ID)
	}
	if fullCfg.Runtime == nil {
		t.Fatal("resolved runtime should not be nil")
	}
	if fullCfg.Runtime.ID != "default" {
		t.Errorf("resolved runtime id=%s, expected default", fullCfg.Runtime.ID)
	}

	t.Logf("FullConfig resolved: provider=%s(%s), runtime=%s(cpu=%.1f, mem=%dMB)",
		fullCfg.Provider.Name, fullCfg.Provider.BaseURL,
		fullCfg.Runtime.Name, fullCfg.Runtime.CPUs, fullCfg.Runtime.MemoryMB)

	// === 更新 Agent ===
	updated := &Agent{
		ID:          "test-codex-zhipu",
		Name:        "Updated Codex Agent",
		Adapter:     AdapterCodex,
		ProviderID:  "zhipu",
		Model:       "glm-4-flash",
		RuntimeID:   "default",
		SystemPrompt: "You are an expert Go developer.",
		Status:      StatusActive,
	}
	err = mgr.Update(updated)
	if err != nil {
		t.Fatalf("Update agent failed: %v", err)
	}

	got, _ = mgr.Get("test-codex-zhipu")
	if got.Name != "Updated Codex Agent" {
		t.Errorf("expected updated name, got '%s'", got.Name)
	}
	if got.Model != "glm-4-flash" {
		t.Errorf("expected model='glm-4-flash', got '%s'", got.Model)
	}

	// === 列表 ===
	agents := mgr.List()
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	// === 删除 ===
	err = mgr.Delete("test-codex-zhipu")
	if err != nil {
		t.Fatalf("Delete agent failed: %v", err)
	}
	agents = mgr.List()
	if len(agents) != 0 {
		t.Errorf("expected 0 agents after delete, got %d", len(agents))
	}

	// 确认 Get 也返回 not found
	_, err = mgr.Get("test-codex-zhipu")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}

// TestCreateAgent_ValidationErrors 测试各种验证失败场景
func TestCreateAgent_ValidationErrors(t *testing.T) {
	mgr, _ := setupTestManager(t)

	tests := []struct {
		name    string
		agent   *Agent
		wantErr error
	}{
		{
			name:    "missing ID",
			agent:   &Agent{Name: "Test", Adapter: AdapterCodex, ProviderID: "zhipu"},
			wantErr: ErrAgentIDRequired,
		},
		{
			name:    "missing name",
			agent:   &Agent{ID: "test", Adapter: AdapterCodex, ProviderID: "zhipu"},
			wantErr: ErrAgentNameRequired,
		},
		{
			name:    "missing adapter",
			agent:   &Agent{ID: "test", Name: "Test", ProviderID: "zhipu"},
			wantErr: ErrAgentAdapterRequired,
		},
		{
			name:    "invalid adapter",
			agent:   &Agent{ID: "test", Name: "Test", Adapter: "invalid", ProviderID: "zhipu"},
			wantErr: ErrAgentInvalidAdapter,
		},
		{
			name:    "missing provider_id",
			agent:   &Agent{ID: "test", Name: "Test", Adapter: AdapterCodex},
			wantErr: ErrAgentProviderRequired,
		},
		{
			name:    "non-existent provider",
			agent:   &Agent{ID: "test", Name: "Test", Adapter: AdapterCodex, ProviderID: "non-existent"},
			wantErr: ErrProviderNotFound,
		},
		{
			name:    "non-existent runtime",
			agent:   &Agent{ID: "test", Name: "Test", Adapter: AdapterCodex, ProviderID: "zhipu", RuntimeID: "non-existent"},
			wantErr: ErrRuntimeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := mgr.Create(tc.agent)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err != tc.wantErr {
				t.Errorf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestCreateAgent_DuplicateID 测试重复 ID
func TestCreateAgent_DuplicateID(t *testing.T) {
	mgr, _ := setupTestManager(t)

	agent := &Agent{
		ID:         "dup-test",
		Name:       "First",
		Adapter:    AdapterCodex,
		ProviderID: "zhipu",
	}
	if err := mgr.Create(agent); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// 尝试用同一 ID 再创建
	agent2 := &Agent{
		ID:         "dup-test",
		Name:       "Second",
		Adapter:    AdapterCodex,
		ProviderID: "zhipu",
	}
	err := mgr.Create(agent2)
	if err != ErrAgentAlreadyExists {
		t.Errorf("expected ErrAgentAlreadyExists, got %v", err)
	}
}

// TestCreateAgent_Persistence 测试持久化：创建后重新加载 Manager 应保留数据
func TestCreateAgent_Persistence(t *testing.T) {
	mgr, tmpDir := setupTestManager(t)

	agent := &Agent{
		ID:           "persist-test",
		Name:         "Persistent Agent",
		Adapter:      AdapterCodex,
		ProviderID:   "zhipu",
		Model:        "glm-4-plus",
		SystemPrompt: "Be helpful",
	}
	if err := mgr.Create(agent); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 验证 agents.json 文件存在
	jsonPath := filepath.Join(tmpDir, "agents", "agents.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Fatal("agents.json should exist after create")
	}

	// 创建新的 Manager（模拟重启），验证数据加载
	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "test-encryption-key-1234")
	runtimeDir := filepath.Join(tmpDir, "runtimes")
	rtMgr := runtime.NewManager(runtimeDir, nil)
	skillDir := filepath.Join(tmpDir, "skills")
	skillMgr, _ := skill.NewManager(skillDir)
	mcpDir := filepath.Join(tmpDir, "mcp")
	mcpMgr, _ := mcp.NewManager(mcpDir)

	agentDir := filepath.Join(tmpDir, "agents")
	mgr2 := NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	got, err := mgr2.Get("persist-test")
	if err != nil {
		t.Fatalf("Get after reload failed: %v", err)
	}
	if got.Name != "Persistent Agent" {
		t.Errorf("expected name='Persistent Agent', got '%s'", got.Name)
	}
	if got.Model != "glm-4-plus" {
		t.Errorf("expected model='glm-4-plus', got '%s'", got.Model)
	}
	if got.SystemPrompt != "Be helpful" {
		t.Errorf("expected system_prompt='Be helpful', got '%s'", got.SystemPrompt)
	}
}

// TestCreateAgent_WithSkillsAndMCP 测试带 Skills 和 MCP 引用的 Agent
func TestCreateAgent_WithSkillsAndMCP(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// 内置 skills 已加载（commit, review-pr 等）
	agent := &Agent{
		ID:           "full-featured",
		Name:         "Full Featured Agent",
		Adapter:      AdapterCodex,
		ProviderID:   "zhipu",
		Model:        "glm-4-plus",
		SkillIDs:     []string{"commit", "review-pr", "non-existent-skill"},
		MCPServerIDs: []string{"non-existent-mcp"},
		SystemPrompt: "You are a senior developer.",
		Permissions: PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "danger-full-access",
			FullAuto:       true,
		},
		Env: map[string]string{
			"CUSTOM_VAR": "custom_value",
		},
	}

	if err := mgr.Create(agent); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// GetFullConfig 应该跳过不存在的 skill/mcp，不报错
	fullCfg, err := mgr.GetFullConfig("full-featured")
	if err != nil {
		t.Fatalf("GetFullConfig failed: %v", err)
	}

	// 只有 commit 和 review-pr 应该被解析成功
	if len(fullCfg.Skills) != 2 {
		t.Errorf("expected 2 resolved skills (commit, review-pr), got %d", len(fullCfg.Skills))
	}
	for _, s := range fullCfg.Skills {
		t.Logf("  resolved skill: %s (%s)", s.ID, s.Name)
	}

	// MCP 全部不存在，应为空
	if len(fullCfg.MCPServers) != 0 {
		t.Errorf("expected 0 resolved MCP servers, got %d", len(fullCfg.MCPServers))
	}
}

// TestCreateAgent_AllAdapters 测试所有三种 adapter
func TestCreateAgent_AllAdapters(t *testing.T) {
	mgr, _ := setupTestManager(t)

	adapters := []struct {
		adapter    string
		providerID string
	}{
		{AdapterClaudeCode, "zhipu"},
		{AdapterCodex, "zhipu"},
		{AdapterOpenCode, "zhipu"},
	}

	for _, tc := range adapters {
		t.Run(tc.adapter, func(t *testing.T) {
			agent := &Agent{
				ID:         "agent-" + tc.adapter,
				Name:       "Agent " + tc.adapter,
				Adapter:    tc.adapter,
				ProviderID: tc.providerID,
			}
			if err := mgr.Create(agent); err != nil {
				t.Fatalf("create with adapter=%s failed: %v", tc.adapter, err)
			}

			got, _ := mgr.Get(agent.ID)
			if got.Adapter != tc.adapter {
				t.Errorf("expected adapter=%s, got %s", tc.adapter, got.Adapter)
			}
		})
	}

	// 应有 3 个 agents
	if len(mgr.List()) != 3 {
		t.Errorf("expected 3 agents, got %d", len(mgr.List()))
	}
}

// TestGetProviderEnvVars 测试获取 Provider 环境变量
func TestGetProviderEnvVars(t *testing.T) {
	mgr, _ := setupTestManager(t)

	agent := &Agent{
		ID:         "env-test",
		Name:       "Env Test Agent",
		Adapter:    AdapterCodex,
		ProviderID: "zhipu",
	}
	if err := mgr.Create(agent); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 未配置 API Key 时，GetProviderEnvVars 应该返回错误或空
	envVars, err := mgr.GetProviderEnvVars("env-test")
	if err != nil {
		// 未配置 key 时预期会报错
		t.Logf("GetProviderEnvVars without configured key: %v (expected)", err)
	} else {
		t.Logf("GetProviderEnvVars: %v", envVars)
	}

	// 不存在的 agent
	_, err = mgr.GetProviderEnvVars("non-existent")
	if err != ErrAgentNotFound {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}
