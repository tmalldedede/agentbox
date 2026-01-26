import { useEffect, useState, useCallback } from 'react'
import { KPISection } from './components/KPISection'
import { PerformanceOverview } from './components/PerformanceOverview'
import { ProviderStatus } from './components/ProviderStatus'
import { ActiveExecutions, ActiveExecution } from './components/ActiveExecutions'
import { TaskPipeline, PipelineTask } from './components/TaskPipeline'
import { TeamStats } from './components/TeamStats'
import { LiveActivity } from './components/LiveActivity'
import './styles.css'

// ==================== MOCK DATA ENGINE ====================
const AGENT_NAMES = ['Alpha-Sentinel', 'Beta-Codex', 'Gamma-Oracle', 'Delta-Forge', 'Omega-Scout']
const ADAPTERS = ['claude-code', 'codex', 'opencode'] as const
const TASK_PROMPTS = [
  'Analyzing security vulnerabilities in auth module',
  'Generating unit tests for payment service',
  'Refactoring database connection pooling',
  'Scanning container images for CVEs',
  'Building CI/CD pipeline configuration',
  'Deploying microservice to staging cluster',
  'Optimizing neural network inference latency',
  'Parsing CloudTrail logs for anomalies',
  'Synthesizing API documentation from source',
  'Running penetration test on REST endpoints',
  'Migrating legacy codebase to TypeScript',
  'Implementing real-time WebSocket notifications',
  'Training sentiment analysis model v2.4',
  'Auditing IAM policies for least privilege',
]
const AGENT_TEAMS = [
  { name: 'Code Generation', desc: 'Write, refactor, test code', icon: '⚡', color: '#22c55e' },
  { name: 'Security & Audit', desc: 'Scan, detect, remediate', icon: '🛡️', color: '#3b82f6' },
  { name: 'DevOps & Deploy', desc: 'Build, deploy, monitor', icon: '🚀', color: '#a78bfa' },
  { name: 'Research & Analysis', desc: 'Analyze data, generate reports', icon: '📊', color: '#eab308' },
]
const ACTIVITY_VERBS = ['Completed code review on', 'Deployed service to', 'Fixed vulnerability in', 'Generated tests for', 'Analyzed logs from']
const ACTIVITY_TARGETS = ['auth-service', 'payment-api', 'user-module', 'gateway-proxy', 'staging-cluster']

function randomInt(min: number, max: number) { return Math.floor(Math.random() * (max - min + 1)) + min }
function randomPick<T>(arr: T[]): T { return arr[Math.floor(Math.random() * arr.length)] }
function generateId() { return Math.random().toString(36).slice(2, 10) }

