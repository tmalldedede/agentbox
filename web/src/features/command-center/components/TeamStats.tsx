import { Users, ChevronRight } from 'lucide-react'
import { useEffect, useState, useRef } from 'react'

interface TeamStat {
  name: string
  desc: string
  icon: string
  color: string
  agents: number
  tasks: number
}

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
      const eased = 1 - Math.pow(1 - t, 3)
      setDisplay(Math.round(from + (to - from) * eased))
      if (t < 1) raf = requestAnimationFrame(step)
    }
    raf = requestAnimationFrame(step)
    prevRef.current = to
    return () => cancelAnimationFrame(raf)
  }, [value])

  return <span className="tabular-nums">{display.toLocaleString()}</span>
}

export function TeamStats({ teams }: { teams: TeamStat[] }) {
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card/50 backdrop-blur-sm">
      <div className="flex items-center justify-between border-b border-border px-5 py-4">
        <div className="flex items-center gap-2">
          <Users className="text-blue-500" size={16} />
          <h3 className="font-semibold text-foreground">Agent Teams</h3>
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-2 p-3 overflow-y-auto custom-scrollbar">
        {teams.map((team) => (
          <div
            key={team.name}
            className="group flex cursor-pointer items-center gap-3 rounded-lg border border-transparent p-2.5 hover:border-border hover:bg-accent/50 transition-all"
          >
            {/* Icon Box */}
            <div
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-lg shadow-sm border border-transparent group-hover:border-border/5 transition-all"
              style={{ backgroundColor: `${team.color}15`, color: team.color }}
            >
              {team.icon}
            </div>

            {/* Main Info */}
            <div className="flex flex-1 min-w-0 flex-col gap-0.5">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-sm text-foreground group-hover:text-primary transition-colors">{team.name}</span>
                <span className="flex items-center justify-center rounded bg-accent border border-border/50 min-w-[50px] px-1.5 py-0.5 text-[9px] font-medium text-muted-foreground group-hover:bg-accent transition-colors">
                  {team.agents} Agents
                </span>
              </div>
              <span className="truncate text-xs text-muted-foreground">{team.desc}</span>
            </div>

            {/* Right Stats */}
            <div className="flex items-center gap-3">
              <div className="flex flex-col items-end">
                <span className="text-sm font-bold text-foreground group-hover:text-primary transition-colors tabular-nums">
                  <AnimatedNumber value={team.tasks} />
                </span>
                <span className="text-[9px] text-muted-foreground/60 uppercase font-medium tracking-wide">Tasks</span>
              </div>
              <ChevronRight size={14} className="text-border group-hover:text-muted-foreground transition-colors" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}