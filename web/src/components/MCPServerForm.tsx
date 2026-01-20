import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Save,
  Server,
  Plus,
  X,
  Loader2,
  Database,
  Globe,
  Wrench,
  Monitor,
  Brain,
  Box,
} from 'lucide-react'
import type { MCPServerType, MCPCategory, CreateMCPServerRequest } from '../types'
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
    author: '',
    version: '1.0.0',
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

    // Validation
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

    // ID validation: lowercase, alphanumeric with hyphens
    if (!/^[a-z0-9-]+$/.test(formData.id)) {
      toast.error('Server ID must be lowercase alphanumeric with hyphens only')
      return
    }

    setSaving(true)
    try {
      await api.createMCPServer(formData)
      toast.success('MCP Server created successfully')
      navigate('/mcp-servers')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create MCP server')
    } finally {
      setSaving(false)
    }
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
            <span className="text-lg font-bold">Create New MCP Server</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button onClick={() => navigate('/mcp-servers')} className="btn btn-ghost">
            Cancel
          </button>
          <button onClick={handleSubmit} className="btn btn-primary" disabled={saving}>
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Create Server
          </button>
        </div>
      </header>

      <form onSubmit={handleSubmit} className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Basic Info */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Basic Information</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">
                Server ID * <span className="text-muted font-normal">(lowercase, alphanumeric, hyphens)</span>
              </label>
              <input
                type="text"
                value={formData.id}
                onChange={e => updateField('id', e.target.value.toLowerCase())}
                className="input font-mono"
                placeholder="my-mcp-server"
                required
              />
              <p className="text-xs text-muted mt-1">
                Unique identifier for this MCP server. Cannot be changed later.
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={e => updateField('name', e.target.value)}
                className="input"
                placeholder="My MCP Server"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={e => updateField('description', e.target.value)}
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
                  className="input"
                  required
                >
                  {serverTypes.map(type => (
                    <option key={type} value={type}>
                      {type}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted mt-1">
                  stdio: Standard input/output • sse: Server-Sent Events • http: HTTP API
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-secondary mb-2">Category *</label>
                <select
                  value={formData.category || 'other'}
                  onChange={e => updateField('category', e.target.value as MCPCategory)}
                  className="input"
                  required
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
          <p className="text-sm text-muted mb-4">
            Configure how to run this MCP server. The command will be executed in the agent's container.
          </p>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Command *</label>
              <input
                type="text"
                value={formData.command}
                onChange={e => updateField('command', e.target.value)}
                className="input font-mono"
                placeholder="npx -y @modelcontextprotocol/server-filesystem"
                required
              />
              <p className="text-xs text-muted mt-1">
                Example: <code className="text-emerald-400">npx -y @modelcontextprotocol/server-filesystem</code>
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Arguments</label>
              <p className="text-xs text-muted mb-2">
                Add command-line arguments. Each argument will be passed to the command separately.
              </p>
              <div className="space-y-2">
                {(formData.args || []).map((arg, index) => (
                  <div key={index} className="flex gap-2">
                    <input
                      type="text"
                      value={arg}
                      disabled
                      className="input flex-1 font-mono text-sm bg-secondary"
                    />
                    <button
                      type="button"
                      onClick={() => removeArg(index)}
                      className="btn btn-ghost btn-icon text-red-400"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newArg}
                    onChange={e => setNewArg(e.target.value)}
                    onKeyDown={e => {
                      if (e.key === 'Enter') {
                        e.preventDefault()
                        addArg()
                      }
                    }}
                    className="input flex-1 font-mono text-sm"
                    placeholder="Add an argument (e.g., /workspace)"
                  />
                  <button type="button" onClick={addArg} className="btn btn-secondary">
                    <Plus className="w-4 h-4" />
                    Add
                  </button>
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Working Directory</label>
              <input
                type="text"
                value={formData.work_dir || ''}
                onChange={e => updateField('work_dir', e.target.value)}
                className="input font-mono"
                placeholder="/path/to/workdir"
              />
              <p className="text-xs text-muted mt-1">Optional: Directory to run the command from</p>
            </div>
          </div>
        </div>

        {/* Environment Variables */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Environment Variables</h3>
          <p className="text-sm text-muted mb-4">
            Set environment variables that will be available to the MCP server process.
          </p>
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
                <button
                  type="button"
                  onClick={() => removeEnv(key)}
                  className="btn btn-ghost btn-icon text-red-400"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
            <div className="flex gap-2">
              <input
                type="text"
                value={newEnvKey}
                onChange={e => setNewEnvKey(e.target.value)}
                className="input flex-1 font-mono text-sm"
                placeholder="VARIABLE_NAME"
              />
              <input
                type="text"
                value={newEnvValue}
                onChange={e => setNewEnvValue(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addEnv()
                  }
                }}
                className="input flex-1 font-mono text-sm"
                placeholder="value"
              />
              <button type="button" onClick={addEnv} className="btn btn-secondary">
                <Plus className="w-4 h-4" />
                Add
              </button>
            </div>
          </div>
        </div>

        {/* Tags */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Tags</h3>
          <p className="text-sm text-muted mb-4">Add tags to help organize and search for this MCP server.</p>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              {(formData.tags || []).map(tag => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-2 px-3 py-1 rounded-lg bg-secondary text-secondary"
                >
                  {tag}
                  <button
                    type="button"
                    onClick={() => removeTag(tag)}
                    className="text-red-400 hover:text-red-300"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={newTag}
                onChange={e => setNewTag(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addTag()
                  }
                }}
                className="input flex-1"
                placeholder="Add a tag..."
              />
              <button type="button" onClick={addTag} className="btn btn-secondary">
                <Plus className="w-4 h-4" />
                Add
              </button>
            </div>
          </div>
        </div>

        {/* Metadata */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Metadata (Optional)</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Author</label>
              <input
                type="text"
                value={formData.author || ''}
                onChange={e => updateField('author', e.target.value)}
                className="input"
                placeholder="Your name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Version</label>
              <input
                type="text"
                value={formData.version || ''}
                onChange={e => updateField('version', e.target.value)}
                className="input"
                placeholder="1.0.0"
              />
            </div>
          </div>
        </div>

        {/* Example Templates */}
        <div className="card p-6 bg-blue-500/5 border-blue-500/20">
          <h3 className="font-semibold text-lg mb-3 flex items-center gap-2">
            <Box className="w-5 h-5 text-blue-400" />
            Common Examples
          </h3>
          <p className="text-sm text-muted mb-4">
            Here are some common MCP server configurations you can use as reference:
          </p>
          <div className="space-y-3 text-sm">
            <div className="p-3 rounded-lg bg-secondary">
              <p className="font-medium text-primary mb-1">Filesystem Server</p>
              <code className="text-xs text-emerald-400">
                npx -y @modelcontextprotocol/server-filesystem /workspace
              </code>
            </div>
            <div className="p-3 rounded-lg bg-secondary">
              <p className="font-medium text-primary mb-1">GitHub Server</p>
              <code className="text-xs text-emerald-400">
                npx -y @modelcontextprotocol/server-github
              </code>
              <p className="text-xs text-muted mt-1">Requires: GITHUB_PERSONAL_ACCESS_TOKEN env var</p>
            </div>
            <div className="p-3 rounded-lg bg-secondary">
              <p className="font-medium text-primary mb-1">Web Browser Server</p>
              <code className="text-xs text-emerald-400">
                npx -y @modelcontextprotocol/server-puppeteer
              </code>
            </div>
          </div>
        </div>
      </form>
    </div>
  )
}
