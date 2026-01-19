import { Sparkline } from './Sparkline'

interface StatCardProps {
  label: string
  value: string | number
  change?: string
  changeLabel?: string
  sparklineData?: number[]
  sparklineColor?: string
}

export function StatCard({
  label,
  value,
  change,
  changeLabel,
  sparklineData,
  sparklineColor,
}: StatCardProps) {
  const isPositive = change?.startsWith('+')
  const isNegative = change?.startsWith('-')

  return (
    <div className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="stat-value">{value}</div>
          {change && (
            <div
              className={`stat-change ${isPositive ? 'stat-change-positive' : ''} ${isNegative ? 'stat-change-negative' : ''}`}
            >
              <span>{change}</span>
              {changeLabel && <span className="text-muted ml-1">{changeLabel}</span>}
            </div>
          )}
        </div>
        {sparklineData && sparklineData.length > 0 && (
          <div className="w-24">
            <Sparkline data={sparklineData} color={sparklineColor} />
          </div>
        )}
      </div>
    </div>
  )
}
