# Playground Agent 选择器修复验证

## 问题描述
- **症状**: API Playground 页面无法选择 agents
- **原因**: 获取 agents 列表时未携带认证 token，导致 401 错误

## 修复内容

### 1. agents 获取认证
```typescript
// 修复前：使用原生 fetch，无认证
fetch('/api/v1/agents')
  .then(res => res.json())

// 修复后：添加 Bearer token
const token = localStorage.getItem('agentbox_token')
const headers = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${token}`
}
fetch('/api/v1/agents', { headers })
```

### 2. loggedFetch 认证增强
```typescript
// 自动在所有 API 调用中添加认证 token
const token = localStorage.getItem('agentbox_token')
const headers = {
  ...(options?.headers || {}),
}
if (token && !headers['Authorization']) {
  headers['Authorization'] = `Bearer ${token}`
}
```

## 验证步骤

### 1. 启动服务
```bash
# 后端（如果未运行）
make dev

# 前端（如果未运行）
cd web && pnpm dev
```

### 2. 测试流程
1. **访问**: http://localhost:5173
2. **登录**: 
   - 用户名: `admin`
   - 密码: `admin` （默认密码）
3. **导航到 API Playground**: 侧边栏 → API Playground
4. **验证 Agent 下拉框**: 
   - ✅ 应该能看到 agents 列表
   - ✅ 可以选择不同的 agent
   - ✅ 默认选中第一个 agent

### 3. 浏览器控制台检查
按 F12 打开开发者工具，检查：

**修复前的错误**:
```
GET /api/v1/agents 401 (Unauthorized)
Failed to fetch agents: Unauthorized
```

**修复后应该看到**:
```
GET /api/v1/agents 200 OK
```

### 4. 完整功能测试
1. **选择 Agent**: 从下拉框选择一个 agent
2. **输入 Prompt**: 例如 "What is 2+2?"
3. **点击 Execute**: 
   - ✅ 应该开始执行
   - ✅ API Calls 标签页显示请求
   - ✅ Container Logs 显示日志
   - ✅ Result 标签页显示结果

## 技术细节

### 修改文件
- `web/src/features/api-playground/components/ApiPlayground.tsx`

### 影响的 API 调用
1. `GET /api/v1/agents` - 获取 agents 列表
2. `POST /api/v1/admin/sessions` - 创建 session
3. `POST /api/v1/admin/sessions/:id/files` - 上传文件
4. `POST /api/v1/tasks` - 创建 task
5. `GET /api/v1/tasks/:id` - 查询任务状态

### Git 提交
```bash
8b49ea5 fix(playground): 修复 Agent 选择器无法加载的问题
```

## 预期结果

### 修复前
- ❌ Agent 下拉框为空
- ❌ 无法执行任务
- ❌ Console 显示 401 错误

### 修复后
- ✅ Agent 下拉框显示可用 agents
- ✅ 可以正常选择和切换 agent
- ✅ Execute 功能正常工作
- ✅ 所有 API 调用携带认证 token

## 附加说明

### 认证机制
AgentBox 使用 JWT Bearer token 认证：
- Token 存储位置: `localStorage.getItem('agentbox_token')`
- 请求头格式: `Authorization: Bearer <token>`
- Token 在登录时获取并保存

### 受影响的其他页面
此修复仅影响 API Playground 页面。其他页面（如 Agents、Tasks、Providers）使用的是 `api.ts` 中的统一 API client，已经正确实现了认证。
