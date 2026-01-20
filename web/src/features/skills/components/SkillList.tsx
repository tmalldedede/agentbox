import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Copy,
  Trash2,
  ChevronRight,
  Zap,
  Code,
  FileSearch,
  FileText,
  Shield,
  TestTube,
  Box,
  RefreshCw,
  AlertCircle,
  Loader2,
  Lock,
  Power,
  PowerOff,
  Download,
  Terminal,
  Store,
  Github,
  Star,
  Check,
  ExternalLink,
  X,
} from 'lucide-react'
import type { Skill, SkillCategory, RemoteSkill, SkillSource, AddSourceRequest } from '@/types'
import {
  useSkills,
  useUpdateSkill,
  useDeleteSkill,
  useCloneSkill,
  useExportSkill,
  useRemoteSkills,
  useSkillSources,
  useInstallSkill,
  useRefreshSkillSource,
  useAddSkillSource,
  useRemoveSkillSource,
} from '@/hooks'

// 类别图标映射
const categoryIcons: Record<SkillCategory, React.ReactNode> = {
  coding: <Code className="w-4 h-4" />,
  review: <FileSearch className="w-4 h-4" />,
  docs: <FileText className="w-4 h-4" />,
  security: <Shield className="w-4 h-4" />,
  testing: <TestTube className="w-4 h-4" />,
  other: <Box className="w-4 h-4" />,
}

// 类别颜色映射
const categoryColors: Record<SkillCategory, string> = {
  coding: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  review: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  docs: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  security: 'bg-red-500/20 text-red-400 border-red-500/30',
  testing: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  other: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
}

// 源类型颜色映射
const sourceTypeColors: Record<string, string> = {
  official: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  community: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  custom: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
}

// Tab 类型
type TabType = 'installed' | 'store' | 'sources'

