import { useState, useEffect } from 'react'
import { AlertCircle } from 'lucide-react'
import { useSessions, useAgents } from '../../hooks'
import { DashboardHeader } from './DashboardHeader'
import { StatsRow } from './StatsRow'
import { SessionsPanel } from './SessionsPanel'
import { AgentsPanel } from './AgentsPanel'
import { ActivityPanel } from './ActivityPanel'
import CreateSessionModal from '../CreateSessionModal'

export default function Dashboard() {
  const [showCreate, setShowCreate] = useState(false)
  const [currentTime, setCurrentTime] = useState(new Date())

  // 使用 React Query hooks
  const {
    data: sessions = [],
    isLoading: sessionsLoading,
    isFetching: sessionsFetching,
    error: sessionsError,
    refetch: refetchSessions,
  } = useSessions()

  const { data: agents = [], isLoading: agentsLoading } = useAgents()

  // 时钟更新
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  // 计算统计数据
  const stats = {
    total: sessions.length,
    running: sessions.filter(s => s.status === 'running').length,
    stopped: sessions.filter(s => s.status === 'stopped').length,
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Top Navigation */}
      <DashboardHeader
        totalSessions={stats.total}
        runningSessions={stats.running}
        currentTime={currentTime}
      />

      <div className="flex-1 overflow-hidden p-6 flex flex-col">
        {/* Error Banner */}
        {sessionsError && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3 flex-shrink-0">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {sessionsError instanceof Error ? sessionsError.message : '加载失败'}
            </span>
          </div>
        )}

        {/* Stats Row */}
        <StatsRow
          totalSessions={stats.total}
          runningSessions={stats.running}
          stoppedSessions={stats.stopped}
          totalAgents={agents.length}
        />

        {/* Three Column Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0">
          {/* Left Column: Sessions */}
          <SessionsPanel
            sessions={sessions}
            isLoading={sessionsLoading}
            isFetching={sessionsFetching}
            onRefresh={() => refetchSessions()}
            onCreateClick={() => setShowCreate(true)}
          />

          {/* Middle Column: Agents */}
          <AgentsPanel agents={agents} isLoading={agentsLoading} />

          {/* Right Column: Activity Log */}
          <ActivityPanel sessions={sessions} />
        </div>
      </div>

      {/* Create Modal */}
      {showCreate && (
        <CreateSessionModal
          agents={agents}
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false)
            refetchSessions()
          }}
        />
      )}
    </div>
  )
}
