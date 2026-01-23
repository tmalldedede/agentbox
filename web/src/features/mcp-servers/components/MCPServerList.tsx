import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  Copy,
  Trash2,
  Server,
  Database,
  Globe,
  Wrench,
  Monitor,
  Brain,
  Box,
  Loader2,
  Lock,
  Power,
  PowerOff,
  Play,
  CheckCircle,
  XCircle,
  MoreVertical,
  Edit,
  Plus,
  LayoutGrid,
  List,
} from 'lucide-react'
import type { MCPServer, MCPCategory } from '@/types'
import { useMCPServers, useUpdateMCPServer } from '@/hooks'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useMCPServersContext } from './mcp-servers-provider'

type ViewMode = 'grid' | 'table'

// Category icon mapping
const categoryIcons: Record<MCPCategory, React.ReactNode> = {
  filesystem: <Server className='w-5 h-5' />,
  database: <Database className='w-5 h-5' />,
  api: <Globe className='w-5 h-5' />,
  tool: <Wrench className='w-5 h-5' />,
  browser: <Monitor className='w-5 h-5' />,
  memory: <Brain className='w-5 h-5' />,
  other: <Box className='w-5 h-5' />,
}

const categoryIconsSmall: Record<MCPCategory, React.ReactNode> = {
  filesystem: <Server className='w-4 h-4' />,
  database: <Database className='w-4 h-4' />,
  api: <Globe className='w-4 h-4' />,
  tool: <Wrench className='w-4 h-4' />,
  browser: <Monitor className='w-4 h-4' />,
  memory: <Brain className='w-4 h-4' />,
  other: <Box className='w-4 h-4' />,
}

