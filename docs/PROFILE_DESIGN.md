# AgentBox Profile 系统设计

## 概述

Profile 是 AgentBox 的核心差异化特性，它在 Adapter 之上提供一个用户可配置、可保存、可复用的抽象层。

```
┌─────────────────────────────────────────────────────────┐
│                      Profile Layer                       │
│  (用户可配置: 模型 + MCP + Skills + 权限 + 资源限制)      │
├─────────────────────────────────────────────────────────┤
│                      Adapter Layer                       │
│  (Claude Code Adapter / Codex Adapter / ...)            │
├─────────────────────────────────────────────────────────┤
│                    Container Runtime                     │
│  (Docker 容器隔离执行)                                   │
└─────────────────────────────────────────────────────────┘
```

---

## CLI 参数完整映射

### Claude Code CLI 参数

| 参数 | 类型 | 说明 | Profile 字段 | 可预配置 |
|------|------|------|--------------|----------|
| `--model <model>` | string | 模型选择 (sonnet, opus, haiku) | `model.name` | ✅ |
| `--mcp-config <configs...>` | []string | MCP 服务器配置 JSON 文件 | `mcp_servers` | ✅ |
| `--allowedTools <tools...>` | []string | 允许的工具列表 | `permissions.allowed_tools` | ✅ |
| `--disallowedTools <tools...>` | []string | 禁止的工具列表 | `permissions.disallowed_tools` | ✅ |
| `--system-prompt <prompt>` | string | 自定义系统提示词 | `system_prompt` | ✅ |
| `--append-system-prompt <prompt>` | string | 追加系统提示词 | `append_system_prompt` | ✅ |
| `--permission-mode <mode>` | enum | 权限模式 | `permissions.mode` | ✅ |
| `--dangerously-skip-permissions` | bool | 跳过权限检查 | `permissions.skip_all` | ✅ |
| `--max-budget-usd <amount>` | float | 预算限制 | `resources.max_budget_usd` | ✅ |
| `--add-dir <directories...>` | []string | 额外目录访问 | `permissions.additional_dirs` | ✅ |
| `--tools <tools...>` | []string | 指定可用工具 | `permissions.tools` | ✅ |
| `--agents <json>` | json | 自定义 Agent 定义 | `custom_agents` | ✅ |
| `--max-turns <n>` | int | 最大对话轮次 | `resources.max_turns` | ✅ |
| `--output-format <format>` | enum | 输出格式 (text/json/stream-json) | `output_format` | ✅ |
| `-p, --print` | bool | 仅打印不执行 | - | ❌ 运行时 |
| `-c, --continue` | bool | 继续上次会话 | - | ❌ 运行时 |
| `--resume <session>` | string | 恢复指定会话 | - | ❌ 运行时 |
| `--verbose` | bool | 详细日志 | `debug.verbose` | ✅ |

**permission-mode 可选值:**
- `acceptEdits` - 自动接受编辑
- `bypassPermissions` - 绕过权限检查
- `default` - 默认交互式
- `delegate` - 委托模式
- `dontAsk` - 不询问直接执行
- `plan` - 计划模式

### Codex CLI 参数

| 参数 | 类型 | 说明 | Profile 字段 | 可预配置 |
|------|------|------|--------------|----------|
| `-m, --model <MODEL>` | string | 模型选择 | `model.name` | ✅ |
| `-s, --sandbox <MODE>` | enum | 沙箱模式 | `permissions.sandbox_mode` | ✅ |
| `-a, --ask-for-approval <POLICY>` | enum | 审批策略 | `permissions.approval_policy` | ✅ |
| `--full-auto` | bool | 全自动模式 | `permissions.full_auto` | ✅ |
| `-p, --profile <PROFILE>` | string | 配置 Profile 名称 | - | ❌ 我们的 Profile |
| `-c, --config <key=value>` | []string | 配置覆盖 | `config_overrides` | ✅ |
| `--add-dir <DIR>` | []string | 额外可写目录 | `permissions.additional_dirs` | ✅ |
| `--search` | bool | 启用网页搜索 | `features.web_search` | ✅ |
| `--output-schema <FILE>` | string | 结构化输出 Schema | `output_schema` | ✅ |
| `--base-instructions <TEXT>` | string | 基础指令 | `base_instructions` | ✅ |
| `--developer-instructions <TEXT>` | string | 开发者指令 | `developer_instructions` | ✅ |
| `--compact-prompt <TEXT>` | string | 压缩提示词 | `compact_prompt` | ✅ |
| `--reasoning-effort <LEVEL>` | enum | 推理强度 (low/medium/high) | `model.reasoning_effort` | ✅ |
| `--timeout` | duration | 超时时间 | `resources.timeout` | ✅ |
| `--max-tokens` | int | 最大 token 数 | `resources.max_tokens` | ✅ |

**sandbox 可选值:**
- `read-only` - 只读模式
- `workspace-write` - 工作空间可写
- `danger-full-access` - 完全访问

