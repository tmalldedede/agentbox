package skill

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
	"gopkg.in/yaml.v3"
)

// 模块日志器
var log *slog.Logger

func init() {
	log = logger.Module("skill")
}

// SkillSource GitHub 仓库源
type SkillSource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`       // GitHub owner
	Repo        string    `json:"repo"`        // GitHub repo
	Branch      string    `json:"branch"`      // 分支，默认 main
	Path        string    `json:"path"`        // skills 目录路径，默认 skills
	Type        string    `json:"type"`        // official / community
	Description string    `json:"description"` // 描述
	Stars       int       `json:"stars"`       // GitHub stars
	UpdatedAt   time.Time `json:"updated_at"`  // 最后更新
	IsEnabled   bool      `json:"is_enabled"`
}

// RemoteSkill 远程 Skill 信息
type RemoteSkill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Category    string `json:"category"`
	Author      string `json:"author"`
	Version     string `json:"version"`
	SourceID    string `json:"source_id"`   // 来源仓库 ID
	SourceName  string `json:"source_name"` // 来源仓库名
	Path        string `json:"path"`        // 在仓库中的路径
	Stars       int    `json:"stars"`       // 估算的使用量
	IsInstalled bool   `json:"is_installed"`
}

// SkillStore Skill 商店管理器
type SkillStore struct {
	sources     map[string]*SkillSource
	cache       map[string][]RemoteSkill // sourceID -> skills
	cacheMu     sync.RWMutex
	cacheTime   map[string]time.Time
	httpClient  *http.Client
	manager     *Manager // 本地 skill 管理器
	githubToken string   // GitHub API Token (可选，用于提高 rate limit)
}

// NewSkillStore 创建商店管理器
func NewSkillStore(manager *Manager) *SkillStore {
	// 从环境变量读取 GitHub Token (可选)
	// 支持 GITHUB_TOKEN 或 GH_TOKEN
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GH_TOKEN")
	}

	store := &SkillStore{
		sources:   make(map[string]*SkillSource),
		cache:     make(map[string][]RemoteSkill),
		cacheTime: make(map[string]time.Time),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		manager:     manager,
		githubToken: githubToken,
	}

	// 添加默认源
	store.AddSource(&SkillSource{
		ID:          "anthropic-official",
		Name:        "Anthropic Official",
		Owner:       "anthropics",
		Repo:        "skills",
		Branch:      "main",
		Path:        "skills",
		Type:        "official",
		Description: "Official Anthropic Agent Skills",
		IsEnabled:   true,
	})

	store.AddSource(&SkillSource{
		ID:          "awesome-claude-skills",
		Name:        "Awesome Claude Skills",
		Owner:       "travisvn",
		Repo:        "awesome-claude-skills",
		Branch:      "main",
		Path:        "skills",
		Type:        "community",
		Description: "Community curated Claude Skills collection",
		IsEnabled:   true,
	})

	return store
}

// setGitHubHeaders 设置 GitHub API 请求头
func (s *SkillStore) setGitHubHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "AgentBox-SkillStore")
	if s.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.githubToken)
	}
}

// AddSource 添加仓库源
func (s *SkillStore) AddSource(source *SkillSource) {
	if source.Branch == "" {
		source.Branch = "main"
	}
	if source.Path == "" {
		source.Path = "skills"
	}
	// 新添加的源默认启用
	source.IsEnabled = true
	s.sources[source.ID] = source
}

// RemoveSource 移除仓库源
func (s *SkillStore) RemoveSource(id string) {
	delete(s.sources, id)
	s.cacheMu.Lock()
	delete(s.cache, id)
	delete(s.cacheTime, id)
	s.cacheMu.Unlock()
}

// ListSources 列出所有源
func (s *SkillStore) ListSources() []*SkillSource {
	sources := make([]*SkillSource, 0, len(s.sources))
	for _, src := range s.sources {
		sources = append(sources, src)
	}
	return sources
}

// GetSource 获取源
func (s *SkillStore) GetSource(id string) (*SkillSource, bool) {
	src, ok := s.sources[id]
	return src, ok
}

// FetchSkills 从 GitHub 获取 Skills
func (s *SkillStore) FetchSkills(ctx context.Context, sourceID string) ([]RemoteSkill, error) {
	source, ok := s.sources[sourceID]
	if !ok {
		return nil, fmt.Errorf("source not found: %s", sourceID)
	}

	// 检查缓存
	s.cacheMu.RLock()
	if cached, ok := s.cache[sourceID]; ok {
		if time.Since(s.cacheTime[sourceID]) < 5*time.Minute {
			s.cacheMu.RUnlock()
			return cached, nil
		}
	}
	s.cacheMu.RUnlock()

	// 获取仓库内容
	skills, err := s.fetchFromGitHub(ctx, source)
	if err != nil {
		return nil, err
	}

	// 标记已安装的
	installedSkills := s.manager.List()
	installedMap := make(map[string]bool)
	for _, sk := range installedSkills {
		installedMap[sk.ID] = true
	}
	for i := range skills {
		skills[i].IsInstalled = installedMap[skills[i].ID]
	}

	// 更新缓存
	s.cacheMu.Lock()
	s.cache[sourceID] = skills
	s.cacheTime[sourceID] = time.Now()
	s.cacheMu.Unlock()

	return skills, nil
}

// FetchAllSkills 从所有启用的源获取 Skills
func (s *SkillStore) FetchAllSkills(ctx context.Context) ([]RemoteSkill, error) {
	var allSkills []RemoteSkill

	for _, source := range s.sources {
		if !source.IsEnabled {
			continue
		}

		skills, err := s.FetchSkills(ctx, source.ID)
		if err != nil {
			// 记录错误但继续
			log.Warn("failed to fetch skills from source", "source_id", source.ID, "error", err)
			continue
		}

		allSkills = append(allSkills, skills...)
	}

	return allSkills, nil
}

// fetchFromGitHub 从 GitHub API 获取 skills
func (s *SkillStore) fetchFromGitHub(ctx context.Context, source *SkillSource) ([]RemoteSkill, error) {
	// 获取目录内容
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		source.Owner, source.Repo, source.Path, source.Branch)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	s.setGitHubHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 403 {
			return nil, fmt.Errorf("GitHub API rate limit exceeded. 请设置 GITHUB_TOKEN 环境变量: %s", string(body))
		}
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var contents []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	var skills []RemoteSkill

	// 遍历目录，查找 SKILL.md
	for _, item := range contents {
		if item.Type != "dir" {
			continue
		}

		// 获取 SKILL.md
		skill, err := s.fetchSkillMD(ctx, source, item.Path)
		if err != nil {
			// 跳过无效的 skill
			continue
		}

		skill.SourceID = source.ID
		skill.SourceName = source.Name
		skill.Path = item.Path
		skills = append(skills, *skill)
	}

	return skills, nil
}

// fetchSkillMD 获取并解析 SKILL.md
func (s *SkillStore) fetchSkillMD(ctx context.Context, source *SkillSource, skillPath string) (*RemoteSkill, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/SKILL.md?ref=%s",
		source.Owner, source.Repo, skillPath, source.Branch)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	s.setGitHubHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			return nil, fmt.Errorf("GitHub API rate limit exceeded")
		}
		return nil, fmt.Errorf("SKILL.md not found")
	}

	var fileContent struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileContent); err != nil {
		return nil, err
	}

	// 解码 base64
	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(fileContent.Content, "\n", ""))
	if err != nil {
		return nil, err
	}

	// 解析 SKILL.md
	return parseSkillMD(string(content), skillPath)
}

// parseSkillMD 解析 SKILL.md 内容
func parseSkillMD(content, path string) (*RemoteSkill, error) {
	skill := &RemoteSkill{
		ID: extractSkillID(path),
	}

	// 解析 YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			var frontmatter struct {
				Name        string `yaml:"name"`
				Description string `yaml:"description"`
				Command     string `yaml:"command"`
				Category    string `yaml:"category"`
				Author      string `yaml:"author"`
				Version     string `yaml:"version"`
			}

			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				skill.Name = frontmatter.Name
				skill.Description = frontmatter.Description
				skill.Command = frontmatter.Command
				skill.Category = frontmatter.Category
				skill.Author = frontmatter.Author
				skill.Version = frontmatter.Version
			}
		}
	}

	// 如果没有从 frontmatter 获取到，尝试从内容解析
	if skill.Name == "" {
		skill.Name = extractSkillID(path)
	}

	if skill.Description == "" {
		// 尝试从第一段提取描述
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
				skill.Description = line
				break
			}
		}
	}

	if skill.Command == "" {
		skill.Command = "/" + skill.ID
	}

	if skill.Category == "" {
		skill.Category = "other"
	}

	return skill, nil
}

// extractSkillID 从路径提取 skill ID
func extractSkillID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// InstallSkill 安装远程 Skill
func (s *SkillStore) InstallSkill(ctx context.Context, sourceID, skillID string) (*Skill, error) {
	source, ok := s.sources[sourceID]
	if !ok {
		return nil, fmt.Errorf("source not found: %s", sourceID)
	}

	// 获取 SKILL.md 完整内容
	skillPath := fmt.Sprintf("%s/%s", source.Path, skillID)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/SKILL.md?ref=%s",
		source.Owner, source.Repo, skillPath, source.Branch)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	s.setGitHubHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch SKILL.md: %d (提示: 设置 GITHUB_TOKEN 环境变量可提高 API 请求限额)", resp.StatusCode)
	}

	var fileContent struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileContent); err != nil {
		return nil, err
	}

	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(fileContent.Content, "\n", ""))
	if err != nil {
		return nil, err
	}

	// 解析为 Skill
	skill, err := s.parseFullSkillMD(string(content), skillID, source)
	if err != nil {
		return nil, err
	}

	// 转换为 CreateSkillRequest
	createReq := &CreateSkillRequest{
		ID:          skill.ID,
		Name:        skill.Name,
		Description: skill.Description,
		Command:     skill.Command,
		Prompt:      skill.Prompt,
		Category:    skill.Category,
		Author:      skill.Author,
		Version:     skill.Version,
	}

	// 保存到本地
	created, err := s.manager.Create(createReq)
	if err != nil {
		// 如果已存在，尝试更新
		if err == ErrSkillAlreadyExists {
			updated, updateErr := s.manager.Update(skill.ID, &UpdateSkillRequest{
				Name:        &skill.Name,
				Description: &skill.Description,
				Prompt:      &skill.Prompt,
				Version:     &skill.Version,
			})
			if updateErr != nil {
				return nil, updateErr
			}
			return updated, nil
		}
		return nil, err
	}

	return created, nil
}

// parseFullSkillMD 解析完整的 SKILL.md 为 Skill
func (s *SkillStore) parseFullSkillMD(content, skillID string, source *SkillSource) (*Skill, error) {
	skill := &Skill{
		ID:        skillID,
		IsBuiltIn: false,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var bodyContent string

	// 解析 YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			var frontmatter struct {
				Name        string `yaml:"name"`
				Description string `yaml:"description"`
				Command     string `yaml:"command"`
				Category    string `yaml:"category"`
				Author      string `yaml:"author"`
				Version     string `yaml:"version"`
			}

			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				skill.Name = frontmatter.Name
				skill.Description = frontmatter.Description
				skill.Command = frontmatter.Command
				skill.Category = Category(frontmatter.Category)
				skill.Author = frontmatter.Author
				skill.Version = frontmatter.Version
			}

			bodyContent = strings.TrimSpace(parts[2])
		}
	} else {
		bodyContent = content
	}

	// 设置默认值
	if skill.Name == "" {
		skill.Name = skillID
	}
	if skill.Command == "" {
		skill.Command = "/" + skillID
	}
	if skill.Category == "" {
		skill.Category = CategoryOther
	}
	if skill.Author == "" {
		skill.Author = source.Name
	}

	// 使用 body 作为 prompt
	skill.Prompt = bodyContent

	return skill, nil
}

// UninstallSkill 卸载 Skill
func (s *SkillStore) UninstallSkill(skillID string) error {
	return s.manager.Delete(skillID)
}

// RefreshCache 刷新缓存
func (s *SkillStore) RefreshCache(ctx context.Context, sourceID string) error {
	s.cacheMu.Lock()
	delete(s.cache, sourceID)
	delete(s.cacheTime, sourceID)
	s.cacheMu.Unlock()

	_, err := s.FetchSkills(ctx, sourceID)
	return err
}

// GetRepoStars 获取仓库 stars（可选，用于排序）
func (s *SkillStore) GetRepoStars(ctx context.Context, owner, repo string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	s.setGitHubHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var repoInfo struct {
		StargazersCount int `json:"stargazers_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return 0, err
	}

	return repoInfo.StargazersCount, nil
}

// 正则表达式用于解析
var frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---`)
