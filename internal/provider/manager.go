package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// Manager manages provider configurations and API keys
type Manager struct {
	mu        sync.RWMutex
	providers map[string]*Provider
	keys      map[string]*ProviderKeyData // providerID -> key data
	rotators  map[string]*ProfileRotator  // providerID -> rotator
	crypto    *Crypto
	dataDir   string
}

// NewManager creates a new provider manager
func NewManager(dataDir string, encryptionKey string) *Manager {
	m := &Manager{
		providers: make(map[string]*Provider),
		keys:      make(map[string]*ProviderKeyData),
		rotators:  make(map[string]*ProfileRotator),
		crypto:    NewCrypto(encryptionKey),
		dataDir:   dataDir,
	}

	// Ensure data directory exists
	os.MkdirAll(dataDir, 0700)

	// Load built-in providers
	for _, p := range BuiltinProviders {
		cp := *p // copy
		m.providers[cp.ID] = &cp
	}

	// Load custom providers from disk
	m.loadCustomProviders()

	// Load API keys and populate provider status
	m.loadKeys()

	return m
}

// List returns all providers
func (m *Manager) List() []*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Provider, 0, len(m.providers))
	for _, p := range m.providers {
		result = append(result, p)
	}
	return result
}

// ListConfigured returns only providers that have API keys configured (active instances)
func (m *Manager) ListConfigured() []*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Provider
	for _, p := range m.providers {
		if p.IsConfigured || !p.IsBuiltIn {
			result = append(result, p)
		}
	}
	return result
}

// ListTemplates returns all built-in providers as templates for the "Add Provider" flow
func (m *Manager) ListTemplates() []*Provider {
	return BuiltinProviders
}

// Stats returns provider statistics
type ProviderStats struct {
	Total         int `json:"total"`          // All active instances (configured built-in + custom)
	Configured    int `json:"configured"`     // Instances with API key set
	Valid         int `json:"valid"`          // Keys verified valid
	Failed        int `json:"failed"`         // Keys configured but invalid
	NotConfigured int `json:"not_configured"` // Instances without key
}

func (m *Manager) Stats() *ProviderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &ProviderStats{}
	for _, p := range m.providers {
		// Only count: custom providers or built-in providers that have keys configured
		if !p.IsBuiltIn || p.IsConfigured {
			stats.Total++
			if p.IsConfigured {
				stats.Configured++
				if p.IsValid {
					stats.Valid++
				} else {
					stats.Failed++
				}
			} else {
				stats.NotConfigured++
			}
		}
	}
	return stats
}

// ListByAgent returns providers for a specific agent
func (m *Manager) ListByAgent(agent string) []*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Provider
	for _, p := range m.providers {
		if p.SupportsAgent(agent) {
			result = append(result, p)
		}
	}
	return result
}

