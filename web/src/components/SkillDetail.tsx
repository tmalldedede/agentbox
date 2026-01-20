import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  Save,
  Trash2,
  Loader2,
  AlertCircle,
  Zap,
  Lock,
  Code,
  FileSearch,
  FileText,
  Shield,
  TestTube,
  Box,
  Terminal,
  Plus,
  X,
} from 'lucide-react'
import type { Skill, SkillCategory, SkillFile } from '../types'
import { useSkills, useUpdateSkill, useDeleteSkill } from '../hooks'
import { toast } from 'sonner'

const categoryIcons: Record<SkillCategory, React.ReactNode> = {
  coding: <Code className="w-4 h-4" />,
  review: <FileSearch className="w-4 h-4" />,
  docs: <FileText className="w-4 h-4" />,
  security: <Shield className="w-4 h-4" />,
  testing: <TestTube className="w-4 h-4" />,
  other: <Box className="w-4 h-4" />,
}

const categories: SkillCategory[] = ['coding', 'review', 'docs', 'security', 'testing', 'other']

export default function SkillDetail() {
  const navigate = useNavigate()
  const { skillId } = useParams<{ skillId: string }>()
  const { data: skills = [], isLoading } = useSkills()
  const updateSkill = useUpdateSkill()
  const deleteSkill = useDeleteSkill()

  const skill = skills.find(s => s.id === skillId)

  const [formData, setFormData] = useState<Partial<Skill>>({})
  const [isDirty, setIsDirty] = useState(false)
  const [newTag, setNewTag] = useState('')
  const [newMcpId, setNewMcpId] = useState('')
  const [newFile, setNewFile] = useState<{ path: string; content: string }>({ path: '', content: '' })
  const [showAddFile, setShowAddFile] = useState(false)

  useEffect(() => {
    if (skill) {
      setFormData(skill)
    }
  }, [skill])

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  if (!skill) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <AlertCircle className="w-16 h-16 text-red-400 mx-auto mb-4" />
          <p className="text-red-400 text-lg">Skill not found</p>
          <button onClick={() => navigate('/skills')} className="btn btn-primary mt-4">
            Back to Skills
          </button>
        </div>
      </div>
    )
  }

  const isBuiltIn = skill.is_built_in
  const isReadOnly = isBuiltIn

  const handleSave = async () => {
    if (!skillId) return

    try {
      await updateSkill.mutateAsync({
        id: skillId,
        data: {
          name: formData.name,
          description: formData.description,
          command: formData.command,
          prompt: formData.prompt,
          files: formData.files,
          allowed_tools: formData.allowed_tools,
          required_mcp: formData.required_mcp,
          category: formData.category,
          tags: formData.tags,
          is_enabled: formData.is_enabled,
        },
      })
      setIsDirty(false)
      toast.success('Skill updated successfully')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update skill')
    }
  }

  const handleDelete = async () => {
    if (!skillId) return
    if (!confirm(`Delete skill "${skill.name}"?`)) return

    try {
      await deleteSkill.mutateAsync(skillId)
      toast.success('Skill deleted')
      navigate('/skills')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete skill')
    }
  }

  const updateField = <K extends keyof Skill>(key: K, value: Skill[K]) => {
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

  const addMcp = () => {
    if (!newMcpId.trim()) return
    const mcps = formData.required_mcp || []
    if (mcps.includes(newMcpId.trim())) {
      toast.error('MCP already added')
      return
    }
    updateField('required_mcp', [...mcps, newMcpId.trim()])
    setNewMcpId('')
  }

  const removeMcp = (mcpId: string) => {
    updateField('required_mcp', (formData.required_mcp || []).filter(m => m !== mcpId))
  }

  const addFile = () => {
    if (!newFile.path.trim() || !newFile.content.trim()) {
      toast.error('Both path and content are required')
      return
    }
    const files = formData.files || []
    if (files.some(f => f.path === newFile.path.trim())) {
      toast.error('File path already exists')
      return
    }
    updateField('files', [...files, { path: newFile.path.trim(), content: newFile.content.trim() }])
    setNewFile({ path: '', content: '' })
    setShowAddFile(false)
  }

  const removeFile = (path: string) => {
    updateField('files', (formData.files || []).filter(f => f.path !== path))
  }

  return (
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/skills')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Zap className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Skill Detail</span>
          </div>
          {isBuiltIn && (
            <span className="badge badge-scaling">
              <Lock className="w-3 h-3" />
              Built-in
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          {!isBuiltIn && (
            <button
              onClick={handleDelete}
              className="btn btn-ghost text-red-400"
              disabled={deleteSkill.isPending}
            >
              <Trash2 className="w-4 h-4" />
              Delete
            </button>
          )}
          {!isReadOnly && (
            <button
              onClick={handleSave}
              className="btn btn-primary"
              disabled={!isDirty || updateSkill.isPending}
            >
              {updateSkill.isPending ? (
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
              <label className="block text-sm font-medium text-secondary mb-2">Skill ID</label>
              <input type="text" value={skill.id} disabled className="input bg-secondary" />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Name *</label>
              <input
                type="text"
                value={formData.name || ''}
                onChange={e => updateField('name', e.target.value)}
                disabled={isReadOnly}
                className="input"
                placeholder="My Custom Skill"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Description</label>
              <textarea
                value={formData.description || ''}
                onChange={e => updateField('description', e.target.value)}
                disabled={isReadOnly}
                className="input min-h-[80px]"
                placeholder="A brief description of what this skill does"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2 flex items-center gap-2">
                <Terminal className="w-4 h-4" />
                Command *
              </label>
              <input
                type="text"
                value={formData.command || ''}
                onChange={e => updateField('command', e.target.value)}
                disabled={isReadOnly}
                className="input font-mono"
                placeholder="/my-skill"
              />
              <p className="text-xs text-muted mt-1">Slash command to invoke this skill</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Category *</label>
              <select
                value={formData.category || 'other'}
                onChange={e => updateField('category', e.target.value as SkillCategory)}
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

        {/* Prompt */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Prompt *</h3>
          <p className="text-sm text-muted mb-4">
            The instruction template that will be sent to the agent when this skill is invoked.
          </p>
          <textarea
            value={formData.prompt || ''}
            onChange={e => updateField('prompt', e.target.value)}
            disabled={isReadOnly}
            className="input font-mono text-sm min-h-[300px]"
            placeholder="You are a helpful assistant that..."
          />
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

        {/* Required MCP Servers */}
        <div className="card p-6">
          <h3 className="font-semibold text-lg mb-4">Required MCP Servers</h3>
          <p className="text-sm text-muted mb-4">
            List of MCP server IDs that must be available for this skill to work.
          </p>
          <div className="space-y-3">
            <div className="flex flex-wrap gap-2">
              {(formData.required_mcp || []).map(mcpId => (
                <span
                  key={mcpId}
                  className="inline-flex items-center gap-2 px-3 py-1 rounded-lg bg-amber-500/20 text-amber-400"
                >
                  <code className="text-xs">{mcpId}</code>
                  {!isReadOnly && (
                    <button onClick={() => removeMcp(mcpId)} className="text-red-400 hover:text-red-300">
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
                  value={newMcpId}
                  onChange={e => setNewMcpId(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && addMcp()}
                  className="input flex-1 font-mono text-sm"
                  placeholder="mcp-server-id"
                />
                <button onClick={addMcp} className="btn btn-secondary">
                  <Plus className="w-4 h-4" />
                  Add
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Files */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="font-semibold text-lg">Files</h3>
              <p className="text-sm text-muted mt-1">
                Additional files that will be created in the agent's workspace when using this skill.
              </p>
            </div>
            {!isReadOnly && (
              <button onClick={() => setShowAddFile(true)} className="btn btn-secondary">
                <Plus className="w-4 h-4" />
                Add File
              </button>
            )}
          </div>

          <div className="space-y-3">
            {(formData.files || []).map(file => (
              <div key={file.path} className="p-4 rounded-lg bg-secondary border border-default">
                <div className="flex items-center justify-between mb-2">
                  <code className="text-sm text-emerald-400">{file.path}</code>
                  {!isReadOnly && (
                    <button
                      onClick={() => removeFile(file.path)}
                      className="btn btn-ghost btn-icon text-red-400"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  )}
                </div>
                <pre className="text-xs text-muted overflow-x-auto bg-background p-2 rounded">
                  {file.content}
                </pre>
              </div>
            ))}
          </div>

          {showAddFile && !isReadOnly && (
            <div className="mt-4 p-4 rounded-lg bg-secondary border border-emerald-500/30">
              <h4 className="font-medium mb-3">Add New File</h4>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-secondary mb-1">File Path</label>
                  <input
                    type="text"
                    value={newFile.path}
                    onChange={e => setNewFile(prev => ({ ...prev, path: e.target.value }))}
                    className="input font-mono text-sm"
                    placeholder=".github/workflows/ci.yml"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-secondary mb-1">Content</label>
                  <textarea
                    value={newFile.content}
                    onChange={e => setNewFile(prev => ({ ...prev, content: e.target.value }))}
                    className="input font-mono text-sm min-h-[150px]"
                    placeholder="File content..."
                  />
                </div>
                <div className="flex gap-2">
                  <button onClick={addFile} className="btn btn-primary">
                    Add File
                  </button>
                  <button
                    onClick={() => {
                      setShowAddFile(false)
                      setNewFile({ path: '', content: '' })
                    }}
                    className="btn btn-ghost"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          )}
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
              <p className="text-primary mt-1">{new Date(skill.created_at).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-muted">Updated</p>
              <p className="text-primary mt-1">{new Date(skill.updated_at).toLocaleString()}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
