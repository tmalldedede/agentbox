package skill

import (
	"testing"
	"time"
)

func createTestSkill(id string, source SkillSource) *Skill {
	return &Skill{
		ID:        id,
		Name:      id,
		Command:   "/" + id,
		Prompt:    "Test prompt for " + id,
		Source:    source,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestMergedSkillSet_Add(t *testing.T) {
	set := NewMergedSkillSet()

	// 添加第一个 Skill
	s1 := createTestSkill("test", SourceWorkspace)
	added := set.Add(s1)
	if !added {
		t.Error("expected first skill to be added")
	}

	// 验证已添加
	if set.Count() != 1 {
		t.Errorf("expected count 1, got %d", set.Count())
	}

	// 添加更高优先级的同 ID Skill（应该覆盖）
	s2 := createTestSkill("test", SourceBundled)
	added = set.Add(s2)
	if !added {
		t.Error("expected higher priority skill to override")
	}

	// 验证被覆盖
	if set.Count() != 1 {
		t.Errorf("expected count still 1, got %d", set.Count())
	}

	result := set.Get("test")
	if result.Source != SourceBundled {
		t.Errorf("expected source 'bundled', got %q", result.Source)
	}

	// 验证 overridden 记录
	overridden := set.GetOverridden("test")
	if len(overridden) != 1 {
		t.Errorf("expected 1 overridden, got %d", len(overridden))
	}
	if overridden[0] != SourceWorkspace {
		t.Errorf("expected overridden source 'workspace', got %q", overridden[0])
	}

	// 添加更低优先级的同 ID Skill（不应覆盖）
	s3 := createTestSkill("test", SourceManaged)
	added = set.Add(s3)
	if added {
		t.Error("expected lower priority skill not to override")
	}

	result = set.Get("test")
	if result.Source != SourceBundled {
		t.Errorf("expected source still 'bundled', got %q", result.Source)
	}
}

func TestMergedSkillSet_AddAll(t *testing.T) {
	set := NewMergedSkillSet()

	skills := []*Skill{
		createTestSkill("a", SourceExtra),
		createTestSkill("b", SourceBundled),
		createTestSkill("c", SourceManaged),
	}

	set.AddAll(skills)

	if set.Count() != 3 {
		t.Errorf("expected count 3, got %d", set.Count())
	}

	// 验证统计
	bySource := set.CountBySource()
	if bySource[SourceExtra] != 1 {
		t.Errorf("expected 1 extra, got %d", bySource[SourceExtra])
	}
	if bySource[SourceBundled] != 1 {
		t.Errorf("expected 1 bundled, got %d", bySource[SourceBundled])
	}
	if bySource[SourceManaged] != 1 {
		t.Errorf("expected 1 managed, got %d", bySource[SourceManaged])
	}
}

func TestMerge(t *testing.T) {
	extra := []*Skill{
		createTestSkill("a", SourceExtra),
		createTestSkill("common", SourceExtra),
	}

	bundled := []*Skill{
		createTestSkill("b", SourceBundled),
		createTestSkill("common", SourceBundled), // 应被 extra 覆盖
	}

	managed := []*Skill{
		createTestSkill("c", SourceManaged),
		createTestSkill("common", SourceManaged), // 应被覆盖
	}

	workspace := []*Skill{
		createTestSkill("d", SourceWorkspace),
		createTestSkill("common", SourceWorkspace), // 应被覆盖
	}

	result := Merge(extra, bundled, managed, workspace)

	// 验证总数
	if result.Count() != 5 { // a, b, c, d, common
		t.Errorf("expected count 5, got %d", result.Count())
	}

	// 验证 common 使用 extra 版本
	common := result.Get("common")
	if common == nil {
		t.Fatal("expected 'common' skill")
	}
	if common.Source != SourceExtra {
		t.Errorf("expected 'common' source 'extra', got %q", common.Source)
	}

	// 验证 overridden
	if !result.WasOverridden("common") {
		t.Error("expected 'common' to have overridden sources")
	}

	overridden := result.GetOverridden("common")
	if len(overridden) != 3 { // bundled, managed, workspace 都被覆盖
		t.Errorf("expected 3 overridden sources, got %d", len(overridden))
	}
}

func TestMergedSkillSet_ListSorted(t *testing.T) {
	set := NewMergedSkillSet()

	// 乱序添加
	set.Add(createTestSkill("z-workspace", SourceWorkspace))
	set.Add(createTestSkill("a-extra", SourceExtra))
	set.Add(createTestSkill("m-bundled", SourceBundled))
	set.Add(createTestSkill("b-extra", SourceExtra))

	sorted := set.ListSorted()

	// 验证排序：先按优先级，再按名称
	expected := []struct {
		id     string
		source SkillSource
	}{
		{"a-extra", SourceExtra},
		{"b-extra", SourceExtra},
		{"m-bundled", SourceBundled},
		{"z-workspace", SourceWorkspace},
	}

	if len(sorted) != len(expected) {
		t.Fatalf("expected %d skills, got %d", len(expected), len(sorted))
	}

	for i, exp := range expected {
		if sorted[i].ID != exp.id {
			t.Errorf("position %d: expected ID %q, got %q", i, exp.id, sorted[i].ID)
		}
		if sorted[i].Source != exp.source {
			t.Errorf("position %d: expected source %q, got %q", i, exp.source, sorted[i].Source)
		}
	}
}

func TestMergedSkillSet_ListBySource(t *testing.T) {
	set := NewMergedSkillSet()

	set.Add(createTestSkill("a", SourceExtra))
	set.Add(createTestSkill("b", SourceExtra))
	set.Add(createTestSkill("c", SourceBundled))
	set.Add(createTestSkill("d", SourceManaged))

	extras := set.ListBySource(SourceExtra)
	if len(extras) != 2 {
		t.Errorf("expected 2 extra skills, got %d", len(extras))
	}

	bundled := set.ListBySource(SourceBundled)
	if len(bundled) != 1 {
		t.Errorf("expected 1 bundled skill, got %d", len(bundled))
	}

	workspace := set.ListBySource(SourceWorkspace)
	if len(workspace) != 0 {
		t.Errorf("expected 0 workspace skills, got %d", len(workspace))
	}
}

func TestPriorityManager(t *testing.T) {
	pm := NewPriorityManager()

	// 设置各层
	pm.SetExtra([]*Skill{
		createTestSkill("user-skill", SourceExtra),
		createTestSkill("override", SourceExtra),
	})

	pm.SetBundled([]*Skill{
		createTestSkill("commit", SourceBundled),
		createTestSkill("override", SourceBundled), // 应被 extra 覆盖
	})

	pm.SetManaged([]*Skill{
		createTestSkill("installed-skill", SourceManaged),
	})

	pm.SetWorkspaceSkills("/workspace/a", []*Skill{
		createTestSkill("ws-skill-a", SourceWorkspace),
	})

	pm.SetWorkspaceSkills("/workspace/b", []*Skill{
		createTestSkill("ws-skill-b", SourceWorkspace),
	})

	// 测试 List
	all := pm.List()
	if len(all) != 6 {
		t.Errorf("expected 6 skills, got %d", len(all))
	}

	// 测试 Get
	s := pm.Get("override")
	if s == nil {
		t.Fatal("expected 'override' skill")
	}
	if s.Source != SourceExtra {
		t.Errorf("expected source 'extra', got %q", s.Source)
	}

	// 测试 Stats
	stats := pm.Stats()
	if stats["total"] != 6 {
		t.Errorf("expected total 6, got %d", stats["total"])
	}
	if stats["extra"] != 2 {
		t.Errorf("expected extra 2, got %d", stats["extra"])
	}
	if stats["bundled"] != 1 { // 1 个被覆盖
		t.Errorf("expected bundled 1, got %d", stats["bundled"])
	}
	if stats["managed"] != 1 {
		t.Errorf("expected managed 1, got %d", stats["managed"])
	}
	if stats["workspace"] != 2 {
		t.Errorf("expected workspace 2, got %d", stats["workspace"])
	}

	// 测试移除工作区
	pm.RemoveWorkspace("/workspace/a")
	all = pm.List()
	if len(all) != 5 {
		t.Errorf("expected 5 skills after remove, got %d", len(all))
	}
}

func TestGetPriority(t *testing.T) {
	tests := []struct {
		source   SkillSource
		expected int
	}{
		{SourceExtra, 1},
		{SourceBundled, 2},
		{SourceManaged, 3},
		{SourceWorkspace, 4},
		{"unknown", 99},
	}

	for _, tt := range tests {
		p := GetPriority(tt.source)
		if p != tt.expected {
			t.Errorf("GetPriority(%q): expected %d, got %d", tt.source, tt.expected, p)
		}
	}
}

func TestComparePriority(t *testing.T) {
	tests := []struct {
		a, b     SkillSource
		expected int
	}{
		{SourceExtra, SourceBundled, -1},   // extra 优先级更高
		{SourceBundled, SourceExtra, 1},    // bundled 优先级更低
		{SourceExtra, SourceExtra, 0},      // 相等
		{SourceWorkspace, SourceManaged, 1}, // workspace 最低
	}

	for _, tt := range tests {
		result := ComparePriority(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("ComparePriority(%q, %q): expected %d, got %d", tt.a, tt.b, tt.expected, result)
		}
	}
}

func TestMergeWithSource(t *testing.T) {
	// 创建不设置 Source 的 Skills
	extra := []*Skill{{ID: "a", Name: "A", Command: "/a", Prompt: "a"}}
	bundled := []*Skill{{ID: "b", Name: "B", Command: "/b", Prompt: "b"}}
	managed := []*Skill{{ID: "c", Name: "C", Command: "/c", Prompt: "c"}}
	workspace := []*Skill{{ID: "d", Name: "D", Command: "/d", Prompt: "d"}}

	result := MergeWithSource(extra, bundled, managed, workspace)

	// 验证 Source 已自动设置
	if result.Get("a").Source != SourceExtra {
		t.Error("expected 'a' source to be 'extra'")
	}
	if result.Get("b").Source != SourceBundled {
		t.Error("expected 'b' source to be 'bundled'")
	}
	if result.Get("c").Source != SourceManaged {
		t.Error("expected 'c' source to be 'managed'")
	}
	if result.Get("d").Source != SourceWorkspace {
		t.Error("expected 'd' source to be 'workspace'")
	}
}
