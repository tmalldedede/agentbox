package smartagent

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/profile"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("smartagent")
}

// Manager manages smart agents
type Manager struct {
	agents     map[string]*Agent
	mu         sync.RWMutex
	configPath string
	profileMgr *profile.Manager
}

// NewManager creates a new agent manager
func NewManager(configPath string, profileMgr *profile.Manager) *Manager {
	m := &Manager{
		agents:     make(map[string]*Agent),
		configPath: configPath,
		profileMgr: profileMgr,
	}
	m.loadFromFile()
	return m
}

// List returns all agents
func (m *Manager) List() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	return agents
}

// Get returns an agent by ID
func (m *Manager) Get(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[id]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// Create creates a new agent
func (m *Manager) Create(agent *Agent) error {
	if err := agent.Validate(); err != nil {
		return err
	}

	// Verify profile exists
	if _, err := m.profileMgr.Get(agent.ProfileID); err != nil {
		return ErrProfileNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agent.ID]; exists {
		return ErrAgentAlreadyExists
	}

	// Set defaults
	if agent.Status == "" {
		agent.Status = StatusActive
	}
	if agent.APIAccess == "" {
		agent.APIAccess = APIAccessPrivate
	}
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()

	m.agents[agent.ID] = agent
	m.saveToFile()

	log.Info("agent created", "id", agent.ID, "name", agent.Name, "profile", agent.ProfileID)
	return nil
}

// Update updates an existing agent
func (m *Manager) Update(agent *Agent) error {
	if err := agent.Validate(); err != nil {
		return err
	}

	// Verify profile exists
	if _, err := m.profileMgr.Get(agent.ProfileID); err != nil {
		return ErrProfileNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.agents[agent.ID]
	if !ok {
		return ErrAgentNotFound
	}

	// Preserve metadata
	agent.CreatedAt = existing.CreatedAt
	agent.UpdatedAt = time.Now()

	m.agents[agent.ID] = agent
	m.saveToFile()

	log.Info("agent updated", "id", agent.ID, "name", agent.Name)
	return nil
}

// Delete deletes an agent
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.agents[id]; !ok {
		return ErrAgentNotFound
	}

	delete(m.agents, id)
	m.saveToFile()

	log.Info("agent deleted", "id", id)
	return nil
}

// GetByProfileID returns all agents using a specific profile
func (m *Manager) GetByProfileID(profileID string) []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agents []*Agent
	for _, a := range m.agents {
		if a.ProfileID == profileID {
			agents = append(agents, a)
		}
	}
	return agents
}

// loadFromFile loads agents from config file
func (m *Manager) loadFromFile() {
	path := filepath.Join(m.configPath, "agents.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug("agents config file not found, starting fresh", "path", path)
			return
		}
		log.Error("failed to read agents config", "error", err)
		return
	}

	var agents []*Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		log.Error("failed to parse agents config", "error", err)
		return
	}

	for _, a := range agents {
		m.agents[a.ID] = a
	}

	log.Info("loaded agents", "count", len(agents))
}

// saveToFile saves agents to config file
func (m *Manager) saveToFile() {
	agents := make([]*Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}

	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		log.Error("failed to marshal agents", "error", err)
		return
	}

	path := filepath.Join(m.configPath, "agents.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Error("failed to create config dir", "error", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Error("failed to write agents config", "error", err)
		return
	}
}
