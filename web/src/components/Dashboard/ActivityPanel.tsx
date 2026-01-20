import type { Session } from '../../types'
import { useLanguage } from '../../contexts/LanguageContext'
import { ActivityLog } from './ActivityLog'

interface ActivityPanelProps {
  sessions: Session[]
}

export function ActivityPanel({ sessions }: ActivityPanelProps) {
  const { t } = useLanguage()

  return (
    <div className="panel flex flex-col min-h-0">
      <div className="panel-header">
        <div>
          <div className="panel-title">{t('activityLog')}</div>
          <div className="panel-subtitle">{t('recentEvents')}</div>
        </div>
      </div>
      <div className="panel-content">
        <ActivityLog sessions={sessions} />
      </div>
    </div>
  )
}
