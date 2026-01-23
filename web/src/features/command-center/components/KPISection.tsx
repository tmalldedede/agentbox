import { Activity, Zap, Layers, Clock, CheckCircle } from 'lucide-react'
import { SparklineChart } from './Charts'
import { useEffect, useState, useRef } from 'react'

function AnimatedNumber({ value }: { value: number }) {
  const [display, setDisplay] = useState(value)
  const prevRef = useRef(value)

  useEffect(() => {
    const from = prevRef.current
    const to = value
    if (from === to) return
    const duration = 600
    const start = performance.now()
    let raf: number
    const step = (now: number) => {
      const t = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - t, 3) // easeOutCubic
      setDisplay(Math.round(from + (to - from) * eased))
      if (t < 1) raf = requestAnimationFrame(step)
    }
    raf = requestAnimationFrame(step)
    prevRef.current = to
    return () => cancelAnimationFrame(raf)
  }, [value])

  return <span className="tabular-nums">{display.toLocaleString()}</span>
}

interface KPICardProps {
  label: string
  value: number | string
  subValue?: string
  color: string
  icon: React.ReactNode
  history: number[]
}

function KPICard({ label, value, subValue, color, icon, history }: KPICardProps) {
  return (
    <div className="relative overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 backdrop-blur-sm transition-all hover:border-zinc-700 hover:bg-zinc-800/80 group">
      {/* Background Glow */}
      <div 
        className="absolute -right-6 -top-6 h-24 w-24 rounded-full opacity-0 blur-3xl transition-opacity duration-500 group-hover:opacity-10" 
        style={{ backgroundColor: color }} 
      />
      
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-1.5 z-10">
          <div className="flex items-center gap-2 text-[11px] font-medium text-zinc-400 uppercase tracking-wide">
             <div className="flex h-5 w-5 items-center justify-center rounded-md bg-zinc-800 text-zinc-300 shadow-sm border border-zinc-700/50">
               {icon}
             </div>
             {label}
          </div>
          <div className="mt-1 text-2xl font-bold text-white tracking-tight flex items-baseline">
             {typeof value === 'number' ? <AnimatedNumber value={value} /> : value}
             {subValue && <span className="text-sm font-medium text-zinc-500 ml-1 translate-y-[-1px]">{subValue}</span>}
          </div>
        </div>
        <div className="opacity-80 group-hover:opacity-100 transition-opacity">
           <SparklineChart data={history} color={color} />
        </div>
      </div>
    </div>
  )
}

interface KPISectionProps {
  kpis: {
    liveTasks: number
    tokens: number
    sessions: number
    successRate: number
    avgDuration: number
  }
  history: {
    tasks: number[]
    tokens: number[]
    sessions: number[]
    success: number[]
    duration: number[]
  }
}

export function KPISection({ kpis, history }: KPISectionProps) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
      <KPICard label="Live Tasks" value={kpis.liveTasks} color="#10b981" icon={<Activity size={12} />} history={history.tasks} />
      <KPICard label="Tokens / min" value={Math.round(kpis.tokens / 1000)} subValue="k" color="#8b5cf6" icon={<Zap size={12} />} history={history.tokens} />
      <KPICard label="Active Sessions" value={kpis.sessions} color="#3b82f6" icon={<Layers size={12} />} history={history.sessions} />
      <KPICard label="Success Rate" value={kpis.successRate.toFixed(1)} subValue="%" color="#f59e0b" icon={<CheckCircle size={12} />} history={history.success} />
      <KPICard label="Avg Duration" value={kpis.avgDuration.toFixed(1)} subValue="s" color="#ef4444" icon={<Clock size={12} />} history={history.duration} />
    </div>
  )
}