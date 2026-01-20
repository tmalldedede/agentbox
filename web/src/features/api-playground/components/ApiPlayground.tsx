import { useState, useEffect, useRef } from 'react'
import { Play, Copy, Check, Loader2, Code, Terminal, Upload, X, FileIcon, Radio } from 'lucide-react'

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

interface UploadedFile {
  file: File
  path: string  // 上传后的路径
  uploaded: boolean
}

export default function ApiPlayground() {
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [selectedProfile, setSelectedProfile] = useState('')
  const [prompt, setPrompt] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TaskResult | null>(null)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState<'curl' | 'python' | null>(null)
  const [activeTab, setActiveTab] = useState<'result' | 'logs' | 'curl' | 'python'>('result')
  const [files, setFiles] = useState<UploadedFile[]>([])
  const [uploadProgress, setUploadProgress] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  // SSE 实时日志
  const [logs, setLogs] = useState<string[]>([])
  const [sseConnected, setSseConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const logsEndRef = useRef<HTMLDivElement>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

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

  // 自动滚动到日志底部
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  // 清理 SSE 连接
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
    }
  }, [])

  // 连接 SSE 日志流
  const connectLogStream = (sessionId: string) => {
    // 关闭之前的连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const es = new EventSource(`/api/v1/sessions/${sessionId}/logs/stream`)
    eventSourceRef.current = es

    es.addEventListener('connected', (event) => {
      setSseConnected(true)
      const data = JSON.parse(event.data)
      setLogs(prev => [...prev, `[SSE] 已连接到 Session: ${data.session_id}`])
    })

    es.onmessage = (event) => {
      if (event.data.trim()) {
        setLogs(prev => [...prev, event.data])
      }
    }

    es.addEventListener('error', (event: MessageEvent) => {
      if (event.data) {
        setLogs(prev => [...prev, `[SSE Error] ${event.data}`])
      }
    })

    es.onerror = () => {
      setSseConnected(false)
      es.close()
    }
  }

  // 断开 SSE 连接
  const disconnectLogStream = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
      setSseConnected(false)
    }
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }
  }

  // SSE 流式执行 (仅限 Codex)
  const executeViaSSE = async (sessionId: string, prompt: string): Promise<{ success: boolean; text: string; error?: string }> => {
    return new Promise((resolve) => {
      abortControllerRef.current = new AbortController()

      let resultText = ''

      fetch(`/api/v1/sessions/${sessionId}/exec/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt }),
        signal: abortControllerRef.current.signal
      }).then(async (response) => {
        if (!response.ok) {
          const err = await response.text()
          resolve({ success: false, text: '', error: err })
          return
        }

        const reader = response.body?.getReader()
        if (!reader) {
          resolve({ success: false, text: '', error: 'No response body' })
          return
        }

        const decoder = new TextDecoder()
        let buffer = ''

        setSseConnected(true)
        setLogs(prev => [...prev, `[SSE] 已连接到执行流`])

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })

          // 解析 SSE 事件
          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // 保留不完整的行

          let eventType = ''
          let eventData = ''

          for (const line of lines) {
            if (line.startsWith('event: ')) {
              eventType = line.slice(7).trim()
            } else if (line.startsWith('data: ')) {
              eventData = line.slice(6).trim()
            } else if (line === '' && eventData) {
              // 事件结束，处理
              try {
                const data = JSON.parse(eventData)

                switch (eventType) {
                  case 'connected':
                    setLogs(prev => [...prev, `[SSE] Execution ID: ${data.execution_id}`])
                    break
                  case 'execution.started':
                    setLogs(prev => [...prev, `[SSE] 执行开始: ${data.execution_id}`])
                    break
                  case 'execution.completed':
                    // 执行完成，提取最终文本
                    if (data.text) {
                      resultText = data.text // 使用最终文本
                    }
                    setLogs(prev => [...prev, `[SSE] 执行完成`])
                    break
                  case 'execution.cancelled':
                    setLogs(prev => [...prev, `[SSE] 执行已取消: ${data.error || ''}`])
                    break
                  case 'item.completed':
                    // Codex item.completed 事件，提取 agent_message 文本
                    if (data.data) {
                      try {
                        const itemData = typeof data.data === 'string' ? JSON.parse(data.data) : data.data
                        const item = itemData.item || itemData
                        if (item.type === 'agent_message' && item.text) {
                          resultText = item.text // 更新为最新的 agent_message
                          setLogs(prev => [...prev, `[Agent] ${item.text.slice(0, 500)}${item.text.length > 500 ? '...' : ''}`])
                        } else if (item.type === 'function_call') {
                          setLogs(prev => [...prev, `[Tool] ${item.name || 'function'}: ${JSON.stringify(item.arguments || {}).slice(0, 100)}...`])
                        } else if (item.type === 'function_call_output') {
                          setLogs(prev => [...prev, `[Tool Result] ${(item.output || '').slice(0, 100)}...`])
                        }
                      } catch {
                        // 如果 data.text 直接有值
                        if (data.text) {
                          resultText = data.text
                          setLogs(prev => [...prev, `[Agent] ${data.text.slice(0, 500)}${data.text.length > 500 ? '...' : ''}`])
                        }
                      }
                    } else if (data.text) {
                      resultText = data.text
                      setLogs(prev => [...prev, `[Agent] ${data.text.slice(0, 500)}${data.text.length > 500 ? '...' : ''}`])
                    }
                    break
                  case 'turn.completed':
                    setLogs(prev => [...prev, `[SSE] Turn 完成`])
                    break
                  case 'turn.failed':
                  case 'error':
                    setLogs(prev => [...prev, `[Error] ${data.error || data.message || 'Unknown error'}`])
                    break
                  case 'done':
                    setLogs(prev => [...prev, `[SSE] 流结束`])
                    break
                  default:
                    // 其他事件类型 - 简化日志
                    if (eventType.startsWith('thread.') || eventType.startsWith('turn.')) {
                      setLogs(prev => [...prev, `[${eventType}]`])
                    } else {
                      setLogs(prev => [...prev, `[${eventType}] ${JSON.stringify(data).slice(0, 150)}`])
                    }
                }
              } catch (e) {
                // 非 JSON 数据
                if (eventData.trim()) {
                  setLogs(prev => [...prev, eventData])
                }
              }

              eventType = ''
              eventData = ''
            }
          }
        }

        setSseConnected(false)
        resolve({ success: true, text: resultText })
      }).catch((err) => {
        setSseConnected(false)
        if (err.name === 'AbortError') {
          resolve({ success: false, text: '', error: '执行已取消' })
        } else {
          resolve({ success: false, text: '', error: err.message })
        }
      })
    })
  }

  // 处理文件选择
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files).map(file => ({
        file,
        path: '',
        uploaded: false
      }))
      setFiles(prev => [...prev, ...newFiles])
    }
    // 清空 input 以允许重复选择同一文件
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  // 移除文件
  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index))
  }

  // 上传文件到 Session
  const uploadFilesToSession = async (sessionId: string): Promise<string[]> => {
    const uploadedPaths: string[] = []

    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      setUploadProgress(`上传文件 (${i + 1}/${files.length}): ${file.file.name}`)

      const formData = new FormData()
      formData.append('file', file.file)

      const res = await fetch(`/api/v1/sessions/${sessionId}/files?path=/`, {
        method: 'POST',
        body: formData
      })

      if (!res.ok) {
        throw new Error(`上传文件失败: ${file.file.name}`)
      }

      const data = await res.json()
      uploadedPaths.push(data.data?.path || `/${file.file.name}`)

      // 更新文件状态
      setFiles(prev => prev.map((f, idx) =>
        idx === i ? { ...f, path: data.data?.path || `/${file.file.name}`, uploaded: true } : f
      ))
    }

    setUploadProgress('')
    return uploadedPaths
  }

  // 格式化文件大小
  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  // 执行任务
  const handleExecute = async () => {
    if (!selectedProfile || !prompt.trim()) {
      setError('请选择 Profile 并输入 Prompt')
      return
    }

    setLoading(true)
    setError('')
    setResult(null)
    setActiveTab('logs') // 切换到日志 tab
    setLogs([`[System] 开始执行任务...`]) // 立即显示日志

    // 获取当前选中的 profile 信息
    const currentProfile = profiles.find(p => p.id === selectedProfile)
    const agentType = currentProfile?.adapter || 'codex'
    const isCodex = agentType === 'codex'

    try {
      // 1. 创建 Session
      const sessionRes = await fetch('/api/v1/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent: agentType,
          profile_id: selectedProfile,
          workspace: `/tmp/playground-${Date.now()}`
        })
      })
      const sessionData = await sessionRes.json()
      if (sessionData.code !== 0) {
        throw new Error(sessionData.message)
      }
      const sessionId = sessionData.data.id
      setLogs(prev => [...prev, `[System] Session 创建成功: ${sessionId} (Agent: ${agentType})`])

      // 2. 上传文件（如果有）
      let uploadedPaths: string[] = []
      if (files.length > 0) {
        uploadedPaths = await uploadFilesToSession(sessionId)
        setLogs(prev => [...prev, `[System] 文件上传完成: ${uploadedPaths.length} 个文件`])
      }

      // 3. 构建 prompt，包含文件信息
      let finalPrompt = prompt
      if (uploadedPaths.length > 0) {
        const fileList = uploadedPaths.map(p => `/workspace${p}`).join('\n')
        finalPrompt = `已上传以下文件到工作区:\n${fileList}\n\n${prompt}`
      }

      // 4. 根据 Agent 类型选择执行方式
      if (isCodex) {
        // Codex: 使用 SSE 流式执行
        setLogs(prev => [...prev, `[System] 使用 SSE 流式执行 (Codex)`])

        const startTime = Date.now()
        const sseResult = await executeViaSSE(sessionId, finalPrompt)

        const elapsed = Math.floor((Date.now() - startTime) / 1000)
        const mins = Math.floor(elapsed / 60)
        const secs = elapsed % 60

        if (sseResult.success) {
          setResult({
            id: `sse-${Date.now()}`,
            status: 'completed',
            result: { text: sseResult.text || '执行完成 (无文本输出)' }
          })
          setLogs(prev => [...prev, `[System] SSE 执行完成 (耗时 ${mins}:${secs.toString().padStart(2, '0')})`])
        } else {
          setResult({
            id: `sse-${Date.now()}`,
            status: 'failed',
            result: { text: sseResult.error || '执行失败' }
          })
          setError(sseResult.error || '执行失败')
        }
        setActiveTab('result')
      } else {
        // 其他 Agent: 使用 Task API 轮询
        setLogs(prev => [...prev, `[System] 使用 Task API 执行 (${agentType})`])

        // 连接 SSE 日志流 (用于实时查看容器日志)
        connectLogStream(sessionId)

        // 创建 Task
        const taskRes = await fetch('/api/v1/tasks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            session_id: sessionId,
            profile_id: selectedProfile,
            prompt: finalPrompt,
            input: uploadedPaths.length > 0 ? { files: uploadedPaths } : undefined
          })
        })
        const taskData = await taskRes.json()
        if (taskData.code !== 0) {
          throw new Error(taskData.message)
        }
        const taskId = taskData.data.id
        setLogs(prev => [...prev, `[System] Task 创建成功: ${taskId}`])

        // 轮询等待结果
        let attempts = 0
        const maxAttempts = 300 // 最多等待 10 分钟 (300 * 2s)
        const startTime = Date.now()
        while (attempts < maxAttempts) {
          await new Promise(resolve => setTimeout(resolve, 2000))
          const statusRes = await fetch(`/api/v1/tasks/${taskId}`)
          const statusData = await statusRes.json()

          // 更新执行时间
          const elapsed = Math.floor((Date.now() - startTime) / 1000)
          const mins = Math.floor(elapsed / 60)
          const secs = elapsed % 60
          if (attempts % 5 === 0) { // 每 10 秒显示一次
            setLogs(prev => [...prev, `[System] 执行中... ${mins}:${secs.toString().padStart(2, '0')}`])
          }

          if (statusData.data.status === 'completed' || statusData.data.status === 'failed') {
            setResult(statusData.data)
            setLogs(prev => [...prev, `[System] Task 完成: ${statusData.data.status} (耗时 ${mins}:${secs.toString().padStart(2, '0')})`])
            setActiveTab('result') // 完成后切换到结果 tab
            break
          }
          attempts++
        }

        if (attempts >= maxAttempts) {
          setError('任务执行超时 (超过 10 分钟)')
          setLogs(prev => [...prev, `[Error] 任务执行超时，但可能仍在后台运行。Task ID: ${taskId}`])
        }
      }
    } catch (err: any) {
      setError(err.message || '执行失败')
      setLogs(prev => [...prev, `[Error] ${err.message || '执行失败'}`])
    } finally {
      setLoading(false)
      setUploadProgress('')
      disconnectLogStream() // 断开 SSE 连接
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

          {/* 文件上传 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              上传文件 <span className="text-gray-400 font-normal">(可选)</span>
            </label>

            {/* 文件列表 */}
            {files.length > 0 && (
              <div className="mb-3 space-y-2">
                {files.map((file, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700/50 rounded-lg"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <FileIcon className="w-4 h-4 text-gray-400 flex-shrink-0" />
                      <span className="text-sm text-gray-700 dark:text-gray-300 truncate">
                        {file.file.name}
                      </span>
                      <span className="text-xs text-gray-400 flex-shrink-0">
                        ({formatFileSize(file.file.size)})
                      </span>
                      {file.uploaded && (
                        <Check className="w-4 h-4 text-green-500 flex-shrink-0" />
                      )}
                    </div>
                    <button
                      onClick={() => removeFile(index)}
                      disabled={loading}
                      className="p-1 text-gray-400 hover:text-red-500 disabled:opacity-50"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* 上传按钮 */}
            <div className="flex items-center gap-3">
              <input
                ref={fileInputRef}
                type="file"
                multiple
                onChange={handleFileSelect}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={loading}
                className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600
                         rounded-lg text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700
                         hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50
                         transition-colors text-sm"
              >
                <Upload className="w-4 h-4" />
                选择文件
              </button>
              {files.length > 0 && (
                <span className="text-sm text-gray-500">
                  已选择 {files.length} 个文件
                </span>
              )}
            </div>

            {/* 上传进度 */}
            {uploadProgress && (
              <div className="mt-2 text-sm text-blue-600 dark:text-blue-400 flex items-center gap-2">
                <Loader2 className="w-4 h-4 animate-spin" />
                {uploadProgress}
              </div>
            )}

            <p className="mt-2 text-xs text-gray-400">
              文件将上传到 Session 工作区，可在 Prompt 中引用 /workspace/文件名
            </p>
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
              onClick={() => setActiveTab('logs')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors relative
                ${activeTab === 'logs'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Radio className="w-4 h-4 inline mr-2" />
              实时日志
              {sseConnected && (
                <span className="absolute top-2 right-1 w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              )}
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

            {activeTab === 'logs' && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full ${sseConnected ? 'bg-green-500' : 'bg-gray-400'}`} />
                    <span className="text-sm text-gray-500">
                      {sseConnected ? 'SSE 已连接' : 'SSE 未连接'}
                    </span>
                  </div>
                  <span className="text-xs text-gray-400">{logs.length} 条日志</span>
                </div>
                <div className="h-80 overflow-y-auto bg-gray-900 rounded-lg p-4 font-mono text-sm">
                  {logs.length === 0 ? (
                    <div className="text-gray-500 text-center py-8">
                      执行任务后将显示实时日志
                    </div>
                  ) : (
                    logs.map((log, index) => (
                      <div
                        key={index}
                        className={`py-0.5 ${
                          log.startsWith('[System]') ? 'text-blue-400' :
                          log.startsWith('[SSE]') ? 'text-green-400' :
                          log.startsWith('[Error]') ? 'text-red-400' :
                          'text-gray-100'
                        }`}
                      >
                        {log}
                      </div>
                    ))
                  )}
                  <div ref={logsEndRef} />
                </div>
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
