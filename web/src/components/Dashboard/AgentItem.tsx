import { CheckCircle } from 'lucide-react'
import type { Agent } from '../../types'

interface AgentItemProps {
  agent: Agent
}

const agentColors: Record<string, string> = {
  'claude-code': 'bg-purple-500/20 text-purple-400',
  codex: 'bg-emerald-500/20 text-emerald-400',
}

export function AgentItem({ agent }: AgentItemProps) {
  const initials = agent.name.slice(0, 2).toUpperCase()

  return (
    <div className="list-item">
      <div className={`agent-avatar ${agentColors[agent.name] || 'bg-blue-500/20 text-blue-400'}`}>
        {initials}
      </div>
      <div className="flex-1 min-w-0">
        <div className="font-medium text-primary">{agent.display_name}</div>
        <div className="text-xs text-muted mt-0.5">{agent.description}</div>
      </div>
      <div className="badge badge-scaling">
        <CheckCircle className="w-3 h-3" />
        Ready
      </div>
    </div>
  )
}
