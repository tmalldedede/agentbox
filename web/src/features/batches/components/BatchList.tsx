import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  Play,
  Pause,
  Square,
  Trash2,
  Plus,
  RotateCcw,
  Download,
  Loader2,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Layers,
  Activity,
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  useBatches,
  useStartBatch,
  usePauseBatch,
  useCancelBatch,
  useDeleteBatch,
} from '@/hooks'
import { CreateBatchModal } from './CreateBatchModal'
import type { Batch, BatchStatus } from '@/types'
import { api } from '@/services/api'

const statusConfig: Record<BatchStatus, { label: string; icon: React.ReactNode; className: string }> = {
  pending: {
    label: 'Pending',
    icon: <Clock className="h-3 w-3" />,
    className: 'bg-yellow-500/10 text-yellow-600 border-yellow-200',
  },
  running: {
    label: 'Running',
    icon: <Loader2 className="h-3 w-3 animate-spin" />,
    className: 'bg-green-500/10 text-green-600 border-green-200',
  },
  paused: {
    label: 'Paused',
    icon: <Pause className="h-3 w-3" />,
    className: 'bg-blue-500/10 text-blue-600 border-blue-200',
  },
  completed: {
    label: 'Completed',
    icon: <CheckCircle2 className="h-3 w-3" />,
    className: 'bg-gray-500/10 text-gray-600 border-gray-200',
  },
  failed: {
    label: 'Failed',
    icon: <XCircle className="h-3 w-3" />,
    className: 'bg-red-500/10 text-red-600 border-red-200',
  },
  cancelled: {
    label: 'Cancelled',
    icon: <AlertCircle className="h-3 w-3" />,
    className: 'bg-orange-500/10 text-orange-600 border-orange-200',
  },
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString()
}