// ==================== MAIN COMPONENT ====================
export default function CommandCenter() {
  const [isFullscreen, setIsFullscreen] = useState(false)

  // --- STATE ---
  const [headerKpis, setHeaderKpis] = useState({ activeAgents: 5, tasksToday: 127, totalExecutions: 2840 })
  const [kpis, setKpis] = useState({ liveTasks: 7, tokens: 1284930, sessions: 12, successRate: 94.2, avgDuration: 23.5 })
  const [kpiHistory, setKpiHistory] = useState({
    tasks: Array.from({ length: 16 }, () => randomInt(3, 12)),
    tokens: Array.from({ length: 16 }, () => randomInt(800, 2000)),
    sessions: Array.from({ length: 16 }, () => randomInt(5, 18)),
    success: Array.from({ length: 16 }, () => randomInt(88, 99)),
    duration: Array.from({ length: 16 }, () => randomInt(10, 40)),
  })

  const [perfKpis, setPerfKpis] = useState({
    tasksCompleted: 1284, tokensUsed: 6400000, avgLatency: 23.5,
    tasksCompletedChange: 24, tokensChange: 18, latencyChange: -12,
  })
  const [perfChartData, setPerfChartData] = useState<number[]>(Array.from({ length: 20 }, (_, i) => Math.round(200 + i * 80 + Math.random() * 200)))

  const [providers, setProviders] = useState([
    { name: 'OpenAI', model: 'gpt-4o', tasks: 423, perDay: 89, status: 'online' as const },
    { name: 'Anthropic', model: 'claude-3.5', tasks: 312, perDay: 67, status: 'online' as const },
    { name: 'ZhiPu', model: 'glm-4-flash', tasks: 156, perDay: 42, status: 'online' as const },
    { name: 'DeepSeek', model: 'deepseek-v3', tasks: 98, perDay: 31, status: 'degraded' as const },
  ])

  const [executions, setExecutions] = useState<ActiveExecution[]>([
    { id: 'e1', prompt: 'Security audit on auth-service', agent: 'Alpha-Sentinel', adapter: 'claude-code', progress: 83, eta: '~2 min' },
    { id: 'e2', prompt: 'Generate API docs', agent: 'Beta-Codex', adapter: 'codex', progress: 45, eta: '~5 min' },
    { id: 'e3', prompt: 'Optimize DB queries', agent: 'Gamma-Oracle', adapter: 'opencode', progress: 67, eta: '~3 min' },
  ])

  const [pipeline, setPipeline] = useState<PipelineTask[]>([
    { id: 'p1', prompt: 'Refactoring auth middleware', status: 'running', adapter: 'claude-code', agent: 'Alpha-Sentinel', progress: 72, elapsed: '2m 34s' },
    { id: 'p2', prompt: 'Generating test coverage report', status: 'running', adapter: 'codex', agent: 'Beta-Codex', progress: 45, elapsed: '1m 12s' },
    { id: 'p3', prompt: 'Deploying to staging environment', status: 'queued', adapter: 'opencode', agent: 'Delta-Forge', progress: 0, elapsed: 'waiting' },
    { id: 'p4', prompt: 'Scanning npm dependencies', status: 'completed', adapter: 'claude-code', agent: 'Gamma-Oracle', progress: 100, elapsed: '45s' },
    { id: 'p5', prompt: 'Writing integration tests', status: 'running', adapter: 'codex', agent: 'Omega-Scout', progress: 31, elapsed: '3m 05s' },
  ])

  const [teamStats, setTeamStats] = useState(
    AGENT_TEAMS.map((t) => ({ ...t, agents: randomInt(3, 12), tasks: randomInt(50, 500) }))
  )

  const [activities, setActivities] = useState(() =>
    Array.from({ length: 8 }, (_, i) => ({
      id: generateId(),
      agent: randomPick(AGENT_NAMES),
      action: `${randomPick(ACTIVITY_VERBS)} ${randomPick(ACTIVITY_TARGETS)}`,
      adapter: randomPick([...ADAPTERS]),
      timestamp: i === 0 ? 'Just now' : `${i}m ago`,
    }))
  )

  // --- EFFECT LOOPS (Mock Data Updates) ---
  useEffect(() => {
    const t = setInterval(() => {
      // Update KPIs
      setKpis(prev => ({
        liveTasks: Math.max(1, prev.liveTasks + randomInt(-2, 3)),
        tokens: prev.tokens + randomInt(500, 5000),
        sessions: Math.max(3, prev.sessions + randomInt(-1, 2)),
        successRate: Math.min(99.9, Math.max(85, prev.successRate + (Math.random() - 0.4) * 2)),
        avgDuration: Math.max(5, prev.avgDuration + (Math.random() - 0.5) * 3),
      }))
      // History for Sparklines
      setKpiHistory(prev => ({
        tasks: [...prev.tasks.slice(1), randomInt(3, 12)],
        tokens: [...prev.tokens.slice(1), randomInt(800, 2000)],
        sessions: [...prev.sessions.slice(1), randomInt(5, 18)],
        success: [...prev.success.slice(1), randomInt(88, 99)],
        duration: [...prev.duration.slice(1), randomInt(10, 40)],
      }))
      // Header
      setHeaderKpis(prev => ({ ...prev, tasksToday: prev.tasksToday + randomInt(0, 2) }))
    }, 3000)
    return () => clearInterval(t)
  }, [])

  // Active Executions Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setExecutions(prev => prev.map(e => {
        const newProgress = Math.min(100, e.progress + randomInt(2, 8))
        if (newProgress >= 100) {
          return {
            id: generateId(),
            prompt: randomPick(TASK_PROMPTS),
            agent: randomPick(AGENT_NAMES),
            adapter: randomPick([...ADAPTERS]),
            progress: 5,
            eta: `~${randomInt(2, 12)} min`,
          }
        }
        return { ...e, progress: newProgress }
      }))
    }, 1500)
    return () => clearInterval(t)
  }, [])

  // Performance Chart Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setPerfChartData(prev => [...prev.slice(1), prev[prev.length - 1] + randomInt(-80, 120)])
      setPerfKpis(prev => ({
        ...prev,
        tasksCompleted: prev.tasksCompleted + randomInt(1, 5),
        tokensUsed: prev.tokensUsed + randomInt(10000, 50000),
        avgLatency: Math.max(8, prev.avgLatency + (Math.random() - 0.5) * 2),
      }))
    }, 4000)
    return () => clearInterval(t)
  }, [])

  // Provider Status Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setProviders(prev => prev.map(p => ({
        ...p,
        tasks: p.tasks + randomInt(0, 3),
        perDay: Math.max(10, p.perDay + randomInt(-2, 4)),
      })))
    }, 5000)
    return () => clearInterval(t)
  }, [])

  // Task Pipeline Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setPipeline(prev => prev.map(task => {
        if (task.status === 'running') {
          const newProgress = task.progress + randomInt(3, 12)
          if (newProgress >= 100) {
            return { ...task, status: 'completed' as const, progress: 100, elapsed: `${randomInt(30, 180)}s` }
          }
          const secs = parseInt(task.elapsed) || randomInt(10, 60)
          return { ...task, progress: newProgress, elapsed: `${Math.floor((secs + 3) / 60)}m ${(secs + 3) % 60}s` }
        }
        if (task.status === 'queued' && Math.random() > 0.7) {
          return { ...task, status: 'running' as const, progress: randomInt(5, 15), elapsed: '0s' }
        }
        if (task.status === 'completed' && Math.random() > 0.85) {
          return {
            id: generateId(),
            prompt: randomPick(TASK_PROMPTS),
            status: 'queued' as const,
            adapter: randomPick([...ADAPTERS]),
            agent: randomPick(AGENT_NAMES),
            progress: 0,
            elapsed: 'waiting',
          }
        }
        return task
      }))
    }, 2500)
    return () => clearInterval(t)
  }, [])

  // Team Stats Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setTeamStats(prev => prev.map(team => ({
        ...team,
        tasks: team.tasks + randomInt(0, 3),
        agents: Math.max(2, team.agents + (Math.random() > 0.8 ? randomInt(-1, 1) : 0)),
      })))
    }, 6000)
    return () => clearInterval(t)
  }, [])

  // Live Activity Feed Simulation
  useEffect(() => {
    const t = setInterval(() => {
      setActivities(prev => {
        const newItem = {
          id: generateId(),
          agent: randomPick(AGENT_NAMES),
          action: `${randomPick(ACTIVITY_VERBS)} ${randomPick(ACTIVITY_TARGETS)}`,
          adapter: randomPick([...ADAPTERS]),
          timestamp: 'Just now',
        }
        return [newItem, ...prev.slice(0, 7).map((a, i) => ({ ...a, timestamp: `${i + 1}m ago` }))]
      })
    }, 4500)
    return () => clearInterval(t)
  }, [])

  // Fullscreen Toggle
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) { document.documentElement.requestFullscreen(); setIsFullscreen(true) }
    else { document.exitFullscreen(); setIsFullscreen(false) }
  }, [])

  return (
    <div data-layout="fixed" className={`h-full bg-background text-foreground p-5 font-sans flex flex-col overflow-hidden ${isFullscreen ? 'fixed inset-0 z-50' : ''}`}>
      <div className="mx-auto flex w-full max-w-[1800px] flex-1 min-h-0 flex-col gap-4">

        {/* HEADER */}
        <header className="flex shrink-0 items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-emerald-500/20 bg-emerald-500/10 text-emerald-500 shadow-[0_0_15px_rgba(16,185,129,0.1)]">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><rect x="3" y="3" width="18" height="18" rx="4" /><path d="M9 12l2 2 4-4" /></svg>
            </div>
            <div>
              <h1 className="text-lg font-bold text-foreground tracking-tight">AgentBox Command</h1>
              <p className="text-[11px] font-medium text-muted-foreground">System Status: <span className="text-emerald-400">Operational</span></p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-3 rounded-lg border border-border bg-card/50 px-3 py-1.5 backdrop-blur-sm">
              <div className="flex h-7 w-7 items-center justify-center rounded bg-emerald-500/10 text-emerald-400">
                <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" /></svg>
              </div>
              <div>
                <div className="text-[9px] uppercase font-bold text-muted-foreground tracking-wider">Active Agents</div>
                <div className="text-base font-bold text-foreground leading-none tabular-nums">{headerKpis.activeAgents}</div>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border border-border bg-card/50 px-3 py-1.5 backdrop-blur-sm">
              <div className="flex h-7 w-7 items-center justify-center rounded bg-blue-500/10 text-blue-400">
                <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path d="M9 11l3 3L22 4" /><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" /></svg>
              </div>
              <div>
                <div className="text-[9px] uppercase font-bold text-muted-foreground tracking-wider">Tasks Today</div>
                <div className="text-base font-bold text-foreground leading-none tabular-nums">{headerKpis.tasksToday}</div>
              </div>
            </div>

            <button
              onClick={toggleFullscreen}
              className="flex h-9 w-9 items-center justify-center rounded-lg border border-border bg-card text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                {isFullscreen ? <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3" /> : <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />}
              </svg>
            </button>
          </div>
        </header>

        {/* KPI GRID */}
        <div className="shrink-0">
          <KPISection kpis={kpis} history={kpiHistory} />
        </div>

        {/* MAIN DASHBOARD GRID - fills remaining space */}
        <div className="flex-1 min-h-0 grid grid-cols-1 gap-4 lg:grid-cols-3 lg:grid-rows-2 relative cc-scanline">
          <div className="min-h-0"><PerformanceOverview kpis={perfKpis} chartData={perfChartData} /></div>
          <div className="min-h-0"><ProviderStatus providers={providers} /></div>
          <div className="min-h-0"><ActiveExecutions executions={executions} /></div>
          <div className="min-h-0"><TaskPipeline tasks={pipeline} /></div>
          <div className="min-h-0"><TeamStats teams={teamStats} /></div>
          <div className="min-h-0"><LiveActivity activities={activities} /></div>
        </div>
      </div>
    </div>
  )
}
