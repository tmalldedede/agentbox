# AgentBox

开源的 AI Agent 容器化运行平台

## 使命

让每个 AI Agent 都能在安全、隔离的环境中自由运行。

## 特性

- 🐳 **Docker 容器隔离** - 每个会话独立容器，安全可控
- 🔌 **多 Agent 支持** - Claude Code, Codex, 可扩展更多
- 🌐 **HTTP API 优先** - RESTful API，易于集成
- 📊 **统一管理后台** - Web UI 管理所有会话
- 🔒 **企业级安全** - 资源限制、网络隔离、审计日志
- 📦 **单二进制部署** - Go 编译，无运行时依赖

## 快速开始

### 前置要求

- Go 1.21+
- Docker Desktop
- Node.js 18+ (Web UI 可选)

### 1. 构建 Agent 镜像

```bash
# 克隆项目
git clone https://github.com/tmalldedede/agentbox.git
cd agentbox

# 构建 Agent 镜像（包含 Claude Code 和 Codex CLI）
docker build -t agentbox/agent:latest -f docker/agent/Dockerfile .
```

### 2. 运行后端

```bash
# 构建
make build

# 运行（默认端口 8080）
./bin/agentbox

# 或指定端口
AGENTBOX_PORT=18080 ./bin/agentbox
```

### 3. 运行 Web UI（可选）

```bash
cd web
npm install
npm run dev
# 访问 http://localhost:5173
```

## API 示例

```bash
# 健康检查
curl http://localhost:8080/health

# 列出可用 Agent
curl http://localhost:8080/api/agents

# 创建会话
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "claude-code",
    "workspace": "/path/to/project",
    "env": {"ANTHROPIC_API_KEY": "your-key"}
  }'

# 列出会话
curl http://localhost:8080/api/sessions

# 删除会话
curl -X DELETE http://localhost:8080/api/sessions/{id}
```

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                     AgentBox Server                      │
├─────────────┬─────────────┬─────────────┬───────────────┤
│   HTTP API  │  WebSocket  │   Session   │    Config     │
│             │   (Logs)    │   Manager   │    Manager    │
├─────────────┴─────────────┴─────────────┴───────────────┤
│                    Container Manager                     │
├─────────────────────────────────────────────────────────┤
│                      Docker Engine                       │
└─────────────────────────────────────────────────────────┘
```

## 支持的 Agent

| Agent | 状态 | 镜像 | 环境变量 |
|-------|------|------|----------|
| Claude Code | ✅ 可用 | `agentbox/agent:latest` | `ANTHROPIC_API_KEY` |
| Codex | ✅ 可用 | `agentbox/agent:latest` | `OPENAI_API_KEY` |

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/agents` | 列出可用 Agent |
| POST | `/api/sessions` | 创建会话 |
| GET | `/api/sessions` | 列出所有会话 |
| GET | `/api/sessions/:id` | 获取会话详情 |
| DELETE | `/api/sessions/:id` | 删除会话 |
| POST | `/api/sessions/:id/stop` | 停止会话 |
| POST | `/api/sessions/:id/start` | 启动会话 |
| POST | `/api/sessions/:id/exec` | 执行任务 |

## 目录结构

```
agentbox/
├── cmd/agentbox/       # 主程序入口
├── internal/
│   ├── api/            # HTTP API
│   ├── agent/          # Agent 适配器
│   ├── container/      # Docker 容器管理
│   ├── session/        # 会话管理
│   └── config/         # 配置管理
├── docker/
│   └── agent/          # Agent 镜像 Dockerfile
└── web/                # Web 管理界面
```

## 文档

- [架构设计](./DESIGN.md)

## 许可证

[Apache-2.0](./LICENSE)
