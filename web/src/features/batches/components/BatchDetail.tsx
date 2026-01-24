import { useState, useEffect, useRef } from 'react'
import { Link } from '@tanstack/react-router'
import {
  ArrowLeft,
  Play,
  Pause,
  Square,
  RotateCcw,
  Download,
  Loader2,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Server,
  Activity,
  Skull,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  useBatch,
  useBatchTasks,
  useBatchStats,
  useBatchDeadTasks,
  useStartBatch,
  usePauseBatch,
  useCancelBatch,
  useRetryBatchFailed,
  useRetryBatchDead,
} from '@/hooks'
import { api } from '@/services/api'
import type { BatchTask, BatchTaskStatus, BatchEvent } from '@/types'

const statusConfig = {
  pending: { label: 'Pending', icon: Clock, className: 'text-yellow-600' },
  running: { label: 'Running', icon: Loader2, className: 'text-green-600' },
  paused: { label: 'Paused', icon: Pause, className: 'text-blue-600' },
  completed: { label: 'Completed', icon: CheckCircle2, className: 'text-gray-600' },
  failed: { label: 'Failed', icon: XCircle, className: 'text-red-600' },
  cancelled: { label: 'Cancelled', icon: AlertCircle, className: 'text-orange-600' },
}

const taskStatusConfig: Record<BatchTaskStatus, { label: string; className: string }> = {
  pending: { label: 'Pending', className: 'bg-yellow-500/10 text-yellow-600 border-yellow-200' },
  running: { label: 'Running', className: 'bg-green-500/10 text-green-600 border-green-200' },
  completed: { label: 'Completed', className: 'bg-gray-500/10 text-gray-600 border-gray-200' },
  failed: { label: 'Failed', className: 'bg-red-500/10 text-red-600 border-red-200' },
  dead: { label: 'Dead', className: 'bg-purple-500/10 text-purple-600 border-purple-200' },
}

