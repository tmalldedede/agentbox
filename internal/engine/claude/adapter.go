package claude

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tmalldedede/agentbox/internal/container"
	"github.com/tmalldedede/agentbox/internal/engine"
)

const (
	// AgentName Agent 名称
	AgentName = "claude-code"

	// DefaultImage 默认镜像 (v2: claude-code 2.1+, codex 0.87+)
	DefaultImage = "agentbox/agent:v2"
)

// Adapter Claude Code 适配器
type Adapter struct {
	image string
}

// New 创建 Claude Code 适配器
func New() *Adapter {
	return &Adapter{
		image: DefaultImage,
	}
}

// NewWithImage 使用自定义镜像创建适配器
func NewWithImage(image string) *Adapter {
	return &Adapter{
		image: image,
	}
}

// Name 返回 Agent 名称
func (a *Adapter) Name() string {
	return AgentName
}

// DisplayName 返回显示名称
func (a *Adapter) DisplayName() string {
	return "Claude Code"
}

// Description 返回描述
func (a *Adapter) Description() string {
	return "Anthropic's official CLI for Claude - an agentic coding assistant"
}

// Image 返回 Docker 镜像
func (a *Adapter) Image() string {
	return a.image
}

// PrepareContainer 准备容器配置
func (a *Adapter) PrepareContainer(session *engine.SessionInfo) *container.CreateConfig {
	env := make(map[string]string)
	for k, v := range session.Env {
		env[k] = v
	}

	return &container.CreateConfig{
		Name:  fmt.Sprintf("agentbox-%s-%s", AgentName, session.ID),
		Image: a.image,
		Cmd:   []string{"sleep", "infinity"},
		Env:   env,
		Mounts: []container.Mount{
			{
				Source:   session.Workspace,
				Target:   "/workspace",
				ReadOnly: false,
			},
		},
		Resources: container.ResourceConfig{
			CPULimit:    2.0,
			MemoryLimit: 4 * 1024 * 1024 * 1024, // 4GB
		},
		NetworkMode: "bridge",
		Labels: map[string]string{
			"agentbox.managed":    "true",
			"agentbox.agent":      AgentName,
			"agentbox.session.id": session.ID,
		},
	}
}

// PrepareContainerWithConfig 使用 AgentConfig 准备容器配置
func (a *Adapter) PrepareContainerWithConfig(session *engine.SessionInfo, cfg *engine.AgentConfig) *container.CreateConfig {
	config := a.PrepareContainer(session)

	// 应用资源限制
	if cfg.Resources.CPUs > 0 {
		config.Resources.CPULimit = cfg.Resources.CPUs
	}
	if cfg.Resources.MemoryMB > 0 {
		config.Resources.MemoryLimit = int64(cfg.Resources.MemoryMB) * 1024 * 1024
	}

	// 添加标签
	config.Labels["agentbox.agent.id"] = cfg.ID
	config.Labels["agentbox.agent.name"] = cfg.Name

	// 应用 Model 环境变量配置
	a.applyModelEnvVars(config.Env, &cfg.Model)

	return config
}

// applyModelEnvVars 根据 ModelConfig 设置环境变量
func (a *Adapter) applyModelEnvVars(env map[string]string, model *engine.ModelConfig) {
	if model.BaseURL != "" {
		env["ANTHROPIC_BASE_URL"] = model.BaseURL
	}
	if model.BearerToken != "" {
		env["ANTHROPIC_AUTH_TOKEN"] = model.BearerToken
	}
	if model.Name != "" {
		env["ANTHROPIC_MODEL"] = model.Name
	}
	if model.HaikuModel != "" {
		env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = model.HaikuModel
	}
	if model.SonnetModel != "" {
		env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = model.SonnetModel
	}
	if model.OpusModel != "" {
		env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = model.OpusModel
	}
	if model.TimeoutMS > 0 {
		env["API_TIMEOUT_MS"] = strconv.Itoa(model.TimeoutMS)
	}
	if model.MaxOutputTokens > 0 {
		env["CLAUDE_CODE_MAX_OUTPUT_TOKENS"] = strconv.Itoa(model.MaxOutputTokens)
	}
	if model.DisableTraffic {
		env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"
	}
}

