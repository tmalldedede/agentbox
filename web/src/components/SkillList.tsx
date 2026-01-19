import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
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
} from 'lucide-react'
import type { Skill, SkillCategory } from '../types'
import { useSkills, useUpdateSkill, useDeleteSkill, useCloneSkill, useExportSkill } from '../hooks'

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
            <span className="font-semibold text-primary truncate">{skill.name}</span>
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
            <Terminal className="w-3 h-3 text-muted" />
            <code className="text-sm text-emerald-400 font-mono">{skill.command}</code>
          </div>

          <p className="text-sm text-muted mt-1 line-clamp-2">
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
                className="text-xs px-2 py-0.5 rounded bg-secondary text-muted"
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
        <ChevronRight className="w-5 h-5 text-muted group-hover:text-emerald-400 transition-colors" />
      </div>
    </div>
  )
}

export default function SkillList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')

  // React Query hooks
  const { data: skills = [], isLoading, isFetching, error, refetch } = useSkills()
  const updateSkill = useUpdateSkill()
  const deleteSkill = useDeleteSkill()
  const cloneSkill = useCloneSkill()
  const exportSkill = useExportSkill()

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

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Zap className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Skills</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
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
          <button className="btn btn-primary" onClick={() => navigate('/skills/new')}>
            <Plus className="w-4 h-4" />
            New Skill
          </button>
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
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-primary mb-2">Skills</h1>
          <p className="text-secondary">
            Skills are reusable task templates that define how agents should handle specific tasks.
            Use commands like <code className="text-emerald-400">/commit</code> or{' '}
            <code className="text-emerald-400">/review-pr</code> to invoke them.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : filteredSkills.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Zap className="w-16 h-16 text-muted mb-4" />
            <p className="text-secondary text-lg">No skills found</p>
            <p className="text-muted mt-2">
              {filter !== 'all'
                ? 'Try changing the filter or create a new skill'
                : 'Create your first skill to get started'}
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
                  <h2 className="text-lg font-semibold text-primary capitalize">{category}</h2>
                  <span className="text-sm text-muted">({groupedSkills[category].length})</span>
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
                      onClick={() => navigate(`/skills/${skill.id}`)}
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
