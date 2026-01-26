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

// LoadLevel 加载级别（渐进式加载）
type LoadLevel string

const (
	LoadLevelMetadata LoadLevel = "metadata" // 仅 name/description/command
	LoadLevelBody     LoadLevel = "body"     // 包含 prompt
	LoadLevelFull     LoadLevel = "full"     // 包含 references 和所有文件
)

// SkillSource Skill 来源
type SkillSource string

const (
	SourceExtra     SkillSource = "extra"     // 用户手动添加
	SourceBundled   SkillSource = "bundled"   // 代码内置
	SourceManaged   SkillSource = "managed"   // 远程仓库安装
	SourceWorkspace SkillSource = "workspace" // 工作区
)

// Requirements 依赖要求
type Requirements struct {
	Bins    []string `yaml:"bins" json:"bins,omitempty"`       // 必需二进制（全部需要）
	AnyBins []string `yaml:"any_bins" json:"any_bins,omitempty"` // 任选其一二进制（满足一个即可）
	Env     []string `yaml:"env" json:"env,omitempty"`         // 必需环境变量
	Config  []string `yaml:"config" json:"config,omitempty"`   // 必需配置项（如 mcpServers.xxx）
	Pip     []string `yaml:"pip" json:"pip,omitempty"`         // Python 依赖
	Npm     []string `yaml:"npm" json:"npm,omitempty"`         // Node.js 依赖
	OS      []string `yaml:"os" json:"os,omitempty"`           // 支持的操作系统（darwin, linux, windows）
}

// InstallKind 安装方式
type InstallKind string

const (
	InstallKindBrew     InstallKind = "brew"     // Homebrew
	InstallKindNode     InstallKind = "node"     // npm/pnpm/yarn
	InstallKindGo       InstallKind = "go"       // go install
	InstallKindUV       InstallKind = "uv"       // uv pip install
	InstallKindPip      InstallKind = "pip"      // pip install
	InstallKindDownload InstallKind = "download" // 下载二进制
)

// InstallSpec 安装规范
type InstallSpec struct {
	ID      string      `yaml:"id" json:"id,omitempty"`           // 安装方式 ID
	Kind    InstallKind `yaml:"kind" json:"kind"`                 // 安装方式类型
	Label   string      `yaml:"label" json:"label,omitempty"`     // 显示标签
	Bins    []string    `yaml:"bins" json:"bins,omitempty"`       // 安装后提供的二进制
	Formula string      `yaml:"formula" json:"formula,omitempty"` // brew formula
	Package string      `yaml:"package" json:"package,omitempty"` // npm/pip/uv 包名
	Module  string      `yaml:"module" json:"module,omitempty"`   // go module 路径
	URL     string      `yaml:"url" json:"url,omitempty"`         // 下载 URL
	OS      []string    `yaml:"os" json:"os,omitempty"`           // 支持的操作系统
}

// HasRequirements 检查是否有依赖要求
func (r *Requirements) HasRequirements() bool {
	if r == nil {
		return false
	}
	return len(r.Bins) > 0 || len(r.AnyBins) > 0 || len(r.Env) > 0 ||
		len(r.Config) > 0 || len(r.Pip) > 0 || len(r.Npm) > 0 || len(r.OS) > 0
}

// InvocationPolicy 调用策略
type InvocationPolicy struct {
	UserInvocable bool     `yaml:"user_invocable" json:"user_invocable,omitempty"` // /command 调用
	AutoInvocable bool     `yaml:"auto_invocable" json:"auto_invocable,omitempty"` // Agent 自动触发
	HookInvocable []string `yaml:"hook_invocable" json:"hook_invocable,omitempty"` // 事件触发（如 pre-commit）
}

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	Python   string `yaml:"python" json:"python,omitempty"`     // Python 版本要求，如 ">=3.8"
	Node     string `yaml:"node" json:"node,omitempty"`         // Node.js 版本要求
	Memory   string `yaml:"memory" json:"memory,omitempty"`     // 内存要求，如 "256Mi"
	Timeout  int    `yaml:"timeout" json:"timeout,omitempty"`   // 执行超时（秒）
	Sandbox  string `yaml:"sandbox" json:"sandbox,omitempty"`   // 沙箱模式
	Network  bool   `yaml:"network" json:"network,omitempty"`   // 是否需要网络
	Image    string `yaml:"image" json:"image,omitempty"`       // 自定义镜像
}

