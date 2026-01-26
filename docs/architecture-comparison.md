# AgentBox vs clawdbot 架构对比分析

## 一、项目定位

| 项目 | 定位 | 核心价值 |
|------|------|----------|
| **clawdbot** | 多渠道 AI 聊天机器人 | WhatsApp/Telegram/Discord 等消息平台的 AI 助手 |
| **AgentBox** | AI Agent 执行平台 | 统一管理和执行多种 Agent 任务的沙箱平台 |

---

## 二、整体架构对比

### 2.1 clawdbot 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         clawdbot                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  消息渠道 (Channels)                                    │     │
│  │  ├── WhatsApp (Baileys)                                │     │
│  │  ├── Telegram (Grammy)                                 │     │
│  │  ├── Discord (Carbon)                                  │     │
│  │  ├── Slack (@slack/bolt)                               │     │
│  │  ├── Signal                                            │     │
│  │  ├── LINE (@line/bot-sdk)                              │     │
│  │  └── iMessage (macOS)                                  │     │
│  └───────────────────────────────────────────────────────┘     │
│                         ↓                                       │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  Gateway 核心 (统一会话管理)                             │     │
│  │  ├── Session Manager (会话状态)                         │     │
│  │  ├── Auto Reply (自动回复触发)                          │     │
│  │  ├── Routing (消息路由)                                 │     │
│  │  └── Security (权限控制)                                │     │
│  └───────────────────────────────────────────────────────┘     │
│                         ↓                                       │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  Agent 引擎 (两种路径)                                   │     │
│  │                                                         │     │
│  │  【主要路径】pi-coding-agent SDK (带工具)                 │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ SessionManager (pi-coding-agent)                │   │     │
│  │  │ ├── 对话历史管理                                  │   │     │
│  │  │ ├── 上下文压缩 (estimateTokens/generateSummary)  │   │     │
│  │  │ └── Extension 机制                               │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Tools (自定义工具)                                │   │     │
│  │  │ ├── Read/Edit/Write (文件操作)                    │   │     │
│  │  │ ├── Bash (命令执行)                               │   │     │
│  │  │ ├── Channel Actions (发消息/管理群组)              │   │     │
│  │  │ ├── Image Tool (图像理解)                         │   │     │
│  │  │ ├── Skills (workspace 技能)                       │   │     │
│  │  │ └── Node Host (运行 JS/TS 代码)                   │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ LLM 调用 (pi-ai)                                  │   │     │
│  │  │ └── streamSimple() → 直接调 Claude/GPT API       │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  【Fallback 路径】CLI Backend (纯文本，无工具)           │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ spawn("claude", [...])  → Claude Code CLI       │   │     │
│  │  │ spawn("codex", [...])   → Codex CLI             │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  └───────────────────────────────────────────────────┘     │
│                         ↓                                       │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  OAuth 令牌复用                                          │     │
│  │  ├── anthropic:claude-cli (Keychain)                   │     │
│  │  ├── openai-codex (Keychain)                           │     │
│  │  └── 双向同步 (Clawdbot ↔ CLI)                          │     │
│  └───────────────────────────────────────────────────────┘     │
│                         ↓                                       │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  ACP Server (IDE 集成)                                  │     │
│  │  └── Agent Client Protocol (@agentclientprotocol/sdk)  │     │
│  └───────────────────────────────────────────────────────┘     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
                    Claude / GPT / Gemini API
