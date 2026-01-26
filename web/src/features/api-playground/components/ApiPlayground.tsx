import { useState, useEffect, useRef, useCallback } from 'react'
import { api } from '@/services/api'
import {
  Play,
  Copy,
  Check,
  Loader2,
  Code,
  Upload,
  X,
  FileIcon,
  Radio,
  ChevronDown,
  ChevronRight,
  ArrowUpRight,
  ArrowDownLeft,
  Clock,
  FlaskConical,
  Network,
} from 'lucide-react'
import type { Agent } from '@/types'
import { useDockerAvailable } from '@/hooks/useSystemHealth'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'

// --- API Call Log Types ---
interface ApiCallEntry {
  id: number
  timestamp: number
  method: string
  url: string
  path: string
  status: number | null
  duration: number | null
  requestBody: string | null
  responseBody: string | null
  error: string | null
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
  path: string
  uploaded: boolean
}

interface ApiPlaygroundProps {
  preselectedAgentId?: string
  initialPrompt?: string
}

export default function ApiPlayground({ preselectedAgentId, initialPrompt }: ApiPlaygroundProps) {
  const dockerAvailable = useDockerAvailable()
  const [agents, setAgents] = useState<Agent[]>([])
  const [selectedAgentId, setSelectedAgentId] = useState('')
  const [prompt, setPrompt] = useState(initialPrompt || '')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TaskResult | null>(null)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState<'curl' | 'python' | null>(null)
  const [files, setFiles] = useState<UploadedFile[]>([])
  const [uploadProgress, setUploadProgress] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  // API Call Log
  const [apiCalls, setApiCalls] = useState<ApiCallEntry[]>([])
  const [expandedCalls, setExpandedCalls] = useState<Set<number>>(new Set())
  const callIdRef = useRef(0)

  // Container Logs (SSE)
  const [containerLogs, setContainerLogs] = useState<string[]>([])
  const [sseConnected, setSseConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const logsEndRef = useRef<HTMLDivElement>(null)

  // Active tab
  const [activeTab, setActiveTab] = useState('result')

  // --- Logged Fetch: 记录所有 API 调用 ---
  const loggedFetch = useCallback(async (
    url: string,
    options?: RequestInit
  ): Promise<Response> => {
    const id = ++callIdRef.current
    const method = options?.method || 'GET'
    const path = url.startsWith('http') ? new URL(url).pathname : url
    const startTime = Date.now()

    // 添加认证 token
    const token = localStorage.getItem('agentbox_token')
    const headers: HeadersInit = {
      ...(options?.headers || {}),
    }
    if (token && !headers['Authorization']) {
      headers['Authorization'] = `Bearer ${token}`
    }
    // 如果没有 FormData，添加 Content-Type
    if (!options?.body || !(options.body instanceof FormData)) {
      if (!headers['Content-Type']) {
        headers['Content-Type'] = 'application/json'
      }
    }

    // 记录请求体
    let requestBody: string | null = null
    if (options?.body) {
      if (typeof options.body === 'string') {
        requestBody = options.body
      } else if (options.body instanceof FormData) {
        requestBody = '[FormData]'
      }
    }

    // 添加进行中的调用
    const entry: ApiCallEntry = {
      id,
      timestamp: startTime,
      method,
      url,
      path,
      status: null,
      duration: null,
      requestBody,
      responseBody: null,
      error: null,
    }
    setApiCalls(prev => [...prev, entry])

    // Include Auth Header if not present
    const token = localStorage.getItem('agentbox_token')
    const finalOptions = {
      ...options,
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json',
        ...options?.headers,
      } as Record<string, string>,
    }

    // Remove Content-Type for FormData
    if (options?.body instanceof FormData) {
      delete finalOptions.headers['Content-Type']
    }

    try {
      const response = await fetch(url, finalOptions)
      const duration = Date.now() - startTime

      // 克隆 response 以读取 body
      const cloned = response.clone()
      let responseBody: string | null = null
      try {
        const text = await cloned.text()
        responseBody = text.length > 2000 ? text.slice(0, 2000) + '...' : text
      } catch {
        responseBody = '[Unable to read body]'
      }

      setApiCalls(prev =>
        prev.map(c => c.id === id ? { ...c, status: response.status, duration, responseBody } : c)
      )

      return response
    } catch (err) {
      const duration = Date.now() - startTime
      const errorMsg = err instanceof Error ? err.message : 'Unknown error'
      setApiCalls(prev =>
        prev.map(c => c.id === id ? { ...c, status: 0, duration, error: errorMsg } : c)
      )
      throw err
    }
  }, [])

  // Fetch agents list
  useEffect(() => {
    api.listAdminAgents()
      .then((data: Agent[]) => {
        setAgents(data || [])
      })
      .catch((err: Error) => {
        console.error('Failed to fetch agents:', err)
        setError('Failed to fetch agents. Please ensure you are logged in.')
      })
  }, [])

  // Sync selected agent and prompt from URL parameters
  useEffect(() => {
    if (agents.length === 0) return

    // 1. Sync selection if URL param provides a valid agent ID
    if (preselectedAgentId) {
      if (agents.find(a => a.id === preselectedAgentId)) {
        setSelectedAgentId(preselectedAgentId)
      }
    } else if (!selectedAgentId && agents.length > 0) {
      // Default selection (e.g. first public agent)
      const publicAgents = agents.filter(a => a.api_access !== 'private')
      setSelectedAgentId(publicAgents.length > 0 ? publicAgents[0].id : agents[0].id)
    }

    // 2. Sync prompt from URL or selected agent's description
    if (initialPrompt !== undefined) {
      setPrompt(initialPrompt)
    } else if (!prompt) {
      // Only default to agent description if prompt is currently empty and no initialPrompt given
      const targetAgentId = preselectedAgentId || selectedAgentId
      const currentAgent = agents.find(a => a.id === targetAgentId)
      if (currentAgent?.description) {
        setPrompt(currentAgent.description)
      }
    }
  }, [preselectedAgentId, initialPrompt, agents]) // Note: selectedAgentId and prompt are NOT deps to avoid loops

  // Auto-scroll container logs
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [containerLogs])

  // Cleanup SSE
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
    }
  }, [])

  // SSE Log Stream
  const connectLogStream = (sessionId: string) => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const token = localStorage.getItem('agentbox_token')
    const url = token
      ? `/api/v1/admin/sessions/${sessionId}/logs/stream?token=${token}`
      : `/api/v1/admin/sessions/${sessionId}/logs/stream`
    const es = new EventSource(url)
    eventSourceRef.current = es

    es.addEventListener('connected', (event) => {
      setSseConnected(true)
      const data = JSON.parse(event.data)
      setContainerLogs(prev => [...prev, `[Connected] Session: ${data.session_id}`])
    })

    es.onmessage = (event) => {
      if (event.data.trim()) {
        setContainerLogs(prev => [...prev, event.data])
      }
    }

    es.addEventListener('error', (event: MessageEvent) => {
      if (event.data) {
        setContainerLogs(prev => [...prev, `[Error] ${event.data}`])
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

      const res = await loggedFetch(`/api/v1/admin/sessions/${sessionId}/files?path=/`, {
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

  // Execute
  const handleExecute = async () => {
    if (!selectedAgentId || !prompt.trim()) {
      setError('Please select an Agent and enter a prompt')
      return
    }

    setLoading(true)
    setError('')
    setResult(null)
    setApiCalls([])
    setContainerLogs([])
    setActiveTab('api-calls')
    callIdRef.current = 0

    const selectedAgent = agents.find(a => a.id === selectedAgentId)

    try {
      // 1. Create Session
      const sessionRes = await loggedFetch('/api/v1/admin/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: selectedAgentId,
          workspace: `playground-${Date.now()}`
        })
      })
      const sessionData = await sessionRes.json()
      if (sessionData.code !== 0) {
        throw new Error(sessionData.message)
      }
      const sessionId = sessionData.data.id
      setContainerLogs(prev => [...prev, `[System] Session created: ${sessionId} (Agent: ${selectedAgent?.name || selectedAgentId})`])

      // 2. Upload files if any
      let uploadedPaths: string[] = []
      if (files.length > 0) {
        uploadedPaths = await uploadFilesToSession(sessionId)
        setContainerLogs(prev => [...prev, `[System] Files uploaded: ${uploadedPaths.length}`])
      }

      // 3. Build final prompt
      let finalPrompt = prompt
      if (uploadedPaths.length > 0) {
        const fileList = uploadedPaths.map(p => `/workspace${p}`).join('\n')
        finalPrompt = `Uploaded files:\n${fileList}\n\n${prompt}`
      }

      // 4. Connect log stream and create task
      connectLogStream(sessionId)

      const taskRes = await loggedFetch('/api/v1/tasks', {
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
      setContainerLogs(prev => [...prev, `[System] Task created: ${taskId}`])

      // 5. Poll for result
      let attempts = 0
      const maxAttempts = 300
      let taskDone = false
      while (attempts < maxAttempts) {
        await new Promise(resolve => setTimeout(resolve, 2000))
        const statusRes = await loggedFetch(`/api/v1/tasks/${taskId}`)
        const statusData = await statusRes.json()
        const taskStatus = statusData.data?.status
        const taskResult = statusData.data?.result

        // 终态 或 已有结果（多轮模式下 status 仍为 running 但已有 result）
        if (['completed', 'failed', 'cancelled'].includes(taskStatus) || taskResult) {
          setResult(statusData.data)
          setActiveTab('result')
          taskDone = true
          break
        }
        attempts++
      }

      if (!taskDone) {
        setError('Task timeout (exceeded 10 minutes)')
      }
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message || 'Execution failed')
    } finally {
      setLoading(false)
      setUploadProgress('')
      disconnectLogStream()
    }
  }

  // Code generation
  const generateCurl = () => {
    const baseUrl = window.location.origin
    return `# 1. Create Session
curl -X POST ${baseUrl}/api/v1/admin/sessions \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_id": "${selectedAgentId}",
    "workspace": "/tmp/my-task"
  }'

# 2. Create Task
curl -X POST ${baseUrl}/api/v1/tasks \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_id": "${selectedAgentId}",
    "prompt": "${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}"
  }'

# 3. Query Result
curl ${baseUrl}/api/v1/tasks/TASK_ID

# Alternative: Agent Run API (synchronous)
curl -X POST ${baseUrl}/api/v1/agents/${selectedAgentId}/run \\
  -H "Content-Type: application/json" \\
  -d '{
    "prompt": "${prompt.replace(/"/g, '\\"').replace(/\n/g, '\\n')}",
    "workspace": "/tmp/my-task"
  }'`
  }

  const generatePython = () => {
    const baseUrl = window.location.origin
    return `import requests
import time

BASE_URL = "${baseUrl}"

def run_agent(prompt: str, agent_id: str = "${selectedAgentId}"):
    """Run an Agent task and return result"""

    # Option 1: Agent Run API (simplest, synchronous)
    resp = requests.post(f"{BASE_URL}/api/v1/agents/{agent_id}/run", json={
        "prompt": prompt,
        "workspace": f"/tmp/task-{int(time.time())}"
    })
    return resp.json()["data"]

    # Option 2: Session + Task (async, more control)
    # session = requests.post(f"{BASE_URL}/api/v1/admin/sessions", json={
    #     "agent_id": agent_id,
    #     "workspace": f"/tmp/task-{int(time.time())}"
    # }).json()["data"]
    #
    # task = requests.post(f"{BASE_URL}/api/v1/tasks", json={
    #     "agent_id": agent_id,
    #     "prompt": prompt
    # }).json()["data"]
    #
    # while True:
    #     status = requests.get(f"{BASE_URL}/api/v1/tasks/{task['id']}").json()["data"]
    #     if status["status"] in ["completed", "failed"]:
    #         return status
    #     time.sleep(2)

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

  const toggleCallExpand = (id: number) => {
    setExpandedCalls(prev => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const getStatusColor = (status: number | null) => {
    if (status === null) return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
    if (status >= 200 && status < 300) return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
    if (status >= 400) return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
    return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
  }

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET': return 'text-blue-600 dark:text-blue-400'
      case 'POST': return 'text-green-600 dark:text-green-400'
      case 'PUT': return 'text-orange-600 dark:text-orange-400'
      case 'DELETE': return 'text-red-600 dark:text-red-400'
      default: return 'text-gray-600 dark:text-gray-400'
    }
  }

  const formatJson = (str: string | null) => {
    if (!str) return ''
    try {
      return JSON.stringify(JSON.parse(str), null, 2)
    } catch {
      return str
    }
  }

  return (
    <div className='space-y-6'>
      {/* Header */}
      <div>
        <h2 className='text-2xl font-bold tracking-tight flex items-center gap-2'>
          <FlaskConical className='h-6 w-6' />
          API Playground
        </h2>
        <p className='text-muted-foreground mt-1'>
          Test the AgentBox API with real-time call tracing
        </p>
      </div>

      {/* Configuration Card */}
      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          {/* Agent Selection */}
          <div className='space-y-2'>
            <Label>Agent</Label>
            <Select value={selectedAgentId} onValueChange={setSelectedAgentId}>
              <SelectTrigger>
                <SelectValue placeholder='Select an Agent' />
              </SelectTrigger>
              <SelectContent>
                {agents.map(a => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name} <span className='text-muted-foreground ml-1'>({a.adapter})</span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Prompt */}
          <div className='space-y-2'>
            <Label>Prompt</Label>
            <Textarea
              value={prompt}
              onChange={e => setPrompt(e.target.value)}
              placeholder='Enter the task for the AI Agent...'
              rows={4}
            />
          </div>

          {/* File Upload */}
          <div className='space-y-2'>
            <Label>
              Attachments <span className='text-muted-foreground font-normal'>(Optional)</span>
            </Label>

            {files.length > 0 && (
              <div className='space-y-1'>
                {files.map((file, index) => (
                  <div
                    key={index}
                    className='flex items-center justify-between px-3 py-1.5 bg-muted rounded-md'
                  >
                    <div className='flex items-center gap-2 min-w-0'>
                      <FileIcon className='w-4 h-4 text-muted-foreground flex-shrink-0' />
                      <span className='text-sm truncate'>{file.file.name}</span>
                      <span className='text-xs text-muted-foreground flex-shrink-0'>
                        {formatFileSize(file.file.size)}
                      </span>
                      {file.uploaded && <Check className='w-4 h-4 text-green-500 flex-shrink-0' />}
                    </div>
                    <Button variant='ghost' size='icon' className='h-6 w-6' onClick={() => removeFile(index)} disabled={loading}>
                      <X className='w-3 h-3' />
                    </Button>
                  </div>
                ))}
              </div>
            )}

            <div className='flex items-center gap-3'>
              <input ref={fileInputRef} type='file' multiple onChange={handleFileSelect} className='hidden' />
              <Button variant='outline' size='sm' onClick={() => fileInputRef.current?.click()} disabled={loading}>
                <Upload className='mr-2 h-4 w-4' />
                Select Files
              </Button>
              {files.length > 0 && (
                <span className='text-sm text-muted-foreground'>{files.length} file(s)</span>
              )}
            </div>

            {uploadProgress && (
              <div className='text-sm text-blue-600 flex items-center gap-2'>
                <Loader2 className='w-4 h-4 animate-spin' />
                {uploadProgress}
              </div>
            )}
          </div>

          {/* Execute */}
          <div className='flex items-center gap-4 pt-2'>
            <Button onClick={handleExecute} disabled={loading || !selectedAgentId || !prompt.trim() || !dockerAvailable}>
              {loading ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Running...
                </>
              ) : (
                <>
                  <Play className='mr-2 h-4 w-4' />
                  Execute
                </>
              )}
            </Button>
            {loading && (
              <span className='text-sm text-muted-foreground'>Task running, check API Calls tab...</span>
            )}
          </div>

          {error && (
            <div className='p-3 bg-destructive/10 text-destructive rounded-md text-sm'>
              {error}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Results Tabs */}
      <Card>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <CardHeader className='pb-0'>
            <TabsList className='w-full justify-start'>
              <TabsTrigger value='result' className='gap-1.5'>
                Result
                {result && (
                  <Badge variant={result.status === 'completed' ? 'default' : 'destructive'} className='ml-1 text-[10px] px-1.5 py-0'>
                    {result.status}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value='api-calls' className='gap-1.5'>
                <Network className='h-3.5 w-3.5' />
                API Calls
                {apiCalls.length > 0 && (
                  <Badge variant='secondary' className='ml-1 text-[10px] px-1.5 py-0'>
                    {apiCalls.length}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value='container-logs' className='gap-1.5'>
                <Radio className='h-3.5 w-3.5' />
                Container Logs
                {sseConnected && (
                  <span className='w-2 h-2 bg-green-500 rounded-full animate-pulse' />
                )}
              </TabsTrigger>
              <TabsTrigger value='curl' className='gap-1.5'>
                <Code className='h-3.5 w-3.5' />
                cURL
              </TabsTrigger>
              <TabsTrigger value='python' className='gap-1.5'>
                <Code className='h-3.5 w-3.5' />
                Python
              </TabsTrigger>
            </TabsList>
          </CardHeader>

          <CardContent className='pt-4'>
            {/* Result Tab */}
            <TabsContent value='result' className='mt-0'>
              {result ? (
                <div className='space-y-3'>
                  <div className='flex items-center gap-2'>
                    <Badge variant={result.status === 'completed' ? 'default' : 'destructive'}>
                      {result.status}
                    </Badge>
                    <span className='text-sm text-muted-foreground font-mono'>ID: {result.id}</span>
                  </div>
                  <pre className='p-4 bg-muted rounded-lg overflow-x-auto text-sm whitespace-pre-wrap max-h-96 overflow-y-auto'>
                    {result.result?.text ? cleanOutput(result.result.text) : 'No output'}
                  </pre>
                </div>
              ) : loading ? (
                <div className='flex flex-col items-center justify-center py-12 text-muted-foreground'>
                  <Loader2 className='w-8 h-8 animate-spin mb-3' />
                  <p>Task running, check API Calls tab for progress...</p>
                </div>
              ) : (
                <div className='text-muted-foreground text-center py-8'>
                  Click Execute to run the task
                </div>
              )}
            </TabsContent>

            {/* API Calls Tab */}
            <TabsContent value='api-calls' className='mt-0'>
              {apiCalls.length === 0 ? (
                <div className='text-muted-foreground text-center py-8'>
                  API calls will appear here after execution
                </div>
              ) : (
                <div className='space-y-1'>
                  {apiCalls.map((call) => {
                    const isExpanded = expandedCalls.has(call.id)
                    return (
                      <div key={call.id} className='border rounded-md overflow-hidden'>
                        {/* Call Summary Row */}
                        <button
                          onClick={() => toggleCallExpand(call.id)}
                          className='w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-muted/50 transition-colors'
                        >
                          {isExpanded ? (
                            <ChevronDown className='h-3.5 w-3.5 text-muted-foreground flex-shrink-0' />
                          ) : (
                            <ChevronRight className='h-3.5 w-3.5 text-muted-foreground flex-shrink-0' />
                          )}
                          <span className={`font-mono font-bold w-12 text-left ${getMethodColor(call.method)}`}>
                            {call.method}
                          </span>
                          <span className='font-mono text-left flex-1 truncate'>
                            {call.path}
                          </span>
                          {call.status !== null ? (
                            <Badge variant='outline' className={`text-[10px] px-1.5 py-0 ${getStatusColor(call.status)}`}>
                              {call.status}
                            </Badge>
                          ) : (
                            <Loader2 className='h-3.5 w-3.5 animate-spin text-muted-foreground' />
                          )}
                          {call.duration !== null && (
                            <span className='text-xs text-muted-foreground flex items-center gap-0.5 w-16 justify-end'>
                              <Clock className='h-3 w-3' />
                              {call.duration}ms
                            </span>
                          )}
                        </button>

                        {/* Expanded Details */}
                        {isExpanded && (
                          <div className='border-t bg-muted/30 px-3 py-3 space-y-3'>
                            <div className='grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-xs'>
                              <span className='text-muted-foreground'>URL:</span>
                              <span className='font-mono break-all'>{call.url}</span>
                              <span className='text-muted-foreground'>Time:</span>
                              <span>{new Date(call.timestamp).toLocaleTimeString()}</span>
                              {call.error && (
                                <>
                                  <span className='text-destructive'>Error:</span>
                                  <span className='text-destructive'>{call.error}</span>
                                </>
                              )}
                            </div>

                            {call.requestBody && call.requestBody !== '[FormData]' && (
                              <div className='space-y-1'>
                                <div className='flex items-center gap-1 text-xs text-muted-foreground'>
                                  <ArrowUpRight className='h-3 w-3' />
                                  Request Body
                                </div>
                                <pre className='p-2 bg-muted rounded text-xs overflow-x-auto max-h-40 overflow-y-auto'>
                                  {formatJson(call.requestBody)}
                                </pre>
                              </div>
                            )}

                            {call.responseBody && (
                              <div className='space-y-1'>
                                <div className='flex items-center gap-1 text-xs text-muted-foreground'>
                                  <ArrowDownLeft className='h-3 w-3' />
                                  Response Body
                                </div>
                                <pre className='p-2 bg-muted rounded text-xs overflow-x-auto max-h-40 overflow-y-auto'>
                                  {formatJson(call.responseBody)}
                                </pre>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </TabsContent>

            {/* Container Logs Tab */}
            <TabsContent value='container-logs' className='mt-0'>
              <div className='space-y-2'>
                <div className='flex items-center justify-between'>
                  <div className='flex items-center gap-2'>
                    <span className={`w-2 h-2 rounded-full ${sseConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
                    <span className='text-sm text-muted-foreground'>
                      {sseConnected ? 'SSE Connected' : 'Disconnected'}
                    </span>
                  </div>
                  <span className='text-xs text-muted-foreground'>{containerLogs.length} entries</span>
                </div>
                <div className='h-80 overflow-y-auto bg-muted rounded-lg p-4 font-mono text-xs'>
                  {containerLogs.length === 0 ? (
                    <div className='text-muted-foreground text-center py-8'>
                      Container logs will appear here during execution
                    </div>
                  ) : (
                    containerLogs.map((log, index) => (
                      <div
                        key={index}
                        className={`py-0.5 ${log.startsWith('[System]') ? 'text-blue-500' :
                          log.startsWith('[Connected]') ? 'text-green-500' :
                            log.startsWith('[Error]') ? 'text-destructive' :
                              ''
                          }`}
                      >
                        {log}
                      </div>
                    ))
                  )}
                  <div ref={logsEndRef} />
                </div>
              </div>
            </TabsContent>

            {/* cURL Tab */}
            <TabsContent value='curl' className='mt-0'>
              <div className='relative'>
                <Button
                  variant='ghost'
                  size='icon'
                  className='absolute top-2 right-2 h-8 w-8'
                  onClick={() => copyToClipboard('curl')}
                >
                  {copied === 'curl' ? <Check className='h-4 w-4 text-green-500' /> : <Copy className='h-4 w-4' />}
                </Button>
                <pre className='p-4 bg-muted rounded-lg overflow-x-auto text-xs max-h-96 overflow-y-auto'>
                  {generateCurl()}
                </pre>
              </div>
            </TabsContent>

            {/* Python Tab */}
            <TabsContent value='python' className='mt-0'>
              <div className='relative'>
                <Button
                  variant='ghost'
                  size='icon'
                  className='absolute top-2 right-2 h-8 w-8'
                  onClick={() => copyToClipboard('python')}
                >
                  {copied === 'python' ? <Check className='h-4 w-4 text-green-500' /> : <Copy className='h-4 w-4' />}
                </Button>
                <pre className='p-4 bg-muted rounded-lg overflow-x-auto text-xs max-h-96 overflow-y-auto'>
                  {generatePython()}
                </pre>
              </div>
            </TabsContent>
          </CardContent>
        </Tabs>
      </Card>

      {/* Quick API Reference */}
      <Card>
        <CardHeader>
          <CardTitle className='text-sm'>API Endpoints</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='grid grid-cols-1 md:grid-cols-2 gap-2 text-xs'>
            <div className='flex items-center gap-2 font-mono'>
              <Badge variant='outline' className='text-[10px] bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'>POST</Badge>
              /api/v1/agents/:id/run
            </div>
            <div className='flex items-center gap-2 font-mono'>
              <Badge variant='outline' className='text-[10px] bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'>POST</Badge>
              /api/v1/admin/sessions
            </div>
            <div className='flex items-center gap-2 font-mono'>
              <Badge variant='outline' className='text-[10px] bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'>POST</Badge>
              /api/v1/tasks
            </div>
            <div className='flex items-center gap-2 font-mono'>
              <Badge variant='outline' className='text-[10px] bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'>GET</Badge>
              /api/v1/tasks/:id
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
