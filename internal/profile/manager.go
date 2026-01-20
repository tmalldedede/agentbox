package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager manages Profiles
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
	dataDir  string
}

// NewManager creates a new Profile manager
func NewManager(dataDir string) (*Manager, error) {
	m := &Manager{
		profiles: make(map[string]*Profile),
		dataDir:  dataDir,
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// Load built-in profiles
	m.loadBuiltInProfiles()

	// Load user profiles from disk
	if err := m.loadFromDisk(); err != nil {
		return nil, err
	}

	return m, nil
}

// List returns all profiles
func (m *Manager) List() []*Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profiles := make([]*Profile, 0, len(m.profiles))
	for _, p := range m.profiles {
		profiles = append(profiles, p)
	}
	return profiles
}

// Get returns a profile by ID
func (m *Manager) Get(id string) (*Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.profiles[id]
	if !ok {
		return nil, ErrProfileNotFound
	}
	return p, nil
}

// GetResolved returns a profile with inheritance resolved
func (m *Manager) GetResolved(id string) (*Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.resolveInheritance(id, make(map[string]bool))
}

// resolveInheritance resolves profile inheritance chain
func (m *Manager) resolveInheritance(id string, visited map[string]bool) (*Profile, error) {
	if visited[id] {
		return nil, ErrProfileCircularExtends
	}
	visited[id] = true

	p, ok := m.profiles[id]
	if !ok {
		return nil, ErrProfileNotFound
	}

	// If no inheritance, return as-is
	if p.Extends == "" {
		resolved := p.Clone(p.ID, p.Name)
		return resolved, nil
	}

	// Resolve parent first
	parent, err := m.resolveInheritance(p.Extends, visited)
	if err != nil {
		if err == ErrProfileNotFound {
			return nil, ErrProfileParentNotFound
		}
		return nil, err
	}

	// Clone and merge
	resolved := p.Clone(p.ID, p.Name)
	resolved.Merge(parent)

	return resolved, nil
}

// Create creates a new profile
func (m *Manager) Create(p *Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.profiles[p.ID]; exists {
		return ErrProfileAlreadyExists
	}

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.IsBuiltIn = false

	m.profiles[p.ID] = p

	return m.saveToDisk(p)
}

// Update updates an existing profile
func (m *Manager) Update(p *Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.profiles[p.ID]
	if !ok {
		return ErrProfileNotFound
	}
	if existing.IsBuiltIn {
		return ErrProfileIsBuiltIn
	}

	p.CreatedAt = existing.CreatedAt
	p.UpdatedAt = time.Now()
	p.IsBuiltIn = false

	m.profiles[p.ID] = p

	return m.saveToDisk(p)
}

// Delete deletes a profile
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.profiles[id]
	if !ok {
		return ErrProfileNotFound
	}
	if p.IsBuiltIn {
		return ErrProfileIsBuiltIn
	}

	delete(m.profiles, id)

	// Remove from disk
	path := filepath.Join(m.dataDir, id+".json")
	return os.Remove(path)
}

// Clone clones a profile with a new ID
func (m *Manager) Clone(id, newID, newName string) (*Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.profiles[id]
	if !ok {
		return nil, ErrProfileNotFound
	}
	if _, exists := m.profiles[newID]; exists {
		return nil, ErrProfileAlreadyExists
	}

	clone := p.Clone(newID, newName)
	m.profiles[newID] = clone

	if err := m.saveToDisk(clone); err != nil {
		delete(m.profiles, newID)
		return nil, err
	}

	return clone, nil
}

// loadFromDisk loads user profiles from disk
func (m *Manager) loadFromDisk() error {
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(m.dataDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var p Profile
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}

		// Don't override built-in profiles
		if _, exists := m.profiles[p.ID]; !exists || !m.profiles[p.ID].IsBuiltIn {
			m.profiles[p.ID] = &p
		}
	}

	return nil
}

// saveToDisk saves a profile to disk
func (m *Manager) saveToDisk(p *Profile) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(m.dataDir, p.ID+".json")
	return os.WriteFile(path, data, 0644)
}