```

**核心特点**：
- **单体应用**：所有逻辑在一个 Node.js 进程中
- **自己实现 Agent**：用 `pi-coding-agent` SDK 构建完整 Agent 能力
- **多渠道接入**：支持 7+ 消息平台
- **OAuth 复用**：可共享 Claude Code CLI 的登录凭证

---

### 2.2 AgentBox 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         AgentBox                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  前端 (Web UI - React + TypeScript)                     │     │
│  │  ├── Chat 界面 (实时 SSE)                               │     │
│  │  ├── Task 管理                                          │     │
│  │  ├── Agent 配置                                         │     │
│  │  ├── Provider 管理                                      │     │
│  │  └── MCP Server 管理                                    │     │
│  └───────────────────────────────────────────────────────┘     │
│                         ↓ HTTP + SSE                            │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  Go 后端 (API Server)                                   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ API 层 (Gin)                                     │   │     │
│  │  │ ├── /api/agents (Agent CRUD)                     │   │     │
│  │  │ ├── /api/providers (Provider + AuthProfile)      │   │     │
│  │  │ ├── /api/tasks (Task 创建/追加)                   │   │     │
│  │  │ ├── /api/tasks/:id/events (SSE 流)               │   │     │
│  │  │ └── /api/files (附件上传)                         │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Task Manager (任务调度)                           │   │     │
│  │  │ ├── Lane Queue (串行队列，防并发)                 │   │     │
│  │  │ ├── Fallback Executor (故障转移)                  │   │     │
│  │  │ ├── Idle Timeout (空闲超时)                       │   │     │
│  │  │ └── Event Broadcast (SSE 事件广播)                │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Session Manager (会话容器管理)                     │   │     │
│  │  │ ├── Docker API 封装                               │   │     │
│  │  │ ├── Workspace 挂载                                │   │     │
│  │  │ ├── 容器生命周期管理                               │   │     │
│  │  │ └── 日志收集                                      │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Agent Manager                                    │   │     │
│  │  │ ├── Agent 配置管理                                │   │     │
│  │  │ ├── Runtime 引用 (codex/claude-code)              │   │     │
│  │  │ ├── Skill 引用                                    │   │     │
│  │  │ └── MCP Server 引用                               │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Provider Manager + AuthProfile Rotation          │   │     │
│  │  │ ├── API Key 轮换                                  │   │     │
│  │  │ ├── 优先级调度                                    │   │     │
│  │  │ ├── 冷却时间管理                                  │   │     │
│  │  │ └── 失败统计                                      │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │                                                         │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Database (GORM)                                  │   │     │
│  │  │ ├── SQLite (开发)                                 │   │     │
│  │  │ ├── PostgreSQL (生产)                             │   │     │
│  │  │ └── 模型: Agent/Provider/Task/File/AuthProfile   │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  └───────────────────────────────────────────────────┘     │
│                         ↓ Docker API                            │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  Agent 容器 (agentbox/agent:latest)                     │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ 基础镜像: node:20                                 │   │     │
│  │  │ 用户: node (非 root)                              │   │     │
│  │  │ Shell: zsh + powerline                           │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ 预装 CLI                                          │   │     │
│  │  │ ├── Claude Code CLI (@anthropic-ai/claude-code)  │   │     │
│  │  │ └── Codex CLI (@openai/codex)                    │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ 开发工具                                          │   │     │
│  │  │ ├── git, gh, fzf, vim, nano                      │   │     │
│  │  │ ├── git-delta (better diffs)                     │   │     │
│  │  │ └── build-essential (编译依赖)                    │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Python + Skill 依赖 (73% 覆盖率)                   │   │     │
│  │  │ ├── requests, aiohttp, httpx (网络)               │   │     │
│  │  │ ├── dnspython, python-whois (DNS)                │   │     │
│  │  │ ├── beautifulsoup4, pandas (数据处理)             │   │     │
│  │  │ ├── pdfplumber, python-docx (文档)                │   │     │
│  │  │ ├── pillow, pyzbar (图像/二维码)                   │   │     │
│  │  │ └── rich, colorama (CLI 美化)                     │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Workspace 挂载                                    │   │     │
│  │  │ └── /workspace → session 隔离目录                 │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ Entrypoint                                       │   │     │
│  │  │ └── sleep infinity (等待后端注入命令)              │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  └───────────────────────────────────────────────────┘     │
│                         ↓                                       │
│  ┌───────────────────────────────────────────────────────┐     │
│  │  后端通过 Docker Exec 注入命令                            │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ codex "prompt" --model xxx                       │   │     │
│  │  │ claude-code "prompt" --model xxx                 │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  │  ┌─────────────────────────────────────────────────┐   │     │
│  │  │ 容器内 CLI 执行 → 调用 LLM API                     │   │     │
│  │  │ ├── 读取 workspace 文件                           │   │     │
│  │  │ ├── 执行工具调用                                  │   │     │
│  │  │ ├── 写入结果                                      │   │     │
│  │  │ └── 输出日志                                      │   │     │
│  │  └─────────────────────────────────────────────────┘   │     │
│  └───────────────────────────────────────────────────┘     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
                    Claude / GPT / Gemini API
```

