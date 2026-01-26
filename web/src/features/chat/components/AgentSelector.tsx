import { Bot, ChevronDown, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Agent } from '@/types'

interface AgentSelectorProps {
  agents: Agent[]
  selectedId: string
  onSelect: (agentId: string) => void
  disabled?: boolean
}

export function AgentSelector({
  agents,
  selectedId,
  onSelect,
  disabled,
}: AgentSelectorProps) {
  const selectedAgent = agents.find((a) => a.id === selectedId)
  const activeAgents = agents.filter((a) => a.status === 'active')

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='outline'
          className='w-full justify-between gap-2 sm:w-auto sm:min-w-[200px]'
          disabled={disabled || activeAgents.length === 0}
        >
          <div className='flex items-center gap-2'>
            <Bot className='h-4 w-4' />
            <span className='truncate'>
              {selectedAgent?.name || 'Select an Agent'}
            </span>
          </div>
          <ChevronDown className='h-4 w-4 shrink-0 opacity-50' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='start' className='w-[300px]'>
        {activeAgents.length === 0 ? (
          <div className='px-2 py-4 text-center text-sm text-muted-foreground'>
            No active agents available
          </div>
        ) : (
          activeAgents.map((agent) => (
            <DropdownMenuItem
              key={agent.id}
              onClick={() => onSelect(agent.id)}
              className='flex items-start gap-3 py-3'
            >
              <div className='flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10'>
                {agent.icon ? (
                  <span className='text-lg'>{agent.icon}</span>
                ) : (
                  <Bot className='h-4 w-4 text-primary' />
                )}
              </div>
              <div className='flex-1 space-y-1'>
                <div className='flex items-center gap-2'>
                  <span className='font-medium'>{agent.name}</span>
                  {agent.id === selectedId && (
                    <Check className='h-4 w-4 text-primary' />
                  )}
                </div>
                {agent.description && (
                  <p className='line-clamp-2 text-xs text-muted-foreground'>
                    {agent.description}
                  </p>
                )}
                <div className='flex items-center gap-2 text-xs text-muted-foreground'>
                  <span className='rounded bg-muted px-1.5 py-0.5'>
                    {agent.adapter}
                  </span>
                  <span>{agent.model || 'default model'}</span>
                </div>
              </div>
            </DropdownMenuItem>
          ))
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
