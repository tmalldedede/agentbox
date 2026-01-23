import { useEffect, useState, useCallback, useRef } from 'react'
import './styles.css'

// ==================== MOCK DATA ====================
const AGENT_NAMES = ['Alpha-Sentinel', 'Beta-Codex', 'Gamma-Oracle', 'Delta-Forge', 'Omega-Scout']
const MODELS = ['gpt-4o', 'claude-3.5', 'glm-4-flash', 'deepseek-v3', 'qwen-turbo']
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
  'Compiling threat intelligence report',
]

const AGENT_TEAMS = [
  { name: 'Code Generation', desc: 'Write, refactor, test code', icon: '⚡', color: '#22c55e' },
  { name: 'Security & Audit', desc: 'Scan, detect, remediate', icon: '🛡️', color: '#3b82f6' },
  { name: 'DevOps & Deploy', desc: 'Build, deploy, monitor', icon: '🚀', color: '#a78bfa' },
  { name: 'Research & Analysis', desc: 'Analyze data, generate reports', icon: '📊', color: '#eab308' },
]

const ACTIVITY_VERBS = [
  'Completed code review on',
  'Deployed service to',
  'Fixed vulnerability in',
  'Generated tests for',
  'Analyzed logs from',
  'Scanned dependencies in',
  'Optimized queries in',
  'Refactored module',
  'Detected anomaly in',
  'Updated pipeline for',
]

const ACTIVITY_TARGETS = [
  'auth-service', 'payment-api', 'user-module', 'gateway-proxy',
  'staging-cluster', 'prod-db', 'redis-cache', 'ml-pipeline',
  'event-bus', 'config-server', 'metrics-collector', 'task-scheduler',
]

type TaskStatus = 'running' | 'queued' | 'completed' | 'failed'

interface PipelineTask {
  id: string
  prompt: string
  status: TaskStatus
  adapter: typeof ADAPTERS[number]
  agent: string
  progress: number
  elapsed: string
}

interface ActiveExecution {
  id: string
  prompt: string
  agent: string
  adapter: typeof ADAPTERS[number]
  progress: number
  eta: string
}

interface ActivityItem {
  id: string
  agent: string
  action: string
  adapter: typeof ADAPTERS[number]
  timestamp: string
}

function randomInt(min: number, max: number) { return Math.floor(Math.random() * (max - min + 1)) + min }
function randomPick<T>(arr: T[]): T { return arr[Math.floor(Math.random() * arr.length)] }
function generateId() { return Math.random().toString(36).slice(2, 10) }

