import type { DashboardRecentTask } from '@/types'

interface Props {
  tasks: DashboardRecentTask[]
  stats?: {
    total: number
    today: number
    by_status: Record<string, number>
    avg_duration_seconds: number
    success_rate: number
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'running':
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <circle cx="12" cy="12" r="10" />
          <path d="M12 6v6l4 2" />
        </svg>
      )
    case 'queued':
    case 'pending':
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <circle cx="12" cy="12" r="10" />
          <path d="M12 8v4M12 16h.01" />
        </svg>
      )
    case 'completed':
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <path d="M20 6L9 17l-5-5" />
        </svg>
      )
    case 'failed':
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <path d="M18 6L6 18M6 6l12 12" />
        </svg>
      )
    case 'cancelled':
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <circle cx="12" cy="12" r="10" />
          <path d="M4.93 4.93l14.14 14.14" />
        </svg>
      )
    default:
      return null
  }
}

function getAdapterTag(adapter: string) {
  const colors: Record<string, { bg: string; color: string }> = {
    'claude-code': { bg: 'rgba(139, 92, 246, 0.15)', color: '#a78bfa' },
    'codex': { bg: 'rgba(16, 185, 129, 0.15)', color: '#6ee7b7' },
    'opencode': { bg: 'rgba(59, 130, 246, 0.15)', color: '#93c5fd' },
  }
  const c = colors[adapter] || { bg: 'rgba(100, 116, 139, 0.15)', color: '#94a3b8' }
  const label = adapter === 'claude-code' ? 'CC' : adapter === 'codex' ? 'CX' : 'OC'
  return (
    <span className="cc-task-agent-tag" style={{ background: c.bg, color: c.color }}>
      {label}
    </span>
  )
}

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return ''
  if (seconds < 60) return `${seconds.toFixed(0)}s`
  return `${(seconds / 60).toFixed(1)}m`
}

function formatTimeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = (now - then) / 1000

  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export function TaskPipeline({ tasks, stats }: Props) {
  const running = stats?.by_status?.running ?? 0
  const queued = (stats?.by_status?.queued ?? 0) + (stats?.by_status?.pending ?? 0)

  // Sort: running first, then queued, then recently completed/failed
  const sortedTasks = [...tasks].sort((a, b) => {
    const priority: Record<string, number> = { running: 0, queued: 1, pending: 2, completed: 3, failed: 4, cancelled: 5 }
    return (priority[a.status] ?? 9) - (priority[b.status] ?? 9)
  })

  return (
    <div className="cc-panel cc-animate-in" style={{ animationDelay: '0.25s' }}>
      <div className="cc-panel-header">
        <span className="cc-panel-title">Task Pipeline</span>
        <span className="cc-panel-badge">
          {queued > 0 && <span style={{ color: '#fbbf24' }}>{queued} queued</span>}
          {running > 0 && <span style={{ color: '#60a5fa' }}>{running} running</span>}
        </span>
      </div>
      <div className="cc-panel-content">
        {sortedTasks.length === 0 ? (
          <div className="cc-empty">
            <div className="cc-empty-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <rect x="3" y="3" width="18" height="18" rx="3" />
                <path d="M9 12h6M12 9v6" />
              </svg>
            </div>
            <span>No tasks in pipeline</span>
          </div>
        ) : (
          sortedTasks.slice(0, 15).map(task => (
            <div key={task.id} className="cc-task-item">
              <div className={`cc-task-icon cc-task-icon-${task.status}`}>
                {getStatusIcon(task.status)}
              </div>
              <div className="cc-task-info">
                <div className="cc-task-prompt">{task.prompt}</div>
                <div className="cc-task-meta">
                  {getAdapterTag(task.adapter)}
                  <span>{task.agent_name || task.agent_id}</span>
                  {task.duration_seconds > 0 && (
                    <span>{formatDuration(task.duration_seconds)}</span>
                  )}
                  <span>{formatTimeAgo(task.created_at)}</span>
                </div>
              </div>
              <span className={`cc-task-status-badge cc-badge-${task.status}`}>
                {task.status}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
