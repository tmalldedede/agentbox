import type { DashboardStats } from '@/types'

interface Props {
  stats: DashboardStats | null
  currentTime: Date
}

export function HeaderBar({ stats, currentTime }: Props) {
  const agentsActive = stats?.agents.active ?? 0
  const tasksRunning = stats?.tasks.by_status?.running ?? 0
  const sessionsRunning = stats?.sessions.running ?? 0

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  return (
    <div className="cc-header cc-animate-in">
      <div className="cc-header-left">
        <div className="cc-header-logo">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
          </svg>
        </div>
        <span className="cc-header-title">AgentBox Command</span>
      </div>

      <div className="cc-header-right">
        <div className="cc-header-stat">
          <span className="cc-header-dot cc-header-dot-green" />
          <span style={{ color: '#e2e8f0', fontWeight: 700, fontSize: '14px' }}>{agentsActive}</span>
          <span>agents active</span>
        </div>
        <div className="cc-header-stat">
          <span className="cc-header-dot cc-header-dot-blue" />
          <span style={{ color: '#e2e8f0', fontWeight: 600 }}>{tasksRunning}</span>
          <span>running</span>
        </div>
        <div className="cc-header-stat">
          <span className="cc-header-dot cc-header-dot-amber" />
          <span style={{ color: '#e2e8f0', fontWeight: 600 }}>{sessionsRunning}</span>
          <span>sessions</span>
        </div>
        <div className="cc-header-time">{formatTime(currentTime)}</div>
      </div>
    </div>
  )
}
