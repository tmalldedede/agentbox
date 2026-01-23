import { GitPullRequest, Circle, CheckCircle2, XCircle, Clock, MoreHorizontal } from 'lucide-react'

type TaskStatus = 'running' | 'queued' | 'completed' | 'failed'

export interface PipelineTask {
  id: string
  prompt: string
  status: TaskStatus
  adapter: string
  agent: string
  progress: number
  elapsed: string
}

const STATUS_CONFIG: Record<TaskStatus, { icon: any; color: string; bg: string; border: string }> = {
  running: { icon: Circle, color: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/20' },
  queued: { icon: Clock, color: 'text-yellow-400', bg: 'bg-yellow-500/10', border: 'border-yellow-500/20' },
  completed: { icon: CheckCircle2, color: 'text-green-400', bg: 'bg-green-500/10', border: 'border-green-500/20' },
  failed: { icon: XCircle, color: 'text-red-400', bg: 'bg-red-500/10', border: 'border-red-500/20' },
}

export function TaskPipeline({ tasks }: { tasks: PipelineTask[] }) {
  const runningCount = tasks.filter(t => t.status === 'running').length

  return (
    <div className="flex h-full flex-col rounded-xl border border-zinc-800 bg-zinc-900/50 backdrop-blur-sm">
      <div className="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
        <div className="flex items-center gap-2">
          <GitPullRequest className="text-yellow-400" size={16} />
          <h3 className="font-semibold text-zinc-100">Task Pipeline</h3>
        </div>
        <div className="flex items-center gap-2 text-xs font-medium text-zinc-500">
          <span className="relative flex h-2 w-2">
             <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
             <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
          </span>
          {runningCount} Running
        </div>
      </div>

      <div className="flex flex-1 flex-col overflow-y-auto custom-scrollbar p-0">
        {tasks.map((task, idx) => {
          const config = STATUS_CONFIG[task.status]
          const StatusIcon = config.icon

          return (
            <div key={task.id} className="relative pl-6 pr-4 py-3 group hover:bg-zinc-800/30 transition-colors">
              {/* Vertical Timeline Line */}
              {idx !== tasks.length - 1 && (
                <div className="absolute left-[29px] top-8 bottom-[-12px] w-px bg-zinc-800 group-hover:bg-zinc-700 transition-colors" />
              )}
              
              <div className="flex items-start gap-3">
                {/* Icon Wrapper */}
                <div className={`relative z-10 flex h-7 w-7 items-center justify-center rounded-lg border ${config.bg} ${config.border} ${config.color} shrink-0`}>
                  <StatusIcon size={14} />
                </div>

                {/* Content */}
                <div className="flex min-w-0 flex-1 flex-col gap-1.5 pt-0.5">
                  <div className="flex items-start justify-between gap-4">
                    <span className="text-sm font-medium text-zinc-200 truncate leading-tight group-hover:text-white transition-colors">
                      {task.prompt}
                    </span>
                    <span className="shrink-0 text-[10px] font-mono text-zinc-500 bg-zinc-800/50 px-1.5 py-0.5 rounded">
                      {task.elapsed}
                    </span>
                  </div>

                  <div className="flex items-center gap-3 text-xs text-zinc-500">
                     <div className="flex items-center gap-1.5">
                       <div className={`h-1.5 w-1.5 rounded-full ${task.status === 'running' ? 'bg-blue-500 animate-pulse' : 'bg-zinc-600'}`} />
                       <span className={`capitalize ${config.color} text-[11px] font-medium`}>{task.status}</span>
                     </div>
                     <span className="text-zinc-700">|</span>
                     <span className="text-[11px] truncate max-w-[120px]">
                       Agent: <span className="text-zinc-400">{task.agent}</span>
                     </span>
                  </div>

                  {/* Progress Bar (Only for Running) */}
                  {task.status === 'running' && (
                    <div className="mt-1 h-1 w-full overflow-hidden rounded-full bg-zinc-800">
                      <div
                        className="h-full rounded-full bg-gradient-to-r from-blue-600 to-indigo-500 relative overflow-hidden"
                        style={{ width: `${task.progress}%`, transition: 'width 0.8s ease-out' }}
                      >
                         <div className="absolute inset-0 bg-white/20 animate-[shimmer_1.5s_infinite]" />
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )
        })}
        {tasks.length === 0 && (
          <div className="flex flex-1 flex-col items-center justify-center gap-2 text-zinc-600">
            <MoreHorizontal size={24} />
            <span className="text-xs">No tasks in pipeline</span>
          </div>
        )}
      </div>
    </div>
  )
}