# AgentBox 产品愿景

## 一句话定位

> **AgentBox** — AI Agent 云交付平台
>
> 将 Claude Code、Codex、OpenCode 等专业 AI Agent，通过可配置的方式，以 Docker 容器稳定交付给用户。

---

## 我们解决什么问题

### 现状痛点

| 痛点 | 说明 |
|------|------|
| **部署复杂** | 每个 Agent 都有独立的安装、配置流程，学习成本高 |
| **环境依赖** | Agent 需要特定运行环境，本地配置繁琐易出错 |
| **资源隔离** | 多个 Agent 或任务共享环境，互相干扰，安全隐患 |
| **能力孤立** | 每个 Agent 能力固定，难以扩展或组合 |
| **交付困难** | 企业想把 Agent 能力提供给用户，缺乏标准化方案 |

### AgentBox 的解决方案

```
用户只需要：提交任务 + 选择 Profile → 获取结果

AgentBox 负责：Agent 管理 + 环境配置 + 容器运行 + 能力扩展
```

---

## 核心价值主张

### 1. Agent 无关（Agent Agnostic）

不绑定特定 Agent，支持多种专业 AI Agent：

| Agent | 说明 | 状态 |
|-------|------|------|
| Claude Code | Anthropic 官方 CLI Agent | ✅ 已支持 |
| Codex | OpenAI Codex CLI Agent | ✅ 已支持 |
| OpenCode | 开源代码 Agent | 🔜 计划中 |
| Aider | AI Pair Programming | 🔜 计划中 |
| 更多... | 可扩展接入 | - |

**价值**：今天用 Claude Code，明天换 Codex，无需改代码。

### 2. 可配置增强（Configurable Enhancement）

通过 Profile 组合 Agent 能力：

```
Profile = Agent + MCP Servers + Skills + Credentials + 资源限制
```

| 组件 | 作用 | 示例 |
|------|------|------|
| **MCP Servers** | 给 Agent 增加工具能力 | 文件系统、数据库、API 调用 |
| **Skills** | 预定义任务模板 | /commit、/review、/refactor |
| **Credentials** | 安全注入密钥 | API Key、Token |
| **资源限制** | 控制执行资源 | CPU、内存、超时时间 |

**价值**：同一个 Agent，不同 Profile，不同能力组合。

### 3. 容器化交付（Containerized Delivery）

每个任务运行在独立 Docker 容器中：

| 特性 | 说明 |
|------|------|
| **隔离性** | 任务间互不干扰，安全可控 |
| **一致性** | 开发、测试、生产环境一致 |
| **可扩展** | 水平扩展，支持高并发 |
| **可追溯** | 完整日志，执行可审计 |

**价值**：稳定、安全、可扩展的生产级交付。

### 4. 统一 API（Unified API）

一套 API 调用所有 Agent：

```bash
# 创建任务，不关心底层是 Claude Code 还是 Codex
curl -X POST https://api.agentbox.io/v1/tasks \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "profile": "code-review-expert",
    "prompt": "Review this pull request",
    "input": {"type": "git", "url": "..."}
  }'
```

**价值**：屏蔽底层差异，降低集成成本。

### 5. 开源自托管（Open Source & Self-Hosted）

| 部署方式 | 适用场景 |
|----------|----------|
| **自托管** | 企业私有部署，数据不出境 |
| **云服务** | 快速开始，免运维（未来） |

**价值**：数据安全，完全可控。

---

## 产品定位对比

### 与 Manus 的区别

| 维度 | Manus | AgentBox |
|------|-------|----------|
| **Agent** | 自研 Agent | 托管已有 Agent |
| **定位** | Agent + 平台（软硬一体） | 纯平台层（开放生态） |
| **扩展性** | 只能用 Manus Agent | 可接入任意 Agent |
| **部署** | 闭源 SaaS | 开源可自托管 |
| **类比** | iPhone | Android |

### 与其他方案的区别

| 方案 | 定位 | AgentBox 优势 |
|------|------|---------------|
| **直接使用 Agent CLI** | 本地运行 | 云端托管、无需配置环境 |
| **自建容器** | 自己搭建 | 开箱即用、统一管理 |
| **Agent SDK** | 嵌入应用 | 独立服务、解耦部署 |

