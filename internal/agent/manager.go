package agent

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/mcp"
	"github.com/tmalldedede/agentbox/internal/provider"
	"github.com/tmalldedede/agentbox/internal/runtime"
	"github.com/tmalldedede/agentbox/internal/skill"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("agent")
}

// AgentFullConfig is the fully resolved configuration for an agent.
// It resolves all ID references (Provider, Runtime, Skills, MCPServers)
// into their actual objects, ready for session creation.
type AgentFullConfig struct {
	Agent      *Agent              `json:"agent"`
	Provider   *provider.Provider  `json:"provider"`
	Runtime    *runtime.AgentRuntime `json:"runtime"`
	Skills     []*skill.Skill      `json:"skills,omitempty"`
	MCPServers []*mcp.Server       `json:"mcp_servers,omitempty"`
}

// Manager manages agents
type Manager struct {
	agents      map[string]*Agent
	mu          sync.RWMutex
	configPath  string
	providerMgr *provider.Manager
	runtimeMgr  *runtime.Manager
	skillMgr    *skill.Manager
	mcpMgr      *mcp.Manager
}

// NewManager creates a new agent manager
func NewManager(configPath string, providerMgr *provider.Manager, runtimeMgr *runtime.Manager, skillMgr *skill.Manager, mcpMgr *mcp.Manager) *Manager {
	m := &Manager{
		agents:      make(map[string]*Agent),
		configPath:  configPath,
		providerMgr: providerMgr,
		runtimeMgr:  runtimeMgr,
		skillMgr:    skillMgr,
		mcpMgr:      mcpMgr,
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

// GetFullConfig resolves all references and returns the complete agent configuration.
// This is used by Session creation to get everything needed to start a container.
func (m *Manager) GetFullConfig(id string) (*AgentFullConfig, error) {
	agent, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	// Resolve Provider
	prov, err := m.providerMgr.Get(agent.ProviderID)
	if err != nil {
		return nil, ErrProviderNotFound
	}

	// Resolve Runtime (default if not specified)
	var rt *runtime.AgentRuntime
	if agent.RuntimeID != "" {
		rt, err = m.runtimeMgr.Get(agent.RuntimeID)
		if err != nil {
			return nil, ErrRuntimeNotFound
		}
	} else {
		rt = m.runtimeMgr.GetDefault()
	}

	// Resolve Skills
	var skills []*skill.Skill
	for _, sid := range agent.SkillIDs {
		s, err := m.skillMgr.Get(sid)
		if err != nil {
			log.Warn("skill not found, skipping", "skill_id", sid, "agent_id", id)
			continue
		}
		skills = append(skills, s)
	}

	// Resolve MCP Servers
	var mcpServers []*mcp.Server
	for _, mid := range agent.MCPServerIDs {
		s, err := m.mcpMgr.Get(mid)
		if err != nil {
			log.Warn("MCP server not found, skipping", "mcp_id", mid, "agent_id", id)
			continue
		}
		mcpServers = append(mcpServers, s)
	}

	return &AgentFullConfig{
		Agent:      agent,
		Provider:   prov,
		Runtime:    rt,
		Skills:     skills,
		MCPServers: mcpServers,
	}, nil
}

// Create creates a new agent
func (m *Manager) Create(agent *Agent) error {
	if err := agent.Validate(); err != nil {
		return err
	}

	// Verify provider exists
	if _, err := m.providerMgr.Get(agent.ProviderID); err != nil {
		return ErrProviderNotFound
	}

	// Verify runtime exists (if specified)
	if agent.RuntimeID != "" {
		if _, err := m.runtimeMgr.Get(agent.RuntimeID); err != nil {
			return ErrRuntimeNotFound
		}
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
	if agent.RuntimeID == "" {
		agent.RuntimeID = "default"
	}
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()

	m.agents[agent.ID] = agent
	m.saveToFile()

	log.Info("agent created", "id", agent.ID, "name", agent.Name, "provider", agent.ProviderID, "runtime", agent.RuntimeID)
	return nil
}

// Update updates an existing agent
func (m *Manager) Update(agent *Agent) error {
	if err := agent.Validate(); err != nil {
		return err
	}

	// Verify provider exists
	if _, err := m.providerMgr.Get(agent.ProviderID); err != nil {
		return ErrProviderNotFound
	}

	// Verify runtime exists (if specified)
	if agent.RuntimeID != "" {
		if _, err := m.runtimeMgr.Get(agent.RuntimeID); err != nil {
			return ErrRuntimeNotFound
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.agents[agent.ID]
	if !ok {
		return ErrAgentNotFound
	}

	// Preserve metadata
	agent.CreatedAt = existing.CreatedAt
	agent.IsBuiltIn = existing.IsBuiltIn
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

// GetProviderEnvVars returns the provider's environment variables (with decrypted API key)
// for the given agent. This is used by the session creation to inject credentials.
func (m *Manager) GetProviderEnvVars(agentID string) (map[string]string, error) {
	ag, err := m.Get(agentID)
	if err != nil {
		return nil, err
	}
	return m.providerMgr.GetEnvVarsWithKey(ag.ProviderID)
}

// --- Storage ---

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
