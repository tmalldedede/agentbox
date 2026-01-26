package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Parser SKILL.md 和 skill.yaml 解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{}
}

// SkillMDFrontmatter SKILL.md frontmatter 结构
type SkillMDFrontmatter struct {
	Name         string           `yaml:"name"`
	Description  string           `yaml:"description"`
	Command      string           `yaml:"command"`
	Category     string           `yaml:"category"`
	Tags         []string         `yaml:"tags"`
	Author       string           `yaml:"author"`
	Version      string           `yaml:"version"`
	AllowedTools []string         `yaml:"allowed_tools"`
	RequiredMCP  []string         `yaml:"required_mcp"`
	Invocation   *InvocationPolicy `yaml:"invocation"`
}

// SkillYAML skill.yaml 结构
type SkillYAML struct {
	Name         string         `yaml:"name"`
	Version      string         `yaml:"version"`
	Description  string         `yaml:"description"`
	Requirements *Requirements  `yaml:"requirements"`
	Runtime      *RuntimeConfig `yaml:"runtime"`
	Invocation   *InvocationPolicy `yaml:"invocation"`
}

// ParseSkillDir 解析 Skill 目录
// 返回完整的 Skill 对象
func (p *Parser) ParseSkillDir(path string, id string) (*Skill, error) {
	// 检查目录是否存在
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("skill directory not found: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	skill := &Skill{
		ID:        id,
		SourceDir: path,
		Source:    SourceWorkspace, // 默认来源
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. 解析 SKILL.md
	skillMDPath := filepath.Join(path, "SKILL.md")
	if _, err := os.Stat(skillMDPath); err == nil {
		if err := p.parseSkillMD(skillMDPath, skill); err != nil {
			return nil, fmt.Errorf("failed to parse SKILL.md: %w", err)
		}
	}

	// 2. 解析 skill.yaml（覆盖 SKILL.md 中的配置）
	skillYAMLPath := filepath.Join(path, "skill.yaml")
	if _, err := os.Stat(skillYAMLPath); err == nil {
		if err := p.parseSkillYAML(skillYAMLPath, skill); err != nil {
			log.Warn("failed to parse skill.yaml", "path", skillYAMLPath, "error", err)
		}
	}

	// 3. 扫描 references/ 目录
	refsDir := filepath.Join(path, "references")
	if _, err := os.Stat(refsDir); err == nil {
		files, err := p.scanReferences(refsDir, path)
		if err != nil {
			log.Warn("failed to scan references", "path", refsDir, "error", err)
		} else {
			skill.Files = append(skill.Files, files...)
		}
	}

	// 4. 设置默认值
	if skill.Name == "" {
		skill.Name = id
	}
	if skill.Command == "" {
		skill.Command = "/" + id
	}
	if skill.Category == "" {
		skill.Category = CategoryOther
	}

	// 验证
	if err := skill.Validate(); err != nil {
		return nil, fmt.Errorf("skill validation failed: %w", err)
	}

	return skill, nil
}

// parseSkillMD 解析 SKILL.md 文件
func (p *Parser) parseSkillMD(path string, skill *Skill) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := string(content)

	// 解析 YAML frontmatter
	var frontmatter SkillMDFrontmatter
	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 3 {
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				skill.Name = frontmatter.Name
				skill.Description = frontmatter.Description
				skill.Command = frontmatter.Command
				skill.Category = Category(frontmatter.Category)
				skill.Tags = frontmatter.Tags
				skill.Author = frontmatter.Author
				skill.Version = frontmatter.Version
				skill.AllowedTools = frontmatter.AllowedTools
				skill.RequiredMCP = frontmatter.RequiredMCP
				if frontmatter.Invocation != nil {
					skill.Invocation = *frontmatter.Invocation
				}
			}

			// Body 是 frontmatter 之后的内容
			skill.Prompt = strings.TrimSpace(parts[2])
		}
	} else {
		// 没有 frontmatter，整个文件作为 prompt
		skill.Prompt = text
	}

	// 如果没有 frontmatter，尝试从 Markdown 内容提取信息
	if skill.Name == "" {
		skill.Name = p.extractTitleFromMarkdown(text)
	}
	if skill.Description == "" {
		skill.Description = p.extractDescriptionFromMarkdown(text)
	}

	return nil
}

