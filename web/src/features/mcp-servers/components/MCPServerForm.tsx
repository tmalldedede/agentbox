import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Save, Plus, X, Loader2, Box } from 'lucide-react'
import type { MCPServerType, MCPCategory, CreateMCPServerRequest } from '@/types'
import { api } from '@/services/api'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'

const categories: MCPCategory[] = ['filesystem', 'database', 'api', 'tool', 'browser', 'memory', 'other']
const serverTypes: MCPServerType[] = ['stdio', 'sse', 'http']

export default function MCPServerForm() {
  const navigate = useNavigate()
  const [saving, setSaving] = useState(false)

  const [formData, setFormData] = useState<CreateMCPServerRequest>({
    id: '',
    name: '',
    description: '',
    command: '',
    args: [],
    env: {},
    work_dir: '',
    type: 'stdio',
    category: 'other',
    tags: [],
  })

  const [newTag, setNewTag] = useState('')
  const [newEnvKey, setNewEnvKey] = useState('')
  const [newEnvValue, setNewEnvValue] = useState('')
  const [newArg, setNewArg] = useState('')

  const updateField = <K extends keyof CreateMCPServerRequest>(
    key: K,
    value: CreateMCPServerRequest[K]
  ) => {
    setFormData(prev => ({ ...prev, [key]: value }))
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
    if (!newEnvKey.trim()) {
      toast.error('Environment variable key is required')
      return
    }
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.id.trim()) {
      toast.error('Server ID is required')
      return
    }
    if (!formData.name.trim()) {
      toast.error('Name is required')
      return
    }
    if (!formData.command.trim()) {
      toast.error('Command is required')
      return
    }

    if (!/^[a-z0-9-]+$/.test(formData.id)) {
      toast.error('Server ID must be lowercase alphanumeric with hyphens only')
      return
    }

    setSaving(true)
    try {
      await api.createMCPServer(formData)
      toast.success('MCP Server created successfully')
      navigate({ to: '/mcp-servers' })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create MCP server')
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className='space-y-6'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-3'>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            onClick={() => navigate({ to: '/mcp-servers' })}
          >
            <ArrowLeft className='h-5 w-5' />
          </Button>
          <h2 className='text-2xl font-bold tracking-tight'>Create MCP Server</h2>
        </div>
        <div className='flex items-center gap-2'>
          <Button type='button' variant='outline' size='sm' onClick={() => navigate({ to: '/mcp-servers' })}>
            Cancel
          </Button>
          <Button type='submit' size='sm' disabled={saving}>
            {saving ? <Loader2 className='mr-2 h-4 w-4 animate-spin' /> : <Save className='mr-2 h-4 w-4' />}
            Create Server
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label>Server ID *</Label>
            <Input
              value={formData.id}
              onChange={e => updateField('id', e.target.value.toLowerCase())}
              placeholder='my-mcp-server'
              className='font-mono'
              required
            />
            <p className='text-xs text-muted-foreground'>
              Lowercase alphanumeric with hyphens. Cannot be changed later.
            </p>
          </div>

          <div className='space-y-2'>
            <Label>Name *</Label>
            <Input
              value={formData.name}
              onChange={e => updateField('name', e.target.value)}
              placeholder='My MCP Server'
              required
            />
          </div>

          <div className='space-y-2'>
            <Label>Description</Label>
            <Textarea
              value={formData.description || ''}
              onChange={e => updateField('description', e.target.value)}
              placeholder='What does this MCP server do?'
              rows={3}
            />
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label>Type *</Label>
              <Select value={formData.type || 'stdio'} onValueChange={v => updateField('type', v as MCPServerType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {serverTypes.map(type => (
                    <SelectItem key={type} value={type}>{type}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className='text-xs text-muted-foreground'>
                stdio / sse / http
              </p>
            </div>

            <div className='space-y-2'>
              <Label>Category *</Label>
              <Select value={formData.category || 'other'} onValueChange={v => updateField('category', v as MCPCategory)}>
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
          <CardDescription>
            Configure how to run this MCP server.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label>Command *</Label>
            <Input
              value={formData.command}
              onChange={e => updateField('command', e.target.value)}
              placeholder='npx -y @modelcontextprotocol/server-filesystem'
              className='font-mono'
              required
            />
          </div>

          <div className='space-y-2'>
            <Label>Arguments</Label>
            <div className='space-y-2'>
              {(formData.args || []).map((arg, index) => (
                <div key={index} className='flex gap-2'>
                  <Input value={arg} disabled className='flex-1 font-mono text-sm' />
                  <Button type='button' variant='ghost' size='icon' onClick={() => removeArg(index)}>
                    <X className='h-4 w-4 text-destructive' />
                  </Button>
                </div>
              ))}
              <div className='flex gap-2'>
                <Input
                  value={newArg}
                  onChange={e => setNewArg(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addArg() } }}
                  placeholder='Add argument...'
                  className='flex-1 font-mono text-sm'
                />
                <Button type='button' variant='secondary' size='sm' onClick={addArg}>
                  <Plus className='mr-1 h-4 w-4' />
                  Add
                </Button>
              </div>
            </div>
          </div>

          <div className='space-y-2'>
            <Label>Working Directory</Label>
            <Input
              value={formData.work_dir || ''}
              onChange={e => updateField('work_dir', e.target.value)}
              placeholder='/path/to/workdir'
              className='font-mono'
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Environment Variables</CardTitle>
          <CardDescription>
            Variables available to the MCP server process.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-3'>
          {Object.entries(formData.env || {}).map(([key, value]) => (
            <div key={key} className='flex gap-2 items-center'>
              <Input value={key} disabled className='flex-1 font-mono text-sm' />
              <Input value={value} disabled className='flex-1 font-mono text-sm' />
              <Button type='button' variant='ghost' size='icon' onClick={() => removeEnv(key)}>
                <X className='h-4 w-4 text-destructive' />
              </Button>
            </div>
          ))}
          <div className='flex gap-2'>
            <Input
              value={newEnvKey}
              onChange={e => setNewEnvKey(e.target.value)}
              placeholder='VARIABLE_NAME'
              className='flex-1 font-mono text-sm'
            />
            <Input
              value={newEnvValue}
              onChange={e => setNewEnvValue(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addEnv() } }}
              placeholder='value'
              className='flex-1 font-mono text-sm'
            />
            <Button type='button' variant='secondary' size='sm' onClick={addEnv}>
              <Plus className='mr-1 h-4 w-4' />
              Add
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Tags</CardTitle>
          <CardDescription>
            Organize and search for this MCP server.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-3'>
          <div className='flex flex-wrap gap-2'>
            {(formData.tags || []).map(tag => (
              <Badge key={tag} variant='secondary' className='gap-1'>
                {tag}
                <button type='button' onClick={() => removeTag(tag)} className='ml-1 hover:text-destructive'>
                  <X className='h-3 w-3' />
                </button>
              </Badge>
            ))}
          </div>
          <div className='flex gap-2'>
            <Input
              value={newTag}
              onChange={e => setNewTag(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addTag() } }}
              placeholder='Add tag...'
              className='flex-1'
            />
            <Button type='button' variant='secondary' size='sm' onClick={addTag}>
              <Plus className='mr-1 h-4 w-4' />
              Add
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card className='border-blue-500/20 bg-blue-500/5'>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <Box className='h-5 w-5 text-blue-500' />
            Common Examples
          </CardTitle>
          <CardDescription>
            Reference configurations for common MCP servers.
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-3 text-sm'>
          <div className='rounded-lg border p-3'>
            <p className='font-medium mb-1'>Filesystem Server</p>
            <code className='text-xs text-emerald-600 dark:text-emerald-400'>
              npx -y @modelcontextprotocol/server-filesystem /workspace
            </code>
          </div>
          <div className='rounded-lg border p-3'>
            <p className='font-medium mb-1'>GitHub Server</p>
            <code className='text-xs text-emerald-600 dark:text-emerald-400'>
              npx -y @modelcontextprotocol/server-github
            </code>
            <p className='text-xs text-muted-foreground mt-1'>Requires: GITHUB_PERSONAL_ACCESS_TOKEN</p>
          </div>
          <div className='rounded-lg border p-3'>
            <p className='font-medium mb-1'>Web Browser Server</p>
            <code className='text-xs text-emerald-600 dark:text-emerald-400'>
              npx -y @modelcontextprotocol/server-puppeteer
            </code>
          </div>
        </CardContent>
      </Card>
    </form>
  )
}
