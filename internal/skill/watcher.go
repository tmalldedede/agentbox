package skill

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WorkspaceSkillPath 工作区 Skill 目录名
const WorkspaceSkillPath = ".claude/skills"

// Watcher 工作区 Skill 热加载监视器
type Watcher struct {
	parser     *Parser
	manager    *Manager
	watchers   map[string]*fsnotify.Watcher // workspace path -> watcher
	skills     map[string]map[string]*Skill // workspace path -> skill id -> skill
	mu         sync.RWMutex
	onChange   func(workspace string, added, updated, removed []string)
	debounceMs int
}

// NewWatcher 创建监视器
func NewWatcher(manager *Manager) *Watcher {
	return &Watcher{
		parser:     NewParser(),
		manager:    manager,
		watchers:   make(map[string]*fsnotify.Watcher),
		skills:     make(map[string]map[string]*Skill),
		debounceMs: 500,
	}
}

// SetOnChange 设置变更回调
func (w *Watcher) SetOnChange(fn func(workspace string, added, updated, removed []string)) {
	w.onChange = fn
}

// LoadWorkspaceSkills 加载工作区 Skills（不启动监视）
func (w *Watcher) LoadWorkspaceSkills(workspacePath string) ([]*Skill, error) {
	skillsDir := filepath.Join(workspacePath, WorkspaceSkillPath)

	// 检查目录是否存在
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, nil
	}

	// 扫描 Skills
	skills, err := w.parser.ScanSkillDirs(skillsDir)
	if err != nil {
		return nil, err
	}

	// 标记来源
	for _, s := range skills {
		s.Source = SourceWorkspace
		s.SourcePath = workspacePath
	}

	// 缓存
	w.mu.Lock()
	if w.skills[workspacePath] == nil {
		w.skills[workspacePath] = make(map[string]*Skill)
	}
	for _, s := range skills {
		w.skills[workspacePath][s.ID] = s
	}
	w.mu.Unlock()

	log.Info("loaded workspace skills", "workspace", workspacePath, "count", len(skills))
	return skills, nil
}

// WatchWorkspace 开始监视工作区
func (w *Watcher) WatchWorkspace(workspacePath string) error {
	skillsDir := filepath.Join(workspacePath, WorkspaceSkillPath)

	// 如果目录不存在，不报错，等待目录创建
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		log.Debug("workspace skills dir not found, skipping watch", "path", skillsDir)
		return nil
	}

	// 检查是否已在监视
	w.mu.Lock()
	if _, exists := w.watchers[workspacePath]; exists {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	// 创建 watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 添加目录到监视列表
	if err := w.addDirRecursive(watcher, skillsDir); err != nil {
		watcher.Close()
		return err
	}

	w.mu.Lock()
	w.watchers[workspacePath] = watcher
	w.mu.Unlock()

	// 启动监视 goroutine
	go w.watchLoop(workspacePath, watcher)

	log.Info("started watching workspace skills", "workspace", workspacePath)
	return nil
}

// UnwatchWorkspace 停止监视工作区
func (w *Watcher) UnwatchWorkspace(workspacePath string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if watcher, exists := w.watchers[workspacePath]; exists {
		watcher.Close()
		delete(w.watchers, workspacePath)
		delete(w.skills, workspacePath)
	}

	return nil
}

// addDirRecursive 递归添加目录到监视列表
func (w *Watcher) addDirRecursive(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// 跳过隐藏目录（.git 等），但不跳过 .claude
			name := filepath.Base(path)
			if strings.HasPrefix(name, ".") && name != ".claude" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
}