// SkillConfig Skill 级别配置（用户可配置）
type SkillConfig struct {
	Enabled bool              `json:"enabled"`           // 是否启用
	APIKey  string            `json:"api_key,omitempty"` // API Key（如 VIRUSTOTAL_API_KEY）
	Env     map[string]string `json:"env,omitempty"`     // 环境变量覆盖
}

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

	// === Phase 1: 渐进式加载 ===
	LoadLevel  LoadLevel `json:"load_level,omitempty"`  // 当前加载级别
	BodyLoaded bool      `json:"body_loaded,omitempty"` // 是否已加载 body

	// === Phase 2: 来源与依赖 ===
	Source       SkillSource    `json:"source,omitempty"`        // 来源类型
	SourcePath   string         `json:"source_path,omitempty"`   // 来源路径（工作区路径或仓库路径）
	Requirements *Requirements  `json:"requirements,omitempty"`  // 依赖要求
	Runtime      *RuntimeConfig `json:"runtime,omitempty"`       // 运行时配置
	Install      []InstallSpec  `json:"install,omitempty"`       // 安装规范
	PrimaryEnv   string         `json:"primary_env,omitempty"`   // 主环境变量（用于 API Key）
	Always       bool           `json:"always,omitempty"`        // 始终包含（忽略依赖检查）
	Emoji        string         `json:"emoji,omitempty"`         // 显示图标
	Homepage     string         `json:"homepage,omitempty"`      // 主页链接

	// === Phase 3: 调用策略 ===
	Invocation InvocationPolicy `json:"invocation,omitempty"` // 调用策略
}

// SkillMetadata Skill 元数据（用于快速列表）
type SkillMetadata struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Command     string       `json:"command"`
	Category    Category     `json:"category"`
	Tags        []string     `json:"tags,omitempty"`
	Author      string       `json:"author,omitempty"`
	Version     string       `json:"version,omitempty"`
	Source      SkillSource  `json:"source,omitempty"`
	IsBuiltIn   bool         `json:"is_built_in"`
	IsEnabled   bool         `json:"is_enabled"`
	HasDeps     bool         `json:"has_deps,omitempty"`     // 是否有依赖要求
	DepsSatisfied *bool      `json:"deps_satisfied,omitempty"` // 依赖是否满足
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ToMetadata 转换为元数据
func (s *Skill) ToMetadata() *SkillMetadata {
	return &SkillMetadata{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Command:     s.Command,
		Category:    s.Category,
		Tags:        s.Tags,
		Author:      s.Author,
		Version:     s.Version,
		Source:      s.Source,
		IsBuiltIn:   s.IsBuiltIn,
		IsEnabled:   s.IsEnabled,
		HasDeps:     s.Requirements.HasRequirements(),
		UpdatedAt:   s.UpdatedAt,
	}
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
	clone.Source = SourceExtra // 克隆的 Skill 变为用户添加

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

	// 深拷贝 Requirements
	if s.Requirements != nil {
		clone.Requirements = &Requirements{}
		if s.Requirements.Bins != nil {
			clone.Requirements.Bins = make([]string, len(s.Requirements.Bins))
			copy(clone.Requirements.Bins, s.Requirements.Bins)
		}
		if s.Requirements.Env != nil {
			clone.Requirements.Env = make([]string, len(s.Requirements.Env))
			copy(clone.Requirements.Env, s.Requirements.Env)
		}
		if s.Requirements.Config != nil {
			clone.Requirements.Config = make([]string, len(s.Requirements.Config))
			copy(clone.Requirements.Config, s.Requirements.Config)
		}
		if s.Requirements.Pip != nil {
			clone.Requirements.Pip = make([]string, len(s.Requirements.Pip))
			copy(clone.Requirements.Pip, s.Requirements.Pip)
		}
		if s.Requirements.Npm != nil {
			clone.Requirements.Npm = make([]string, len(s.Requirements.Npm))
			copy(clone.Requirements.Npm, s.Requirements.Npm)
		}
	}

	// 深拷贝 Runtime
	if s.Runtime != nil {
		cloneRuntime := *s.Runtime
		clone.Runtime = &cloneRuntime
	}

	// 深拷贝 Invocation.HookInvocable
	if s.Invocation.HookInvocable != nil {
		clone.Invocation.HookInvocable = make([]string, len(s.Invocation.HookInvocable))
		copy(clone.Invocation.HookInvocable, s.Invocation.HookInvocable)
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

	// 新增字段
	Source       SkillSource      `json:"source,omitempty"`
	Requirements *Requirements    `json:"requirements,omitempty"`
	Runtime      *RuntimeConfig   `json:"runtime,omitempty"`
	Invocation   *InvocationPolicy `json:"invocation,omitempty"`
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

	// 新增字段
	Requirements *Requirements     `json:"requirements,omitempty"`
	Runtime      *RuntimeConfig    `json:"runtime,omitempty"`
	Invocation   *InvocationPolicy `json:"invocation,omitempty"`
}

// 错误定义 - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrSkillNotFound      = apperr.NotFound("skill")
	ErrSkillAlreadyExists = apperr.AlreadyExists("skill")
	ErrSkillIsBuiltIn     = apperr.Forbidden("cannot modify built-in skill")
)
