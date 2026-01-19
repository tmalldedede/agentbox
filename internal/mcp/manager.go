package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager MCP Server 管理器
type Manager struct {
	dataDir string
	servers map[string]*Server
	mu      sync.RWMutex
}

// NewManager 创建 Manager
func NewManager(dataDir string) (*Manager, error) {
	m := &Manager{
		dataDir: dataDir,
		servers: make(map[string]*Server),
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// 加载内置 MCP Servers
	m.loadBuiltInServers()

	// 加载用户自定义 MCP Servers
	if err := m.loadServers(); err != nil {
		return nil, err
	}

	return m, nil
}

// loadBuiltInServers 加载内置 MCP Servers
func (m *Manager) loadBuiltInServers() {
	builtIns := []*Server{
		{
			ID:          "filesystem",
			Name:        "Filesystem",
			Description: "Access and manage files on the local filesystem",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "/workspace"},
			Type:        ServerTypeStdio,
			Category:    CategoryFilesystem,
			Tags:        []string{"files", "read", "write"},
			IsBuiltIn:   true,
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "fetch",
			Name:        "Fetch",
			Description: "Make HTTP requests to external APIs",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-fetch"},
			Type:        ServerTypeStdio,
			Category:    CategoryAPI,
			Tags:        []string{"http", "api", "request"},
			IsBuiltIn:   true,
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "memory",
			Name:        "Memory",
			Description: "Knowledge graph-based memory for persistent context",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
			Type:        ServerTypeStdio,
			Category:    CategoryMemory,
			Tags:        []string{"memory", "knowledge", "graph"},
			IsBuiltIn:   true,
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "puppeteer",
			Name:        "Puppeteer",
			Description: "Browser automation and web scraping",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-puppeteer"},
			Type:        ServerTypeStdio,
			Category:    CategoryBrowser,
			Tags:        []string{"browser", "automation", "scraping"},
			IsBuiltIn:   true,
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "postgres",
			Name:        "PostgreSQL",
			Description: "Query and manage PostgreSQL databases",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-postgres"},
			Env: map[string]string{
				"POSTGRES_CONNECTION_STRING": "",
			},
			Type:      ServerTypeStdio,
			Category:  CategoryDatabase,
			Tags:      []string{"database", "sql", "postgres"},
			IsBuiltIn: true,
			IsEnabled: false, // 需要配置连接字符串
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "sqlite",
			Name:        "SQLite",
			Description: "Query and manage SQLite databases",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-sqlite"},
			Type:        ServerTypeStdio,
			Category:    CategoryDatabase,
			Tags:        []string{"database", "sql", "sqlite"},
			IsBuiltIn:   true,
			IsEnabled:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "github",
			Name:        "GitHub",
			Description: "Interact with GitHub repositories, issues, and pull requests",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-github"},
			Env: map[string]string{
				"GITHUB_PERSONAL_ACCESS_TOKEN": "",
			},
			Type:      ServerTypeStdio,
			Category:  CategoryAPI,
			Tags:      []string{"github", "git", "repository"},
			IsBuiltIn: true,
			IsEnabled: false, // 需要配置 token
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "slack",
			Name:        "Slack",
			Description: "Interact with Slack workspaces",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-slack"},
			Env: map[string]string{
				"SLACK_BOT_TOKEN": "",
				"SLACK_TEAM_ID":   "",
			},
			Type:      ServerTypeStdio,
			Category:  CategoryAPI,
			Tags:      []string{"slack", "chat", "messaging"},
			IsBuiltIn: true,
			IsEnabled: false, // 需要配置 token
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, s := range builtIns {
		m.servers[s.ID] = s
	}
}

// loadServers 从文件加载自定义 Servers
func (m *Manager) loadServers() error {
	indexPath := filepath.Join(m.dataDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，忽略
		}
		return err
	}

	var servers []*Server
	if err := json.Unmarshal(data, &servers); err != nil {
		return err
	}

	for _, s := range servers {
		// 不覆盖内置 Server
		if existing, ok := m.servers[s.ID]; ok && existing.IsBuiltIn {
			continue
		}
		m.servers[s.ID] = s
	}

	return nil
}

// saveServers 保存自定义 Servers 到文件
func (m *Manager) saveServers() error {
	var servers []*Server
	for _, s := range m.servers {
		if !s.IsBuiltIn {
			servers = append(servers, s)
		}
	}

	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(m.dataDir, "index.json")
	return os.WriteFile(indexPath, data, 0644)
}

// List 列出所有 MCP Servers
func (m *Manager) List() []*Server {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*Server, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	return servers
}

// ListEnabled 列出所有启用的 MCP Servers
func (m *Manager) ListEnabled() []*Server {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []*Server
	for _, s := range m.servers {
		if s.IsEnabled {
			servers = append(servers, s)
		}
	}
	return servers
}

// ListByCategory 按类别列出 MCP Servers
func (m *Manager) ListByCategory(category Category) []*Server {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []*Server
	for _, s := range m.servers {
		if s.Category == category {
			servers = append(servers, s)
		}
	}
	return servers
}

// Get 获取指定 MCP Server
func (m *Manager) Get(id string) (*Server, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.servers[id]
	if !ok {
		return nil, ErrServerNotFound
	}
	return s, nil
}

// GetMultiple 批量获取 MCP Servers
func (m *Manager) GetMultiple(ids []string) ([]*Server, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*Server, 0, len(ids))
	for _, id := range ids {
		if s, ok := m.servers[id]; ok {
			servers = append(servers, s)
		}
	}
	return servers, nil
}

