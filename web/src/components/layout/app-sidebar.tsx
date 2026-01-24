import { useMemo } from 'react'
import { useLayout } from '@/context/layout-provider'
import { useAuthStore } from '@/stores/auth-store'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { ThemeSwitch } from '@/components/theme-switch'
import { getSidebarData } from './data/sidebar-data'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { TeamSwitcher } from './team-switcher'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const { auth } = useAuthStore()

  const data = useMemo(() => {
    const role = auth.user?.role || 'user'
    const sidebar = getSidebarData(role)
    // Override user info from auth state
    if (auth.user) {
      sidebar.user = {
        name: auth.user.username,
        email: `${auth.user.username}@agentbox`,
        avatar: '/avatars/shadcn.jpg',
      }
    }
    return sidebar
  }, [auth.user])

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <TeamSwitcher teams={data.teams} />
      </SidebarHeader>
      <SidebarContent>
        {data.navGroups.map((props, index) => (
          <NavGroup key={props.title || `nav-group-${index}`} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <div className='flex items-center justify-end px-2'>
          <ThemeSwitch />
        </div>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
