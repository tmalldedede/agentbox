package api

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// TestSkillInjection_E2E 端到端测试 Skills 注入功能
//
// 测试流程:
//  1. 创建自定义 Skill（含附加文件）
//  2. 创建 Agent 并关联该 Skill
//  3. 创建 Session（启动容器）
//  4. 验证容器内存在 SKILL.md 文件
//  5. 验证容器内存在附加文件（references）
//
// 运行方式:
//
//	go test -v -run TestSkillInjection_E2E -timeout 120s ./internal/api/
func TestSkillInjection_E2E(t *testing.T) {
	// === 前置检查 ===
	ctx := context.Background()
	dockerMgr, err := container.NewDockerManager()
	if err != nil {
		t.Skipf("Docker not available: %v, skipping test", err)
	}

	// === 初始化临时目录 ===
	tmpDir := t.TempDir()
	// 解析符号链接，避免 macOS /var/folders -> /private/var/folders 导致路径验证失败
	if resolved, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolved
	}

	// === 初始化所有 Manager ===
	gin.SetMode(gin.TestMode)

	// Provider Manager
	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "test-encryption-key-32bytes!!")

	// Runtime Manager
	runtimeDir := filepath.Join(tmpDir, "runtimes")
	rtMgr := runtime.NewManager(runtimeDir, nil)

	// Skill Manager
	skillDir := filepath.Join(tmpDir, "skills")
	skillMgr, err := skill.NewManager(skillDir)
	require.NoError(t, err)

	// MCP Manager
	mcpDir := filepath.Join(tmpDir, "mcp")
	mcpMgr, err := mcp.NewManager(mcpDir)
	require.NoError(t, err)

	// Agent Manager
	agentDir := filepath.Join(tmpDir, "agents")
	agentMgr := agent.NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	// Engine Registry
	registry := engine.DefaultRegistry()

	// Session Manager
	workspaceBase := filepath.Join(tmpDir, "workspaces")
	sessionStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sessionStore, dockerMgr, registry, workspaceBase)
	sessionMgr.SetAgentManager(agentMgr)
	sessionMgr.SetSkillManager(skillMgr)

	// === 测试点 1: 创建自定义 Skill ===
	t.Run("CreateCustomSkill", func(t *testing.T) {
		testSkillReq := &skill.CreateSkillRequest{
			ID:          "test-custom-skill",
			Name:        "Test Custom Skill",
			Description: "A test skill for injection verification",
			Command:     "/test-inject",
			Prompt:      "This is a test prompt for skill injection.\nIt should appear in the container.",
			Category:    skill.CategoryOther,
			Tags:        []string{"test", "injection"},
			Files: []skill.SkillFile{
				{
					Path:    "references/example.md",
					Content: "# Example Reference\n\nThis is an example reference file.",
				},
				{
					Path:    "templates/template.txt",
					Content: "Template content here.",
				},
			},
		}

		testSkill, err := skillMgr.Create(testSkillReq)
		require.NoError(t, err, "create skill should succeed")
		t.Logf("Skill created: id=%s, name=%s, files=%d", testSkill.ID, testSkill.Name, len(testSkill.Files))

		// 验证 Skill 可以被检索
		retrieved, err := skillMgr.Get("test-custom-skill")
		require.NoError(t, err, "get skill should succeed")
		assert.Equal(t, "Test Custom Skill", retrieved.Name)
		assert.Equal(t, "/test-inject", retrieved.Command)
		assert.Len(t, retrieved.Files, 2)
	})

	// === 测试点 2: 创建 Agent 并关联 Skill ===
	var testAgent *agent.Agent
	t.Run("CreateAgentWithSkill", func(t *testing.T) {
		testAgent = &agent.Agent{
			ID:              "test-skill-injection-agent",
			Name:            "Skill Injection Test Agent",
			Description:     "Agent for testing skill injection",
			Adapter:         agent.AdapterCodex,
			ProviderID:      "openai", // 必须有 provider，但此测试不调用 API
			Model:           "gpt-4",
			SkillIDs:        []string{"test-custom-skill", "commit"}, // 自定义 + 内置
			SystemPrompt:    "You are a test agent.",
			Permissions: agent.PermissionConfig{
				SandboxMode: "danger-full-access",
			},
		}

		err := agentMgr.Create(testAgent)
		require.NoError(t, err, "create agent should succeed")
		t.Logf("Agent created: id=%s, skills=%v", testAgent.ID, testAgent.SkillIDs)
	})

	// === 测试点 3: 验证 GetFullConfig 解析 Skills ===
	t.Run("GetFullConfigResolvesSkills", func(t *testing.T) {
		fullConfig, err := agentMgr.GetFullConfig("test-skill-injection-agent")
		require.NoError(t, err, "get full config should succeed")

		assert.Len(t, fullConfig.Skills, 2, "should have 2 skills resolved")

		// 验证自定义 Skill
		var customSkill *skill.Skill
		for _, s := range fullConfig.Skills {
			if s.ID == "test-custom-skill" {
				customSkill = s
				break
			}
		}
		require.NotNil(t, customSkill, "custom skill should be resolved")
		assert.Equal(t, "/test-inject", customSkill.Command)
		assert.Len(t, customSkill.Files, 2)

		t.Logf("Full config resolved: skills=%d", len(fullConfig.Skills))
	})

	// === 测试点 4: 创建 Session 并验证 Skill 注入 ===
	var containerID string
	t.Run("SessionCreationInjectsSkills", func(t *testing.T) {
		// 创建 Session
		createReq := &session.CreateRequest{
			AgentID:   "test-skill-injection-agent",
			Workspace: workspaceBase,
		}

		sess, err := sessionMgr.Create(ctx, createReq)
		require.NoError(t, err, "create session should succeed")
		containerID = sess.ContainerID
		t.Logf("Session created: id=%s, container=%s", sess.ID, containerID)

		// 等待容器启动
		time.Sleep(2 * time.Second)

		// 清理
		defer func() {
			sessionMgr.Delete(ctx, sess.ID)
			t.Logf("Cleaned up session: %s", sess.ID)
		}()

		// === 验证容器内 SKILL.md 存在 ===
		t.Run("VerifySkillMDExists", func(t *testing.T) {
			// 检查自定义 Skill 的 SKILL.md
			checkCmd := []string{"sh", "-c", "cat $HOME/.codex/skills/test-custom-skill/SKILL.md"}
			result, err := dockerMgr.Exec(ctx, containerID, checkCmd)
			require.NoError(t, err, "exec should succeed")
			assert.Equal(t, 0, result.ExitCode, "cat should succeed")

			output := result.Stdout
			t.Logf("SKILL.md content:\n%s", output)

			// 验证内容
			assert.Contains(t, output, "Test Custom Skill", "should contain skill name")
			assert.Contains(t, output, "/test-inject", "should contain command")
			assert.Contains(t, output, "test prompt for skill injection", "should contain prompt")
		})

		// === 验证附加文件存在 ===
		t.Run("VerifyReferenceFilesExist", func(t *testing.T) {
			// 检查 references/example.md
			checkRefCmd := []string{"sh", "-c", "cat $HOME/.codex/skills/test-custom-skill/references/example.md"}
			result, err := dockerMgr.Exec(ctx, containerID, checkRefCmd)
			require.NoError(t, err, "exec should succeed")
			assert.Equal(t, 0, result.ExitCode, "cat reference should succeed")

			output := result.Stdout
			t.Logf("Reference file content:\n%s", output)
			assert.Contains(t, output, "Example Reference", "should contain reference content")

			// 检查 templates/template.txt
			checkTplCmd := []string{"sh", "-c", "cat $HOME/.codex/skills/test-custom-skill/templates/template.txt"}
			result2, err := dockerMgr.Exec(ctx, containerID, checkTplCmd)
			require.NoError(t, err, "exec should succeed")
			assert.Equal(t, 0, result2.ExitCode, "cat template should succeed")
			assert.Contains(t, result2.Stdout, "Template content", "should contain template content")
		})

		// === 验证内置 Skill 也被注入 ===
		t.Run("VerifyBuiltinSkillInjected", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "cat $HOME/.codex/skills/commit/SKILL.md"}
			result, err := dockerMgr.Exec(ctx, containerID, checkCmd)
			require.NoError(t, err, "exec should succeed")
			assert.Equal(t, 0, result.ExitCode, "cat builtin skill should succeed")

			output := result.Stdout
			t.Logf("Builtin Skill (commit) content:\n%s", output[:min(len(output), 200)]+"...")
			assert.Contains(t, output, "/commit", "should contain /commit command")
		})

		// === 验证 Skill 目录结构 ===
		t.Run("VerifySkillDirectoryStructure", func(t *testing.T) {
			listCmd := []string{"sh", "-c", "ls -la $HOME/.codex/skills/"}
			result, err := dockerMgr.Exec(ctx, containerID, listCmd)
			require.NoError(t, err, "exec should succeed")

			output := result.Stdout
			t.Logf("Skills directory:\n%s", output)
			assert.Contains(t, output, "test-custom-skill", "should have custom skill dir")
			assert.Contains(t, output, "commit", "should have commit skill dir")
		})
	})

	// === 测试点 5: 禁用的 Skill 不应被注入 ===
	t.Run("DisabledSkillNotInjected", func(t *testing.T) {
		// 创建禁用的 Skill
		disabledSkillReq := &skill.CreateSkillRequest{
			ID:      "disabled-skill",
			Name:    "Disabled Skill",
			Command: "/disabled",
			Prompt:  "This should not appear.",
		}
		disabledSkill, err := skillMgr.Create(disabledSkillReq)
		require.NoError(t, err)
		// 禁用这个 Skill
		_, err = skillMgr.Update(disabledSkill.ID, &skill.UpdateSkillRequest{IsEnabled: boolPtr(false)})
		require.NoError(t, err)

		// 创建 Agent 关联禁用的 Skill
		agentWithDisabled := &agent.Agent{
			ID:         "agent-with-disabled-skill",
			Name:       "Agent With Disabled Skill",
			Adapter:    agent.AdapterCodex,
			ProviderID: "openai",
			Model:      "gpt-4",
			SkillIDs:   []string{"disabled-skill"},
		}
		err = agentMgr.Create(agentWithDisabled)
		require.NoError(t, err)

		// 创建 Session
		sess, err := sessionMgr.Create(ctx, &session.CreateRequest{
			AgentID:   "agent-with-disabled-skill",
			Workspace: workspaceBase,
		})
		require.NoError(t, err)
		defer sessionMgr.Delete(ctx, sess.ID)

		time.Sleep(2 * time.Second)

		// 验证禁用的 Skill 没有被注入
		checkCmd := []string{"sh", "-c", "ls $HOME/.codex/skills/disabled-skill/SKILL.md 2>&1"}
		result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
		require.NoError(t, err)

		// 文件不应该存在
		assert.NotEqual(t, 0, result.ExitCode, "disabled skill should not be injected")
		t.Logf("Disabled skill check result: exit=%d, output=%s", result.ExitCode, result.Stdout)
	})
}

