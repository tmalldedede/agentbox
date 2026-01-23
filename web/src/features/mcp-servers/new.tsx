import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import MCPServerForm from './components/MCPServerForm'

export function MCPServerNewPage() {
  return (
    <>
      <Header fixed className='md:hidden' />
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <MCPServerForm />
      </Main>
    </>
  )
}
