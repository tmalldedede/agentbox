import { useState } from 'react'
import { Loader2, Plus } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import type { Agent } from '@/types'
import { useSessions } from '@/hooks'
import { api } from '@/services/api'
import { SessionsProvider } from './sessions-provider'
import { SessionsTable } from './sessions-table'
import { SessionsDialogs } from './sessions-dialogs'
import CreateSessionModal from './CreateSessionModal'

export default function SessionList() {
  const { data: sessions = [], isLoading } = useSessions()

  const [showCreateModal, setShowCreateModal] = useState(false)
  const [agents, setAgents] = useState<Agent[]>([])
  const [agentsLoading, setAgentsLoading] = useState(false)

  const handleOpenCreate = async () => {
    setAgentsLoading(true)
    try {
      const fetchedAgents = await api.listAgents()
      setAgents(fetchedAgents)
      setShowCreateModal(true)
    } catch (err) {
      console.error('Failed to fetch agents:', err)
    } finally {
      setAgentsLoading(false)
    }
  }

  return (
    <SessionsProvider>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Sessions</h2>
            <p className='text-muted-foreground'>
              Manage active and stopped agent sessions. Click on a row to view details.
            </p>
          </div>
          <Button size='sm' onClick={handleOpenCreate} disabled={agentsLoading}>
            {agentsLoading ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Plus className='mr-2 h-4 w-4' />
            )}
            New Session
          </Button>
        </div>

        {isLoading ? (
          <div className='flex flex-1 items-center justify-center'>
            <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
          </div>
        ) : (
          <SessionsTable data={sessions} />
        )}
      </Main>

      <SessionsDialogs />

      {showCreateModal && (
        <CreateSessionModal
          agents={agents}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => setShowCreateModal(false)}
        />
      )}
    </SessionsProvider>
  )
}
