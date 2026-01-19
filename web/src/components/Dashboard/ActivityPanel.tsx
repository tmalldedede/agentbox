import type { Session } from '../../types'
import { ActivityLog } from './ActivityLog'

interface ActivityPanelProps {
  sessions: Session[]
}

export function ActivityPanel({ sessions }: ActivityPanelProps) {
  return (
    <div className="panel flex flex-col min-h-0">
      <div className="panel-header">
        <div>
          <div className="panel-title">Activity Log</div>
          <div className="panel-subtitle">Recent events</div>
        </div>
      </div>
      <div className="panel-content">
        <ActivityLog sessions={sessions} />
      </div>
    </div>
  )
}