// Skill 卡片组件
function SkillCard({
  skill,
  onClone,
  onDelete,
  onToggle,
  onExport,
  onClick,
}: {
  skill: Skill
  onClone: () => void
  onDelete: () => void
  onToggle: () => void
  onExport: () => void
  onClick: () => void
}) {
  const colors = categoryColors[skill.category] || categoryColors.other
  const icon = categoryIcons[skill.category] || categoryIcons.other

  return (
    <div
      className={`card p-4 cursor-pointer group transition-colors ${
        skill.is_enabled
          ? 'hover:border-emerald-500/50'
          : 'opacity-60 hover:border-gray-500/50'
      }`}
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colors}`}>
          {icon}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{skill.name}</span>
            {skill.is_built_in && (
              <span className="badge badge-scaling text-xs">
                <Lock className="w-3 h-3" />
                Built-in
              </span>
            )}
            {!skill.is_enabled && (
              <span className="text-xs px-2 py-0.5 rounded bg-gray-500/20 text-gray-400">
                Disabled
              </span>
            )}
          </div>

          {/* Command */}
          <div className="flex items-center gap-2 mt-1">
            <Terminal className="w-3 h-3 text-muted-foreground" />
            <code className="text-sm text-emerald-400 font-mono">{skill.command}</code>
          </div>

          <p className="text-sm text-foreground/90-foreground mt-1 line-clamp-2">
            {skill.description || skill.prompt.slice(0, 100)}
          </p>

          {/* Tags */}
          <div className="flex items-center gap-2 mt-3 flex-wrap">
            <span className={`text-xs px-2 py-0.5 rounded border ${colors}`}>
              {skill.category}
            </span>
            {skill.tags?.slice(0, 3).map(tag => (
              <span
                key={tag}
                className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/90"
              >
                {tag}
              </span>
            ))}
            {skill.required_mcp && skill.required_mcp.length > 0 && (
              <span className="text-xs px-2 py-0.5 rounded bg-amber-500/20 text-amber-400">
                {skill.required_mcp.length} MCP
              </span>
            )}
          </div>
        </div>

        {/* Actions */}
        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          {/* Export Button */}
          <button
            onClick={e => {
              e.stopPropagation()
              onExport()
            }}
            className="btn btn-ghost btn-icon"
            title="Export as SKILL.md"
          >
            <Download className="w-4 h-4" />
          </button>

          {/* Toggle Button */}
          <button
            onClick={e => {
              e.stopPropagation()
              onToggle()
            }}
            className="btn btn-ghost btn-icon"
            title={skill.is_enabled ? 'Disable' : 'Enable'}
          >
            {skill.is_enabled ? (
              <Power className="w-4 h-4 text-emerald-400" />
            ) : (
              <PowerOff className="w-4 h-4 text-gray-400" />
            )}
          </button>

          {/* Clone Button */}
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

          {/* Delete Button */}
          {!skill.is_built_in && (
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

        {/* Arrow */}
        <ChevronRight className="w-5 h-5 text-muted-foreground group-hover:text-emerald-400 transition-colors" />
      </div>
    </div>
  )
}

// 远程 Skill 卡片组件
function RemoteSkillCard({
  skill,
  onInstall,
  installing,
}: {
  skill: RemoteSkill
  onInstall: () => void
  installing: boolean
}) {
  const colors = categoryColors[skill.category as SkillCategory] || categoryColors.other
  const icon = categoryIcons[skill.category as SkillCategory] || categoryIcons.other

  return (
    <div className="card p-4 group transition-colors hover:border-blue-500/50">
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colors}`}>
          {icon}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{skill.name}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/90">
              {skill.source_name}
            </span>
            {skill.is_installed && (
              <span className="text-xs px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 flex items-center gap-1">
                <Check className="w-3 h-3" />
                已安装
              </span>
            )}
          </div>

          {/* Command */}
          <div className="flex items-center gap-2 mt-1">
            <Terminal className="w-3 h-3 text-muted-foreground" />
            <code className="text-sm text-blue-400 font-mono">{skill.command}</code>
          </div>

          <p className="text-sm text-foreground/90-foreground mt-1 line-clamp-2">
            {skill.description || 'No description'}
          </p>

          {/* Meta */}
          <div className="flex items-center gap-3 mt-3 text-xs text-muted-foreground">
            {skill.author && <span>by {skill.author}</span>}
            {skill.version && <span>v{skill.version}</span>}
            <span className={`px-2 py-0.5 rounded border ${colors}`}>
              {skill.category}
            </span>
          </div>
        </div>

        {/* Install Button */}
        <div className="flex items-center">
          {skill.is_installed ? (
            <button
              className="btn btn-ghost text-emerald-400"
              disabled
            >
              <Check className="w-4 h-4" />
              已安装
            </button>
          ) : (
            <button
              onClick={onInstall}
              disabled={installing}
              className="btn btn-primary"
            >
              {installing ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Download className="w-4 h-4" />
              )}
              安装
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

// 源卡片组件
function SourceCard({
  source,
  onRefresh,
  onDelete,
  refreshing,
  deleting,
}: {
  source: SkillSource
  onRefresh: () => void
  onDelete: () => void
  refreshing: boolean
  deleting: boolean
}) {
  const colors = sourceTypeColors[source.type] || sourceTypeColors.custom
  const isOfficial = source.type === 'official'

  return (
    <div className="card p-4 group transition-colors hover:border-emerald-500/50">
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div className="w-12 h-12 rounded-xl bg-gray-700 flex items-center justify-center">
          <Github className="w-6 h-6 text-white" />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{source.name}</span>
            <span className={`text-xs px-2 py-0.5 rounded border ${colors}`}>
              {source.type}
            </span>
            {source.stars !== undefined && source.stars > 0 && (
              <span className="flex items-center gap-1 text-xs text-amber-400">
                <Star className="w-3 h-3" />
                {source.stars}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2 mt-1 text-sm text-foreground/90">
            <span>{source.owner}/{source.repo}</span>
            <span className="text-muted-foreground">·</span>
            <span>{source.branch}</span>
            <span className="text-muted-foreground">·</span>
            <span>{source.path}</span>
          </div>

          {source.description && (
            <p className="text-sm text-foreground/90-foreground mt-1 line-clamp-2">
              {source.description}
            </p>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          <a
            href={`https://github.com/${source.owner}/${source.repo}`}
            target="_blank"
            rel="noopener noreferrer"
            className="btn btn-ghost btn-icon"
            title="在 GitHub 上查看"
          >
            <ExternalLink className="w-4 h-4" />
          </a>
          <button
            onClick={onRefresh}
            disabled={refreshing}
            className="btn btn-ghost btn-icon"
            title="刷新"
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
          </button>
          {!isOfficial && (
            <button
              onClick={onDelete}
              disabled={deleting}
              className="btn btn-ghost btn-icon text-red-400"
              title="删除"
            >
              {deleting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Trash2 className="w-4 h-4" />
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

// 添加源对话框
function AddSourceModal({
  isOpen,
  onClose,
  onSubmit,
  submitting,
}: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: AddSourceRequest) => void
  submitting: boolean
}) {
  const [formData, setFormData] = useState<AddSourceRequest>({
    id: '',
    name: '',
    owner: '',
    repo: '',
    branch: 'main',
    path: 'skills',
    type: 'community',
    description: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  const handleGitHubUrlParse = (url: string) => {
    // 解析 GitHub URL: https://github.com/owner/repo
    const match = url.match(/github\.com\/([^/]+)\/([^/]+)/)
    if (match) {
      const owner = match[1]
      const repo = match[2].replace(/\.git$/, '')
      setFormData(prev => ({
        ...prev,
        owner,
        repo,
        id: `${owner}-${repo}`,
        name: repo,
      }))
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative bg-primary border border-border rounded-xl shadow-xl w-full max-w-lg mx-4">
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">添加 Skill 源</h2>
          <button onClick={onClose} className="btn btn-ghost btn-icon">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {/* GitHub URL 快速填充 */}
          <div>
            <label className="block text-sm font-medium text-foreground/90 mb-1">
              GitHub URL（快速填充）
            </label>
            <input
              type="text"
              placeholder="https://github.com/owner/repo"
              className="input w-full"
              onChange={e => handleGitHubUrlParse(e.target.value)}
            />
            <p className="text-xs text-muted-foreground mt-1">粘贴 GitHub 仓库地址自动解析</p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                Owner <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={formData.owner}
                onChange={e => setFormData(prev => ({ ...prev, owner: e.target.value }))}
                placeholder="anthropics"
                className="input w-full"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                Repo <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={formData.repo}
                onChange={e => setFormData(prev => ({ ...prev, repo: e.target.value }))}
                placeholder="skills"
                className="input w-full"
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                Branch
              </label>
              <input
                type="text"
                value={formData.branch}
                onChange={e => setFormData(prev => ({ ...prev, branch: e.target.value }))}
                placeholder="main"
                className="input w-full"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                Skills 路径
              </label>
              <input
                type="text"
                value={formData.path}
                onChange={e => setFormData(prev => ({ ...prev, path: e.target.value }))}
                placeholder="skills"
                className="input w-full"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                ID <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={formData.id}
                onChange={e => setFormData(prev => ({ ...prev, id: e.target.value }))}
                placeholder="my-skills"
                className="input w-full"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-1">
                名称 <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={e => setFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="My Skills"
                className="input w-full"
                required
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground/90 mb-1">
              描述
            </label>
            <input
              type="text"
              value={formData.description}
              onChange={e => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Skills 仓库描述"
              className="input w-full"
            />
          </div>

          <div className="flex justify-end gap-2 pt-4 border-t border-border">
            <button type="button" onClick={onClose} className="btn btn-ghost">
              取消
            </button>
            <button
              type="submit"
              disabled={submitting || !formData.id || !formData.owner || !formData.repo}
              className="btn btn-primary"
            >
              {submitting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Plus className="w-4 h-4" />
              )}
              添加源
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function SkillList() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<TabType>('installed')
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [installingId, setInstallingId] = useState<string | null>(null)
  const [refreshingId, setRefreshingId] = useState<string | null>(null)
  const [deletingSourceId, setDeletingSourceId] = useState<string | null>(null)
  const [showAddSourceModal, setShowAddSourceModal] = useState(false)

  // React Query hooks - 本地 Skills
  const { data: skills = [], isLoading, isFetching, error, refetch } = useSkills()
  const updateSkill = useUpdateSkill()
  const deleteSkill = useDeleteSkill()
  const cloneSkill = useCloneSkill()
  const exportSkill = useExportSkill()

  // React Query hooks - 远程 Skills
  const { data: remoteSkills = [], isLoading: loadingRemote, isFetching: fetchingRemote } = useRemoteSkills()
  const { data: sources = [], isLoading: loadingSources } = useSkillSources()
  const installSkill = useInstallSkill()
  const refreshSource = useRefreshSkillSource()
  const addSource = useAddSkillSource()
  const removeSource = useRemoveSkillSource()

  const handleClone = (skill: Skill) => {
    const newId = `${skill.id}-copy-${Date.now()}`
    const newName = `${skill.name} (Copy)`
    cloneSkill.mutate({ id: skill.id, newId, newName })
  }

  const handleDelete = (skill: Skill) => {
    if (!confirm(`Delete skill "${skill.name}"?`)) return
    deleteSkill.mutate(skill.id)
  }

  const handleToggle = (skill: Skill) => {
    updateSkill.mutate({ id: skill.id, data: { is_enabled: !skill.is_enabled } })
  }

  const handleExport = async (skill: Skill) => {
    exportSkill.mutate(skill.id, {
      onSuccess: content => {
        const blob = new Blob([content], { type: 'text/markdown' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${skill.id}-SKILL.md`
        a.click()
        URL.revokeObjectURL(url)
      },
    })
  }

  const handleInstall = (skill: RemoteSkill) => {
    setInstallingId(skill.id)
    installSkill.mutate(
      { sourceId: skill.source_id, skillId: skill.id },
      {
        onSettled: () => setInstallingId(null),
      }
    )
  }

  const handleRefresh = (sourceId: string) => {
    setRefreshingId(sourceId)
    refreshSource.mutate(sourceId, {
      onSettled: () => setRefreshingId(null),
    })
  }

  const handleDeleteSource = (sourceId: string) => {
    if (!confirm('确定要删除这个源吗？')) return
    setDeletingSourceId(sourceId)
    removeSource.mutate(sourceId, {
      onSettled: () => setDeletingSourceId(null),
    })
  }

  const handleAddSource = (data: AddSourceRequest) => {
    addSource.mutate(data, {
      onSuccess: () => setShowAddSourceModal(false),
    })
  }

  // 过滤技能
  const filteredSkills = skills.filter(s => {
    if (filter === 'enabled') return s.is_enabled
    if (filter === 'disabled') return !s.is_enabled
    return true
  })

  // 按类别分组
  const categories = Array.from(new Set(filteredSkills.map(s => s.category)))
  const groupedSkills = categories.reduce(
    (acc, category) => {
      acc[category] = filteredSkills.filter(s => s.category === category)
      return acc
    },
    {} as Record<SkillCategory, Skill[]>
  )

  // 按源分组远程 Skills
  const groupedRemoteSkills = remoteSkills.reduce(
    (acc, skill) => {
      if (!acc[skill.source_id]) {
        acc[skill.source_id] = []
      }
      acc[skill.source_id].push(skill)
      return acc
    },
    {} as Record<string, RemoteSkill[]>
  )

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Zap className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Skills</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {activeTab === 'installed' && (
            <>
              {/* Filter */}
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
              <button className="btn btn-primary" onClick={() => navigate({ to: '/skills/new' })}>
                <Plus className="w-4 h-4" />
                New Skill
              </button>
            </>
          )}
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to fetch skills'}
            </span>
          </div>
        )}

        {/* Description */}
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-foreground mb-2">Skills</h1>
          <p className="text-foreground/90">
            Skills are reusable task templates that define how agents should handle specific tasks.
            Use commands like <code className="text-emerald-400">/commit</code> or{' '}
            <code className="text-emerald-400">/review-pr</code> to invoke them.
          </p>
        </div>

        {/* Tabs */}
        <div className="flex items-center gap-1 mb-6 border-b border-border">
          <button
            onClick={() => setActiveTab('installed')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'installed'
                ? 'border-emerald-500 text-emerald-400'
                : 'border-transparent text-muted-foreground hover:text-primary'
            }`}
          >
            <Zap className="w-4 h-4 inline mr-2" />
            已安装 ({skills.length})
          </button>
          <button
            onClick={() => setActiveTab('store')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'store'
                ? 'border-emerald-500 text-emerald-400'
                : 'border-transparent text-muted-foreground hover:text-primary'
            }`}
          >
            <Store className="w-4 h-4 inline mr-2" />
            商店 ({remoteSkills.filter(s => !s.is_installed).length})
          </button>
          <button
            onClick={() => setActiveTab('sources')}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'sources'
                ? 'border-emerald-500 text-emerald-400'
                : 'border-transparent text-muted-foreground hover:text-primary'
            }`}
          >
            <Github className="w-4 h-4 inline mr-2" />
            源 ({sources.length})
          </button>
        </div>

        {/* Tab Content */}
        {activeTab === 'installed' && (
          <>
            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
              </div>
            ) : filteredSkills.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Zap className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-foreground/90 text-lg">No skills found</p>
                <p className="text-muted-foreground mt-2">
                  {filter !== 'all'
                    ? 'Try changing the filter or create a new skill'
                    : 'Create your first skill or install from the store'}
                </p>
                <button
                  onClick={() => setActiveTab('store')}
                  className="btn btn-primary mt-4"
                >
                  <Store className="w-4 h-4" />
                  浏览商店
                </button>
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
                      <span className="text-sm text-foreground/90">({groupedSkills[category].length})</span>
                    </div>
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                      {groupedSkills[category].map(skill => (
                        <SkillCard
                          key={skill.id}
                          skill={skill}
                          onClone={() => handleClone(skill)}
                          onDelete={() => handleDelete(skill)}
                          onToggle={() => handleToggle(skill)}
                          onExport={() => handleExport(skill)}
                          onClick={() => navigate({ to: `/skills/${skill.id}` })}
                        />
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {activeTab === 'store' && (
          <>
            {loadingRemote ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
              </div>
            ) : remoteSkills.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Store className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-foreground/90 text-lg">商店暂无 Skills</p>
                <p className="text-muted-foreground mt-2">
                  请检查网络连接或添加新的 Skill 源
                </p>
              </div>
            ) : (
              <div className="space-y-8">
                {Object.entries(groupedRemoteSkills).map(([sourceId, sourceSkills]) => {
                  const source = sources.find(s => s.id === sourceId)
                  return (
                    <div key={sourceId}>
                      <div className="flex items-center gap-3 mb-4">
                        <div className="w-8 h-8 rounded-lg bg-gray-700 flex items-center justify-center">
                          <Github className="w-4 h-4 text-white" />
                        </div>
                        <h2 className="text-lg font-semibold text-foreground">
                          {source?.name || sourceId}
                        </h2>
                        <span className="text-sm text-foreground/90">({sourceSkills.length})</span>
                        {fetchingRemote && (
                          <Loader2 className="w-4 h-4 text-muted-foreground animate-spin" />
                        )}
                      </div>
                      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                        {sourceSkills.map(skill => (
                          <RemoteSkillCard
                            key={`${skill.source_id}-${skill.id}`}
                            skill={skill}
                            onInstall={() => handleInstall(skill)}
                            installing={installingId === skill.id}
                          />
                        ))}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </>
        )}

        {activeTab === 'sources' && (
          <>
            {/* 添加源按钮 */}
            <div className="flex justify-end mb-4">
              <button
                onClick={() => setShowAddSourceModal(true)}
                className="btn btn-primary"
              >
                <Plus className="w-4 h-4" />
                添加源
              </button>
            </div>

            {loadingSources ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
              </div>
            ) : sources.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Github className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-foreground/90 text-lg">暂无 Skill 源</p>
                <p className="text-muted-foreground mt-2">
                  添加 GitHub 仓库作为 Skill 源
                </p>
                <button
                  onClick={() => setShowAddSourceModal(true)}
                  className="btn btn-primary mt-4"
                >
                  <Plus className="w-4 h-4" />
                  添加源
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                {sources.map(source => (
                  <SourceCard
                    key={source.id}
                    source={source}
                    onRefresh={() => handleRefresh(source.id)}
                    onDelete={() => handleDeleteSource(source.id)}
                    refreshing={refreshingId === source.id}
                    deleting={deletingSourceId === source.id}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </div>

      {/* 添加源对话框 */}
      <AddSourceModal
        isOpen={showAddSourceModal}
        onClose={() => setShowAddSourceModal(false)}
        onSubmit={handleAddSource}
        submitting={addSource.isPending}
      />
    </div>
  )
}
