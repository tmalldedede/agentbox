package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_ParseSkillMDContent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		content  string
		id       string
		expected struct {
			name        string
			command     string
			description string
			category    Category
		}
	}{
		{
			name: "with frontmatter",
			content: `---
name: Test Skill
description: A test skill
command: /test
category: coding
author: test
version: 1.0.0
---

This is the skill prompt.`,
			id: "test",
			expected: struct {
				name        string
				command     string
				description string
				category    Category
			}{
				name:        "Test Skill",
				command:     "/test",
				description: "A test skill",
				category:    CategoryCoding,
			},
		},
		{
			name:    "without frontmatter",
			content: "# My Skill\n\nThis is a simple skill.",
			id:      "my-skill",
			expected: struct {
				name        string
				command     string
				description string
				category    Category
			}{
				name:        "my-skill", // falls back to ID
				command:     "/my-skill",
				description: "",
				category:    CategoryOther,
			},
		},
		{
			name: "with invocation policy",
			content: `---
name: Auto Skill
command: /auto
invocation:
  user_invocable: true
  auto_invocable: true
  hook_invocable:
    - pre-commit
---

Auto trigger skill.`,
			id: "auto-skill",
			expected: struct {
				name        string
				command     string
				description string
				category    Category
			}{
				name:        "Auto Skill",
				command:     "/auto",
				description: "",
				category:    CategoryOther,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := parser.ParseSkillMDContent(tt.content, tt.id)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			if s.Name != tt.expected.name {
				t.Errorf("expected name %q, got %q", tt.expected.name, s.Name)
			}
			if s.Command != tt.expected.command {
				t.Errorf("expected command %q, got %q", tt.expected.command, s.Command)
			}
			if s.Description != tt.expected.description {
				t.Errorf("expected description %q, got %q", tt.expected.description, s.Description)
			}
			if s.Category != tt.expected.category {
				t.Errorf("expected category %q, got %q", tt.expected.category, s.Category)
			}
		})
	}
}

func TestParser_ParseSkillDir(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 skill 目录
	skillDir := filepath.Join(tmpDir, "my-skill")
	refsDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 SKILL.md
	skillMD := `---
name: My Skill
description: A test skill
command: /myskill
category: security
author: tester
version: 2.0.0
tags:
  - test
  - example
---

This is the main prompt content.

## Usage

Use this skill for testing.`

	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 skill.yaml
	skillYAML := `name: My Skill Override
version: 2.1.0
requirements:
  bins:
    - python3
    - curl
  env:
    - API_KEY
  pip:
    - requests
runtime:
  python: ">=3.8"
  memory: 512Mi`

	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建引用文件
	if err := os.WriteFile(filepath.Join(refsDir, "examples.md"), []byte("# Examples\n\nExample 1..."), 0644); err != nil {
		t.Fatal(err)
	}

	// 解析
	parser := NewParser()
	skill, err := parser.ParseSkillDir(skillDir, "my-skill")
	if err != nil {
		t.Fatalf("failed to parse skill dir: %v", err)
	}

	// 验证基本字段（skill.yaml 覆盖）
	if skill.Name != "My Skill Override" {
		t.Errorf("expected name 'My Skill Override', got %q", skill.Name)
	}
	if skill.Version != "2.1.0" {
		t.Errorf("expected version '2.1.0', got %q", skill.Version)
	}

	// 验证 command（从 SKILL.md）
	if skill.Command != "/myskill" {
		t.Errorf("expected command '/myskill', got %q", skill.Command)
	}

	// 验证 Requirements
	if skill.Requirements == nil {
		t.Fatal("expected requirements not nil")
	}
	if len(skill.Requirements.Bins) != 2 {
		t.Errorf("expected 2 bins, got %d", len(skill.Requirements.Bins))
	}
	if skill.Requirements.Bins[0] != "python3" {
		t.Errorf("expected bin 'python3', got %q", skill.Requirements.Bins[0])
	}
	if len(skill.Requirements.Env) != 1 || skill.Requirements.Env[0] != "API_KEY" {
		t.Errorf("unexpected env: %v", skill.Requirements.Env)
	}

	// 验证 Runtime
	if skill.Runtime == nil {
		t.Fatal("expected runtime not nil")
	}
	if skill.Runtime.Python != ">=3.8" {
		t.Errorf("expected python '>=3.8', got %q", skill.Runtime.Python)
	}
	if skill.Runtime.Memory != "512Mi" {
		t.Errorf("expected memory '512Mi', got %q", skill.Runtime.Memory)
	}

	// 验证引用文件
	found := false
	for _, f := range skill.Files {
		if f.Path == "references/examples.md" {
			found = true
			if f.Content == "" {
				t.Error("expected file content not empty")
			}
		}
	}
	if !found {
		t.Error("expected reference file to be loaded")
	}

	// 验证来源
	if skill.Source != SourceWorkspace {
		t.Errorf("expected source 'workspace', got %q", skill.Source)
	}
}

