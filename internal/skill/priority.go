package skill

import (
	"sort"
)

// Priority 优先级顺序（数值越小优先级越高）
var sourcePriority = map[SkillSource]int{
	SourceExtra:     1, // 用户手动添加 - 最高优先级
	SourceBundled:   2, // 代码内置
	SourceManaged:   3, // 远程仓库安装
	SourceWorkspace: 4, // 工作区 - 最低优先级
}

// MergedSkillSet 合并后的 Skill 集合
type MergedSkillSet struct {
	skills      map[string]*Skill
	overridden  map[string][]SkillSource // 记录被覆盖的来源
	sourceCount map[SkillSource]int
}

// NewMergedSkillSet 创建合并集合
func NewMergedSkillSet() *MergedSkillSet {
	return &MergedSkillSet{
		skills:      make(map[string]*Skill),
		overridden:  make(map[string][]SkillSource),
		sourceCount: make(map[SkillSource]int),
	}
}

// Add 添加 Skill 到集合
// 高优先级来源的 Skill 会覆盖低优先级来源的同 ID Skill
func (m *MergedSkillSet) Add(s *Skill) bool {
	if s == nil {
		return false
	}

	existing, exists := m.skills[s.ID]
	if !exists {
		// 新 Skill，直接添加
		m.skills[s.ID] = s
		m.sourceCount[s.Source]++
		return true
	}

	// 比较优先级
	newPriority := sourcePriority[s.Source]
	existingPriority := sourcePriority[existing.Source]

	if newPriority < existingPriority {
		// 新 Skill 优先级更高，覆盖
		m.overridden[s.ID] = append(m.overridden[s.ID], existing.Source)
		m.sourceCount[existing.Source]--
		m.skills[s.ID] = s
		m.sourceCount[s.Source]++
		return true
	}

	// 新 Skill 优先级更低，记录被覆盖
	m.overridden[s.ID] = append(m.overridden[s.ID], s.Source)
	return false
}

// AddAll 批量添加 Skills
func (m *MergedSkillSet) AddAll(skills []*Skill) {
	for _, s := range skills {
		m.Add(s)
	}
}

// Get 获取 Skill
func (m *MergedSkillSet) Get(id string) *Skill {
	return m.skills[id]
}

// List 列出所有 Skills
func (m *MergedSkillSet) List() []*Skill {
	result := make([]*Skill, 0, len(m.skills))
	for _, s := range m.skills {
		result = append(result, s)
	}
	return result
}

// ListSorted 按优先级和名称排序列出
func (m *MergedSkillSet) ListSorted() []*Skill {
	result := m.List()
	sort.Slice(result, func(i, j int) bool {
		// 先按优先级排序
		pi := sourcePriority[result[i].Source]
		pj := sourcePriority[result[j].Source]
		if pi != pj {
			return pi < pj
		}
		// 同优先级按名称排序
		return result[i].Name < result[j].Name
	})
	return result
}

// ListBySource 按来源列出
func (m *MergedSkillSet) ListBySource(source SkillSource) []*Skill {
	var result []*Skill
	for _, s := range m.skills {
		if s.Source == source {
			result = append(result, s)
		}
	}
	return result
}

// Count 返回总数
func (m *MergedSkillSet) Count() int {
	return len(m.skills)
}

// CountBySource 按来源统计
func (m *MergedSkillSet) CountBySource() map[SkillSource]int {
	return m.sourceCount
}

// GetOverridden 获取被覆盖的来源
func (m *MergedSkillSet) GetOverridden(id string) []SkillSource {
	return m.overridden[id]
}

// WasOverridden 检查 Skill 是否覆盖了其他来源
func (m *MergedSkillSet) WasOverridden(id string) bool {
	return len(m.overridden[id]) > 0
}

// Merge 合并多个层级的 Skills
// 参数顺序：优先级从高到低
// 例如：Merge(extraSkills, bundledSkills, managedSkills, workspaceSkills)
func Merge(layers ...[]*Skill) *MergedSkillSet {
	result := NewMergedSkillSet()

	// 按顺序添加（高优先级在前）
	for _, layer := range layers {
		result.AddAll(layer)
	}

	return result
}

