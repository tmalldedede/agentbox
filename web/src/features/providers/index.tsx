import { Loader2, Plus, ShieldCheck } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { useConfiguredProviders, useProviderStats, useVerifyAllProviders } from '@/hooks/useProviders'
import { ProvidersDialogs } from './components/providers-dialogs'
import { ProvidersProvider, useProvidersContext } from './components/providers-provider'
import { ProvidersTable } from './components/providers-table'

function ProvidersContent() {
  const { setOpen } = useProvidersContext()
  const { data: providers = [], isLoading } = useConfiguredProviders()
  const { data: stats } = useProviderStats()
  const verifyAll = useVerifyAllProviders()

  return (
    <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
      <div className='flex flex-wrap items-end justify-between gap-2'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>Providers</h2>
          <p className='text-muted-foreground'>
            Manage AI service providers and their API keys.
          </p>
        </div>
        <div className='flex items-center gap-2'>
          {stats && stats.configured > 0 && (
            <Button
              variant='outline'
              size='sm'
              onClick={() => verifyAll.mutate()}
              disabled={verifyAll.isPending}
            >
              {verifyAll.isPending ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <ShieldCheck className='mr-2 h-4 w-4' />
              )}
              Test All
            </Button>
          )}
          <Button size='sm' onClick={() => setOpen('add')}>
            <Plus className='mr-2 h-4 w-4' />
            Add Provider
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className='grid grid-cols-2 md:grid-cols-4 gap-4'>
          <Card>
            <CardContent className='pt-4 pb-3'>
              <div className='text-2xl font-bold'>{stats.total}</div>
              <div className='text-sm text-muted-foreground'>Total Providers</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className='pt-4 pb-3'>
              <div className='text-2xl font-bold text-blue-500'>{stats.configured}</div>
              <div className='text-sm text-muted-foreground'>Configured</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className='pt-4 pb-3'>
              <div className='text-2xl font-bold text-emerald-500'>{stats.valid}</div>
              <div className='text-sm text-muted-foreground'>Valid</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className='pt-4 pb-3'>
              <div className='text-2xl font-bold text-red-500'>{stats.failed}</div>
              <div className='text-sm text-muted-foreground'>Failed</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Provider Table */}
      {isLoading ? (
        <div className='flex flex-1 items-center justify-center'>
          <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
        </div>
      ) : providers.length === 0 ? (
        <Card className='flex flex-col items-center justify-center py-16'>
          <CardContent className='text-center space-y-4'>
            <div className='text-4xl'>☁️</div>
            <div>
              <h3 className='text-lg font-medium'>No providers configured</h3>
              <p className='text-sm text-muted-foreground mt-1'>
                Add a provider to get started with AI agent sessions.
              </p>
            </div>
            <Button onClick={() => setOpen('add')}>
              <Plus className='mr-2 h-4 w-4' />
              Add Your First Provider
            </Button>
          </CardContent>
        </Card>
      ) : (
        <ProvidersTable data={providers} />
      )}
    </Main>
  )
}

export default function Providers() {
  return (
    <ProvidersProvider>
      <Header fixed className='md:hidden' />
      <ProvidersContent />
      <ProvidersDialogs />
    </ProvidersProvider>
  )
}
