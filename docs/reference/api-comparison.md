# API 对比分析: AgentBox vs ClaudeBox-Server vs Manus

## 概念映射

| 概念 | AgentBox | ClaudeBox-Server | Manus |
|------|----------|------------------|-------|
| 顶层组织 | - | - | Project |
| 执行环境 | Session | Session | - |
| 任务执行 | Execution | Execution | Task |
| 文件存储 | Workspace | Workspace | Files (S3) |
| 异步通知 | - | - | Webhook |

---

## 功能对比矩阵

### Session/Project 管理

| 功能 | AgentBox | ClaudeBox | Manus | 说明 |
|------|:--------:|:---------:|:-----:|------|
| 创建 | ✅ | ✅ | ✅ | POST /sessions 或 /projects |
| 列表 (分页) | ✅ | ✅ | ✅ | 支持 limit/offset |
| 获取详情 | ✅ | ✅ | - | GET /sessions/:id |
| 删除 | ✅ | ✅ | - | DELETE /sessions/:id |
| 启动 | ✅ | - | - | POST /sessions/:id/start |
| 停止 | ✅ | - | - | POST /sessions/:id/stop |
| 重连 | ❌ | ✅ | - | POST /sessions/:id/reconnect |

### 任务执行

| 功能 | AgentBox | ClaudeBox | Manus | 说明 |
|------|:--------:|:---------:|:-----:|------|
| 执行任务 | ✅ | ✅ | ✅ | POST /exec 或 /tasks |
| 执行历史 | ✅ | ✅ | ✅ | GET /executions 或 /history |
| 执行详情 | ✅ | - | ✅ | GET /executions/:id |
| 更新任务 | - | - | ✅ | PUT /tasks/:id |
| 删除任务 | - | - | ✅ | DELETE /tasks/:id |
| 执行超时 | ❌ | ✅ | ✅ | timeout 参数 |
| 工具控制 | ❌ | ✅ | - | allowed/disallowed_tools |
| 最大轮数 | ❌ | ✅ | ✅ | max_turns 参数 |

### 文件管理

| 功能 | AgentBox | ClaudeBox | Manus | 说明 |
|------|:--------:|:---------:|:-----:|------|
| 列出文件 | ❌ | ✅ | ✅ | GET /files |
| 上传文件 | ❌ | ✅ | ✅ | POST /files |
| 下载文件 | ❌ | ✅ | - | GET /files/:path |
| 删除文件 | ❌ | ✅ | ✅ | DELETE /files/:path |
| 创建目录 | ❌ | ✅ | - | POST /directories |
| 读取内容 | ❌ | ✅ | - | GET /files/:path/content |

### 实时通信

| 功能 | AgentBox | ClaudeBox | Manus | 说明 |
|------|:--------:|:---------:|:-----:|------|
| WebSocket 执行流 | ❌ | ✅ | - | WS /sessions/:id/stream |
| WebSocket 日志流 | ❌ | ✅ | - | WS /logs/stream |
| Webhook 回调 | ❌ | - | ✅ | POST /webhooks |

### 日志

| 功能 | AgentBox | ClaudeBox | Manus | 说明 |
|------|:--------:|:---------:|:-----:|------|
| 容器/会话日志 | ✅ | ✅ | - | GET /sessions/:id/logs |
| 服务器日志 | ❌ | ✅ | - | GET /logs/server |
| 全局日志 | ❌ | ✅ | - | GET /logs |

---

## AgentBox 需要增补的 API

### 🔴 P0 - 核心功能 (必须)

#### 1. 文件管理模块

```
GET    /api/v1/sessions/:id/files              # 列出文件
GET    /api/v1/sessions/:id/files/*path        # 下载文件
POST   /api/v1/sessions/:id/files              # 上传文件
DELETE /api/v1/sessions/:id/files/*path        # 删除文件
POST   /api/v1/sessions/:id/directories        # 创建目录
GET    /api/v1/sessions/:id/files/*path/content # 读取文本内容
```

#### 2. 执行参数增强

```go
type ExecRequest struct {
    Prompt          string   `json:"prompt" binding:"required"`
    MaxTurns        int      `json:"max_turns"`        // 新增: 最大轮数
    Timeout         int      `json:"timeout"`          // 新增: 超时秒数
    AllowedTools    []string `json:"allowed_tools"`    // 新增: 允许的工具
    DisallowedTools []string `json:"disallowed_tools"` // 新增: 禁用的工具
}
```

### 🟡 P1 - 重要功能

#### 3. WebSocket 实时流

```
WS /api/v1/sessions/:id/stream   # 实时执行流
```

消息类型:
- 客户端: `execute`, `ping`
- 服务端: `execution_started`, `execution_completed`, `error`, `pong`, `heartbeat`

#### 4. 会话重连

```
POST /api/v1/sessions/:id/reconnect
```

### 🟢 P2 - 增强功能

#### 5. Webhook 支持

```
POST   /api/v1/webhooks              # 创建 Webhook
GET    /api/v1/webhooks              # 列出 Webhook
DELETE /api/v1/webhooks/:id          # 删除 Webhook
```

#### 6. 日志增强

```
GET /api/v1/logs                     # 全局日志
GET /api/v1/logs/server              # 服务器日志
WS  /api/v1/logs/stream              # 实时日志流
```

---

## 建议的实现顺序

1. **Phase 1: 文件管理** (6 个端点)
   - 这是最关键的缺失功能
   - 用户无法查看/管理工作区文件

2. **Phase 2: 执行增强** (参数扩展)
   - 添加 timeout, max_turns, tools 控制
   - 提升执行可控性

3. **Phase 3: WebSocket 流** (2 个端点)
   - 实时执行输出
   - 实时日志推送

4. **Phase 4: Webhook** (3 个端点)
   - 异步任务完成通知
   - 集成第三方系统

---

## 数据模型扩展建议

### FileInfo

```go
type FileInfo struct {
    Name         string    `json:"name"`
    Type         string    `json:"type"` // file, directory
    Size         int64     `json:"size,omitempty"`
    ModifiedAt   time.Time `json:"modified_at,omitempty"`
    ChildrenCount int      `json:"children_count,omitempty"` // 仅目录
}
```

### Webhook

```go
type Webhook struct {
    ID        string    `json:"id"`
    URL       string    `json:"url"`
    Events    []string  `json:"events"` // task.completed, task.failed
    CreatedAt time.Time `json:"created_at"`
}
```

### WebSocket 消息

```go
// 客户端发送
type WSExecuteMessage struct {
    Type            string   `json:"type"` // execute, ping
    Prompt          string   `json:"prompt,omitempty"`
    MaxTurns        int      `json:"max_turns,omitempty"`
    Timeout         int      `json:"timeout,omitempty"`
    AllowedTools    []string `json:"allowed_tools,omitempty"`
    DisallowedTools []string `json:"disallowed_tools,omitempty"`
}

// 服务端发送
type WSResultMessage struct {
    Type        string `json:"type"` // execution_started, execution_completed, error, pong
    ExecutionID string `json:"execution_id,omitempty"`
    Success     bool   `json:"success,omitempty"`
    Response    string `json:"response,omitempty"`
    ExitCode    int    `json:"exit_code,omitempty"`
    Error       string `json:"error,omitempty"`
    DurationMs  int64  `json:"duration_ms,omitempty"`
    Timestamp   string `json:"timestamp"`
}
```
