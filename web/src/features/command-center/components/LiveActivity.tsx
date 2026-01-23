import { Zap } from 'lucide-react'

interface ActivityItem {
  id: string
  agent: string
  action: string
  adapter: string
  timestamp: string
}

export function LiveActivity({ activities }: { activities: ActivityItem[] }) {
  return (
    <div className="flex h-full flex-col rounded-xl border border-zinc-800 bg-zinc-900/50 backdrop-blur-sm">
      <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
        <div className="flex items-center gap-2">
          <Zap className="text-red-400" size={16} />
          <h3 className="font-semibold text-zinc-100">Live Feed</h3>
        </div>
        <div className="flex items-center gap-1.5 rounded-full bg-red-500/10 px-2.5 py-1 text-[10px] font-bold text-red-400 border border-red-500/20 shadow-[0_0_10px_rgba(239,68,68,0.1)]">
          <span className="relative flex h-1.5 w-1.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-red-500"></span>
          </span>
          LIVE
        </div>
      </div>

      <div className="flex flex-1 flex-col overflow-y-auto custom-scrollbar p-0">
        {activities.map((item, idx) => (
          <div 
            key={item.id} 
            className={`group relative flex gap-3 px-5 py-3.5 transition-colors hover:bg-zinc-800/30 ${idx === 0 ? 'cc-fade-in bg-zinc-800/20' : ''}`}
          >
            {/* Timeline Line */}
            {idx !== activities.length - 1 && (
               <div className="absolute left-[23px] top-[30px] bottom-0 w-px bg-zinc-800 group-hover:bg-zinc-700 transition-colors" />
            )}

            {/* Dot */}
            <div className="relative flex flex-col items-center pt-1.5">
              <div className={`h-2 w-2 rounded-full border border-zinc-900 ${idx === 0 ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.5)]' : 'bg-zinc-700'}`} />
            </div>
            
            <div className="flex flex-1 flex-col min-w-0 gap-0.5">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs font-semibold text-zinc-200 truncate group-hover:text-white transition-colors">
                  {item.agent}
                </span>
                <span className="text-[10px] text-zinc-600 whitespace-nowrap font-mono">
                  {item.timestamp}
                </span>
              </div>
              <p className="text-xs text-zinc-500 truncate group-hover:text-zinc-400 transition-colors">
                {item.action}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}