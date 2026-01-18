# AgentBox 架构设计

## 项目定位

**AgentBox** - 开源的 AI Agent 容器化运行平台

为 AI Agent 提供安全、隔离、可管理的执行环境。支持 Claude Code、Codex 等主流 Agent，未来可扩展支持更多。

---

## 竞品分析

| 项目 | Stars | 语言 | 定位 | 优势 | 不足 |
|------|-------|------|------|------|------|
| **agent-infra/sandbox** | 2.1k | Python/TS/Go | All-in-one 沙箱 (Browser+Shell+VSCode+MCP) | 功能最全，SDK 完善 | 过于重量级，资源消耗大 |
| **anthropic/sandbox-runtime** | 2.5k | TypeScript | OS 级沙箱 (非 Docker) | Anthropic 官方，轻量 | 只支持 Claude Code，无管理界面 |
| **cloudflare/sandbox-sdk** | 806 | TypeScript | 边缘执行沙箱 | 企业级，边缘网络 | 依赖 Cloudflare，无法自托管 |
| **SWE-agent/SWE-ReX** | 409 | Python | 代码执行运行时 | 并行执行，多平台 | 专注 SWE-agent，不通用 |
| **textcortex/claude-code-sandbox** | 261 | TypeScript | Claude Code Docker 沙箱 | 有 Web UI，易用 | 只支持 Claude Code |
| **zzev/aibox** | 6 | JavaScript | 多 CLI 容器 | 支持 Claude/Codex/Gemini | 无 API，无管理界面 |
| **libops/cli-sandbox** | 3 | Shell | 简单多 CLI 容器 | 简单直接 | 功能单一 |

### 市场空白

1. **没有 Go 实现** - 所有竞品都是 Python/TypeScript/Shell
2. **没有统一管理后台** - 大多只是 CLI 工具或 SDK
3. **没有 HTTP API 优先** - 集成不便
4. **没有多 Agent 统一管理** - 各管各的

### AgentBox 差异化

| 差异点 | AgentBox | 竞品 |
|--------|----------|------|
| **语言** | Go (单二进制，高性能) | Python/TS (需运行时) |
| **部署** | 单文件部署 | 依赖 Node.js/Python 环境 |
| **API** | HTTP API 优先 | CLI/SDK 优先 |
| **管理** | 完整管理后台 | 大多无 UI |
| **多 Agent** | 统一架构，可扩展 | 单一 Agent 或简单堆叠 |
| **企业特性** | 会话管理、审计、RBAC | 基本无 |

---

## 核心价值

| 价值 | 说明 |
|------|------|
| **隔离执行** | 每个 Agent 会话运行在独立 Docker 容器中 |
| **统一管理** | 一个平台管理多种 Agent |
| **安全可控** | 资源限制、网络隔离、文件系统隔离 |
| **开源自托管** | 完全开源，可私有化部署 |

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         AgentBox Server                          │
├─────────────┬─────────────┬─────────────┬───────────────────────┤
│   HTTP API  │  WebSocket  │   Session   │       Config          │
│   (Gin)     │   (Logs)    │   Manager   │       Manager         │
├─────────────┴─────────────┴─────────────┴───────────────────────┤
│                       Agent Adapters                             │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐        │
│  │  Claude Code  │  │    Codex      │  │    Future     │        │
│  │   Adapter     │  │   Adapter     │  │   Adapters    │        │
│  └───────────────┘  └───────────────┘  └───────────────┘        │
├─────────────────────────────────────────────────────────────────┤
│                      Container Manager                           │
│           (Docker SDK: create/start/stop/delete)                 │
├─────────────────────────────────────────────────────────────────┤
│                        Docker Engine                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │
│  │Session 1│  │Session 2│  │Session 3│  │   ...   │            │
│  │Container│  │Container│  │Container│  │         │            │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 核心模块

### 1. 容器管理器 (ContainerManager)

