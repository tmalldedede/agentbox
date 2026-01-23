import { useNavigate } from '@tanstack/react-router'
import { Plus } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { MCPServersProvider } from './components/mcp-servers-provider'
import { MCPServersDialogs } from './components/mcp-servers-dialogs'
import MCPServerList from './components/MCPServerList'

export function MCPServers() {
  const navigate = useNavigate()

  return (
    <MCPServersProvider>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>MCP Servers</h2>
            <p className='text-muted-foreground'>
              Configure MCP servers to extend agent capabilities with external tools and integrations.
            </p>
          </div>
          <Button size='sm' onClick={() => navigate({ to: '/mcp-servers/new' })}>
            <Plus className='mr-2 h-4 w-4' />
            New Server
          </Button>
        </div>

        <MCPServerList />
      </Main>

      <MCPServersDialogs />
    </MCPServersProvider>
  )
}
