import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  Save,
  Trash2,
  Loader2,
  AlertCircle,
  Server,
  Lock,
  Database,
  Globe,
  Wrench,
  Monitor,
  Brain,
  Box,
  Plus,
  X,
  Play,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import type { MCPServer, MCPServerType, MCPCategory } from '../types'
import { useMCPServers, useUpdateMCPServer, useDeleteMCPServer } from '../hooks'
import { api } from '../services/api'
import { toast } from 'sonner'

const categoryIcons: Record<MCPCategory, React.ReactNode> = {
  filesystem: <Server className="w-4 h-4" />,
  database: <Database className="w-4 h-4" />,
  api: <Globe className="w-4 h-4" />,
  tool: <Wrench className="w-4 h-4" />,
  browser: <Monitor className="w-4 h-4" />,
  memory: <Brain className="w-4 h-4" />,
  other: <Box className="w-4 h-4" />,
}

const categories: MCPCategory[] = ['filesystem', 'database', 'api', 'tool', 'browser', 'memory', 'other']
const serverTypes: MCPServerType[] = ['stdio', 'sse', 'http']

export default function MCPServerDetail() {
  const navigate = useNavigate()
  const { serverId } = useParams<{ serverId: string }>()
  const { data: servers = [], isLoading } = useMCPServers()
  const updateServer = useUpdateMCPServer()
  const deleteServer = useDeleteMCPServer()

  const server = servers.find(s => s.id === serverId)

  const [formData, setFormData] = useState<Partial<MCPServer>>({})
  const [isDirty, setIsDirty] = useState(false)
  const [newTag, setNewTag] = useState('')
  const [newEnvKey, setNewEnvKey] = useState('')
  const [newEnvValue, setNewEnvValue] = useState('')
  const [newArg, setNewArg] = useState('')
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<'ok' | 'error' | null>(null)

  useEffect(() => {
    if (server) {
      setFormData(server)
    }
  }, [server])

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  if (!server) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <AlertCircle className="w-16 h-16 text-red-400 mx-auto mb-4" />
          <p className="text-red-400 text-lg">MCP Server not found</p>
          <button onClick={() => navigate('/mcp-servers')} className="btn btn-primary mt-4">
            Back to MCP Servers
          </button>
        </div>
      </div>
    )
  }

  const isBuiltIn = server.is_built_in
  const isReadOnly = isBuiltIn

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
    if (!confirm(`Delete MCP server "${server.name}"?`)) return

    try {
      await deleteServer.mutateAsync(serverId)
      toast.success('MCP Server deleted')
      navigate('/mcp-servers')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete MCP server')
    }
  }

  const handleTest = async () => {
    if (!serverId) return
    setTesting(true)
    setTestResult(null)
    try {
      await api.testMCPServer(serverId)
      setTestResult('ok')
      toast.success('Connection test successful')
    } catch (err) {
      setTestResult('error')
      toast.error(err instanceof Error ? err.message : 'Connection test failed')
    } finally {
      setTesting(false)
      setTimeout(() => setTestResult(null), 3000)
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
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/mcp-servers')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Server className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">MCP Server Detail</span>
          </div>
          {isBuiltIn && (
            <span className="badge badge-scaling">
              <Lock className="w-3 h-3" />
              Built-in
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          <button onClick={handleTest} className="btn btn-ghost" disabled={testing}>
            {testing ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : testResult === 'ok' ? (
              <CheckCircle className="w-4 h-4 text-emerald-400" />
            ) : testResult === 'error' ? (
              <XCircle className="w-4 h-4 text-red-400" />
            ) : (
              <Play className="w-4 h-4" />
            )}
            Test
          </button>
          {!isBuiltIn && (
            <button
              onClick={handleDelete}
              className="btn btn-ghost text-red-400"
              disabled={deleteServer.isPending}
            >
              <Trash2 className="w-4 h-4" />
              Delete
            </button>
          )}
          {!isReadOnly && (
            <button
              onClick={handleSave}
              className="btn btn-primary"
              disabled={!isDirty || updateServer.isPending}
            >
              {updateServer.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              Save
            </button>
          )}
        </div>
      </header>

      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Basic Info */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Basic Information</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Server ID</label>
              <input type="text" value={server.id} disabled className="input bg-secondary" />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Name *</label>
              <input
                type="text"
                value={formData.name || ''}
                onChange={e => updateField('name', e.target.value)}
                disabled={isReadOnly}
                className="input"
                placeholder="My MCP Server"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={e => updateField('description', e.target.value)}
                disabled={isReadOnly}
                className="input min-h-[80px]"
                placeholder="A brief description of what this MCP server does"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-secondary mb-2">Type *</label>
                <select
                  value={formData.type || 'stdio'}
                  onChange={e => updateField('type', e.target.value as MCPServerType)}
                  disabled={isReadOnly}
                  className="input"
                >
                  {serverTypes.map(type => (
                    <option key={type} value={type}>
                      {type}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-secondary mb-2">Category *</label>
                <select
                  value={formData.category || 'other'}
                  onChange={e => updateField('category', e.target.value as MCPCategory)}
                  disabled={isReadOnly}
                  className="input"
                >
                  {categories.map(cat => (
                    <option key={cat} value={cat}>
                      {cat}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Command & Args */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Command & Arguments</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Command *</label>
              <input
                type="text"
                value={formData.command || ''}
                onChange={e => updateField('command', e.target.value)}
                disabled={isReadOnly}
                className="input font-mono"
                placeholder="npx -y @modelcontextprotocol/server-filesystem"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Arguments</label>
              <div className="space-y-2">
                {(formData.args || []).map((arg, index) => (
                  <div key={index} className="flex gap-2">
                    <input
                      type="text"
                      value={arg}
                      disabled
                      className="input flex-1 font-mono text-sm bg-secondary"
                    />
                    {!isReadOnly && (
                      <button
                        onClick={() => removeArg(index)}
                        className="btn btn-ghost btn-icon text-red-400"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                ))}
                {!isReadOnly && (
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={newArg}
                      onChange={e => setNewArg(e.target.value)}
                      onKeyDown={e => e.key === 'Enter' && addArg()}
                      className="input flex-1 font-mono text-sm"
                      placeholder="Add an argument..."
                    />
                    <button onClick={addArg} className="btn btn-secondary">
                      <Plus className="w-4 h-4" />
                      Add
                    </button>
                  </div>
                )}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Working Directory</label>
              <input
                type="text"
                value={formData.work_dir || ''}
                onChange={e => updateField('work_dir', e.target.value)}
                disabled={isReadOnly}
                className="input font-mono"
                placeholder="/path/to/workdir"
              />
            </div>
          </div>
        </div>

        {/* Environment Variables */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Environment Variables</h3>
          <div className="space-y-3">
            {Object.entries(formData.env || {}).map(([key, value]) => (
              <div key={key} className="flex gap-2 items-start">
                <input
                  type="text"
                  value={key}
                  disabled
                  className="input flex-1 font-mono text-sm bg-secondary"
                  placeholder="KEY"
                />
                <input
                  type="text"
                  value={value}
                  disabled
                  className="input flex-1 font-mono text-sm bg-secondary"
                  placeholder="value"
                />
                {!isReadOnly && (
                  <button onClick={() => removeEnv(key)} className="btn btn-ghost btn-icon text-red-400">
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>
            ))}
            {!isReadOnly && (
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newEnvKey}
                  onChange={e => setNewEnvKey(e.target.value)}
                  className="input flex-1 font-mono text-sm"
                  placeholder="KEY"
                />
                <input
                  type="text"
                  value={newEnvValue}
                  onChange={e => setNewEnvValue(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && addEnv()}
                  className="input flex-1 font-mono text-sm"
                  placeholder="value"
                />
                <button onClick={addEnv} className="btn btn-secondary">
                  <Plus className="w-4 h-4" />
                  Add
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Tags */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Tags</h3>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              {(formData.tags || []).map(tag => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-2 px-3 py-1 rounded-lg bg-secondary text-secondary"
                >
                  {tag}
                  {!isReadOnly && (
                    <button onClick={() => removeTag(tag)} className="text-red-400 hover:text-red-300">
                      <X className="w-3 h-3" />
                    </button>
                  )}
                </span>
              ))}
            </div>
            {!isReadOnly && (
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newTag}
                  onChange={e => setNewTag(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && addTag()}
                  className="input flex-1"
                  placeholder="Add a tag..."
                />
                <button onClick={addTag} className="btn btn-secondary">
                  <Plus className="w-4 h-4" />
                  Add
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Metadata */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Metadata</h3>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted">Author</p>
              <p className="text-primary mt-1">{formData.author || 'N/A'}</p>
            </div>
            <div>
              <p className="text-muted">Version</p>
              <p className="text-primary mt-1">{formData.version || 'N/A'}</p>
            </div>
            <div>
              <p className="text-muted">Created</p>
              <p className="text-primary mt-1">{new Date(server.created_at).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-muted">Updated</p>
              <p className="text-primary mt-1">{new Date(server.updated_at).toLocaleString()}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
