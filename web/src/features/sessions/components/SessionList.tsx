import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Terminal,
  RefreshCw,
  AlertCircle,
  Loader2,
  Play,
  Square,
  Trash2,
  ChevronRight,
  Activity,
  Cpu,
  Shield,
  Code2,
} from 'lucide-react'
import type { Session, Agent } from '@/types'
import { useSessions, useDeleteSession, useStartSession, useStopSession } from '@/hooks'
import { api } from '@/services/api'
import { useLanguage } from '@/contexts/LanguageContext'
import CreateSessionModal from './CreateSessionModal'

// Agent colors
const agentColors: Record<string, string> = {
  'claude-code': 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  codex: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  opencode: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
}

// Status colors
const statusColors: Record<string, string> = {
  running: 'bg-emerald-500/20 text-emerald-400',
  stopped: 'bg-gray-500/20 text-gray-400',
  creating: 'bg-blue-500/20 text-blue-400',
  error: 'bg-red-500/20 text-red-400',
}

// Session Card Component (matching ProfileCard style)
function SessionCard({
  session,
  onStart,
  onStop,
  onDelete,
  onClick,
  isStarting,
  isStopping,
  isDeleting,
}: {
  session: Session
  onStart: () => void
  onStop: () => void
  onDelete: () => void
  onClick: () => void
  isStarting: boolean
  isStopping: boolean
  isDeleting: boolean
}) {
  const colors = agentColors[session.agent] || 'bg-gray-500/20 text-gray-400 border-gray-500/30'
  const initials = session.agent.slice(0, 2).toUpperCase()

  return (
    <div
      className="card p-4 cursor-pointer group hover:border-emerald-500/50 transition-colors"
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        {/* Avatar */}
        <div
          className={`w-12 h-12 rounded-xl flex items-center justify-center text-sm font-bold ${colors}`}
        >
          {initials}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">
              {session.id.slice(0, 16)}...
            </span>
            <span className={`text-xs px-2 py-0.5 rounded ${statusColors[session.status]}`}>
              {session.status}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1 line-clamp-1">
            {session.workspace}
          </p>

          {/* Tags */}
          <div className="flex items-center gap-2 mt-3 flex-wrap">
            <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/80">
              {session.agent}
            </span>
            {session.profile_id && (
              <span className="text-xs px-2 py-0.5 rounded bg-amber-500/20 text-amber-600 dark:text-amber-400">
                Profile
              </span>
            )}
            <span className="text-xs text-muted-foreground">
              {new Date(session.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>

        {/* Actions */}
        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          {session.status === 'running' ? (
            <button
              onClick={onStop}
              className="btn btn-ghost btn-icon"
              title="Stop"
              disabled={isStopping}
            >
              <Square className={`w-4 h-4 ${isStopping ? 'animate-pulse' : ''}`} />
            </button>
          ) : (
            <button
              onClick={onStart}
              className="btn btn-ghost btn-icon text-emerald-400"
              title="Start"
              disabled={isStarting || session.status === 'creating'}
            >
              <Play className={`w-4 h-4 ${isStarting ? 'animate-pulse' : ''}`} />
            </button>
          )}
          {session.status !== 'running' && (
            <button
              onClick={onDelete}
              className="btn btn-ghost btn-icon text-red-400"
              title="Delete"
              disabled={isDeleting}
            >
              <Trash2 className={`w-4 h-4 ${isDeleting ? 'animate-pulse' : ''}`} />
            </button>
          )}
        </div>

        {/* Arrow */}
        <ChevronRight className="w-5 h-5 text-muted-foreground group-hover:text-emerald-400 transition-colors" />
      </div>
    </div>
  )
}

// Session Group Component (matching ProfileGroup style)
function SessionGroup({
  title,
  icon,
  iconBgColor,
  sessions,
  onStart,
  onStop,
  onDelete,
  onClick,
  actioningId,
  actionType,
}: {
  title: string
  icon: React.ReactNode
  iconBgColor: string
  sessions: Session[]
  onStart: (session: Session) => void
  onStop: (session: Session) => void
  onDelete: (session: Session) => void
  onClick: (session: Session) => void
  actioningId?: string
  actionType?: 'start' | 'stop' | 'delete'
}) {
  if (sessions.length === 0) return null

  return (
    <div>
      <div className="flex items-center gap-3 mb-4">
        <div className={`w-8 h-8 rounded-lg ${iconBgColor} flex items-center justify-center`}>
          {icon}
        </div>
        <h2 className="text-lg font-semibold text-foreground">{title}</h2>
        <span className="text-sm text-muted-foreground">({sessions.length})</span>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {sessions.map(session => (
          <SessionCard
            key={session.id}
            session={session}
            onStart={() => onStart(session)}
            onStop={() => onStop(session)}
            onDelete={() => onDelete(session)}
            onClick={() => onClick(session)}
            isStarting={actioningId === session.id && actionType === 'start'}
            isStopping={actioningId === session.id && actionType === 'stop'}
            isDeleting={actioningId === session.id && actionType === 'delete'}
          />
        ))}
      </div>
    </div>
  )
}

export default function SessionList() {
  const navigate = useNavigate()
  const { t } = useLanguage()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [actioningId, setActioningId] = useState<string | undefined>()
  const [actionType, setActionType] = useState<'start' | 'stop' | 'delete' | undefined>()
  const [agents, setAgents] = useState<Agent[]>([])
  const [agentsLoading, setAgentsLoading] = useState(false)

  // React Query hooks
  const { data: sessions = [], isLoading, isFetching, error, refetch } = useSessions()

  const deleteSession = useDeleteSession()
  const startSession = useStartSession()
  const stopSession = useStopSession()

  const handleOpenCreate = async () => {
    setAgentsLoading(true)
    try {
      const fetchedAgents = await api.listAgents()
      setAgents(fetchedAgents)
      setShowCreateModal(true)
    } catch (err) {
      console.error('Failed to fetch agents:', err)
    } finally {
      setAgentsLoading(false)
    }
  }

  const handleStart = (session: Session) => {
    setActioningId(session.id)
    setActionType('start')
    startSession.mutate(session.id, {
      onSettled: () => {
        setActioningId(undefined)
        setActionType(undefined)
      },
    })
  }

  const handleStop = (session: Session) => {
    setActioningId(session.id)
    setActionType('stop')
    stopSession.mutate(session.id, {
      onSettled: () => {
        setActioningId(undefined)
        setActionType(undefined)
      },
    })
  }

  const handleDelete = (session: Session) => {
    if (!confirm(`Delete session "${session.id.slice(0, 12)}..."?`)) return
    setActioningId(session.id)
    setActionType('delete')
    deleteSession.mutate(session.id, {
      onSettled: () => {
        setActioningId(undefined)
        setActionType(undefined)
      },
    })
  }

  const handleClick = (session: Session) => {
    navigate({ to: `/sessions/${session.id}` })
  }

  // Group sessions by agent
  const claudeSessions = sessions.filter(s => s.agent === 'claude-code')
  const codexSessions = sessions.filter(s => s.agent === 'codex')
  const opencodeSessions = sessions.filter(s => s.agent === 'opencode')
  const otherSessions = sessions.filter(
    s => !['claude-code', 'codex', 'opencode'].includes(s.agent)
  )

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Activity className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">{t('sessions')}</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="btn btn-ghost btn-icon"
            disabled={isFetching}
          >
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button
            className="btn btn-primary"
            onClick={handleOpenCreate}
            disabled={agentsLoading}
          >
            {agentsLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Plus className="w-4 h-4" />
            )}
            {t('createSession')}
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load sessions'}
            </span>
          </div>
        )}

        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">{t('sessions')}</h1>
          <p className="text-muted-foreground">
            Manage your active and stopped agent sessions. Click on a session to view details and
            interact with the terminal.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Terminal className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg">No sessions found</p>
            <p className="text-muted-foreground mt-2">Create your first session to get started</p>
          </div>
        ) : (
          <div className="space-y-8">
            <SessionGroup
              title="Claude Code"
              icon={<Cpu className="w-4 h-4 text-purple-400" />}
              iconBgColor="bg-purple-500/20"
              sessions={claudeSessions}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              onClick={handleClick}
              actioningId={actioningId}
              actionType={actionType}
            />

            <SessionGroup
              title="Codex"
              icon={<Shield className="w-4 h-4 text-emerald-400" />}
              iconBgColor="bg-emerald-500/20"
              sessions={codexSessions}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              onClick={handleClick}
              actioningId={actioningId}
              actionType={actionType}
            />

            <SessionGroup
              title="OpenCode"
              icon={<Code2 className="w-4 h-4 text-blue-400" />}
              iconBgColor="bg-blue-500/20"
              sessions={opencodeSessions}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              onClick={handleClick}
              actioningId={actioningId}
              actionType={actionType}
            />

            <SessionGroup
              title="Other"
              icon={<Terminal className="w-4 h-4 text-gray-400" />}
              iconBgColor="bg-gray-500/20"
              sessions={otherSessions}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              onClick={handleClick}
              actioningId={actioningId}
              actionType={actionType}
            />
          </div>
        )}
      </div>

      {/* Create Session Modal */}
      {showCreateModal && (
        <CreateSessionModal
          agents={agents}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false)
            refetch()
          }}
        />
      )}
    </div>
  )
}