// PrepareExec 准备执行命令
func (a *Adapter) PrepareExec(req *engine.ExecOptions) []string {
	args := []string{
		"claude",
		"-p", req.Prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
	}

	// 多轮对话：resume 使用上一轮的 session_id
	if req.ThreadID != "" {
		args = append(args, "--resume", req.ThreadID)
	}

	if req.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}
	for _, tool := range req.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}
	for _, tool := range req.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}

	return args
}

// PrepareExecWithConfig 使用 AgentConfig 准备执行命令
func (a *Adapter) PrepareExecWithConfig(req *engine.ExecOptions, cfg *engine.AgentConfig) []string {
	args := []string{"claude", "-p", req.Prompt}

	// 多轮对话：resume 使用上一轮的 session_id
	if req.ThreadID != "" {
		args = append(args, "--resume", req.ThreadID)
	}

	// 模型配置
	if cfg.Model.Name != "" {
		args = append(args, "--model", cfg.Model.Name)
	}

	// 权限模式
	if cfg.Permissions.SkipAll {
		args = append(args, "--dangerously-skip-permissions")
	} else if cfg.Permissions.Mode != "" {
		args = append(args, "--permission-mode", cfg.Permissions.Mode)
	}

	// 工具配置
	for _, tool := range cfg.Permissions.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}
	for _, tool := range cfg.Permissions.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}
	for _, tool := range cfg.Permissions.Tools {
		args = append(args, "--tools", tool)
	}

	// 目录访问
	for _, dir := range cfg.Permissions.AdditionalDirs {
		args = append(args, "--add-dir", dir)
	}

	// 系统提示词
	if cfg.SystemPrompt != "" {
		args = append(args, "--system-prompt", cfg.SystemPrompt)
	}
	if cfg.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", cfg.AppendSystemPrompt)
	}

	// MCP 服务器
	if len(cfg.MCPServers) > 0 {
		mcpConfig := make(map[string]interface{})
		mcpConfig["mcpServers"] = make(map[string]interface{})

		for _, server := range cfg.MCPServers {
			serverConfig := map[string]interface{}{
				"command": server.Command,
			}
			if len(server.Args) > 0 {
				serverConfig["args"] = server.Args
			}
			if len(server.Env) > 0 {
				serverConfig["env"] = server.Env
			}
			mcpConfig["mcpServers"].(map[string]interface{})[server.Name] = serverConfig
		}

		if configJSON, err := json.Marshal(mcpConfig); err == nil {
			args = append(args, "--mcp-config", string(configJSON))
		}
	}

	// 资源限制
	if cfg.Resources.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", cfg.Resources.MaxTurns))
	}
	if cfg.Resources.MaxBudgetUSD > 0 {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", cfg.Resources.MaxBudgetUSD))
	}

	// 自定义 Agent
	if len(cfg.CustomAgents) > 0 {
		args = append(args, "--agents", string(cfg.CustomAgents))
	}

	// 输出格式：默认使用 stream-json 以便解析 session_id 和结构化结果
	outputFormat := cfg.OutputFormat
	if outputFormat == "" {
		outputFormat = "stream-json"
	}
	args = append(args, "--output-format", outputFormat)

	// stream-json 在 print 模式下必须加 --verbose
	if outputFormat == "stream-json" || cfg.Debug.Verbose {
		args = append(args, "--verbose")
	}

	// 请求级别覆盖
	if req.MaxTurns > 0 {
		args = removeArg(args, "--max-turns")
		args = append(args, "--max-turns", fmt.Sprintf("%d", req.MaxTurns))
	}
	for _, tool := range req.AllowedTools {
		args = append(args, "--allowedTools", tool)
	}
	for _, tool := range req.DisallowedTools {
		args = append(args, "--disallowedTools", tool)
	}

	return args
}

// removeArg 从参数列表中移除指定参数及其值
func removeArg(args []string, argName string) []string {
	result := make([]string, 0, len(args))
	skip := false
	for _, arg := range args {
		if skip {
			skip = false
			continue
		}
		if arg == argName {
			skip = true
			continue
		}
		result = append(result, arg)
	}
	return result
}

// RequiredEnvVars 返回必需的环境变量
func (a *Adapter) RequiredEnvVars() []string {
	return []string{"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"}
}

