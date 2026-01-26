import { BarChart2, TrendingUp, TrendingDown, ChevronDown } from 'lucide-react'
import { DashboardChart } from './Charts'

interface PerformanceProps {
  kpis: {
    tasksCompleted: number
    tokensUsed: number
    avgLatency: number
    tasksCompletedChange: number
    tokensChange: number
    latencyChange: number
  }
  chartData: number[]
}

export function PerformanceOverview({ kpis, chartData }: PerformanceProps) {
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card/50 backdrop-blur-sm">
      <div className="flex items-center justify-between border-b border-border px-5 py-4">
        <div className="flex items-center gap-2">
          <BarChart2 className="text-emerald-500" size={16} />
          <h3 className="font-semibold text-foreground">Performance Overview</h3>
        </div>
        <div className="relative">
          <select className="appearance-none bg-accent border border-border text-xs font-medium text-foreground rounded-lg pl-3 pr-8 py-1.5 outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/20 cursor-pointer transition-all hover:bg-accent/80">
            <option>Last 6 Months</option>
            <option>Last 30 Days</option>
            <option>Last 7 Days</option>
          </select>
          <ChevronDown size={12} className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
        </div>
      </div>

      <div className="flex flex-col flex-1 p-5 gap-6">
        {/* KPI Row */}
        <div className="grid grid-cols-3 gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Tasks Completed</span>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-foreground tracking-tight">{kpis.tasksCompleted.toLocaleString()}</span>
              <span className="inline-flex items-center rounded-full bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-bold text-emerald-500 border border-emerald-500/20">
                <TrendingUp size={10} className="mr-1" />
                {kpis.tasksCompletedChange}%
              </span>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Tokens Used</span>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-foreground tracking-tight">{(kpis.tokensUsed / 1000000).toFixed(1)}M</span>
              <span className="inline-flex items-center rounded-full bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-bold text-emerald-500 border border-emerald-500/20">
                <TrendingUp size={10} className="mr-1" />
                {kpis.tokensChange}%
              </span>
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Avg Latency</span>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-foreground tracking-tight">{kpis.avgLatency.toFixed(1)}s</span>
              <span className={`inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-bold border ${kpis.latencyChange <= 0 ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' : 'bg-red-500/10 text-red-500 border-red-500/20'}`}>
                {kpis.latencyChange <= 0 ? <TrendingDown size={10} className="mr-1" /> : <TrendingUp size={10} className="mr-1" />}
                {Math.abs(kpis.latencyChange)}%
              </span>
            </div>
          </div>
        </div>

        {/* Chart Area */}
        <div className="flex-1 w-full min-h-0">
          <DashboardChart data={chartData} color="#10b981" />
        </div>
      </div>
    </div>
  )
}