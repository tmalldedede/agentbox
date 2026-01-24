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
	"github.com/tmalldedede/agentbox/internal/auth"
	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *auth.Manager, func()) {
	// Create temp DB
	tempFile, err := os.CreateTemp("", "auth_test_*.db")
	require.NoError(t, err)
	tempFile.Close()

	db, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
	require.NoError(t, err)

	// Migrate user and api_key tables
	err = db.AutoMigrate(&database.UserModel{}, &database.APIKeyModel{})
	require.NoError(t, err)

	mgr := auth.NewManager(db)
	handler := NewAuthHandler(mgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Public routes
	router.POST("/api/v1/auth/login", handler.Login)

	// Authenticated routes (simulate middleware by setting user context)
	authenticated := router.Group("/api/v1")
	authenticated.Use(func(c *gin.Context) {
		// Simulate auth middleware - get user from header for testing
		userID := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		role := c.GetHeader("X-Role")
		if userID != "" {
			c.Set("user_id", userID)
			c.Set("username", username)
			c.Set("role", role)
		}
		c.Next()
	})
	authenticated.GET("/auth/me", handler.Me)
	authenticated.PUT("/auth/password", handler.ChangePassword)
	authenticated.POST("/auth/api-keys", handler.CreateAPIKey)
	authenticated.GET("/auth/api-keys", handler.ListAPIKeys)
	authenticated.DELETE("/auth/api-keys/:id", handler.DeleteAPIKey)

	// Admin routes
	admin := router.Group("/api/v1/admin")
	admin.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		username := c.GetHeader("X-Username")
		role := c.GetHeader("X-Role")
		if userID != "" {
			c.Set("user_id", userID)
			c.Set("username", username)
			c.Set("role", role)
		}
		c.Next()
	})
	admin.POST("/users", handler.CreateUser)
	admin.GET("/users", handler.ListUsers)
	admin.DELETE("/users/:id", handler.DeleteUser)

	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return router, mgr, cleanup
}

func TestAuthLogin(t *testing.T) {
	router, _, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Test login with default admin credentials (admin / admin123)
	loginReq := auth.LoginRequest{
		Username: "admin",
		Password: "admin123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
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
	assert.NotEmpty(t, data["token"])
	assert.NotEmpty(t, data["user"])
}

func TestAuthLoginInvalidCredentials(t *testing.T) {
	router, _, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	loginReq := auth.LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEqual(t, 0, resp.Code)
}

func TestAuthMe(t *testing.T) {
	router, _, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Username", "testuser")
	req.Header.Set("X-Role", "user")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "user-123", data["id"])
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, "user", data["role"])
}

func TestAuthChangePassword(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// First create a user
	user, err := mgr.CreateUser("testuser", "oldpass", "user")
	require.NoError(t, err)

	// Change password
	changeReq := auth.ChangePasswordRequest{
		OldPassword: "oldpass",
		NewPassword: "newpass",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user.ID)
	req.Header.Set("X-Username", user.Username)
	req.Header.Set("X-Role", user.Role)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// Verify new password works
	_, err = mgr.Login("testuser", "newpass")
	assert.NoError(t, err)
}

func TestAuthChangePasswordWrongOld(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	user, err := mgr.CreateUser("testuser2", "oldpass", "user")
	require.NoError(t, err)

	changeReq := auth.ChangePasswordRequest{
		OldPassword: "wrongold",
		NewPassword: "newpass",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user.ID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code)
}

func TestAuthCreateUser(t *testing.T) {
	router, _, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	createReq := auth.CreateUserRequest{
		Username: "newuser",
		Password: "password123",
		Role:     "user",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "admin-id")
	req.Header.Set("X-Role", "admin")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "newuser", data["username"])
	assert.Equal(t, "user", data["role"])
}

func TestAuthListUsers(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Create some users
	_, _ = mgr.CreateUser("user1", "pass1", "user")
	_, _ = mgr.CreateUser("user2", "pass2", "user")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("X-User-ID", "admin-id")
	req.Header.Set("X-Role", "admin")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Code)

	users, ok := resp.Data.([]interface{})
	require.True(t, ok)
	// admin + 2 created users
	assert.GreaterOrEqual(t, len(users), 3)
}

func TestAuthDeleteUser(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Create a user to delete
	user, err := mgr.CreateUser("deleteuser", "pass", "user")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+user.ID, nil)
	req.Header.Set("X-User-ID", "admin-id")
	req.Header.Set("X-Role", "admin")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestAuthDeleteSelf(t *testing.T) {
	router, _, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Try to delete yourself - should fail
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/admin-id", nil)
	req.Header.Set("X-User-ID", "admin-id")
	req.Header.Set("X-Role", "admin")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code, "Should not be able to delete yourself")
}

func TestAuthAPIKeyCreate(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	// Create a user first
	user, err := mgr.CreateUser("apiuser", "pass", "user")
	require.NoError(t, err)

	createReq := auth.CreateAPIKeyRequest{
		Name:      "Test API Key",
		ExpiresIn: 30, // 30 days
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user.ID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data["key"], "Should return the API key")
}

func TestAuthAPIKeyList(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	user, err := mgr.CreateUser("apiuser2", "pass", "user")
	require.NoError(t, err)

	// Create some API keys
	_, _ = mgr.CreateAPIKey(user.ID, "Key1", 30)
	_, _ = mgr.CreateAPIKey(user.ID, "Key2", 30)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/api-keys", nil)
	req.Header.Set("X-User-ID", user.ID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	keys, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(keys))
}

func TestAuthAPIKeyDelete(t *testing.T) {
	router, mgr, cleanup := setupAuthTestRouter(t)
	defer cleanup()

	user, err := mgr.CreateUser("apiuser3", "pass", "user")
	require.NoError(t, err)

	key, err := mgr.CreateAPIKey(user.ID, "ToDelete", 30)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/api-keys/"+key.ID, nil)
	req.Header.Set("X-User-ID", user.ID)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// Verify it's gone
	keys, _ := mgr.ListAPIKeys(user.ID)
	assert.Equal(t, 0, len(keys))
}
