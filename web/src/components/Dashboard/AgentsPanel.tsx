import { Cpu } from 'lucide-react'
import type { Agent } from '../../types'
import { useLanguage } from '../../contexts/LanguageContext'
import { AgentItem } from './AgentItem'

interface AgentsPanelProps {
  agents: Agent[]
  isLoading: boolean
}

export function AgentsPanel({ agents, isLoading }: AgentsPanelProps) {
  const { t } = useLanguage()

  return (
    <div className="panel flex flex-col min-h-0">
      <div className="panel-header">
        <div>
          <div className="panel-title">{t('supportedAgents')}</div>
          <div className="panel-subtitle">{agents.length} {t('available')}</div>
        </div>
      </div>
      <div className="panel-content">
        {isLoading && agents.length === 0 ? (
          <div className="p-5 space-y-4">
            {[1, 2].map(i => (
              <div key={i} className="h-16 skeleton rounded-lg" />
            ))}
          </div>
        ) : agents.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center p-8">
            <Cpu className="w-12 h-12 text-muted mb-4" />
            <p className="text-secondary">{t('noAgentsAvailable')}</p>
          </div>
        ) : (
          agents.map(agent => <AgentItem key={agent.name} agent={agent} />)
        )}
      </div>
    </div>
  )
}
