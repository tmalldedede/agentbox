import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import MCPServerDetail from './components/MCPServerDetail'

type MCPServerDetailPageProps = {
  serverId: string
}

export function MCPServerDetailPage({ serverId }: MCPServerDetailPageProps) {
  return (
    <>
      <Header fixed className='md:hidden' />
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <MCPServerDetail serverId={serverId} />
      </Main>
    </>
  )
}
