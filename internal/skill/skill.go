package skill

import (
	"errors"
	"strings"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// Category Skill 类别
type Category string

const (
	CategoryCoding   Category = "coding"
	CategoryReview   Category = "review"
	CategoryDocs     Category = "docs"
	CategorySecurity Category = "security"
	CategoryTesting  Category = "testing"
	CategoryOther    Category = "other"
)

// Skill 技能定义
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command"` // 触发命令，如 "/commit"

	// Skill 内容
	Prompt string      `json:"prompt"`          // 主提示词
	Files  []SkillFile `json:"files,omitempty"` // 附加文件（如 references）

	// 源目录 - 本地文件系统路径，用于复制 scripts/tools 等目录到容器
	// 如果设置，注入时会复制整个目录（排除 SKILL.md，因为会动态生成）
	SourceDir string `json:"source_dir,omitempty"`

	// 配置
	AllowedTools []string `json:"allowed_tools,omitempty"` // 允许使用的工具
	RequiredMCP  []string `json:"required_mcp,omitempty"`  // 依赖的 MCP Server

	// 元数据
	Category Category `json:"category"`
	Tags     []string `json:"tags,omitempty"`
	Author   string   `json:"author,omitempty"`
	Version  string   `json:"version,omitempty"`

	// 状态
	IsBuiltIn bool `json:"is_built_in"`
	IsEnabled bool `json:"is_enabled"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SkillFile Skill 附加文件
type SkillFile struct {
	Path    string `json:"path"`    // 相对路径，如 "references/examples.md"
	Content string `json:"content"` // 文件内容
}

// Validate 验证 Skill 配置
func (s *Skill) Validate() error {
	if s.ID == "" {
		return errors.New("id is required")
	}
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.Command == "" {
		return errors.New("command is required")
	}
	if s.Prompt == "" {
		return errors.New("prompt is required")
	}

	// 验证 ID 格式
	if strings.ContainsAny(s.ID, " \t\n/\\") {
		return errors.New("id cannot contain whitespace or slashes")
	}

	// 验证 Command 格式（应以 / 开头）
	if !strings.HasPrefix(s.Command, "/") {
		return errors.New("command must start with /")
	}

	return nil
}

// Clone 克隆 Skill
func (s *Skill) Clone() *Skill {
	clone := *s
	clone.IsBuiltIn = false

	// 深拷贝 slice
	if s.Files != nil {
		clone.Files = make([]SkillFile, len(s.Files))
		copy(clone.Files, s.Files)
	}
	if s.AllowedTools != nil {
		clone.AllowedTools = make([]string, len(s.AllowedTools))
		copy(clone.AllowedTools, s.AllowedTools)
	}
	if s.RequiredMCP != nil {
		clone.RequiredMCP = make([]string, len(s.RequiredMCP))
		copy(clone.RequiredMCP, s.RequiredMCP)
	}
	if s.Tags != nil {
		clone.Tags = make([]string, len(s.Tags))
		copy(clone.Tags, s.Tags)
	}

	return &clone
}

// ToSkillMD 生成 SKILL.md 格式内容
func (s *Skill) ToSkillMD() string {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(s.Name)
	sb.WriteString("\n\n")

	if s.Description != "" {
		sb.WriteString(s.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Command\n\n")
	sb.WriteString("`")
	sb.WriteString(s.Command)
	sb.WriteString("`\n\n")

	sb.WriteString("## Instructions\n\n")
	sb.WriteString(s.Prompt)
	sb.WriteString("\n")

	return sb.String()
}

// CreateSkillRequest 创建 Skill 请求
type CreateSkillRequest struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	Command      string      `json:"command"`
	Prompt       string      `json:"prompt"`
	Files        []SkillFile `json:"files,omitempty"`
	SourceDir    string      `json:"source_dir,omitempty"` // 本地源目录路径
	AllowedTools []string    `json:"allowed_tools,omitempty"`
	RequiredMCP  []string    `json:"required_mcp,omitempty"`
	Category     Category    `json:"category"`
	Tags         []string    `json:"tags,omitempty"`
	Author       string      `json:"author,omitempty"`
	Version      string      `json:"version,omitempty"`
}

// UpdateSkillRequest 更新 Skill 请求
type UpdateSkillRequest struct {
	Name         *string     `json:"name,omitempty"`
	Description  *string     `json:"description,omitempty"`
	Command      *string     `json:"command,omitempty"`
	Prompt       *string     `json:"prompt,omitempty"`
	Files        []SkillFile `json:"files,omitempty"`
	SourceDir    *string     `json:"source_dir,omitempty"` // 本地源目录路径
	AllowedTools []string    `json:"allowed_tools,omitempty"`
	RequiredMCP  []string    `json:"required_mcp,omitempty"`
	Category     *Category   `json:"category,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
	Author       *string     `json:"author,omitempty"`
	Version      *string     `json:"version,omitempty"`
	IsEnabled    *bool       `json:"is_enabled,omitempty"`
}

// 错误定义 - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrSkillNotFound      = apperr.NotFound("skill")
	ErrSkillAlreadyExists = apperr.AlreadyExists("skill")
	ErrSkillIsBuiltIn     = apperr.Forbidden("cannot modify built-in skill")
)
