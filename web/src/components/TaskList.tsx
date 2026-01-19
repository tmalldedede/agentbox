import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  RefreshCw,
  AlertCircle,
  Loader2,
  ListTodo,
  Clock,
  CheckCircle2,
  XCircle,
  PlayCircle,
  PauseCircle,
  Trash2,
  FileText,
  ChevronRight,
  Calendar,
  Timer,
  DollarSign,
  Zap,
} from 'lucide-react'
import type { Task, TaskStatus, Profile } from '../types'
import { useTasks, useTaskLogs, useCreateTask, useCancelTask, useProfiles } from '../hooks'

// 状态图标映射
const statusIcons: Record<TaskStatus, React.ReactNode> = {
  pending: <Clock className="w-4 h-4" />,
  queued: <PauseCircle className="w-4 h-4" />,
  running: <PlayCircle className="w-4 h-4 animate-pulse" />,
  completed: <CheckCircle2 className="w-4 h-4" />,
  failed: <XCircle className="w-4 h-4" />,
  cancelled: <XCircle className="w-4 h-4" />,
}

// 状态颜色映射
const statusColors: Record<TaskStatus, string> = {
  pending: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  queued: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  running: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  completed: 'bg-green-500/20 text-green-400 border-green-500/30',
  failed: 'bg-red-500/20 text-red-400 border-red-500/30',
  cancelled: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
}

