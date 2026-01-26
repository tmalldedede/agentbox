import { Cpu, Terminal, ArrowRight } from 'lucide-react'

export interface ActiveExecution {
  id: string
  prompt: string
  agent: string
  adapter: string
  progress: number
  eta: string
}

function AdapterBadge({ adapter }: { adapter: string }) {
  const config: Record<string, { bg: string; color: string; label: string; border: string }> = {
    'claude-code': { bg: 'bg-purple-500/10', color: 'text-purple-400', border: 'border-purple-500/20', label: 'Claude' },
    'codex': { bg: 'bg-emerald-500/10', color: 'text-emerald-400', border: 'border-emerald-500/20', label: 'Codex' },
    'opencode': { bg: 'bg-blue-500/10', color: 'text-blue-400', border: 'border-blue-500/20', label: 'OpenCode' },
  }
  const c = config[adapter] || config['opencode']

  return (
    <span className={`inline-flex items-center rounded border px-1.5 py-0.5 text-[10px] font-medium ${c.bg} ${c.color} ${c.border}`}>
      {c.label}
    </span>
  )
}

export function ActiveExecutions({ executions }: { executions: ActiveExecution[] }) {
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card/50 backdrop-blur-sm">
      <div className="flex items-center justify-between border-b border-border px-5 py-4">
        <div className="flex items-center gap-2">
          <Cpu className="text-purple-500" size={16} />
          <h3 className="font-semibold text-foreground">Active Executions</h3>
        </div>
        <span className="flex items-center gap-1.5 rounded-full bg-purple-500/10 border border-purple-500/20 px-2.5 py-1 text-[10px] font-bold text-purple-500">
          <span className="relative flex h-1.5 w-1.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-purple-500"></span>
          </span>
          {executions.length} Active
        </span>
      </div>

      <div className="flex flex-1 flex-col gap-3 overflow-y-auto p-4 custom-scrollbar">
        {executions.map((exec) => (
          <div
            key={exec.id}
            className="group relative flex flex-col gap-3 rounded-xl border border-border bg-accent/40 p-4 transition-all hover:border-border/80 hover:bg-accent/60"
          >
            {/* Header */}
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-2.5 min-w-0">
                <div className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded bg-accent text-muted-foreground">
                  <Terminal size={10} />
                </div>
                <span className="text-xs font-medium text-foreground leading-relaxed line-clamp-2" title={exec.prompt}>
                  {exec.prompt}
                </span>
              </div>
              <span className="shrink-0 text-[10px] font-medium text-muted-foreground bg-accent/50 px-2 py-1 rounded-md border border-border font-mono">
                ETA {exec.eta}
              </span>
            </div>

            {/* Meta & Progress */}
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between text-[11px]">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <span>{exec.agent}</span>
                  <ArrowRight size={10} className="text-border" />
                  <AdapterBadge adapter={exec.adapter} />
                </div>
                <span className="font-mono font-bold text-purple-500">{exec.progress}%</span>
              </div>

              <div className="h-1.5 w-full overflow-hidden rounded-full bg-accent">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-purple-600 via-violet-500 to-indigo-500 relative"
                  style={{ width: `${exec.progress}%`, transition: 'width 0.5s ease-out' }}
                >
                  <div className="absolute inset-0 bg-white/25 animate-[shimmer_2s_infinite]" />
                </div>
              </div>
            </div>
          </div>
        ))}

        {executions.length === 0 && (
          <div className="flex flex-1 items-center justify-center text-xs text-muted-foreground font-medium">
            System Idle. Waiting for tasks...
          </div>
        )}
      </div>
    </div>
  )
}