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
    <div className="flex flex-col rounded-xl border border-zinc-800 bg-zinc-900/50 backdrop-blur-sm h-full">
      <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
        <div className="flex items-center gap-2">
          <Server className="text-blue-400" size={16} />
          <h3 className="font-semibold text-zinc-100">Model Providers</h3>
        </div>
        <div className="flex items-center gap-2 rounded-full bg-zinc-800 px-2.5 py-1 text-[10px] font-medium text-zinc-400">
          <span className="h-1.5 w-1.5 rounded-full bg-green-500" />
          {onlineCount} / {providers.length} Online
        </div>
      </div>

      <div className="flex flex-col gap-3 p-4 overflow-y-auto custom-scrollbar">
        {providers.map((p) => (
          <div 
            key={p.name} 
            className="group flex items-center justify-between gap-3 rounded-lg border border-zinc-800 bg-zinc-900/40 p-3 transition-all hover:border-zinc-700 hover:bg-zinc-800/60"
          >
            {/* Left: Info */}
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <StatusDot status={p.status} />
                <span className="font-semibold text-sm text-zinc-100">{p.name}</span>
              </div>
              <div className="text-[11px] text-zinc-500 font-mono pl-4">{p.model}</div>
            </div>
            
            {/* Right: Stats */}
            <div className="flex items-center gap-4">
               <div className="flex flex-col items-end">
                 <span className="text-sm font-bold text-white tabular-nums">{p.tasks}</span>
                 <span className="text-[9px] text-zinc-600 uppercase font-medium">Tasks</span>
               </div>
               <div className="h-8 w-px bg-zinc-800" />
               <div className="flex flex-col items-end min-w-[40px]">
                 <span className="text-xs font-bold text-green-400 tabular-nums">+{p.perDay}</span>
                 <span className="text-[9px] text-zinc-600 uppercase font-medium">Daily</span>
               </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}