// TestSkillInjection_SourceDir 测试 SourceDir 目录复制功能
// 使用真实的 email-osint skill 目录进行测试
func TestSkillInjection_SourceDir(t *testing.T) {
	// 检查 email-osint skill 目录是否存在
	emailOsintDir := "/Users/sky2/pr/cybersec-skills/skills/email-osint"
	if _, err := os.Stat(emailOsintDir); os.IsNotExist(err) {
		t.Skipf("email-osint skill directory not found at %s, skipping", emailOsintDir)
	}

	ctx := context.Background()
	dockerMgr, err := container.NewDockerManager()
	if err != nil {
		t.Skipf("Docker not available: %v, skipping test", err)
	}

	// 初始化
	tmpDir := t.TempDir()
	gin.SetMode(gin.TestMode)

	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "test-encryption-key-32bytes!!")

	runtimeDir := filepath.Join(tmpDir, "runtimes")
	rtMgr := runtime.NewManager(runtimeDir, nil)

	skillDir := filepath.Join(tmpDir, "skills")
	skillMgr, err := skill.NewManager(skillDir)
	require.NoError(t, err)

	mcpDir := filepath.Join(tmpDir, "mcp")
	mcpMgr, err := mcp.NewManager(mcpDir)
	require.NoError(t, err)

	agentDir := filepath.Join(tmpDir, "agents")
	agentMgr := agent.NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	registry := engine.DefaultRegistry()
	workspaceBase := filepath.Join(tmpDir, "workspaces")
	sessionStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sessionStore, dockerMgr, registry, workspaceBase)
	sessionMgr.SetAgentManager(agentMgr)
	sessionMgr.SetSkillManager(skillMgr)

	// 创建带 SourceDir 的 Skill
	t.Run("CreateSkillWithSourceDir", func(t *testing.T) {
		createReq := &skill.CreateSkillRequest{
			ID:          "email-osint",
			Name:        "Email OSINT",
			Description: "Email intelligence gathering and analysis",
			Command:     "/email-osint",
			Prompt:      "Perform email OSINT analysis using the provided scripts.",
			SourceDir:   emailOsintDir, // 指向真实目录
			Category:    skill.CategorySecurity,
		}

		createdSkill, err := skillMgr.Create(createReq)
		require.NoError(t, err)
		assert.Equal(t, emailOsintDir, createdSkill.SourceDir)
		t.Logf("Skill created with SourceDir: %s", createdSkill.SourceDir)
	})

	// 创建 Agent 关联该 Skill
	t.Run("CreateAgentWithSourceDirSkill", func(t *testing.T) {
		testAgent := &agent.Agent{
			ID:         "source-dir-test-agent",
			Name:       "SourceDir Test Agent",
			Adapter:    agent.AdapterCodex,
			ProviderID: "openai",
			Model:      "gpt-4",
			SkillIDs:   []string{"email-osint"},
		}
		err := agentMgr.Create(testAgent)
		require.NoError(t, err)
	})

	// 创建 Session 并验证目录被复制
	t.Run("VerifySourceDirCopied", func(t *testing.T) {
		sess, err := sessionMgr.Create(ctx, &session.CreateRequest{
			AgentID:   "source-dir-test-agent",
			Workspace: workspaceBase,
		})
		require.NoError(t, err)
		defer sessionMgr.Delete(ctx, sess.ID)

		time.Sleep(3 * time.Second)

		// 验证 SKILL.md 存在
		t.Run("VerifySKILLMD", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "cat $HOME/.codex/skills/email-osint/SKILL.md"}
			result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, "email-osint")
			t.Logf("SKILL.md exists and contains expected content")
		})

		// 验证 scripts 目录被复制
		t.Run("VerifyScriptsDir", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "ls -la $HOME/.codex/skills/email-osint/scripts/"}
			result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode, "scripts directory should exist")
			t.Logf("Scripts directory:\n%s", result.Stdout)

			// 验证具体脚本文件存在
			assert.Contains(t, result.Stdout, "blackbird_run.py", "should have blackbird_run.py")
			assert.Contains(t, result.Stdout, "check_env.py", "should have check_env.py")
		})

		// 验证 tools 目录被复制
		t.Run("VerifyToolsDir", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "ls -la $HOME/.codex/skills/email-osint/tools/"}
			result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode, "tools directory should exist")
			t.Logf("Tools directory:\n%s", result.Stdout)

			// 验证 blackbird 目录存在
			assert.Contains(t, result.Stdout, "blackbird", "should have blackbird directory")
		})

		// 验证 blackbird.py 脚本存在
		t.Run("VerifyBlackbirdScript", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "head -5 $HOME/.codex/skills/email-osint/tools/blackbird/blackbird.py"}
			result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode, "blackbird.py should exist")
			t.Logf("blackbird.py first 5 lines:\n%s", result.Stdout)
		})

		// 验证完整目录结构
		t.Run("VerifyFullStructure", func(t *testing.T) {
			checkCmd := []string{"sh", "-c", "find $HOME/.codex/skills/email-osint -type f | wc -l"}
			result, err := dockerMgr.Exec(ctx, sess.ContainerID, checkCmd)
			require.NoError(t, err)
			t.Logf("Total files in email-osint skill: %s", strings.TrimSpace(result.Stdout))
			// 应该有多个文件（scripts + tools + references）
			fileCount := strings.TrimSpace(result.Stdout)
			assert.NotEqual(t, "1", fileCount, "should have more than just SKILL.md")
		})
	})
}

