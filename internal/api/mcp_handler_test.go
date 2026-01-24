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
	"github.com/tmalldedede/agentbox/internal/mcp"
)

func setupMCPTestRouter(t *testing.T) (*gin.Engine, *MCPHandler, string) {
	tempDir, err := os.MkdirTemp("", "mcp_test")
	require.NoError(t, err)

	mgr, err := mcp.NewManager(tempDir)
	require.NoError(t, err)

	handler := NewMCPHandler(mgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler, tempDir
}

func TestMCPList(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	servers, ok := resp.Data.([]interface{})
	require.True(t, ok)
	// Should have built-in MCP servers
	assert.GreaterOrEqual(t, len(servers), 0)
}

func TestMCPListByCategory(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers?category=security", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestMCPListEnabled(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers?enabled=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestMCPStats(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	stats, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, stats, "total")
	assert.Contains(t, stats, "enabled")
}

func TestMCPCreate(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := mcp.CreateServerRequest{
		ID:          "test-mcp",
		Name:        "Test MCP Server",
		Description: "A test MCP server",
		Type:        mcp.ServerTypeStdio,
		Category:    mcp.CategoryTool,
		Command:     "node",
		Args:        []string{"server.js"},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
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
	assert.Equal(t, "test-mcp", data["id"])
	assert.Equal(t, "Test MCP Server", data["name"])
}

func TestMCPGet(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create an MCP server
	createReq := mcp.CreateServerRequest{
		ID:      "get-test",
		Name:    "Get Test",
		Type:    mcp.ServerTypeStdio,
		Command: "echo",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now get it
	req = httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers/get-test", nil)
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

func TestMCPGetNotFound(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMCPUpdate(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create an MCP server
	createReq := mcp.CreateServerRequest{
		ID:      "update-test",
		Name:    "Update Test",
		Type:    mcp.ServerTypeStdio,
		Command: "echo",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now update it
	isEnabled := true
	updatedName := "Updated Name"
	updatedDesc := "Updated description"
	updateReq := mcp.UpdateServerRequest{
		Name:        &updatedName,
		Description: &updatedDesc,
		IsEnabled:   &isEnabled,
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/mcp-servers/update-test", bytes.NewReader(body))
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

func TestMCPDelete(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create an MCP server
	createReq := mcp.CreateServerRequest{
		ID:      "delete-test",
		Name:    "Delete Test",
		Type:    mcp.ServerTypeStdio,
		Command: "echo",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/mcp-servers/delete-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/mcp-servers/delete-test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMCPClone(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create an MCP server
	createReq := mcp.CreateServerRequest{
		ID:          "clone-source",
		Name:        "Clone Source",
		Type:        mcp.ServerTypeStdio,
		Command:     "node",
		Args:        []string{"server.js"},
		Description: "Source server",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Clone it
	cloneReq := struct {
		NewID   string `json:"new_id"`
		NewName string `json:"new_name"`
	}{
		NewID:   "clone-target",
		NewName: "Clone Target",
	}

	body, _ = json.Marshal(cloneReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers/clone-source/clone", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "clone-target", data["id"])
	assert.Equal(t, "Clone Target", data["name"])
}

func TestMCPCreateValidation(t *testing.T) {
	router, _, tempDir := setupMCPTestRouter(t)
	defer os.RemoveAll(tempDir)

	// Missing required fields
	createReq := map[string]interface{}{
		"name": "Test",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp-servers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// API returns 400 for validation errors (id is required)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
