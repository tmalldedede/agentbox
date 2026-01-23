import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Session } from '@/types'
import { DataTableRowActions } from './data-table-row-actions'

const statusColors: Record<string, string> = {
  running: 'bg-green-500/10 text-green-600 border-green-200',
  stopped: 'bg-gray-500/10 text-gray-600 border-gray-200',
  creating: 'bg-blue-500/10 text-blue-600 border-blue-200',
  error: 'bg-red-500/10 text-red-600 border-red-200',
}

export const sessionsColumns: ColumnDef<Session>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Session ID' />
    ),
    cell: ({ row }) => (
      <code className='text-xs font-mono text-muted-foreground'>
        {row.original.id.slice(0, 12)}...
      </code>
    ),
    enableHiding: false,
  },
  {
    accessorKey: 'agent',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Adapter' />
    ),
    cell: ({ row }) => (
      <Badge variant='outline' className='text-xs capitalize'>
        {row.getValue('agent')}
      </Badge>
    ),
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
  },
  {
    accessorKey: 'agent_id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Agent' />
    ),
    cell: ({ row }) => {
      const agentId = row.getValue('agent_id') as string
      return agentId ? (
        <span className='text-sm'>{agentId}</span>
      ) : (
        <span className='text-xs text-muted-foreground'>-</span>
      )
    },
  },
  {
    accessorKey: 'workspace',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Workspace' />
    ),
    cell: ({ row }) => {
      const workspace = row.getValue('workspace') as string
      return (
        <code className='text-xs font-mono max-w-[200px] truncate block'>
          {workspace}
        </code>
      )
    },
  },
  {
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const status = row.getValue('status') as string
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
    enableSorting: false,
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Created' />
    ),
    cell: ({ row }) => (
      <span className='text-xs text-muted-foreground'>
        {new Date(row.getValue('created_at')).toLocaleString()}
      </span>
    ),
  },
  {
    id: 'actions',
    cell: DataTableRowActions,
  },
]
