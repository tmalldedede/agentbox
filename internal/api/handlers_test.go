package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/agent"

	// Import agent adapters to register them
	_ "github.com/tmalldedede/agentbox/internal/agent/claude"
	_ "github.com/tmalldedede/agentbox/internal/agent/codex"
	_ "github.com/tmalldedede/agentbox/internal/agent/opencode"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestRouter creates a test router with handler
func setupTestRouter() (*gin.Engine, *Handler) {
	registry := agent.DefaultRegistry()
	handler := NewHandler(nil, registry) // sessionMgr can be nil for some tests

	router := gin.New()
	return router, handler
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/health", handler.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ok", data["status"])
	assert.Equal(t, "0.1.0", data["version"])
}

// TestListAgents tests the list agents endpoint
func TestListAgents(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/agents", handler.ListAgents)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	// Should have at least claude-code, codex, opencode
	agents, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(agents), 3)
}

// TestResponseFormat tests that all responses follow the standard format
func TestResponseFormat(t *testing.T) {
	tests := []struct {
		name           string
		setupRoute     func(*gin.Engine, *Handler)
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedCode   int
	}{
		{
			name: "health check returns standard format",
			setupRoute: func(r *gin.Engine, h *Handler) {
				r.GET("/health", h.HealthCheck)
			},
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedCode:   0,
		},
		{
			name: "agents list returns standard format",
			setupRoute: func(r *gin.Engine, h *Handler) {
				r.GET("/agents", h.ListAgents)
			},
			method:         http.MethodGet,
			path:           "/agents",
			expectedStatus: http.StatusOK,
			expectedCode:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, handler := setupTestRouter()
			tt.setupRoute(router, handler)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err, "Response should be valid JSON")

			assert.Equal(t, tt.expectedCode, resp.Code)
			assert.NotEmpty(t, resp.Message)
		})
	}
}