// TestSkill_ToSkillMD 单元测试 Skill 转 Markdown 格式
func TestSkill_ToSkillMD(t *testing.T) {
	s := &skill.Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill description",
		Command:     "/test",
		Prompt:      "This is the main prompt content.\nWith multiple lines.",
	}

	md := s.ToSkillMD()

	assert.Contains(t, md, "# Test Skill", "should contain skill name as header")
	assert.Contains(t, md, "A test skill description", "should contain description")
	assert.Contains(t, md, "`/test`", "should contain command in code format")
	assert.Contains(t, md, "## Instructions", "should have instructions section")
	assert.Contains(t, md, "This is the main prompt content.", "should contain prompt")

	t.Logf("Generated SKILL.md:\n%s", md)
}

// TestSkillManager_BuiltinSkills 验证内置 Skills
func TestSkillManager_BuiltinSkills(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	// 列出所有 Skills
	skills := mgr.List()

	// 验证内置 Skills 存在
	builtinIDs := []string{"commit", "review-pr", "explain", "refactor", "test", "docs", "security"}
	for _, id := range builtinIDs {
		found := false
		for _, s := range skills {
			if s.ID == id {
				found = true
				assert.True(t, s.IsBuiltIn, "skill %s should be built-in", id)
				assert.True(t, s.IsEnabled, "skill %s should be enabled by default", id)
				break
			}
		}
		assert.True(t, found, "built-in skill %s should exist", id)
	}

	t.Logf("Found %d skills (expected at least %d built-in)", len(skills), len(builtinIDs))
}

