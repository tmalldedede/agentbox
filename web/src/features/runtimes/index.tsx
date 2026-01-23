import { Loader2, Plus } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { useRuntimes } from '@/hooks/useRuntimes'
import { RuntimesDialogs } from './components/runtimes-dialogs'
import { RuntimesProvider, useRuntimesContext } from './components/runtimes-provider'
import { RuntimesTable } from './components/runtimes-table'

function RuntimesPrimaryButtons() {
  const { setOpen } = useRuntimesContext()
  return (
    <Button size='sm' onClick={() => setOpen('create')}>
      <Plus className='mr-2 h-4 w-4' />
      New Runtime
    </Button>
  )
}

export default function Runtimes() {
  const { data: runtimes = [], isLoading } = useRuntimes()

  return (
    <RuntimesProvider>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Runtimes</h2>
            <p className='text-muted-foreground'>
              Configure runtime environments for agents. Each runtime defines a Docker image and resource limits.
            </p>
          </div>
          <RuntimesPrimaryButtons />
        </div>

        {isLoading ? (
          <div className='flex flex-1 items-center justify-center'>
            <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
          </div>
        ) : (
          <RuntimesTable data={runtimes} />
        )}
      </Main>

      <RuntimesDialogs />
    </RuntimesProvider>
  )
}
