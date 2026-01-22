# AgentBox 详细实施计划

> 基于项目诊断报告制定，目标：将 AgentBox 从"本地可用"推进到"生产就绪"

## 阶段概览

```
Week 1-2     Week 3-4     Month 2      Month 3+
───────────────────────────────────────────────────
   P0           P1           P2           P3
 立即修复     短期完善     中期增强     长期规划
───────────────────────────────────────────────────
• 数据持久化  • Profile UI  • 认证授权   • 计费系统
• 依赖统一    • Demo 演示   • 多租户     • K8s 部署
• 核心测试    • CI/CD       • 历史回放   • 更多 Agent
```

---

## P0 阶段：立即修复（Week 1-2）

### 任务 1.1：数据持久化

**问题**：当前所有数据 in-memory 存储，重启丢失

**方案**：SQLite（开发）+ PostgreSQL（生产）

**实施步骤**：

```
1.1.1 定义数据库 Schema
      ├── profiles 表
      ├── mcp_servers 表
      ├── skills 表
      ├── credentials 表
      ├── sessions 表
      ├── tasks 表
      └── executions 表

1.1.2 引入数据库层
      ├── 添加 GORM 依赖
      ├── 创建 internal/database/ 模块
      ├── 实现 Repository 接口
      └── 迁移脚本

1.1.3 改造现有 Manager
      ├── ProfileManager → ProfileRepository
      ├── MCPServerManager → MCPServerRepository
      ├── SkillManager → SkillRepository
      ├── CredentialManager → CredentialRepository
      └── SessionManager → SessionRepository

1.1.4 配置管理
      ├── config.yaml 添加数据库配置
      ├── 支持 SQLite/PostgreSQL 切换
      └── 自动迁移
```

**预估时间**：3-4 天

**验收标准**：
- [ ] 重启后数据不丢失
- [ ] 支持 SQLite 和 PostgreSQL
- [ ] 迁移脚本可重复执行

---

### 任务 1.2：前端依赖统一

**问题**：pnpm-lock.yaml 与 package-lock.json 共存导致冲突

**方案**：统一使用 pnpm

**实施步骤**：

```
1.2.1 清理冲突文件
      ├── 删除 package-lock.json
      ├── 删除 node_modules/
      └── 清理 .pnpm-store

1.2.2 锁定包管理器
      ├── 添加 .npmrc: engine-strict=true
      ├── package.json 添加 "packageManager": "pnpm@9.x"
      └── 更新 README 安装说明

1.2.3 验证依赖
      ├── pnpm install
      ├── pnpm run build
      └── pnpm run dev
```

**预估时间**：0.5 天

**验收标准**：
- [ ] 只有 pnpm-lock.yaml，无 package-lock.json
- [ ] `pnpm install && pnpm run build` 成功
- [ ] 前端正常启动

---

### 任务 1.3：核心测试补充

**问题**：测试覆盖不足，API 变更可能引入回归

**方案**：补充单元测试 + 集成测试

**实施步骤**：

```
1.3.1 后端单元测试
      ├── internal/profile/ 测试
      │   ├── profile_test.go (Validate, Clone, Merge)
      │   └── manager_test.go (CRUD, 继承解析)
      ├── internal/mcp/ 测试
      ├── internal/skill/ 测试
      ├── internal/credential/ 测试
      └── internal/session/ 测试

1.3.2 API 集成测试
      ├── 扩展 tests/integration/api.hurl
      │   ├── Profile 完整 CRUD
      │   ├── MCP Server 完整 CRUD
      │   ├── Skill 完整 CRUD
      │   ├── Session 创建和执行
      │   └── 错误场景测试
      └── 添加测试数据 fixtures

1.3.3 覆盖率报告
      ├── 配置 codecov
      └── 目标覆盖率 > 60%
```

**预估时间**：2-3 天

**验收标准**：
- [ ] `make test` 全部通过
- [ ] `make test-coverage` > 60%
- [ ] Hurl 集成测试覆盖核心 API

---

### 任务 1.4：前端代码整合

**问题**：PR #2 合并后代码风格不一致，部分组件重复

**方案**：统一组件结构和代码风格

**实施步骤**：

```
1.4.1 目录结构规范化
      web/src/
      ├── components/          # 通用组件
      │   ├── ui/              # shadcn/ui 组件
      │   └── layout/          # 布局组件
      ├── features/            # 功能模块
      │   ├── profiles/
      │   ├── mcp-servers/
      │   ├── skills/
      │   ├── credentials/
      │   ├── sessions/
      │   ├── agents/
      │   └── history/
      ├── hooks/               # 自定义 Hooks
      ├── services/            # API 服务
      ├── types/               # 类型定义
      └── routes/              # TanStack Router

1.4.2 删除重复/废弃组件
      ├── 检查 components/ vs features/ 重复
      ├── 统一 List/Detail/Form 模式
      └── 更新路由引用

1.4.3 代码风格统一
      ├── ESLint 规则检查
      ├── Prettier 格式化
      └── TypeScript 严格模式
```

**预估时间**：2 天