// parseSkillYAML 解析 skill.yaml 文件
func (p *Parser) parseSkillYAML(path string, skill *Skill) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var skillYAML SkillYAML
	if err := yaml.Unmarshal(content, &skillYAML); err != nil {
		return err
	}

	// skill.yaml 中的字段覆盖 SKILL.md
	if skillYAML.Name != "" {
		skill.Name = skillYAML.Name
	}
	if skillYAML.Version != "" {
		skill.Version = skillYAML.Version
	}
	if skillYAML.Description != "" {
		skill.Description = skillYAML.Description
	}
	if skillYAML.Requirements != nil {
		skill.Requirements = skillYAML.Requirements
	}
	if skillYAML.Runtime != nil {
		skill.Runtime = skillYAML.Runtime
	}
	if skillYAML.Invocation != nil {
		skill.Invocation = *skillYAML.Invocation
	}

	return nil
}

// scanReferences 扫描 references 目录
func (p *Parser) scanReferences(refsDir string, baseDir string) ([]SkillFile, error) {
	var files []SkillFile

	err := filepath.Walk(refsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 跳过隐藏文件
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		files = append(files, SkillFile{
			Path:    relPath,
			Content: string(content),
		})
		return nil
	})

	return files, err
}

// extractTitleFromMarkdown 从 Markdown 提取标题
func (p *Parser) extractTitleFromMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// extractDescriptionFromMarkdown 从 Markdown 提取描述（第一段非标题文本）
func (p *Parser) extractDescriptionFromMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 跳过 frontmatter
		if trimmed == "---" {
			inFrontmatter = !inFrontmatter
			continue
		}
		if inFrontmatter {
			continue
		}

		// 跳过空行和标题
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 返回第一段非空文本
		return trimmed
	}
	return ""
}

// ParseSkillMDContent 解析 SKILL.md 内容字符串
func (p *Parser) ParseSkillMDContent(content string, id string) (*Skill, error) {
	skill := &Skill{
		ID:        id,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 解析 YAML frontmatter
	var frontmatter SkillMDFrontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err == nil {
				skill.Name = frontmatter.Name
				skill.Description = frontmatter.Description
				skill.Command = frontmatter.Command
				skill.Category = Category(frontmatter.Category)
				skill.Tags = frontmatter.Tags
				skill.Author = frontmatter.Author
				skill.Version = frontmatter.Version
				skill.AllowedTools = frontmatter.AllowedTools
				skill.RequiredMCP = frontmatter.RequiredMCP
				if frontmatter.Invocation != nil {
					skill.Invocation = *frontmatter.Invocation
				}
			}
			skill.Prompt = strings.TrimSpace(parts[2])
		}
	} else {
		skill.Prompt = content
	}

	// 设置默认值
	if skill.Name == "" {
		skill.Name = id
	}
	if skill.Command == "" {
		skill.Command = "/" + id
	}
	if skill.Category == "" {
		skill.Category = CategoryOther
	}

	return skill, nil
}

// ScanSkillDirs 扫描目录下所有 Skill
func (p *Parser) ScanSkillDirs(baseDir string) ([]*Skill, error) {
	var skills []*Skill

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return skills, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 跳过隐藏目录
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		skillDir := filepath.Join(baseDir, entry.Name())
		skillID := entry.Name()

		// 检查是否有 SKILL.md
		skillMDPath := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
			continue
		}

		skill, err := p.ParseSkillDir(skillDir, skillID)
		if err != nil {
			log.Warn("failed to parse skill", "id", skillID, "error", err)
			continue
		}

		skills = append(skills, skill)
	}

	return skills, nil
}

// ValidateSkillDir 验证 Skill 目录结构
func (p *Parser) ValidateSkillDir(path string) error {
	// 必须存在 SKILL.md
	skillMDPath := filepath.Join(path, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		return fmt.Errorf("SKILL.md not found in %s", path)
	}

	// 验证 SKILL.md 不为空
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return fmt.Errorf("failed to read SKILL.md: %w", err)
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("SKILL.md is empty")
	}

	return nil
}

// ExtractReferencePaths 从 SKILL.md 内容中提取引用的文件路径
// 支持 Markdown 链接格式：[text](path) 和 直接路径引用
func (p *Parser) ExtractReferencePaths(content string) []string {
	var paths []string
	seen := make(map[string]bool)

	// 匹配 Markdown 链接
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			path := match[2]
			// 只处理相对路径
			if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
				if !seen[path] {
					paths = append(paths, path)
					seen[path] = true
				}
			}
		}
	}

	// 匹配 references/ 目录下的路径
	refsRegex := regexp.MustCompile(`references/[^\s\)\"\']+`)
	refsMatches := refsRegex.FindAllString(content, -1)
	for _, path := range refsMatches {
		if !seen[path] {
			paths = append(paths, path)
			seen[path] = true
		}
	}

	return paths
}
