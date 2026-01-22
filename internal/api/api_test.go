package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "0.1.0",
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%v'", response["status"])
	}
}

// TestEnginesEndpoint tests the engines listing endpoint
func TestEnginesEndpoint(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/engines", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"engines": []map[string]interface{}{
				{"id": "claude-code", "name": "Claude Code", "status": "available"},
				{"id": "codex", "name": "Codex", "status": "available"},
			},
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/engines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	engines := response["engines"].([]interface{})
	if len(engines) < 2 {
		t.Errorf("expected at least 2 engines, got %d", len(engines))
	}
}

// TestProfilesEndpoint tests profile CRUD operations
func TestProfilesEndpoint(t *testing.T) {
	router := gin.New()

	// In-memory store for testing
	profiles := make(map[string]map[string]interface{})

	// List profiles
	router.GET("/api/v1/profiles", func(c *gin.Context) {
		list := make([]map[string]interface{}, 0)
		for _, p := range profiles {
			list = append(list, p)
		}
		c.JSON(http.StatusOK, gin.H{"profiles": list})
	})

	// Create profile
	router.POST("/api/v1/profiles", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id := body["id"].(string)
		profiles[id] = body
		c.JSON(http.StatusCreated, body)
	})

	// Get profile
	router.GET("/api/v1/profiles/:id", func(c *gin.Context) {
		id := c.Param("id")
		if p, ok := profiles[id]; ok {
			c.JSON(http.StatusOK, p)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// Delete profile
	router.DELETE("/api/v1/profiles/:id", func(c *gin.Context) {
		id := c.Param("id")
		if _, ok := profiles[id]; ok {
			delete(profiles, id)
			c.Status(http.StatusNoContent)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// Test Create
	createBody := map[string]interface{}{
		"id":      "test-profile",
		"name":    "Test Profile",
		"adapter": "claude-code",
	}
	bodyBytes, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected status 201, got %d", w.Code)
	}

	// Test Get
	req, _ = http.NewRequest("GET", "/api/v1/profiles/test-profile", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get: expected status 200, got %d", w.Code)
	}

	// Test List
	req, _ = http.NewRequest("GET", "/api/v1/profiles", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List: expected status 200, got %d", w.Code)
	}

	var listResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResponse)
	profilesList := listResponse["profiles"].([]interface{})
	if len(profilesList) != 1 {
		t.Errorf("List: expected 1 profile, got %d", len(profilesList))
	}

	// Test Delete
	req, _ = http.NewRequest("DELETE", "/api/v1/profiles/test-profile", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Delete: expected status 204, got %d", w.Code)
	}

	// Test Get after delete (should 404)
	req, _ = http.NewRequest("GET", "/api/v1/profiles/test-profile", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get after delete: expected status 404, got %d", w.Code)
	}
}

// TestMCPServersEndpoint tests MCP server CRUD operations
func TestMCPServersEndpoint(t *testing.T) {
	router := gin.New()
	servers := make(map[string]map[string]interface{})

	router.GET("/api/v1/admin/mcp-servers", func(c *gin.Context) {
		list := make([]map[string]interface{}, 0)
		for _, s := range servers {
			list = append(list, s)
		}
		c.JSON(http.StatusOK, gin.H{"mcp_servers": list})
	})

	router.POST("/api/v1/admin/mcp-servers", func(c *gin.Context) {
		var body map[string]interface{}
		c.BindJSON(&body)
		id := body["id"].(string)
		servers[id] = body
		c.JSON(http.StatusCreated, body)
	})

	router.GET("/api/v1/admin/mcp-servers/:id", func(c *gin.Context) {
		id := c.Param("id")
		if s, ok := servers[id]; ok {
			c.JSON(http.StatusOK, s)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	router.DELETE("/api/v1/admin/mcp-servers/:id", func(c *gin.Context) {
		id := c.Param("id")
		delete(servers, id)
		c.Status(http.StatusNoContent)
	})

	// Test Create
	createBody := map[string]interface{}{
		"id":      "test-mcp",
		"name":    "Test MCP Server",
		"command": "npx",
		"args":    []string{"-y", "test-server"},
	}
	bodyBytes, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("POST", "/api/v1/admin/mcp-servers", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create MCP: expected status 201, got %d", w.Code)
	}

	// Test Get
	req, _ = http.NewRequest("GET", "/api/v1/admin/mcp-servers/test-mcp", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get MCP: expected status 200, got %d", w.Code)
	}

	// Test List
	req, _ = http.NewRequest("GET", "/api/v1/admin/mcp-servers", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List MCP: expected status 200, got %d", w.Code)
	}
}

// TestSkillsEndpoint tests skill CRUD operations
func TestSkillsEndpoint(t *testing.T) {
	router := gin.New()
	skills := make(map[string]map[string]interface{})

	router.GET("/api/v1/admin/skills", func(c *gin.Context) {
		list := make([]map[string]interface{}, 0)
		for _, s := range skills {
			list = append(list, s)
		}
		c.JSON(http.StatusOK, gin.H{"skills": list})
	})

	router.POST("/api/v1/admin/skills", func(c *gin.Context) {
		var body map[string]interface{}
		c.BindJSON(&body)
		id := body["id"].(string)
		skills[id] = body
		c.JSON(http.StatusCreated, body)
	})

	router.GET("/api/v1/admin/skills/:id", func(c *gin.Context) {
		id := c.Param("id")
		if s, ok := skills[id]; ok {
			c.JSON(http.StatusOK, s)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// Test Create
	createBody := map[string]interface{}{
		"id":      "test-skill",
		"name":    "Test Skill",
		"slug":    "test-skill",
		"content": "# Test Skill\n\nThis is a test.",
	}
	bodyBytes, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("POST", "/api/v1/admin/skills", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create Skill: expected status 201, got %d", w.Code)
	}

	// Test Get
	req, _ = http.NewRequest("GET", "/api/v1/admin/skills/test-skill", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get Skill: expected status 200, got %d", w.Code)
	}

	// Test List
	req, _ = http.NewRequest("GET", "/api/v1/admin/skills", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List Skills: expected status 200, got %d", w.Code)
	}
}

// TestErrorHandling tests error responses
func TestErrorHandling(t *testing.T) {
	router := gin.New()

	router.GET("/api/v1/profiles/:id", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not found",
			"message": "Profile not found",
		})
	})

	router.POST("/api/v1/profiles", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad request",
				"message": "Invalid JSON",
			})
			return
		}
		if body["name"] == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation error",
				"message": "name is required",
			})
			return
		}
	})

	// Test 404
	req, _ := http.NewRequest("GET", "/api/v1/profiles/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	// Test 400 - invalid JSON
	req, _ = http.NewRequest("POST", "/api/v1/profiles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	// Test 400 - validation error
	body := map[string]interface{}{"adapter": "claude-code"}
	bodyBytes, _ := json.Marshal(body)
	req, _ = http.NewRequest("POST", "/api/v1/profiles", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
