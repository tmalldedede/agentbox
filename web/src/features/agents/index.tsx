import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import AgentList from './components/AgentList'
import AgentDetailComponent from './components/AgentDetail'

export function Agents() {
  return (
    <>
      <Header>
        <Search />
        <div className='ml-auto flex items-center gap-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <AgentList />
      </Main>
    </>
  )
}

export function AgentDetail({ agentId }: { agentId: string }) {
  return (
    <>
      <Header>
        <Search />
        <div className='ml-auto flex items-center gap-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <AgentDetailComponent agentId={agentId} />
      </Main>
    </>
  )
}
