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
	"github.com/tmalldedede/agentbox/internal/webhook"
)

func setupWebhookTestRouter(t *testing.T) (*gin.Engine, *WebhookHandler, string) {
	// Create temp directory for test data
	tempDir, err := os.MkdirTemp("", "webhook_test")
	require.NoError(t, err)

	mgr, err := webhook.NewManager(tempDir)
	require.NoError(t, err)

	handler := NewWebhookHandler(mgr)

	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler, tempDir
}

func TestWebhookList(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)
}

func TestWebhookCreate(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := webhook.CreateWebhookRequest{
		URL:    "https://example.com/webhook",
		Secret: "test-secret",
		Events: []string{"task.created", "task.completed"},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	webhookData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, webhookData["id"])
	assert.Equal(t, "https://example.com/webhook", webhookData["url"])
	assert.Equal(t, true, webhookData["is_active"])
}

func TestWebhookGet(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a webhook
	createReq := webhook.CreateWebhookRequest{
		URL: "https://example.com/get-test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	webhookData := createResp.Data.(map[string]interface{})
	webhookID := webhookData["id"].(string)

	// Now get it
	req = httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+webhookID, nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, webhookID, data["id"])
}

func TestWebhookGetNotFound(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/nonexistent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.Code)
}

func TestWebhookUpdate(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a webhook
	createReq := webhook.CreateWebhookRequest{
		URL: "https://example.com/update-test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	webhookData := createResp.Data.(map[string]interface{})
	webhookID := webhookData["id"].(string)

	// Now update it
	isActive := false
	updateReq := webhook.UpdateWebhookRequest{
		URL:      "https://example.com/updated",
		IsActive: &isActive,
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/webhooks/"+webhookID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com/updated", data["url"])
	assert.Equal(t, false, data["is_active"])
}

func TestWebhookDelete(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a webhook
	createReq := webhook.CreateWebhookRequest{
		URL: "https://example.com/delete-test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	webhookData := createResp.Data.(map[string]interface{})
	webhookID := webhookData["id"].(string)

	// Now delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+webhookID, nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+webhookID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWebhookValidation(t *testing.T) {
	router, _, tempDir := setupWebhookTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Missing URL
	createReq := map[string]interface{}{
		"secret": "test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
