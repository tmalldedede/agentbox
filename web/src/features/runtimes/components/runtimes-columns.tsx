import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { AgentRuntime } from '@/types'
import { DataTableRowActions } from './data-table-row-actions'

export const runtimesColumns: ColumnDef<AgentRuntime>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name, id, is_built_in, is_default } = row.original
      return (
        <div className='flex items-center gap-2'>
          <div>
            <div className='font-medium'>{name}</div>
            <div className='text-xs text-muted-foreground font-mono'>{id}</div>
          </div>
          {is_default && (
            <Badge variant='default' className='text-xs'>Default</Badge>
          )}
          {is_built_in && (
            <Badge variant='secondary' className='text-xs'>Built-in</Badge>
          )}
        </div>
      )
    },
    enableHiding: false,
  },
  {
    accessorKey: 'image',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Image' />
    ),
    cell: ({ row }) => (
      <code className='text-xs bg-muted px-2 py-1 rounded font-mono'>
        {row.getValue('image')}
      </code>
    ),
  },
  {
    accessorKey: 'cpus',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='CPUs' />
    ),
    cell: ({ row }) => (
      <span className='text-sm'>{row.getValue('cpus')}</span>
    ),
  },
  {
    accessorKey: 'memory_mb',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Memory' />
    ),
    cell: ({ row }) => {
      const mb = row.getValue('memory_mb') as number
      return <span className='text-sm'>{mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb} MB`}</span>
    },
  },
  {
    accessorKey: 'privileged',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Privileged' />
    ),
    cell: ({ row }) => {
      const privileged = row.getValue('privileged') as boolean
      return privileged ? (
        <Badge variant='destructive' className='text-xs'>Yes</Badge>
      ) : (
        <span className='text-muted-foreground text-xs'>No</span>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'network',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Network' />
    ),
    cell: ({ row }) => {
      const network = row.getValue('network') as string
      return network ? (
        <Badge variant='outline' className='text-xs'>{network}</Badge>
      ) : (
        <span className='text-muted-foreground text-xs'>-</span>
      )
    },
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: DataTableRowActions,
  },
]
