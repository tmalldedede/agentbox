# Logger 使用指南

## 快速开始

```go
import "github.com/tmalldedede/agentbox/internal/logger"

// 初始化（在 main 中调用一次）
logger.Init(&logger.Config{
    Level:  "debug",  // debug, info, warn, error
    Format: "text",   // text, json
})

// 创建模块日志器（推荐）
log := logger.Module("session")
log.Info("session created", "id", "abc123", "agent", "codex")
log.Error("failed to start", "error", err)

// 或使用便捷方法
logger.Info("server started", "port", 18080)
```

## 输出格式

### Text 格式（开发环境）
```
15:04:05.000 INFO session created module=session id=abc123 agent=codex
15:04:05.100 ERROR failed to start module=session error="connection refused"
```

### JSON 格式（生产环境）
```json
{"time":"15:04:05.000","level":"INFO","msg":"session created","module":"session","id":"abc123"}
```

## 日志级别使用规范

| 级别 | 使用场景 | 示例 |
|------|----------|------|
| DEBUG | 调试信息，开发时使用 | 函数参数、中间状态 |
| INFO | 正常业务流程 | 请求处理、任务完成 |
| WARN | 可恢复的异常 | 重试、降级 |
| ERROR | 需要关注的错误 | 请求失败、资源不可用 |

## 模块命名规范

```go
// 按功能模块命名
logger.Module("api")        // HTTP API 处理
logger.Module("session")    // Session 管理
logger.Module("task")       // Task 管理
logger.Module("container")  // 容器操作
logger.Module("webhook")    // Webhook 处理
logger.Module("codex")      // Codex 适配器
```

## 键值对命名规范

```go
// 使用 snake_case，保持一致
log.Info("request handled",
    "request_id", reqID,
    "method", "POST",
    "path", "/api/v1/tasks",
    "status", 200,
    "duration_ms", 42,
)

// 错误始终用 "error" 键
log.Error("operation failed", "error", err)

// ID 类字段
log.Info("created", "session_id", sid, "task_id", tid)
```

## 迁移指南

### 从 log.Printf 迁移

```go
// Before
log.Printf("[Session] Created session %s", sessionID)

// After
log := logger.Module("session")
log.Info("created session", "id", sessionID)
```

### 从 fmt.Printf 迁移

```go
// Before (调试输出)
fmt.Printf("[DEBUG] parsing line: %s\n", line)

// After
log := logger.Module("parser")
log.Debug("parsing line", "content", line)
```