---

## 目标用户

### 1. 开发者

> "我想用 AI Agent 帮我写代码、Review PR，但不想折腾环境配置"

- 通过 API 直接调用
- 无需本地安装 Agent
- 按需选择不同 Profile

### 2. 企业 DevOps

> "我们想把 AI Agent 能力集成到 CI/CD 流程中"

- 标准化 API 集成
- 容器化部署，适配现有基础设施
- 审计日志，合规需求

### 3. 平台提供商

> "我们想在自己的产品中提供 AI Agent 能力"

- 白标部署
- 多租户支持
- 灵活的 Profile 配置

### 4. AI Agent 开发者

> "我开发了新的 Agent，想让更多人使用"

- 标准化的 Agent 接入规范
- 复用平台的调度、监控能力
- 专注 Agent 核心能力开发

---

## 核心概念模型

```
┌─────────────────────────────────────────────────────────────────┐
│                        AgentBox 概念模型                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Agent         底层执行引擎                                      │
│                 Claude Code / Codex / OpenCode / ...            │
│       │                                                         │
│       ▼                                                         │
│   Profile       Agent 配置模板                                    │
│                 = Agent + MCP Servers + Skills + Credentials    │
│       │           + 权限 + 资源限制 + 系统提示词                   │
│       ▼                                                         │
│   Task          用户提交的任务                                    │
│                 = Profile + Prompt + Input + Output 配置         │
│       │                                                         │
│       ▼                                                         │
│   Session       运行中的任务实例                                   │
│                 = Docker 容器 + 执行状态 + 日志                   │
│       │                                                         │
│       ▼                                                         │
│   Output        任务执行结果                                      │
│                 = 文件 + 日志 + 元数据                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 概念说明

| 概念 | 说明 | 生命周期 |
|------|------|----------|
| **Agent** | 底层 AI Agent 引擎，如 Claude Code | 系统预置 |
| **Profile** | Agent 配置模板，定义能力组合 | 管理员配置 |
| **Task** | 用户提交的具体任务 | 用户创建 → 执行 → 完成 |
| **Session** | 任务的运行实例（容器） | 任务开始 → 运行 → 结束 |
| **Output** | 任务执行产出 | 任务完成后生成 |

---

## 系统架构

```
                              用户 / 应用
                                   │
                                   │ API
                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                          API Gateway                             │
│         认证 (API Key) │ 限流 │ 路由 │ 日志                       │
└─────────────────────────────────────────────────────────────────┘
                                   │
         ┌─────────────────────────┼─────────────────────────┐
         │                         │                         │
         ▼                         ▼                         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Task API     │     │    File API     │     │   Webhook API   │
│   任务管理服务   │     │   文件管理服务   │     │   回调通知服务   │
└────────┬────────┘     └─────────────────┘     └─────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Task Scheduler                            │
│          队列管理 │ 优先级调度 │ 并发控制 │ 失败重试               │
└─────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Profile Resolver                           │
│     解析 Profile → 组装完整配置（Agent + MCP + Skill + Cred）     │
└─────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Container Manager                           │
│      Docker SDK │ 容器生命周期 │ 资源限制 │ 网络隔离 │ 日志收集    │
└─────────────────────────────────────────────────────────────────┘
         │
         ├─────────────────────┬─────────────────────┐
         ▼                     ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  Claude Code    │  │     Codex       │  │   OpenCode      │
│   Container     │  │   Container     │  │   Container     │
│                 │  │                 │  │                 │
│  + MCP Server A │  │  + MCP Server B │  │  + MCP Server C │
│  + Skill X      │  │  + Skill Y      │  │  + Skill Z      │
│  + Credential   │  │  + Credential   │  │  + Credential   │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## API 设计概览

### 对外服务 API（给用户）