职责：
- 创建/启动/停止/删除容器
- 资源限制 (CPU/内存/磁盘)
- 卷挂载 (工作目录持久化)
- 网络隔离

```go
type ContainerManager interface {
    Create(ctx context.Context, config ContainerConfig) (string, error)
    Start(ctx context.Context, containerID string) error
    Stop(ctx context.Context, containerID string) error
    Remove(ctx context.Context, containerID string) error
    Exec(ctx context.Context, containerID string, cmd []string) (io.Reader, error)
    Logs(ctx context.Context, containerID string) (io.ReadCloser, error)
}

type ContainerConfig struct {
    Image       string
    Cmd         []string
    Env         []string
    Mounts      []Mount
    Resources   ResourceConfig
    NetworkMode string
}

type ResourceConfig struct {
    CPULimit    float64  // CPU 核心数
    MemoryLimit int64    // 内存限制 (bytes)
    DiskLimit   int64    // 磁盘限制 (bytes)
}
```

### 2. 会话管理器 (SessionManager)

职责：
- 会话生命周期管理
- 工作空间持久化
- 执行历史记录

```go
type SessionManager interface {
    Create(ctx context.Context, req CreateSessionRequest) (*Session, error)
    Get(ctx context.Context, id string) (*Session, error)
    List(ctx context.Context, filter SessionFilter) ([]*Session, error)
    Delete(ctx context.Context, id string) error
    Execute(ctx context.Context, id string, prompt string) (*ExecutionResult, error)
}

type Session struct {
    ID          string
    Agent       string         // "claude-code", "codex", etc.
    Status      SessionStatus  // creating, running, stopped, error
    ContainerID string
    Workspace   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type SessionStatus string

const (
    SessionStatusCreating SessionStatus = "creating"
    SessionStatusRunning  SessionStatus = "running"
    SessionStatusStopped  SessionStatus = "stopped"
    SessionStatusError    SessionStatus = "error"
)
```

### 3. Agent 适配器 (AgentAdapter)

职责：
- 统一接口定义
- 各 Agent 的具体实现
- 可扩展架构

```go
type AgentAdapter interface {
    Name() string
    Image() string
    PrepareCommand(prompt string) []string
    PrepareEnv(session *Session) []string
    ParseOutput(output []byte) (*AgentOutput, error)
}

// Claude Code 适配器
type ClaudeCodeAdapter struct {
    apiKey string
}

func (a *ClaudeCodeAdapter) Name() string { return "claude-code" }
func (a *ClaudeCodeAdapter) Image() string { return "agentbox/claude-code:latest" }

// Codex 适配器
type CodexAdapter struct {
    apiKey string
}

func (a *CodexAdapter) Name() string { return "codex" }
func (a *CodexAdapter) Image() string { return "agentbox/codex:latest" }
```

### 4. HTTP API

RESTful API 设计：

```
# 会话管理
POST   /api/sessions              # 创建会话
GET    /api/sessions              # 列出会话
GET    /api/sessions/:id          # 获取会话详情
DELETE /api/sessions/:id          # 删除会话
POST   /api/sessions/:id/start    # 启动会话
POST   /api/sessions/:id/stop     # 停止会话

# 任务执行
POST   /api/sessions/:id/exec     # 执行 Agent 任务

# 日志
GET    /api/sessions/:id/logs     # 获取日志 (WebSocket 支持)

# 系统
GET    /api/health                # 健康检查
GET    /api/agents                # 列出支持的 Agent
```

---

## 数据模型

### Session

```go
type Session struct {
    ID          string         `json:"id"`
    Agent       string         `json:"agent"`
    Status      SessionStatus  `json:"status"`
    Workspace   string         `json:"workspace"`
    ContainerID string         `json:"container_id,omitempty"`
    Config      SessionConfig  `json:"config"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