// Category color mapping
const categoryBgColors: Record<MCPCategory, string> = {
  filesystem: 'bg-blue-500/20 text-blue-500',
  database: 'bg-purple-500/20 text-purple-500',
  api: 'bg-emerald-500/20 text-emerald-500',
  tool: 'bg-amber-500/20 text-amber-500',
  browser: 'bg-cyan-500/20 text-cyan-500',
  memory: 'bg-pink-500/20 text-pink-500',
  other: 'bg-gray-500/20 text-gray-500',
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
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between'>
          <div className='flex items-center gap-3'>
            <div className={`w-10 h-10 rounded-lg ${bgColor} flex items-center justify-center`}>
              {icon}
            </div>
            <div>
              <div className='flex items-center gap-2'>
                <CardTitle className='text-base'>{server.name}</CardTitle>
                {server.is_built_in && (
                  <Badge variant='secondary' className='text-xs'>
                    <Lock className='w-3 h-3 mr-1' />
                    Built-in
                  </Badge>
                )}
              </div>
              <p className='text-xs text-muted-foreground font-mono'>{server.id}</p>
            </div>
          </div>
          <MCPServerActions
            server={server}
            onClone={onClone}
            onDelete={onDelete}
            onToggle={onToggle}
            onTest={handleTest}
            onClick={onClick}
            testing={testing}
            testResult={testResult}
          />
        </div>
      </CardHeader>
      <CardContent>
        {server.description && (
          <CardDescription className='mb-3 line-clamp-2'>
            {server.description}
          </CardDescription>
        )}
        <div className='flex items-center gap-2 flex-wrap'>
          {server.is_enabled ? (
            <Badge variant='default' className='bg-green-500 text-xs'>
              <Power className='w-3 h-3 mr-1' />
              Enabled
            </Badge>
          ) : (
            <Badge variant='secondary' className='text-xs'>
              <PowerOff className='w-3 h-3 mr-1' />
              Disabled
            </Badge>
          )}
          <Badge variant='outline' className='text-xs capitalize'>
            {server.category}
          </Badge>
          <Badge variant='outline' className='text-xs'>
            {server.type}
          </Badge>
          {server.tags?.slice(0, 2).map(tag => (
            <Badge key={tag} variant='outline' className='text-xs'>
              {tag}
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

// Shared dropdown actions
function MCPServerActions({
  server,
  onClone,
  onDelete,
  onToggle,
  onTest,
  onClick,
  testing,
  testResult,
}: {
  server: MCPServer
  onClone: () => void
  onDelete: () => void
  onToggle: () => void
  onTest: () => void
  onClick: () => void
  testing: boolean
  testResult: 'ok' | 'error' | null
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
        <Button variant='ghost' size='icon' className='h-8 w-8'>
          <MoreVertical className='w-4 h-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onClick() }}>
          <Edit className='w-4 h-4 mr-2' />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onTest() }} disabled={testing}>
          {testing ? (
            <Loader2 className='w-4 h-4 mr-2 animate-spin' />
          ) : testResult === 'ok' ? (
            <CheckCircle className='w-4 h-4 mr-2 text-green-500' />
          ) : testResult === 'error' ? (
            <XCircle className='w-4 h-4 mr-2 text-red-500' />
          ) : (
            <Play className='w-4 h-4 mr-2' />
          )}
          Test Connection
        </DropdownMenuItem>
        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onToggle() }}>
          {server.is_enabled ? (
            <>
              <PowerOff className='w-4 h-4 mr-2' />
              Disable
            </>
          ) : (
            <>
              <Power className='w-4 h-4 mr-2' />
              Enable
            </>
          )}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onClone() }}>
          <Copy className='w-4 h-4 mr-2' />
          Clone
        </DropdownMenuItem>
        {!server.is_built_in && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className='text-red-600'
              onClick={(e) => { e.stopPropagation(); onDelete() }}
            >
              <Trash2 className='w-4 h-4 mr-2' />
              Delete
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// Table row with actions
function MCPServerTableRow({
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

  const icon = categoryIconsSmall[server.category] || categoryIconsSmall.other
  const bgColor = categoryBgColors[server.category] || categoryBgColors.other

  return (
    <TableRow
      className={`cursor-pointer ${!server.is_enabled ? 'opacity-60' : ''}`}
      onClick={onClick}
    >
      <TableCell>
        <div className='flex items-center gap-3'>
          <div className={`w-8 h-8 rounded-md ${bgColor} flex items-center justify-center shrink-0`}>
            {icon}
          </div>
          <div className='min-w-0'>
            <div className='flex items-center gap-2'>
              <span className='font-medium truncate'>{server.name}</span>
              {server.is_built_in && (
                <Badge variant='secondary' className='text-xs shrink-0'>
                  <Lock className='w-3 h-3 mr-1' />
                  Built-in
                </Badge>
              )}
            </div>
            <p className='text-xs text-muted-foreground font-mono truncate'>{server.id}</p>
          </div>
        </div>
      </TableCell>
      <TableCell>
        <Badge variant='outline' className='text-xs capitalize'>{server.category}</Badge>
      </TableCell>
      <TableCell>
        <Badge variant='outline' className='text-xs'>{server.type}</Badge>
      </TableCell>
      <TableCell>
        {server.is_enabled ? (
          <Badge variant='default' className='bg-green-500 text-xs'>Enabled</Badge>
        ) : (
          <Badge variant='secondary' className='text-xs'>Disabled</Badge>
        )}
      </TableCell>
      <TableCell className='text-right'>
        <MCPServerActions
          server={server}
          onClone={onClone}
          onDelete={onDelete}
          onToggle={onToggle}
          onTest={handleTest}
          onClick={onClick}
          testing={testing}
          testResult={testResult}
        />
      </TableCell>
    </TableRow>
  )
}

export default function MCPServerList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [viewMode, setViewMode] = useState<ViewMode>('grid')
  const { setCurrentRow, setOpen } = useMCPServersContext()

  const { data: servers = [], isLoading } = useMCPServers()
  const updateServer = useUpdateMCPServer()

  const handleClone = async (server: MCPServer) => {
    const newId = `${server.id}-copy-${Date.now()}`
    const newName = `${server.name} (Copy)`
    try {
      await api.cloneMCPServer(server.id, { new_id: newId, new_name: newName })
      toast.success('MCP Server cloned')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to clone MCP server')
    }
  }

  const handleDelete = (server: MCPServer) => {
    setCurrentRow(server)
    setOpen('delete')
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

  // Group by category (for grid view)
  const categories = Array.from(new Set(filteredServers.map(s => s.category)))
  const groupedServers = categories.reduce(
    (acc, category) => {
      acc[category] = filteredServers.filter(s => s.category === category)
      return acc
    },
    {} as Record<MCPCategory, MCPServer[]>
  )

  return (
    <div className='space-y-4'>
      {/* Toolbar */}
      <div className='flex items-center justify-between gap-2'>
        <Select value={filter} onValueChange={(v) => setFilter(v as typeof filter)}>
          <SelectTrigger className='w-[120px]'>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value='all'>All</SelectItem>
            <SelectItem value='enabled'>Enabled</SelectItem>
            <SelectItem value='disabled'>Disabled</SelectItem>
          </SelectContent>
        </Select>

        <div className='flex items-center rounded-md border'>
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size='icon'
            className='h-8 w-8 rounded-r-none'
            onClick={() => setViewMode('grid')}
          >
            <LayoutGrid className='h-4 w-4' />
          </Button>
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size='icon'
            className='h-8 w-8 rounded-l-none'
            onClick={() => setViewMode('table')}
          >
            <List className='h-4 w-4' />
          </Button>
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className='flex items-center justify-center h-48'>
          <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
        </div>
      ) : filteredServers.length === 0 ? (
        <div className='flex flex-col items-center justify-center h-48 text-center'>
          <Server className='w-16 h-16 text-muted-foreground mb-4' />
          <p className='text-muted-foreground text-lg'>No MCP servers found</p>
          <p className='text-muted-foreground mt-2 text-sm'>
            {filter !== 'all'
              ? 'Try changing the filter or create a new server'
              : 'Create your first MCP server to get started'}
          </p>
          <Button className='mt-4' size='sm' onClick={() => navigate({ to: '/mcp-servers/new' })}>
            <Plus className='w-4 h-4 mr-2' />
            Create MCP Server
          </Button>
        </div>
      ) : viewMode === 'grid' ? (
        /* Grid View - grouped by category */
        <div className='space-y-8'>
          {categories.map(category => (
            <div key={category}>
              <div className='flex items-center gap-3 mb-4'>
                <div
                  className={`w-8 h-8 rounded-lg flex items-center justify-center ${categoryBgColors[category]}`}
                >
                  {categoryIcons[category]}
                </div>
                <h3 className='text-lg font-semibold capitalize'>{category}</h3>
                <span className='text-sm text-muted-foreground'>({groupedServers[category].length})</span>
              </div>
              <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
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
      ) : (
        /* Table View */
        <div className='rounded-md border'>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Server</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className='text-right w-[50px]' />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredServers.map(server => (
                <MCPServerTableRow
                  key={server.id}
                  server={server}
                  onClone={() => handleClone(server)}
                  onDelete={() => handleDelete(server)}
                  onToggle={() => handleToggle(server)}
                  onTest={() => handleTest(server)}
                  onClick={() => navigate({ to: `/mcp-servers/${server.id}` })}
                />
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
