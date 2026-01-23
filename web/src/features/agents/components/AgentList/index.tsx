import { useNavigate } from '@tanstack/react-router'
import { Loader2, Plus } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { useAgents } from '@/hooks'
import { AgentsProvider } from './agents-provider'
import { AgentsTable } from './agents-table'
import { AgentsDialogs } from './agents-dialogs'

export default function AgentList() {
  const navigate = useNavigate()
  const { data: agents = [], isLoading } = useAgents()

  return (
    <AgentsProvider>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Agents</h2>
            <p className='text-muted-foreground'>
              Agents combine a Provider, Runtime, Skills, and system prompts into a deployable AI assistant.
            </p>
          </div>
          <Button size='sm' onClick={() => navigate({ to: '/agents/new' })}>
            <Plus className='mr-2 h-4 w-4' />
            New Agent
          </Button>
        </div>

        {isLoading ? (
          <div className='flex flex-1 items-center justify-center'>
            <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
          </div>
        ) : (
          <AgentsTable data={agents} />
        )}
      </Main>

      <AgentsDialogs />
    </AgentsProvider>
  )
}
