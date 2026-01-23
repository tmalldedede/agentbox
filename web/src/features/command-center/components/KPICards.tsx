import type { DashboardStats } from '@/types'

interface Props {
  stats: DashboardStats | null
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) return (num / 1_000_000).toFixed(1) + 'M'
  if (num >= 1_000) return (num / 1_000).toFixed(1) + 'K'
  return num.toString()
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return seconds.toFixed(0) + 's'
  if (seconds < 3600) return (seconds / 60).toFixed(1) + 'm'
  return (seconds / 3600).toFixed(1) + 'h'
}

export function KPICards({ stats }: Props) {
  const cards = [
    {
      label: 'LIVE TASKS',
      value: formatNumber(stats?.tasks.total ?? 0),
      sub: `+${stats?.tasks.today ?? 0} today`,
      color: '#6366f1',
      trendUp: (stats?.tasks.today ?? 0) > 0,
    },
    {
      label: 'TOKENS USED',
      value: formatNumber(stats?.tokens.total_tokens ?? 0),
      sub: `${formatNumber(stats?.tokens.total_input ?? 0)} in / ${formatNumber(stats?.tokens.total_output ?? 0)} out`,
      color: '#8b5cf6',
    },
    {
      label: 'ACTIVE SESSIONS',
      value: (stats?.sessions.running ?? 0).toString(),
      sub: `${stats?.sessions.creating ?? 0} creating`,
      color: '#3b82f6',
    },
    {
      label: 'SUCCESS RATE',
      value: (stats?.tasks.success_rate ?? 0).toFixed(1) + '%',
      sub: stats?.tasks.success_rate && stats.tasks.success_rate >= 90 ? 'Healthy' : 'Needs attention',
      color: '#10b981',
      trendUp: (stats?.tasks.success_rate ?? 0) >= 90,
    },
    {
      label: 'AVG DURATION',
      value: formatDuration(stats?.tasks.avg_duration_seconds ?? 0),
      sub: `${stats?.containers.running ?? 0} containers`,
      color: '#f59e0b',
    },
  ]

  return (
    <div className="cc-kpi-grid cc-animate-in" style={{ animationDelay: '0.1s' }}>
      {cards.map((card, i) => (
        <div
          key={card.label}
          className="cc-kpi-card"
          style={{ '--kpi-color': card.color, animationDelay: `${i * 0.05}s` } as React.CSSProperties}
        >
          <div className="cc-kpi-label">{card.label}</div>
          <div className="cc-kpi-value" style={{ color: card.color }}>{card.value}</div>
          <div className={`cc-kpi-sub ${card.trendUp ? 'cc-kpi-trend-up' : ''}`}>
            {card.trendUp && (
              <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
                <path d="M5 1L9 6H1L5 1Z" />
              </svg>
            )}
            {card.sub}
          </div>
        </div>
      ))}
    </div>
  )
}
