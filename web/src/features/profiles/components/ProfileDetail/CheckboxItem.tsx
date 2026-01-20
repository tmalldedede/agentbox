interface CheckboxItemProps {
  checked: boolean
  onChange: () => void
  label: string
  sublabel?: string
  badge?: string
  activeColor?: string
}

export function CheckboxItem({
  checked,
  onChange,
  label,
  sublabel,
  badge,
  activeColor = 'emerald',
}: CheckboxItemProps) {
  const activeClasses: Record<string, string> = {
    emerald: 'bg-emerald-500/20 border border-emerald-500/30',
    amber: 'bg-amber-500/20 border border-amber-500/30',
    purple: 'bg-purple-500/20 border border-purple-500/30',
  }

  const checkboxClasses: Record<string, string> = {
    emerald: 'bg-emerald-500 border-emerald-500',
    amber: 'bg-amber-500 border-amber-500',
    purple: 'bg-purple-500 border-purple-500',
  }

  return (
    <label
      className={`flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors ${
        checked ? activeClasses[activeColor] : 'bg-secondary hover:bg-secondary/80'
      }`}
    >
      <input type="checkbox" checked={checked} onChange={onChange} className="sr-only" />
      <div
        className={`w-4 h-4 rounded border flex items-center justify-center ${
          checked ? checkboxClasses[activeColor] : 'border-muted'
        }`}
      >
        {checked && (
          <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path
              fillRule="evenodd"
              d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
              clipRule="evenodd"
            />
          </svg>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="font-medium text-foreground">{label}</div>
        {sublabel && <div className="text-xs text-muted-foreground truncate">{sublabel}</div>}
      </div>
      {badge && (
        <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">{badge}</span>
      )}
    </label>
  )
}