function formatDuration(startedAt?: string, completedAt?: string) {
  if (!startedAt) return '-'
  const start = new Date(startedAt)
  const end = completedAt ? new Date(completedAt) : new Date()
  const seconds = Math.floor((end.getTime() - start.getTime()) / 1000)
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

export function BatchList() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const { data, isLoading } = useBatches()
  const startBatch = useStartBatch()
  const pauseBatch = usePauseBatch()
  const cancelBatch = useCancelBatch()
  const deleteBatch = useDeleteBatch()

  const batches = data?.batches || []

  // 计算统计数据
  const stats = {
    total: batches.length,
    running: batches.filter(b => b.status === 'running').length,
    completed: batches.filter(b => b.status === 'completed').length,
    failed: batches.filter(b => b.status === 'failed').length,
  }

  const handleStart = (id: string) => {
    startBatch.mutate(id)
  }

  const handlePause = (id: string) => {
    pauseBatch.mutate(id)
  }

  const handleCancel = (id: string) => {
    cancelBatch.mutate(id)
  }

  const handleDelete = () => {
    if (deleteId) {
      deleteBatch.mutate(deleteId)
      setDeleteId(null)
    }
  }

  const handleExport = (id: string, format: 'json' | 'csv') => {
    window.open(api.getBatchExportUrl(id, format), '_blank')
  }

  return (
    <>
      <Header fixed>
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Batches</h2>
            <p className="text-muted-foreground">
              Manage batch processing jobs with worker pools
            </p>
          </div>
          <Button onClick={() => setShowCreateModal(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Batch
          </Button>
        </div>

        {/* Stats Cards */}
        <div className='grid gap-4 grid-cols-2 sm:grid-cols-4'>
          <div className='rounded-lg border p-3'>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <Activity className='h-4 w-4' />
              Total
            </div>
            <p className='text-2xl font-bold mt-1'>{stats.total}</p>
          </div>
          <div className='rounded-lg border p-3'>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <Loader2 className='h-4 w-4 text-green-500' />
              Running
            </div>
            <p className='text-2xl font-bold mt-1'>{stats.running}</p>
          </div>
          <div className='rounded-lg border p-3'>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <CheckCircle2 className='h-4 w-4 text-gray-500' />
              Completed
            </div>
            <p className='text-2xl font-bold mt-1'>{stats.completed}</p>
          </div>
          <div className='rounded-lg border p-3'>
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <XCircle className='h-4 w-4 text-red-500' />
              Failed
            </div>
            <p className='text-2xl font-bold mt-1'>{stats.failed}</p>
          </div>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : batches.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <div className="rounded-full bg-muted p-4 mb-4">
                <Layers className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-semibold mb-2">No batches yet</h3>
              <p className="text-muted-foreground text-center mb-4">
                Create a batch to process multiple tasks efficiently with worker pools.
              </p>
              <Button onClick={() => setShowCreateModal(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create First Batch
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>All Batches</CardTitle>
              <CardDescription>
                {batches.length} batch{batches.length !== 1 ? 'es' : ''} total
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Progress</TableHead>
                    <TableHead>Workers</TableHead>
                    <TableHead>Duration</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {batches.map((batch) => (
                    <BatchRow
                      key={batch.id}
                      batch={batch}
                      onStart={handleStart}
                      onPause={handlePause}
                      onCancel={handleCancel}
                      onDelete={(id) => setDeleteId(id)}
                      onExport={handleExport}
                    />
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        )}

        <CreateBatchModal
          open={showCreateModal}
          onOpenChange={setShowCreateModal}
        />

        <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete Batch</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to delete this batch? This action cannot be undone.
                All tasks and results will be permanently deleted.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground">
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </Main>
    </>
  )
}

function BatchRow({
  batch,
  onStart,
  onPause,
  onCancel,
  onDelete,
  onExport,
}: {
  batch: Batch
  onStart: (id: string) => void
  onPause: (id: string) => void
  onCancel: (id: string) => void
  onDelete: (id: string) => void
  onExport: (id: string, format: 'json' | 'csv') => void
}) {
  const status = statusConfig[batch.status]
  const progress = batch.total_tasks > 0
    ? ((batch.completed + batch.failed) / batch.total_tasks) * 100
    : 0

  return (
    <TableRow>
      <TableCell>
        <Link
          to="/batches/$batchId"
          params={{ batchId: batch.id }}
          className="font-medium hover:underline"
        >
          {batch.name || batch.id}
        </Link>
        <div className="text-xs text-muted-foreground">{batch.id}</div>
      </TableCell>
      <TableCell>
        <Badge variant="outline" className={status.className}>
          {status.icon}
          <span className="ml-1">{status.label}</span>
        </Badge>
      </TableCell>
      <TableCell>
        <div className="w-32 space-y-1">
          <Progress value={progress} className="h-2" />
          <div className="text-xs text-muted-foreground">
            {batch.completed}/{batch.total_tasks}
            {batch.failed > 0 && (
              <span className="text-red-500 ml-1">({batch.failed} failed)</span>
            )}
          </div>
        </div>
      </TableCell>
      <TableCell>
        <span className="text-sm">{batch.concurrency}</span>
        {batch.workers && batch.status === 'running' && (
          <span className="text-xs text-muted-foreground ml-1">
            ({batch.workers.filter(w => w.status === 'busy').length} busy)
          </span>
        )}
      </TableCell>
      <TableCell>
        <span className="text-sm">
          {formatDuration(batch.started_at, batch.completed_at)}
        </span>
        {batch.estimated_eta && batch.status === 'running' && (
          <div className="text-xs text-muted-foreground">ETA: {batch.estimated_eta}</div>
        )}
      </TableCell>
      <TableCell>
        <span className="text-sm">{formatDate(batch.created_at)}</span>
      </TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-1">
          {batch.status === 'pending' && (
            <Button size="icon" variant="ghost" onClick={() => onStart(batch.id)}>
              <Play className="h-4 w-4" />
            </Button>
          )}
          {batch.status === 'running' && (
            <Button size="icon" variant="ghost" onClick={() => onPause(batch.id)}>
              <Pause className="h-4 w-4" />
            </Button>
          )}
          {batch.status === 'paused' && (
            <Button size="icon" variant="ghost" onClick={() => onStart(batch.id)}>
              <Play className="h-4 w-4" />
            </Button>
          )}
          {(batch.status === 'running' || batch.status === 'paused') && (
            <Button size="icon" variant="ghost" onClick={() => onCancel(batch.id)}>
              <Square className="h-4 w-4" />
            </Button>
          )}
          {(batch.status === 'completed' || batch.status === 'failed') && (
            <>
              <Button size="icon" variant="ghost" onClick={() => onExport(batch.id, 'json')}>
                <Download className="h-4 w-4" />
              </Button>
              {batch.failed > 0 && (
                <Button size="icon" variant="ghost">
                  <RotateCcw className="h-4 w-4" />
                </Button>
              )}
            </>
          )}
          <Button size="icon" variant="ghost" onClick={() => onDelete(batch.id)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}