**验收标准**：
- [ ] `pnpm run lint` 无错误
- [ ] `pnpm run build` 成功
- [ ] 无重复组件文件

---

## P1 阶段：短期完善（Week 3-4）

### 任务 2.1：Profile 编辑器 UI

**问题**：核心差异化特性缺少可视化编辑

**方案**：实现完整的 Profile 创建/编辑界面

**功能设计**：

```
Profile 编辑器页面
├── 基础信息
│   ├── 名称、描述、图标
│   ├── Adapter 选择 (Claude Code / Codex)
│   └── 继承自 (extends)
├── 模型配置
│   ├── 模型名称
│   ├── Provider / BaseURL
│   ├── Credential 选择
│   └── 高级参数 (timeout, max_tokens)
├── MCP 服务器
│   ├── 已添加列表
│   ├── 从库中选择
│   └── 快速创建
├── Skills
│   ├── 已添加列表
│   ├── 从库中选择
│   └── Skill Store 入口
├── 权限配置
│   ├── Claude Code: mode, tools, dangerously-skip
│   └── Codex: sandbox_mode, approval_policy
├── 资源限制
│   ├── CPU / Memory / Disk
│   ├── Max Turns / Timeout
│   └── Budget
└── 系统提示词
    ├── System Prompt
    ├── Append System Prompt
    └── Developer Instructions
```

**实施步骤**：

```
2.1.1 Profile Form 组件
      ├── ProfileForm.tsx (主表单)
      ├── ModelConfigSection.tsx
      ├── MCPServerSelector.tsx
      ├── SkillSelector.tsx
      ├── PermissionsSection.tsx
      └── ResourcesSection.tsx

2.1.2 Profile Create/Edit 页面
      ├── /profiles/new
      ├── /profiles/:id/edit
      └── 表单验证

2.1.3 继承预览
      ├── 显示解析后的完整配置
      └── 高亮覆盖的字段
```

**预估时间**：3-4 天

**验收标准**：
- [ ] 可视化创建 Profile
- [ ] 编辑现有 Profile
- [ ] 继承关系正确解析

---

### 任务 2.2：端到端 Demo

**问题**：缺少开箱即用的演示

**方案**：预置可运行的示例任务

**实施步骤**：

```
2.2.1 预置 Demo Profile
      ├── demo-code-review (代码审查)
      ├── demo-doc-generator (文档生成)
      └── demo-security-scan (安全扫描)

2.2.2 示例仓库
      ├── 准备测试用的 Git 仓库
      └── 或使用公开仓库 URL

2.2.3 一键演示
      ├── Dashboard 添加 "Try Demo" 按钮
      ├── 自动创建 Session
      ├── 执行预置 Prompt
      └── 展示执行结果

2.2.4 Demo 视频/GIF
      └── 录制完整流程
```

**预估时间**：2 天

**验收标准**：
- [ ] 新用户可一键体验核心功能
- [ ] Demo 在 30 秒内完成
- [ ] 有清晰的结果展示

---

### 任务 2.3：Docker 镜像优化

**问题**：Agent 镜像构建不完善

**方案**：优化 Dockerfile，预构建常用镜像

**实施步骤**：

```
2.3.1 Claude Code 镜像
      docker/agent/Dockerfile.claude-code
      ├── 基础镜像: node:20-slim
      ├── 安装 Claude Code CLI
      ├── 预配置环境变量
      ├── 非 root 用户
      └── 健康检查

2.3.2 Codex 镜像
      docker/agent/Dockerfile.codex
      ├── 基础镜像: rust:slim (或预编译)
      ├── 安装 Codex CLI
      ├── 预配置 config.toml
      └── 健康检查

2.3.3 多架构构建
      ├── linux/amd64
      ├── linux/arm64
      └── 发布到 Docker Hub / GHCR

2.3.4 镜像管理 UI
      ├── 镜像列表页面
      ├── 拉取状态显示
      └── 镜像配置
```

**预估时间**：2-3 天

**验收标准**：
- [ ] `docker pull agentbox/claude-code:latest` 可用
- [ ] `docker pull agentbox/codex:latest` 可用
- [ ] 镜像大小 < 1GB

---

### 任务 2.4：CI/CD 配置

**问题**：无自动化构建测试

**方案**：GitHub Actions

**实施步骤**：

```
2.4.1 后端 CI
      .github/workflows/backend.yml
      ├── Trigger: push/PR to main
      ├── Go setup
      ├── make lint
      ├── make test
      ├── make build
      └── 上传覆盖率

2.4.2 前端 CI
      .github/workflows/frontend.yml
      ├── Trigger: push/PR to main
      ├── pnpm setup
      ├── pnpm install
      ├── pnpm run lint
      ├── pnpm run build
      └── (可选) 截图测试

2.4.3 Docker 构建
      .github/workflows/docker.yml
      ├── Trigger: release tag
      ├── 构建多架构镜像
      └── 推送到 GHCR

2.4.4 Release 自动化
      .github/workflows/release.yml
      ├── GoReleaser
      ├── 生成 Changelog
      └── 发布二进制
```

**预估时间**：1-2 天

