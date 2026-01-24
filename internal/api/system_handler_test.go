package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/batch"
	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/session"
)

// MockContainerManager for testing
type MockContainerManager struct {
	containers []*container.Container
	images     []*container.Image
	pingErr    error
}

func (m *MockContainerManager) Create(ctx context.Context, config *container.CreateConfig) (*container.Container, error) {
	return &container.Container{ID: "mock-container-id"}, nil
}

func (m *MockContainerManager) Start(ctx context.Context, id string) error {
	return nil
}

func (m *MockContainerManager) Stop(ctx context.Context, id string) error {
	return nil
}

func (m *MockContainerManager) Remove(ctx context.Context, id string) error {
	// Remove from mock list
	for i, c := range m.containers {
		if c.ID == id {
			m.containers = append(m.containers[:i], m.containers[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockContainerManager) Exec(ctx context.Context, id string, cmd []string) (*container.ExecResult, error) {
	return &container.ExecResult{ExitCode: 0, Stdout: "mock output"}, nil
}

func (m *MockContainerManager) ExecStream(ctx context.Context, id string, cmd []string) (*container.ExecStream, error) {
	return &container.ExecStream{}, nil
}

func (m *MockContainerManager) CopyToContainer(ctx context.Context, containerID, srcPath, dstPath string) error {
	return nil
}

func (m *MockContainerManager) Inspect(ctx context.Context, id string) (*container.Container, error) {
	return &container.Container{ID: id, Status: container.StatusRunning}, nil
}

func (m *MockContainerManager) ListContainers(ctx context.Context) ([]*container.Container, error) {
	return m.containers, nil
}

func (m *MockContainerManager) ListImages(ctx context.Context) ([]*container.Image, error) {
	return m.images, nil
}

func (m *MockContainerManager) PullImage(ctx context.Context, imageName string) error {
	return nil
}

func (m *MockContainerManager) RemoveImage(ctx context.Context, id string) error {
	for i, img := range m.images {
		if img.ID == id {
			m.images = append(m.images[:i], m.images[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockContainerManager) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *MockContainerManager) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock logs")), nil
}

func (m *MockContainerManager) Close() error {
	return nil
}

// MockSessionLister implements container.SessionLister for testing
type MockSessionLister struct {
	containerIDs []string
}

func (m *MockSessionLister) ListContainerIDs(ctx context.Context) ([]string, error) {
	return m.containerIDs, nil
}

func setupSystemTestRouter(t *testing.T) (*gin.Engine, *SystemHandler, *MockContainerManager, *container.GarbageCollector) {
	mockContainer := &MockContainerManager{
		containers: []*container.Container{
			{ID: "container123456789001", Name: "test-1", Status: container.StatusRunning},
			{ID: "container123456789002", Name: "test-2", Status: container.StatusExited},
		},
		images: []*container.Image{
			{ID: "image123456789001", Tags: []string{"test:v1"}, Size: 100 * 1024 * 1024, IsAgentImage: true},
			{ID: "image123456789002", Tags: []string{"test:v2"}, Size: 200 * 1024 * 1024, InUse: true},
		},
	}

	mockSessionLister := &MockSessionLister{
		containerIDs: []string{"container123456789001"}, // container123456789001 is active, container123456789002 is orphan
	}

	// Create a properly initialized session.Manager with memory store
	sessionStore := session.NewMemoryStore()
	sessionMgr := session.NewManager(sessionStore, mockContainer, nil, t.TempDir())

	mockBatch := &batch.Manager{}

	gcConfig := container.GCConfig{
		Interval:     5 * time.Minute,
		ContainerTTL: time.Hour,
		IdleTimeout:  30 * time.Minute,
	}
	gc := container.NewGarbageCollector(mockContainer, mockSessionLister, gcConfig)

	handler := NewSystemHandler(mockContainer, sessionMgr, mockBatch, gc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1/admin")
	handler.RegisterRoutes(v1)

	return router, handler, mockContainer, gc
}

func TestSystemHealth(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "healthy", data["status"])
	assert.Contains(t, data, "uptime")
	assert.Contains(t, data, "docker")
	assert.Contains(t, data, "resources")
	assert.Contains(t, data, "checks")
}

func TestSystemStats(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "sessions")
	assert.Contains(t, data, "containers")
	assert.Contains(t, data, "images")
	assert.Contains(t, data, "system")
}

func TestSystemCleanupContainers(t *testing.T) {
	router, _, mockContainer, _ := setupSystemTestRouter(t)

	// Add an orphan container (not in any session)
	mockContainer.containers = append(mockContainer.containers, &container.Container{
		ID:     "orphancontainer123456789",
		Name:   "orphan",
		Status: container.StatusExited,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/cleanup/containers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "removed")
}

func TestSystemCleanupImages(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	cleanupReq := CleanupImagesRequest{
		UnusedOnly: true,
	}

	body, _ := json.Marshal(cleanupReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/cleanup/images", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "removed")
	assert.Contains(t, data, "space_freed")
}

func TestSystemGCStats(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/system/gc/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "running")
	assert.Contains(t, data, "containers_removed")
	assert.Contains(t, data, "total_runs")
	assert.Contains(t, data, "config")
}

func TestSystemGCTrigger(t *testing.T) {
	router, _, _, gc := setupSystemTestRouter(t)

	initialRuns := gc.Stats().TotalRuns

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/gc/trigger", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "removed")

	// Verify GC was triggered
	assert.Equal(t, initialRuns+1, gc.Stats().TotalRuns)
}

func TestSystemGCPreview(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/gc/preview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// Data should be an array (may be empty or have candidates)
	candidates, ok := resp.Data.([]interface{})
	require.True(t, ok)
	// container-2 is not in active list, so it should be a candidate
	assert.GreaterOrEqual(t, len(candidates), 1)
}

func TestSystemGCUpdateConfig(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	interval := 120
	ttl := 3600
	idle := 600

	updateReq := UpdateGCConfigRequest{
		IntervalSeconds:     &interval,
		ContainerTTLSeconds: &ttl,
		IdleTimeoutSeconds:  &idle,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/system/gc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "config")
}

func TestSystemGCUpdateConfigValidation(t *testing.T) {
	router, _, _, _ := setupSystemTestRouter(t)

	tests := []struct {
		name   string
		config UpdateGCConfigRequest
	}{
		{
			name: "interval_too_small",
			config: UpdateGCConfigRequest{
				IntervalSeconds: ptrInt(5), // < 10
			},
		},
		{
			name: "ttl_too_small",
			config: UpdateGCConfigRequest{
				ContainerTTLSeconds: ptrInt(30), // < 60
			},
		},
		{
			name: "idle_too_small",
			config: UpdateGCConfigRequest{
				IdleTimeoutSeconds: ptrInt(10), // < 30
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.config)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/system/gc/config", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Helper function
func ptrInt(i int) *int {
	return &i
}
