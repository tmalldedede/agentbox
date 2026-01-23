import type { DashboardRecentTask } from '@/types'

interface Props {
  tasks: DashboardRecentTask[]
}

function getEventType(status: string): { label: string; className: string } {
  switch (status) {
    case 'completed': return { label: 'COMPLETED', className: 'cc-event-type-completed' }
    case 'failed': return { label: 'FAILED', className: 'cc-event-type-failed' }
    case 'running': return { label: 'RUNNING', className: 'cc-event-type-running' }
    case 'queued': return { label: 'QUEUED', className: 'cc-event-type-queued' }
    case 'cancelled': return { label: 'CANCELLED', className: 'cc-event-type-completed' }
    default: return { label: status.toUpperCase(), className: 'cc-event-type-running' }
  }
}

function getAdapterAbbr(adapter: string): string {
  switch (adapter) {
    case 'claude-code': return 'CC'
    case 'codex': return 'CX'
    case 'opencode': return 'OC'
    default: return adapter.slice(0, 2).toUpperCase()
  }
}

function formatEventTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function truncatePrompt(prompt: string, max: number = 60): string {
  if (prompt.length <= max) return prompt
  return prompt.slice(0, max) + '...'
}

export function AgentEvents({ tasks }: Props) {
  // Sort by created_at descending to show newest first
  const sortedTasks = [...tasks].sort((a, b) =>
    new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="cc-panel cc-animate-in" style={{ animationDelay: '0.3s' }}>
      <div className="cc-panel-header">
        <span className="cc-panel-title">Agent Events</span>
        <span className="cc-panel-badge">{tasks.length} recent</span>
      </div>
      <div className="cc-panel-content">
        {sortedTasks.length === 0 ? (
          <div className="cc-empty">
            <div className="cc-empty-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
                <path d="M22 6l-10 7L2 6" />
              </svg>
            </div>
            <span>No recent events</span>
          </div>
        ) : (
          sortedTasks.slice(0, 20).map(task => {
            const eventType = getEventType(task.status)
            return (
              <div
                key={task.id}
                className={`cc-event-item cc-event-border-${task.status}`}
              >
                <div className="cc-event-header">
                  <div className="cc-event-agents">
                    <span style={{
                      padding: '1px 5px',
                      borderRadius: '3px',
                      background: 'rgba(99, 102, 241, 0.1)',
                      color: '#a5b4fc',
                      fontSize: '10px',
                    }}>
                      {getAdapterAbbr(task.adapter)}
                    </span>
                    <span className={`cc-event-type ${eventType.className}`}>
                      [{eventType.label}]
                    </span>
                  </div>
                  <span className="cc-event-time">{formatEventTime(task.created_at)}</span>
                </div>
                <div className="cc-event-message">
                  {truncatePrompt(task.prompt)}
                  {task.duration_seconds > 0 && (
                    <span style={{ color: '#475569', marginLeft: '6px' }}>
                      ({task.duration_seconds.toFixed(1)}s)
                    </span>
                  )}
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