export function BatchDetail({ batchId }: { batchId: string }) {
  const [taskFilter, setTaskFilter] = useState<BatchTaskStatus | undefined>()
  const [_events, setEvents] = useState<BatchEvent[]>([])
  const eventSourceRef = useRef<EventSource | null>(null)

  const { data: batch, isLoading } = useBatch(batchId)
  const { data: tasksData } = useBatchTasks(batchId, { status: taskFilter === 'dead' ? undefined : taskFilter, limit: 50 })
  const { data: stats } = useBatchStats(batchId)
  const { data: deadTasksData } = useBatchDeadTasks(batchId)

  const startBatch = useStartBatch()
  const pauseBatch = usePauseBatch()
  const cancelBatch = useCancelBatch()
  const retryFailed = useRetryBatchFailed()
  const retryDead = useRetryBatchDead()

  // SSE connection for running batches
  useEffect(() => {
    if (batch?.status === 'running') {
      eventSourceRef.current = api.streamBatchEvents(batchId)

      eventSourceRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as BatchEvent
          setEvents((prev) => [...prev.slice(-99), data])
        } catch (e) {
          console.error('Failed to parse event:', e)
        }
      }

      eventSourceRef.current.onerror = () => {
        eventSourceRef.current?.close()
      }

      return () => {
        eventSourceRef.current?.close()
      }
    }
  }, [batch?.status, batchId])

  if (isLoading) {
    return (
      <>
        <Header fixed>
          <div className='ms-auto flex items-center space-x-4'>
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
        </Header>
        <Main className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </Main>
      </>
    )
  }

  if (!batch) {
    return (
      <>
        <Header fixed>
          <div className='ms-auto flex items-center space-x-4'>
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
        </Header>
        <Main className="text-center py-12">
          <h2 className="text-xl font-semibold">Batch not found</h2>
          <Link to="/batches" className="text-primary hover:underline">
            Back to batches
          </Link>
        </Main>
      </>
    )
  }

  const StatusIcon = statusConfig[batch.status].icon
  const progress = batch.total_tasks > 0
    ? ((batch.completed + batch.failed) / batch.total_tasks) * 100
    : 0

  const tasks = tasksData?.tasks || []

  return (
    <>
      <Header fixed>
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Batch Header */}
        <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/batches">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{batch.name || batch.id}</h1>
            <p className="text-muted-foreground">{batch.id}</p>
          </div>
          <Badge variant="outline" className={`${statusConfig[batch.status].className} ml-4`}>
            <StatusIcon className={`h-3 w-3 mr-1 ${batch.status === 'running' ? 'animate-spin' : ''}`} />
            {statusConfig[batch.status].label}
          </Badge>
        </div>

        <div className="flex items-center gap-2">
          {batch.status === 'pending' && (
            <Button onClick={() => startBatch.mutate(batch.id)}>
              <Play className="h-4 w-4 mr-2" />
              Start
            </Button>
          )}
          {batch.status === 'running' && (
            <>
              <Button variant="outline" onClick={() => pauseBatch.mutate(batch.id)}>
                <Pause className="h-4 w-4 mr-2" />
                Pause
              </Button>
              <Button variant="destructive" onClick={() => cancelBatch.mutate(batch.id)}>
                <Square className="h-4 w-4 mr-2" />
                Cancel
              </Button>
            </>
          )}
          {batch.status === 'paused' && (
            <>
              <Button onClick={() => startBatch.mutate(batch.id)}>
                <Play className="h-4 w-4 mr-2" />
                Resume
              </Button>
              <Button variant="destructive" onClick={() => cancelBatch.mutate(batch.id)}>
                <Square className="h-4 w-4 mr-2" />
                Cancel
              </Button>
            </>
          )}
          {(batch.status === 'completed' || batch.status === 'failed') && (
            <>
              <Button variant="outline" onClick={() => window.open(api.getBatchExportUrl(batch.id, 'csv'), '_blank')}>
                <Download className="h-4 w-4 mr-2" />
                Export CSV
              </Button>
              {batch.failed > 0 && (
                <Button onClick={() => retryFailed.mutate(batch.id)}>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Retry Failed
                </Button>
              )}
            </>
          )}
        </div>
      </div>

      {/* Progress Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Progress</p>
                <p className="text-2xl font-bold">{progress.toFixed(1)}%</p>
              </div>
              <div className="text-right">
                <p className="text-sm text-muted-foreground">Tasks</p>
                <p className="text-lg">
                  <span className="text-green-600">{batch.completed}</span>
                  {batch.failed > 0 && (
                    <span className="text-red-600 ml-2">+{batch.failed} failed</span>
                  )}
                  <span className="text-muted-foreground"> / {batch.total_tasks}</span>
                </p>
              </div>
              {batch.tasks_per_sec && batch.tasks_per_sec > 0 && (
                <div className="text-right">
                  <p className="text-sm text-muted-foreground">Speed</p>
                  <p className="text-lg">{batch.tasks_per_sec.toFixed(2)} tasks/s</p>
                </div>
              )}
              {batch.estimated_eta && (
                <div className="text-right">
                  <p className="text-sm text-muted-foreground">ETA</p>
                  <p className="text-lg">{batch.estimated_eta}</p>
                </div>
              )}
            </div>
            <Progress value={progress} className="h-3" />
          </div>
        </CardContent>
      </Card>

      {/* Stats and Workers */}
      <div className="grid grid-cols-2 gap-6">
        {/* Stats */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Statistics
            </CardTitle>
          </CardHeader>
          <CardContent>
            {stats ? (
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">Pending</p>
                  <p className="text-xl font-semibold">{stats.pending}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Running</p>
                  <p className="text-xl font-semibold">{stats.running}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Completed</p>
                  <p className="text-xl font-semibold text-green-600">{stats.completed}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Failed</p>
                  <p className="text-xl font-semibold text-red-600">{stats.failed}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground flex items-center gap-1">
                    <Skull className="h-3 w-3" />
                    Dead Letter
                  </p>
                  <p className="text-xl font-semibold text-purple-600">{stats.dead || 0}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Avg Duration</p>
                  <p className="text-xl font-semibold">{(stats.avg_duration_ms / 1000).toFixed(1)}s</p>
                </div>
              </div>
            ) : (
              <p className="text-muted-foreground">Loading stats...</p>
            )}
          </CardContent>
        </Card>

        {/* Workers */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Server className="h-4 w-4" />
              Workers ({batch.concurrency})
            </CardTitle>
          </CardHeader>
          <CardContent>
            {batch.workers && batch.workers.length > 0 ? (
              <div className="space-y-2">
                {batch.workers.map((worker) => (
                  <div key={worker.id} className="flex items-center justify-between p-2 bg-muted rounded">
                    <div className="flex items-center gap-2">
                      <div className={`w-2 h-2 rounded-full ${
                        worker.status === 'busy' ? 'bg-green-500' :
                        worker.status === 'idle' ? 'bg-yellow-500' :
                        worker.status === 'error' ? 'bg-red-500' : 'bg-gray-500'
                      }`} />
                      <span className="font-mono text-sm">{worker.id}</span>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {worker.completed} completed
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">
                {batch.status === 'pending' ? 'Workers will be created when batch starts' : 'No workers'}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tasks */}
      <Card>
        <CardHeader>
          <CardTitle>Tasks</CardTitle>
          <CardDescription>
            View and filter batch tasks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={taskFilter || 'all'} onValueChange={(v) => setTaskFilter(v === 'all' ? undefined : v as BatchTaskStatus)}>
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger value="all">All</TabsTrigger>
                <TabsTrigger value="pending">Pending</TabsTrigger>
                <TabsTrigger value="running">Running</TabsTrigger>
                <TabsTrigger value="completed">Completed</TabsTrigger>
                <TabsTrigger value="failed">Failed</TabsTrigger>
                <TabsTrigger value="dead" className="flex items-center gap-1">
                  <Skull className="h-3 w-3" />
                  Dead ({stats?.dead || 0})
                </TabsTrigger>
              </TabsList>

              {taskFilter === 'dead' && (deadTasksData?.count || 0) > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => retryDead.mutate({ batchId })}
                  disabled={retryDead.isPending}
                >
                  <RotateCcw className={`h-4 w-4 mr-2 ${retryDead.isPending ? 'animate-spin' : ''}`} />
                  Retry All Dead
                </Button>
              )}
            </div>

            <TabsContent value={taskFilter || 'all'} className="mt-4">
              {taskFilter === 'dead' ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Index</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Attempts</TableHead>
                      <TableHead>Dead At</TableHead>
                      <TableHead>Dead Reason</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {!deadTasksData?.tasks?.length ? (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center text-muted-foreground">
                          No dead tasks
                        </TableCell>
                      </TableRow>
                    ) : (
                      deadTasksData.tasks.map((task) => (
                        <DeadTaskRow
                          key={task.id}
                          task={task}
                          onRetry={() => retryDead.mutate({ batchId, taskIds: [task.id] })}
                        />
                      ))
                    )}
                  </TableBody>
                </Table>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Index</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Worker</TableHead>
                      <TableHead>Duration</TableHead>
                      <TableHead>Attempts</TableHead>
                      <TableHead>Result/Error</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tasks.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center text-muted-foreground">
                          No tasks found
                        </TableCell>
                      </TableRow>
                    ) : (
                      tasks.map((task) => (
                        <TaskRow key={task.id} task={task} />
                      ))
                    )}
                  </TableBody>
                </Table>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Agent ID</p>
              <p className="font-mono">{batch.agent_id}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Concurrency</p>
              <p>{batch.concurrency} workers</p>
            </div>
            <div>
              <p className="text-muted-foreground">Timeout</p>
              <p>{batch.template.timeout}s per task</p>
            </div>
            <div>
              <p className="text-muted-foreground">Max Retries</p>
              <p>{batch.template.max_retries}</p>
            </div>
            <div className="col-span-2">
              <p className="text-muted-foreground">Prompt Template</p>
              <pre className="mt-1 p-2 bg-muted rounded text-xs overflow-x-auto">
                {batch.template.prompt_template}
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>
      </Main>
    </>
  )
}

function DeadTaskRow({ task, onRetry }: { task: BatchTask; onRetry: () => void }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <>
      <TableRow className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
        <TableCell className="font-mono">#{task.index}</TableCell>
        <TableCell>
          <Badge variant="outline" className={taskStatusConfig.dead.className}>
            <Skull className="h-3 w-3 mr-1" />
            Dead
          </Badge>
        </TableCell>
        <TableCell>{task.attempts}</TableCell>
        <TableCell className="text-xs">
          {task.dead_at ? new Date(task.dead_at).toLocaleString() : '-'}
        </TableCell>
        <TableCell className="max-w-md truncate text-red-600">
          {task.dead_reason || task.error || '-'}
        </TableCell>
        <TableCell>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onRetry()
            }}
          >
            <RotateCcw className="h-4 w-4" />
          </Button>
        </TableCell>
      </TableRow>
      {expanded && (
        <TableRow>
          <TableCell colSpan={6} className="bg-muted/50">
            <div className="p-4 space-y-4">
              <div>
                <p className="text-sm font-medium mb-1">Input</p>
                <pre className="text-xs bg-background p-2 rounded overflow-x-auto">
                  {JSON.stringify(task.input, null, 2)}
                </pre>
              </div>
              {task.prompt && (
                <div>
                  <p className="text-sm font-medium mb-1">Rendered Prompt</p>
                  <pre className="text-xs bg-background p-2 rounded overflow-x-auto whitespace-pre-wrap">
                    {task.prompt}
                  </pre>
                </div>
              )}
              <div>
                <p className="text-sm font-medium mb-1 text-purple-600">Dead Reason</p>
                <pre className="text-xs bg-purple-50 dark:bg-purple-950 p-2 rounded overflow-x-auto">
                  {task.dead_reason || 'Unknown'}
                </pre>
              </div>
              {task.error && (
                <div>
                  <p className="text-sm font-medium mb-1 text-red-600">Last Error</p>
                  <pre className="text-xs bg-red-50 dark:bg-red-950 p-2 rounded overflow-x-auto">
                    {task.error}
                  </pre>
                </div>
              )}
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  )
}

