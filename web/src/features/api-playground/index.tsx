import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import ApiPlayground from './components/ApiPlayground'

export function APIPlayground({
  preselectedAgentId,
  initialPrompt
}: {
  preselectedAgentId?: string,
  initialPrompt?: string
}) {
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
        <ApiPlayground preselectedAgentId={preselectedAgentId} initialPrompt={initialPrompt} />
      </Main>
    </>
  )
}
