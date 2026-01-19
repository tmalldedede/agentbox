# AgentBox 管理系统架构

## 概述

AgentBox 管理系统提供完整的 AI Agent 运行平台管理能力，包括资源配置、会话管理、凭证管理等核心功能。

---

## 模块架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        AgentBox 管理系统                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Session    │  │   Profile    │  │  Credential  │          │
│  │   会话管理    │  │  配置模板    │  │   凭证管理   │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                  │                  │
│         │    ┌────────────┴────────────┐    │                  │
│         │    │                         │    │                  │
│         ▼    ▼                         ▼    ▼                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  MCP Server  │  │    Skills    │  │   Adapter    │          │
│  │   MCP 服务   │  │   技能管理   │  │   适配器     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│                        基础设施层                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │    Image     │  │  Container   │  │   Storage    │          │
│  │   镜像管理   │  │   容器运行   │  │   数据存储   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 一、MCP Server 管理

MCP (Model Context Protocol) Server 为 Agent 提供扩展能力，如文件系统、数据库、API 调用等。

### 数据模型

```go
type MCPServer struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Command     string            `json:"command"`           // 启动命令
    Args        []string          `json:"args,omitempty"`    // 命令参数
    Env         map[string]string `json:"env,omitempty"`     // 环境变量
    WorkDir     string            `json:"work_dir,omitempty"`

    // 元数据
    Type        string            `json:"type"`              // stdio | sse | http
    Category    string            `json:"category"`          // filesystem | database | api | tool
    Tags        []string          `json:"tags,omitempty"`

    // 状态
    IsBuiltIn   bool              `json:"is_built_in"`
    IsEnabled   bool              `json:"is_enabled"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/mcp-servers | 列出所有 MCP Server |
| POST | /api/v1/mcp-servers | 创建 MCP Server |
| GET | /api/v1/mcp-servers/:id | 获取详情 |
| PUT | /api/v1/mcp-servers/:id | 更新配置 |
| DELETE | /api/v1/mcp-servers/:id | 删除 |
| POST | /api/v1/mcp-servers/:id/test | 测试连接 |

### 内置 MCP Servers

| ID | 名称 | 说明 |
|----|------|------|
| filesystem | Filesystem | 文件系统访问 |
| fetch | Fetch | HTTP 请求 |
| memory | Memory | 知识图谱记忆 |
| puppeteer | Puppeteer | 浏览器自动化 |
| postgres | PostgreSQL | 数据库操作 |

---

## 二、Skills 管理

Skills 是可复用的任务模板，定义 Agent 如何执行特定类型的任务。

### 数据模型

