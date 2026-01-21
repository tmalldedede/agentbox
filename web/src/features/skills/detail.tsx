import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import SkillDetail from './components/SkillDetail'

type SkillDetailPageProps = {
  skillId: string
}

export function SkillDetailPage({ skillId }: SkillDetailPageProps) {
  return (
    <>
      <Header>
        <Search />
        <div className="ml-auto flex items-center gap-4">
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <SkillDetail skillId={skillId} />
      </Main>
    </>
  )
}