type SessionConfig struct {
    CPULimit    float64           `json:"cpu_limit"`
    MemoryLimit int64             `json:"memory_limit"`
    Env         map[string]string `json:"env"`
    Mounts      []MountConfig     `json:"mounts"`
}
```

### Execution

```go
type Execution struct {
    ID        string          `json:"id"`
    SessionID string          `json:"session_id"`
    Prompt    string          `json:"prompt"`
    Status    ExecutionStatus `json:"status"`
    Output    string          `json:"output,omitempty"`
    Error     string          `json:"error,omitempty"`
    StartedAt time.Time       `json:"started_at"`
    EndedAt   *time.Time      `json:"ended_at,omitempty"`
}

type ExecutionStatus string

const (
    ExecutionStatusPending  ExecutionStatus = "pending"
    ExecutionStatusRunning  ExecutionStatus = "running"
    ExecutionStatusSuccess  ExecutionStatus = "success"
    ExecutionStatusFailed   ExecutionStatus = "failed"
)
```

---

## 安全设计

### 容器隔离

- 每个会话独立容器
- 默认网络隔离 (无外网访问，可配置)
- 只读文件系统 (工作目录除外)
- 非 root 用户运行

### 资源限制

```yaml
defaults:
  cpu_limit: 2.0        # 2 核
  memory_limit: 4GB     # 4GB 内存
  disk_limit: 10GB      # 10GB 磁盘
  timeout: 1h           # 1 小时超时
```

### 审计日志

记录所有操作：
- 会话创建/删除
- 任务执行
- 文件访问

---

## 扩展机制

### 添加新 Agent

1. 实现 `AgentAdapter` 接口
2. 创建 Dockerfile 构建镜像
3. 注册到 AdapterRegistry

```go
// 注册新 Agent
func init() {
    registry.Register("my-agent", &MyAgentAdapter{})
}
```

---

## 技术栈

| 组件 | 技术选型 |
|------|----------|
| 语言 | Go 1.21+ |
| HTTP 框架 | Gin |
| 容器管理 | Docker SDK (github.com/docker/docker) |
| 数据存储 | SQLite (开发) / PostgreSQL (生产) |
| 前端 | React + Vite + TailwindCSS |
| 配置 | Viper |
| 日志 | Zap |

---

## 项目结构

```
agentbox/
├── cmd/
│   └── agentbox/
│       └── main.go            # 入口
├── internal/
│   ├── container/             # 容器管理
│   │   ├── manager.go
│   │   └── docker.go
│   ├── session/               # 会话管理
│   │   ├── manager.go
│   │   └── store.go
│   ├── agent/                 # Agent 适配器
│   │   ├── adapter.go         # 接口定义
│   │   ├── registry.go        # 注册表
│   │   ├── claude/            # Claude Code
│   │   │   └── adapter.go
│   │   └── codex/             # Codex
│   │       └── adapter.go
│   ├── api/                   # HTTP API
│   │   ├── router.go
│   │   ├── handlers.go
│   │   └── middleware.go
│   └── config/                # 配置
│       └── config.go
├── web/                       # 前端
├── docker/                    # Dockerfile
│   ├── Dockerfile             # AgentBox 服务
│   ├── claude-code/           # Claude Code 镜像
│   └── codex/                 # Codex 镜像
├── docs/                      # 文档
├── README.md
├── DESIGN.md
├── LICENSE
├── Makefile
└── go.mod
```

---

## 实施路线图

### Phase 1: 基础框架 ✅
- [x] 项目初始化
- [x] 目录结构
- [x] 文档

### Phase 2: 核心实现
- [ ] 容器管理模块
- [ ] 会话管理模块
- [ ] Agent 适配器接口
- [ ] Claude Code 适配器
- [ ] HTTP API

### Phase 3: 完善
- [ ] Codex 适配器
- [ ] Web 管理后台
- [ ] Docker 镜像构建
- [ ] 测试用例
- [ ] CI/CD
