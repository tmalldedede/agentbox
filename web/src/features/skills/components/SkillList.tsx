import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Copy,
  Trash2,
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
  MoreVertical,
  Edit,
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

// 类别图标映射
const categoryIcons: Record<SkillCategory, React.ReactNode> = {
  coding: <Code className="w-5 h-5" />,
  review: <FileSearch className="w-5 h-5" />,
  docs: <FileText className="w-5 h-5" />,
  security: <Shield className="w-5 h-5" />,
  testing: <TestTube className="w-5 h-5" />,
  other: <Box className="w-5 h-5" />,
}

// 类别颜色映射
const categoryBgColors: Record<SkillCategory, string> = {
  coding: 'bg-blue-500/20 text-blue-400',
  review: 'bg-purple-500/20 text-purple-400',
  docs: 'bg-emerald-500/20 text-emerald-400',
  security: 'bg-red-500/20 text-red-400',
  testing: 'bg-amber-500/20 text-amber-400',
  other: 'bg-gray-500/20 text-gray-400',
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
  const bgColor = categoryBgColors[skill.category] || categoryBgColors.other
  const icon = categoryIcons[skill.category] || categoryIcons.other

  return (
    <Card
      className={`cursor-pointer transition-colors ${
        skill.is_enabled
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
                <CardTitle className="text-base">{skill.name}</CardTitle>
                {skill.is_built_in && (
                  <Badge variant="secondary" className="text-xs">
                    <Lock className="w-3 h-3 mr-1" />
                    Built-in
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <Terminal className="w-3 h-3 text-muted-foreground" />
                <code className="text-xs text-emerald-400 font-mono">{skill.command}</code>
              </div>
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
                onExport()
              }}>
                <Download className="w-4 h-4 mr-2" />
                Export
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onToggle()
              }}>
                {skill.is_enabled ? (
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
              {!skill.is_built_in && (
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
        <CardDescription className="mb-3 line-clamp-2">
          {skill.description || skill.prompt.slice(0, 100)}
        </CardDescription>
        <div className="flex items-center gap-2 flex-wrap">
          {skill.is_enabled ? (
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
            {skill.category}
          </Badge>
          {skill.tags?.slice(0, 2).map(tag => (
            <Badge key={tag} variant="outline" className="text-xs">
              {tag}
            </Badge>
          ))}
          {skill.required_mcp && skill.required_mcp.length > 0 && (
            <Badge variant="outline" className="text-xs text-amber-600">
              {skill.required_mcp.length} MCP
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
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
  const bgColor = categoryBgColors[skill.category as SkillCategory] || categoryBgColors.other
  const icon = categoryIcons[skill.category as SkillCategory] || categoryIcons.other

  return (
    <Card className="transition-colors hover:border-blue-500/50">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-lg ${bgColor} flex items-center justify-center`}>
              {icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{skill.name}</CardTitle>
                <Badge variant="secondary" className="text-xs">
                  {skill.source_name}
                </Badge>
                {skill.is_installed && (
                  <Badge variant="default" className="bg-green-500 text-xs">
                    <Check className="w-3 h-3 mr-1" />
                    Installed
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <Terminal className="w-3 h-3 text-muted-foreground" />
                <code className="text-xs text-blue-400 font-mono">{skill.command}</code>
              </div>
            </div>
          </div>
          {skill.is_installed ? (
            <Button variant="ghost" disabled className="text-green-500">
              <Check className="w-4 h-4 mr-2" />
              Installed
            </Button>
          ) : (
            <Button onClick={onInstall} disabled={installing}>
              {installing ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Download className="w-4 h-4 mr-2" />
              )}
              Install
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <CardDescription className="mb-3 line-clamp-2">
          {skill.description || 'No description'}
        </CardDescription>
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="outline" className="text-xs capitalize">
            {skill.category}
          </Badge>
          {skill.author && (
            <Badge variant="outline" className="text-xs">
              by {skill.author}
            </Badge>
          )}
          {skill.version && (
            <Badge variant="outline" className="text-xs">
              v{skill.version}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
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
  const isOfficial = source.type === 'official'

  return (
    <Card className="transition-colors hover:border-primary/50">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gray-700 flex items-center justify-center">
              <Github className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{source.name}</CardTitle>
                <Badge
                  variant={source.type === 'official' ? 'default' : 'secondary'}
                  className={`text-xs ${source.type === 'official' ? 'bg-emerald-500' : ''}`}
                >
                  {source.type}
                </Badge>
                {source.stars !== undefined && source.stars > 0 && (
                  <span className="flex items-center gap-1 text-xs text-amber-400">
                    <Star className="w-3 h-3" />
                    {source.stars}
                  </span>
                )}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {source.owner}/{source.repo} · {source.branch} · {source.path}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              asChild
              className="h-8 w-8"
            >
              <a
                href={`https://github.com/${source.owner}/${source.repo}`}
                target="_blank"
                rel="noopener noreferrer"
                title="View on GitHub"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={onRefresh}
              disabled={refreshing}
              title="Refresh"
            >
              <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
            </Button>
            {!isOfficial && (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-red-400 hover:text-red-500"
                onClick={onDelete}
                disabled={deleting}
                title="Delete"
              >
                {deleting ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Trash2 className="w-4 h-4" />
                )}
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      {source.description && (
        <CardContent>
          <CardDescription>{source.description}</CardDescription>
        </CardContent>
      )}
    </Card>
  )
}

// 添加源对话框
function AddSourceDialog({
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

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Add Skill Source</DialogTitle>
          <DialogDescription>
            Add a GitHub repository as a skill source
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>GitHub URL (Quick Fill)</Label>
            <Input
              placeholder="https://github.com/owner/repo"
              onChange={e => handleGitHubUrlParse(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">Paste GitHub URL to auto-fill</p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Owner <span className="text-red-400">*</span></Label>
              <Input
                value={formData.owner}
                onChange={e => setFormData(prev => ({ ...prev, owner: e.target.value }))}
                placeholder="anthropics"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>Repo <span className="text-red-400">*</span></Label>
              <Input
                value={formData.repo}
                onChange={e => setFormData(prev => ({ ...prev, repo: e.target.value }))}
                placeholder="skills"
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Branch</Label>
              <Input
                value={formData.branch}
                onChange={e => setFormData(prev => ({ ...prev, branch: e.target.value }))}
                placeholder="main"
              />
            </div>
            <div className="space-y-2">
              <Label>Skills Path</Label>
              <Input
                value={formData.path}
                onChange={e => setFormData(prev => ({ ...prev, path: e.target.value }))}
                placeholder="skills"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>ID <span className="text-red-400">*</span></Label>
              <Input
                value={formData.id}
                onChange={e => setFormData(prev => ({ ...prev, id: e.target.value }))}
                placeholder="my-skills"
                required
              />
            </div>
            <div className="space-y-2">
              <Label>Name <span className="text-red-400">*</span></Label>
              <Input
                value={formData.name}
                onChange={e => setFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="My Skills"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Description</Label>
            <Input
              value={formData.description}
              onChange={e => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Skills repository description"
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={submitting || !formData.id || !formData.owner || !formData.repo}
            >
              {submitting ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Plus className="w-4 h-4 mr-2" />
              )}
              Add Source
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function SkillList() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('installed')
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [installingId, setInstallingId] = useState<string | null>(null)
  const [refreshingId, setRefreshingId] = useState<string | null>(null)
  const [deletingSourceId, setDeletingSourceId] = useState<string | null>(null)
  const [showAddSourceDialog, setShowAddSourceDialog] = useState(false)

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
    if (!confirm('Are you sure you want to delete this source?')) return
    setDeletingSourceId(sourceId)
    removeSource.mutate(sourceId, {
      onSettled: () => setDeletingSourceId(null),
    })
  }

  const handleAddSource = (data: AddSourceRequest) => {
    addSource.mutate(data, {
      onSuccess: () => setShowAddSourceDialog(false),
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
              <button className="btn btn-primary" onClick={() => navigate({ to: '/skills/new' })}>
                <Plus className="w-4 h-4" />
                New Skill
              </button>
            </>
          )}
          {activeTab === 'sources' && (
            <Button onClick={() => setShowAddSourceDialog(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Source
            </Button>
          )}
        </div>
      </header>

      <div className="p-6">
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to fetch skills'}
            </span>
          </div>
        )}

        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">Skills</h1>
          <p className="text-muted-foreground">
            Skills are reusable task templates that define how agents should handle specific tasks.
            Use commands like <code className="text-emerald-400">/commit</code> or{' '}
            <code className="text-emerald-400">/review-pr</code> to invoke them.
          </p>
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="mb-6">
            <TabsTrigger value="installed" className="gap-2">
              <Zap className="w-4 h-4" />
              Installed ({skills.length})
            </TabsTrigger>
            <TabsTrigger value="store" className="gap-2">
              <Store className="w-4 h-4" />
              Store ({remoteSkills.filter(s => !s.is_installed).length})
            </TabsTrigger>
            <TabsTrigger value="sources" className="gap-2">
              <Github className="w-4 h-4" />
              Sources ({sources.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="installed">
            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
              </div>
            ) : filteredSkills.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Zap className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-muted-foreground text-lg">No skills found</p>
                <p className="text-muted-foreground mt-2">
                  {filter !== 'all'
                    ? 'Try changing the filter or create a new skill'
                    : 'Create your first skill or install from the store'}
                </p>
                <Button className="mt-4" onClick={() => setActiveTab('store')}>
                  <Store className="w-4 h-4 mr-2" />
                  Browse Store
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
                      <span className="text-sm text-muted-foreground">({groupedSkills[category].length})</span>
                    </div>
                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
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
          </TabsContent>

          <TabsContent value="store">
            {loadingRemote ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
              </div>
            ) : remoteSkills.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Store className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-muted-foreground text-lg">No skills in store</p>
                <p className="text-muted-foreground mt-2">
                  Check your network connection or add a new skill source
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
                        <span className="text-sm text-muted-foreground">({sourceSkills.length})</span>
                        {fetchingRemote && (
                          <Loader2 className="w-4 h-4 text-muted-foreground animate-spin" />
                        )}
                      </div>
                      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
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
          </TabsContent>

          <TabsContent value="sources">
            {loadingSources ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
              </div>
            ) : sources.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-64 text-center">
                <Github className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-muted-foreground text-lg">No skill sources</p>
                <p className="text-muted-foreground mt-2">
                  Add a GitHub repository as a skill source
                </p>
                <Button className="mt-4" onClick={() => setShowAddSourceDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  Add Source
                </Button>
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
          </TabsContent>
        </Tabs>
      </div>

      <AddSourceDialog
        isOpen={showAddSourceDialog}
        onClose={() => setShowAddSourceDialog(false)}
        onSubmit={handleAddSource}
        submitting={addSource.isPending}
      />
    </div>
  )
}
