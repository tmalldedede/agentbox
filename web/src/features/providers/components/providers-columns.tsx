import { type ColumnDef } from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { CheckCircle2, AlertCircle, KeyRound } from 'lucide-react'
import type { Provider } from '@/types'
import { categories, agents, categoryColorMap, agentColorMap } from '../data/data'
import { getProviderIcon } from '../data/icons'
import { DataTableRowActions } from './data-table-row-actions'

export const providersColumns: ColumnDef<Provider>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { id, name, icon } = row.original
      const iconInfo = getProviderIcon(icon)
      return (
        <div className='flex items-center gap-3'>
          <div className={cn('w-9 h-9 rounded-lg flex items-center justify-center text-sm font-bold shrink-0', iconInfo.color)}>
            {iconInfo.label}
          </div>
          <div className='min-w-0'>
            <div className='font-medium truncate'>{name}</div>
            <div className='text-xs text-muted-foreground font-mono truncate'>{id}</div>
          </div>
        </div>
      )
    },
    enableHiding: false,
  },
  {
    accessorKey: 'category',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Category' />
    ),
    cell: ({ row }) => {
      const category = row.getValue('category') as string
      const info = categories.find((c) => c.value === category)
      return (
        <Badge variant='outline' className={cn('capitalize', categoryColorMap[category])}>
          {info?.label || category}
        </Badge>
      )
    },
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
  },
  {
    accessorKey: 'agents',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Adapters' />
    ),
    cell: ({ row }) => {
      const adapterList = row.getValue('agents') as string[]
      return (
        <div className='flex flex-wrap gap-1'>
          {adapterList?.map((a) => {
            const info = agents.find((ag) => ag.value === a)
            return (
              <Badge key={a} variant='outline' className={cn('capitalize text-xs', agentColorMap[a])}>
                {info?.label || a}
              </Badge>
            )
          })}
        </div>
      )
    },
    filterFn: (row, id, value: string[]) => {
      const rowAgents = row.getValue(id) as string[]
      return value.some((v) => rowAgents?.includes(v))
    },
    enableSorting: false,
  },
  {
    accessorKey: 'default_model',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Default Model' />
    ),
    cell: ({ row }) => {
      const model = row.getValue('default_model') as string
      return (
        <code className='text-sm text-muted-foreground bg-muted px-1.5 py-0.5 rounded'>
          {model || '-'}
        </code>
      )
    },
    enableSorting: false,
  },
  {
    id: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Key Status' />
    ),
    cell: ({ row }) => {
      const { is_configured, is_valid, api_key_masked } = row.original
      if (!is_configured) {
        return (
          <div className='flex items-center gap-1.5 text-amber-600 dark:text-amber-400'>
            <AlertCircle className='h-4 w-4' />
            <span className='text-sm'>Not configured</span>
          </div>
        )
      }
      return (
        <div className='flex items-center gap-1.5'>
          {is_valid ? (
            <CheckCircle2 className='h-4 w-4 text-emerald-500' />
          ) : (
            <KeyRound className='h-4 w-4 text-red-500' />
          )}
          <span className='text-sm font-mono text-muted-foreground'>
            {api_key_masked || (is_valid ? 'Valid' : 'Invalid')}
          </span>
        </div>
      )
    },
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: DataTableRowActions,
  },
]