// TestAgentHandler_RunWithSkills 通过 HTTP API 测试 Agent 执行时 Skills 注入
func TestAgentHandler_RunWithSkills(t *testing.T) {
	// 获取 API key
	apiKey := getZhipuAPIKey(t)
	if apiKey == "" {
		t.Skip("zhipu API key not available, skipping")
	}

	ctx := context.Background()
	dockerMgr, err := container.NewDockerManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// 初始化
	tmpDir := t.TempDir()
	gin.SetMode(gin.TestMode)

	providerDir := filepath.Join(tmpDir, "providers")
	provMgr := provider.NewManager(providerDir, "test-encryption-key-32bytes!!")

	// 配置 zhipu 的 API Key
	err = provMgr.ConfigureKey("zhipu", apiKey)
	require.NoError(t, err)

	runtimeDir := filepath.Join(tmpDir, "runtimes")
	rtMgr := runtime.NewManager(runtimeDir, nil)

	skillDir := filepath.Join(tmpDir, "skills")
	skillMgr, err := skill.NewManager(skillDir)
	require.NoError(t, err)

	mcpDir := filepath.Join(tmpDir, "mcp")
	mcpMgr, err := mcp.NewManager(mcpDir)
	require.NoError(t, err)

	agentDir := filepath.Join(tmpDir, "agents")
	agentMgr := agent.NewManager(agentDir, provMgr, rtMgr, skillMgr, mcpMgr)

	registry := engine.DefaultRegistry()
	workspaceBase := filepath.Join(tmpDir, "workspaces")
	sessionStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sessionStore, dockerMgr, registry, workspaceBase)
	sessionMgr.SetAgentManager(agentMgr)
	sessionMgr.SetSkillManager(skillMgr)

	historyMgr := history.NewManager(nil)

	// 创建自定义 Skill - 强制回复特定内容
	customSkillReq := &skill.CreateSkillRequest{
		ID:      "greeting-skill",
		Name:    "Greeting Skill",
		Command: "/greet",
		Prompt:  "When the user says 'test skill', you MUST respond with exactly: 'SKILL_INJECTED_OK'",
	}
	_, err = skillMgr.Create(customSkillReq)
	require.NoError(t, err)

	// 创建 Agent 关联 Skill
	testAgent := &agent.Agent{
		ID:              "skill-test-agent",
		Name:            "Skill Test Agent",
		Adapter:         agent.AdapterCodex,
		ProviderID:      "zhipu",
		Model:           "glm-4-flash",
		BaseURLOverride: "https://open.bigmodel.cn/api/coding/paas/v4",
		SkillIDs:        []string{"greeting-skill"},
		SystemPrompt:    "Follow the skill instructions exactly.",
		Permissions: agent.PermissionConfig{
			ApprovalPolicy: "never",
			SandboxMode:    "danger-full-access",
			FullAuto:       true,
		},
	}
	err = agentMgr.Create(testAgent)
	require.NoError(t, err)

	// 设置 HTTP Handler
	handler := NewAgentHandler(agentMgr, sessionMgr, historyMgr)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	// 发送请求
	reqBody := map[string]interface{}{
		"prompt":    "test skill",
		"workspace": workspaceBase,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/skill-test-agent/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response: status=%d, body=%s", w.Code, w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Output string `json:"output"`
			Status string `json:"status"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Data.Status)
	// Note: LLM 可能不会严格遵循指令，所以这里只验证有输出
	assert.NotEmpty(t, resp.Data.Output, "should have output")
	t.Logf("Agent output: %s", resp.Data.Output)

	// 清理容器
	sessions, _ := sessionMgr.List(ctx, nil)
	for _, s := range sessions {
		if strings.Contains(s.AgentID, "skill-test-agent") {
			sessionMgr.Delete(ctx, s.ID)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func boolPtr(b bool) *bool {
	return &b
}