func TestParser_ScanSkillDirs(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "parser-scan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建多个 skill 目录
	for _, name := range []string{"skill-a", "skill-b", "skill-c"} {
		skillDir := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}
		skillMD := "---\nname: " + name + "\ncommand: /" + name + "\n---\n\nPrompt for " + name
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 创建无效目录（没有 SKILL.md）
	invalidDir := filepath.Join(tmpDir, "invalid-skill")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 扫描
	parser := NewParser()
	skills, err := parser.ScanSkillDirs(tmpDir)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}

	// 验证每个 skill
	ids := make(map[string]bool)
	for _, s := range skills {
		ids[s.ID] = true
	}

	for _, expected := range []string{"skill-a", "skill-b", "skill-c"} {
		if !ids[expected] {
			t.Errorf("expected skill %q", expected)
		}
	}
}

func TestParser_ValidateSkillDir(t *testing.T) {
	parser := NewParser()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "parser-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试目录不存在
	err = parser.ValidateSkillDir(filepath.Join(tmpDir, "not-exist"))
	if err == nil {
		t.Error("expected error for non-existent directory")
	}

	// 测试没有 SKILL.md
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}
	err = parser.ValidateSkillDir(emptyDir)
	if err == nil {
		t.Error("expected error for missing SKILL.md")
	}

	// 测试空 SKILL.md
	emptySkillDir := filepath.Join(tmpDir, "empty-skill")
	if err := os.MkdirAll(emptySkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(emptySkillDir, "SKILL.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	err = parser.ValidateSkillDir(emptySkillDir)
	if err == nil {
		t.Error("expected error for empty SKILL.md")
	}

	// 测试有效目录
	validDir := filepath.Join(tmpDir, "valid")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validDir, "SKILL.md"), []byte("# Valid\n\nContent"), 0644); err != nil {
		t.Fatal(err)
	}
	err = parser.ValidateSkillDir(validDir)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestParser_ExtractReferencePaths(t *testing.T) {
	parser := NewParser()

	content := `# Skill

See [examples](references/examples.md) for more.

Also check [template](references/templates/base.md).

External link: [Google](https://google.com)

Direct reference: references/data.json
`

	paths := parser.ExtractReferencePaths(content)

	expected := map[string]bool{
		"references/examples.md":        true,
		"references/templates/base.md":  true,
		"references/data.json":          true,
	}

	for _, p := range paths {
		if !expected[p] {
			t.Errorf("unexpected path: %s", p)
		}
		delete(expected, p)
	}

	for p := range expected {
		t.Errorf("missing expected path: %s", p)
	}
}

func TestParser_extractTitleFromMarkdown(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		content  string
		expected string
	}{
		{"# Hello World\n\nContent", "Hello World"},
		{"## Not a title\n\n# Actual Title", "Actual Title"},
		{"No title here", ""},
		{"---\nfrontmatter\n---\n# Title After", "Title After"},
	}

	for _, tt := range tests {
		result := parser.extractTitleFromMarkdown(tt.content)
		if result != tt.expected {
			t.Errorf("for content starting with %q: expected %q, got %q", tt.content[:20], tt.expected, result)
		}
	}
}