**核心特点**：
- **平台架构**：前后端分离 + 容器沙箱
- **调用现成 CLI**：不自己实现 Agent，依赖 Claude Code / Codex
- **容器隔离**：每个任务独立容器，资源隔离
- **多租户支持**：可扩展为多用户平台

---

## 三、数据流转对比

### 3.1 clawdbot 消息处理流程

```
用户 (WhatsApp)
    ↓ 发送消息
Baileys SDK 接收
    ↓
Channel Plugin 解析
    ↓
Gateway Router 路由
    ↓ 判断是否触发 auto-reply
Auto Reply 模块
    ↓ 如果触发
Session Manager 查找/创建会话
    ↓
pi-coding-agent SessionManager
    ↓ 加载历史 + 工具定义
pi-ai streamSimple()
    ↓ HTTP POST /v1/messages
Claude API
    ↓ SSE Stream
    ↓ tool_use / text
pi-coding-agent 执行工具
    ↓ read_file / edit_file / bash
文件系统 / Shell
    ↓ 结果返回
Agent 继续对话
    ↓ 完成后
Gateway 格式化输出
    ↓
Channel Plugin 发送
    ↓
用户收到消息 (WhatsApp)
```

**关键点**：
- **同步模式**：用户发消息 → 立即处理 → 返回结果
- **进程内执行**：所有逻辑在 Node.js 进程内
- **工具直接调用**：文件操作、Shell 直接在宿主机执行
- **无沙箱**：运行在服务器本地环境

---

### 3.2 AgentBox 任务处理流程

```
用户 (Web UI)
    ↓ 发送消息 (POST /api/tasks)
Go 后端接收
    ↓
Task Manager 创建任务
    ↓ 写入数据库
Database (tasks 表)
    ↓
Task Manager 调度器轮询
    ↓ 获取 queued 任务
Lane Queue 排队
    ↓ 防止同一 Provider 并发
Fallback Executor 选择可用 Provider
    ↓ 根据优先级 + 冷却时间
Session Manager 创建容器
    ↓ docker run agentbox/agent:latest
Agent 容器启动
    ↓ sleep infinity
Session Manager 注入命令
    ↓ docker exec session-xxx codex "prompt"
Codex CLI 执行
    ↓ 读取 workspace 文件
    ↓ HTTP POST /v1/chat/completions
OpenAI API
    ↓ SSE Stream
    ↓ tool_use / text
Codex CLI 执行工具
    ↓ 在容器内隔离执行
容器文件系统 / Shell
    ↓ 结果返回
Codex CLI 输出
    ↓ stdout/stderr
Session Manager 收集日志
    ↓ 解析 SSE 事件
Task Manager 广播事件
    ↓ SSE: agent.thinking / agent.message
前端 (useChat hook)
    ↓ 实时显示流式输出
用户看到响应 (Chat 界面)
    ↓ 任务完成
Task Manager 更新状态
    ↓
Database (tasks.status = completed)
    ↓ 清理容器
Session Manager 删除容器
```

