import { useState, useEffect, useRef } from 'react'
import { Play, Copy, Check, Loader2, Code, Terminal, Upload, X, FileIcon, Radio } from 'lucide-react'
import type { Agent } from '@/types'

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
  path: string
  uploaded: boolean
}

interface ApiPlaygroundProps {
  preselectedAgentId?: string
}

export default function ApiPlayground({ preselectedAgentId }: ApiPlaygroundProps) {
  const [agents, setAgents] = useState<Agent[]>([])
  const [selectedAgentId, setSelectedAgentId] = useState('')
  const [prompt, setPrompt] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TaskResult | null>(null)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState<'curl' | 'python' | null>(null)
  const [activeTab, setActiveTab] = useState<'result' | 'logs' | 'curl' | 'python'>('result')
  const [files, setFiles] = useState<UploadedFile[]>([])
  const [uploadProgress, setUploadProgress] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  // SSE logs
  const [logs, setLogs] = useState<string[]>([])
  const [sseConnected, setSseConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const logsEndRef = useRef<HTMLDivElement>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  // Fetch agents with public/api_key access
  useEffect(() => {
    fetch('/api/v1/agents')
      .then(res => res.json())
      .then(data => {
        const allAgents = (data.data || []) as Agent[]
        const publicAgents = allAgents.filter((a: Agent) => a.api_access === 'public' || a.api_access === 'api_key')
        const displayAgents = publicAgents.length > 0 ? publicAgents : allAgents
        setAgents(displayAgents)
        if (preselectedAgentId && allAgents.find(a => a.id === preselectedAgentId)) {
          setSelectedAgentId(preselectedAgentId)
        } else if (displayAgents.length > 0) {
          setSelectedAgentId(displayAgents[0].id)
        }
      })
      .catch(err => console.error('Failed to fetch agents:', err))
  }, [preselectedAgentId])

  // Auto-scroll logs
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  // Cleanup SSE
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
    }
  }, [])

  // Connect to SSE log stream
  const connectLogStream = (sessionId: string) => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const es = new EventSource(`/api/v1/sessions/${sessionId}/logs/stream`)
    eventSourceRef.current = es

    es.addEventListener('connected', (event) => {
      setSseConnected(true)
      const data = JSON.parse(event.data)
      setLogs(prev => [...prev, `[SSE] Connected to Session: ${data.session_id}`])
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

  // File handling
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files).map(file => ({
        file,
        path: '',
        uploaded: false
      }))
      setFiles(prev => [...prev, ...newFiles])
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index))
  }

  const uploadFilesToSession = async (sessionId: string): Promise<string[]> => {
    const uploadedPaths: string[] = []

    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      setUploadProgress(`Uploading (${i + 1}/${files.length}): ${file.file.name}`)

      const formData = new FormData()
      formData.append('file', file.file)

      const res = await fetch(`/api/v1/sessions/${sessionId}/files?path=/`, {
        method: 'POST',
        body: formData
      })

      if (!res.ok) {
        throw new Error(`Failed to upload: ${file.file.name}`)
      }

      const data = await res.json()
      uploadedPaths.push(data.data?.path || `/${file.file.name}`)

      setFiles(prev => prev.map((f, idx) =>
        idx === i ? { ...f, path: data.data?.path || `/${file.file.name}`, uploaded: true } : f
      ))
    }

    setUploadProgress('')
    return uploadedPaths
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  // Execute task using Agent Run API
  const handleExecute = async () => {
    if (!selectedAgentId || !prompt.trim()) {
      setError('Please select an Agent and enter a prompt')
      return
    }

    setLoading(true)
    setError('')
    setResult(null)
    setActiveTab('logs')
    setLogs([`[System] Starting execution...`])

    const selectedAgent = agents.find(a => a.id === selectedAgentId)

    try {
      // 1. Create Session
      const sessionRes = await fetch('/api/v1/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: selectedAgentId,
          workspace: `/tmp/playground-${Date.now()}`
        })
      })
      const sessionData = await sessionRes.json()
      if (sessionData.code !== 0) {
        throw new Error(sessionData.message)
      }
      const sessionId = sessionData.data.id
      setLogs(prev => [...prev, `[System] Session created: ${sessionId} (Agent: ${selectedAgent?.name || selectedAgentId})`])

      // 2. Upload files if any
      let uploadedPaths: string[] = []
      if (files.length > 0) {
        uploadedPaths = await uploadFilesToSession(sessionId)
        setLogs(prev => [...prev, `[System] Files uploaded: ${uploadedPaths.length}`])
      }

      // 3. Build final prompt
      let finalPrompt = prompt
      if (uploadedPaths.length > 0) {
        const fileList = uploadedPaths.map(p => `/workspace${p}`).join('\n')
        finalPrompt = `Uploaded files:\n${fileList}\n\n${prompt}`
      }

      // 4. Connect log stream and create task
      connectLogStream(sessionId)

      const taskRes = await fetch('/api/v1/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: selectedAgentId,
          prompt: finalPrompt,
          input: uploadedPaths.length > 0 ? { files: uploadedPaths } : undefined
        })
      })
      const taskData = await taskRes.json()
      if (taskData.code !== 0) {
        throw new Error(taskData.message)
      }
      const taskId = taskData.data.id
      setLogs(prev => [...prev, `[System] Task created: ${taskId}`])

      // 5. Poll for result
      let attempts = 0
      const maxAttempts = 300
      const startTime = Date.now()
      while (attempts < maxAttempts) {
        await new Promise(resolve => setTimeout(resolve, 2000))
        const statusRes = await fetch(`/api/v1/tasks/${taskId}`)
        const statusData = await statusRes.json()

        const elapsed = Math.floor((Date.now() - startTime) / 1000)
        const mins = Math.floor(elapsed / 60)
        const secs = elapsed % 60
        if (attempts % 5 === 0) {
          setLogs(prev => [...prev, `[System] Running... ${mins}:${secs.toString().padStart(2, '0')}`])
        }

        if (statusData.data.status === 'completed' || statusData.data.status === 'failed') {
          setResult(statusData.data)
          setLogs(prev => [...prev, `[System] Task ${statusData.data.status} (${mins}:${secs.toString().padStart(2, '0')})`])
          setActiveTab('result')
          break
        }
        attempts++
      }

      if (attempts >= maxAttempts) {
        setError('Task timeout (exceeded 10 minutes)')
        setLogs(prev => [...prev, `[Error] Task timeout. Task ID: ${taskId}`])
      }
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message || 'Execution failed')
      setLogs(prev => [...prev, `[Error] ${error.message || 'Execution failed'}`])
    } finally {
      setLoading(false)
      setUploadProgress('')
      disconnectLogStream()
    }
  }

  // Generate cURL command
  const generateCurl = () => {
    const baseUrl = window.location.origin
    return `# 1. Create Session
curl -X POST ${baseUrl}/api/v1/sessions \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_id": "${selectedAgentId}",
    "workspace": "/tmp/my-task"
  }'

# Returns session_id, e.g.: "abc123"

# 2. Create Task
curl -X POST ${baseUrl}/api/v1/tasks \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_id": "${selectedAgentId}",
    "prompt": "${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}"
  }'

# Returns task_id, e.g.: "task-xyz789"

# 3. Query Result
curl ${baseUrl}/api/v1/tasks/TASK_ID

# Alternative: Use Agent Run API directly
curl -X POST ${baseUrl}/api/v1/agents/${selectedAgentId}/run \\
  -H "Content-Type: application/json" \\
  -d '{
    "prompt": "${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}",
    "workspace": "/tmp/my-task"
  }'`
  }

  // Generate Python code
  const generatePython = () => {
    const baseUrl = window.location.origin
    return `import requests
import time

BASE_URL = "${baseUrl}"

def run_agent(prompt: str, agent_id: str = "${selectedAgentId}"):
    """Run an Agent task and return result"""

    # Option 1: Use Agent Run API (simplest)
    run_resp = requests.post(f"{BASE_URL}/api/v1/agents/{agent_id}/run", json={
        "prompt": prompt,
        "workspace": f"/tmp/task-{int(time.time())}"
    })
    return run_resp.json()["data"]

    # Option 2: Session + Task (more control)
    # 1. Create Session
    session_resp = requests.post(f"{BASE_URL}/api/v1/sessions", json={
        "agent_id": agent_id,
        "workspace": f"/tmp/task-{int(time.time())}"
    })
    session_id = session_resp.json()["data"]["id"]

    # 2. Create Task
    task_resp = requests.post(f"{BASE_URL}/api/v1/tasks", json={
        "agent_id": agent_id,
        "prompt": prompt
    })
    task_id = task_resp.json()["data"]["id"]

    # 3. Wait for completion
    while True:
        status_resp = requests.get(f"{BASE_URL}/api/v1/tasks/{task_id}")
        status = status_resp.json()["data"]["status"]
        if status in ["completed", "failed"]:
            return status_resp.json()["data"]
        time.sleep(2)

# Usage
result = run_agent("""${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}""")
print(result)`
  }

  const copyToClipboard = (type: 'curl' | 'python') => {
    const text = type === 'curl' ? generateCurl() : generatePython()
    navigator.clipboard.writeText(text)
    setCopied(type)
    setTimeout(() => setCopied(null), 2000)
  }

  const cleanOutput = (text: string) => {
    return text.replace(/[\x00-\x1F\x7F]/g, '').trim()
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">API Playground</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          Test the AgentBox API online
        </p>
      </div>

      {/* Configuration */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="space-y-4">
          {/* Agent Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Select Agent
            </label>
            <select
              value={selectedAgentId}
              onChange={(e) => setSelectedAgentId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              {agents.length === 0 ? (
                <option value="">No agents available</option>
              ) : (
                agents.map(a => (
                  <option key={a.id} value={a.id}>
                    {a.name} ({a.adapter})
                  </option>
                ))
              )}
            </select>
          </div>

          {/* Prompt */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Prompt
            </label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Enter the task for the AI Agent..."
              rows={4}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg
                       bg-white dark:bg-gray-700 text-gray-900 dark:text-white
                       focus:ring-2 focus:ring-blue-500 focus:border-transparent
                       placeholder-gray-400"
            />
          </div>

          {/* File Upload */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Upload Files <span className="text-gray-400 font-normal">(Optional)</span>
            </label>

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
                Select Files
              </button>
              {files.length > 0 && (
                <span className="text-sm text-gray-500">
                  {files.length} file(s) selected
                </span>
              )}
            </div>

            {uploadProgress && (
              <div className="mt-2 text-sm text-blue-600 dark:text-blue-400 flex items-center gap-2">
                <Loader2 className="w-4 h-4 animate-spin" />
                {uploadProgress}
              </div>
            )}
          </div>

          {/* Execute Button */}
          <div className="flex items-center gap-4">
            <button
              onClick={handleExecute}
              disabled={loading || !selectedAgentId || !prompt.trim()}
              className="flex items-center gap-2 px-6 py-2.5 bg-blue-600 text-white rounded-lg
                       hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed
                       transition-colors font-medium"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Running...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" />
                  Execute
                </>
              )}
            </button>

            {loading && (
              <span className="text-sm text-gray-500">
                Task running, please wait...
              </span>
            )}
          </div>

          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg">
              {error}
            </div>
          )}
        </div>
      </div>

      {/* Results */}
      {(result || prompt) && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
          {/* Tabs */}
          <div className="flex border-b border-gray-200 dark:border-gray-700">
            <button
              onClick={() => setActiveTab('result')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors
                ${activeTab === 'result'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Terminal className="w-4 h-4 inline mr-2" />
              Result
            </button>
            <button
              onClick={() => setActiveTab('logs')}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors relative
                ${activeTab === 'logs'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'}`}
            >
              <Radio className="w-4 h-4 inline mr-2" />
              Live Logs
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

          {/* Tab Content */}
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
                      {result.result?.text ? cleanOutput(result.result.text) : 'No output'}
                    </pre>
                  </div>
                ) : loading ? (
                  <div className="flex flex-col items-center justify-center py-12 text-gray-500">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500 mb-3" />
                    <p>Task running, check Live Logs for progress...</p>
                  </div>
                ) : (
                  <div className="text-gray-500 text-center py-8">
                    Click Execute to run the task
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
                      {sseConnected ? 'SSE Connected' : 'SSE Disconnected'}
                    </span>
                  </div>
                  <span className="text-xs text-gray-400">{logs.length} entries</span>
                </div>
                <div className="h-80 overflow-y-auto bg-gray-900 rounded-lg p-4 font-mono text-sm">
                  {logs.length === 0 ? (
                    <div className="text-gray-500 text-center py-8">
                      Logs will appear here after execution
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

      {/* API Documentation */}
      <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
        <h3 className="font-medium text-blue-900 dark:text-blue-300 mb-2">API Usage</h3>
        <ol className="list-decimal list-inside text-sm text-blue-800 dark:text-blue-400 space-y-1">
          <li><code>POST /api/v1/agents/:id/run</code> - Run agent directly (simplest)</li>
          <li><code>POST /api/v1/sessions</code> - Create session (for interactive use)</li>
          <li><code>POST /api/v1/tasks</code> - Create async task</li>
          <li><code>GET /api/v1/tasks/:id</code> - Query task result</li>
        </ol>
      </div>
    </div>
  )
}
