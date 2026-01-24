package auth

import "github.com/golang-jwt/jwt/v5"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse 登录成功响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      UserInfo `json:"user"`
}

// UserInfo 用户信息（返回给前端）
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Claims JWT Claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// CreateUserRequest 创建用户请求（admin）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"`
}

// UserResponse 用户信息响应（admin 列表）
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// CreateAPIKeyRequest 创建 API Key 请求
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=255"`
	ExpiresIn int    `json:"expires_in,omitempty"` // 过期天数，0 表示永不过期
}

// APIKeyResponse API Key 响应（创建时返回完整 key，之后只返回前缀）
type APIKeyResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Key        string  `json:"key,omitempty"` // 仅创建时返回完整 key
	LastUsedAt *string `json:"last_used_at"`
	ExpiresAt  *string `json:"expires_at"`
	CreatedAt  string  `json:"created_at"`
}

// APIKeyClaims API Key 验证后的 Claims（类似 JWT Claims）
type APIKeyClaims struct {
	UserID   string
	Username string
	Role     string
	KeyID    string
}
