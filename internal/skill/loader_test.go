package skill

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoader_ListMetadata(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 Manager
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	loader := mgr.GetLoader()

	// 测试列出元数据
	metas := loader.ListMetadata()
	if len(metas) == 0 {
		t.Error("expected at least one skill (built-in)")
	}

	// 验证元数据字段
	for _, m := range metas {
		if m.ID == "" {
			t.Error("metadata ID should not be empty")
		}
		if m.Name == "" {
			t.Error("metadata Name should not be empty")
		}
		if m.Command == "" {
			t.Error("metadata Command should not be empty")
		}
	}
}

func TestLoader_LoadBody(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	loader := mgr.GetLoader()

	// 加载内置 Skill
	s, err := loader.LoadBody("commit")
	if err != nil {
		t.Fatalf("failed to load body: %v", err)
	}

	if s.ID != "commit" {
		t.Errorf("expected ID 'commit', got %s", s.ID)
	}
	if s.Prompt == "" {
		t.Error("prompt should not be empty after LoadBody")
	}
	if s.LoadLevel != LoadLevelBody {
		t.Errorf("expected LoadLevel 'body', got %s", s.LoadLevel)
	}
	if !s.BodyLoaded {
		t.Error("BodyLoaded should be true")
	}
}

func TestLoader_LoadFull(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 Skill 目录和文件
	skillDir := filepath.Join(tmpDir, "skills", "test-skill")
	refsDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 SKILL.md
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill\n\nTest prompt"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建引用文件
	if err := os.WriteFile(filepath.Join(refsDir, "example.md"), []byte("Example content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 Manager 并添加 Skill
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgr.Create(&CreateSkillRequest{
		ID:        "test-skill",
		Name:      "Test Skill",
		Command:   "/test",
		Prompt:    "Test prompt",
		SourceDir: skillDir,
	})
	if err != nil {
		t.Fatalf("failed to create skill: %v", err)
	}

	loader := mgr.GetLoader()

	// 完整加载
	s, err := loader.LoadFull("test-skill")
	if err != nil {
		t.Fatalf("failed to load full: %v", err)
	}

	if s.LoadLevel != LoadLevelFull {
		t.Errorf("expected LoadLevel 'full', got %s", s.LoadLevel)
	}

	// 验证引用文件已加载
	found := false
	for _, f := range s.Files {
		if f.Path == "references/example.md" {
			found = true
			if f.Content != "Example content" {
				t.Errorf("unexpected file content: %s", f.Content)
			}
		}
	}
	if !found {
		t.Error("reference file not loaded")
	}
}

func TestLoader_Cache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	loader := mgr.GetLoader()

	// 第一次加载
	s1, err := loader.LoadBody("commit")
	if err != nil {
		t.Fatal(err)
	}

	// 第二次加载（应该从缓存）
	s2, err := loader.LoadBody("commit")
	if err != nil {
		t.Fatal(err)
	}

	// 应该是同一个指针
	if s1 != s2 {
		t.Error("expected same pointer from cache")
	}

	// 清除缓存
	loader.InvalidateCache("commit")

	// 再次加载（应该是新实例）
	s3, err := loader.LoadBody("commit")
	if err != nil {
		t.Fatal(err)
	}

	if s1 == s3 {
		t.Error("expected different pointer after cache invalidation")
	}
}

func TestLoader_ListMetadataByCategory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	loader := mgr.GetLoader()

	// 测试按类别筛选
	codingMetas := loader.ListMetadataByCategory(CategoryCoding)
	for _, m := range codingMetas {
		if m.Category != CategoryCoding {
			t.Errorf("expected category 'coding', got %s", m.Category)
		}
	}

	// commit 和 refactor 应该在 coding 类别
	found := false
	for _, m := range codingMetas {
		if m.ID == "commit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'commit' skill in coding category")
	}
}

func TestLoader_ListMetadataBySource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-loader-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	loader := mgr.GetLoader()

	// 测试按来源筛选
	bundledMetas := loader.ListMetadataBySource(SourceBundled)
	if len(bundledMetas) == 0 {
		t.Error("expected bundled skills")
	}

	for _, m := range bundledMetas {
		if m.Source != SourceBundled {
			t.Errorf("expected source 'bundled', got %s", m.Source)
		}
	}
}

func TestSkillMetadata_HasDeps(t *testing.T) {
	s := &Skill{
		ID:      "test",
		Name:    "Test",
		Command: "/test",
		Prompt:  "Test",
		Requirements: &Requirements{
			Bins: []string{"python3"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	meta := s.ToMetadata()
	if !meta.HasDeps {
		t.Error("expected HasDeps to be true")
	}

	// 无依赖
	s2 := &Skill{
		ID:        "test2",
		Name:      "Test2",
		Command:   "/test2",
		Prompt:    "Test2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	meta2 := s2.ToMetadata()
	if meta2.HasDeps {
		t.Error("expected HasDeps to be false")
	}
}
