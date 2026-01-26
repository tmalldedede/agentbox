# AgentBox 项目说明

## 项目概述

AgentBox 是一个 Agent 执行平台，采用 Go 后端 + TypeScript 前端（前后端分离）架构。

## 参考项目

| 项目 | 路径 | 说明 |
|------|------|------|
| **clawdbot** | `/Users/sky2/pr/clawdbot` |  |

## 开发命令

### 后端（Go）
```bash
make build    # 构建
make run      # 运行
make dev      # 开发模式（热重载）
make test     # 测试
```

### 前端（TypeScript + Vite）
```bash
cd web
pnpm dev      # 开发模式
pnpm build    # 构建
pnpm preview  # 预览
```

## 目录结构

```
agentbox/
├── cmd/agentbox/          # 后端入口
├── internal/              # 后端业务逻辑
│   ├── api/               # API 路由
│   ├── database/          # 数据库模型
│   └── task/              # 任务执行
├── web/                   # 前端代码
│   ├── src/
│   │   ├── components/    # 通用组件
│   │   ├── features/      # 功能模块
│   │   │   ├── chat/      # Chat 界面
│   │   │   ├── providers/ # Provider 管理
│   │   │   └── ...
│   │   ├── hooks/         # 自定义 hooks
│   │   ├── routes/        # 路由页面
│   │   ├── services/      # API 服务
│   │   ├── stores/        # Zustand stores
│   │   └── types/         # TypeScript 类型
│   └── ...
└── ...
```