**验收标准**：
- [ ] PR 自动运行测试
- [ ] main 分支构建状态徽章
- [ ] Release 自动发布

---

## P2 阶段：中期增强（Month 2）

### 任务 3.1：认证授权

**方案**：JWT + 基础 RBAC

```
3.1.1 用户系统
      ├── User 模型
      ├── 注册/登录 API
      ├── JWT 生成/验证
      └── Refresh Token

3.1.2 RBAC
      ├── Role: admin, user, viewer
      ├── Permission: profile:*, session:*, etc.
      └── 中间件

3.1.3 前端集成
      ├── 登录页面
      ├── Token 存储
      ├── 请求拦截器
      └── 权限控制
```

**预估时间**：1 周

---

### 任务 3.2：多租户隔离

**方案**：Workspace 概念

```
3.2.1 Workspace 模型
      ├── workspace_id 关联所有资源
      ├── Workspace 成员管理
      └── 资源配额

3.2.2 数据隔离
      ├── 所有查询加 workspace_id 过滤
      ├── 容器隔离标签
      └── 存储目录隔离

3.2.3 UI 切换
      ├── Workspace 选择器
      └── 工作区管理页面
```

**预估时间**：1 周

---

### 任务 3.3：执行历史

**方案**：完整的任务日志和回放

```
3.3.1 日志存储
      ├── 结构化日志表
      ├── 文件附件存储
      └── 索引优化

3.3.2 历史查询
      ├── 按时间/状态/Profile 过滤
      ├── 全文搜索
      └── 导出功能

3.3.3 回放 UI
      ├── 时间线展示
      ├── 步骤详情
      └── 输出文件预览
```

**预估时间**：1 周

---

### 任务 3.4：Skill Store

**方案**：社区技能市场雏形

```
3.4.1 Skill 格式规范
      ├── SKILL.md 标准
      ├── 元数据 schema
      └── 打包格式 (.skill)

3.4.2 Store 功能
      ├── 浏览/搜索
      ├── 一键安装
      ├── 版本管理
      └── 评分/反馈

3.4.3 贡献流程
      ├── 提交 Skill
      ├── 审核机制
      └── 发布
```

**预估时间**：1-2 周

---

## P3 阶段：长期规划（Month 3+）

### 任务 4.1：计费系统

```
├── Token 计量
├── 使用统计
├── 配额管理
├── 计费规则
└── 报表导出
```

### 任务 4.2：Kubernetes 部署

```
├── Helm Chart
├── Operator (可选)
├── 自动扩缩
├── 持久化卷
└── 监控集成 (Prometheus/Grafana)
```

### 任务 4.3：更多 Agent 支持

```
├── OpenCode Adapter
├── Aider Adapter
├── 自定义 Agent SDK
└── Agent 市场
```

### 任务 4.4：Profile 市场

```
├── 公开分享
├── 订阅机制
├── 版本同步
└── 社区治理
```

---

## 依赖关系图

```
P0.1 数据持久化 ─────┬──→ P1.1 Profile 编辑器
                     │
P0.2 依赖统一 ───────┼──→ P1.4 CI/CD
                     │
P0.3 核心测试 ───────┤
                     │
P0.4 前端整合 ───────┴──→ P1.2 Demo 演示

P1.* ────────────────────→ P2.1 认证授权
                           │
                           ├──→ P2.2 多租户
                           │
                           └──→ P2.3 执行历史

P2.* ────────────────────→ P3.* 长期规划
```

---

## 里程碑

| 里程碑 | 目标日期 | 交付物 |
|--------|----------|--------|
| **M1: 稳定可用** | Week 2 | 数据持久化 + 测试 + 依赖统一 |
| **M2: 功能完整** | Week 4 | Profile UI + Demo + CI/CD |
| **M3: 企业就绪** | Month 2 | 认证 + 多租户 + 历史 |
| **M4: 生态繁荣** | Month 3+ | Skill Store + 更多 Agent |

---

## 资源估算

| 阶段 | 工作量 | 建议人力 |
|------|--------|----------|
| P0 | 8-10 人天 | 1 全栈 + 1 后端 |
| P1 | 8-10 人天 | 1 全栈 + 1 前端 |
| P2 | 15-20 人天 | 2 全栈 |
| P3 | 30+ 人天 | 团队 |

---

## 下一步行动

### 本周任务（P0 启动）

- [ ] **Day 1-2**: 数据持久化 Schema 设计 + GORM 集成
- [ ] **Day 2**: 前端依赖统一（pnpm）
- [ ] **Day 3-4**: 数据库 Repository 实现
- [ ] **Day 4-5**: 核心测试补充
- [ ] **Day 5**: 前端代码整合

### 立即执行

```bash
# 1. 删除 npm 残留
cd /Users/sky2/pr/agentbox/web
rm -f package-lock.json
echo "engine-strict=true" > .npmrc

# 2. 创建数据库模块
mkdir -p internal/database
touch internal/database/database.go
touch internal/database/migrations.go

# 3. 添加 GORM 依赖
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get -u gorm.io/driver/postgres
```

---

*计划制定日期: 2026-01-22*
*版本: 1.0*