// Create 创建 MCP Server
func (m *Manager) Create(req *CreateServerRequest) (*Server, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查 ID 是否已存在
	if _, ok := m.servers[req.ID]; ok {
		return nil, ErrServerAlreadyExists
	}

	now := time.Now()
	server := &Server{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Command:     req.Command,
		Args:        req.Args,
		Env:         req.Env,
		WorkDir:     req.WorkDir,
		Type:        req.Type,
		Category:    req.Category,
		Tags:        req.Tags,
		URL:         req.URL,
		IsBuiltIn:   false,
		IsEnabled:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 默认类型
	if server.Type == "" {
		server.Type = ServerTypeStdio
	}
	if server.Category == "" {
		server.Category = CategoryOther
	}

	if err := server.Validate(); err != nil {
		return nil, err
	}

	m.servers[server.ID] = server

	if err := m.saveServers(); err != nil {
		delete(m.servers, server.ID)
		return nil, err
	}

	return server, nil
}

// Update 更新 MCP Server
func (m *Manager) Update(id string, req *UpdateServerRequest) (*Server, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[id]
	if !ok {
		return nil, ErrServerNotFound
	}

	if server.IsBuiltIn {
		// 内置 Server 只允许更新 Env 和 IsEnabled
		if req.Env != nil {
			server.Env = *req.Env
		}
		if req.IsEnabled != nil {
			server.IsEnabled = *req.IsEnabled
		}
		server.UpdatedAt = time.Now()

		if err := m.saveServers(); err != nil {
			return nil, err
		}
		return server, nil
	}

	// 非内置 Server 可以更新所有字段
	if req.Name != nil {
		server.Name = *req.Name
	}
	if req.Description != nil {
		server.Description = *req.Description
	}
	if req.Command != nil {
		server.Command = *req.Command
	}
	if req.Args != nil {
		server.Args = req.Args
	}
	if req.Env != nil {
		server.Env = *req.Env
	}
	if req.WorkDir != nil {
		server.WorkDir = *req.WorkDir
	}
	if req.Type != nil {
		server.Type = *req.Type
	}
	if req.Category != nil {
		server.Category = *req.Category
	}
	if req.Tags != nil {
		server.Tags = req.Tags
	}
	if req.URL != nil {
		server.URL = *req.URL
	}
	if req.IsEnabled != nil {
		server.IsEnabled = *req.IsEnabled
	}

	server.UpdatedAt = time.Now()

	if err := server.Validate(); err != nil {
		return nil, err
	}

	if err := m.saveServers(); err != nil {
		return nil, err
	}

	return server, nil
}

// Delete 删除 MCP Server
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[id]
	if !ok {
		return ErrServerNotFound
	}

	if server.IsBuiltIn {
		return ErrServerIsBuiltIn
	}

	delete(m.servers, id)

	return m.saveServers()
}

// Clone 克隆 MCP Server
func (m *Manager) Clone(id, newID, newName string) (*Server, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[id]
	if !ok {
		return nil, ErrServerNotFound
	}

	if _, exists := m.servers[newID]; exists {
		return nil, ErrServerAlreadyExists
	}

	clone := server.Clone()
	clone.ID = newID
	clone.Name = newName
	clone.IsBuiltIn = false
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()

	if err := clone.Validate(); err != nil {
		return nil, err
	}

	m.servers[clone.ID] = clone

	if err := m.saveServers(); err != nil {
		delete(m.servers, clone.ID)
		return nil, err
	}

	return clone, nil
}

// Test 测试 MCP Server 连接
func (m *Manager) Test(id string) error {
	m.mu.RLock()
	server, ok := m.servers[id]
	m.mu.RUnlock()

	if !ok {
		return ErrServerNotFound
	}

	// TODO: 实现实际的连接测试
	// 目前只是检查配置是否有效
	return server.Validate()
}