// ==================== ANIMATED NUMBER ====================
function AnimatedNumber({ value, duration = 1200 }: { value: number; duration?: number }) {
  const [display, setDisplay] = useState(0)
  const prevRef = useRef(0)

  useEffect(() => {
    const start = prevRef.current
    const diff = value - start
    if (diff === 0) return
    const startTime = performance.now()
    const animate = (now: number) => {
      const elapsed = now - startTime
      const progress = Math.min(elapsed / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      setDisplay(Math.round(start + diff * eased))
      if (progress < 1) requestAnimationFrame(animate)
      else prevRef.current = value
    }
    requestAnimationFrame(animate)
  }, [value, duration])

  return <>{display.toLocaleString()}</>
}

// ==================== SPARKLINE ====================
function Sparkline({ data, color, width = 100, height = 32 }: { data: number[]; color: string; width?: number; height?: number }) {
  const max = Math.max(...data, 1)
  const min = Math.min(...data, 0)
  const range = max - min || 1
  const points = data.map((v, i) => ({
    x: (i / (data.length - 1)) * width,
    y: height - ((v - min) / range) * (height - 4) - 2,
  }))

  // Smooth bezier path
  const linePath = points.map((p, i) => {
    if (i === 0) return `M${p.x},${p.y}`
    const prev = points[i - 1]
    const cpx1 = prev.x + (p.x - prev.x) * 0.4
    const cpx2 = prev.x + (p.x - prev.x) * 0.6
    return `C${cpx1},${prev.y} ${cpx2},${p.y} ${p.x},${p.y}`
  }).join(' ')

  const areaPath = `${linePath} L${width},${height} L0,${height} Z`

  return (
    <svg width={width} height={height} className="cc-sparkline">
      <defs>
        <linearGradient id={`sp-${color.replace('#', '')}-${width}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.25" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={areaPath} fill={`url(#sp-${color.replace('#', '')}-${width})`} />
      <path d={linePath} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
      <circle cx={points[points.length - 1].x} cy={points[points.length - 1].y} r="2.5" fill={color}>
        <animate attributeName="opacity" values="1;0.4;1" dur="2s" repeatCount="indefinite" />
      </circle>
    </svg>
  )
}

// ==================== PERFORMANCE CHART ====================
function PerformanceChart({ data, color }: { data: number[]; color: string }) {
  const width = 440
  const height = 140
  const pad = { top: 10, right: 10, bottom: 24, left: 40 }
  const cw = width - pad.left - pad.right
  const ch = height - pad.top - pad.bottom

  const max = Math.max(...data, 10)
  const points = data.map((v, i) => ({
    x: pad.left + (i / (data.length - 1)) * cw,
    y: pad.top + ch - (v / max) * ch,
  }))

  const linePath = points.map((p, i) => {
    if (i === 0) return `M${p.x},${p.y}`
    const prev = points[i - 1]
    const cpx1 = prev.x + (p.x - prev.x) * 0.4
    const cpx2 = prev.x + (p.x - prev.x) * 0.6
    return `C${cpx1},${prev.y} ${cpx2},${p.y} ${p.x},${p.y}`
  }).join(' ')

  const areaPath = `${linePath} L${pad.left + cw},${pad.top + ch} L${pad.left},${pad.top + ch} Z`
  const yTicks = [0, Math.round(max * 0.33), Math.round(max * 0.66), max]
  const months = ['Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan']

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="cc-perf-chart">
      <defs>
        <linearGradient id="perf-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.2" />
          <stop offset="100%" stopColor={color} stopOpacity="0.02" />
        </linearGradient>
      </defs>
      {yTicks.map((val, i) => {
        const y = pad.top + ch - (val / max) * ch
        return (
          <g key={i}>
            <line x1={pad.left} y1={y} x2={pad.left + cw} y2={y} stroke="rgba(255,255,255,0.04)" />
            <text x={pad.left - 6} y={y + 3} fill="#4b5563" fontSize="9" textAnchor="end">{val >= 1000 ? `${Math.round(val / 1000)}k` : val}</text>
          </g>
        )
      })}
      {months.map((m, i) => {
        const x = pad.left + (i / (months.length - 1)) * cw
        return <text key={m} x={x} y={height - 6} fill="#4b5563" fontSize="9" textAnchor="middle">{m}</text>
      })}
      <path d={areaPath} fill="url(#perf-grad)" />
      <path d={linePath} fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" />
      <circle cx={points[points.length - 1].x} cy={points[points.length - 1].y} r="3.5" fill="#0a0f1a" stroke={color} strokeWidth="2">
        <animate attributeName="r" values="3;5;3" dur="2.5s" repeatCount="indefinite" />
      </circle>
    </svg>
  )
}

// ==================== MAIN COMPONENT ====================
export default function CommandCenter() {
  const [isFullscreen, setIsFullscreen] = useState(false)

  // Header KPIs
  const [headerKpis, setHeaderKpis] = useState({
    activeAgents: 5,
    tasksToday: 127,
    totalExecutions: 2840,
  })

  // KPI Cards (5 cards with sparklines)
  const [kpis, setKpis] = useState({ liveTasks: 7, tokens: 1284930, sessions: 12, successRate: 94.2, avgDuration: 23.5 })
  const [kpiHistory, setKpiHistory] = useState({
    tasks: Array.from({ length: 16 }, () => randomInt(3, 12)),
    tokens: Array.from({ length: 16 }, () => randomInt(800, 2000)),
    sessions: Array.from({ length: 16 }, () => randomInt(5, 18)),
    success: Array.from({ length: 16 }, () => randomInt(88, 99)),
    duration: Array.from({ length: 16 }, () => randomInt(10, 40)),
  })

  // Performance panel
  const [perfKpis, setPerfKpis] = useState({
    tasksCompleted: 1284,
    tokensUsed: 6400000,
    avgLatency: 23.5,
    tasksCompletedChange: 24,
    tokensChange: 18,
    latencyChange: -12,
  })
  const [perfChartData, setPerfChartData] = useState<number[]>(
    Array.from({ length: 20 }, (_, i) => Math.round(200 + i * 80 + Math.random() * 200))
  )

  // Provider Status
  const [providers, setProviders] = useState([
    { name: 'OpenAI', model: 'gpt-4o', tasks: 423, perDay: 89, status: 'online' as const },
    { name: 'Anthropic', model: 'claude-3.5', tasks: 312, perDay: 67, status: 'online' as const },
    { name: 'ZhiPu', model: 'glm-4-flash', tasks: 156, perDay: 42, status: 'online' as const },
    { name: 'DeepSeek', model: 'deepseek-v3', tasks: 98, perDay: 31, status: 'degraded' as const },
  ])

  // Active Executions
  const [executions, setExecutions] = useState<ActiveExecution[]>([
    { id: 'e1', prompt: 'Security audit on auth-service', agent: 'Alpha-Sentinel', adapter: 'claude-code', progress: 83, eta: '~2 min' },
    { id: 'e2', prompt: 'Generate API docs', agent: 'Beta-Codex', adapter: 'codex', progress: 45, eta: '~5 min' },
    { id: 'e3', prompt: 'Optimize DB queries', agent: 'Gamma-Oracle', adapter: 'opencode', progress: 67, eta: '~3 min' },
    { id: 'e4', prompt: 'Container vulnerability scan', agent: 'Delta-Forge', adapter: 'claude-code', progress: 28, eta: '~8 min' },
  ])

  // Pipeline tasks
  const [pipeline, setPipeline] = useState<PipelineTask[]>([
    { id: 'p1', prompt: 'Refactoring auth middleware', status: 'running', adapter: 'claude-code', agent: 'Alpha-Sentinel', progress: 72, elapsed: '2m 34s' },
    { id: 'p2', prompt: 'Generating test coverage report', status: 'running', adapter: 'codex', agent: 'Beta-Codex', progress: 45, elapsed: '1m 12s' },
    { id: 'p3', prompt: 'Deploying to staging environment', status: 'queued', adapter: 'opencode', agent: 'Delta-Forge', progress: 0, elapsed: 'waiting' },
    { id: 'p4', prompt: 'Scanning npm dependencies', status: 'completed', adapter: 'claude-code', agent: 'Gamma-Oracle', progress: 100, elapsed: '45s' },
    { id: 'p5', prompt: 'Writing integration tests', status: 'running', adapter: 'codex', agent: 'Omega-Scout', progress: 31, elapsed: '3m 05s' },
    { id: 'p6', prompt: 'Fixing CORS configuration', status: 'failed', adapter: 'opencode', agent: 'Alpha-Sentinel', progress: 60, elapsed: '1m 50s' },
  ])

  // Agent Teams
  const [teamStats, setTeamStats] = useState(
    AGENT_TEAMS.map((t, i) => ({ ...t, agents: randomInt(3, 12), tasks: randomInt(50, 500) }))
  )

  // Live Activity
  const [activities, setActivities] = useState<ActivityItem[]>(() =>
    Array.from({ length: 8 }, (_, i) => ({
      id: generateId(),
      agent: randomPick(AGENT_NAMES),
      action: `${randomPick(ACTIVITY_VERBS)} ${randomPick(ACTIVITY_TARGETS)}`,
      adapter: randomPick([...ADAPTERS]),
      timestamp: i === 0 ? 'Just now' : `${i}m ago`,
    }))
  )

  // ==================== TIMERS ====================
  // Header KPIs
  useEffect(() => {
    const t = setInterval(() => {
      setHeaderKpis(prev => ({
        activeAgents: Math.max(3, prev.activeAgents + randomInt(-1, 1)),
        tasksToday: prev.tasksToday + randomInt(0, 3),
        totalExecutions: prev.totalExecutions + randomInt(1, 5),
      }))
    }, 4000)
    return () => clearInterval(t)
  }, [])

  // KPI Cards timer
  useEffect(() => {
    const t = setInterval(() => {
      setKpis(prev => ({
        liveTasks: Math.max(1, prev.liveTasks + randomInt(-2, 3)),
        tokens: prev.tokens + randomInt(500, 5000),
        sessions: Math.max(3, prev.sessions + randomInt(-1, 2)),
        successRate: Math.min(99.9, Math.max(85, prev.successRate + (Math.random() - 0.4) * 2)),
        avgDuration: Math.max(5, prev.avgDuration + (Math.random() - 0.5) * 3),
      }))
      setKpiHistory(prev => ({
        tasks: [...prev.tasks.slice(1), randomInt(3, 12)],
        tokens: [...prev.tokens.slice(1), randomInt(800, 2000)],
        sessions: [...prev.sessions.slice(1), randomInt(5, 18)],
        success: [...prev.success.slice(1), randomInt(88, 99)],
        duration: [...prev.duration.slice(1), randomInt(10, 40)],
      }))
    }, 3000)
    return () => clearInterval(t)
  }, [])

  // Performance KPIs
  useEffect(() => {
    const t = setInterval(() => {
      setPerfKpis(prev => ({
        tasksCompleted: prev.tasksCompleted + randomInt(1, 5),
        tokensUsed: prev.tokensUsed + randomInt(10000, 80000),
        avgLatency: Math.max(8, prev.avgLatency + (Math.random() - 0.5) * 3),
        tasksCompletedChange: Math.min(50, Math.max(5, prev.tasksCompletedChange + randomInt(-2, 3))),
        tokensChange: Math.min(40, Math.max(5, prev.tokensChange + randomInt(-2, 2))),
        latencyChange: Math.max(-30, Math.min(5, prev.latencyChange + randomInt(-2, 2))),
      }))
      setPerfChartData(prev => [...prev.slice(1), prev[prev.length - 1] + randomInt(-100, 200)])
    }, 3000)
    return () => clearInterval(t)
  }, [])

  // Provider updates
  useEffect(() => {
    const t = setInterval(() => {
      setProviders(prev => prev.map(p => ({
        ...p,
        tasks: p.tasks + randomInt(0, 5),
        perDay: Math.max(10, p.perDay + randomInt(-3, 5)),
      })))
    }, 5000)
    return () => clearInterval(t)
  }, [])

  // Execution progress
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
            progress: randomInt(5, 20),
            eta: `~${randomInt(2, 12)} min`,
          }
        }
        return { ...e, progress: newProgress }
      }))
    }, 2000)
    return () => clearInterval(t)
  }, [])

  // Pipeline updates
  useEffect(() => {
    const t = setInterval(() => {
      setPipeline(prev => {
        const next = [...prev]
        const idx = randomInt(0, next.length - 1)
        const item = next[idx]
        if (item.status === 'running') {
          const newProg = Math.min(100, item.progress + randomInt(5, 15))
          if (newProg >= 100) {
            next[idx] = { ...item, status: 'completed', progress: 100 }
          } else {
            next[idx] = { ...item, progress: newProg }
          }
        } else if (item.status === 'queued') {
          next[idx] = { ...item, status: 'running', progress: randomInt(5, 15) }
        } else if (item.status === 'completed' || item.status === 'failed') {
          next[idx] = {
            id: generateId(),
            prompt: randomPick(TASK_PROMPTS),
            status: 'queued',
            adapter: randomPick([...ADAPTERS]),
            agent: randomPick(AGENT_NAMES),
            progress: 0,
            elapsed: 'waiting',
          }
        }
        return next
      })
    }, 3000)
    return () => clearInterval(t)
  }, [])

  // Team stats
  useEffect(() => {
    const t = setInterval(() => {
      setTeamStats(prev => prev.map(team => ({
        ...team,
        agents: Math.max(2, team.agents + randomInt(-1, 1)),
        tasks: team.tasks + randomInt(0, 5),
      })))
    }, 4000)
    return () => clearInterval(t)
  }, [])

  // Live activity
  useEffect(() => {
    const t = setInterval(() => {
      setActivities(prev => {
        const newItem: ActivityItem = {
          id: generateId(),
          agent: randomPick(AGENT_NAMES),
          action: `${randomPick(ACTIVITY_VERBS)} ${randomPick(ACTIVITY_TARGETS)}`,
          adapter: randomPick([...ADAPTERS]),
          timestamp: 'Just now',
        }
        return [newItem, ...prev.slice(0, 7).map((a, i) => ({ ...a, timestamp: `${i + 1}m ago` }))]
      })
    }, 4000)
    return () => clearInterval(t)
  }, [])

  // Fullscreen
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) { document.documentElement.requestFullscreen(); setIsFullscreen(true) }
    else { document.exitFullscreen(); setIsFullscreen(false) }
  }, [])
  useEffect(() => {
    const handler = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', handler)
    return () => document.removeEventListener('fullscreenchange', handler)
  }, [])

  // Helpers
  function adapterLabel(adapter: string) {
    const config: Record<string, { bg: string; color: string; label: string }> = {
      'claude-code': { bg: 'rgba(167,139,250,0.12)', color: '#a78bfa', label: 'Claude Code' },
      'codex': { bg: 'rgba(34,197,94,0.12)', color: '#4ade80', label: 'Codex' },
      'opencode': { bg: 'rgba(96,165,250,0.12)', color: '#60a5fa', label: 'OpenCode' },
    }
    const c = config[adapter] || config['opencode']
    return <span className="cc-tag" style={{ background: c.bg, color: c.color }}>{c.label}</span>
  }

  function statusColor(status: string) {
    switch (status) {
      case 'running': return '#3b82f6'
      case 'queued': return '#eab308'
      case 'completed': return '#22c55e'
      case 'failed': return '#ef4444'
      case 'online': return '#22c55e'
      case 'degraded': return '#eab308'
      default: return '#6b7280'
    }
  }

  const totalAgents = teamStats.reduce((s, t) => s + t.agents, 0)
  const totalTeamTasks = teamStats.reduce((s, t) => s + t.tasks, 0)
  const activeTeamAgents = totalAgents - randomInt(0, 3)

  return (
    <div className={`cc-container ${isFullscreen ? 'cc-fullscreen' : ''}`}>
      <div className="cc-layout">
        {/* ===== HEADER ===== */}
        <header className="cc-header">
          <div className="cc-header-left">
            <div className="cc-logo-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2">
                <rect x="3" y="3" width="18" height="18" rx="3" />
                <path d="M9 12l2 2 4-4" />
              </svg>
            </div>
            <div>
              <div className="cc-header-title">AgentBox</div>
              <div className="cc-header-sub">Your AI agents are working</div>
            </div>
          </div>
          <div className="cc-header-right">
            <div className="cc-header-kpi">
              <div className="cc-header-kpi-icon" style={{ background: 'rgba(34,197,94,0.1)' }}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" /></svg>
              </div>
              <div>
                <div className="cc-header-kpi-label">Active Agents</div>
                <div className="cc-header-kpi-value"><AnimatedNumber value={headerKpis.activeAgents} duration={600} /></div>
              </div>
            </div>
            <div className="cc-header-kpi">
              <div className="cc-header-kpi-icon" style={{ background: 'rgba(59,130,246,0.1)' }}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12" /></svg>
              </div>
              <div>
                <div className="cc-header-kpi-label">Tasks Today</div>
                <div className="cc-header-kpi-value"><AnimatedNumber value={headerKpis.tasksToday} duration={600} /></div>
              </div>
            </div>
            <div className="cc-header-kpi">
              <div className="cc-header-kpi-icon" style={{ background: 'rgba(167,139,250,0.1)' }}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg>
              </div>
              <div>
                <div className="cc-header-kpi-label">Total Executions</div>
                <div className="cc-header-kpi-value"><AnimatedNumber value={headerKpis.totalExecutions} duration={600} /></div>
              </div>
            </div>
            <button className="cc-fullscreen-btn" onClick={toggleFullscreen} title="Toggle Fullscreen">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                {isFullscreen ? <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3" />
                  : <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3" />}
              </svg>
            </button>
          </div>
        </header>

        {/* ===== KPI CARDS ===== */}
        <div className="cc-kpi-grid">
          <div className="cc-kpi-card">
            <div className="cc-kpi-top">
              <div>
                <div className="cc-kpi-label">Live Tasks</div>
                <div className="cc-kpi-value" style={{ color: '#4ade80' }}><AnimatedNumber value={kpis.liveTasks} duration={800} /></div>
              </div>
              <Sparkline data={kpiHistory.tasks} color="#4ade80" width={72} height={28} />
            </div>
          </div>
          <div className="cc-kpi-card">
            <div className="cc-kpi-top">
              <div>
                <div className="cc-kpi-label">Tokens / min</div>
                <div className="cc-kpi-value" style={{ color: '#a78bfa' }}><AnimatedNumber value={Math.round(kpis.tokens / 1000)} />K</div>
              </div>
              <Sparkline data={kpiHistory.tokens} color="#a78bfa" width={72} height={28} />
            </div>
          </div>
          <div className="cc-kpi-card">
            <div className="cc-kpi-top">
              <div>
                <div className="cc-kpi-label">Sessions</div>
                <div className="cc-kpi-value" style={{ color: '#60a5fa' }}><AnimatedNumber value={kpis.sessions} duration={800} /></div>
              </div>
              <Sparkline data={kpiHistory.sessions} color="#60a5fa" width={72} height={28} />
            </div>
          </div>
          <div className="cc-kpi-card">
            <div className="cc-kpi-top">
              <div>
                <div className="cc-kpi-label">Success Rate</div>
                <div className="cc-kpi-value" style={{ color: '#facc15' }}>{kpis.successRate.toFixed(1)}%</div>
              </div>
              <Sparkline data={kpiHistory.success} color="#facc15" width={72} height={28} />
            </div>
          </div>
          <div className="cc-kpi-card">
            <div className="cc-kpi-top">
              <div>
                <div className="cc-kpi-label">Avg Duration</div>
                <div className="cc-kpi-value" style={{ color: '#f87171' }}>{kpis.avgDuration.toFixed(1)}s</div>
              </div>
              <Sparkline data={kpiHistory.duration} color="#f87171" width={72} height={28} />
            </div>
          </div>
        </div>

        {/* ===== ROW 1 ===== */}
        <div className="cc-row-1">
          {/* Performance Overview */}
          <div className="cc-card cc-card-perf">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18" /><polyline points="17 6 23 6 23 12" /></svg>
                <span>Performance Overview</span>
              </div>
              <span className="cc-card-subtitle">Last 6 months</span>
            </div>
            <div className="cc-perf-kpis">
              <div className="cc-perf-kpi">
                <div className="cc-perf-kpi-icon">⚡</div>
                <div className="cc-perf-kpi-info">
                  <div className="cc-perf-kpi-label">Tasks Completed</div>
                  <div className="cc-perf-kpi-row">
                    <span className="cc-perf-kpi-value"><AnimatedNumber value={perfKpis.tasksCompleted} /></span>
                    <span className="cc-change cc-change-up">+{perfKpis.tasksCompletedChange}%</span>
                  </div>
                </div>
              </div>
              <div className="cc-perf-kpi">
                <div className="cc-perf-kpi-icon">🔤</div>
                <div className="cc-perf-kpi-info">
                  <div className="cc-perf-kpi-label">Tokens Used</div>
                  <div className="cc-perf-kpi-row">
                    <span className="cc-perf-kpi-value">{(perfKpis.tokensUsed / 1000000).toFixed(1)}M</span>
                    <span className="cc-change cc-change-up">+{perfKpis.tokensChange}%</span>
                  </div>
                </div>
              </div>
              <div className="cc-perf-kpi">
                <div className="cc-perf-kpi-icon">⏱️</div>
                <div className="cc-perf-kpi-info">
                  <div className="cc-perf-kpi-label">Avg Latency</div>
                  <div className="cc-perf-kpi-row">
                    <span className="cc-perf-kpi-value">{perfKpis.avgLatency.toFixed(1)}s</span>
                    <span className={`cc-change ${perfKpis.latencyChange <= 0 ? 'cc-change-up' : 'cc-change-down'}`}>{perfKpis.latencyChange}%</span>
                  </div>
                </div>
              </div>
            </div>
            <PerformanceChart data={perfChartData} color="#22c55e" />
          </div>

          {/* Provider Status */}
          <div className="cc-card cc-card-providers">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><path d="M2 20h.01" /><path d="M7 20v-4" /><path d="M12 20v-8" /><path d="M17 20V8" /><path d="M22 4v16" /></svg>
                <span>Provider Status</span>
              </div>
              <span className="cc-card-badge cc-badge-online">{providers.filter(p => p.status === 'online').length} Online</span>
            </div>
            <div className="cc-provider-grid">
              {providers.map(p => (
                <div key={p.name} className="cc-provider-item">
                  <div className="cc-provider-top">
                    <span className="cc-provider-dot" style={{ background: statusColor(p.status) }} />
                    <span className="cc-provider-name">{p.name}</span>
                  </div>
                  <div className="cc-provider-model">{p.model}</div>
                  <div className="cc-provider-stats">
                    <span className="cc-provider-stat-val"><AnimatedNumber value={p.tasks} duration={800} /></span>
                    <span className="cc-provider-stat-label"> tasks</span>
                  </div>
                  <div className="cc-provider-daily">
                    <span style={{ color: '#4ade80' }}>+{p.perDay}/day</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Active Executions */}
          <div className="cc-card cc-card-executions">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" strokeWidth="2"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
                <span>Active Executions</span>
              </div>
              <span className="cc-card-badge cc-badge-active">{executions.length} Active</span>
            </div>
            <div className="cc-exec-list">
              {executions.map(exec => (
                <div key={exec.id} className="cc-exec-item">
                  <div className="cc-exec-top">
                    <span className="cc-exec-prompt">{exec.prompt}</span>
                    <span className="cc-exec-eta">{exec.eta}</span>
                  </div>
                  <div className="cc-exec-meta">
                    <span className="cc-exec-agent">{exec.agent}</span>
                    {adapterLabel(exec.adapter)}
                  </div>
                  <div className="cc-exec-progress-row">
                    <div className="cc-exec-bar">
                      <div className="cc-exec-bar-fill" style={{ width: `${exec.progress}%` }} />
                    </div>
                    <span className="cc-exec-pct">{exec.progress}%</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* ===== ROW 2 ===== */}
        <div className="cc-row-2">
          {/* Task Pipeline */}
          <div className="cc-card cc-card-pipeline">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#eab308" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /></svg>
                <span>Task Pipeline</span>
              </div>
              <span className="cc-card-subtitle">{pipeline.filter(t => t.status === 'running').length} running</span>
            </div>
            <div className="cc-pipeline-list">
              {pipeline.map(task => (
                <div key={task.id} className="cc-pipeline-item">
                  <div className="cc-pipeline-indicator" style={{ background: statusColor(task.status) }} />
                  <div className="cc-pipeline-content">
                    <div className="cc-pipeline-top">
                      <span className="cc-pipeline-prompt">{task.prompt}</span>
                      <span className="cc-pipeline-elapsed">{task.elapsed}</span>
                    </div>
                    <div className="cc-pipeline-meta">
                      <span className="cc-pipeline-agent">Agent: <b>{task.agent}</b></span>
                      {adapterLabel(task.adapter)}
                    </div>
                    {task.status === 'running' && (
                      <div className="cc-pipeline-bar-wrap">
                        <div className="cc-pipeline-bar">
                          <div className="cc-pipeline-bar-fill" style={{ width: `${task.progress}%`, background: statusColor(task.status) }} />
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Agent Teams */}
          <div className="cc-card cc-card-teams">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg>
                <span>Agent Teams</span>
              </div>
              <div className="cc-teams-badge">
                <span>{activeTeamAgents}/{totalAgents} active</span>
                <div className="cc-teams-indicator" />
              </div>
            </div>
            <div className="cc-teams-list">
              {teamStats.map(team => (
                <div key={team.name} className="cc-team-item">
                  <div className="cc-team-icon" style={{ background: `${team.color}15`, color: team.color }}>{team.icon}</div>
                  <div className="cc-team-info">
                    <div className="cc-team-name-row">
                      <span className="cc-team-name">{team.name}</span>
                      <span className="cc-team-agents">{team.agents}</span>
                    </div>
                    <div className="cc-team-desc">{team.desc}</div>
                  </div>
                  <div className="cc-team-tasks">
                    <span className="cc-team-tasks-val"><AnimatedNumber value={team.tasks} duration={800} /></span>
                    <span className="cc-team-tasks-label">tasks</span>
                  </div>
                  <div className="cc-team-arrow">›</div>
                </div>
              ))}
            </div>
          </div>

          {/* Live Agent Activity */}
          <div className="cc-card cc-card-activity">
            <div className="cc-card-header">
              <div className="cc-card-title">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#f87171" strokeWidth="2"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
                <span>Live Agent Activity</span>
              </div>
              <span className="cc-live-badge">LIVE</span>
            </div>
            <div className="cc-activity-list">
              {activities.map((item, idx) => (
                <div key={item.id} className={`cc-activity-item ${idx === 0 ? 'cc-activity-new' : ''}`}>
                  <div className="cc-activity-dot-wrap">
                    <span className="cc-activity-dot" style={{ background: statusColor(item.adapter === 'claude-code' ? 'running' : item.adapter === 'codex' ? 'completed' : 'queued') }} />
                    {idx < activities.length - 1 && <span className="cc-activity-line" />}
                  </div>
                  <div className="cc-activity-content">
                    <div className="cc-activity-top">
                      <span className="cc-activity-agent">{item.agent}</span>
                      {adapterLabel(item.adapter)}
                      <span className="cc-activity-time">{item.timestamp}</span>
                    </div>
                    <div className="cc-activity-action">{item.action}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