// ValidateConfig 验证 AgentConfig 是否与此适配器兼容
func (a *Adapter) ValidateConfig(cfg *engine.AgentConfig) error {
	if cfg.Adapter != engine.AdapterClaudeCode {
		return fmt.Errorf("adapter %q is not compatible with Claude Code adapter", cfg.Adapter)
	}

	validModes := map[string]bool{
		"":                                    true,
		engine.PermissionModeAcceptEdits:      true,
		engine.PermissionModeBypassPermissions: true,
		engine.PermissionModeDefault:          true,
		engine.PermissionModeDelegate:         true,
		engine.PermissionModeDontAsk:          true,
		engine.PermissionModePlan:             true,
	}
	if !validModes[cfg.Permissions.Mode] {
		return fmt.Errorf("invalid permission mode %q for Claude Code", cfg.Permissions.Mode)
	}

	if cfg.Permissions.SandboxMode != "" {
		return fmt.Errorf("sandbox_mode is a Codex-specific option, not valid for Claude Code")
	}
	if cfg.Permissions.ApprovalPolicy != "" {
		return fmt.Errorf("approval_policy is a Codex-specific option, not valid for Claude Code")
	}
	if cfg.Permissions.FullAuto {
		return fmt.Errorf("full_auto is a Codex-specific option, not valid for Claude Code")
	}

	return nil
}

// ParseJSONLOutput 解析 Claude Code stream-json 输出
// 实现 engine.JSONOutputParser 接口
func (a *Adapter) ParseJSONLOutput(output string, includeEvents bool) (*engine.ExecResult, error) {
	var (
		message   string
		sessionID string
		events    []engine.ExecEvent
		usage     *engine.TokenUsage
		execErr   string
	)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// 跳过 Docker 流头和空行，找到 JSON 开始
		jsonStart := strings.Index(line, "{")
		if jsonStart < 0 {
			continue
		}
		line = line[jsonStart:]

		var msg claudeStreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if includeEvents {
			events = append(events, engine.ExecEvent{
				Type: msg.Type,
				Raw:  json.RawMessage(line),
			})
		}

		switch msg.Type {
		case "system":
			// 系统初始化消息包含 session_id
			if msg.SessionID != "" {
				sessionID = msg.SessionID
			}

		case "assistant":
			// 提取 assistant 的文本内容
			if msg.Message != nil {
				for _, block := range msg.Message.Content {
					if block.Type == "text" && block.Text != "" {
						if message != "" {
							message += "\n"
						}
						message += block.Text
					}
				}
			}

		case "result":
			// 最终结果：包含 session_id 和 usage
			if msg.SessionID != "" {
				sessionID = msg.SessionID
			}
			if msg.Subtype == "error" || msg.Subtype == "error_max_turns" {
				execErr = msg.ErrorMessage
			}
			if msg.Usage != nil {
				usage = &engine.TokenUsage{
					InputTokens:  msg.Usage.InputTokens,
					OutputTokens: msg.Usage.OutputTokens,
				}
			}
		}
	}

	// 如果没有解析到任何 JSON，回退到纯文本模式
	if message == "" && sessionID == "" && len(events) == 0 && execErr == "" {
		message = strings.TrimSpace(output)
	}

	result := &engine.ExecResult{
		Message:  message,
		ThreadID: sessionID, // Claude Code 的 session_id 对应 ThreadID 字段
		Usage:    usage,
		Error:    execErr,
	}

	if includeEvents {
		result.Events = events
	}

	return result, nil
}

// Claude Code stream-json 消息结构
type claudeStreamMessage struct {
	Type         string               `json:"type"`
	Subtype      string               `json:"subtype,omitempty"`
	SessionID    string               `json:"session_id,omitempty"`
	Message      *claudeMessage       `json:"message,omitempty"`
	Usage        *claudeUsage         `json:"usage,omitempty"`
	ErrorMessage string               `json:"error,omitempty"`
}

type claudeMessage struct {
	Role    string         `json:"role"`
	Content []claudeBlock  `json:"content"`
}

type claudeBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// SupportedFeatures 返回此适配器支持的功能列表
func (a *Adapter) SupportedFeatures() []string {
	return []string{
		"model",
		"permission_mode",
		"allowed_tools",
		"disallowed_tools",
		"tools",
		"additional_dirs",
		"system_prompt",
		"append_system_prompt",
		"mcp_servers",
		"max_turns",
		"max_budget_usd",
		"custom_agents",
		"output_format",
		"verbose",
		"custom_provider",
		"base_url",
		"auth_token",
	}
}

// init 自动注册到默认注册表
func init() {
	engine.Register(New())
}
