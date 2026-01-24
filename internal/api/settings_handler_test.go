package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSettingsTestRouter(t *testing.T) (*gin.Engine, *settings.Manager, func()) {
	// Create temp DB
	tempFile, err := os.CreateTemp("", "settings_test_*.db")
	require.NoError(t, err)
	tempFile.Close()

	db, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
	require.NoError(t, err)

	mgr, err := settings.NewManager(db)
	require.NoError(t, err)

	handler := NewSettingsHandler(mgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1/admin")
	handler.RegisterRoutes(v1)

	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return router, mgr, cleanup
}

func TestSettingsGetAll(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	// Should have all setting categories
	assert.Contains(t, data, "agent")
	assert.Contains(t, data, "task")
	assert.Contains(t, data, "batch")
	assert.Contains(t, data, "storage")
	assert.Contains(t, data, "notify")
}

func TestSettingsUpdateAll(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	// Get current settings first
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var getResp Response
	json.Unmarshal(w.Body.Bytes(), &getResp)
	currentSettings := getResp.Data.(map[string]interface{})

	// Modify and update
	currentSettings["agent"].(map[string]interface{})["default_timeout"] = float64(120)

	body, _ := json.Marshal(currentSettings)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestSettingsReset(t *testing.T) {
	router, mgr, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	// Modify settings first
	s := mgr.Get()
	s.Agent.DefaultTimeout = 999
	_ = mgr.Update(s)

	// Reset
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/reset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// Verify reset to default
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	agentSettings := data["agent"].(map[string]interface{})
	// Should be back to default (3600 seconds)
	assert.NotEqual(t, float64(999), agentSettings["default_timeout"])
}

func TestSettingsGetAgent(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/agent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "default_timeout")
	assert.Contains(t, data, "default_provider_id")
}

func TestSettingsUpdateAgent(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	updateReq := settings.AgentSettings{
		DefaultTimeout:    180,
		DefaultProviderID: "test-provider",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/agent", bytes.NewReader(body))
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
	assert.Equal(t, float64(180), data["default_timeout"])
}

func TestSettingsGetTask(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/task", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestSettingsUpdateTask(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	updateReq := settings.TaskSettings{
		DefaultIdleTimeout:  600,
		DefaultPollInterval: 1000,
		MaxTurns:            20,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/task", bytes.NewReader(body))
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
	assert.Equal(t, float64(600), data["default_idle_timeout"])
}

func TestSettingsGetBatch(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/batch", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestSettingsUpdateBatch(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	updateReq := settings.BatchSettings{
		DefaultWorkers: 5,
		MaxWorkers:     20,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/batch", bytes.NewReader(body))
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
	assert.Equal(t, float64(5), data["default_workers"])
}

func TestSettingsGetStorage(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/storage", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestSettingsUpdateStorage(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	updateReq := settings.StorageSettings{
		HistoryRetentionDays: 60,
		SessionRetentionDays: 14,
		AutoCleanup:          true,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/storage", bytes.NewReader(body))
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
	assert.Equal(t, float64(60), data["history_retention_days"])
}

func TestSettingsGetNotify(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/notify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestSettingsUpdateNotify(t *testing.T) {
	router, _, cleanup := setupSettingsTestRouter(t)
	defer cleanup()

	updateReq := settings.NotifySettings{
		WebhookURL:       "https://example.com/webhook",
		NotifyOnComplete: true,
		NotifyOnFailed:   true,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/notify", bytes.NewReader(body))
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
	assert.Equal(t, "https://example.com/webhook", data["webhook_url"])
	assert.Equal(t, true, data["notify_on_complete"])
}
