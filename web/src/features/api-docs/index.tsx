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
// Auth
export { LoginPage } from './components/LoginPage'
export { ApiKeysPage } from './components/ApiKeysPage'
export { CreateApiKeyPage } from './components/CreateApiKeyPage'
export { DeleteApiKeyPage } from './components/DeleteApiKeyPage'
// Tasks
export { CreateTaskPage } from './components/CreateTaskPage'
export { GetTasksPage } from './components/GetTasksPage'
export { GetTaskPage } from './components/GetTaskPage'
export { CancelTaskPage } from './components/CancelTaskPage'
export { StreamEventsPage } from './components/StreamEventsPage'
// Files
export { UploadFilePage } from './components/UploadFilePage'
export { ListFilesPage } from './components/ListFilesPage'
export { GetFilePage } from './components/GetFilePage'
export { DeleteFilePage } from './components/DeleteFilePage'
export { DownloadFilePage } from './components/DownloadFilePage'
// Batches
export { CreateBatchPage } from './components/CreateBatchPage'
export { ListBatchesPage } from './components/ListBatchesPage'
export { GetBatchPage } from './components/GetBatchPage'
export { StartBatchPage, PauseBatchPage, CancelBatchPage, DeleteBatchPage } from './components/BatchActionsPage'
export { ListBatchTasksPage } from './components/ListBatchTasksPage'
export { StreamBatchEventsPage } from './components/StreamBatchEventsPage'
// Webhooks
export { CreateWebhookPage } from './components/CreateWebhookPage'
export { ListWebhooksPage } from './components/ListWebhooksPage'
export { GetWebhookPage } from './components/GetWebhookPage'
export { DeleteWebhookPage } from './components/DeleteWebhookPage'
