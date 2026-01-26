package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager Skill 管理器
type Manager struct {
	dataDir string
	skills  map[string]*Skill
	mu      sync.RWMutex

	// 新组件
	loader   *Loader
	watcher  *Watcher
	priority *PriorityManager
}

// NewManager 创建 Manager
func NewManager(dataDir string) (*Manager, error) {
	m := &Manager{
		dataDir:  dataDir,
		skills:   make(map[string]*Skill),
		priority: NewPriorityManager(),
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// 加载内置 Skills
	m.loadBuiltInSkills()

	// 加载用户自定义 Skills
	if err := m.loadSkills(); err != nil {
		return nil, err
	}

	// 初始化 Loader
	m.loader = NewLoader(m, dataDir)

	// 初始化 Watcher
	m.watcher = NewWatcher(m)

	// 设置优先级管理器
	m.rebuildPriority()

	return m, nil
}

// rebuildPriority 重建优先级管理器
func (m *Manager) rebuildPriority() {
	var bundled, managed, extra []*Skill

	for _, s := range m.skills {
		switch {
		case s.IsBuiltIn:
			bundled = append(bundled, s)
		case s.Source == SourceExtra:
			extra = append(extra, s)
		case s.Source == SourceManaged:
			managed = append(managed, s)
		default:
			// 未设置来源的自定义 Skill 视为用户添加
			extra = append(extra, s)
		}
	}

	m.priority.SetBundled(bundled)
	m.priority.SetManaged(managed)
	m.priority.SetExtra(extra)
}

// GetLoader 获取 Loader
func (m *Manager) GetLoader() *Loader {
	return m.loader
}

// GetWatcher 获取 Watcher
func (m *Manager) GetWatcher() *Watcher {
	return m.watcher
}

// GetPriorityManager 获取优先级管理器
func (m *Manager) GetPriorityManager() *PriorityManager {
	return m.priority
}

// loadBuiltInSkills 加载内置 Skills
func (m *Manager) loadBuiltInSkills() {
	builtIns := []*Skill{
		{
			ID:          "commit",
			Name:        "Commit",
			Description: "Generate intelligent commit messages based on staged changes",
			Command:     "/commit",
			Prompt: `Analyze the staged changes and generate a commit message following conventional commits format.

## Instructions
1. Run 'git diff --cached' to see staged changes
2. Analyze the changes and determine the type (feat, fix, refactor, docs, test, chore)
3. Write a concise commit message with:
   - Type and optional scope
   - Brief description (50 chars or less)
   - Optional body with more details
4. Execute the commit with the generated message`,
			Category:  CategoryCoding,
			Tags:      []string{"git", "commit", "automation"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "review-pr",
			Name:        "Review PR",
			Description: "Review pull request code changes and provide feedback",
			Command:     "/review-pr",
			Prompt: `Review the pull request and provide constructive feedback.

## Instructions
1. Fetch the PR diff or file changes
2. Review code for:
   - Correctness and logic errors
   - Security vulnerabilities
   - Performance issues
   - Code style and best practices
   - Test coverage
3. Provide specific, actionable feedback
4. Highlight both issues and good practices`,
			Category:  CategoryReview,
			Tags:      []string{"review", "pr", "code-quality"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "explain",
			Name:        "Explain Code",
			Description: "Explain how the selected code works",
			Command:     "/explain",
			Prompt: `Explain the selected code in detail.

## Instructions
1. Read the provided code carefully
2. Explain:
   - What the code does at a high level
   - How each part works step by step
   - Any algorithms or patterns used
   - Dependencies and side effects
3. Use clear, simple language
4. Provide examples if helpful`,
			Category:  CategoryDocs,
			Tags:      []string{"explain", "documentation", "learning"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "refactor",
			Name:        "Refactor",
			Description: "Suggest refactoring improvements for code",
			Command:     "/refactor",
			Prompt: `Analyze the code and suggest refactoring improvements.

## Instructions
1. Review the provided code
2. Identify:
   - Code smells and anti-patterns
   - Duplication
   - Complexity issues
   - Naming improvements
3. Suggest specific refactoring:
   - Extract methods/functions
   - Simplify conditionals
   - Improve abstractions
4. Explain the benefits of each change`,
			Category:  CategoryCoding,
			Tags:      []string{"refactor", "clean-code", "improvement"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "test",
			Name:        "Generate Tests",
			Description: "Generate unit tests for code",
			Command:     "/test",
			Prompt: `Generate comprehensive unit tests for the provided code.

## Instructions
1. Analyze the code to understand its functionality
2. Identify test cases:
   - Happy path scenarios
   - Edge cases
   - Error handling
   - Boundary conditions
3. Write tests using the project's testing framework
4. Include setup/teardown if needed
5. Use descriptive test names`,
			Category:  CategoryTesting,
			Tags:      []string{"test", "unit-test", "quality"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "docs",
			Name:        "Generate Docs",
			Description: "Generate documentation for code",
			Command:     "/docs",
			Prompt: `Generate documentation for the provided code.

## Instructions
1. Analyze the code structure
2. Generate:
   - Function/method documentation
   - Parameter descriptions
   - Return value documentation
   - Usage examples
3. Follow the project's documentation style
4. Include any important notes or warnings`,
			Category:  CategoryDocs,
			Tags:      []string{"docs", "documentation", "comments"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "security",
			Name:        "Security Audit",
			Description: "Perform security audit on code",
			Command:     "/security",
			Prompt: `Perform a security audit on the provided code.

## Instructions
1. Scan for common vulnerabilities:
   - Injection attacks (SQL, XSS, Command)
   - Authentication/Authorization issues
   - Sensitive data exposure
   - Security misconfigurations
   - OWASP Top 10
2. Identify:
   - Hardcoded secrets
   - Insecure dependencies
   - Missing input validation
3. Provide remediation advice
4. Rate severity of each issue`,
			Category:  CategorySecurity,
			Tags:      []string{"security", "audit", "vulnerability"},
			IsBuiltIn: true,
			IsEnabled: true,
			Source:    SourceBundled,
			Invocation: InvocationPolicy{
				UserInvocable: true,
				AutoInvocable: false,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, s := range builtIns {
		m.skills[s.ID] = s
	}
}

// loadSkills 从文件加载自定义 Skills
func (m *Manager) loadSkills() error {
	indexPath := filepath.Join(m.dataDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var skills []*Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		return err
	}

	for _, s := range skills {
		// 不覆盖内置 Skill
		if existing, ok := m.skills[s.ID]; ok && existing.IsBuiltIn {
			continue
		}
		m.skills[s.ID] = s
	}

	return nil
}

// saveSkills 保存自定义 Skills 到文件
func (m *Manager) saveSkills() error {
	var skills []*Skill
	for _, s := range m.skills {
		if !s.IsBuiltIn {
			skills = append(skills, s)
		}
	}

	data, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(m.dataDir, "index.json")
	return os.WriteFile(indexPath, data, 0644)
}

// List 列出所有 Skills
func (m *Manager) List() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skills := make([]*Skill, 0, len(m.skills))
	for _, s := range m.skills {
		skills = append(skills, s)
	}
	return skills
}

// ListEnabled 列出所有启用的 Skills
func (m *Manager) ListEnabled() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var skills []*Skill
	for _, s := range m.skills {
		if s.IsEnabled {
			skills = append(skills, s)
		}
	}
	return skills
}

// ListByCategory 按类别列出 Skills
func (m *Manager) ListByCategory(category Category) []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var skills []*Skill
	for _, s := range m.skills {
		if s.Category == category {
			skills = append(skills, s)
		}
	}
	return skills
}

// Get 获取指定 Skill
func (m *Manager) Get(id string) (*Skill, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.skills[id]
	if !ok {
		return nil, ErrSkillNotFound
	}
	return s, nil
}

// GetByCommand 通过命令获取 Skill
func (m *Manager) GetByCommand(command string) (*Skill, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.skills {
		if s.Command == command && s.IsEnabled {
			return s, nil
		}
	}
	return nil, ErrSkillNotFound
}

// Create 创建 Skill
func (m *Manager) Create(req *CreateSkillRequest) (*Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查 ID 是否已存在
	if _, ok := m.skills[req.ID]; ok {
		return nil, ErrSkillAlreadyExists
	}

	now := time.Now()
	skill := &Skill{
		ID:           req.ID,
		Name:         req.Name,
		Description:  req.Description,
		Command:      req.Command,
		Prompt:       req.Prompt,
		Files:        req.Files,
		SourceDir:    req.SourceDir,
		AllowedTools: req.AllowedTools,
		RequiredMCP:  req.RequiredMCP,
		Category:     req.Category,
		Tags:         req.Tags,
		Author:       req.Author,
		Version:      req.Version,
		IsBuiltIn:    false,
		IsEnabled:    true,
		CreatedAt:    now,
		UpdatedAt:    now,

		// 新字段
		Source:       req.Source,
		Requirements: req.Requirements,
		Runtime:      req.Runtime,
	}

	// 设置调用策略
	if req.Invocation != nil {
		skill.Invocation = *req.Invocation
	} else {
		// 默认可用户调用
		skill.Invocation = InvocationPolicy{
			UserInvocable: true,
			AutoInvocable: false,
		}
	}

	// 默认来源
	if skill.Source == "" {
		skill.Source = SourceExtra
	}

	// 默认类别
	if skill.Category == "" {
		skill.Category = CategoryOther
	}

	if err := skill.Validate(); err != nil {
		return nil, err
	}

	m.skills[skill.ID] = skill

	if err := m.saveSkills(); err != nil {
		delete(m.skills, skill.ID)
		return nil, err
	}

	// 更新优先级管理器
	m.rebuildPriority()

	// 清除 Loader 缓存
	if m.loader != nil {
		m.loader.InvalidateCache(skill.ID)
	}

	return skill, nil
}

// Update 更新 Skill
func (m *Manager) Update(id string, req *UpdateSkillRequest) (*Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[id]
	if !ok {
		return nil, ErrSkillNotFound
	}

	if skill.IsBuiltIn {
		// 内置 Skill 只允许更新 IsEnabled
		if req.IsEnabled != nil {
			skill.IsEnabled = *req.IsEnabled
		}
		skill.UpdatedAt = time.Now()
		return skill, nil
	}

	// 非内置 Skill 可以更新所有字段
	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}
	if req.Command != nil {
		skill.Command = *req.Command
	}
	if req.Prompt != nil {
		skill.Prompt = *req.Prompt
	}
	if req.Files != nil {
		skill.Files = req.Files
	}
	if req.SourceDir != nil {
		skill.SourceDir = *req.SourceDir
	}
	if req.AllowedTools != nil {
		skill.AllowedTools = req.AllowedTools
	}
	if req.RequiredMCP != nil {
		skill.RequiredMCP = req.RequiredMCP
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if req.Tags != nil {
		skill.Tags = req.Tags
	}
	if req.Author != nil {
		skill.Author = *req.Author
	}
	if req.Version != nil {
		skill.Version = *req.Version
	}
	if req.IsEnabled != nil {
		skill.IsEnabled = *req.IsEnabled
	}

	// 新字段
	if req.Requirements != nil {
		skill.Requirements = req.Requirements
	}
	if req.Runtime != nil {
		skill.Runtime = req.Runtime
	}
	if req.Invocation != nil {
		skill.Invocation = *req.Invocation
	}

	skill.UpdatedAt = time.Now()

	if err := skill.Validate(); err != nil {
		return nil, err
	}

	if err := m.saveSkills(); err != nil {
		return nil, err
	}

	// 更新优先级管理器
	m.rebuildPriority()

	// 清除 Loader 缓存
	if m.loader != nil {
		m.loader.InvalidateCache(skill.ID)
	}

	return skill, nil
}

// Delete 删除 Skill
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[id]
	if !ok {
		return ErrSkillNotFound
	}

	if skill.IsBuiltIn {
		return ErrSkillIsBuiltIn
	}

	delete(m.skills, id)

	if err := m.saveSkills(); err != nil {
		return err
	}

	// 更新优先级管理器
	m.rebuildPriority()

	// 清除 Loader 缓存
	if m.loader != nil {
		m.loader.InvalidateCache(id)
	}

	return nil
}

// Clone 克隆 Skill
func (m *Manager) Clone(id, newID, newName string) (*Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	skill, ok := m.skills[id]
	if !ok {
		return nil, ErrSkillNotFound
	}

	if _, exists := m.skills[newID]; exists {
		return nil, ErrSkillAlreadyExists
	}

	clone := skill.Clone()
	clone.ID = newID
	clone.Name = newName
	clone.IsBuiltIn = false
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()

	if err := clone.Validate(); err != nil {
		return nil, err
	}

	m.skills[clone.ID] = clone

	if err := m.saveSkills(); err != nil {
		delete(m.skills, clone.ID)
		return nil, err
	}

	// 更新优先级管理器
	m.rebuildPriority()

	return clone, nil
}

// LoadWorkspaceSkills 加载工作区 Skills
func (m *Manager) LoadWorkspaceSkills(workspacePath string) ([]*Skill, error) {
	if m.watcher == nil {
		return nil, nil
	}
	skills, err := m.watcher.LoadWorkspaceSkills(workspacePath)
	if err != nil {
		return nil, err
	}

	// 更新优先级管理器
	m.priority.SetWorkspaceSkills(workspacePath, skills)

	return skills, nil
}

// GetMergedSkills 获取合并后的所有 Skills（包含工作区）
func (m *Manager) GetMergedSkills() []*Skill {
	return m.priority.List()
}

// GetMergedSkill 从合并集合中获取 Skill
func (m *Manager) GetMergedSkill(id string) *Skill {
	return m.priority.Get(id)
}

// GetSkillStats 获取 Skill 统计信息
func (m *Manager) GetSkillStats() map[string]int {
	return m.priority.Stats()
}

// ListBySource 按来源列出 Skills
func (m *Manager) ListBySource(source SkillSource) []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var skills []*Skill
	for _, s := range m.skills {
		if s.Source == source {
			skills = append(skills, s)
		}
	}
	return skills
}

// WatchWorkspace 开始监视指定工作区的 Skills 目录
func (m *Manager) WatchWorkspace(workspacePath string) error {
	if m.watcher == nil {
		return nil
	}

	// 先加载工作区 Skills
	skills, err := m.watcher.LoadWorkspaceSkills(workspacePath)
	if err != nil {
		log.Warn("failed to load workspace skills", "path", workspacePath, "error", err)
	}

	// 更新优先级管理器
	if len(skills) > 0 {
		m.priority.SetWorkspaceSkills(workspacePath, skills)
	}

	// 设置变更回调
	m.watcher.SetOnChange(func(workspace string, added, updated, removed []string) {
		log.Info("workspace skills changed",
			"workspace", workspace,
			"added", added,
			"updated", updated,
			"removed", removed,
		)

		// 刷新工作区 Skills
		refreshed, _ := m.watcher.GetWorkspaceSkills(workspace), error(nil)
		m.priority.SetWorkspaceSkills(workspace, refreshed)

		// 清除相关缓存
		if m.loader != nil {
			for _, id := range append(append(added, updated...), removed...) {
				m.loader.InvalidateCache(id)
			}
		}
	})

	// 开始监视
	return m.watcher.WatchWorkspace(workspacePath)
}

// UnwatchWorkspace 停止监视指定工作区
func (m *Manager) UnwatchWorkspace(workspacePath string) error {
	if m.watcher == nil {
		return nil
	}

	// 从优先级管理器移除
	m.priority.RemoveWorkspace(workspacePath)

	return m.watcher.UnwatchWorkspace(workspacePath)
}

// Stop 停止 Manager（清理资源）
func (m *Manager) Stop() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}
