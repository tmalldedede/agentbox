import { useState } from 'react'
import {
  Play,
  Square,
  Trash2,
  Clock,
  Terminal,
  Activity,
  MoreVertical,
  AlertCircle,
  Loader2,
  ChevronRight,
} from 'lucide-react'
import type { Session } from '../../types'
import { useLanguage } from '../../contexts/LanguageContext'

interface SessionItemProps {
  session: Session
  onStart: () => void
  onStop: () => void
  onDelete: () => void
  onClick: () => void
}

const statusConfig: Record<string, { badge: string; icon: React.ReactNode; color: string }> = {
  running: {
    badge: 'badge-running',
    icon: <Activity className="w-3 h-3" />,
    color: '#10b981',
  },
  stopped: {
    badge: 'badge-stopped',
    icon: <Clock className="w-3 h-3" />,
    color: '#6b7280',
  },
  creating: {
    badge: 'badge-creating',
    icon: <Loader2 className="w-3 h-3 animate-spin" />,
    color: '#f59e0b',
  },
  error: {
    badge: 'badge-error',
    icon: <AlertCircle className="w-3 h-3" />,
    color: '#ef4444',
  },
}

const agentColors: Record<string, string> = {
  'claude-code': 'bg-purple-500/20 text-purple-400',
  codex: 'bg-emerald-500/20 text-emerald-400',
}

export function SessionItem({ session, onStart, onStop, onDelete, onClick }: SessionItemProps) {
  const { t } = useLanguage()
  const [showMenu, setShowMenu] = useState(false)

  const config = statusConfig[session.status] || statusConfig.stopped
  const agentInitials = session.agent.slice(0, 2).toUpperCase()

  return (
    <div className="list-item cursor-pointer group" onClick={onClick}>
      {/* Agent Avatar */}
      <div className={`agent-avatar ${agentColors[session.agent] || 'bg-blue-500/20 text-blue-400'}`}>
        {agentInitials}
      </div>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-3">
          <span className="font-medium text-primary truncate">{session.id}</span>
          <span className={`badge ${config.badge}`}>
            {config.icon}
            {t(session.status as 'running' | 'stopped' | 'creating' | 'error')}
          </span>
        </div>
        <div className="flex items-center gap-4 mt-1 text-xs text-muted">
          <span className="flex items-center gap-1">
            <Terminal className="w-3 h-3" />
            {session.agent}
          </span>
          <span className="truncate">{session.workspace}</span>
        </div>
      </div>

      {/* Actions */}
      <div className="relative" onClick={e => e.stopPropagation()}>
        <button
          onClick={() => setShowMenu(!showMenu)}
          className="btn btn-ghost btn-icon opacity-0 group-hover:opacity-100"
        >
          <MoreVertical className="w-4 h-4" />
        </button>

        {showMenu && (
          <>
            <div className="fixed inset-0 z-40" onClick={() => setShowMenu(false)} />
            <div className="dropdown-menu open">
              {session.status === 'running' && (
                <button
                  className="dropdown-item w-full text-left"
                  onClick={() => {
                    onStop()
                    setShowMenu(false)
                  }}
                >
                  <Square className="w-4 h-4 text-amber-400" />
                  <span>{t('stop')}</span>
                </button>
              )}
              {session.status === 'stopped' && (
                <button
                  className="dropdown-item w-full text-left"
                  onClick={() => {
                    onStart()
                    setShowMenu(false)
                  }}
                >
                  <Play className="w-4 h-4 text-emerald-400" />
                  <span>{t('start')}</span>
                </button>
              )}
              <button
                className="dropdown-item w-full text-left text-red-400"
                onClick={() => {
                  onDelete()
                  setShowMenu(false)
                }}
              >
                <Trash2 className="w-4 h-4" />
                <span>{t('delete')}</span>
              </button>
            </div>
          </>
        )}
      </div>

      {/* Arrow */}
      <ChevronRight className="w-5 h-5 text-muted group-hover:text-emerald-400 transition-colors" />
    </div>
  )
}