**关键点**：
- **异步模式**：创建任务 → 后台调度 → SSE 推送进度
- **容器隔离**：每个任务在独立 Docker 容器中运行
- **CLI 黑盒调用**：不直接控制 Agent 逻辑
- **沙箱安全**：容器限制资源和权限

---

## 四、关键组件对比

### 4.1 Agent 实现方式

| 对比项 | clawdbot | AgentBox |
|--------|----------|----------|
| **核心依赖** | `@mariozechner/pi-coding-agent` (SDK) | `claude-code` / `codex` (CLI) |
| **LLM 调用** | 自己实现 (`pi-ai` 直接调 API) | CLI 封装 |
| **工具定义** | 代码中自定义 (`createClawdbotCodingTools`) | CLI 内置 |
| **Session 管理** | 自己实现 (`SessionManager`) | CLI 管理 |
| **上下文压缩** | 自己实现 (`estimateTokens`, `generateSummary`) | CLI 内置 |
| **扩展性** | 极高（完全可控） | 低（CLI 黑盒） |
| **开发成本** | 高（需要理解 Agent 原理） | 低（开箱即用） |

---

### 4.2 会话管理

| 对比项 | clawdbot | AgentBox |
|--------|----------|----------|
| **存储方式** | 文件 (~/.clawdbot/sessions/) | Docker 容器 + 数据库 |
| **持久化** | JSON 文件 | PostgreSQL / SQLite |
| **隔离级别** | 进程内 (按 sessionKey) | 容器级别 (每任务独立容器) |
| **多轮对话** | SessionManager 自动管理 | 通过 appendTurn API + 容器复用 |
| **上下文限制** | 手动压缩 (compaction) | CLI 自动处理 |

---

### 4.3 工具执行

| 工具类型 | clawdbot | AgentBox |
|---------|----------|----------|
| **文件读写** | 直接操作宿主机 | 容器内 workspace（挂载） |
| **Shell 命令** | 直接执行（危险） | 容器内执行（隔离） |
| **自定义工具** | TypeScript 代码实现 | Skill 机制（预装 Python 依赖） |
| **消息平台操作** | 直接调 SDK (Baileys/Grammy) | 不支持 |
| **安全性** | 需要手动控制 | 容器天然隔离 |

---

### 4.4 多 Provider 支持

| 特性 | clawdbot | AgentBox |
|------|----------|----------|
| **Provider 定义** | config.json5 + auth profiles | 数据库 (providers 表) |
| **API Key 管理** | 文件存储 + Keychain | 数据库 + AuthProfile 轮换 |
| **Fallback** | CLI Backend (spawn CLI 进程) | Fallback Executor (多 Provider 故障转移) |
| **OAuth** | 复用 Claude Code CLI 令牌 | 不支持 OAuth |
| **轮换策略** | 无自动轮换 | 优先级 + 冷却时间 + 失败统计 |

---

### 4.5 部署模式

| 对比项 | clawdbot | AgentBox |
|--------|----------|----------|
| **部署方式** | Docker 单容器 / 本地进程 | Go 二进制 + Docker (Agent 容器) |
| **依赖** | Node.js 22 + pnpm | Go 1.22 + Docker |
| **横向扩展** | 不支持（单体） | 可扩展（无状态后端） |
| **存储** | 本地文件 | 数据库 (SQLite/PostgreSQL) |
| **高可用** | 单点 | 可负载均衡（数据库共享） |

---

## 五、技术栈对比

### 5.1 clawdbot 技术栈

| 层级 | 技术选型 |
|------|----------|
| **语言** | TypeScript |
| **运行时** | Node.js 22 |
| **包管理** | pnpm |
| **Agent 框架** | `@mariozechner/pi-coding-agent` |
| **LLM 调用** | `@mariozechner/pi-ai` |
| **消息渠道** | Baileys, Grammy, Carbon, @slack/bolt 等 |
| **IDE 集成** | `@agentclientprotocol/sdk` |
| **UI** | TUI (@mariozechner/pi-tui) + Web (Lit) |
| **配置** | JSON5 |
| **测试** | Vitest |

