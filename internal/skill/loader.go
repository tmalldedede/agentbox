package skill

import (
	"os"
	"path/filepath"
	"sync"
)

// Loader 渐进式加载器
// 支持三级加载：metadata -> body -> full
type Loader struct {
	manager *Manager
	dataDir string

	// 缓存
	metadataCache map[string]*SkillMetadata
	bodyCache     map[string]*Skill
	fullCache     map[string]*Skill
	cacheMu       sync.RWMutex
}

// NewLoader 创建加载器
func NewLoader(manager *Manager, dataDir string) *Loader {
	return &Loader{
		manager:       manager,
		dataDir:       dataDir,
		metadataCache: make(map[string]*SkillMetadata),
		bodyCache:     make(map[string]*Skill),
		fullCache:     make(map[string]*Skill),
	}
}

// ListMetadata 快速列出所有 Skill 元数据
// 仅返回 ID, Name, Description, Command, Category 等基本信息
func (l *Loader) ListMetadata() []*SkillMetadata {
	skills := l.manager.List()
	result := make([]*SkillMetadata, len(skills))
	for i, s := range skills {
		result[i] = s.ToMetadata()
	}
	return result
}

// ListMetadataByCategory 按类别列出 Skill 元数据
func (l *Loader) ListMetadataByCategory(category Category) []*SkillMetadata {
	skills := l.manager.ListByCategory(category)
	result := make([]*SkillMetadata, len(skills))
	for i, s := range skills {
		result[i] = s.ToMetadata()
	}
	return result
}

// ListMetadataBySource 按来源列出 Skill 元数据
func (l *Loader) ListMetadataBySource(source SkillSource) []*SkillMetadata {
	skills := l.manager.List()
	var result []*SkillMetadata
	for _, s := range skills {
		if s.Source == source {
			result = append(result, s.ToMetadata())
		}
	}
	return result
}

// ListEnabledMetadata 列出启用的 Skill 元数据
func (l *Loader) ListEnabledMetadata() []*SkillMetadata {
	skills := l.manager.ListEnabled()
	result := make([]*SkillMetadata, len(skills))
	for i, s := range skills {
		result[i] = s.ToMetadata()
	}
	return result
}

// GetMetadata 获取单个 Skill 元数据
func (l *Loader) GetMetadata(id string) (*SkillMetadata, error) {
	s, err := l.manager.Get(id)
	if err != nil {
		return nil, err
	}
	return s.ToMetadata(), nil
}

// LoadBody 加载 Skill body（包含 prompt）
// 对于内置 Skill，直接返回；对于文件系统 Skill，按需读取
func (l *Loader) LoadBody(id string) (*Skill, error) {
	// 检查缓存
	l.cacheMu.RLock()
	if cached, ok := l.bodyCache[id]; ok {
		l.cacheMu.RUnlock()
		return cached, nil
	}
	l.cacheMu.RUnlock()

	// 从 Manager 获取
	s, err := l.manager.Get(id)
	if err != nil {
		return nil, err
	}

	// 创建副本以便缓存（避免返回 manager 的内部指针）
	copy := *s
	copy.LoadLevel = LoadLevelBody
	copy.BodyLoaded = true

	// 更新缓存
	l.cacheMu.Lock()
	l.bodyCache[id] = &copy
	l.cacheMu.Unlock()

	return &copy, nil
}