// MergeWithSource 合并多个层级并自动设置来源
func MergeWithSource(extra, bundled, managed, workspace []*Skill) *MergedSkillSet {
	result := NewMergedSkillSet()

	// Extra（最高优先级）
	for _, s := range extra {
		s.Source = SourceExtra
		result.Add(s)
	}

	// Bundled
	for _, s := range bundled {
		s.Source = SourceBundled
		result.Add(s)
	}

	// Managed
	for _, s := range managed {
		s.Source = SourceManaged
		result.Add(s)
	}

	// Workspace（最低优先级）
	for _, s := range workspace {
		s.Source = SourceWorkspace
		result.Add(s)
	}

	return result
}

// PriorityManager 优先级管理器
// 统一管理不同来源的 Skills
type PriorityManager struct {
	extra     []*Skill
	bundled   []*Skill
	managed   []*Skill
	workspace map[string][]*Skill // workspace path -> skills
	merged    *MergedSkillSet
	dirty     bool // 是否需要重新合并
}

// NewPriorityManager 创建优先级管理器
func NewPriorityManager() *PriorityManager {
	return &PriorityManager{
		workspace: make(map[string][]*Skill),
	}
}

// SetExtra 设置用户手动添加的 Skills
func (p *PriorityManager) SetExtra(skills []*Skill) {
	for _, s := range skills {
		s.Source = SourceExtra
	}
	p.extra = skills
	p.dirty = true
}

// SetBundled 设置内置 Skills
func (p *PriorityManager) SetBundled(skills []*Skill) {
	for _, s := range skills {
		s.Source = SourceBundled
	}
	p.bundled = skills
	p.dirty = true
}

// SetManaged 设置远程安装的 Skills
func (p *PriorityManager) SetManaged(skills []*Skill) {
	for _, s := range skills {
		s.Source = SourceManaged
	}
	p.managed = skills
	p.dirty = true
}

// SetWorkspaceSkills 设置工作区 Skills
func (p *PriorityManager) SetWorkspaceSkills(workspacePath string, skills []*Skill) {
	for _, s := range skills {
		s.Source = SourceWorkspace
		s.SourcePath = workspacePath
	}
	p.workspace[workspacePath] = skills
	p.dirty = true
}

// RemoveWorkspace 移除工作区
func (p *PriorityManager) RemoveWorkspace(workspacePath string) {
	delete(p.workspace, workspacePath)
	p.dirty = true
}

// GetMerged 获取合并后的 Skill 集合
func (p *PriorityManager) GetMerged() *MergedSkillSet {
	if p.dirty || p.merged == nil {
		p.rebuild()
	}
	return p.merged
}

// rebuild 重新合并
func (p *PriorityManager) rebuild() {
	// 收集所有工作区 Skills
	var allWorkspace []*Skill
	for _, skills := range p.workspace {
		allWorkspace = append(allWorkspace, skills...)
	}

	p.merged = MergeWithSource(p.extra, p.bundled, p.managed, allWorkspace)
	p.dirty = false
}

// List 列出所有合并后的 Skills
func (p *PriorityManager) List() []*Skill {
	return p.GetMerged().ListSorted()
}

// Get 获取 Skill
func (p *PriorityManager) Get(id string) *Skill {
	return p.GetMerged().Get(id)
}

// Stats 返回统计信息
func (p *PriorityManager) Stats() map[string]int {
	merged := p.GetMerged()
	bySource := merged.CountBySource()

	return map[string]int{
		"total":     merged.Count(),
		"extra":     bySource[SourceExtra],
		"bundled":   bySource[SourceBundled],
		"managed":   bySource[SourceManaged],
		"workspace": bySource[SourceWorkspace],
	}
}

// GetPriority 获取来源的优先级值
func GetPriority(source SkillSource) int {
	if p, ok := sourcePriority[source]; ok {
		return p
	}
	return 99 // 未知来源，最低优先级
}

// ComparePriority 比较两个来源的优先级
// 返回: -1 if a < b (a优先级更高), 0 if equal, 1 if a > b
func ComparePriority(a, b SkillSource) int {
	pa, pb := GetPriority(a), GetPriority(b)
	if pa < pb {
		return -1
	}
	if pa > pb {
		return 1
	}
	return 0
}
