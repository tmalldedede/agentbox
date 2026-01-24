package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/auth"
)

// AuthHandler 认证相关 API
type AuthHandler struct {
	authManager *auth.Manager
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(authManager *auth.Manager) *AuthHandler {
	return &AuthHandler{authManager: authManager}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.authManager.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": resp})
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": auth.UserInfo{
			ID:       userID,
			Username: username,
			Role:     role,
		},
	})
}

// ChangePassword 修改当前用户密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request: " + err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if err := h.authManager.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"message": "password changed successfully"}})
}

// CreateUser 创建用户（仅 admin）
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req auth.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request: " + err.Error()})
		return
	}

	user, err := h.authManager.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": auth.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// ListUsers 列出所有用户（仅 admin）
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authManager.ListUsers()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	result := make([]auth.UserResponse, len(users))
	for i, u := range users {
		result[i] = auth.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// DeleteUser 删除用户（仅 admin）
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 不允许删除自己
	currentUserID := c.GetString("user_id")
	if id == currentUserID {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "cannot delete yourself"})
		return
	}

	if err := h.authManager.DeleteUser(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "deleted": true}})
}

// ==================== API Key 管理 ====================

// CreateAPIKey 创建 API Key
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req auth.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "invalid request: " + err.Error()})
		return
	}

	userID := c.GetString("user_id")
	key, err := h.authManager.CreateAPIKey(userID, req.Name, req.ExpiresIn)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": key})
}

// ListAPIKeys 列出当前用户的 API Keys
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	keys, err := h.authManager.ListAPIKeys(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": keys})
}

// DeleteAPIKey 删除 API Key
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID := c.GetString("user_id")
	keyID := c.Param("id")

	if err := h.authManager.DeleteAPIKey(userID, keyID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": keyID, "deleted": true}})
}
