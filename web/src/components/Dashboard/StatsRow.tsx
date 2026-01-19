import { useMemo } from 'react'
import { StatCard } from './StatCard'
import { useLanguage } from '../../contexts/LanguageContext'

interface StatsRowProps {
  totalSessions: number
  runningSessions: number
  stoppedSessions: number
  totalAgents: number
}

export function StatsRow({
  totalSessions,
  runningSessions,
  stoppedSessions,
  totalAgents,
}: StatsRowProps) {
  const { t } = useLanguage()

  // 生成模拟的 sparkline 数据
  const sparklineData = useMemo(
    () => Array.from({ length: 20 }, () => Math.floor(Math.random() * 100)),
    []
  )

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-6 flex-shrink-0">
      <StatCard
        label={t('totalSessions')}
        value={totalSessions}
        change={`+${runningSessions}`}
        changeLabel="active"
        sparklineData={sparklineData}
        sparklineColor="#10b981"
      />
      <StatCard label={t('runningSessions')} value={runningSessions} />
      <StatCard label={t('stoppedSessions')} value={stoppedSessions} />
      <StatCard label={t('availableAgents')} value={totalAgents} />
      <StatCard label={t('successRate')} value="100%" />
    </div>
  )
}
