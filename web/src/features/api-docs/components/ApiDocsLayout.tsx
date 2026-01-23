import { Link, useLocation } from '@tanstack/react-router'
import { MethodBadge } from './shared'

interface NavSection {
  title: string
  items: { path: string; label: string; method?: string }[]
}

const NAV_SECTIONS: NavSection[] = [
  {
    title: 'Tasks',
    items: [
      { path: '/api-docs/create-task', label: 'Create Task', method: 'POST' },
      { path: '/api-docs/get-tasks', label: 'List Tasks', method: 'GET' },
      { path: '/api-docs/get-task', label: 'Get Task', method: 'GET' },
      { path: '/api-docs/cancel-task', label: 'Cancel Task', method: 'DELETE' },
      { path: '/api-docs/stream-events', label: 'Stream Events', method: 'GET' },
    ],
  },
  {
    title: 'Files',
    items: [
      { path: '/api-docs/upload-file', label: 'Upload File', method: 'POST' },
      { path: '/api-docs/list-files', label: 'List Files', method: 'GET' },
      { path: '/api-docs/get-file', label: 'Get File', method: 'GET' },
      { path: '/api-docs/delete-file', label: 'Delete File', method: 'DELETE' },
      { path: '/api-docs/download-file', label: 'Download File', method: 'GET' },
    ],
  },
]

export function ApiDocsLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  const currentPath = location.pathname.replace(/\/$/, '') || '/api-docs'

  return (
    <div className="flex gap-6 max-w-6xl mx-auto">
      {/* Sidebar Navigation */}
      <nav className="hidden lg:block w-56 shrink-0 sticky top-4 self-start py-2">
        {/* Overview */}
        <Link
          to="/api-docs"
          className={`flex items-center gap-2 px-3 py-2 text-sm rounded-md transition-colors mb-4 ${
            currentPath === '/api-docs'
              ? 'bg-primary/10 text-primary font-medium'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted'
          }`}
        >
          Overview
        </Link>

        {/* API Sections */}
        {NAV_SECTIONS.map((section) => (
          <div key={section.title} className="mb-6">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 px-3">
              {section.title}
            </p>
            <ul className="space-y-0.5">
              {section.items.map((item) => {
                const isActive = currentPath === item.path
                return (
                  <li key={item.path}>
                    <Link
                      to={item.path}
                      className={`flex items-center gap-2 px-3 py-2 text-sm rounded-md transition-colors ${
                        isActive
                          ? 'bg-primary/10 text-primary font-medium'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                      }`}
                    >
                      {item.method && <MethodBadge method={item.method} small />}
                      {item.label}
                    </Link>
                  </li>
                )
              })}
            </ul>
          </div>
        ))}

        {/* Resources */}
        <div className="px-3">
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
            Resources
          </p>
          <ul className="space-y-0.5">
            <li>
              <Link
                to="/api-docs"
                hash="auth"
                className="block px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
              >
                Authentication
              </Link>
            </li>
            <li>
              <Link
                to="/api-docs"
                hash="errors"
                className="block px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
              >
                Errors
              </Link>
            </li>
          </ul>
        </div>
      </nav>

      {/* Main Content */}
      <div className="flex-1 min-w-0 py-2 pb-20">
        {children}
      </div>
    </div>
  )
}
