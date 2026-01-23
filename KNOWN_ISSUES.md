# Known Issues

## Issue 1: Docker 流解复用不完整 (低优先级)

**位置**: `internal/container/docker.go:140-145`

**问题**: `Exec` 函数使用 `io.Copy(&stdout, attachResp.Reader)` 直接读取 Docker 多路复用流，没有使用 `stdcopy.StdCopy` 正确分离 stdout/stderr。导致 `result.Stdout` 包含 8 字节 Docker 帧头（stream_type + size）。

**当前修补**: `internal/session/manager.go` 的 `stripDockerStreamHeaders()` 在应用层剥离帧头。JSONL 解析器通过 `strings.Index(line, "{")` 跳过帧头。

**正确修复**:
```go
import "github.com/docker/docker/pkg/stdcopy"

var stdout, stderr bytes.Buffer
stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
```

---

## Issue 2: Claude Code 输出包含控制字符前缀

**位置**: `internal/engine/claude/adapter.go`

**问题**: Claude Code CLI 输出的每行前面带有 `\x01` 或 `\x02` 等控制字符前缀（可能是 Docker TTY 模式或 Claude Code 自身的格式化标记），导致解析结果包含不可见字符。

**复现**: 使用 claude-code adapter 执行任意 prompt，观察 output 中的控制字符。

**待调查**: 确认是 Docker multiplexing 问题（与 Issue 1 同源）还是 Claude Code 自身输出格式。

---

## Issue 3: Claude Code 多轮对话未实现

**位置**: `internal/engine/claude/adapter.go`

**问题**: Claude Code 的多轮对话机制与 Codex 不同，当前未实现 resume/thread 支持。Claude Code 可能需要 `--resume` 参数或不同的 session 管理方式。

**依赖**: 需要先解决 Issue 2（控制字符问题），确认 Claude Code 的基本输出解析正常后再实现多轮。

---

## Issue 4: Codex 容器版本过旧 (0.77.0)

**位置**: Docker 镜像 `agentbox/agent:latest`

**问题**: 容器内 codex 版本为 0.77.0，主机为 0.89.0。旧版本的 `resume` 子命令不支持任何 flag（包括 `--json`、`--skip-git-repo-check`）。

**影响**:
- Resume 输出为纯文本，无法获取 token usage 等结构化信息
- 需要 `sh -c "cd /tmp && git init -q; ..."` 包装来绕过 git repo 检查

**修复**: 升级容器镜像中的 codex 版本到 0.89.0+，之后可恢复 `--json` 输出。

---

## Issue 5: 前端 Task 详情页未实现

**位置**: `web/src/features/tasks/`

**问题**: Tasks 列表页使用真实 API 数据，但缺少 Task 详情页（点击 task 查看多轮对话历史、SSE 实时输出、追加轮次输入框）。

**需要**:
- `web/src/routes/_authenticated/tasks/$id.tsx` — 路由
- `web/src/features/tasks/components/TaskDetail.tsx` — 详情组件
- SSE hook 连接 `/api/v1/tasks/:id/events`
- `useAppendTurn` hook 追加轮次