```
POST   /v1/tasks                    创建任务
GET    /v1/tasks                    列出任务
GET    /v1/tasks/{id}               获取任务详情
DELETE /v1/tasks/{id}               取消任务
GET    /v1/tasks/{id}/logs          获取执行日志
GET    /v1/tasks/{id}/output        获取任务输出

POST   /v1/files                    上传文件（获取上传 URL）
GET    /v1/files/{id}               获取文件信息
DELETE /v1/files/{id}               删除文件

POST   /v1/webhooks                 注册 Webhook
GET    /v1/webhooks                 列出 Webhook
DELETE /v1/webhooks/{id}            删除 Webhook

GET    /v1/profiles                 列出可用 Profile（公开的）
```

### 管理 API（给运维）

```
# Profile 管理
POST/PUT/DELETE /v1/profiles/{id}

# MCP Server 管理
POST/PUT/DELETE /v1/mcp-servers/{id}

# Skill 管理
POST/PUT/DELETE /v1/skills/{id}

# Credential 管理
POST/PUT/DELETE /v1/credentials/{id}

# 系统监控
GET /v1/system/health
GET /v1/system/stats
```

---

## 任务生命周期

```
            ┌──────────┐
            │ pending  │  用户刚提交，等待入队
            └────┬─────┘
                 │
                 ▼
            ┌──────────┐
            │  queued  │  已入队，等待调度
            └────┬─────┘
                 │
                 ▼
            ┌──────────┐
            │ running  │  容器启动，Agent 执行中
            └────┬─────┘
                 │
     ┌───────────┼───────────┐
     ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│completed │ │  failed  │ │cancelled │
│ 执行成功  │ │ 执行失败  │ │ 用户取消  │
└──────────┘ └──────────┘ └──────────┘
```

---

## 创建任务示例

### 请求

```json
POST /v1/tasks
{
  "profile": "claude-code-with-github",
  "prompt": "分析这个仓库的代码架构，生成 ARCHITECTURE.md 文档",
  "input": {
    "type": "git",
    "url": "https://github.com/example/repo",
    "branch": "main"
  },
  "output": {
    "type": "files"
  },
  "webhook": "https://my-app.com/webhook/agentbox",
  "timeout": 1800
}
```

### 响应

```json
{
  "id": "task_abc123",
  "status": "queued",
  "profile": "claude-code-with-github",
  "agent": "claude-code",
  "created_at": 1705641600,
  "estimated_start": 1705641660
}
```

### Webhook 回调

```json
POST https://my-app.com/webhook/agentbox
{
  "event": "task.completed",
  "task_id": "task_abc123",
  "status": "completed",
  "output": {
    "files": [
      {"name": "ARCHITECTURE.md", "size": 4521, "url": "..."}
    ]
  },
  "usage": {
    "duration_seconds": 120,
    "tokens": 15000
  }
}
```

---

## 未来规划

### Phase 1：核心平台（当前）
- [x] Agent 适配器（Claude Code、Codex）
- [x] Profile 配置管理
- [x] MCP Server 集成
- [x] Skill 管理
- [x] Credential 管理
- [x] Docker 容器管理
- [ ] Task API（进行中）
- [ ] 任务调度器
- [ ] Webhook 回调

### Phase 2：生产就绪
- [ ] 多租户支持
- [ ] API Key 认证
- [ ] 用量统计与配额
- [ ] 持久化队列
- [ ] 失败重试机制

### Phase 3：生态扩展
- [ ] 更多 Agent 支持（OpenCode、Aider）
- [ ] CLI 工具
- [ ] Python/Node.js SDK
- [ ] GitHub App 集成
- [ ] Profile 市场

### Phase 4：企业特性
- [ ] RBAC 权限控制
- [ ] 审计日志
- [ ] SSO 集成
- [ ] 私有 Agent 仓库

---

## 总结

**AgentBox** 不是另一个 AI Agent，而是 **AI Agent 的运行平台**。

我们的使命：

> 让任何 AI Agent 都能被轻松、安全、可靠地交付给用户。

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   用户提交任务  →  AgentBox 调度执行  →  Agent 完成任务     │
│                                                             │
│   用户不需要关心：                                           │
│   - Agent 怎么安装                                          │
│   - 环境怎么配置                                            │
│   - 容器怎么运行                                            │
│   - 资源怎么隔离                                            │
│                                                             │
│   用户只需要关心：                                           │
│   - 选择合适的 Profile                                      │
│   - 描述清楚任务                                            │
│   - 获取执行结果                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```