**approval_policy 可选值:**
- `untrusted` - 不信任，每次询问
- `on-failure` - 失败时询问
- `on-request` - 请求时询问
- `never` - 从不询问

---

## Profile 数据模型

```go
// Profile 是用户可配置的 Agent 运行模板
type Profile struct {
    // 基础信息
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Icon        string    `json:"icon"`
    Tags        []string  `json:"tags"`

    // 适配器选择
    Adapter     string    `json:"adapter"`  // "claude-code" | "codex"

    // 继承
    Extends     string    `json:"extends,omitempty"`  // 继承自哪个 Profile

    // 模型配置
    Model       ModelConfig       `json:"model"`

    // MCP 服务器
    MCPServers  []MCPServerConfig `json:"mcp_servers,omitempty"`

    // 权限配置
    Permissions PermissionConfig  `json:"permissions"`

    // 资源限制
    Resources   ResourceConfig    `json:"resources"`

    // 系统提示词
    SystemPrompt       string `json:"system_prompt,omitempty"`
    AppendSystemPrompt string `json:"append_system_prompt,omitempty"`
    BaseInstructions   string `json:"base_instructions,omitempty"`
    DeveloperInstructions string `json:"developer_instructions,omitempty"`

    // 功能开关
    Features    FeatureConfig     `json:"features"`

    // 自定义 Agent (Claude Code)
    CustomAgents json.RawMessage  `json:"custom_agents,omitempty"`

    // 配置覆盖 (Codex)
    ConfigOverrides map[string]string `json:"config_overrides,omitempty"`

    // 输出配置
    OutputFormat string `json:"output_format,omitempty"`  // text/json/stream-json
    OutputSchema string `json:"output_schema,omitempty"`  // JSON Schema 文件路径

    // 调试
    Debug       DebugConfig       `json:"debug,omitempty"`

    // 元数据
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CreatedBy   string    `json:"created_by,omitempty"`
    IsBuiltIn   bool      `json:"is_built_in"`
    IsPublic    bool      `json:"is_public"`
}

// ModelConfig 模型配置
type ModelConfig struct {
    Name            string `json:"name"`             // e.g., "sonnet", "opus", "o3"
    Provider        string `json:"provider,omitempty"` // e.g., "anthropic", "openai"
    ReasoningEffort string `json:"reasoning_effort,omitempty"` // low/medium/high
}

// MCPServerConfig MCP 服务器配置
type MCPServerConfig struct {
    Name        string            `json:"name"`
    Command     string            `json:"command"`
    Args        []string          `json:"args,omitempty"`
    Env         map[string]string `json:"env,omitempty"`
    Description string            `json:"description,omitempty"`
}

// PermissionConfig 权限配置
type PermissionConfig struct {
    // Claude Code 特有
    Mode             string   `json:"mode,omitempty"`  // permission-mode
    AllowedTools     []string `json:"allowed_tools,omitempty"`
    DisallowedTools  []string `json:"disallowed_tools,omitempty"`
    Tools            []string `json:"tools,omitempty"`
    SkipAll          bool     `json:"skip_all,omitempty"`  // dangerously-skip-permissions

    // Codex 特有
    SandboxMode      string   `json:"sandbox_mode,omitempty"`  // read-only/workspace-write/danger-full-access
    ApprovalPolicy   string   `json:"approval_policy,omitempty"` // untrusted/on-failure/on-request/never
    FullAuto         bool     `json:"full_auto,omitempty"`

    // 通用
    AdditionalDirs   []string `json:"additional_dirs,omitempty"`
}

// ResourceConfig 资源限制
type ResourceConfig struct {
    MaxBudgetUSD float64       `json:"max_budget_usd,omitempty"`
    MaxTurns     int           `json:"max_turns,omitempty"`
    MaxTokens    int           `json:"max_tokens,omitempty"`
    Timeout      time.Duration `json:"timeout,omitempty"`

    // 容器资源
    CPUs         float64 `json:"cpus,omitempty"`
    MemoryMB     int     `json:"memory_mb,omitempty"`
    DiskGB       int     `json:"disk_gb,omitempty"`
}

// FeatureConfig 功能开关
type FeatureConfig struct {
    WebSearch bool `json:"web_search,omitempty"`
}

// DebugConfig 调试配置
type DebugConfig struct {
    Verbose bool `json:"verbose,omitempty"`
}
```

---

## Profile 继承机制

Profile 支持继承，子 Profile 可以覆盖父 Profile 的配置：

```yaml
# 基础 Profile: security-research
id: security-research
name: 安全研究
adapter: claude-code
model:
  name: opus
permissions:
  mode: dontAsk
  allowed_tools:
    - Bash
    - Read
    - Write
    - Grep
mcp_servers:
  - name: cybersec-cloud
    command: npx
    args: ["-y", "@anthropic/mcp-cybersec"]

---
# 子 Profile: 继承并扩展
id: malware-analysis
name: 恶意软件分析
extends: security-research
description: 专用于恶意软件分析的配置
permissions:
  sandbox_mode: read-only  # 覆盖: 更严格的沙箱
  additional_dirs:
    - /samples  # 新增: 样本目录
mcp_servers:
  - name: virustotal
    command: npx
    args: ["-y", "@mcp/virustotal"]
```

