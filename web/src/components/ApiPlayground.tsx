import { useState, useEffect } from 'react'
import { Play, Copy, Check, Loader2, Code, Terminal } from 'lucide-react'

interface Profile {
  id: string
  name: string
  adapter: string
  is_public: boolean
}

interface TaskResult {
  id: string
  status: string
  result?: {
    text: string
    summary?: string
  }
}

export default function ApiPlayground() {
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [selectedProfile, setSelectedProfile] = useState('')
  const [prompt, setPrompt] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TaskResult | null>(null)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState<'curl' | 'python' | null>(null)
  const [activeTab, setActiveTab] = useState<'result' | 'curl' | 'python'>('result')

  // 获取 public profiles
  useEffect(() => {
    fetch('/api/v1/profiles')
      .then(res => res.json())
      .then(data => {
        const publicProfiles = (data.data || []).filter((p: Profile) => p.is_public)
        setProfiles(publicProfiles)
        if (publicProfiles.length > 0) {
          setSelectedProfile(publicProfiles[0].id)
        }
      })
      .catch(err => console.error('Failed to fetch profiles:', err))
  }, [])

  // 执行任务
  const handleExecute = async () => {
    if (!selectedProfile || !prompt.trim()) {
      setError('请选择 Profile 并输入 Prompt')
      return
    }

    setLoading(true)
    setError('')
    setResult(null)

    try {
      // 1. 创建 Session
      const sessionRes = await fetch('/api/v1/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent: profiles.find(p => p.id === selectedProfile)?.adapter || 'codex',
          profile_id: selectedProfile,
          workspace: `/tmp/playground-${Date.now()}`
        })
      })
      const sessionData = await sessionRes.json()
      if (sessionData.code !== 0) {
        throw new Error(sessionData.message)
      }
      const sessionId = sessionData.data.id

      // 2. 创建 Task
      const taskRes = await fetch('/api/v1/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sessionId,
          profile_id: selectedProfile,
          prompt: prompt
        })
      })
      const taskData = await taskRes.json()
      if (taskData.code !== 0) {
        throw new Error(taskData.message)
      }
      const taskId = taskData.data.id

      // 3. 轮询等待结果
      let attempts = 0
      const maxAttempts = 60 // 最多等待 2 分钟
      while (attempts < maxAttempts) {
        await new Promise(resolve => setTimeout(resolve, 2000))
        const statusRes = await fetch(`/api/v1/tasks/${taskId}`)
        const statusData = await statusRes.json()

        if (statusData.data.status === 'completed' || statusData.data.status === 'failed') {
          setResult(statusData.data)
          break
        }
        attempts++
      }

      if (attempts >= maxAttempts) {
        setError('任务执行超时')
      }
    } catch (err: any) {
      setError(err.message || '执行失败')
    } finally {
      setLoading(false)
    }
  }

  // 生成 cURL 命令
  const generateCurl = () => {
    const baseUrl = window.location.origin
    return `# 1. 创建 Session
curl -X POST ${baseUrl}/api/v1/sessions \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent": "${profiles.find(p => p.id === selectedProfile)?.adapter || 'codex'}",
    "profile_id": "${selectedProfile}",
    "workspace": "/tmp/my-task"
  }'

# 返回 session_id，如: "abc123"

# 2. 创建 Task
curl -X POST ${baseUrl}/api/v1/tasks \\
  -H "Content-Type: application/json" \\
  -d '{
    "session_id": "SESSION_ID",
    "profile_id": "${selectedProfile}",
    "prompt": "${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}"
  }'

# 返回 task_id，如: "task-xyz789"

# 3. 查询结果
curl ${baseUrl}/api/v1/tasks/TASK_ID`
  }

  // 生成 Python 代码
  const generatePython = () => {
    const baseUrl = window.location.origin
    return `import requests
import time

BASE_URL = "${baseUrl}"

def run_agent(prompt: str, profile_id: str = "${selectedProfile}"):
    """执行 Agent 任务并返回结果"""

    # 1. 创建 Session
    session_resp = requests.post(f"{BASE_URL}/api/v1/sessions", json={
        "agent": "${profiles.find(p => p.id === selectedProfile)?.adapter || 'codex'}",
        "profile_id": profile_id,
        "workspace": f"/tmp/task-{int(time.time())}"
    })
    session_id = session_resp.json()["data"]["id"]

    # 2. 创建 Task
    task_resp = requests.post(f"{BASE_URL}/api/v1/tasks", json={
        "session_id": session_id,
        "profile_id": profile_id,
        "prompt": prompt
    })
    task_id = task_resp.json()["data"]["id"]

    # 3. 等待完成
    while True:
        status_resp = requests.get(f"{BASE_URL}/api/v1/tasks/{task_id}")
        status = status_resp.json()["data"]["status"]
        if status in ["completed", "failed"]:
            return status_resp.json()["data"]
        time.sleep(2)

# 使用示例
result = run_agent("""${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}""")
print(result["result"]["text"])`
  }

  // 复制到剪贴板
  const copyToClipboard = (type: 'curl' | 'python') => {
    const text = type === 'curl' ? generateCurl() : generatePython()
    navigator.clipboard.writeText(text)
    setCopied(type)
    setTimeout(() => setCopied(null), 2000)
  }

  // 清理结果文本中的控制字符
  const cleanOutput = (text: string) => {
    return text.replace(/[\x00-\x1F\x7F]/g, '').trim()
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">API Playground</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          在线测试 AgentBox API，快速体验 AI Agent 能力
        </p>
      </div>

      {/* 配置区域 */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="space-y-4">
          {/* Profile 选择 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              选择 Profile
            </label>
            <select
              value={selectedProfile}
              onChange={(e) => setSelectedProfile(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              {profiles.length === 0 ? (
                <option value="">无可用 Profile (需设置 is_public=true)</option>
              ) : (
                profiles.map(p => (
                  <option key={p.id} value={p.id}>
                    {p.name} ({p.adapter})
                  </option>
                ))
              )}
            </select>
          </div>

          {/* Prompt 输入 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Prompt
            </label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="输入你想让 AI Agent 执行的任务..."
              rows={4}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent
                       placeholder-gray-400"
            />
          </div>

          {/* 执行按钮 */}
          <div className="flex items-center gap-4">
            <button
              onClick={handleExecute}
              disabled={loading || !selectedProfile || !prompt.trim()}
              className="flex items-center gap-2 px-6 py-2.5 bg-blue-600 text-white rounded-lg
                       hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed
                       transition-colors font-medium"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  执行中...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" />
                  执行
                </>
              )}
            </button>

            {loading && (
              <span className="text-sm text-gray-500">
                任务执行中，请稍候...
              </span>
            )}
          </div>

          {/* 错误提示 */}
          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg">
              {error}
            </div>
          )}
        </div>
      </div>

      {/* 结果区域 */}
      {(result || prompt) && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
          {/* Tab 切换 */}
          <div className="flex border-b border-gray-200 dark:border-gray-700">
            <button
              onClick={() => setActiveTab('result')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors
                ${activeTab === 'result'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Terminal className="w-4 h-4 inline mr-2" />
              执行结果
            </button>
            <button
              onClick={() => setActiveTab('curl')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors
                ${activeTab === 'curl'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Code className="w-4 h-4 inline mr-2" />
              cURL
            </button>
            <button
              onClick={() => setActiveTab('python')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors
                ${activeTab === 'python'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Code className="w-4 h-4 inline mr-2" />
              Python
            </button>
          </div>

          {/* Tab 内容 */}
          <div className="p-4">
            {activeTab === 'result' && (
              <div>
                {result ? (
                  <div className="space-y-3">
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-1 text-xs rounded ${
                        result.status === 'completed'
                          ? 'bg-green-100 text-green-700'
                          : 'bg-red-100 text-red-700'
                      }`}>
                        {result.status}
                      </span>
                      <span className="text-sm text-gray-500">Task ID: {result.id}</span>
                    </div>
                    <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg overflow-x-auto text-sm whitespace-pre-wrap">
                      {result.result?.text ? cleanOutput(result.result.text) : '无输出'}
                    </pre>
                  </div>
                ) : (
                  <div className="text-gray-500 text-center py-8">
                    点击「执行」按钮运行任务
                  </div>
                )}
              </div>
            )}

            {activeTab === 'curl' && (
              <div className="relative">
                <button
                  onClick={() => copyToClipboard('curl')}
                  className="absolute top-2 right-2 p-2 text-gray-400 hover:text-white
                           bg-gray-700 rounded transition-colors"
                >
                  {copied === 'curl' ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                </button>
                <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg overflow-x-auto text-sm">
                  {generateCurl()}
                </pre>
              </div>
            )}

            {activeTab === 'python' && (
              <div className="relative">
                <button
                  onClick={() => copyToClipboard('python')}
                  className="absolute top-2 right-2 p-2 text-gray-400 hover:text-white
                           bg-gray-700 rounded transition-colors"
                >
                  {copied === 'python' ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                </button>
                <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg overflow-x-auto text-sm">
                  {generatePython()}
                </pre>
              </div>
            )}
          </div>
        </div>
      )}

      {/* API 说明 */}
      <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
        <h3 className="font-medium text-blue-900 dark:text-blue-300 mb-2">API 调用流程</h3>
        <ol className="list-decimal list-inside text-sm text-blue-800 dark:text-blue-400 space-y-1">
          <li><code>POST /api/v1/sessions</code> - 创建会话（启动容器）</li>
          <li><code>POST /api/v1/tasks</code> - 创建任务</li>
          <li><code>GET /api/v1/tasks/:id</code> - 查询结果（轮询直到完成）</li>
        </ol>
      </div>
    </div>
  )
}