function TaskRow({ task }: { task: BatchTask }) {
  const [expanded, setExpanded] = useState(false)
  const status = taskStatusConfig[task.status]

  return (
    <>
      <TableRow className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
        <TableCell className="font-mono">#{task.index}</TableCell>
        <TableCell>
          <Badge variant="outline" className={status.className}>
            {status.label}
          </Badge>
        </TableCell>
        <TableCell className="font-mono text-xs">
          {task.worker_id || '-'}
        </TableCell>
        <TableCell>
          {task.duration_ms ? `${(task.duration_ms / 1000).toFixed(1)}s` : '-'}
        </TableCell>
        <TableCell>{task.attempts}</TableCell>
        <TableCell className="max-w-md truncate">
          {task.error ? (
            <span className="text-red-600">{task.error}</span>
          ) : task.result ? (
            <span className="text-muted-foreground">{task.result.slice(0, 100)}...</span>
          ) : (
            '-'
          )}
        </TableCell>
      </TableRow>
      {expanded && (
        <TableRow>
          <TableCell colSpan={6} className="bg-muted/50">
            <div className="p-4 space-y-4">
              <div>
                <p className="text-sm font-medium mb-1">Input</p>
                <pre className="text-xs bg-background p-2 rounded overflow-x-auto">
                  {JSON.stringify(task.input, null, 2)}
                </pre>
              </div>
              {task.prompt && (
                <div>
                  <p className="text-sm font-medium mb-1">Rendered Prompt</p>
                  <pre className="text-xs bg-background p-2 rounded overflow-x-auto whitespace-pre-wrap">
                    {task.prompt}
                  </pre>
                </div>
              )}
              {task.result && (
                <div>
                  <p className="text-sm font-medium mb-1">Result</p>
                  <pre className="text-xs bg-background p-2 rounded overflow-x-auto whitespace-pre-wrap">
                    {task.result}
                  </pre>
                </div>
              )}
              {task.error && (
                <div>
                  <p className="text-sm font-medium mb-1 text-red-600">Error</p>
                  <pre className="text-xs bg-red-50 dark:bg-red-950 p-2 rounded overflow-x-auto">
                    {task.error}
                  </pre>
                </div>
              )}
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  )
}
