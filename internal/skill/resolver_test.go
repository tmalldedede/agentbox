package skill

import (
	"testing"
	"time"
)

func TestRequirements_HasRequirements(t *testing.T) {
	tests := []struct {
		name     string
		reqs     *Requirements
		expected bool
	}{
		{
			name:     "nil requirements",
			reqs:     nil,
			expected: false,
		},
		{
			name:     "empty requirements",
			reqs:     &Requirements{},
			expected: false,
		},
		{
			name: "has bins",
			reqs: &Requirements{
				Bins: []string{"python3"},
			},
			expected: true,
		},
		{
			name: "has env",
			reqs: &Requirements{
				Env: []string{"API_KEY"},
			},
			expected: true,
		},
		{
			name: "has config",
			reqs: &Requirements{
				Config: []string{"mcpServers.test"},
			},
			expected: true,
		},
		{
			name: "has pip",
			reqs: &Requirements{
				Pip: []string{"requests"},
			},
			expected: true,
		},
		{
			name: "has npm",
			reqs: &Requirements{
				Npm: []string{"lodash"},
			},
			expected: true,
		},
		{
			name: "has multiple",
			reqs: &Requirements{
				Bins: []string{"python3"},
				Env:  []string{"API_KEY"},
				Pip:  []string{"requests"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reqs.HasRequirements()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMissingDeps_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		missing  *MissingDeps
		expected bool
	}{
		{
			name:     "empty",
			missing:  &MissingDeps{},
			expected: true,
		},
		{
			name: "has bins",
			missing: &MissingDeps{
				Bins: []string{"python3"},
			},
			expected: false,
		},
		{
			name: "has env",
			missing: &MissingDeps{
				Env: []string{"API_KEY"},
			},
			expected: false,
		},
		{
			name: "has config",
			missing: &MissingDeps{
				Config: []string{"test"},
			},
			expected: false,
		},
		{
			name: "has pip",
			missing: &MissingDeps{
				Pip: []string{"requests"},
			},
			expected: false,
		},
		{
			name: "has npm",
			missing: &MissingDeps{
				Npm: []string{"lodash"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.missing.IsEmpty()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResolver_QuickCheck(t *testing.T) {
	resolver := NewResolver(nil)

	// 测试无依赖的 Skill
	skillNoDeps := &Skill{
		ID:        "no-deps",
		Name:      "No Deps",
		Command:   "/no-deps",
		Prompt:    "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := resolver.QuickCheck(skillNoDeps)
	if !result.Satisfied {
		t.Error("expected satisfied for skill without requirements")
	}
	if result.Missing != nil && !result.Missing.IsEmpty() {
		t.Error("expected empty missing deps")
	}

	// 测试有依赖的 Skill
	skillWithDeps := &Skill{
		ID:      "with-deps",
		Name:    "With Deps",
		Command: "/with-deps",
		Prompt:  "Test",
		Requirements: &Requirements{
			Bins: []string{"python3", "curl"},
			Env:  []string{"API_KEY"},
			Pip:  []string{"requests"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result = resolver.QuickCheck(skillWithDeps)
	if result.Satisfied {
		t.Error("expected not satisfied for skill with requirements")
	}
	if result.Missing == nil {
		t.Fatal("expected missing deps")
	}
	if len(result.Missing.Bins) != 2 {
		t.Errorf("expected 2 missing bins, got %d", len(result.Missing.Bins))
	}
	if len(result.Missing.Env) != 1 {
		t.Errorf("expected 1 missing env, got %d", len(result.Missing.Env))
	}
	if len(result.Missing.Pip) != 1 {
		t.Errorf("expected 1 missing pip, got %d", len(result.Missing.Pip))
	}
}

func TestCheckResult(t *testing.T) {
	result := &CheckResult{
		SkillID:   "test",
		Satisfied: false,
		Missing: &MissingDeps{
			Bins: []string{"python3"},
		},
	}

	if result.SkillID != "test" {
		t.Errorf("expected skill_id 'test', got %q", result.SkillID)
	}
	if result.Satisfied {
		t.Error("expected not satisfied")
	}
	if len(result.Missing.Bins) != 1 {
		t.Errorf("expected 1 missing bin, got %d", len(result.Missing.Bins))
	}
}

// 注意：以下测试在本地环境运行，会使用真实的环境检查

func TestResolver_checkEnvLocal(t *testing.T) {
	// 测试本地环境变量检查逻辑
	// 使用真实的 os.LookupEnv，PATH 和 HOME 通常存在
	t.Setenv("HOME", "test-home")
	resolver := &Resolver{}

	// 只有 NONEXISTENT_VAR_12345 应该缺失
	missing := resolver.checkEnvLocal([]string{"PATH", "HOME", "NONEXISTENT_VAR_12345"})
	if len(missing) != 1 {
		t.Errorf("expected 1 missing (NONEXISTENT_VAR_12345), got %d: %v", len(missing), missing)
	}
	if len(missing) > 0 && missing[0] != "NONEXISTENT_VAR_12345" {
		t.Errorf("expected NONEXISTENT_VAR_12345 to be missing, got %s", missing[0])
	}
}

// TestResolver_Check_NoContainer 测试不指定容器时的本地检查行为
func TestResolver_Check_NoContainer(t *testing.T) {
	resolver := NewResolver(nil)

	// 使用本地存在的依赖（bash 和 HOME 在 Unix 系统通常存在）
	skill := &Skill{
		ID:      "test",
		Name:    "Test",
		Command: "/test",
		Prompt:  "Test",
		Requirements: &Requirements{
			Bins: []string{"bash"},
			Env:  []string{"HOME"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 本地检查：bash 和 HOME 通常存在，应该满足
	result, err := resolver.Check(nil, skill, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 在大多数 Unix 系统上应该满足
	if !result.Satisfied {
		t.Logf("missing: bins=%v, env=%v", result.Missing.Bins, result.Missing.Env)
		// 不报错，因为某些 CI 环境可能没有 bash
	}
}

// TestResolver_Check_MissingDeps 测试缺失依赖的场景
func TestResolver_Check_MissingDeps(t *testing.T) {
	resolver := NewResolver(nil)

	skill := &Skill{
		ID:      "test",
		Name:    "Test",
		Command: "/test",
		Prompt:  "Test",
		Requirements: &Requirements{
			Bins: []string{"nonexistent_binary_12345"},
			Env:  []string{"NONEXISTENT_ENV_VAR_12345"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := resolver.Check(nil, skill, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该不满足
	if result.Satisfied {
		t.Error("expected not satisfied when deps are missing")
	}
	if len(result.Missing.Bins) != 1 {
		t.Errorf("expected 1 missing bin, got %d", len(result.Missing.Bins))
	}
	if len(result.Missing.Env) != 1 {
		t.Errorf("expected 1 missing env, got %d", len(result.Missing.Env))
	}
}
