import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Copy,
  Trash2,
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
  MoreVertical,
  Edit,
} from 'lucide-react'
import type { MCPServer, MCPCategory } from '@/types'
import { useMCPServers, useUpdateMCPServer, useDeleteMCPServer } from '@/hooks'
import { api } from '@/services/api'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

// Category icon mapping
const categoryIcons: Record<MCPCategory, React.ReactNode> = {
  filesystem: <Server className="w-5 h-5" />,
  database: <Database className="w-5 h-5" />,
  api: <Globe className="w-5 h-5" />,
  tool: <Wrench className="w-5 h-5" />,
  browser: <Monitor className="w-5 h-5" />,
  memory: <Brain className="w-5 h-5" />,
  other: <Box className="w-5 h-5" />,
}

// Category color mapping
const categoryBgColors: Record<MCPCategory, string> = {
  filesystem: 'bg-blue-500/20 text-blue-400',
  database: 'bg-purple-500/20 text-purple-400',
  api: 'bg-emerald-500/20 text-emerald-400',
  tool: 'bg-amber-500/20 text-amber-400',
  browser: 'bg-cyan-500/20 text-cyan-400',
  memory: 'bg-pink-500/20 text-pink-400',
  other: 'bg-gray-500/20 text-gray-400',
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

  const bgColor = categoryBgColors[server.category] || categoryBgColors.other
  const icon = categoryIcons[server.category] || categoryIcons.other

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      await onTest()
      setTestResult('ok')
      toast.success('Connection test successful')
    } catch {
      setTestResult('error')
      toast.error('Connection test failed')
    } finally {
      setTesting(false)
      setTimeout(() => setTestResult(null), 3000)
    }
  }

  return (
    <Card
      className={`cursor-pointer transition-colors ${
        server.is_enabled
          ? 'hover:border-primary/50'
          : 'opacity-60 hover:border-muted-foreground/50'
      }`}
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-lg ${bgColor} flex items-center justify-center`}>
              {icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{server.name}</CardTitle>
                {server.is_built_in && (
                  <Badge variant="secondary" className="text-xs">
                    <Lock className="w-3 h-3 mr-1" />
                    Built-in
                  </Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground font-mono">{server.id}</p>
            </div>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onClick()
              }}>
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                handleTest()
              }} disabled={testing}>
                {testing ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : testResult === 'ok' ? (
                  <CheckCircle className="w-4 h-4 mr-2 text-green-500" />
                ) : testResult === 'error' ? (
                  <XCircle className="w-4 h-4 mr-2 text-red-500" />
                ) : (
                  <Play className="w-4 h-4 mr-2" />
                )}
                Test Connection
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onToggle()
              }}>
                {server.is_enabled ? (
                  <>
                    <PowerOff className="w-4 h-4 mr-2" />
                    Disable
                  </>
                ) : (
                  <>
                    <Power className="w-4 h-4 mr-2" />
                    Enable
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onClone()
              }}>
                <Copy className="w-4 h-4 mr-2" />
                Clone
              </DropdownMenuItem>
              {!server.is_built_in && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-red-600"
                    onClick={(e) => {
                      e.stopPropagation()
                      onDelete()
                    }}
                  >
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent>
        {server.description && (
          <CardDescription className="mb-3 line-clamp-2">
            {server.description}
          </CardDescription>
        )}
        <div className="flex items-center gap-2 flex-wrap">
          {server.is_enabled ? (
            <Badge variant="default" className="bg-green-500 text-xs">
              <Power className="w-3 h-3 mr-1" />
              Enabled
            </Badge>
          ) : (
            <Badge variant="secondary" className="text-xs">
              <PowerOff className="w-3 h-3 mr-1" />
              Disabled
            </Badge>
          )}
          <Badge variant="outline" className="text-xs capitalize">
            {server.category}
          </Badge>
          <Badge variant="outline" className="text-xs">
            {server.type}
          </Badge>
          {server.tags?.slice(0, 2).map(tag => (
            <Badge key={tag} variant="outline" className="text-xs">
              {tag}
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
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
            <Server className="w-6 h-6 text-blue-400" />
            <span className="text-lg font-bold">MCP Servers</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Select value={filter} onValueChange={(v) => setFilter(v as typeof filter)}>
            <SelectTrigger className="w-[120px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="enabled">Enabled</SelectItem>
              <SelectItem value="disabled">Disabled</SelectItem>
            </SelectContent>
          </Select>

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
            <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
          </div>
        ) : filteredServers.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Server className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg">No MCP servers found</p>
            <p className="text-muted-foreground mt-2">
              {filter !== 'all'
                ? 'Try changing the filter or create a new server'
                : 'Create your first MCP server to get started'}
            </p>
            <Button className="mt-4" onClick={() => navigate({ to: '/mcp-servers/new' })}>
              <Plus className="w-4 h-4 mr-2" />
              Create MCP Server
            </Button>
          </div>
        ) : (
          <div className="space-y-8">
            {categories.map(category => (
              <div key={category}>
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center ${categoryBgColors[category]}`}
                  >
                    {categoryIcons[category]}
                  </div>
                  <h2 className="text-lg font-semibold text-foreground capitalize">{category}</h2>
                  <span className="text-sm text-muted-foreground">({groupedServers[category].length})</span>
                </div>
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
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
