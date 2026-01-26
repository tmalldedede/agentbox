import { Server } from 'lucide-react'

interface Provider {
  name: string
  model: string
  tasks: number
  perDay: number
  status: 'online' | 'degraded' | 'offline'
}

function StatusDot({ status }: { status: Provider['status'] }) {
  const color = status === 'online'
    ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] animate-pulse'
    : status === 'degraded'
      ? 'bg-yellow-500 animate-pulse'
      : 'bg-red-500'
  return <div className={`h-2 w-2 rounded-full ${color}`} />
}

export function ProviderStatus({ providers }: { providers: Provider[] }) {
  const onlineCount = providers.filter(p => p.status === 'online').length

  return (
    <div className="flex flex-col rounded-xl border border-border bg-card/50 backdrop-blur-sm h-full">
      <div className="flex items-center justify-between border-b border-border px-5 py-4">
        <div className="flex items-center gap-2">
          <Server className="text-blue-500" size={16} />
          <h3 className="font-semibold text-foreground">Model Providers</h3>
        </div>
        <div className="flex items-center gap-2 rounded-full bg-accent px-2.5 py-1 text-[10px] font-medium text-muted-foreground">
          <span className="h-1.5 w-1.5 rounded-full bg-green-500" />
          {onlineCount} / {providers.length} Online
        </div>
      </div>

      <div className="flex flex-col gap-3 p-4 overflow-y-auto custom-scrollbar">
        {providers.map((p) => (
          <div
            key={p.name}
            className="group flex items-center justify-between gap-3 rounded-lg border border-border bg-accent/40 p-3 transition-all hover:border-border/80 hover:bg-accent/60"
          >
            {/* Left: Info */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <StatusDot status={p.status} />
                <span className="font-semibold text-sm text-foreground">{p.name}</span>
              </div>
              <div className="text-[11px] text-muted-foreground font-mono pl-4">{p.model}</div>
            </div>

            {/* Right: Stats */}
            <div className="flex items-center gap-4">
              <div className="flex flex-col items-end">
                <span className="text-sm font-bold text-foreground tabular-nums">{p.tasks}</span>
                <span className="text-[9px] text-muted-foreground/60 uppercase font-medium">Tasks</span>
              </div>
              <div className="h-8 w-px bg-border" />
              <div className="flex flex-col items-end min-w-[40px]">
                <span className="text-xs font-bold text-green-500 tabular-nums">+{p.perDay}</span>
                <span className="text-[9px] text-muted-foreground/60 uppercase font-medium">Daily</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}