// LoadFull 完整加载 Skill（包含所有引用文件）
// 读取 SourceDir 中的所有文件（如 references/）
func (l *Loader) LoadFull(id string) (*Skill, error) {
	// 检查缓存
	l.cacheMu.RLock()
	if cached, ok := l.fullCache[id]; ok {
		l.cacheMu.RUnlock()
		return cached, nil
	}
	l.cacheMu.RUnlock()

	// 从 Manager 获取（不使用 LoadBody 以避免双重缓存）
	s, err := l.manager.Get(id)
	if err != nil {
		return nil, err
	}

	// 创建副本
	copy := *s
	copy.LoadLevel = LoadLevelFull
	copy.BodyLoaded = true

	// 如果有 SourceDir，读取引用文件
	if copy.SourceDir != "" {
		files, err := l.loadReferences(copy.SourceDir)
		if err != nil {
			log.Warn("failed to load references", "skill_id", id, "error", err)
		} else if len(files) > 0 {
			copy.Files = append(copy.Files, files...)
		}
	}

	// 更新缓存
	l.cacheMu.Lock()
	l.fullCache[id] = &copy
	l.cacheMu.Unlock()

	return &copy, nil
}

// loadReferences 从 SourceDir 加载引用文件
func (l *Loader) loadReferences(sourceDir string) ([]SkillFile, error) {
	refsDir := filepath.Join(sourceDir, "references")
	if _, err := os.Stat(refsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var files []SkillFile
	err := filepath.Walk(refsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}
		relPath = filepath.ToSlash(relPath)

		files = append(files, SkillFile{
			Path:    relPath,
			Content: string(content),
		})
		return nil
	})

	return files, err
}

// InvalidateCache 清除指定 Skill 的缓存
func (l *Loader) InvalidateCache(id string) {
	l.cacheMu.Lock()
	defer l.cacheMu.Unlock()

	delete(l.metadataCache, id)
	delete(l.bodyCache, id)
	delete(l.fullCache, id)
}

// InvalidateAll 清除所有缓存
func (l *Loader) InvalidateAll() {
	l.cacheMu.Lock()
	defer l.cacheMu.Unlock()

	l.metadataCache = make(map[string]*SkillMetadata)
	l.bodyCache = make(map[string]*Skill)
	l.fullCache = make(map[string]*Skill)
}

// LoadSkillsByLevel 根据加载级别批量加载
func (l *Loader) LoadSkillsByLevel(level LoadLevel) ([]*Skill, error) {
	switch level {
	case LoadLevelMetadata:
		// 对于 metadata 级别，返回精简的 Skill
		metas := l.ListMetadata()
		skills := make([]*Skill, len(metas))
		for i, m := range metas {
			skills[i] = &Skill{
				ID:          m.ID,
				Name:        m.Name,
				Description: m.Description,
				Command:     m.Command,
				Category:    m.Category,
				Tags:        m.Tags,
				Author:      m.Author,
				Version:     m.Version,
				Source:      m.Source,
				IsBuiltIn:   m.IsBuiltIn,
				IsEnabled:   m.IsEnabled,
				LoadLevel:   LoadLevelMetadata,
				UpdatedAt:   m.UpdatedAt,
			}
		}
		return skills, nil

	case LoadLevelBody:
		// 加载所有 body
		allSkills := l.manager.List()
		for _, s := range allSkills {
			s.LoadLevel = LoadLevelBody
			s.BodyLoaded = true
		}
		return allSkills, nil

	case LoadLevelFull:
		// 完整加载所有 Skill
		allSkills := l.manager.List()
		for _, s := range allSkills {
			if s.SourceDir != "" {
				files, err := l.loadReferences(s.SourceDir)
				if err == nil && len(files) > 0 {
					s.Files = append(s.Files, files...)
				}
			}
			s.LoadLevel = LoadLevelFull
			s.BodyLoaded = true
		}
		return allSkills, nil

	default:
		return l.manager.List(), nil
	}
}

// PreloadAll 预加载所有 Skill 到缓存
func (l *Loader) PreloadAll() {
	skills := l.manager.List()
	l.cacheMu.Lock()
	defer l.cacheMu.Unlock()

	for _, s := range skills {
		s.LoadLevel = LoadLevelBody
		s.BodyLoaded = true
		l.bodyCache[s.ID] = s
		l.metadataCache[s.ID] = s.ToMetadata()
	}
}
