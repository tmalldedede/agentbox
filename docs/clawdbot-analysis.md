# Clawdbot 项目分析报告

> 分析时间: 2026-01-25
> 项目地址: https://github.com/clawdbot/clawdbot

## 项目概述

**Clawdbot** 是一个功能完整的个人 AI 助手平台（TypeScript/Node.js）：
- **2463** 个 TypeScript 文件
- **52** 个预构建技能
- **28** 个扩展
- **13+** 个通讯平台集成（WhatsApp、Telegram、Slack、Discord 等）
- **版本**: 2026.1.24-2
- **许可证**: MIT

## 核心架构

```
客户端 (macOS/iOS/Android/CLI/WebChat)
          ↓
    Gateway WebSocket (单一控制中心)
   ↙  ↓  ↙  ↓  ↙  ↓  ↙  ↓
通道  Sessions  Tools  Events  Cron
```

### 1. Gateway（中央控制平面）

位置：`/src/gateway/`

- WebSocket 服务器（默认 127.0.0.1:18789）
- 单一控制平面管理所有会话、通道、工具和事件
- 协议版本控制
- 设备认证和 TLS 指纹验证

### 2. Agent 运行时

位置：`/src/agents/`

- 基于 Pi 框架
- RPC 模式：从 Gateway 单独进程运行
- 系统提示动态构建
- 支持多家 LLM 厂商

### 3. 多通道架构

位置：`/src/channels/`

支持的通道：
- WhatsApp（Baileys 库）
- Telegram（grammY SDK）
- Slack（Bolt 框架）
- Discord（discord.js）
- Google Chat
- Signal
- iMessage（macOS）
- Microsoft Teams
- Matrix
- Zalo

### 4. 会话与路由系统

- **会话模式**: main（直聊）vs group（群组隔离）
- **激活模式**: always vs mention（群组中需要提及）
- **队列模式**: 顺序处理、并行处理、忽略重复
- **安全沙箱**: 非 main 会话可运行在 Docker 容器中

## 技能系统

位置：`/skills/` 目录，包含 **52 个预构建技能**

### 技能分类

| 类型 | 示例 |
|------|------|
| 通讯工具 | github, imsg, 1password, discord |
| 生产力 | apple-reminders, bear-notes, blogwatcher |
| 媒体 | nano-pdf, camsnap, gifgrep |
| 系统 | eightctl, goplaces |
| AI/编码 | coding-agent, gemini |

### 技能文件结构

```
skills/{name}/
├── SKILL.md              # 主文档（<500 行）
├── SKILL.json            # 元数据（可选）
├── scripts/              # 工具脚本（可选）
└── references/           # 参考文档（可选）
```

## 扩展系统

位置：`/extensions/` 目录，包含 **28 个扩展**

### Plugin SDK 接口

```typescript
export type ClawdbotPluginApi = {
  registerTool(name: string, factory: ToolFactory): void;
  registerHook(name: string, handler: HookHandler): void;
  registerHttpRoute(method: string, path: string, handler: Handler): void;
  registerAuthProvider(kind: AuthKind, handler: AuthHandler): void;
};
```

## 可借鉴功能

### 优先级高 ⭐⭐⭐

| 功能 | 说明 |
|------|------|
| 多通道集成 | WhatsApp/Telegram/Slack/飞书 统一收件箱 |
| 会话路由系统 | main/group 隔离、mention 激活模式 |
| Plugin SDK | 工具工厂、钩子系统、认证提供者 |

### 优先级中 ⭐⭐

| 功能 | 说明 |
|------|------|
| 跨会话协调 | sessions_list/history/send 工具 |
| Cron + Webhooks | 定时任务、外部触发 |
| 媒体处理管道 | 图像/视频/音频统一处理 |

### 优先级低 ⭐

| 功能 | 说明 |
|------|------|
| Browser 控制 | CDP 驱动的浏览器自动化 |
| Canvas/A2UI | 代理驱动的可视工作区 |
| 内存系统 | 向量数据库集成 |

## 实施路线

### Phase 1 (短期)
- Plugin SDK 基础架构
- Cron 定时任务
- Webhook 触发器

### Phase 2 (中期)
- 多通道适配器（飞书优先）
- 跨会话协调工具
- 内存/向量数据库

### Phase 3 (长期)
- Browser 控制
- 语音对话
- Canvas 可视工作区

## 技术亮点

### 消息协议设计

```typescript
type RequestFrame = {
  method: string;
  params: Record<string, unknown>;
  id?: string;  // 配对响应
};

type EventFrame = {
  method: string;
  params: unknown;
  seq: number;  // 序列化
};
```

### 设备认证

- 设备生成 Ed25519 密钥对
- 首次连接时签署有效载荷
- Gateway 验证设备指纹

### 系统提示动态生成

- 基础：身份行
- 可选：用户、时区、技能、内存
- 模式：full/minimal/none
