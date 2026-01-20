package credential

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// CredentialType 凭证类型
type CredentialType string

const (
	TypeAPIKey CredentialType = "api_key"
	TypeToken  CredentialType = "token"
	TypeOAuth  CredentialType = "oauth"
)

// Provider 凭证提供商
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
	ProviderGitHub    Provider = "github"
	ProviderCustom    Provider = "custom"
)

// Scope 凭证作用域
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProfile Scope = "profile"
	ScopeSession Scope = "session"
)

// Credential 凭证
type Credential struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     CredentialType `json:"type"`
	Provider Provider       `json:"provider"`

	// 凭证值（加密存储，API 返回时掩码）
	Value       string `json:"value,omitempty"`       // 加密后的值
	ValueMasked string `json:"value_masked,omitempty"` // 掩码显示

	// 作用域
	Scope     Scope  `json:"scope"`
	ProfileID string `json:"profile_id,omitempty"`

	// 环境变量名
	EnvVar string `json:"env_var,omitempty"`

	// 状态
	IsValid    bool       `json:"is_valid"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate 验证凭证配置
func (c *Credential) Validate() error {
	if c.ID == "" {
		return errors.New("id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Value == "" {
		return errors.New("value is required")
	}

	// 验证 ID 格式
	if strings.ContainsAny(c.ID, " \t\n/\\") {
		return errors.New("id cannot contain whitespace or slashes")
	}

	return nil
}

// MaskValue 生成掩码值
func (c *Credential) MaskValue(decryptedValue string) string {
	if len(decryptedValue) <= 8 {
		return "****"
	}
	return decryptedValue[:4] + "..." + decryptedValue[len(decryptedValue)-4:]
}

// GetEnvVar 获取环境变量名
func (c *Credential) GetEnvVar() string {
	if c.EnvVar != "" {
		return c.EnvVar
	}
	// 默认环境变量名
	switch c.Provider {
	case ProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	case ProviderOpenAI:
		return "OPENAI_API_KEY"
	case ProviderGitHub:
		return "GITHUB_TOKEN"
	default:
		return strings.ToUpper(string(c.Provider)) + "_API_KEY"
	}
}

// CreateCredentialRequest 创建凭证请求
type CreateCredentialRequest struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      CredentialType `json:"type"`
	Provider  Provider       `json:"provider"`
	Value     string         `json:"value"` // 原始明文值
	Scope     Scope          `json:"scope"`
	ProfileID string         `json:"profile_id,omitempty"`
	EnvVar    string         `json:"env_var,omitempty"`
}

// UpdateCredentialRequest 更新凭证请求
type UpdateCredentialRequest struct {
	Name      *string         `json:"name,omitempty"`
	Type      *CredentialType `json:"type,omitempty"`
	Provider  *Provider       `json:"provider,omitempty"`
	Value     *string         `json:"value,omitempty"` // 原始明文值
	Scope     *Scope          `json:"scope,omitempty"`
	ProfileID *string         `json:"profile_id,omitempty"`
	EnvVar    *string         `json:"env_var,omitempty"`
}

// 错误定义 - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrCredentialNotFound      = apperr.NotFound("credential")
	ErrCredentialAlreadyExists = apperr.AlreadyExists("credential")
	ErrEncryptionFailed        = apperr.Internal("encryption failed")
	ErrDecryptionFailed        = apperr.Internal("decryption failed")
)

// Crypto 加密工具
type Crypto struct {
	key []byte
}

// NewCrypto 创建加密工具
func NewCrypto(key string) (*Crypto, error) {
	// 将 key 调整为 32 字节 (AES-256)
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// 补齐
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	return &Crypto{key: keyBytes}, nil
}

// Encrypt 加密
func (c *Crypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (c *Crypto) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", ErrDecryptionFailed
	}

	nonce, cipherData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