// ListByCategory returns providers by category
func (m *Manager) ListByCategory(category ProviderCategory) []*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Provider
	for _, p := range m.providers {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// Get returns a provider by ID
func (m *Manager) Get(id string) (*Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[id]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return p, nil
}

// CreateFromTemplate creates a new provider instance from a built-in template
func (m *Manager) CreateFromTemplate(templateID string, id string, name string, baseURL string, apiKey string, models []string) (*Provider, error) {
	// Find template
	var tmpl *Provider
	for _, t := range BuiltinProviders {
		if t.ID == templateID {
			tmpl = t
			break
		}
	}
	if tmpl == nil {
		return nil, ErrProviderNotFound
	}

	// Create instance from template
	p := &Provider{
		ID:          id,
		Name:        name,
		Description: tmpl.Description,
		TemplateID:  templateID,
		Agents:      tmpl.Agents,
		Category:    tmpl.Category,
		WebsiteURL:  tmpl.WebsiteURL,
		APIKeyURL:   tmpl.APIKeyURL,
		DocsURL:     tmpl.DocsURL,
		BaseURL:     baseURL,
		EnvConfig:   tmpl.EnvConfig,
		Icon:        tmpl.Icon,
		IconColor:   tmpl.IconColor,
		IsBuiltIn:   false,
		RequiresAK:  tmpl.RequiresAK,
		IsEnabled:   true,
	}

	if baseURL == "" {
		p.BaseURL = tmpl.BaseURL
	}

	if len(models) > 0 {
		p.DefaultModels = models
		p.DefaultModel = models[0]
	} else {
		p.DefaultModels = tmpl.DefaultModels
		p.DefaultModel = tmpl.DefaultModel
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	if _, exists := m.providers[id]; exists {
		m.mu.Unlock()
		return nil, errors.New("provider ID already exists")
	}
	m.providers[p.ID] = p
	if err := m.saveCustomProviders(); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	m.mu.Unlock()

	// If API key provided, configure it immediately (separate lock)
	if apiKey != "" {
		if err := m.ConfigureKey(id, apiKey); err == nil {
			// Re-read the provider to get updated key status
			m.mu.RLock()
			p = m.providers[id]
			m.mu.RUnlock()
		}
	}

	return p, nil
}

// Create creates a new custom provider
func (m *Manager) Create(p *Provider) error {
	if err := p.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[p.ID]; exists {
		return errors.New("provider ID already exists")
	}

	p.IsBuiltIn = false
	m.providers[p.ID] = p

	return m.saveCustomProviders()
}

// Update updates an existing provider
func (m *Manager) Update(id string, updates *Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.providers[id]
	if !ok {
		return ErrProviderNotFound
	}

	// Cannot update built-in providers
	if existing.IsBuiltIn {
		updated := *existing
		if updates.Name != "" {
			updated.Name = updates.Name
		}
		if updates.Description != "" {
			updated.Description = updates.Description
		}
		if updates.BaseURL != "" {
			updated.BaseURL = updates.BaseURL
		}
		if updates.EnvConfig != nil {
			updated.EnvConfig = updates.EnvConfig
		}
		updated.IsBuiltIn = false
		updated.ID = id + "-custom"
		m.providers[updated.ID] = &updated
		return m.saveCustomProviders()
	}

	// Update custom provider
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.BaseURL != "" {
		existing.BaseURL = updates.BaseURL
	}
	if updates.EnvConfig != nil {
		existing.EnvConfig = updates.EnvConfig
	}
	if updates.DefaultModel != "" {
		existing.DefaultModel = updates.DefaultModel
	}
	if updates.DefaultModels != nil {
		existing.DefaultModels = updates.DefaultModels
	}

	return m.saveCustomProviders()
}

// Delete deletes a custom provider
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.providers[id]
	if !ok {
		return ErrProviderNotFound
	}

	if p.IsBuiltIn {
		return nil
	}

	delete(m.providers, id)
	delete(m.keys, id)
	m.saveKeys()
	return m.saveCustomProviders()
}

// --- API Key Management ---

// ConfigureKey sets or updates the API key for a provider
func (m *Manager) ConfigureKey(id string, apiKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.providers[id]
	if !ok {
		return ErrProviderNotFound
	}

	encrypted, err := m.crypto.Encrypt(apiKey)
	if err != nil {
		return ErrEncryptionFailed
	}

	now := time.Now()
	keyData := &ProviderKeyData{
		ProviderID:   id,
		EncryptedKey: encrypted,
		KeyMasked:    MaskAPIKey(apiKey),
		IsValid:      true,
		UpdatedAt:    now,
	}
	m.keys[id] = keyData

	// Update provider status
	p.APIKeyMasked = keyData.KeyMasked
	p.IsConfigured = true
	p.IsValid = true
	p.LastValidatedAt = &now

	return m.saveKeys()
}

// DeleteKey removes the API key for a provider
func (m *Manager) DeleteKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.providers[id]
	if !ok {
		return ErrProviderNotFound
	}

	delete(m.keys, id)

	p.APIKeyMasked = ""
	p.IsConfigured = false
	p.IsValid = false
	p.LastValidatedAt = nil

	return m.saveKeys()
}

// ValidateKey validates the API key for a provider by making a real API call
func (m *Manager) ValidateKey(id string) (bool, error) {
	m.mu.RLock()
	p, ok := m.providers[id]
	if !ok {
		m.mu.RUnlock()
		return false, ErrProviderNotFound
	}

	keyData, ok := m.keys[id]
	if !ok {
		m.mu.RUnlock()
		return false, ErrKeyNotConfigured
	}

	apiKey, err := m.crypto.Decrypt(keyData.EncryptedKey)
	if err != nil {
		m.mu.RUnlock()
		m.markKeyInvalid(id)
		return false, nil
	}

	baseURL := p.BaseURL
	agents := p.Agents
	m.mu.RUnlock()

	// Make actual API call to validate the key
	valid := m.probeAPI(baseURL, apiKey, agents)

	// Update status
	m.mu.Lock()
	now := time.Now()
	if kd, ok := m.keys[id]; ok {
		kd.IsValid = valid
		kd.LastValidatedAt = &now
	}
	if prov, ok := m.providers[id]; ok {
		prov.IsValid = valid
		prov.LastValidatedAt = &now
	}
	m.saveKeys()
	m.mu.Unlock()

	return valid, nil
}

