import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  type ColumnFiltersState,
  type SortingState,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  flexRender,
  type ColumnDef,
} from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination, DataTableToolbar, DataTableColumnHeader } from '@/components/data-table'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { useTasks, useTaskStats, useRetryTask, useDeleteTask } from '@/hooks/useTasks'
import { Loader2, RotateCw, Trash2, CheckCircle, XCircle, Clock, Activity } from 'lucide-react'
import type { Task } from '@/types'

const statusColors: Record<string, string> = {
  pending: 'bg-yellow-500/10 text-yellow-600 border-yellow-200',
  queued: 'bg-blue-500/10 text-blue-600 border-blue-200',
  running: 'bg-green-500/10 text-green-600 border-green-200',
  completed: 'bg-gray-500/10 text-gray-600 border-gray-200',
  failed: 'bg-red-500/10 text-red-600 border-red-200',
  cancelled: 'bg-orange-500/10 text-orange-600 border-orange-200',
}

const tasksColumns: ColumnDef<Task>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Task ID' />
    ),
    cell: ({ row }) => (
      <code className='text-xs font-mono text-muted-foreground'>
        {row.original.id}
      </code>
    ),
    enableHiding: false,
  },
  {
    accessorKey: 'agent_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Agent' />
    ),
    cell: ({ row }) => (
      <span className='text-sm font-medium'>
        {row.original.agent_name || row.original.agent_id}
      </span>
    ),
  },
  {
    accessorKey: 'prompt',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Prompt' />
    ),
    cell: ({ row }) => (
      <span className='text-sm max-w-[300px] truncate block'>
        {row.original.prompt}
      </span>
    ),
    enableSorting: false,
  },
  {
    accessorKey: 'turn_count',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Turns' />
    ),
    cell: ({ row }) => (
      <span className='text-sm text-muted-foreground'>
        {row.original.turn_count}
      </span>
    ),
  },
  {
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const status = row.original.status
      return (
        <Badge
          variant='outline'
          className={`text-xs capitalize ${statusColors[status] || ''}`}
        >
          {status}
        </Badge>
      )
    },
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Created' />
    ),
    cell: ({ row }) => (
      <span className='text-xs text-muted-foreground'>
        {new Date(row.original.created_at).toLocaleString()}
      </span>
    ),
  },
  {
    id: 'actions',
    header: 'Actions',
    cell: ({ row }) => <TaskActions task={row.original} />,
    enableSorting: false,
  },
]

function TaskActions({ task }: { task: Task }) {
  const navigate = useNavigate()
  const retryTask = useRetryTask()
  const deleteTask = useDeleteTask()
  const canRetry = task.status === 'failed' || task.status === 'cancelled'
  const canDelete = task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled'

  return (
    <div className='flex items-center gap-1' onClick={(e) => e.stopPropagation()}>
      {canRetry && (
        <Button
          variant='ghost'
          size='icon'
          className='h-7 w-7'
          title='Retry'
          onClick={() => retryTask.mutate(task.id, {
            onSuccess: (newTask) => navigate({ to: `/tasks/${newTask.id}` }),
          })}
          disabled={retryTask.isPending}
        >
          <RotateCw className='h-3.5 w-3.5' />
        </Button>
      )}
      {canDelete && (
        <Button
          variant='ghost'
          size='icon'
          className='h-7 w-7 text-destructive hover:text-destructive'
          title='Delete'
          onClick={() => deleteTask.mutate(task.id)}
          disabled={deleteTask.isPending}
        >
          <Trash2 className='h-3.5 w-3.5' />
        </Button>
      )}
    </div>
  )
}

export function Tasks() {
  const navigate = useNavigate()
  const { data, isLoading } = useTasks({ limit: 100 })
  const { data: stats } = useTaskStats()
  const tasks = data?.tasks || []

  const [sorting, setSorting] = useState<SortingState>([
    { id: 'created_at', desc: true },
  ])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  const table = useReactTable({
    data: tasks,
    columns: tasksColumns,
    state: { sorting, columnFilters },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  return (
    <>
      <Header fixed>
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>Tasks</h2>
          <p className='text-muted-foreground'>
            Agent task execution list
          </p>
        </div>

        {/* Stats Cards */}
        {stats && (
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
                <Clock className='h-4 w-4 text-blue-500' />
                Running
              </div>
              <p className='text-2xl font-bold mt-1'>
                {(stats.by_status?.running || 0) + (stats.by_status?.queued || 0) + (stats.by_status?.pending || 0)}
              </p>
            </div>
            <div className='rounded-lg border p-3'>
              <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                <CheckCircle className='h-4 w-4 text-green-500' />
                Completed
              </div>
              <p className='text-2xl font-bold mt-1'>{stats.by_status?.completed || 0}</p>
            </div>
            <div className='rounded-lg border p-3'>
              <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                <XCircle className='h-4 w-4 text-red-500' />
                Failed
              </div>
              <p className='text-2xl font-bold mt-1'>{stats.by_status?.failed || 0}</p>
            </div>
          </div>
        )}

        {isLoading ? (
          <div className='flex items-center justify-center py-12'>
            <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
          </div>
        ) : (
          <div className='space-y-4'>
            <DataTableToolbar
              table={table}
              searchPlaceholder='Search tasks...'
              filters={[
                {
                  columnId: 'status',
                  title: 'Status',
                  options: [
                    { label: 'Pending', value: 'pending' },
                    { label: 'Queued', value: 'queued' },
                    { label: 'Running', value: 'running' },
                    { label: 'Completed', value: 'completed' },
                    { label: 'Failed', value: 'failed' },
                    { label: 'Cancelled', value: 'cancelled' },
                  ],
                },
              ]}
            />
            <div className='rounded-md border'>
              <Table>
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(
                                header.column.columnDef.header,
                                header.getContext()
                              )}
                        </TableHead>
                      ))}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {table.getRowModel().rows?.length ? (
                    table.getRowModel().rows.map((row) => (
                      <TableRow
                        key={row.id}
                        className='cursor-pointer'
                        onClick={() =>
                          navigate({ to: `/tasks/${row.original.id}` })
                        }
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={tasksColumns.length}
                        className='h-24 text-center'
                      >
                        No tasks found.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
            <DataTablePagination table={table} />
          </div>
        )}
      </Main>
    </>
  )
}
