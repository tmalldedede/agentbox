import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Agent } from '@/types'
import { DataTableRowActions } from './data-table-row-actions'

const statusColors: Record<string, string> = {
  active: 'bg-green-500/10 text-green-600 border-green-200',
  inactive: 'bg-gray-500/10 text-gray-600 border-gray-200',
}

const accessLabels: Record<string, string> = {
  public: 'Public',
  api_key: 'API Key',
  private: 'Private',
}

export const agentsColumns: ColumnDef<Agent>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name, id, icon } = row.original
      return (
        <div className='flex items-center gap-3'>
          <div className='flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500/10'>
            {icon ? (
              <span className='text-lg'>{icon}</span>
            ) : (
              <span className='text-xs font-bold text-blue-600'>
                {name.slice(0, 2).toUpperCase()}
              </span>
            )}
          </div>
          <div>
            <div className='font-medium'>{name}</div>
            <div className='text-xs text-muted-foreground font-mono'>{id}</div>
          </div>
        </div>
      )
    },
    enableHiding: false,
  },
  {
    accessorKey: 'adapter',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Adapter' />
    ),
    cell: ({ row }) => (
      <Badge variant='outline' className='text-xs capitalize'>
        {row.getValue('adapter')}
      </Badge>
    ),
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
  },
  {
    accessorKey: 'provider_id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Provider' />
    ),
    cell: ({ row }) => (
      <span className='text-sm'>{row.getValue('provider_id')}</span>
    ),
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
  },
  {
    accessorKey: 'model',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Model' />
    ),
    cell: ({ row }) => {
      const model = row.getValue('model') as string
      return model ? (
        <code className='text-xs bg-muted px-1.5 py-0.5 rounded font-mono'>
          {model}
        </code>
      ) : (
        <span className='text-muted-foreground text-xs'>default</span>
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
    accessorKey: 'api_access',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Access' />
    ),
    cell: ({ row }) => {
      const access = row.getValue('api_access') as string
      return access ? (
        <Badge variant='secondary' className='text-xs'>
          {accessLabels[access] || access}
        </Badge>
      ) : null
    },
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: DataTableRowActions,
  },
]
