import type { MCPServer } from '@/types'
import { CheckboxItem } from './CheckboxItem'

interface MCPServerSelectorProps {
  selectedIds: string[]
  onChange: (ids: string[]) => void
  availableServers: MCPServer[]
  disabled?: boolean
}

export function MCPServerSelector({
  selectedIds,
  onChange,
  availableServers,
  disabled,
}: MCPServerSelectorProps) {
  const toggleServer = (id: string) => {
    if (disabled) return
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter(s => s !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  const enabledServers = availableServers.filter(s => s.is_enabled)

  return (
    <div className="space-y-2 mt-3">
      {enabledServers.length === 0 ? (
        <p className="text-muted-foreground text-sm">No MCP servers available</p>
      ) : (
        enabledServers.map(server => (
          <CheckboxItem
            key={server.id}
            checked={selectedIds.includes(server.id)}
            onChange={() => toggleServer(server.id)}
            label={server.name}
            sublabel={server.command}
            badge={server.category}
          />
        ))
      )}
    </div>
  )
}
