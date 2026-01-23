package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/provider"
)

func setupProviderTestRouter(t *testing.T) (*gin.Engine, *ProviderHandler, string) {
	// Create temp directory for test data
	tempDir, err := os.MkdirTemp("", "provider_test")
	require.NoError(t, err)

	mgr := provider.NewManager(tempDir, "test-encryption-key-32bytes!!")
	handler := NewProviderHandler(mgr)

	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler, tempDir
}

func TestProviderList(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	// Should have built-in providers
	providers, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(providers), 5, "Should have at least 5 built-in providers")
}

func TestProviderListFilterByAgent(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		agent       string
		minExpected int
	}{
		{"claude-code", 5},
		{"codex", 2},
		{"opencode", 1},
	}

	for _, tt := range tests {
		t.Run("filter_by_"+tt.agent, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/providers?agent="+tt.agent, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			providers, ok := resp.Data.([]interface{})
			require.True(t, ok)
			assert.GreaterOrEqual(t, len(providers), tt.minExpected,
				"Should have at least %d providers for agent %s", tt.minExpected, tt.agent)
		})
	}
}

func TestProviderListFilterByCategory(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		category    string
		minExpected int
	}{
		{"official", 1},
		{"cn_official", 3},
		{"aggregator", 1},
	}

	for _, tt := range tests {
		t.Run("filter_by_"+tt.category, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/providers?category="+tt.category, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			providers, ok := resp.Data.([]interface{})
			require.True(t, ok)
			assert.GreaterOrEqual(t, len(providers), tt.minExpected,
				"Should have at least %d providers for category %s", tt.minExpected, tt.category)
		})
	}
}

func TestProviderGet(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Get a built-in provider
	req := httptest.NewRequest(http.MethodGet, "/api/v1/providers/anthropic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "anthropic", p["id"])
	assert.Equal(t, "Anthropic", p["name"])
}

func TestProviderGetNotFound(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/providers/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.Code)
}

func TestProviderCreate(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := CreateProviderRequest{
		ID:          "test-provider",
		Name:        "Test Provider",
		Description: "A test provider",
		Agents:      []string{"claude-code"},
		Category:    "third_party",
		BaseURL:     "https://api.test.com",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-provider", p["id"])
	assert.Equal(t, "Test Provider", p["name"])
	assert.Equal(t, false, p["is_built_in"])

	// Verify providers.json was created (all custom providers are saved in one file)
	filePath := filepath.Join(tempDir, "providers.json")
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "Providers file should exist")
}

func TestProviderCreateDuplicate(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Creating a provider with a duplicate ID (e.g., built-in "anthropic") should fail
	createReq := CreateProviderRequest{
		ID:     "anthropic",
		Name:   "Custom Anthropic",
		Agents: []string{"claude-code"},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Duplicate ID should be rejected
	assert.NotEqual(t, http.StatusCreated, w.Code)
}

func TestProviderUpdate(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a custom provider
	createReq := CreateProviderRequest{
		ID:    "update-test",
		Name:  "Update Test",
		Agents: []string{"claude-code"},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now update it
	updateReq := UpdateProviderRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/providers/update-test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Updated Name", p["name"])
	assert.Equal(t, "Updated description", p["description"])
}

func TestProviderDelete(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a custom provider
	createReq := CreateProviderRequest{
		ID:    "delete-test",
		Name:  "Delete Test",
		Agents: []string{"claude-code"},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/providers/delete-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/providers/delete-test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProviderDeleteBuiltIn(t *testing.T) {
	router, _, tempDir := setupProviderTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Try to delete a built-in provider
	// Current implementation silently ignores deletion of built-in providers
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/providers/anthropic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Current implementation returns 200 OK (silently ignores)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	// Verify the built-in provider still exists
	req = httptest.NewRequest(http.MethodGet, "/api/v1/providers/anthropic", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
