# AgentBox 产品愿景

## 一句话定位

> **AgentBox = 开源的企业版 Cowork + Agent 无关的 E2B + 可自托管的 Manus**
>
> 将 Claude Code、Codex、OpenCode 等专业 AI Agent，通过可配置的 Profile 系统（MCP Servers + Skills + Credentials），以 Docker 容器安全交付给企业用户。开源可自托管，数据不出境。

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

### 核心竞品对比

| 维度 | Claude Cowork | E2B | Manus | **AgentBox** |
|------|---------------|-----|-------|--------------|
| **Agent** | 仅 Claude | Agent 无关 | 自研 Agent | **Agent 无关** |
| **目标用户** | 消费者 ($20/月) | 开发者/企业 | 消费者 | **企业** |
| **部署方式** | macOS 本地 | 云托管/私有部署 | 闭源 SaaS | **开源自托管** |
| **隔离技术** | Apple VM 沙盒 | Firecracker microVM | 未知 | **Docker 容器** |
| **扩展能力** | Connectors | SDK | 无 | **Profile + MCP + Skills** |
| **数据安全** | 本地 | 可私有部署 | 数据在云端 | **完全可控** |
| **类比** | 个人助手 | 开发者工具 | iPhone | **Android** |

### AgentBox 的差异化

| 差异点 | 说明 |
|--------|------|
| **不造轮子** | 托管已有最强 Agent（Claude Code、Codex），而非自研 |
| **企业级可控** | 开源、私有部署、审计日志、资源隔离 |
| **统一能力增强** | 一套 Profile 系统为所有 Agent 注入 MCP/Skills/Credentials |
| **Agent 可替换** | 今天用 Claude Code，明天换 Codex，无需改代码 |

### 与其他方案的区别

| 方案 | 定位 | AgentBox 优势 |
|------|------|---------------|
| **直接使用 Agent CLI** | 本地运行 | 云端托管、无需配置环境 |
| **自建容器** | 自己搭建 | 开箱即用、统一管理 |
| **Agent SDK** | 嵌入应用 | 独立服务、解耦部署 |

---

## 目标用户画像

### 用户画像总览

| 用户类型 | 核心诉求 | 痛点 | AgentBox 价值 |
|----------|----------|------|---------------|
| **企业技术团队** | 将 AI Agent 集成到内部系统 | 环境配置复杂、数据安全顾虑 | 私有部署、统一 API |
| **SaaS 平台商** | 为用户提供 AI Agent 能力 | 自研成本高、多租户管理难 | 白标部署、Profile 隔离 |
| **安全合规企业** | 使用 AI 但数据不能出境 | 云服务无法满足合规 | 开源自托管、审计日志 |
| **AI 应用开发者** | 快速集成 Agent 能力 | Agent 选型困难、切换成本高 | Agent 无关、统一接口 |

---

### 画像 1：企业技术团队负责人

**角色**：Tech Lead / 架构师 / DevOps 负责人

**公司特征**：
- 中大型企业（100+ 人技术团队）
- 有自己的 CI/CD 流水线
- 对数据安全和合规有严格要求

**核心痛点**：
| 痛点 | 详情 |
|------|------|
| 环境一致性 | 开发、测试、生产环境配置不一致，Agent 表现不稳定 |
| 资源隔离 | 多个项目共用 Agent，担心资源竞争和数据泄露 |
| 审计合规 | 需要记录 AI Agent 的所有操作，满足安全审计要求 |
| 私有部署 | 代码和数据不能传到外部服务，必须自托管 |

**使用场景**：
```
1. 将 AI Agent 集成到代码审查流程
   - PR 提交 → 触发 AgentBox Task → Agent 执行 Review → 结果回写 GitHub

2. 自动化文档生成
   - 代码变更 → 触发 AgentBox Task → Agent 更新文档 → 生成 PR

3. 安全漏洞扫描
   - 定时任务 → AgentBox 批量扫描仓库 → 输出漏洞报告
```

