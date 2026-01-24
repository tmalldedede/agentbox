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
	"github.com/tmalldedede/agentbox/internal/runtime"
)

func setupRuntimeTestRouter(t *testing.T) (*gin.Engine, *RuntimeHandler, string) {
	tempDir, err := os.MkdirTemp("", "runtime_test")
	require.NoError(t, err)

	mgr := runtime.NewManager(tempDir)
	handler := NewRuntimeHandler(mgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler, tempDir
}

func TestRuntimeList(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtimes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// Should have default runtime
	runtimes, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(runtimes), 1, "Should have at least 1 default runtime")
}

func TestRuntimeCreate(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := CreateRuntimeRequest{
		ID:          "test-runtime",
		Name:        "Test Runtime",
		Description: "A test runtime",
		Image:       "ubuntu:22.04",
		CPUs:        2.0,
		MemoryMB:    1024,
		Network:     "bridge",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-runtime", data["id"])
	assert.Equal(t, "Test Runtime", data["name"])
	assert.Equal(t, "ubuntu:22.04", data["image"])
}

func TestRuntimeGet(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a runtime
	createReq := CreateRuntimeRequest{
		ID:    "get-test",
		Name:  "Get Test",
		Image: "alpine:latest",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now get it
	req = httptest.NewRequest(http.MethodGet, "/api/v1/runtimes/get-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "get-test", data["id"])
}

func TestRuntimeGetNotFound(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtimes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRuntimeUpdate(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a runtime
	createReq := CreateRuntimeRequest{
		ID:    "update-test",
		Name:  "Update Test",
		Image: "alpine:latest",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now update it
	updateReq := UpdateRuntimeRequest{
		Name:        "Updated Name",
		Description: "Updated description",
		MemoryMB:    2048,
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/runtimes/update-test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Updated Name", data["name"])
	assert.Equal(t, "Updated description", data["description"])
}

func TestRuntimeDelete(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a runtime
	createReq := CreateRuntimeRequest{
		ID:    "delete-test",
		Name:  "Delete Test",
		Image: "alpine:latest",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/runtimes/delete-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/runtimes/delete-test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRuntimeSetDefault(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a runtime
	createReq := CreateRuntimeRequest{
		ID:    "default-test",
		Name:  "Default Test",
		Image: "alpine:latest",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Set as default
	req = httptest.NewRequest(http.MethodPost, "/api/v1/runtimes/default-test/set-default", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, data["is_default"])
}

func TestRuntimeCreateValidation(t *testing.T) {
	router, _, tempDir := setupRuntimeTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Missing required fields
	createReq := map[string]interface{}{
		"name": "Test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtimes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