---

### 5.2 AgentBox 技术栈

| 层级 | 技术选型 |
|------|----------|
| **后端语言** | Go 1.22 |
| **Web 框架** | Gin |
| **ORM** | GORM |
| **数据库** | SQLite (开发) / PostgreSQL (生产) |
| **容器** | Docker API |
| **Agent 引擎** | Claude Code CLI / Codex CLI |
| **前端语言** | TypeScript |
| **前端框架** | React 18 + TanStack Router |
| **状态管理** | Zustand + TanStack Query |
| **UI 组件** | shadcn/ui (Radix UI + Tailwind) |
| **实时通信** | SSE (Server-Sent Events) |
| **构建工具** | Vite |

---

## 六、优缺点分析

### 6.1 clawdbot

**优点**：
1. ✅ **功能完整**：自己实现 Agent，完全可控
2. ✅ **多渠道接入**：支持 7+ 消息平台
3. ✅ **灵活扩展**：可自定义工具、Extension
4. ✅ **OAuth 复用**：共享 Claude Code CLI 凭证
5. ✅ **IDE 集成**：支持 ACP 协议
6. ✅ **低延迟**：进程内执行，无容器开销

**缺点**：
1. ❌ **单体架构**：难以横向扩展
2. ❌ **安全风险**：Shell 命令直接在宿主机执行
3. ❌ **资源隔离差**：所有会话共享进程资源
4. ❌ **开发门槛高**：需要深入理解 Agent 原理
5. ❌ **部署复杂**：依赖 Node.js 环境配置

---

### 6.2 AgentBox

**优点**：
1. ✅ **容器隔离**：每个任务独立沙箱，安全
2. ✅ **易扩展**：无状态后端，可负载均衡
3. ✅ **开箱即用**：调用 Claude Code/Codex CLI
4. ✅ **Web UI**：用户友好的界面
5. ✅ **Provider 轮换**：自动故障转移
6. ✅ **多租户架构**：支持多用户隔离

**缺点**：
1. ❌ **容器开销**：启动慢（~5秒）
2. ❌ **CLI 黑盒**：无法深度定制 Agent 行为
3. ❌ **功能受限**：依赖 CLI 支持的功能
4. ❌ **无消息平台**：仅 Web UI，不支持 WhatsApp 等
5. ❌ **资源消耗大**：每个任务占用一个容器

---

## 七、适用场景

### 7.1 clawdbot 适合

- ✅ 需要在 WhatsApp/Telegram 等平台部署 AI 助手
- ✅ 需要完全控制 Agent 行为和工具
- ✅ 单租户/个人使用场景
- ✅ 对延迟要求高（秒级响应）
- ✅ 有前端开发能力，可自定义扩展

---

### 7.2 AgentBox 适合

- ✅ 企业级 Agent 执行平台
- ✅ 多租户 SaaS 场景
- ✅ 需要沙箱隔离保证安全
- ✅ 快速集成现有 LLM（调用 CLI 即可）
- ✅ 需要 Web UI 管理 Agent

---

## 八、总结

| 维度 | clawdbot | AgentBox |
|------|----------|----------|
| **定位** | 多渠道聊天机器人 | Agent 执行平台 |
| **架构** | 单体应用 | 平台 + 沙箱 |
| **Agent 实现** | 自己写（SDK） | 调用 CLI |
| **部署** | 简单（单容器） | 复杂（前后端 + Docker） |
| **扩展性** | 低（单体） | 高（可集群） |
| **安全性** | 低（宿主机执行） | 高（容器隔离） |
| **灵活性** | 极高（完全可控） | 中等（CLI 限制） |
| **开发难度** | 高（需懂原理） | 低（开箱即用） |

**选择建议**：
- **个人/小团队**：用 clawdbot，功能强大、灵活
- **企业/平台**：用 AgentBox，安全、可扩展
