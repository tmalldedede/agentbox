# WebSocket Gateway

AgentBox WebSocket Gateway 提供实时双向通信能力，支持任务事件订阅、会话流式执行和系统事件推送。

## 连接

WebSocket 端点:
```
ws://localhost:18080/api/v1/gateway/ws
```

## 消息协议

所有消息采用 JSON 格式:

```json
{
  "id": "msg-123",         // 可选，用于请求/响应匹配
  "type": "subscribe",     // 消息类型
  "payload": {...},        // 消息负载
  "timestamp": 1706234567890
}
```

## 认证

连接后需要发送认证消息:

```json
{
  "type": "auth",
  "payload": {
    "token": "ab_xxx...",   // API Key
    "device_id": "device-1" // 可选，设备标识
  }
}
```

响应:
```json
{
  "type": "auth.result",
  "payload": {
    "success": true,
    "user_id": "user-123",
    "device_id": "device-1"
  }
}
```

## 订阅事件

### 订阅任务事件

```json
{
  "type": "subscribe",
  "payload": {
    "channel": "task",
    "topics": ["task-123"]  // 空数组表示订阅所有
  }
}
```

### 订阅系统事件

```json
{
  "type": "subscribe",
  "payload": {
    "channel": "system",
    "topics": ["tasks", "agents"]  // 订阅任务和 Agent 创建事件
  }
}
```

### 取消订阅

```json
{
  "type": "unsubscribe",
  "payload": {
    "channel": "task",
    "topics": ["task-123"]  // 空数组表示取消整个频道
  }
}
```

## 事件推送

订阅成功后，会收到事件推送:

```json
{
  "type": "event",
  "payload": {
    "channel": "task",
    "topic": "task-123",
    "event_type": "agent.message",
    "data": {
      "content": "Hello! I can help you with that."
    }
  }
}
```

### 任务事件类型

| 事件类型 | 说明 |
|---------|------|
| `task.started` | 任务开始执行 |
| `agent.thinking` | Agent 正在思考 |
| `agent.message` | Agent 输出消息 |
| `agent.tool_call` | Agent 调用工具 |
| `task.completed` | 任务完成 |
| `task.failed` | 任务失败 |
| `task.cancelled` | 任务取消 |

### 系统事件类型

| 事件类型 | 说明 |
|---------|------|
| `task.created` | 新任务创建 |
| `agent.created` | 新 Agent 创建 |
| `session.created` | 新会话创建 |
| `batch.created` | 新批处理创建 |
| `alert` | 系统告警 |

## 任务操作

### 取消任务

```json
{
  "type": "task.action",
  "payload": {
    "task_id": "task-123",
    "action": "cancel"
  }
}
```

### 追加对话轮次

```json
{
  "type": "task.action",
  "payload": {
    "task_id": "task-123",
    "action": "append_turn",
    "data": "请继续分析..."
  }
}
```

## 心跳

客户端应定期发送心跳:

```json
{"type": "ping"}
```

响应:
```json
{"type": "pong"}
```

## JavaScript 客户端示例

```javascript
class AgentBoxGateway {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.handlers = new Map();
    this.pendingRequests = new Map();
    this.msgId = 0;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.auth().then(resolve).catch(reject);
      };

      this.ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        this.handleMessage(msg);
      };

      this.ws.onerror = reject;
    });
  }

  auth() {
    return this.send('auth', { token: this.token });
  }

  subscribe(channel, topics = []) {
    return this.send('subscribe', { channel, topics });
  }

  unsubscribe(channel, topics = []) {
    return this.send('unsubscribe', { channel, topics });
  }

  cancelTask(taskId) {
    return this.send('task.action', {
      task_id: taskId,
      action: 'cancel'
    });
  }

  appendTurn(taskId, prompt) {
    return this.send('task.action', {
      task_id: taskId,
      action: 'append_turn',
      data: prompt
    });
  }

  on(eventType, handler) {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, []);
    }
    this.handlers.get(eventType).push(handler);
  }

  send(type, payload) {
    return new Promise((resolve, reject) => {
      const id = `msg-${++this.msgId}`;
      const msg = { id, type, payload, timestamp: Date.now() };

      this.pendingRequests.set(id, { resolve, reject });
      this.ws.send(JSON.stringify(msg));

      // 超时处理
      setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error('Request timeout'));
        }
      }, 30000);
    });
  }

  handleMessage(msg) {
    // 处理请求响应
    if (msg.id && this.pendingRequests.has(msg.id)) {
      const { resolve, reject } = this.pendingRequests.get(msg.id);
      this.pendingRequests.delete(msg.id);

      if (msg.type === 'error') {
        reject(new Error(msg.payload.message));
      } else {
        resolve(msg.payload);
      }
      return;
    }

    // 处理事件推送
    if (msg.type === 'event') {
      const { event_type, data } = msg.payload;
      const handlers = this.handlers.get(event_type) || [];
      handlers.forEach(h => h(data, msg.payload));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// 使用示例
const gateway = new AgentBoxGateway(
  'ws://localhost:18080/api/v1/gateway/ws',
  'ab_xxx...'
);

await gateway.connect();
await gateway.subscribe('task', ['task-123']);

gateway.on('agent.message', (data) => {
  console.log('Agent:', data.content);
});

gateway.on('task.completed', (data) => {
  console.log('Task completed:', data.result);
});
```

## 架构说明

```
┌──────────────────────────────────────────────────────────────┐
│                         Clients                               │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐         │
│  │ Web App │  │ Mobile  │  │   CLI   │  │  Bot    │         │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘         │
└───────┼────────────┼────────────┼────────────┼───────────────┘
        │            │            │            │
        └────────────┴─────┬──────┴────────────┘
                           │ WebSocket
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                     WebSocket Gateway                         │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Client Manager  │  Subscription Manager  │  Auth      │  │
│  └────────────────────────────────────────────────────────┘  │
└───────────────────────────┬──────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Task Manager  │   │Session Manager│   │ System Events │
│               │   │               │   │               │
│ - Subscribe   │   │ - Execute     │   │ - Alerts      │
│ - Cancel      │   │ - Stream      │   │ - Stats       │
│ - AppendTurn  │   │               │   │               │
└───────────────┘   └───────────────┘   └───────────────┘
```

## 与 SSE 的对比

| 特性 | SSE (当前) | WebSocket Gateway |
|------|-----------|-------------------|
| 通信方向 | 单向 (服务端→客户端) | 双向 |
| 连接数 | 每个任务一个连接 | 所有任务共享一个连接 |
| 任务操作 | 需要 HTTP API | 直接通过 WS 发送 |
| 多端同步 | 不支持 | 支持（同用户多设备） |
| 系统事件 | 不支持 | 支持 |
| 资源消耗 | 高（多连接） | 低（单连接） |
