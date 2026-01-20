package mcp

import (
	"errors"
	"strings"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// ServerType MCP Server 通信类型
type ServerType string

const (
	ServerTypeStdio ServerType = "stdio"
	ServerTypeSSE   ServerType = "sse"
	ServerTypeHTTP  ServerType = "http"
)

// Category MCP Server 类别
type Category string

const (
	CategoryFilesystem Category = "filesystem"
	CategoryDatabase   Category = "database"
	CategoryAPI        Category = "api"
	CategoryTool       Category = "tool"
	CategoryBrowser    Category = "browser"
	CategoryMemory     Category = "memory"
	CategoryOther      Category = "other"
)

// Server MCP Server 配置
type Server struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty"`

	// 类型信息
	Type     ServerType `json:"type"`
	Category Category   `json:"category"`
	Tags     []string   `json:"tags,omitempty"`

	// URL (for SSE/HTTP type)
	URL string `json:"url,omitempty"`

	// 状态
	IsBuiltIn bool `json:"is_built_in"`
	IsEnabled bool `json:"is_enabled"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate 验证 Server 配置
func (s *Server) Validate() error {
	if s.ID == "" {
		return errors.New("id is required")
	}
	if s.Name == "" {
		return errors.New("name is required")
	}

	// stdio 类型需要 command
	if s.Type == ServerTypeStdio || s.Type == "" {
		if s.Command == "" {
			return errors.New("command is required for stdio type")
		}
	}

	// sse/http 类型需要 url
	if s.Type == ServerTypeSSE || s.Type == ServerTypeHTTP {
		if s.URL == "" {
			return errors.New("url is required for sse/http type")
		}
	}

	// 验证 ID 格式
	if strings.ContainsAny(s.ID, " \t\n/\\") {
		return errors.New("id cannot contain whitespace or slashes")
	}

	return nil
}

// Clone 克隆 Server
func (s *Server) Clone() *Server {
	clone := *s
	clone.IsBuiltIn = false

	// 深拷贝 slice 和 map
	if s.Args != nil {
		clone.Args = make([]string, len(s.Args))
		copy(clone.Args, s.Args)
	}
	if s.Env != nil {
		clone.Env = make(map[string]string)
		for k, v := range s.Env {
			clone.Env[k] = v
		}
	}
	if s.Tags != nil {
		clone.Tags = make([]string, len(s.Tags))
		copy(clone.Tags, s.Tags)
	}

	return &clone
}

// ToMCPConfig 转换为 Claude Code MCP 配置格式
func (s *Server) ToMCPConfig() map[string]interface{} {
	config := map[string]interface{}{
		"command": s.Command,
	}

	if len(s.Args) > 0 {
		config["args"] = s.Args
	}
	if len(s.Env) > 0 {
		config["env"] = s.Env
	}

	return config
}

// CreateServerRequest 创建 Server 请求
type CreateServerRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty"`
	Type        ServerType        `json:"type"`
	Category    Category          `json:"category"`
	Tags        []string          `json:"tags,omitempty"`
	URL         string            `json:"url,omitempty"`
}

// UpdateServerRequest 更新 Server 请求
type UpdateServerRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Command     *string            `json:"command,omitempty"`
	Args        []string           `json:"args,omitempty"`
	Env         *map[string]string `json:"env,omitempty"`
	WorkDir     *string            `json:"work_dir,omitempty"`
	Type        *ServerType        `json:"type,omitempty"`
	Category    *Category          `json:"category,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	URL         *string            `json:"url,omitempty"`
	IsEnabled   *bool              `json:"is_enabled,omitempty"`
}

// 错误定义 - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrServerNotFound      = apperr.NotFound("mcp server")
	ErrServerAlreadyExists = apperr.AlreadyExists("mcp server")
	ErrServerIsBuiltIn     = apperr.Forbidden("cannot modify built-in mcp server")
)
