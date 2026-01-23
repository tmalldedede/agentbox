import { useState, useMemo } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Route } from '@/routes/_authenticated/sessions/$id'
import {
  ArrowLeft,
  Play,
  Square,
  Trash2,
  Send,
  Terminal,
  Clock,
  Folder,
  Cpu,
  Activity,
  AlertCircle,
  Loader2,
  RefreshCw,
  Box,
  HardDrive,
  MemoryStick,
} from 'lucide-react'
import type { ExecResponse } from '@/types'
import { useLanguage } from '@/contexts/LanguageContext'
import {
  useSession,
  useStartSession,
  useStopSession,
  useDeleteSession,
  useExecSession,
} from '@/hooks'
import TerminalViewer, { formatExecutionLogs } from './TerminalViewer'

interface ExecutionHistory {
  id: string
  prompt: string
  output: string
  exitCode: number
  error?: string
  timestamp: Date
}

export default function SessionDetail() {
  const params = Route.useParams()
  const sessionId = params?.id
  const navigate = useNavigate()
  const { t } = useLanguage()

  // 使用 React Query hooks
  const {
    data: session,
    isLoading: sessionLoading,
    error: sessionError,
    refetch: refetchSession,
  } = useSession(sessionId)

  // Mutations
  const startSession = useStartSession()
  const stopSession = useStopSession()
  const deleteSession = useDeleteSession()
  const execSession = useExecSession(sessionId)

  const [prompt, setPrompt] = useState('')
  const [history, setHistory] = useState<ExecutionHistory[]>([])

  // 格式化执行历史为终端显示内容
  const terminalContent = useMemo(() => {
    if (history.length === 0) return ''

    return formatExecutionLogs(
      history.map(item => ({
        id: item.id,
        prompt: item.prompt,
        output: item.output || undefined,
        error: item.error,
        exitCode: item.exitCode,
        status: item.exitCode === -1 ? 'running' : item.exitCode === 0 ? 'success' : 'failed',
        startedAt: item.timestamp,
        endedAt: item.exitCode === -1 ? undefined : item.timestamp,
      }))
    )
  }, [history])

  const handleExecute = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim() || !sessionId || execSession.isPending) return

    const currentPrompt = prompt
    setPrompt('')

    // Add to history immediately
    const historyItem: ExecutionHistory = {
      id: Date.now().toString(),
      prompt: currentPrompt,
      output: '',
      exitCode: -1,
      timestamp: new Date(),
    }
    setHistory(prev => [...prev, historyItem])

    try {
      const result: ExecResponse = await execSession.mutateAsync(currentPrompt)

      // Update history with result
      setHistory(prev =>
        prev.map(item =>
          item.id === historyItem.id
            ? {
                ...item,
                output: result.output || '',
                exitCode: result.exit_code,
                error: result.error,
              }
            : item
        )
      )
    } catch (err) {
      setHistory(prev =>
        prev.map(item =>
          item.id === historyItem.id
            ? {
                ...item,
                output: '',
                exitCode: 1,
                error: err instanceof Error ? err.message : 'Execution failed',
              }
            : item
        )
      )
    }
  }

  const handleStart = () => {
    if (sessionId) {
      startSession.mutate(sessionId)
    }
  }

  const handleStop = () => {
    if (sessionId) {
      stopSession.mutate(sessionId)
    }
  }

  const handleDelete = () => {
    if (!sessionId || !confirm(t('confirmDelete'))) return
    deleteSession.mutate(sessionId, {
      onSuccess: () => navigate({ to: '/' }),
    })
  }

  const statusConfig: Record<string, { badge: string; color: string }> = {
    running: { badge: 'badge-running', color: '#10b981' },
    stopped: { badge: 'badge-stopped', color: '#6b7280' },
    creating: { badge: 'badge-creating', color: '#f59e0b' },
    error: { badge: 'badge-error', color: '#ef4444' },
  }

  if (sessionLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  if (sessionError || !session) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center">
        <AlertCircle className="w-12 h-12 text-red-400 mb-4" />
        <p className="text-muted-foreground">
          {sessionError instanceof Error ? sessionError.message : 'Session not found'}
        </p>
        <button onClick={() => navigate({ to: '/' })} className="btn btn-secondary mt-4">
          <ArrowLeft className="w-4 h-4" />
          Back to Dashboard
        </button>
      </div>
    )
  }

  const config = statusConfig[session.status] || statusConfig.stopped

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="app-header flex-shrink-0">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Terminal className="w-5 h-5 text-emerald-400" />
            <span className="font-semibold">{session.id}</span>
            <span className={`badge ${config.badge}`}>
              {session.status === 'running' && <Activity className="w-3 h-3" />}
              {t(session.status as 'running' | 'stopped' | 'creating' | 'error')}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button onClick={() => refetchSession()} className="btn btn-ghost btn-icon">
            <RefreshCw className="w-4 h-4" />
          </button>
          {session.status === 'running' ? (
            <button
              onClick={handleStop}
              className="btn btn-secondary"
              disabled={stopSession.isPending}
            >
              {stopSession.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Square className="w-4 h-4" />
              )}
              {t('stop')}
            </button>
          ) : session.status === 'stopped' ? (
            <button
              onClick={handleStart}
              className="btn btn-primary"
              disabled={startSession.isPending}
            >
              {startSession.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              {t('start')}
            </button>
          ) : null}
          <button
            onClick={handleDelete}
            className="btn btn-danger"
            disabled={deleteSession.isPending}
          >
            {deleteSession.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Trash2 className="w-4 h-4" />
            )}
            {t('delete')}
          </button>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {/* Left: Session Info */}
        <div
          className="w-80 border-r p-6 flex-shrink-0 overflow-y-auto bg-card"
          style={{ borderColor: 'var(--border-color)' }}
        >
          <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            Session Info
          </h3>

          <div className="space-y-4">
            {/* Agent */}
            <div
              className="card p-4 cursor-pointer hover:border-emerald-500/50 transition-colors"
              onClick={() => session.agent_id && navigate({ to: `/agents/${session.agent_id}` })}
            >
              <div className="flex items-center gap-3 mb-3">
                <Cpu className="w-5 h-5 text-purple-400" />
                <span className="font-medium">Agent</span>
              </div>
              <p className="text-muted-foreground">{session.agent_id || session.agent}</p>
              <p className="text-xs text-muted-foreground mt-1">{session.agent}</p>
            </div>

            {/* Workspace */}
            <div className="card p-4">
              <div className="flex items-center gap-3 mb-3">
                <Folder className="w-5 h-5 text-blue-400" />
                <span className="font-medium">Workspace</span>
              </div>
              <p className="text-muted-foreground text-sm font-mono break-all">{session.workspace}</p>
            </div>

            {/* Container */}
            {session.container_id && (
              <div className="card p-4">
                <div className="flex items-center gap-3 mb-3">
                  <Box className="w-5 h-5 text-emerald-400" />
                  <span className="font-medium">Container</span>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={`w-2 h-2 rounded-full ${
                      session.status === 'running' ? 'bg-emerald-400' : 'bg-gray-400'
                    }`}
                  />
                  <p className="text-muted-foreground text-sm font-mono">
                    {session.container_id.slice(0, 12)}
                  </p>
                </div>
              </div>
            )}

            {/* Resources */}
            <div className="card p-4">
              <div className="flex items-center gap-3 mb-3">
                <HardDrive className="w-5 h-5 text-amber-400" />
                <span className="font-medium">Resources</span>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Cpu className="w-3 h-3" /> CPU
                  </span>
                  <span className="text-muted-foreground">{session.config.cpu_limit} cores</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <MemoryStick className="w-3 h-3" /> Memory
                  </span>
                  <span className="text-muted-foreground">
                    {Math.round(session.config.memory_limit / (1024 * 1024 * 1024))} GB
                  </span>
                </div>
              </div>
            </div>

            {/* Created */}
            <div className="card p-4">
              <div className="flex items-center gap-3 mb-3">
                <Clock className="w-5 h-5 text-gray-400" />
                <span className="font-medium">Created</span>
              </div>
              <p className="text-muted-foreground text-sm">
                {new Date(session.created_at).toLocaleString()}
              </p>
            </div>
          </div>
        </div>

        {/* Right: Terminal */}
        <div className="flex-1 flex flex-col">
          {/* Output Area */}
          <div className="flex-1 overflow-hidden">
            {history.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-muted-foreground bg-[#0d1117]">
                <Terminal className="w-12 h-12 mb-4 opacity-50" />
                <p>Execute a prompt to see output here</p>
                {session.status !== 'running' && (
                  <p className="text-amber-400 mt-2">
                    Session is {session.status}. Start it first.
                  </p>
                )}
              </div>
            ) : (
              <TerminalViewer
                content={terminalContent}
                loading={execSession.isPending}
                placeholder="Execute a prompt to see output here"
                minHeight="100%"
              />
            )}
          </div>

          {/* Input Area */}
          <form
            onSubmit={handleExecute}
            className="p-4 border-t border-default flex items-center gap-3 terminal-input-area"
          >
            <span className="text-emerald-400 font-mono">❯</span>
            <input
              type="text"
              value={prompt}
              onChange={e => setPrompt(e.target.value)}
              placeholder={
                session.status === 'running' ? 'Enter a prompt to execute...' : 'Session not running'
              }
              disabled={session.status !== 'running' || execSession.isPending}
              className="flex-1 bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground font-mono"
              autoFocus
            />
            <button
              type="submit"
              disabled={!prompt.trim() || session.status !== 'running' || execSession.isPending}
              className="btn btn-primary"
            >
              {execSession.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Send className="w-4 h-4" />
              )}
              {t('execute')}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