// 格式化时间
function formatTime(dateStr?: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 格式化持续时间
function formatDuration(seconds?: number): string {
  if (!seconds) return '-'
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${Math.floor(seconds % 60)}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

// Task 卡片组件
function TaskCard({
  task,
  onCancel,
  onViewLogs,
  onClick,
}: {
  task: Task
  onCancel: () => void
  onViewLogs: () => void
  onClick: () => void
}) {
  const statusColor = statusColors[task.status] || statusColors.pending
  const statusIcon = statusIcons[task.status] || statusIcons.pending
  const isActive = task.status === 'running' || task.status === 'queued' || task.status === 'pending'

  return (
    <div
      className={`card p-4 cursor-pointer group transition-colors ${
        isActive ? 'border-emerald-500/30 hover:border-emerald-500/50' : 'hover:border-gray-500/50'
      }`}
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${statusColor}`}>
          {statusIcon}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-primary truncate">{task.id.slice(0, 8)}</span>
            <span className={`text-xs px-2 py-0.5 rounded border ${statusColor}`}>{task.status}</span>
          </div>

          {/* Profile & Agent */}
          <div className="flex items-center gap-2 mt-1">
            <Zap className="w-3 h-3 text-muted" />
            <span className="text-sm text-emerald-400">{task.profile_name}</span>
            <span className="text-muted">·</span>
            <span className="text-sm text-muted">{task.agent_type}</span>
          </div>

          {/* Prompt */}
          <p className="text-sm text-secondary mt-1 line-clamp-2">{task.prompt}</p>

          {/* Meta */}
          <div className="flex items-center gap-4 mt-3 text-xs text-muted">
            <div className="flex items-center gap-1">
              <Calendar className="w-3 h-3" />
              <span>{formatTime(task.created_at)}</span>
            </div>
            {task.result?.usage && (
              <>
                <div className="flex items-center gap-1">
                  <Timer className="w-3 h-3" />
                  <span>{formatDuration(task.result.usage.duration_seconds)}</span>
                </div>
                {task.result.usage.cost_usd && (
                  <div className="flex items-center gap-1">
                    <DollarSign className="w-3 h-3" />
                    <span>${task.result.usage.cost_usd.toFixed(4)}</span>
                  </div>
                )}
              </>
            )}
          </div>

          {/* Error message */}
          {task.error_message && (
            <div className="mt-2 p-2 rounded bg-red-500/10 border border-red-500/20">
              <p className="text-xs text-red-400 line-clamp-2">{task.error_message}</p>
            </div>
          )}
        </div>

        {/* Actions */}
        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          {/* View Logs */}
          <button
            onClick={e => {
              e.stopPropagation()
              onViewLogs()
            }}
            className="btn btn-ghost btn-icon"
            title="View Logs"
          >
            <FileText className="w-4 h-4" />
          </button>

          {/* Cancel Button (only for active tasks) */}
          {isActive && (
            <button
              onClick={e => {
                e.stopPropagation()
                onCancel()
              }}
              className="btn btn-ghost btn-icon text-red-400"
              title="Cancel"
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

// 创建任务模态框
function CreateTaskModal({
  isOpen,
  onClose,
  onCreate,
  profiles,
  isCreating,
}: {
  isOpen: boolean
  onClose: () => void
  onCreate: (profileId: string, prompt: string) => void
  profiles: Profile[]
  isCreating: boolean
}) {
  const [profileId, setProfileId] = useState('')
  const [prompt, setPrompt] = useState('')

  useEffect(() => {
    if (profiles.length > 0 && !profileId) {
      setProfileId(profiles[0].id)
    }
  }, [profiles, profileId])

  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (profileId && prompt.trim()) {
      onCreate(profileId, prompt.trim())
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="card w-full max-w-lg p-6">
        <h2 className="text-xl font-bold text-primary mb-4">Create New Task</h2>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Profile</label>
              <select
                value={profileId}
                onChange={e => setProfileId(e.target.value)}
                className="input w-full"
                required
              >
                {profiles.map(profile => (
                  <option key={profile.id} value={profile.id}>
                    {profile.name} ({profile.adapter})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">Prompt</label>
              <textarea
                value={prompt}
                onChange={e => setPrompt(e.target.value)}
                className="input w-full h-32 resize-none"
                placeholder="Describe what you want the agent to do..."
                required
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 mt-6">
            <button type="button" onClick={onClose} className="btn btn-ghost" disabled={isCreating}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={isCreating}>
              {isCreating ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
              Create Task
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// 日志查看模态框
function LogsModal({
  isOpen,
  onClose,
  taskId,
}: {
  isOpen: boolean
  onClose: () => void
  taskId: string | null
}) {
  const { data: logsData, isLoading } = useTaskLogs(taskId || undefined)

  if (!isOpen || !taskId) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="card w-full max-w-4xl max-h-[80vh] p-6 flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-primary">Task Logs: {taskId.slice(0, 8)}</h2>
          <button onClick={onClose} className="btn btn-ghost btn-icon">
            ✕
          </button>
        </div>
        <div className="flex-1 overflow-auto bg-secondary rounded-lg p-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-32">
              <Loader2 className="w-6 h-6 text-emerald-400 animate-spin" />
            </div>
          ) : logsData?.logs ? (
            <pre className="text-sm text-secondary font-mono whitespace-pre-wrap">
              {logsData.logs}
            </pre>
          ) : (
            <p className="text-muted text-center">No logs available</p>
          )}
        </div>
      </div>
    </div>
  )
}

export default function TaskList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | TaskStatus>('all')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showLogsModal, setShowLogsModal] = useState(false)
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  // React Query hooks
  const { data: tasksData, isLoading, isFetching, error, refetch } = useTasks({ limit: 100 })
  const { data: profiles = [] } = useProfiles()
  const createTask = useCreateTask()
  const cancelTask = useCancelTask()

  const tasks = tasksData?.tasks || []

  const handleCreate = (profileId: string, prompt: string) => {
    createTask.mutate(
      { profile_id: profileId, prompt },
      {
        onSuccess: () => {
          setShowCreateModal(false)
        },
      }
    )
  }

  const handleCancel = (task: Task) => {
    if (!confirm(`Cancel task "${task.id.slice(0, 8)}"?`)) return
    cancelTask.mutate(task.id)
  }

  const handleViewLogs = (task: Task) => {
    setSelectedTaskId(task.id)
    setShowLogsModal(true)
  }

  // 过滤任务
  const filteredTasks = tasks.filter(t => {
    if (filter === 'all') return true
    return t.status === filter
  })

  // 统计
  const stats = {
    total: tasks.length,
    running: tasks.filter(t => t.status === 'running').length,
    queued: tasks.filter(t => t.status === 'queued' || t.status === 'pending').length,
    completed: tasks.filter(t => t.status === 'completed').length,
    failed: tasks.filter(t => t.status === 'failed').length,
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <ListTodo className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Tasks</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Filter */}
          <select
            value={filter}
            onChange={e => setFilter(e.target.value as typeof filter)}
            className="input py-2 px-3 text-sm"
          >
            <option value="all">All ({stats.total})</option>
            <option value="running">Running ({stats.running})</option>
            <option value="queued">Queued ({stats.queued})</option>
            <option value="completed">Completed ({stats.completed})</option>
            <option value="failed">Failed ({stats.failed})</option>
          </select>

          <button onClick={() => refetch()} className="btn btn-ghost btn-icon" disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            <Plus className="w-4 h-4" />
            New Task
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to fetch tasks'}
            </span>
          </div>
        )}

        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
          <div className="card p-4">
            <div className="text-2xl font-bold text-primary">{stats.total}</div>
            <div className="text-sm text-muted">Total Tasks</div>
          </div>
          <div className="card p-4 border-emerald-500/30">
            <div className="text-2xl font-bold text-emerald-400">{stats.running}</div>
            <div className="text-sm text-muted">Running</div>
          </div>
          <div className="card p-4 border-blue-500/30">
            <div className="text-2xl font-bold text-blue-400">{stats.queued}</div>
            <div className="text-sm text-muted">Queued</div>
          </div>
          <div className="card p-4 border-green-500/30">
            <div className="text-2xl font-bold text-green-400">{stats.completed}</div>
            <div className="text-sm text-muted">Completed</div>
          </div>
          <div className="card p-4 border-red-500/30">
            <div className="text-2xl font-bold text-red-400">{stats.failed}</div>
            <div className="text-sm text-muted">Failed</div>
          </div>
        </div>

        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-primary mb-2">Async Tasks</h1>
          <p className="text-secondary">
            Tasks are background jobs that run agents autonomously. Create a task with a prompt, and
            the system will queue it, execute it in a container, and store the results.
          </p>
        </div>

        {isLoading && tasks.length === 0 ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : filteredTasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <ListTodo className="w-16 h-16 text-muted mb-4" />
            <p className="text-secondary text-lg">No tasks found</p>
            <p className="text-muted mt-2">
              {filter !== 'all'
                ? 'Try changing the filter or create a new task'
                : 'Create your first task to get started'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {filteredTasks.map(task => (
              <TaskCard
                key={task.id}
                task={task}
                onCancel={() => handleCancel(task)}
                onViewLogs={() => handleViewLogs(task)}
                onClick={() => {}} // TODO: Task detail page
              />
            ))}
          </div>
        )}
      </div>

      {/* Create Modal */}
      <CreateTaskModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreate={handleCreate}
        profiles={profiles}
        isCreating={createTask.isPending}
      />

      {/* Logs Modal */}
      <LogsModal
        isOpen={showLogsModal}
        onClose={() => {
          setShowLogsModal(false)
          setSelectedTaskId(null)
        }}
        taskId={selectedTaskId}
      />
    </div>
  )
}