// watchLoop 监视循环
func (w *Watcher) watchLoop(workspacePath string, watcher *fsnotify.Watcher) {
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()
	pendingEvents := make(map[string]fsnotify.Op)

	processEvents := func() {
		if len(pendingEvents) == 0 {
			return
		}

		// 处理事件
		skillsDir := filepath.Join(workspacePath, WorkspaceSkillPath)
		affectedSkills := make(map[string]bool)

		for path := range pendingEvents {
			// 找到受影响的 skill
			relPath, err := filepath.Rel(skillsDir, path)
			if err != nil {
				continue
			}

			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 0 {
				skillID := parts[0]
				if skillID != "" && skillID != "." {
					affectedSkills[skillID] = true
				}
			}
		}

		// 重新加载受影响的 skills
		var added, updated, removed []string

		for skillID := range affectedSkills {
			skillDir := filepath.Join(skillsDir, skillID)
			if _, err := os.Stat(skillDir); os.IsNotExist(err) {
				// Skill 被删除
				w.mu.Lock()
				if w.skills[workspacePath] != nil {
					delete(w.skills[workspacePath], skillID)
				}
				w.mu.Unlock()
				removed = append(removed, skillID)
			} else {
				// Skill 被添加或更新
				skill, err := w.parser.ParseSkillDir(skillDir, skillID)
				if err != nil {
					log.Warn("failed to parse skill", "id", skillID, "error", err)
					continue
				}

				skill.Source = SourceWorkspace
				skill.SourcePath = workspacePath

				w.mu.Lock()
				if w.skills[workspacePath] == nil {
					w.skills[workspacePath] = make(map[string]*Skill)
				}
				_, exists := w.skills[workspacePath][skillID]
				w.skills[workspacePath][skillID] = skill
				w.mu.Unlock()

				if exists {
					updated = append(updated, skillID)
				} else {
					added = append(added, skillID)
				}
			}
		}

		// 触发回调
		if w.onChange != nil && (len(added) > 0 || len(updated) > 0 || len(removed) > 0) {
			w.onChange(workspacePath, added, updated, removed)
		}

		pendingEvents = make(map[string]fsnotify.Op)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 只关心 SKILL.md 和 skill.yaml 的变更
			fileName := filepath.Base(event.Name)
			if fileName != "SKILL.md" && fileName != "skill.yaml" &&
				!strings.Contains(event.Name, "references") {
				continue
			}

			pendingEvents[event.Name] = event.Op

			// 重置 debounce timer
			debounceTimer.Reset(time.Duration(w.debounceMs) * time.Millisecond)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("watcher error", "error", err)

		case <-debounceTimer.C:
			processEvents()
		}
	}
}

// GetWorkspaceSkills 获取指定工作区的 Skills
func (w *Watcher) GetWorkspaceSkills(workspacePath string) []*Skill {
	w.mu.RLock()
	defer w.mu.RUnlock()

	skillMap := w.skills[workspacePath]
	if skillMap == nil {
		return nil
	}

	skills := make([]*Skill, 0, len(skillMap))
	for _, s := range skillMap {
		skills = append(skills, s)
	}
	return skills
}

// GetWorkspaceSkill 获取工作区中指定的 Skill
func (w *Watcher) GetWorkspaceSkill(workspacePath, skillID string) *Skill {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if skillMap := w.skills[workspacePath]; skillMap != nil {
		return skillMap[skillID]
	}
	return nil
}

// GetAllWorkspaceSkills 获取所有工作区的 Skills
func (w *Watcher) GetAllWorkspaceSkills() []*Skill {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var skills []*Skill
	for _, skillMap := range w.skills {
		for _, s := range skillMap {
			skills = append(skills, s)
		}
	}
	return skills
}

// RefreshWorkspace 刷新工作区 Skills
func (w *Watcher) RefreshWorkspace(workspacePath string) ([]*Skill, error) {
	// 清除缓存
	w.mu.Lock()
	delete(w.skills, workspacePath)
	w.mu.Unlock()

	// 重新加载
	return w.LoadWorkspaceSkills(workspacePath)
}

// Close 关闭所有监视器
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for workspace, watcher := range w.watchers {
		watcher.Close()
		delete(w.watchers, workspace)
	}

	w.skills = make(map[string]map[string]*Skill)
	return nil
}

// IsWatching 检查是否正在监视指定工作区
func (w *Watcher) IsWatching(workspacePath string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, exists := w.watchers[workspacePath]
	return exists
}

// ListWatchedWorkspaces 列出所有正在监视的工作区
func (w *Watcher) ListWatchedWorkspaces() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	workspaces := make([]string, 0, len(w.watchers))
	for ws := range w.watchers {
		workspaces = append(workspaces, ws)
	}
	return workspaces
}
