package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "agentbox-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		Driver:   "sqlite",
		DSN:      dbPath,
		LogLevel: "silent",
	}

	err = Initialize(cfg)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}

	// Verify DB is not nil
	if DB == nil {
		t.Error("DB is nil after initialization")
	}

	// Verify tables were created
	tables := []string{"mcp_servers", "skills", "sessions", "tasks", "executions", "webhooks", "images", "history"}
	for _, table := range tables {
		if !DB.Migrator().HasTable(table) {
			t.Errorf("table %s was not created", table)
		}
	}
}

func TestMCPServerRepository(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "agentbox-test-*")
	defer os.RemoveAll(tmpDir)

	Initialize(Config{Driver: "sqlite", DSN: filepath.Join(tmpDir, "test.db"), LogLevel: "silent"})
	defer Close()

	repo := NewMCPServerRepository()

	server := &MCPServerModel{
		BaseModel:   BaseModel{ID: "test-mcp"},
		Name:        "Test MCP",
		Command:     "npx",
		Args:        `["-y", "test-server"]`,
		IsEnabled:   true,
	}

	// Create
	if err := repo.Create(server); err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	// Get
	got, err := repo.Get("test-mcp")
	if err != nil {
		t.Fatalf("failed to get MCP server: %v", err)
	}
	if got.Name != "Test MCP" {
		t.Errorf("expected name 'Test MCP', got '%s'", got.Name)
	}

	// GetByName
	byName, err := repo.GetByName("Test MCP")
	if err != nil {
		t.Fatalf("failed to get MCP server by name: %v", err)
	}
	if byName.ID != "test-mcp" {
		t.Errorf("expected ID 'test-mcp', got '%s'", byName.ID)
	}

	// List
	servers, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list MCP servers: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}

	// Delete
	if err := repo.Delete("test-mcp"); err != nil {
		t.Fatalf("failed to delete MCP server: %v", err)
	}
}

func TestSkillRepository(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "agentbox-test-*")
	defer os.RemoveAll(tmpDir)

	Initialize(Config{Driver: "sqlite", DSN: filepath.Join(tmpDir, "test.db"), LogLevel: "silent"})
	defer Close()

	repo := NewSkillRepository()

	skill := &SkillModel{
		BaseModel:   BaseModel{ID: "test-skill"},
		Name:        "Test Skill",
		Slug:        "test-skill",
		Description: "A test skill",
		Content:     "# Test Skill\n\nThis is a test.",
		IsEnabled:   true,
	}

	// Create
	if err := repo.Create(skill); err != nil {
		t.Fatalf("failed to create skill: %v", err)
	}

	// Get
	got, err := repo.Get("test-skill")
	if err != nil {
		t.Fatalf("failed to get skill: %v", err)
	}
	if got.Name != "Test Skill" {
		t.Errorf("expected name 'Test Skill', got '%s'", got.Name)
	}

	// GetBySlug
	bySlug, err := repo.GetBySlug("test-skill")
	if err != nil {
		t.Fatalf("failed to get skill by slug: %v", err)
	}
	if bySlug.ID != "test-skill" {
		t.Errorf("expected ID 'test-skill', got '%s'", bySlug.ID)
	}

	// List
	skills, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list skills: %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(skills))
	}

	// Delete
	if err := repo.Delete("test-skill"); err != nil {
		t.Fatalf("failed to delete skill: %v", err)
	}
}

func TestSeedBuiltInData(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "agentbox-test-*")
	defer os.RemoveAll(tmpDir)

	Initialize(Config{Driver: "sqlite", DSN: filepath.Join(tmpDir, "test.db"), LogLevel: "silent"})
	defer Close()

	// Seed data
	err := SeedBuiltInData()
	if err != nil {
		t.Fatalf("failed to seed built-in data: %v", err)
	}

	// Verify MCP servers were seeded
	var mcpCount int64
	DB.Model(&MCPServerModel{}).Where("is_built_in = ?", true).Count(&mcpCount)
	t.Logf("built-in MCP servers seeded: %d", mcpCount)

	// Verify idempotency - running again should not create duplicates
	err = SeedBuiltInData()
	if err != nil {
		t.Fatalf("second seed failed: %v", err)
	}

	var newMCPCount int64
	DB.Model(&MCPServerModel{}).Where("is_built_in = ?", true).Count(&newMCPCount)
	if newMCPCount != mcpCount {
		t.Errorf("seed is not idempotent: count changed from %d to %d", mcpCount, newMCPCount)
	}
}
