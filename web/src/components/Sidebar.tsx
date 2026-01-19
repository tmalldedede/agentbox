import { useLocation, useNavigate } from 'react-router-dom'
import {
  Zap,
  Terminal,
  Layers,
  Server,
  Key,
  Box,
  Activity,
  Settings,
  ChevronLeft,
  ChevronRight,
  ListTodo,
  Webhook,
  Wrench,
  Rocket,
} from 'lucide-react'
import { useState, useEffect } from 'react'
import { api } from '../services/api'

interface NavItem {
  path: string
  icon: React.ReactNode
  label: string
  badge?: number
}

interface NavSection {
  title: string
  items: NavItem[]
}

export default function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const [collapsed, setCollapsed] = useState(false)
  const [sessionCount, setSessionCount] = useState(0)
  const [taskCount, setTaskCount] = useState(0)

  useEffect(() => {
    const fetchCounts = async () => {
      try {
        const [sessions, tasksResult] = await Promise.all([
          api.listSessions(),
          api.listTasks({ limit: 100 }),
        ])
        setSessionCount(sessions?.length || 0)
        const runningTasks = tasksResult?.tasks?.filter(
          t => t.status === 'running' || t.status === 'queued'
        ) || []
        setTaskCount(runningTasks.length)
      } catch {
        // Ignore errors
      }
    }
    fetchCounts()
    const interval = setInterval(fetchCounts, 5000)
    return () => clearInterval(interval)
  }, [])

  const sections: NavSection[] = [
    {
      title: 'Workspace',
      items: [
        { path: '/quick-start', icon: <Rocket className="w-5 h-5" />, label: 'Quick Start' },
        { path: '/', icon: <Terminal className="w-5 h-5" />, label: 'Sessions', badge: sessionCount },
        { path: '/profiles', icon: <Layers className="w-5 h-5" />, label: 'Profiles' },
        { path: '/tasks', icon: <ListTodo className="w-5 h-5" />, label: 'Tasks', badge: taskCount },
        { path: '/webhooks', icon: <Webhook className="w-5 h-5" />, label: 'Webhooks' },
      ],
    },
    {
      title: 'Admin',
      items: [
        { path: '/mcp-servers', icon: <Server className="w-5 h-5" />, label: 'MCP Servers' },
        { path: '/skills', icon: <Zap className="w-5 h-5" />, label: 'Skills' },
        { path: '/credentials', icon: <Key className="w-5 h-5" />, label: 'Credentials' },
        { path: '/images', icon: <Box className="w-5 h-5" />, label: 'Images' },
        { path: '/system', icon: <Activity className="w-5 h-5" />, label: 'System' },
      ],
    },
  ]

  const bottomItems: NavItem[] = [
    { path: '/settings', icon: <Settings className="w-5 h-5" />, label: 'Settings' },
  ]

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/' || location.pathname.startsWith('/sessions')
    }
    return location.pathname.startsWith(path)
  }

  return (
    <aside
      className={`sidebar ${collapsed ? 'sidebar-collapsed' : ''}`}
      style={{
        width: collapsed ? '64px' : '220px',
        minWidth: collapsed ? '64px' : '220px',
      }}
    >
      {/* Logo */}
      <div className="sidebar-header">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-emerald-500/20 flex items-center justify-center flex-shrink-0">
            <Zap className="w-5 h-5 text-emerald-400" />
          </div>
          {!collapsed && (
            <span className="text-lg font-bold text-primary">AgentBox</span>
          )}
        </div>
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="sidebar-toggle"
        >
          {collapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <ChevronLeft className="w-4 h-4" />
          )}
        </button>
      </div>

      {/* Navigation */}
      <nav className="sidebar-nav">
        {sections.map((section, sectionIndex) => (
          <div key={section.title} className={sectionIndex > 0 ? 'mt-4' : ''}>
            {!collapsed && (
              <div className="px-3 py-2">
                <span className="text-xs font-semibold text-muted uppercase tracking-wider flex items-center gap-2">
                  {section.title === 'Admin' && <Wrench className="w-3 h-3" />}
                  {section.title}
                </span>
              </div>
            )}
            {collapsed && sectionIndex > 0 && (
              <div className="mx-3 my-2 border-t border-border" />
            )}
            {section.items.map((item) => (
              <button
                key={item.path}
                onClick={() => navigate(item.path)}
                className={`sidebar-item ${isActive(item.path) ? 'sidebar-item-active' : ''}`}
                title={collapsed ? item.label : undefined}
              >
                <span className="sidebar-item-icon">{item.icon}</span>
                {!collapsed && (
                  <>
                    <span className="sidebar-item-label">{item.label}</span>
                    {item.badge !== undefined && item.badge > 0 && (
                      <span className="sidebar-badge">{item.badge}</span>
                    )}
                  </>
                )}
                {collapsed && item.badge !== undefined && item.badge > 0 && (
                  <span className="sidebar-badge-dot" />
                )}
              </button>
            ))}
          </div>
        ))}

        {/* Bottom items */}
        <div className="mt-auto pt-4">
          {!collapsed && (
            <div className="mx-3 mb-2 border-t border-border" />
          )}
          {bottomItems.map((item) => (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className={`sidebar-item ${isActive(item.path) ? 'sidebar-item-active' : ''}`}
              title={collapsed ? item.label : undefined}
            >
              <span className="sidebar-item-icon">{item.icon}</span>
              {!collapsed && (
                <span className="sidebar-item-label">{item.label}</span>
              )}
            </button>
          ))}
        </div>
      </nav>

      {/* Footer */}
      <div className="sidebar-footer">
        {!collapsed && (
          <div className="text-xs text-muted">
            <p>AgentBox v0.1.0</p>
          </div>
        )}
      </div>
    </aside>
  )
}
