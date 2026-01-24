package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/history"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/session"
	"github.com/tmalldedede/agentbox/internal/skill"
	"github.com/tmalldedede/agentbox/internal/task"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockContainerManagerForDashboard implements container.Manager for dashboard tests
type MockContainerManagerForDashboard struct {
	containers []*container.Container
}

func (m *MockContainerManagerForDashboard) Create(ctx context.Context, config *container.CreateConfig) (*container.Container, error) {
	return &container.Container{ID: "mock-id"}, nil
}
func (m *MockContainerManagerForDashboard) Start(ctx context.Context, id string) error   { return nil }
func (m *MockContainerManagerForDashboard) Stop(ctx context.Context, id string) error    { return nil }
func (m *MockContainerManagerForDashboard) Remove(ctx context.Context, id string) error  { return nil }
func (m *MockContainerManagerForDashboard) Exec(ctx context.Context, id string, cmd []string) (*container.ExecResult, error) {
	return &container.ExecResult{ExitCode: 0}, nil
}
func (m *MockContainerManagerForDashboard) ExecStream(ctx context.Context, id string, cmd []string) (*container.ExecStream, error) {
	return &container.ExecStream{}, nil
}
func (m *MockContainerManagerForDashboard) CopyToContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	return nil
}
func (m *MockContainerManagerForDashboard) Inspect(ctx context.Context, id string) (*container.Container, error) {
	return &container.Container{ID: id, Status: container.StatusRunning}, nil
}
func (m *MockContainerManagerForDashboard) ListContainers(ctx context.Context) ([]*container.Container, error) {
	return m.containers, nil
}
func (m *MockContainerManagerForDashboard) ListImages(ctx context.Context) ([]*container.Image, error) {
	return nil, nil
}
func (m *MockContainerManagerForDashboard) PullImage(ctx context.Context, imageName string) error { return nil }
func (m *MockContainerManagerForDashboard) RemoveImage(ctx context.Context, id string) error { return nil }
func (m *MockContainerManagerForDashboard) Ping(ctx context.Context) error                   { return nil }
func (m *MockContainerManagerForDashboard) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (m *MockContainerManagerForDashboard) Close() error { return nil }

func setupDashboardTestRouter(t *testing.T) (*gin.Engine, *DashboardHandler, func()) {
	// Create temp directories and DB
	tempDir, err := os.MkdirTemp("", "dashboard_test")
	require.NoError(t, err)

	tempDB, err := os.CreateTemp("", "dashboard_test_*.db")
	require.NoError(t, err)
	tempDB.Close()

	db, err := gorm.Open(sqlite.Open(tempDB.Name()), &gorm.Config{})
	require.NoError(t, err)

	// Setup managers
	taskStore, err := task.NewGormStore(db)
	require.NoError(t, err)
	providerMgr := provider.NewManager(tempDir, "test-encryption-key-32bytes!!")
	runtimeMgr := runtime.NewManager(tempDir)
	skillMgr, _ := skill.NewManager(tempDir)
	mcpMgr, _ := mcp.NewManager(tempDir)
	agentMgr := agent.NewManager(tempDir, providerMgr, runtimeMgr, skillMgr, mcpMgr)

	mockContainer := &MockContainerManagerForDashboard{
		containers: []*container.Container{
			{ID: "container123456789001", Status: container.StatusRunning},
			{ID: "container123456789002", Status: container.StatusExited},
		},
	}

	sessionStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sessionStore, mockContainer, nil, tempDir)

	taskMgr := task.NewManager(taskStore, agentMgr, sessionMgr, nil)
	historyStore := history.NewMemoryStore()
	historyMgr := history.NewManager(historyStore)

	handler := NewDashboardHandler(taskMgr, agentMgr, sessionMgr, providerMgr, mcpMgr, mockContainer, historyMgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1/admin")
	handler.RegisterRoutes(v1)

	cleanup := func() {
		os.RemoveAll(tempDir)
		os.Remove(tempDB.Name())
	}

	return router, handler, cleanup
}

func TestDashboardStats(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	// Verify all sections are present
	assert.Contains(t, data, "agents")
	assert.Contains(t, data, "tasks")
	assert.Contains(t, data, "sessions")
	assert.Contains(t, data, "tokens")
	assert.Contains(t, data, "containers")
	assert.Contains(t, data, "mcp_servers")
	assert.Contains(t, data, "providers")
	assert.Contains(t, data, "system")
	assert.Contains(t, data, "recent_tasks")
}

func TestDashboardStatsAgentSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	agents := data["agents"].(map[string]interface{})

	assert.Contains(t, agents, "total")
	assert.Contains(t, agents, "active")
	assert.Contains(t, agents, "by_adapter")
	assert.Contains(t, agents, "details")
}

func TestDashboardStatsTaskSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	tasks := data["tasks"].(map[string]interface{})

	assert.Contains(t, tasks, "total")
	assert.Contains(t, tasks, "today")
	assert.Contains(t, tasks, "by_status")
	assert.Contains(t, tasks, "avg_duration_seconds")
	assert.Contains(t, tasks, "success_rate")
}

func TestDashboardStatsSessionSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	sessions := data["sessions"].(map[string]interface{})

	assert.Contains(t, sessions, "total")
	assert.Contains(t, sessions, "running")
	assert.Contains(t, sessions, "creating")
	assert.Contains(t, sessions, "stopped")
	assert.Contains(t, sessions, "error")
}

func TestDashboardStatsContainerSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	containers := data["containers"].(map[string]interface{})

	assert.Contains(t, containers, "total")
	assert.Contains(t, containers, "running")
	assert.Contains(t, containers, "stopped")

	// Should have 2 containers from mock
	assert.Equal(t, float64(2), containers["total"])
	assert.Equal(t, float64(1), containers["running"])
	assert.Equal(t, float64(1), containers["stopped"])
}

func TestDashboardStatsMCPSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	mcpServers := data["mcp_servers"].(map[string]interface{})

	assert.Contains(t, mcpServers, "total")
	assert.Contains(t, mcpServers, "enabled")
	assert.Contains(t, mcpServers, "configured")
	assert.Contains(t, mcpServers, "by_category")
	assert.Contains(t, mcpServers, "details")
}

func TestDashboardStatsProviderSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	providers, ok := data["providers"].([]interface{})
	require.True(t, ok)

	// Should have built-in providers
	assert.GreaterOrEqual(t, len(providers), 1)

	if len(providers) > 0 {
		firstProvider := providers[0].(map[string]interface{})
		assert.Contains(t, firstProvider, "id")
		assert.Contains(t, firstProvider, "name")
		assert.Contains(t, firstProvider, "status")
		assert.Contains(t, firstProvider, "is_configured")
	}
}

func TestDashboardStatsSystemSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	system := data["system"].(map[string]interface{})

	assert.Contains(t, system, "uptime")
	assert.Contains(t, system, "started_at")
}

func TestDashboardStatsTokenSection(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	tokens := data["tokens"].(map[string]interface{})

	assert.Contains(t, tokens, "total_input")
	assert.Contains(t, tokens, "total_output")
	assert.Contains(t, tokens, "total_tokens")
}

func TestDashboardStatsRecentTasks(t *testing.T) {
	router, _, cleanup := setupDashboardTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	recentTasks, ok := data["recent_tasks"].([]interface{})
	require.True(t, ok)

	// Should be an empty array or have tasks
	assert.NotNil(t, recentTasks)
}
