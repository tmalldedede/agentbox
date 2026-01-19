import { Activity } from 'lucide-react'
import type { Session } from '../../types'

interface ActivityLogProps {
  sessions: Session[]
}

export function ActivityLog({ sessions }: ActivityLogProps) {
  if (sessions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <Activity className="w-12 h-12 text-muted mb-4" />
        <p className="text-secondary">No activity yet</p>
      </div>
    )
  }

  return (
    <>
      {sessions.slice(0, 10).map((session, i) => (
        <div key={`${session.id}-${i}`} className="message-item">
          <div className="flex items-center gap-2 mb-1">
            <span
              className={`w-6 h-6 rounded text-xs font-bold flex items-center justify-center ${
                session.status === 'running'
                  ? 'bg-emerald-500/20 text-emerald-400'
                  : session.status === 'error'
                    ? 'bg-red-500/20 text-red-400'
                    : 'bg-secondary text-muted'
              }`}
            >
              {session.agent.slice(0, 2).toUpperCase()}
            </span>
            <span className="text-xs text-muted">
              {new Date(session.created_at).toLocaleTimeString()}
            </span>
          </div>
          <p className="text-sm text-secondary">
            Session <span className="text-primary font-medium">{session.id}</span> {session.status}
          </p>
        </div>
      ))}
    </>
  )
}
