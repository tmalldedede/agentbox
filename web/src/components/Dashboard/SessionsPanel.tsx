import { useNavigate } from 'react-router-dom'
import { Plus, RefreshCw, Server } from 'lucide-react'
import { toast } from 'sonner'
import type { Session } from '../../types'
import { useLanguage } from '../../contexts/LanguageContext'
import { useDeleteSession, useStopSession } from '../../hooks'
import { api } from '../../services/api'
import { SessionItem } from './SessionItem'

interface SessionsPanelProps {
  sessions: Session[]
  isLoading: boolean
  isFetching: boolean
  onRefresh: () => void
  onCreateClick: () => void
}

export function SessionsPanel({
  sessions,
  isLoading,
  isFetching,
  onRefresh,
  onCreateClick,
}: SessionsPanelProps) {
  const navigate = useNavigate()
  const { t } = useLanguage()
  const deleteSession = useDeleteSession()
  const stopSession = useStopSession()

  const runningSessions = sessions.filter(s => s.status === 'running').length

  const handleStart = async (id: string) => {
    try {
      await api.startSession(id)
      onRefresh()
      toast.success('会话已启动')
    } catch (err) {
      toast.error(`启动失败: ${err instanceof Error ? err.message : '未知错误'}`)
    }
  }

  const handleStop = (id: string) => {
    stopSession.mutate(id)
  }

  const handleDelete = (id: string) => {
    if (!confirm(t('confirmDelete'))) return
    deleteSession.mutate(id)
  }

  return (
    <div className="panel flex flex-col min-h-0">
      <div className="panel-header">
        <div>
          <div className="panel-title">{t('sessions')}</div>
          <div className="panel-subtitle">
            {runningSessions} {t('running').toLowerCase()}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={onRefresh} className="btn btn-ghost btn-icon">
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button onClick={onCreateClick} className="btn btn-primary">
            <Plus className="w-4 h-4" />
            {t('newSession')}
          </button>
        </div>
      </div>
      <div className="panel-content">
        {isLoading && sessions.length === 0 ? (
          <div className="p-5 space-y-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-16 skeleton rounded-lg" />
            ))}
          </div>
        ) : sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center p-8">
            <Server className="w-12 h-12 text-muted mb-4" />
            <p className="text-secondary">{t('noSessions')}</p>
          </div>
        ) : (
          sessions.map(session => (
            <SessionItem
              key={session.id}
              session={session}
              onStart={() => handleStart(session.id)}
              onStop={() => handleStop(session.id)}
              onDelete={() => handleDelete(session.id)}
              onClick={() => navigate(`/sessions/${session.id}`)}
            />
          ))
        )}
      </div>
    </div>
  )
}
