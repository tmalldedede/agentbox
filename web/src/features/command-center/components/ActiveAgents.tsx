import type { DashboardAgentDetail } from '@/types'

interface Props {
  agents: DashboardAgentDetail[]
}

function getAdapterAbbr(adapter: string): string {
  switch (adapter) {
    case 'claude-code': return 'CC'
    case 'codex': return 'CX'
    case 'opencode': return 'OC'
    default: return adapter.slice(0, 2).toUpperCase()
  }
}

function getAdapterClass(adapter: string): string {
  switch (adapter) {
    case 'claude-code': return 'cc-agent-avatar-cc'
    case 'codex': return 'cc-agent-avatar-cx'
    case 'opencode': return 'cc-agent-avatar-oc'
    default: return 'cc-agent-avatar-cc'
  }
}

export function ActiveAgents({ agents }: Props) {
  const activeAgents = agents.filter(a => a.status === 'active')
  const totalRunning = agents.reduce((sum, a) => sum + a.running, 0)

  return (
    <div className="cc-panel cc-animate-in" style={{ animationDelay: '0.2s' }}>
      <div className="cc-panel-header">
        <span className="cc-panel-title">Active Agents</span>
        <span className="cc-panel-badge">
          {activeAgents.length} active
          {totalRunning > 0 && (
            <span style={{ color: '#60a5fa' }}>{totalRunning} running</span>
          )}
        </span>
      </div>
      <div className="cc-panel-content">
        {agents.length === 0 ? (
          <div className="cc-empty">
            <div className="cc-empty-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <path d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
            </div>
            <span>No agents configured</span>
          </div>
        ) : (
          agents.map(agent => {
            const total = agent.running + agent.queued + agent.completed + agent.failed
            const progressPercent = total > 0
              ? Math.min(((agent.running + agent.completed) / Math.max(total, 1)) * 100, 100)
              : 0

            return (
              <div key={agent.id} className="cc-agent-card">
                <div className="cc-agent-card-header">
                  <div className="cc-agent-card-left">
                    <div className={`cc-agent-avatar ${getAdapterClass(agent.adapter)}`}>
                      {getAdapterAbbr(agent.adapter)}
                    </div>
                    <div>
                      <div className="cc-agent-name">{agent.name}</div>
                      <div style={{ fontSize: '10px', color: '#64748b', marginTop: '1px' }}>
                        {agent.model || agent.adapter}
                      </div>
                    </div>
                  </div>
                  <div className={`cc-agent-status ${agent.status === 'active' ? 'cc-agent-status-active' : 'cc-agent-status-inactive'}`}>
                    <span className={`cc-header-dot ${agent.status === 'active' ? 'cc-header-dot-green' : ''}`}
                      style={{ width: '5px', height: '5px' }} />
                    {agent.status === 'active' ? 'Active' : 'Inactive'}
                  </div>
                </div>

                <div className="cc-agent-stats">
                  <span>
                    <span className="cc-agent-stat-value cc-agent-stat-running">{agent.running}</span>
                    running
                  </span>
                  <span>
                    <span className="cc-agent-stat-value cc-agent-stat-queued">{agent.queued}</span>
                    queued
                  </span>
                </div>

                <div className="cc-agent-progress">
                  <div
                    className="cc-agent-progress-bar"
                    style={{ width: `${progressPercent}%` }}
                  />
                </div>

                <div className="cc-agent-footer">
                  <span>{agent.completed} completed</span>
                  {agent.failed > 0 && (
                    <span style={{ color: '#fca5a5' }}>{agent.failed} failed</span>
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