```go
type Skill struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Command     string            `json:"command"`           // 触发命令，如 "/commit"

    // Skill 内容
    Prompt      string            `json:"prompt"`            // 主提示词
    Files       []SkillFile       `json:"files,omitempty"`   // 附加文件

    // 配置
    AllowedTools   []string       `json:"allowed_tools,omitempty"`
    RequiredMCP    []string       `json:"required_mcp,omitempty"`  // 依赖的 MCP Server

    // 元数据
    Category    string            `json:"category"`          // coding | review | docs | security
    Tags        []string          `json:"tags,omitempty"`
    Author      string            `json:"author,omitempty"`
    Version     string            `json:"version,omitempty"`

    // 状态
    IsBuiltIn   bool              `json:"is_built_in"`
    IsEnabled   bool              `json:"is_enabled"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type SkillFile struct {
    Path    string `json:"path"`     // 相对路径
    Content string `json:"content"`  // 文件内容
}
```

### API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/skills | 列出所有 Skills |
| POST | /api/v1/skills | 创建 Skill |
| GET | /api/v1/skills/:id | 获取详情 |
| PUT | /api/v1/skills/:id | 更新 |
| DELETE | /api/v1/skills/:id | 删除 |
| POST | /api/v1/skills/import | 从文件导入 |
| GET | /api/v1/skills/:id/export | 导出为文件 |

### 内置 Skills

| ID | 命令 | 说明 |
|----|------|------|
| commit | /commit | 智能生成 commit message |
| review-pr | /review-pr | PR 代码审查 |
| explain | /explain | 解释代码 |
| refactor | /refactor | 代码重构建议 |
| test | /test | 生成测试用例 |
| docs | /docs | 生成文档 |

---

## 三、Credentials 凭证管理

安全存储和管理 API Keys、Tokens 等敏感凭证。

### 数据模型

```go
type Credential struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Type        string            `json:"type"`              // api_key | token | oauth
    Provider    string            `json:"provider"`          // anthropic | openai | github

    // 凭证值（加密存储，API 返回时掩码）
    Value       string            `json:"value,omitempty"`
    ValueMasked string            `json:"value_masked,omitempty"`

    // 作用域
    Scope       string            `json:"scope"`             // global | profile | session
    ProfileID   string            `json:"profile_id,omitempty"`

    // 状态
    IsValid     bool              `json:"is_valid"`
    LastUsedAt  *time.Time        `json:"last_used_at,omitempty"`
    ExpiresAt   *time.Time        `json:"expires_at,omitempty"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/credentials | 列出凭证（值掩码） |
| POST | /api/v1/credentials | 创建凭证 |
| GET | /api/v1/credentials/:id | 获取详情 |
| PUT | /api/v1/credentials/:id | 更新 |
| DELETE | /api/v1/credentials/:id | 删除 |
| POST | /api/v1/credentials/:id/verify | 验证有效性 |

### 凭证类型

| Provider | Type | 环境变量 |
|----------|------|----------|
| anthropic | api_key | ANTHROPIC_API_KEY |
| openai | api_key | OPENAI_API_KEY |
| github | token | GITHUB_TOKEN |

---

## 四、Profile 管理（完善）

Profile 是预配置模板，组合 Adapter、Model、MCP Servers、Skills、Permissions、Resources。

### 更新后的数据模型

```go
type Profile struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Icon        string            `json:"icon,omitempty"`
    Tags        []string          `json:"tags,omitempty"`

    // 适配器
    Adapter     string            `json:"adapter"`           // claude-code | codex
    Extends     string            `json:"extends,omitempty"` // 继承的 Profile ID

    // 模型配置
    Model       ModelConfig       `json:"model"`

    // 关联资源（引用 ID）
    MCPServerIDs  []string        `json:"mcp_server_ids,omitempty"`
    SkillIDs      []string        `json:"skill_ids,omitempty"`
    CredentialID  string          `json:"credential_id,omitempty"`

    // 内联配置（向后兼容）
    MCPServers    []MCPServerConfig `json:"mcp_servers,omitempty"`

    // 权限与资源
    Permissions   PermissionConfig  `json:"permissions"`
    Resources     ResourceConfig    `json:"resources"`

    // 提示词
    SystemPrompt       string      `json:"system_prompt,omitempty"`
    AppendSystemPrompt string      `json:"append_system_prompt,omitempty"`

    // 元数据
    IsBuiltIn   bool              `json:"is_built_in"`
    IsPublic    bool              `json:"is_public"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### API 补充

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/profiles/:id/resolved | 获取完全解析后的配置（含继承） |
| POST | /api/v1/profiles/:id/clone | 克隆 Profile |
| POST | /api/v1/profiles/import | 从 YAML/JSON 导入 |
| GET | /api/v1/profiles/:id/export | 导出为 YAML/JSON |

---

## 五、Session 管理（完善）

Session 是 Agent 的运行实例，关联一个 Profile。

### 更新后的数据模型

```go
type Session struct {
    ID          string            `json:"id"`
    Name        string            `json:"name,omitempty"`

    // 关联
    ProfileID   string            `json:"profile_id"`
    Profile     *Profile          `json:"profile,omitempty"`  // 展开的 Profile

    // 工作区
    Workspace   string            `json:"workspace"`

    // 容器信息
    ContainerID string            `json:"container_id,omitempty"`
    Status      string            `json:"status"`            // creating | running | stopped | error

    // 容器状态（实时）
    ContainerStatus *ContainerStatus `json:"container_status,omitempty"`

    // 凭证覆盖
    CredentialID string           `json:"credential_id,omitempty"`

    // 执行统计
    ExecCount   int               `json:"exec_count"`
    LastExecAt  *time.Time        `json:"last_exec_at,omitempty"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type ContainerStatus struct {
    State       string            `json:"state"`             // running | exited | paused
    CPUPercent  float64           `json:"cpu_percent"`
    MemoryUsage int64             `json:"memory_usage"`
    MemoryLimit int64             `json:"memory_limit"`
    NetworkRx   int64             `json:"network_rx"`
    NetworkTx   int64             `json:"network_tx"`
}
```

### API 补充

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/sessions | 创建会话（需 profile_id） |
| GET | /api/v1/sessions/:id/stats | 获取容器资源使用 |
| POST | /api/v1/sessions/:id/restart | 重启容器 |
| GET | /api/v1/sessions/:id/logs | 获取容器日志 |
| WS | /api/v1/sessions/:id/logs/stream | 流式日志 |

---

## 六、镜像管理

管理 Agent 运行所需的 Docker 镜像。

### 数据模型

```go
type Image struct {
    ID          string            `json:"id"`
    Repository  string            `json:"repository"`
    Tag         string            `json:"tag"`
    Digest      string            `json:"digest,omitempty"`
    Size        int64             `json:"size"`

    // 关联
    Adapter     string            `json:"adapter"`           // claude-code | codex

    // 状态
    Status      string            `json:"status"`            // available | pulling | error
    PullProgress float64          `json:"pull_progress,omitempty"`

    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/images | 列出可用镜像 |
| POST | /api/v1/images/pull | 拉取镜像 |
| DELETE | /api/v1/images/:id | 删除镜像 |
| GET | /api/v1/images/pull/:id/progress | 获取拉取进度 |

---

## 七、系统维护

系统级别的维护操作。

### API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/system/health | 健康检查 |
| GET | /api/v1/system/stats | 系统统计 |
| POST | /api/v1/system/cleanup | 清理过期资源 |
| POST | /api/v1/system/cleanup/containers | 清理孤儿容器 |
| POST | /api/v1/system/cleanup/images | 清理无用镜像 |
| POST | /api/v1/system/cleanup/sessions | 清理过期会话 |

---

## 八、前端页面规划

### 导航结构

```
Dashboard (/)
├── Sessions (/sessions)
│   ├── Session List
│   └── Session Detail (/sessions/:id)
├── Profiles (/profiles)
│   ├── Profile List
│   └── Profile Detail (/profiles/:id)
├── MCP Servers (/mcp-servers)
│   ├── MCP Server List
│   └── MCP Server Detail (/mcp-servers/:id)
├── Skills (/skills)
│   ├── Skill List
│   └── Skill Detail (/skills/:id)
├── Credentials (/credentials)
│   └── Credential List & Management
├── Images (/images)
│   └── Image Management
└── Settings (/settings)
    ├── General
    └── System Maintenance
```

### 页面清单

| 页面 | 路径 | 状态 | 说明 |
|------|------|------|------|
| Dashboard | / | ✅ 有 | 总览 |
| Session List | /sessions | ✅ 在 Dashboard | 会话列表 |
| Session Detail | /sessions/:id | ✅ 有 | 会话详情 |
| Profile List | /profiles | ✅ 有 | Profile 列表 |
| Profile Detail | /profiles/:id | ❌ 无 | Profile 编辑 |
| MCP Server List | /mcp-servers | ❌ 无 | MCP 列表 |
| MCP Server Detail | /mcp-servers/:id | ❌ 无 | MCP 编辑 |
| Skill List | /skills | ❌ 无 | Skill 列表 |
| Skill Detail | /skills/:id | ❌ 无 | Skill 编辑 |
| Credentials | /credentials | ❌ 无 | 凭证管理 |
| Images | /images | ❌ 无 | 镜像管理 |
| Settings | /settings | ✅ 有 | 设置 |

---

## 九、实现优先级

### P0 - 核心管理闭环

1. **MCP Server 管理** - 后端 + 前端
2. **Skills 管理** - 后端 + 前端
3. **Credentials 管理** - 后端 + 前端
4. **Profile 详情页** - 关联 MCP/Skills/Credential
5. **Session 创建关联 Profile**

### P1 - 完善体验

6. **镜像管理**
7. **容器状态监控**
8. **系统维护**

### P2 - 增强功能

9. **导入/导出**
10. **WebSocket 实时日志**
11. **执行中断**

---

## 十、目录结构

```
agentbox/
├── internal/
│   ├── mcp/                    # MCP Server 管理
│   │   ├── server.go           # 数据模型
│   │   ├── manager.go          # 业务逻辑
│   │   └── builtin.go          # 内置 MCP
│   ├── skill/                  # Skills 管理
│   │   ├── skill.go
│   │   ├── manager.go
│   │   └── builtin.go
│   ├── credential/             # 凭证管理
│   │   ├── credential.go
│   │   ├── manager.go
│   │   └── crypto.go           # 加密工具
│   ├── profile/                # Profile (已有)
│   ├── session/                # Session (已有)
│   ├── image/                  # 镜像管理
│   │   ├── image.go
│   │   └── manager.go
│   └── api/
│       ├── mcp_handler.go
│       ├── skill_handler.go
│       ├── credential_handler.go
│       ├── image_handler.go
│       └── system_handler.go
└── web/src/
    ├── components/
    │   ├── MCPServerList.tsx
    │   ├── MCPServerDetail.tsx
    │   ├── SkillList.tsx
    │   ├── SkillDetail.tsx
    │   ├── CredentialList.tsx
    │   ├── ImageList.tsx
    │   └── ProfileDetail.tsx
    └── types/
        └── index.ts            # 更新类型定义
```

---

## 十一、数据存储

当前使用文件存储（JSON），后续可迁移到 SQLite/PostgreSQL。

```
data/
├── profiles/
│   ├── index.json
│   └── {id}.json
├── mcp-servers/
│   ├── index.json
│   └── {id}.json
├── skills/
│   ├── index.json
│   └── {id}.json
├── credentials/
│   └── credentials.json        # 加密存储
└── sessions/
    └── {id}.json
```
