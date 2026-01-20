import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import Layout from './components/Layout'
import QuickStart from './components/QuickStart'
import Dashboard from './components/Dashboard'
import SessionDetail from './components/SessionDetail'
import Settings from './components/Settings'
import ProfileList from './components/ProfileList'
import ProfileDetail from './components/ProfileDetail'
import TaskList from './components/TaskList'
import WebhookList from './components/WebhookList'
import MCPServerList from './components/MCPServerList'
import SkillList from './components/SkillList'
import CredentialList from './components/CredentialList'
import ImageList from './components/ImageList'
import SystemMaintenance from './components/SystemMaintenance'
import ApiPlayground from './components/ApiPlayground'

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
        <Routes>
          <Route element={<Layout />}>
            {/* Workspace */}
            <Route path="/quick-start" element={<QuickStart />} />
            <Route path="/playground" element={<ApiPlayground />} />
            <Route path="/" element={<Dashboard />} />
            <Route path="/sessions/:sessionId" element={<SessionDetail />} />
            <Route path="/profiles" element={<ProfileList />} />
            <Route path="/profiles/:profileId" element={<ProfileDetail />} />
            <Route path="/tasks" element={<TaskList />} />
            <Route path="/webhooks" element={<WebhookList />} />
            {/* Admin */}
            <Route path="/mcp-servers" element={<MCPServerList />} />
            <Route path="/skills" element={<SkillList />} />
            <Route path="/credentials" element={<CredentialList />} />
            <Route path="/images" element={<ImageList />} />
            <Route path="/system" element={<SystemMaintenance />} />
            {/* Settings */}
            <Route path="/settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App