// probeAPI makes a lightweight API call to verify the key works
func (m *Manager) probeAPI(baseURL string, apiKey string, agents []string) bool {
	if baseURL == "" {
		// Official providers without base_url — try to determine protocol
		for _, a := range agents {
			if a == AgentClaudeCode {
				return m.probeAnthropic("https://api.anthropic.com", apiKey)
			}
			if a == AgentCodex {
				return m.probeOpenAI("https://api.openai.com/v1", apiKey)
			}
		}
		return true // Can't determine, assume valid
	}

	// Try to determine protocol from agents
	for _, a := range agents {
		if a == AgentClaudeCode {
			// Anthropic-compatible: strip trailing path if it ends with /anthropic
			url := baseURL
			if !strings.HasSuffix(url, "/anthropic") {
				url = strings.TrimRight(url, "/")
			}
			return m.probeAnthropic(url, apiKey)
		}
		if a == AgentCodex {
			return m.probeOpenAI(baseURL, apiKey)
		}
	}

	return true
}

// probeAnthropic calls Anthropic /v1/messages with a minimal request
func (m *Manager) probeAnthropic(baseURL string, apiKey string) bool {
	url := strings.TrimRight(baseURL, "/") + "/v1/messages"
	body := `{"model":"claude-haiku-3-5-20241022","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 200 = success, 401/403 = invalid key, others (400, 429, etc.) = key is valid but request issue
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return false
	}
	return true
}

// probeOpenAI calls OpenAI /models to verify the key
func (m *Manager) probeOpenAI(baseURL string, apiKey string) bool {
	url := strings.TrimRight(baseURL, "/") + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return false
	}
	return true
}

// FetchModels fetches the available model list for a configured provider
func (m *Manager) FetchModels(id string) ([]string, error) {
	m.mu.RLock()
	p, ok := m.providers[id]
	if !ok {
		m.mu.RUnlock()
		return nil, ErrProviderNotFound
	}

	keyData, ok := m.keys[id]
	if !ok {
		m.mu.RUnlock()
		return nil, ErrKeyNotConfigured
	}

	apiKey, err := m.crypto.Decrypt(keyData.EncryptedKey)
	if err != nil {
		m.mu.RUnlock()
		return nil, ErrEncryptionFailed
	}

	baseURL := p.BaseURL
	agents := p.Agents
	m.mu.RUnlock()

	return m.fetchModelList(baseURL, apiKey, agents)
}

// ProbeModels fetches model list from a given base_url + api_key without requiring a saved provider
func (m *Manager) ProbeModels(baseURL string, apiKey string, agents []string) ([]string, error) {
	return m.fetchModelList(baseURL, apiKey, agents)
}

// fetchModelList fetches models from the appropriate API
func (m *Manager) fetchModelList(baseURL string, apiKey string, agents []string) ([]string, error) {
	if baseURL == "" {
		for _, a := range agents {
			if a == AgentClaudeCode {
				return m.fetchAnthropicModels("https://api.anthropic.com", apiKey)
			}
			if a == AgentCodex {
				return m.fetchOpenAIModels("https://api.openai.com/v1", apiKey)
			}
		}
		return nil, errors.New("cannot determine API protocol")
	}

	for _, a := range agents {
		if a == AgentClaudeCode {
			url := baseURL
			if !strings.HasSuffix(url, "/anthropic") {
				url = strings.TrimRight(url, "/")
			}
			return m.fetchAnthropicModels(url, apiKey)
		}
		if a == AgentCodex {
			return m.fetchOpenAIModels(baseURL, apiKey)
		}
	}

	// Default: try OpenAI-compatible
	return m.fetchOpenAIModels(baseURL, apiKey)
}

// fetchOpenAIModels fetches models from OpenAI-compatible /models endpoint
func (m *Manager) fetchOpenAIModels(baseURL string, apiKey string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, errors.New("authentication failed: invalid API key")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	sort.Strings(models)
	return models, nil
}

// fetchAnthropicModels fetches models from Anthropic /v1/models endpoint
func (m *Manager) fetchAnthropicModels(baseURL string, apiKey string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/v1/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, errors.New("authentication failed: invalid API key")
	}
	if resp.StatusCode != 200 {
		// Anthropic /v1/models may not be available on all compatible endpoints
		// Fall back to empty list
		return nil, fmt.Errorf("models endpoint returned status %d (may not be supported)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (m *Manager) markKeyInvalid(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if kd, ok := m.keys[id]; ok {
		kd.IsValid = false
	}
	if p, ok := m.providers[id]; ok {
		p.IsValid = false
	}
	m.saveKeys()
}

// GetDecryptedKey returns the decrypted API key for a provider
func (m *Manager) GetDecryptedKey(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyData, ok := m.keys[id]
	if !ok {
		return "", ErrKeyNotConfigured
	}

	return m.crypto.Decrypt(keyData.EncryptedKey)
}

// GetEnvVarsWithKey returns environment variables for a provider with the actual API key injected
func (m *Manager) GetEnvVarsWithKey(id string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[id]
	if !ok {
		return nil, ErrProviderNotFound
	}

	apiKey := ""
	if keyData, ok := m.keys[id]; ok {
		decrypted, err := m.crypto.Decrypt(keyData.EncryptedKey)
		if err == nil {
			apiKey = decrypted
		}
	}

	return p.GetEnvVars(apiKey), nil
}

// --- Storage ---

func (m *Manager) keysFile() string {
	return filepath.Join(m.dataDir, "provider_keys.json")
}

func (m *Manager) customProvidersFile() string {
	return filepath.Join(m.dataDir, "providers.json")
}

func (m *Manager) loadKeys() {
	data, err := os.ReadFile(m.keysFile())
	if err != nil {
		return
	}

	var keys []*ProviderKeyData
	if err := json.Unmarshal(data, &keys); err != nil {
		return
	}

	for _, k := range keys {
		m.keys[k.ProviderID] = k
		// Populate provider status
		if p, ok := m.providers[k.ProviderID]; ok {
			p.APIKeyMasked = k.KeyMasked
			p.IsConfigured = true
			p.IsValid = k.IsValid
			p.LastValidatedAt = k.LastValidatedAt
		}
	}
}

func (m *Manager) saveKeys() error {
	keys := make([]*ProviderKeyData, 0, len(m.keys))
	for _, k := range m.keys {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		os.Remove(m.keysFile())
		return nil
	}

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.keysFile(), data, 0600)
}

func (m *Manager) loadCustomProviders() {
	data, err := os.ReadFile(m.customProvidersFile())
	if err != nil {
		return
	}

	var providers []*Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		return
	}

	for _, p := range providers {
		p.IsBuiltIn = false
		m.providers[p.ID] = p
	}
}

func (m *Manager) saveCustomProviders() error {
	var custom []*Provider
	for _, p := range m.providers {
		if !p.IsBuiltIn {
			custom = append(custom, p)
		}
	}

	if len(custom) == 0 {
		os.Remove(m.customProvidersFile())
		return nil
	}

	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.customProvidersFile(), data, 0644)
}

// --- Auth Profile Rotation ---

// GetDecryptedKeyWithRotation returns the decrypted API key using rotation
// Returns: apiKey, profileID, error
// If no profiles configured, falls back to single key mode
func (m *Manager) GetDecryptedKeyWithRotation(providerID string) (string, string, error) {
	m.mu.RLock()
	rotator, hasRotator := m.rotators[providerID]
	m.mu.RUnlock()

	// If rotator exists and has profiles, use it
	if hasRotator && rotator != nil {
		profile := rotator.GetActiveProfile()
		if profile != nil {
			apiKey, err := m.crypto.Decrypt(profile.EncryptedKey)
			if err != nil {
				return "", "", fmt.Errorf("failed to decrypt profile key: %w", err)
			}
			return apiKey, profile.ID, nil
		}
		// All profiles in cooldown
		if rotator.AllInCooldown() {
			waitTime := rotator.NextAvailableIn()
			return "", "", fmt.Errorf("all API keys in cooldown, next available in %v", waitTime)
		}
	}

	// Fall back to single key mode
	apiKey, err := m.GetDecryptedKey(providerID)
	if err != nil {
		return "", "", err
	}
	return apiKey, "", nil // Empty profileID means single key mode
}

// MarkProfileFailed marks a profile as failed (for rotation)
func (m *Manager) MarkProfileFailed(providerID, profileID string, reason string) {
	m.mu.RLock()
	rotator, ok := m.rotators[providerID]
	m.mu.RUnlock()

	if ok && rotator != nil && profileID != "" {
		// Convert reason string to FailoverReason
		var failReason = classifyReasonString(reason)
		rotator.MarkFailed(profileID, failReason)
	}
}

// MarkProfileSuccess marks a profile as successful (for rotation)
func (m *Manager) MarkProfileSuccess(providerID, profileID string) {
	m.mu.RLock()
	rotator, ok := m.rotators[providerID]
	m.mu.RUnlock()

	if ok && rotator != nil && profileID != "" {
		rotator.MarkSuccess(profileID)
	}
}

// AddAuthProfile adds a new auth profile for a provider
func (m *Manager) AddAuthProfile(providerID string, apiKey string, priority int) (*AuthProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[providerID]; !ok {
		return nil, ErrProviderNotFound
	}

	encrypted, err := m.crypto.Encrypt(apiKey)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	profile := &AuthProfile{
		ID:           fmt.Sprintf("prof-%s", time.Now().Format("20060102150405")),
		ProviderID:   providerID,
		EncryptedKey: encrypted,
		KeyMasked:    MaskAPIKey(apiKey),
		Priority:     priority,
		IsEnabled:    true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Get or create rotator
	rotator, ok := m.rotators[providerID]
	if !ok {
		rotator = NewProfileRotator(nil, nil)
		m.rotators[providerID] = rotator
	}
	rotator.AddProfile(profile)

	return profile, nil
}

// ListAuthProfiles returns all auth profiles for a provider
// If no profiles exist but a legacy key is configured, automatically migrates it
func (m *Manager) ListAuthProfiles(providerID string) []*AuthProfile {
	m.mu.Lock()
	defer m.mu.Unlock()

	rotator, ok := m.rotators[providerID]

	// Auto-migrate: if no rotator or empty profiles, check for legacy key
	if (!ok || rotator.GetAllProfiles() == nil || len(rotator.GetAllProfiles()) == 0) {
		if keyData, hasKey := m.keys[providerID]; hasKey && keyData.EncryptedKey != "" {
			// Decrypt the legacy key
			decrypted, err := m.crypto.Decrypt(keyData.EncryptedKey)
			if err == nil {
				// Create auth profile with priority 0
				encrypted, _ := m.crypto.Encrypt(decrypted)
				profile := &AuthProfile{
					ID:           fmt.Sprintf("prof-migrated-%d", time.Now().Unix()),
					ProviderID:   providerID,
					EncryptedKey: encrypted,
					KeyMasked:    keyData.KeyMasked,
					Priority:     0,
					IsEnabled:    true,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				// Get or create rotator
				if !ok {
					rotator = NewProfileRotator(nil, nil)
					m.rotators[providerID] = rotator
				}
				rotator.AddProfile(profile)
			}
		}
	}

	if !ok {
		return nil
	}
	return rotator.GetAllProfiles()
}

// RemoveAuthProfile removes an auth profile
func (m *Manager) RemoveAuthProfile(providerID, profileID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rotator, ok := m.rotators[providerID]
	if !ok {
		return ErrProviderNotFound
	}
	if !rotator.RemoveProfile(profileID) {
		return errors.New("profile not found")
	}
	return nil
}

// GetRotatorStats returns rotation statistics for a provider
func (m *Manager) GetRotatorStats(providerID string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rotator, ok := m.rotators[providerID]
	if !ok {
		return nil
	}

	profiles := rotator.GetAllProfiles()
	stats := map[string]interface{}{
		"total_profiles":    len(profiles),
		"active_profiles":   rotator.ActiveCount(),
		"all_in_cooldown":   rotator.AllInCooldown(),
		"next_available_in": rotator.NextAvailableIn().String(),
	}
	return stats
}

// classifyReasonString converts a reason string to apperr.FailoverReason
func classifyReasonString(reason string) apperr.FailoverReason {
	reason = strings.ToLower(reason)
	if strings.Contains(reason, "rate") || strings.Contains(reason, "429") {
		return apperr.ReasonRateLimit
	}
	if strings.Contains(reason, "auth") || strings.Contains(reason, "401") || strings.Contains(reason, "403") {
		return apperr.ReasonAuthFailed
	}
	if strings.Contains(reason, "timeout") {
		return apperr.ReasonTimeout
	}
	if strings.Contains(reason, "overload") || strings.Contains(reason, "503") {
		return apperr.ReasonOverloaded
	}
	return apperr.ReasonUnknown
}
