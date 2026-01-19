# AgentBox 前端改进 - 快速开始

## 安装依赖

```bash
cd web
npm install
```

这将安装所有新增的依赖：
- `@tanstack/react-query` - 数据管理
- `@tanstack/react-query-devtools` - 开发工具
- `sonner` - Toast 通知
- `prettier` - 代码格式化
- `eslint-config-prettier` - ESLint 和 Prettier 集成
- `eslint-plugin-react` - React ESLint 规则

## 使用指南

### 1. 代码格式化

```bash
# 格式化所有代码
npm run format

# 检查格式（不修改）
npm run format:check
```

### 2. 代码检查

```bash
# 运行 ESLint
npm run lint

# 自动修复问题
npm run lint:fix
```

### 3. 开发

```bash
npm run dev
```

打开浏览器访问 http://localhost:5173

你会看到：
- ✅ React Query Devtools（右下角浮动按钮）
- ✅ 改进的错误处理
- ✅ Toast 通知

### 4. 使用新的 Hooks

#### 示例 1：获取会话列表

```typescript
import { useSessions } from './hooks'

function SessionsPage() {
  const { data: sessions, isLoading, error } = useSessions()

  if (isLoading) return <div>加载中...</div>
  if (error) return <div>错误: {error.message}</div>

  return (
    <div>
      {sessions?.map(session => (
        <div key={session.id}>{session.id}</div>
      ))}
    </div>
  )
}
```

#### 示例 2：创建会话

```typescript
import { useCreateSession } from './hooks'
import { toast } from 'sonner'

function CreateButton() {
  const createSession = useCreateSession()

  const handleCreate = async () => {
    try {
      await createSession.mutateAsync({
        agent: 'claude-code',
        profile_id: 'default',
      })
      // 成功后自动显示 toast（已在 hook 中处理）
    } catch (error) {
      // 错误也会自动显示 toast
    }
  }

  return (
    <button
      onClick={handleCreate}
      disabled={createSession.isPending}
    >
      {createSession.isPending ? '创建中...' : '创建会话'}
    </button>
  )
}
```

#### 示例 3：显示 Toast

```typescript
import { toast } from 'sonner'

// 成功提示
toast.success('操作成功')

// 错误提示
toast.error('操作失败')

// 信息提示
toast.info('这是一条信息')

// 警告提示
toast.warning('这是一条警告')

// 带描述的提示
toast.success('操作成功', {
  description: '会话已创建',
})

// 自定义持续时间
toast.success('操作成功', {
  duration: 5000, // 5 秒
})
```

### 5. 使用错误处理

```typescript
import { getErrorMessage } from './lib/errors'
import { toast } from 'sonner'

try {
  await someApiCall()
} catch (error) {
  // 统一的错误消息提取
  const message = getErrorMessage(error)
  toast.error(`操作失败: ${message}`)
}
```

### 6. 使用增强的 API 客户端

```typescript
import { get, post, put, del } from './lib/api-client'

// GET 请求
const user = await get<User>('/users/123')

// POST 请求
const newUser = await post<User>('/users', {
  name: 'John',
  email: 'john@example.com',
})

// PUT 请求
const updated = await put<User>('/users/123', {
  name: 'John Doe',
})

// DELETE 请求
await del('/users/123')

// 自定义超时
const data = await get<Data>('/slow-endpoint', {
  timeout: 60000, // 60 秒
})
```

## 迁移现有组件

### 步骤 1: 移除手动状态管理

```typescript
// ❌ 旧代码
const [sessions, setSessions] = useState<Session[]>([])
const [loading, setLoading] = useState(true)
const [error, setError] = useState<string | null>(null)

useEffect(() => {
  const fetchData = async () => {
    try {
      setLoading(true)
      const data = await api.listSessions()
      setSessions(data || [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
    } finally {
      setLoading(false)
    }
  }

  fetchData()
  const interval = setInterval(fetchData, 5000)
  return () => clearInterval(interval)
}, [])

// ✅ 新代码
const { data: sessions, isLoading, error } = useSessions()
```

### 步骤 2: 移除手动错误处理

```typescript
// ❌ 旧代码
const handleDelete = async (id: string) => {
  try {
    await api.deleteSession(id)
    // 手动刷新列表
    await fetchData()
  } catch (err) {
    alert('删除失败')
  }
}

// ✅ 新代码
const deleteSession = useDeleteSession()

const handleDelete = async (id: string) => {
  // Toast 和缓存刷新都自动处理
  await deleteSession.mutateAsync(id)
}
```

### 步骤 3: 添加 Error Boundary

```typescript
// ❌ 旧代码
function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
      </Routes>
    </BrowserRouter>
  )
}

// ✅ 新代码
import { ErrorBoundary } from './components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Dashboard />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}
```

## 调试工具

### React Query Devtools

打开开发服务器后，你会在右下角看到一个浮动按钮。点击它可以：

- 📊 查看所有查询状态
- 🔄 手动触发重新获取
- 🗑️ 清除缓存
- ⏱️ 查看查询时间线
- 🔍 检查查询详情

### 浏览器开发工具

打开 Console 查看：
- ❌ 捕获的错误（来自 ErrorBoundary）
- 📝 网络请求日志
- ⚠️ ESLint 警告

## 常见问题

### Q: 安装依赖失败？

```bash
# 清理缓存
rm -rf node_modules package-lock.json
npm install
```

### Q: ESLint 报错太多？

```bash
# 自动修复大部分问题
npm run lint:fix

# 格式化代码
npm run format
```

### Q: 如何禁用某个 ESLint 规则？

在 `.eslintrc.json` 中修改：

```json
{
  "rules": {
    "规则名称": "off"
  }
}
```

### Q: React Query 不工作？

确保：
1. ✅ 已安装依赖 `npm install`
2. ✅ `main.tsx` 中已添加 `QueryClientProvider`
3. ✅ 使用正确的 hooks（如 `useSessions()`）

### Q: Toast 不显示？

确保：
1. ✅ 已在 `main.tsx` 中添加 `<Toaster />`
2. ✅ 正确导入 `import { toast } from 'sonner'`

## 下一步

查看 `IMPROVEMENTS.md` 了解：
- 📋 完整的改进列表
- 📈 性能提升数据
- 🗺️ 后续优化计划
- 📚 更多示例代码

## 需要帮助？

- 📖 阅读 `IMPROVEMENTS.md`
- 💬 查看源代码注释
- 🔍 使用 React Query Devtools 调试
