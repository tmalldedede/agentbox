package credential

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager Credential 管理器
type Manager struct {
	dataDir     string
	credentials map[string]*Credential
	crypto      *Crypto
	mu          sync.RWMutex
}

// NewManager 创建 Manager
func NewManager(dataDir string, encryptionKey string) (*Manager, error) {
	crypto, err := NewCrypto(encryptionKey)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		dataDir:     dataDir,
		credentials: make(map[string]*Credential),
		crypto:      crypto,
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0700); err != nil { // 更严格的权限
		return nil, err
	}

	// 加载凭证
	if err := m.loadCredentials(); err != nil {
		return nil, err
	}

	return m, nil
}

// loadCredentials 从文件加载凭证
func (m *Manager) loadCredentials() error {
	filePath := filepath.Join(m.dataDir, "credentials.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var credentials []*Credential
	if err := json.Unmarshal(data, &credentials); err != nil {
		return err
	}

	for _, c := range credentials {
		m.credentials[c.ID] = c
	}

	return nil
}

// saveCredentials 保存凭证到文件
func (m *Manager) saveCredentials() error {
	credentials := make([]*Credential, 0, len(m.credentials))
	for _, c := range m.credentials {
		credentials = append(credentials, c)
	}

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(m.dataDir, "credentials.json")
	return os.WriteFile(filePath, data, 0600) // 更严格的权限
}

// List 列出所有凭证（掩码值）
func (m *Manager) List() []*Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()

	credentials := make([]*Credential, 0, len(m.credentials))
	for _, c := range m.credentials {
		// 返回副本，不包含原始值
		cred := *c
		cred.Value = "" // 不返回加密值
		credentials = append(credentials, &cred)
	}
	return credentials
}

// ListByScope 按作用域列出凭证
func (m *Manager) ListByScope(scope Scope) []*Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var credentials []*Credential
	for _, c := range m.credentials {
		if c.Scope == scope {
			cred := *c
			cred.Value = ""
			credentials = append(credentials, &cred)
		}
	}
	return credentials
}

// ListByProvider 按提供商列出凭证
func (m *Manager) ListByProvider(provider Provider) []*Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var credentials []*Credential
	for _, c := range m.credentials {
		if c.Provider == provider {
			cred := *c
			cred.Value = ""
			credentials = append(credentials, &cred)
		}
	}
	return credentials
}

// Get 获取凭证（掩码值）
func (m *Manager) Get(id string) (*Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.credentials[id]
	if !ok {
		return nil, ErrCredentialNotFound
	}

	// 返回副本，不包含原始值
	cred := *c
	cred.Value = ""
	return &cred, nil
}

// GetDecrypted 获取解密后的凭证值
func (m *Manager) GetDecrypted(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.credentials[id]
	if !ok {
		return "", ErrCredentialNotFound
	}

	// 解密
	value, err := m.crypto.Decrypt(c.Value)
	if err != nil {
		return "", err
	}

	// 更新最后使用时间
	go m.updateLastUsed(id)

	return value, nil
}

// updateLastUsed 更新最后使用时间
func (m *Manager) updateLastUsed(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.credentials[id]; ok {
		now := time.Now()
		c.LastUsedAt = &now
		m.saveCredentials()
	}
}

// Create 创建凭证
func (m *Manager) Create(req *CreateCredentialRequest) (*Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查 ID 是否已存在
	if _, ok := m.credentials[req.ID]; ok {
		return nil, ErrCredentialAlreadyExists
	}

	// 加密值
	encryptedValue, err := m.crypto.Encrypt(req.Value)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	now := time.Now()
	credential := &Credential{
		ID:          req.ID,
		Name:        req.Name,
		Type:        req.Type,
		Provider:    req.Provider,
		Value:       encryptedValue,
		ValueMasked: (&Credential{}).MaskValue(req.Value),
		Scope:       req.Scope,
		ProfileID:   req.ProfileID,
		EnvVar:      req.EnvVar,
		IsValid:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 默认值
	if credential.Type == "" {
		credential.Type = TypeAPIKey
	}
	if credential.Scope == "" {
		credential.Scope = ScopeGlobal
	}

	if err := credential.Validate(); err != nil {
		return nil, err
	}

	m.credentials[credential.ID] = credential

	if err := m.saveCredentials(); err != nil {
		delete(m.credentials, credential.ID)
		return nil, err
	}

	// 返回副本，不包含加密值
	cred := *credential
	cred.Value = ""
	return &cred, nil
}

// Update 更新凭证
func (m *Manager) Update(id string, req *UpdateCredentialRequest) (*Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	credential, ok := m.credentials[id]
	if !ok {
		return nil, ErrCredentialNotFound
	}

	if req.Name != nil {
		credential.Name = *req.Name
	}
	if req.Type != nil {
		credential.Type = *req.Type
	}
	if req.Provider != nil {
		credential.Provider = *req.Provider
	}
	if req.Value != nil {
		// 加密新值
		encryptedValue, err := m.crypto.Encrypt(*req.Value)
		if err != nil {
			return nil, ErrEncryptionFailed
		}
		credential.Value = encryptedValue
		credential.ValueMasked = credential.MaskValue(*req.Value)
	}
	if req.Scope != nil {
		credential.Scope = *req.Scope
	}
	if req.ProfileID != nil {
		credential.ProfileID = *req.ProfileID
	}
	if req.EnvVar != nil {
		credential.EnvVar = *req.EnvVar
	}

	credential.UpdatedAt = time.Now()

	if err := m.saveCredentials(); err != nil {
		return nil, err
	}

	// 返回副本，不包含加密值
	cred := *credential
	cred.Value = ""
	return &cred, nil
}

// Delete 删除凭证
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.credentials[id]; !ok {
		return ErrCredentialNotFound
	}

	delete(m.credentials, id)

	return m.saveCredentials()
}

// Verify 验证凭证有效性
func (m *Manager) Verify(id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.credentials[id]
	if !ok {
		return false, ErrCredentialNotFound
	}

	// 检查过期
	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	// TODO: 实际验证 API Key 有效性
	// 目前只返回存储的 IsValid 状态
	return c.IsValid, nil
}

// GetEnvMap 获取凭证的环境变量映射
func (m *Manager) GetEnvMap(ids []string) (map[string]string, error) {
	envMap := make(map[string]string)

	for _, id := range ids {
		value, err := m.GetDecrypted(id)
		if err != nil {
			return nil, err
		}

		c, err := m.Get(id)
		if err != nil {
			return nil, err
		}

		envMap[c.GetEnvVar()] = value
	}

	return envMap, nil
}
