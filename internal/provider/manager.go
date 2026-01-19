package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Manager manages provider configurations
type Manager struct {
	mu        sync.RWMutex
	providers map[string]*Provider
	dataDir   string
}

// NewManager creates a new provider manager
func NewManager(dataDir string) *Manager {
	m := &Manager{
		providers: make(map[string]*Provider),
		dataDir:   dataDir,
	}

	// Load built-in providers
	for _, p := range BuiltinProviders {
		m.providers[p.ID] = p
	}

	// Load custom providers from disk
	m.loadCustomProviders()

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

// Create creates a new custom provider
func (m *Manager) Create(p *Provider) error {
	if err := p.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

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
		// Create a copy with modifications
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

	// Cannot delete built-in providers
	if p.IsBuiltIn {
		return nil // Silently ignore
	}

	delete(m.providers, id)
	return m.saveCustomProviders()
}

// customProvidersFile returns the path to the custom providers file
func (m *Manager) customProvidersFile() string {
	return filepath.Join(m.dataDir, "providers.json")
}

// loadCustomProviders loads custom providers from disk
func (m *Manager) loadCustomProviders() {
	data, err := os.ReadFile(m.customProvidersFile())
	if err != nil {
		return // File doesn't exist or can't be read
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

// saveCustomProviders saves custom providers to disk
func (m *Manager) saveCustomProviders() error {
	var custom []*Provider
	for _, p := range m.providers {
		if !p.IsBuiltIn {
			custom = append(custom, p)
		}
	}

	if len(custom) == 0 {
		// Remove file if no custom providers
		os.Remove(m.customProvidersFile())
		return nil
	}

	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.customProvidersFile(), data, 0644)
}
