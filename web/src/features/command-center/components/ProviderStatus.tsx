import type { DashboardProviderInfo } from '@/types'

interface Props {
  providers: DashboardProviderInfo[]
}

function getProviderInitial(name: string): string {
  return name.slice(0, 2).toUpperCase()
}

function getStatusLabel(status: string): { text: string; color: string } {
  switch (status) {
    case 'online': return { text: 'Online', color: '#6ee7b7' }
    case 'degraded': return { text: 'Degraded', color: '#fbbf24' }
    default: return { text: 'Offline', color: '#64748b' }
  }
}

export function ProviderStatus({ providers }: Props) {
  if (providers.length === 0) return null

  const onlineCount = providers.filter(p => p.status === 'online').length

  return (
    <div className="cc-panel cc-animate-in" style={{ animationDelay: '0.35s' }}>
      <div className="cc-panel-header">
        <span className="cc-panel-title">Provider Status</span>
        <span className="cc-panel-badge">
          <span style={{ color: '#6ee7b7' }}>{onlineCount} online</span>
          <span>/ {providers.length} total</span>
        </span>
      </div>
      <div className="cc-provider-grid">
        {providers.map(provider => {
          const statusInfo = getStatusLabel(provider.status)
          return (
            <div key={provider.id} className="cc-provider-card">
              <div
                className="cc-provider-icon"
                style={{
                  background: provider.icon_color
                    ? `${provider.icon_color}20`
                    : 'rgba(99, 102, 241, 0.1)',
                  color: provider.icon_color || '#a5b4fc',
                }}
              >
                {provider.icon || getProviderInitial(provider.name)}
              </div>
              <div className="cc-provider-info">
                <div className="cc-provider-name">{provider.name}</div>
                <div className="cc-provider-status">
                  <span className={`cc-provider-dot cc-provider-dot-${provider.status}`} />
                  <span style={{ color: statusInfo.color }}>{statusInfo.text}</span>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
