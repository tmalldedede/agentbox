import { Outlet } from '@tanstack/react-router'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { ApiDocsLayout } from './components/ApiDocsLayout'

export function ApiDocsPage() {
  return (
    <>
      <Header>
        <h1 className="text-lg font-semibold">API Reference</h1>
        <div className='ml-auto flex items-center gap-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <ApiDocsLayout>
          <Outlet />
        </ApiDocsLayout>
      </Main>
    </>
  )
}

export { OverviewPage } from './components/OverviewPage'
export { CreateTaskPage } from './components/CreateTaskPage'
export { GetTasksPage } from './components/GetTasksPage'
export { GetTaskPage } from './components/GetTaskPage'
export { CancelTaskPage } from './components/CancelTaskPage'
// Files
export { UploadFilePage } from './components/UploadFilePage'
export { ListFilesPage } from './components/ListFilesPage'
export { GetFilePage } from './components/GetFilePage'
export { DeleteFilePage } from './components/DeleteFilePage'
export { DownloadFilePage } from './components/DownloadFilePage'
