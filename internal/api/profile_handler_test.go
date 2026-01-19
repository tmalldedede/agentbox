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
	"github.com/tmalldedede/agentbox/internal/profile"
)

func setupProfileTestRouter(t *testing.T) (*gin.Engine, *ProfileHandler, string) {
	// Create temp directory for test data
	tempDir, err := os.MkdirTemp("", "profile_test")
	require.NoError(t, err)

	mgr, err := profile.NewManager(tempDir)
	require.NoError(t, err)

	handler := NewProfileHandler(mgr)

	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, handler, tempDir
}

func TestProfileList(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	// Should have built-in profiles
	profiles, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(profiles), 1, "Should have at least 1 built-in profile")
}

func TestProfileCreate(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := CreateProfileRequest{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "A test profile",
		Adapter:     "claude-code",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
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
	assert.Equal(t, "test-profile", p["id"])
	assert.Equal(t, "Test Profile", p["name"])
}

func TestProfileGet(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a profile
	createReq := CreateProfileRequest{
		ID:      "get-test",
		Name:    "Get Test Profile",
		Adapter: "claude-code",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now get it
	req = httptest.NewRequest(http.MethodGet, "/api/v1/profiles/get-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "get-test", p["id"])
}

func TestProfileGetNotFound(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.Code)
}

func TestProfileUpdate(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a profile
	createReq := CreateProfileRequest{
		ID:      "update-test",
		Name:    "Update Test",
		Adapter: "claude-code",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now update it
	updateReq := CreateProfileRequest{
		ID:          "update-test",
		Name:        "Updated Name",
		Description: "Updated description",
		Adapter:     "claude-code",
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/profiles/update-test", bytes.NewReader(body))
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

func TestProfileDelete(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a profile
	createReq := CreateProfileRequest{
		ID:      "delete-test",
		Name:    "Delete Test",
		Adapter: "claude-code",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/profiles/delete-test", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/profiles/delete-test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProfileClone(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	// First create a profile
	createReq := CreateProfileRequest{
		ID:          "clone-source",
		Name:        "Clone Source",
		Description: "Source profile for cloning",
		Adapter:     "claude-code",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now clone it
	cloneReq := CloneRequest{
		NewID:   "cloned-profile",
		NewName: "Cloned Profile",
	}

	body, _ = json.Marshal(cloneReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/profiles/clone-source/clone", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "cloned-profile", p["id"])
	assert.Equal(t, "Cloned Profile", p["name"])
	// Description should be copied from source
	assert.Equal(t, "Source profile for cloning", p["description"])
}

func TestProfileWithModelConfig(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	createReq := CreateProfileRequest{
		ID:      "model-test",
		Name:    "Model Test Profile",
		Adapter: "claude-code",
		Model: profile.ModelConfig{
			Name:            "claude-3-5-sonnet-20241022",
			Provider:        "anthropic",
			BaseURL:         "https://api.anthropic.com",
			HaikuModel:      "claude-3-5-haiku-20241022",
			SonnetModel:     "claude-3-5-sonnet-20241022",
			OpusModel:       "claude-3-opus-20240229",
			TimeoutMS:       60000,
			MaxOutputTokens: 8192,
		},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	p, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	model, ok := p["model"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "claude-3-5-sonnet-20241022", model["name"])
	assert.Equal(t, "anthropic", model["provider"])
	assert.Equal(t, "claude-3-5-haiku-20241022", model["haiku_model"])
}

func TestProfileValidation(t *testing.T) {
	router, _, tempDir := setupProfileTestRouter(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name       string
		request    interface{}
		statusCode int
	}{
		{
			name: "missing id",
			request: map[string]string{
				"name":    "Test",
				"adapter": "claude-code",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "missing name",
			request: map[string]string{
				"id":      "test",
				"adapter": "claude-code",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "missing adapter",
			request: map[string]string{
				"id":   "test",
				"name": "Test",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