// loadBuiltInProfiles loads the built-in profiles
func (m *Manager) loadBuiltInProfiles() {
	// Claude Code Development
	m.profiles["claude-code-dev"] = &Profile{
		ID:          "claude-code-dev",
		Name:        "Claude Code 开发",
		Description: "Claude Code 通用开发配置",
		Adapter:     AdapterClaudeCode,
		Model: ModelConfig{
			Name: "sonnet",
		},
		Permissions: PermissionConfig{
			Mode: PermissionModeDefault,
		},
		Resources: ResourceConfig{
			MaxBudgetUSD: 10,
			CPUs:         4,
			MemoryMB:     4096,
		},
		IsBuiltIn: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Claude Code Full Auto
	m.profiles["claude-code-auto"] = &Profile{
		ID:          "claude-code-auto",
		Name:        "Claude Code 全自动",
		Description: "Claude Code 全自动模式，无需确认",
		Adapter:     AdapterClaudeCode,
		Model: ModelConfig{
			Name: "sonnet",
		},
		Permissions: PermissionConfig{
			Mode: PermissionModeDontAsk,
		},
		Resources: ResourceConfig{
			MaxBudgetUSD: 20,
			CPUs:         4,
			MemoryMB:     4096,
		},
		IsBuiltIn: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Codex Full Auto
	m.profiles["codex-full-auto"] = &Profile{
		ID:          "codex-full-auto",
		Name:        "Codex 全自动",
		Description: "Codex 全自动模式",
		Adapter:     AdapterCodex,
		Model: ModelConfig{
			Name: "o3",
		},
		Permissions: PermissionConfig{
			FullAuto:       true,
			SandboxMode:    SandboxModeWorkspaceWrite,
			ApprovalPolicy: ApprovalPolicyOnRequest,
		},
		Resources: ResourceConfig{
			CPUs:     2,
			MemoryMB: 2048,
		},
		IsBuiltIn: true,
		IsPublic:  true, // 在 API Playground 可见
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Codex Read-Only
	m.profiles["codex-readonly"] = &Profile{
		ID:          "codex-readonly",
		Name:        "Codex 只读",
		Description: "Codex 只读模式，安全分析代码",
		Adapter:     AdapterCodex,
		Model: ModelConfig{
			Name: "o3",
		},
		Permissions: PermissionConfig{
			SandboxMode:    SandboxModeReadOnly,
			ApprovalPolicy: ApprovalPolicyUntrusted,
		},
		Resources: ResourceConfig{
			CPUs:     2,
			MemoryMB: 2048,
		},
		IsBuiltIn: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Security Research
	m.profiles["security-research"] = &Profile{
		ID:          "security-research",
		Name:        "安全研究",
		Description: "安全研究与分析配置",
		Adapter:     AdapterClaudeCode,
		IsPublic:    true, // 在 API Playground 可见
		Model: ModelConfig{
			Name: "opus",
		},
		Permissions: PermissionConfig{
			Mode: PermissionModeDontAsk,
			AllowedTools: []string{
				"Bash",
				"Read",
				"Write",
				"Grep",
				"Glob",
				"WebFetch",
				"WebSearch",
			},
		},
		MCPServers: []MCPServerConfig{
			{
				Name:        "cybersec-cloud",
				Command:     "cybersec-mcp",
				Description: "网络安全工具集",
			},
		},
		Features: FeatureConfig{
			WebSearch: true,
		},
		Resources: ResourceConfig{
			MaxBudgetUSD: 50,
			CPUs:         4,
			MemoryMB:     8192,
		},
		IsBuiltIn: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Data Analysis
	m.profiles["data-analysis"] = &Profile{
		ID:          "data-analysis",
		Name:        "数据分析",
		Description: "数据分析与可视化配置",
		Adapter:     AdapterClaudeCode,
		IsPublic:    true, // 在 API Playground 可见
		Model: ModelConfig{
			Name: "sonnet",
		},
		Permissions: PermissionConfig{
			Mode: PermissionModeDefault,
			AllowedTools: []string{
				"Read",
				"Write",
				"Bash",
				"NotebookEdit",
			},
		},
		Resources: ResourceConfig{
			MaxBudgetUSD: 20,
			CPUs:         4,
			MemoryMB:     8192,
		},
		IsBuiltIn: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
