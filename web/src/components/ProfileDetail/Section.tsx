import { useState, type ReactNode } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'

interface SectionProps {
  title: string
  icon: ReactNode
  children: ReactNode
  defaultOpen?: boolean
}

export function Section({ title, icon, children, defaultOpen = true }: SectionProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className="card">
      <button
        onClick={() => setOpen(!open)}
        className="w-full px-4 py-3 flex items-center gap-3 hover:bg-secondary/50 transition-colors"
      >
        <span className="text-emerald-400">{icon}</span>
        <span className="font-semibold text-primary flex-1 text-left">{title}</span>
        {open ? (
          <ChevronDown className="w-4 h-4 text-muted" />
        ) : (
          <ChevronRight className="w-4 h-4 text-muted" />
        )}
      </button>
      {open && <div className="px-4 pb-4 border-t border-primary">{children}</div>}
    </div>
  )
}