继承规则：
1. 子 Profile 的字段完全覆盖父 Profile 同名字段
2. 数组类型字段可选择覆盖或追加（通过 `+` 前缀）
3. 嵌套对象进行深度合并

---

## 内置 Profiles

### 1. 通用开发 (claude-code-dev)

```yaml
id: claude-code-dev
name: Claude Code 开发
adapter: claude-code
model:
  name: sonnet
permissions:
  mode: default
resources:
  max_budget_usd: 10
  cpus: 4
  memory_mb: 4096
```

### 2. 全自动编码 (codex-full-auto)

```yaml
id: codex-full-auto
name: Codex 全自动
adapter: codex
model:
  name: o3
permissions:
  full_auto: true
  sandbox_mode: workspace-write
resources:
  cpus: 2
  memory_mb: 2048
```

### 3. 安全研究 (security-research)

```yaml
id: security-research
name: 安全研究
adapter: claude-code
model:
  name: opus
permissions:
  mode: dontAsk
  allowed_tools:
    - Bash
    - Read
    - Grep
    - WebFetch
    - WebSearch
mcp_servers:
  - name: cybersec-cloud
    command: cybersec-mcp
features:
  web_search: true
```

### 4. 数据分析 (data-analysis)

```yaml
id: data-analysis
name: 数据分析
adapter: claude-code
model:
  name: sonnet
permissions:
  allowed_tools:
    - Read
    - Write
    - Bash
    - NotebookEdit
mcp_servers:
  - name: jupyter
    command: jupyter-mcp
resources:
  memory_mb: 8192
```

---

## API 设计

### Profiles CRUD

```
GET    /api/profiles              # 列出所有 Profiles
POST   /api/profiles              # 创建 Profile
GET    /api/profiles/:id          # 获取 Profile 详情
PUT    /api/profiles/:id          # 更新 Profile
DELETE /api/profiles/:id          # 删除 Profile
POST   /api/profiles/:id/clone    # 克隆 Profile
```

### 使用 Profile 创建会话

```
POST /api/sessions
{
  "profile_id": "security-research",
  "workspace": "/path/to/project",
  "env": {
    "CUSTOM_VAR": "value"
  }
}
```

### Profile 导入导出

```
GET  /api/profiles/:id/export     # 导出为 YAML/JSON
POST /api/profiles/import         # 导入 Profile
```

---

## 目录结构

```
internal/
├── profile/
│   ├── profile.go           # Profile 数据模型
│   ├── manager.go           # Profile 管理器
│   ├── resolver.go          # 继承解析器
│   ├── validator.go         # Profile 验证
│   └── builtin/             # 内置 Profiles
│       ├── claude_dev.yaml
│       ├── codex_auto.yaml
│       ├── security.yaml
│       └── data_analysis.yaml
├── agent/
│   ├── adapter.go           # Adapter 接口
│   ├── claude/
│   │   └── adapter.go       # Claude Code Adapter
│   └── codex/
│       └── adapter.go       # Codex Adapter
```

---

## Adapter 接口扩展

```go
// Adapter 接口需要支持 Profile
type Adapter interface {
    // 现有方法
    Name() string
    Image() string
    ValidateEnv(env map[string]string) error
    BuildExecCommand(prompt string) []string

    // 新增: 从 Profile 构建命令
    BuildExecCommandFromProfile(prompt string, profile *Profile) []string

    // 新增: 验证 Profile 兼容性
    ValidateProfile(profile *Profile) error

    // 新增: 获取适配器支持的功能
    SupportedFeatures() []string
}
```

---

## 实现优先级

### Phase 1: 核心 Profile 系统
1. Profile 数据模型
2. Profile 存储 (SQLite/文件)
3. Profile CRUD API
4. 基础内置 Profiles

### Phase 2: 适配器集成
1. Claude Code Adapter 支持 Profile
2. Codex Adapter 支持 Profile
3. 会话创建使用 Profile

### Phase 3: 高级特性
1. Profile 继承
2. Profile 导入导出
3. Profile 市场/分享

---

## 竞争优势

| 特性 | AgentBox | Claudebox | 其他竞品 |
|------|----------|-----------|----------|
| 用户可配置 Profile | ✅ | ❌ | ❌ |
| Profile 继承 | ✅ | ❌ | ❌ |
| 多 Agent 统一管理 | ✅ | ❌ | 部分 |
| MCP 预配置 | ✅ | ❌ | ❌ |
| 权限预设 | ✅ | ❌ | ❌ |
| Profile 分享/市场 | ✅ (计划) | ❌ | ❌ |
