import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Copy,
  Trash2,
  ChevronRight,
  Server,
  Database,
  Globe,
  Wrench,
  Monitor,
  Brain,
  Box,
  RefreshCw,
  AlertCircle,
  Loader2,
  Lock,
  Power,
  PowerOff,
  Play,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import type { MCPServer, MCPCategory } from '@/types'
import { useMCPServers, useUpdateMCPServer, useDeleteMCPServer } from '@/hooks'
import { api } from '@/services/api'
import { toast } from 'sonner'

// Category icon mapping
const categoryIcons: Record<MCPCategory, React.ReactNode> = {
  filesystem: <Server className="w-4 h-4" />,
  database: <Database className="w-4 h-4" />,
  api: <Globe className="w-4 h-4" />,
  tool: <Wrench className="w-4 h-4" />,
  browser: <Monitor className="w-4 h-4" />,
  memory: <Brain className="w-4 h-4" />,
  other: <Box className="w-4 h-4" />,
}

// Category color mapping
const categoryColors: Record<MCPCategory, string> = {
  filesystem: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  database: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  api: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  tool: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  browser: 'bg-cyan-500/20 text-cyan-400 border-cyan-500/30',
  memory: 'bg-pink-500/20 text-pink-400 border-pink-500/30',
  other: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
}

