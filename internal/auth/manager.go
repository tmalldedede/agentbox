package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/database"
	"github.com/tmalldedede/agentbox/internal/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var log *slog.Logger

func init() {
	log = logger.Module("auth")
}

// Manager 认证管理器
type Manager struct {
	db        *gorm.DB
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewManager 创建认证管理器
func NewManager(db *gorm.DB) *Manager {
	secret := os.Getenv("AGENTBOX_JWT_SECRET")
	if secret == "" {
		secret = loadOrCreateSecret()
	}

	m := &Manager{
		db:        db,
		jwtSecret: []byte(secret),
		tokenTTL:  24 * time.Hour,
	}

	// 自动创建默认 admin 用户
	m.ensureDefaultAdmin()

	return m
}

// loadOrCreateSecret 从文件加载或生成并持久化 JWT 密钥
func loadOrCreateSecret() string {
	secretFile := jwtSecretPath()

	// 尝试从文件读取
	data, err := os.ReadFile(secretFile)
	if err == nil {
		s := strings.TrimSpace(string(data))
		if s != "" {
			log.Info("JWT secret loaded from file", "path", secretFile)
			return s
		}
	}

	// 生成新密钥
	b := make([]byte, 32)
	rand.Read(b)
	secret := hex.EncodeToString(b)

	// 持久化到文件
	dir := filepath.Dir(secretFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Warn("Failed to create directory for JWT secret", "error", err)
		return secret
	}
	if err := os.WriteFile(secretFile, []byte(secret), 0600); err != nil {
		log.Warn("Failed to persist JWT secret to file", "error", err)
	} else {
		log.Info("JWT secret generated and persisted", "path", secretFile)
	}

	return secret
}

// jwtSecretPath 返回 JWT 密钥文件路径
func jwtSecretPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".agentbox", ".jwt_secret")
}

// Login 用户登录
func (m *Manager) Login(username, password string) (*TokenResponse, error) {
	var user database.UserModel
	if err := m.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 生成 JWT
	now := time.Now()
	expiresAt := now.Add(m.tokenTTL)

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &TokenResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	}, nil
}

// ValidateToken 验证 JWT token
func (m *Manager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// CreateUser 创建用户
func (m *Manager) CreateUser(username, password, role string) (*database.UserModel, error) {
	if role == "" {
		role = "user"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &database.UserModel{
		BaseModel: database.BaseModel{
			ID: uuid.New().String(),
		},
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
	}

	if err := m.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Info("user created", "username", username, "role", role)
	return user, nil
}

// ListUsers 列出所有用户
func (m *Manager) ListUsers() ([]database.UserModel, error) {
	var users []database.UserModel
	if err := m.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUser 删除用户
func (m *Manager) DeleteUser(id string) error {
	result := m.db.Delete(&database.UserModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// GetUserByID 根据 ID 获取用户
func (m *Manager) GetUserByID(id string) (*database.UserModel, error) {
	var user database.UserModel
	if err := m.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ChangePassword 修改密码
func (m *Manager) ChangePassword(userID, oldPassword, newPassword string) error {
	var user database.UserModel
	if err := m.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := m.db.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	log.Info("password changed", "user_id", userID)
	return nil
}

// ensureDefaultAdmin 确保默认管理员用户存在
func (m *Manager) ensureDefaultAdmin() {
	var count int64
	m.db.Model(&database.UserModel{}).Count(&count)
	if count > 0 {
		return
	}

	_, err := m.CreateUser("admin", "admin123", "admin")
	if err != nil {
		log.Error("failed to create default admin user", "error", err)
		return
	}

	log.Warn("=== Default admin user created: admin / admin123 ===")
	log.Warn("=== Please change the default password immediately! ===")
}

// ==================== API Key 管理 ====================

// CreateAPIKey 创建 API Key
func (m *Manager) CreateAPIKey(userID, name string, expiresInDays int) (*APIKeyResponse, error) {
	// 生成随机 key: ab_<32字节hex>
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	fullKey := "ab_" + hex.EncodeToString(keyBytes)
	keyPrefix := fullKey[:10] + "..."

	// 使用 bcrypt hash 存储
	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	now := time.Now()
	var expiresAt *time.Time
	if expiresInDays > 0 {
		exp := now.AddDate(0, 0, expiresInDays)
		expiresAt = &exp
	}

	apiKey := &database.APIKeyModel{
		BaseModel: database.BaseModel{
			ID: "key-" + uuid.New().String()[:8],
		},
		UserID:    userID,
		Name:      name,
		KeyPrefix: keyPrefix,
		KeyHash:   string(keyHash),
		ExpiresAt: expiresAt,
	}

	if err := m.db.Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	log.Info("API key created", "key_id", apiKey.ID, "user_id", userID, "name", name)

	resp := &APIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		KeyPrefix: apiKey.KeyPrefix,
		Key:       fullKey, // 仅创建时返回
		CreatedAt: apiKey.CreatedAt.Format(time.RFC3339),
	}
	if expiresAt != nil {
		exp := expiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &exp
	}

	return resp, nil
}

// ListAPIKeys 列出用户的 API Keys
func (m *Manager) ListAPIKeys(userID string) ([]APIKeyResponse, error) {
	var keys []database.APIKeyModel
	if err := m.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}

	result := make([]APIKeyResponse, len(keys))
	for i, k := range keys {
		result[i] = APIKeyResponse{
			ID:        k.ID,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			CreatedAt: k.CreatedAt.Format(time.RFC3339),
		}
		if k.LastUsedAt != nil {
			t := k.LastUsedAt.Format(time.RFC3339)
			result[i].LastUsedAt = &t
		}
		if k.ExpiresAt != nil {
			t := k.ExpiresAt.Format(time.RFC3339)
			result[i].ExpiresAt = &t
		}
	}
	return result, nil
}

// DeleteAPIKey 删除 API Key
func (m *Manager) DeleteAPIKey(userID, keyID string) error {
	result := m.db.Where("id = ? AND user_id = ?", keyID, userID).Delete(&database.APIKeyModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	log.Info("API key deleted", "key_id", keyID, "user_id", userID)
	return nil
}

// ValidateAPIKey 验证 API Key，返回 Claims
func (m *Manager) ValidateAPIKey(key string) (*APIKeyClaims, error) {
	if !strings.HasPrefix(key, "ab_") {
		return nil, fmt.Errorf("invalid API key format")
	}

	// 查找所有未过期的 keys 并逐个验证 hash
	var keys []database.APIKeyModel
	if err := m.db.Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	for _, k := range keys {
		if err := bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(key)); err == nil {
			// 找到匹配的 key，获取用户信息
			var user database.UserModel
			if err := m.db.Where("id = ? AND is_active = ?", k.UserID, true).First(&user).Error; err != nil {
				return nil, fmt.Errorf("user not found or inactive")
			}

			// 更新最后使用时间（异步）
			go func(keyID string) {
				now := time.Now()
				m.db.Model(&database.APIKeyModel{}).Where("id = ?", keyID).Update("last_used_at", &now)
			}(k.ID)

			return &APIKeyClaims{
				UserID:   user.ID,
				Username: user.Username,
				Role:     user.Role,
				KeyID:    k.ID,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}