**期望价值**：
- 统一 API，无需为每个 Agent 写适配代码
- Docker 容器隔离，与现有 K8s 基础设施兼容
- 完整审计日志，满足 SOC2/等保合规

---

### 画像 2：SaaS 平台产品经理

**角色**：产品经理 / 技术合伙人

**公司特征**：
- SaaS 创业公司或中型 ISV
- 想在自己产品中提供 AI 能力
- 有多租户、计费需求

**核心痛点**：
| 痛点 | 详情 |
|------|------|
| 自研成本高 | 自己开发 Agent 太复杂，不如借用现有能力 |
| 多租户隔离 | 不同客户的任务需要完全隔离 |
| 能力可配置 | 不同客户套餐需要不同的 Agent 能力组合 |
| 用量计费 | 需要精确统计每个租户的资源消耗 |

**使用场景**：
```
1. 代码托管平台提供 AI 代码审查
   - 用户点击"AI Review"→ 调用 AgentBox API → 返回审查结果

2. 文档平台提供 AI 写作助手
   - 用户上传素材 → AgentBox 执行生成任务 → 返回文档

3. 低代码平台集成 AI 生成
   - 用户描述需求 → AgentBox 生成代码 → 导入低代码编辑器
```

**期望价值**：
- 白标部署，用户无感知底层 Agent
- Profile 实现套餐差异化（基础版/专业版/企业版）
- 多租户 + 用量统计，支撑计费系统

---

### 画像 3：安全合规企业 IT 主管

**角色**：IT 主管 / 信息安全负责人

**公司特征**：
- 金融、医疗、政府等强监管行业
- 数据不能出境，禁止使用公有云 AI 服务
- 有严格的软件准入审批流程

**核心痛点**：
| 痛点 | 详情 |
|------|------|
| 数据出境 | 监管禁止代码/文档传到境外服务器 |
| 审批流程 | 引入新软件需要安全评估，闭源软件难以通过 |
| 网络隔离 | 内网环境无法访问外部 API |
| 可追溯 | 所有 AI 操作必须有完整审计记录 |

**使用场景**：
```
1. 内网部署 AI 编程助手
   - 开发人员使用内网 AgentBox → 代码不出内网

2. 合规文档自动化
   - 输入业务数据 → AgentBox 生成合规报告 → 数据本地存储

3. 代码安全扫描
   - 私有仓库 → AgentBox 扫描 → 结果存入内部系统
```

**期望价值**：
- 开源可审计，安全团队可以审查代码
- 完全私有部署，数据不出内网
- 详细的操作日志，满足审计要求

---

### 画像 4：AI 应用开发者

**角色**：独立开发者 / 技术创业者

**公司特征**：
- 小团队或个人开发者
- 想快速构建 AI 驱动的应用
- 对成本敏感，希望按需付费

**核心痛点**：
| 痛点 | 详情 |
|------|------|
| Agent 选型 | Claude Code vs Codex vs 其他，不知道选哪个 |
| 切换成本 | 换 Agent 要改大量代码 |
| 环境配置 | 每个 Agent 安装配置方式不同 |
| 能力扩展 | 想给 Agent 加工具，但不知道怎么做 |

**使用场景**：
```
1. 快速验证产品 MVP
   - 调用 AgentBox API → 验证 AI 能力 → 决定是否深入开发

2. 多 Agent 对比测试
   - 同一任务分别用 Claude Code 和 Codex → 对比效果

3. 自定义 Agent 能力
   - 通过 Profile 组合 MCP + Skills → 构建专属 Agent
```

**期望价值**：
- 一套 API 调用多个 Agent，轻松切换
- Profile 模板快速上手
- 开源免费，可自托管控制成本

---

### 用户优先级

| 优先级 | 用户类型 | 原因 |
|--------|----------|------|
| **P0** | 企业技术团队 | 核心目标用户，付费意愿强 |
| **P0** | 安全合规企业 | 开源自托管是核心差异化 |
| **P1** | SaaS 平台商 | 高价值用户，但需求复杂度高 |
| **P2** | AI 应用开发者 | 社区推广，但商业价值有限 |

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