// MCP Server Card component
function MCPServerCard({
  server,
  onClone,
  onDelete,
  onToggle,
  onTest,
  onClick,
}: {
  server: MCPServer
  onClone: () => void
  onDelete: () => void
  onToggle: () => void
  onTest: () => Promise<void>
  onClick: () => void
}) {
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<'ok' | 'error' | null>(null)

  const colors = categoryColors[server.category] || categoryColors.other
  const icon = categoryIcons[server.category] || categoryIcons.other

  const handleTest = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setTesting(true)
    setTestResult(null)
    try {
      await onTest()
      setTestResult('ok')
    } catch {
      setTestResult('error')
    } finally {
      setTesting(false)
      setTimeout(() => setTestResult(null), 3000)
    }
  }

  return (
    <div
      className={`card p-4 cursor-pointer group transition-colors ${
        server.is_enabled
          ? 'hover:border-emerald-500/50'
          : 'opacity-60 hover:border-gray-500/50'
      }`}
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colors}`}>
          {icon}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{server.name}</span>
            {server.is_built_in && (
              <span className="badge badge-scaling text-xs">
                <Lock className="w-3 h-3" />
                Built-in
              </span>
            )}
            {!server.is_enabled && (
              <span className="text-xs px-2 py-0.5 rounded bg-gray-500/20 text-gray-400">
                Disabled
              </span>
            )}
          </div>
          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
            {server.description || `${server.type} MCP server`}
          </p>

          <div className="flex items-center gap-2 mt-3 flex-wrap">
            <span className={`text-xs px-2 py-0.5 rounded border ${colors}`}>{server.category}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/80">
              {server.type}
            </span>
            {server.tags?.slice(0, 3).map(tag => (
              <span key={tag} className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">
                {tag}
              </span>
            ))}
          </div>
        </div>

        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          <button
            onClick={handleTest}
            className="btn btn-ghost btn-icon"
            title="Test Connection"
            disabled={testing}
          >
            {testing ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : testResult === 'ok' ? (
              <CheckCircle className="w-4 h-4 text-emerald-400" />
            ) : testResult === 'error' ? (
              <XCircle className="w-4 h-4 text-red-400" />
            ) : (
              <Play className="w-4 h-4" />
            )}
          </button>

          <button
            onClick={e => {
              e.stopPropagation()
              onToggle()
            }}
            className="btn btn-ghost btn-icon"
            title={server.is_enabled ? 'Disable' : 'Enable'}
          >
            {server.is_enabled ? (
              <Power className="w-4 h-4 text-emerald-400" />
            ) : (
              <PowerOff className="w-4 h-4 text-gray-400" />
            )}
          </button>

          <button
            onClick={e => {
              e.stopPropagation()
              onClone()
            }}
            className="btn btn-ghost btn-icon"
            title="Clone"
          >
            <Copy className="w-4 h-4" />
          </button>

          {!server.is_built_in && (
            <button
              onClick={e => {
                e.stopPropagation()
                onDelete()
              }}
              className="btn btn-ghost btn-icon text-red-400"
              title="Delete"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          )}
        </div>

        <ChevronRight className="w-5 h-5 text-muted-foreground group-hover:text-emerald-400 transition-colors" />
      </div>
    </div>
  )
}

export default function MCPServerList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')

  // React Query hooks
  const { data: servers = [], isLoading, isFetching, error, refetch } = useMCPServers()
  const updateServer = useUpdateMCPServer()
  const deleteServer = useDeleteMCPServer()

  const handleClone = async (server: MCPServer) => {
    const newId = `${server.id}-copy-${Date.now()}`
    const newName = `${server.name} (Copy)`
    try {
      await api.cloneMCPServer(server.id, { new_id: newId, new_name: newName })
      refetch()
      toast.success('MCP Server cloned')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to clone MCP server')
    }
  }

  const handleDelete = (server: MCPServer) => {
    if (!confirm(`Delete MCP server "${server.name}"?`)) return
    deleteServer.mutate(server.id)
  }

  const handleToggle = (server: MCPServer) => {
    updateServer.mutate({ id: server.id, data: { is_enabled: !server.is_enabled } })
  }

  const handleTest = async (server: MCPServer) => {
    await api.testMCPServer(server.id)
  }

  // Filter servers
  const filteredServers = servers.filter(s => {
    if (filter === 'enabled') return s.is_enabled
    if (filter === 'disabled') return !s.is_enabled
    return true
  })

  // Group by category
  const categories = Array.from(new Set(filteredServers.map(s => s.category)))
  const groupedServers = categories.reduce(
    (acc, category) => {
      acc[category] = filteredServers.filter(s => s.category === category)
      return acc
    },
    {} as Record<MCPCategory, MCPServer[]>
  )

  return (
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Server className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">MCP Servers</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={filter}
            onChange={e => setFilter(e.target.value as typeof filter)}
            className="input py-2 px-3 text-sm"
          >
            <option value="all">All</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>

          <button onClick={() => refetch()} className="btn btn-ghost btn-icon" disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button className="btn btn-primary" onClick={() => navigate({ to: '/mcp-servers/new' })}>
            <Plus className="w-4 h-4" />
            New Server
          </button>
        </div>
      </header>

      <div className="p-6">
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load MCP servers'}
            </span>
          </div>
        )}

        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">MCP Servers</h1>
          <p className="text-muted-foreground">
            Model Context Protocol (MCP) servers extend Agent capabilities with external tools, data
            sources, and integrations. Configure and manage your MCP servers here.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : filteredServers.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Server className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-foreground/90 text-lg">No MCP servers found</p>
            <p className="text-muted-foreground mt-2">
              {filter !== 'all'
                ? 'Try changing the filter or create a new server'
                : 'Create your first MCP server to get started'}
            </p>
          </div>
        ) : (
          <div className="space-y-8">
            {categories.map(category => (
              <div key={category}>
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center ${categoryColors[category]}`}
                  >
                    {categoryIcons[category]}
                  </div>
                  <h2 className="text-lg font-semibold text-foreground capitalize">{category}</h2>
                  <span className="text-sm text-muted-foreground">({groupedServers[category].length})</span>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  {groupedServers[category].map(server => (
                    <MCPServerCard
                      key={server.id}
                      server={server}
                      onClone={() => handleClone(server)}
                      onDelete={() => handleDelete(server)}
                      onToggle={() => handleToggle(server)}
                      onTest={() => handleTest(server)}
                      onClick={() => navigate({ to: `/mcp-servers/${server.id}` })}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
