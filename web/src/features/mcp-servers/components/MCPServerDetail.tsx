import { useEffect, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Save,
  Trash2,
  Loader2,
  AlertCircle,
  Lock,
  Plus,
  X,
  Play,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import type { MCPServer, MCPServerType, MCPCategory, MCPTestResult } from '@/types'
import { useMCPServers, useUpdateMCPServer, useDeleteMCPServer } from '@/hooks'
import { api } from '@/services/api'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
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
  Alert,
  AlertDescription,
  AlertTitle,
} from '@/components/ui/alert'
import { Switch } from '@/components/ui/switch'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

const categories: MCPCategory[] = ['filesystem', 'database', 'api', 'tool', 'browser', 'memory', 'other']
const serverTypes: MCPServerType[] = ['stdio', 'sse', 'http']

type MCPServerDetailProps = {
  serverId: string
}

export default function MCPServerDetail({ serverId }: MCPServerDetailProps) {
  const navigate = useNavigate()
  const { data: servers = [], isLoading } = useMCPServers()
  const updateServer = useUpdateMCPServer()
  const deleteServer = useDeleteMCPServer()
  const resetTimeoutRef = useRef<number | null>(null)

  const server = servers.find(s => s.id === serverId)

  const [formData, setFormData] = useState<Partial<MCPServer>>({})
  const [isDirty, setIsDirty] = useState(false)
  const [newTag, setNewTag] = useState('')
  const [newEnvKey, setNewEnvKey] = useState('')
  const [newEnvValue, setNewEnvValue] = useState('')
  const [newArg, setNewArg] = useState('')
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<MCPTestResult | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  useEffect(() => {
    if (server) {
      setFormData(server)
    }
  }, [server])

  useEffect(() => {
    return () => {
      if (resetTimeoutRef.current) {
        window.clearTimeout(resetTimeoutRef.current)
      }
    }
  }, [])

  if (isLoading) {
    return (
      <div className='flex items-center justify-center h-48'>
        <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  if (!server) {
    return (
      <div className='flex flex-col items-center justify-center h-48 text-center'>
        <AlertCircle className='h-16 w-16 text-muted-foreground mb-4' />
        <p className='text-lg text-muted-foreground'>MCP Server not found</p>
        <Button variant='outline' className='mt-4' onClick={() => navigate({ to: '/mcp-servers' })}>
          Back to MCP Servers
        </Button>
      </div>
    )
  }

  const isBuiltIn = server.is_built_in
  // Built-in 服务器只允许修改 Env 和 IsEnabled，其他字段只读
  const isReadOnly = isBuiltIn
  const missingEnvKeys = Object.entries(formData.env || {}).filter(([, v]) => !v).map(([k]) => k)

  const handleSave = async () => {
    if (!serverId) return
    try {
      await updateServer.mutateAsync({
        id: serverId,
        data: {
          name: formData.name,
          description: formData.description,
          command: formData.command,
          args: formData.args,
          env: formData.env,
          work_dir: formData.work_dir,
          type: formData.type,
          category: formData.category,
          tags: formData.tags,
          is_enabled: formData.is_enabled,
        },
      })
      setIsDirty(false)
      toast.success('MCP Server updated successfully')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update MCP server')
    }
  }

  const handleDelete = async () => {
    if (!serverId) return
    try {
      await deleteServer.mutateAsync(serverId)
      toast.success('MCP Server deleted')
      navigate({ to: '/mcp-servers' })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete MCP server')
    }
  }

  const handleTest = async () => {
    if (!serverId) return
    setTesting(true)
    setTestResult(null)
    try {
      const result = await api.testMCPServer(serverId)
      setTestResult(result)
      if (result.status === 'ok') {
        const latencyInfo = result.latency_ms ? ` (${result.latency_ms}ms)` : ''
        toast.success(`Connection test passed${latencyInfo}`)
      } else {
        toast.error(`Test failed: ${result.error || 'Unknown error'}`)
      }
    } catch (err) {
      setTestResult({ status: 'error', latency_ms: 0, error: err instanceof Error ? err.message : 'Connection test failed' })
      toast.error(err instanceof Error ? err.message : 'Connection test failed')
    } finally {
      setTesting(false)
      if (resetTimeoutRef.current) {
        window.clearTimeout(resetTimeoutRef.current)
      }
      resetTimeoutRef.current = window.setTimeout(() => setTestResult(null), 10000)
    }
  }

  const updateField = <K extends keyof MCPServer>(key: K, value: MCPServer[K]) => {
    setFormData(prev => ({ ...prev, [key]: value }))
    setIsDirty(true)
  }

  const addTag = () => {
    if (!newTag.trim()) return
    const tags = formData.tags || []
    if (tags.includes(newTag.trim())) {
      toast.error('Tag already exists')
      return
    }
    updateField('tags', [...tags, newTag.trim()])
    setNewTag('')
  }

  const removeTag = (tag: string) => {
    updateField('tags', (formData.tags || []).filter(t => t !== tag))
  }

  const addEnv = () => {
    if (!newEnvKey.trim()) return
    const env = formData.env || {}
    if (env[newEnvKey.trim()]) {
      toast.error('Environment variable already exists')
      return
    }
    updateField('env', { ...env, [newEnvKey.trim()]: newEnvValue.trim() })
    setNewEnvKey('')
    setNewEnvValue('')
  }

  const removeEnv = (key: string) => {
    const env = { ...(formData.env || {}) }
    delete env[key]
    updateField('env', env)
  }

  const addArg = () => {
    if (!newArg.trim()) return
    const args = formData.args || []
    updateField('args', [...args, newArg.trim()])
    setNewArg('')
  }

  const removeArg = (index: number) => {
    updateField('args', (formData.args || []).filter((_, i) => i !== index))
  }

  return (
    <>
      <div className='space-y-6'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div className='flex items-center gap-3'>
            <Button variant='ghost' size='icon' onClick={() => navigate({ to: '/mcp-servers' })}>
              <ArrowLeft className='h-5 w-5' />
            </Button>
            <h2 className='text-2xl font-bold tracking-tight'>MCP Server Detail</h2>
            {isBuiltIn && (
              <Badge variant='secondary'>
                <Lock className='mr-1 h-3 w-3' />
                Built-in
              </Badge>
            )}
            {!server.is_configured && (
              <Badge variant='destructive'>Needs Config</Badge>
            )}
            <div className='flex items-center gap-2 ml-2'>
              <Switch
                checked={formData.is_enabled ?? server.is_enabled}
                onCheckedChange={v => updateField('is_enabled', v)}
              />
              <span className='text-sm text-muted-foreground'>
                {formData.is_enabled ?? server.is_enabled ? 'Enabled' : 'Disabled'}
              </span>
            </div>
          </div>
          <div className='flex items-center gap-2'>
            <Button variant='outline' size='sm' onClick={handleTest} disabled={testing}>
              {testing ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : testResult?.status === 'ok' ? (
                <CheckCircle className='mr-2 h-4 w-4 text-green-500' />
              ) : testResult?.status === 'error' ? (
                <XCircle className='mr-2 h-4 w-4 text-red-500' />
              ) : (
                <Play className='mr-2 h-4 w-4' />
              )}
              Test
            </Button>
            {!isBuiltIn && (
              <Button
                variant='outline'
                size='sm'
                onClick={() => setShowDeleteDialog(true)}
                disabled={deleteServer.isPending}
                className='text-destructive hover:text-destructive'
              >
                <Trash2 className='mr-2 h-4 w-4' />
                Delete
              </Button>
            )}
            <Button
              size='sm'
              onClick={handleSave}
              disabled={!isDirty || updateServer.isPending}
            >
              {updateServer.isPending ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <Save className='mr-2 h-4 w-4' />
              )}
              Save
            </Button>
          </div>
        </div>

        {/* 配置状态警告 */}
        {!server.is_configured && (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Configuration Required</AlertTitle>
            <AlertDescription>
              Missing environment variables: <span className='font-mono font-semibold'>{missingEnvKeys.join(', ')}</span>.
              Please fill in all required values to use this MCP server.
            </AlertDescription>
          </Alert>
        )}

        {/* 测试结果详情 */}
        {testResult && (
          <Alert variant={testResult.status === 'ok' ? 'default' : 'destructive'}>
            {testResult.status === 'ok' ? (
              <CheckCircle className='h-4 w-4' />
            ) : (
              <XCircle className='h-4 w-4' />
            )}
            <AlertTitle>
              {testResult.status === 'ok' ? 'Connection Successful' : 'Connection Failed'}
              {testResult.latency_ms > 0 && (
                <span className='ml-2 text-sm font-normal text-muted-foreground'>
                  ({testResult.latency_ms}ms)
                </span>
              )}
            </AlertTitle>
            <AlertDescription className='space-y-1'>
              {testResult.error && <p>{testResult.error}</p>}
              {testResult.server_info && (
                <p className='text-sm'>
                  Server: {(testResult.server_info as Record<string, string>).name || 'unknown'}
                  {(testResult.server_info as Record<string, string>).version && ` v${(testResult.server_info as Record<string, string>).version}`}
                </p>
              )}
              {testResult.capabilities && Object.keys(testResult.capabilities).length > 0 && (
                <p className='text-sm'>
                  Capabilities: {Object.keys(testResult.capabilities).join(', ')}
                </p>
              )}
            </AlertDescription>
          </Alert>
        )}

        <Card>
          <CardHeader>
            <CardTitle>Basic Information</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <Label>Server ID</Label>
              <Input value={server.id} disabled className='font-mono' />
            </div>

            <div className='space-y-2'>
              <Label>Name *</Label>
              <Input
                value={formData.name || ''}
                onChange={e => updateField('name', e.target.value)}
                disabled={isReadOnly}
                placeholder='Server name'
              />
            </div>

            <div className='space-y-2'>
              <Label>Description</Label>
              <Textarea
                value={formData.description || ''}
                onChange={e => updateField('description', e.target.value)}
                disabled={isReadOnly}
                placeholder='What does this server do?'
                rows={3}
              />
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <div className='space-y-2'>
                <Label>Type *</Label>
                <Select
                  value={formData.type || 'stdio'}
                  onValueChange={v => updateField('type', v as MCPServerType)}
                  disabled={isReadOnly}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {serverTypes.map(type => (
                      <SelectItem key={type} value={type}>{type}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className='space-y-2'>
                <Label>Category *</Label>
                <Select
                  value={formData.category || 'other'}
                  onValueChange={v => updateField('category', v as MCPCategory)}
                  disabled={isReadOnly}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {categories.map(cat => (
                      <SelectItem key={cat} value={cat} className='capitalize'>{cat}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Command & Arguments</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='space-y-2'>
              <Label>Command *</Label>
              <Input
                value={formData.command || ''}
                onChange={e => updateField('command', e.target.value)}
                disabled={isReadOnly}
                className='font-mono'
                placeholder='npx -y @modelcontextprotocol/server-filesystem'
              />
            </div>

            <div className='space-y-2'>
              <Label>Arguments</Label>
              <div className='space-y-2'>
                {(formData.args || []).map((arg, index) => (
                  <div key={index} className='flex gap-2'>
                    <Input value={arg} disabled className='flex-1 font-mono text-sm' />
                    {!isReadOnly && (
                      <Button variant='ghost' size='icon' onClick={() => removeArg(index)}>
                        <X className='h-4 w-4 text-destructive' />
                      </Button>
                    )}
                  </div>
                ))}
                {!isReadOnly && (
                  <div className='flex gap-2'>
                    <Input
                      value={newArg}
                      onChange={e => setNewArg(e.target.value)}
                      onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addArg() } }}
                      placeholder='Add argument...'
                      className='flex-1 font-mono text-sm'
                    />
                    <Button variant='secondary' size='sm' onClick={addArg}>
                      <Plus className='mr-1 h-4 w-4' />
                      Add
                    </Button>
                  </div>
                )}
              </div>
            </div>

            <div className='space-y-2'>
              <Label>Working Directory</Label>
              <Input
                value={formData.work_dir || ''}
                onChange={e => updateField('work_dir', e.target.value)}
                disabled={isReadOnly}
                className='font-mono'
                placeholder='/path/to/workdir'
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className='flex items-center gap-2'>
              Environment Variables
              {missingEnvKeys.length > 0 && (
                <Badge variant='destructive' className='text-xs'>
                  {missingEnvKeys.length} missing
                </Badge>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-3'>
            {Object.entries(formData.env || {}).map(([key, value]) => (
              <div key={key} className='flex gap-2 items-center'>
                <Input value={key} disabled className='flex-1 font-mono text-sm' />
                <div className='flex-1 relative'>
                  <Input
                    value={value}
                    onChange={e => {
                      const env = { ...(formData.env || {}) }
                      env[key] = e.target.value
                      updateField('env', env)
                    }}
                    className={`font-mono text-sm ${!value ? 'border-destructive bg-destructive/5' : ''}`}
                    placeholder='Enter value...'
                  />
                  {!value && (
                    <AlertCircle className='absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-destructive' />
                  )}
                </div>
                {!isReadOnly && (
                  <Button variant='ghost' size='icon' onClick={() => removeEnv(key)}>
                    <X className='h-4 w-4 text-destructive' />
                  </Button>
                )}
              </div>
            ))}
            {!isReadOnly && (
              <div className='flex gap-2'>
                <Input
                  value={newEnvKey}
                  onChange={e => setNewEnvKey(e.target.value)}
                  placeholder='KEY'
                  className='flex-1 font-mono text-sm'
                />
                <Input
                  value={newEnvValue}
                  onChange={e => setNewEnvValue(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addEnv() } }}
                  placeholder='value'
                  className='flex-1 font-mono text-sm'
                />
                <Button variant='secondary' size='sm' onClick={addEnv}>
                  <Plus className='mr-1 h-4 w-4' />
                  Add
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Tags</CardTitle>
          </CardHeader>
          <CardContent className='space-y-3'>
            <div className='flex flex-wrap gap-2'>
              {(formData.tags || []).map(tag => (
                <Badge key={tag} variant='secondary' className='gap-1'>
                  {tag}
                  {!isReadOnly && (
                    <button onClick={() => removeTag(tag)} className='ml-1 hover:text-destructive'>
                      <X className='h-3 w-3' />
                    </button>
                  )}
                </Badge>
              ))}
            </div>
            {!isReadOnly && (
              <div className='flex gap-2'>
                <Input
                  value={newTag}
                  onChange={e => setNewTag(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addTag() } }}
                  placeholder='Add tag...'
                  className='flex-1'
                />
                <Button variant='secondary' size='sm' onClick={addTag}>
                  <Plus className='mr-1 h-4 w-4' />
                  Add
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
          </CardHeader>
          <CardContent>
            <div className='grid grid-cols-2 gap-4 text-sm'>
              <div>
                <p className='text-muted-foreground'>Created</p>
                <p className='mt-1'>{new Date(server.created_at).toLocaleString()}</p>
              </div>
              <div>
                <p className='text-muted-foreground'>Updated</p>
                <p className='mt-1'>{new Date(server.updated_at).toLocaleString()}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete MCP Server</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete MCP server &quot;{server.name}&quot;?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <Button variant='outline' onClick={() => setShowDeleteDialog(false)}>
              Cancel
            </Button>
            <Button
              variant='destructive'
              onClick={handleDelete}
              disabled={deleteServer.isPending}
            >
              {deleteServer.isPending && (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              )}
              Delete
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
