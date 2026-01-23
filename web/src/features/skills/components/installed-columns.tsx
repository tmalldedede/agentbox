import { type ColumnDef } from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { Lock, Power, PowerOff, TrendingUp } from 'lucide-react'
import type { Skill } from '@/types'
import { categoryOptions, categoryColorMap, getSkillStats } from '../data/data'
import { InstalledRowActions } from './installed-row-actions'

export const installedColumns: ColumnDef<Skill>[] = [
  {
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name, command, is_built_in } = row.original
      return (
        <div>
          <div className='flex items-center gap-2'>
            <span className='font-medium'>{name}</span>
            {is_built_in && (
              <Badge variant='secondary' className='text-xs gap-1'>
                <Lock className='h-3 w-3' />
                Built-in
              </Badge>
            )}
          </div>
          <code className='text-xs text-emerald-600 dark:text-emerald-400 font-mono'>
            /{command}
          </code>
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
      const info = categoryOptions.find((c) => c.value === category)
      return (
        <Badge variant='outline' className={cn('capitalize text-xs', categoryColorMap[category as keyof typeof categoryColorMap])}>
          {info?.label || category}
        </Badge>
      )
    },
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
  },
  {
    id: 'usage',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Usage' />
    ),
    cell: ({ row }) => {
      const stats = getSkillStats(row.original)
      return (
        <div className='flex items-center gap-2'>
          <div className='flex flex-col'>
            <span className='text-sm font-semibold tabular-nums'>{stats.usageCount}</span>
            <span className='text-[10px] text-muted-foreground flex items-center gap-0.5'>
              <TrendingUp className='h-2.5 w-2.5 text-emerald-500' />
              {stats.lastUsed}
            </span>
          </div>
        </div>
      )
    },
    sortingFn: (a, b) => {
      const sa = getSkillStats(a.original)
      const sb = getSkillStats(b.original)
      return sa.usageCount - sb.usageCount
    },
  },
  {
    id: 'successRate',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Success' />
    ),
    cell: ({ row }) => {
      const stats = getSkillStats(row.original)
      const color = stats.successRate >= 95
        ? 'text-emerald-600 dark:text-emerald-400'
        : stats.successRate >= 90
        ? 'text-amber-600 dark:text-amber-400'
        : 'text-red-600 dark:text-red-400'
      return (
        <div className='flex items-center gap-2'>
          <div className='h-1.5 w-16 overflow-hidden rounded-full bg-muted'>
            <div
              className={cn('h-full rounded-full', stats.successRate >= 95 ? 'bg-emerald-500' : stats.successRate >= 90 ? 'bg-amber-500' : 'bg-red-500')}
              style={{ width: `${stats.successRate}%` }}
            />
          </div>
          <span className={cn('text-xs font-medium tabular-nums', color)}>
            {stats.successRate}%
          </span>
        </div>
      )
    },
    sortingFn: (a, b) => {
      const sa = getSkillStats(a.original)
      const sb = getSkillStats(b.original)
      return sa.successRate - sb.successRate
    },
  },
  {
    id: 'status',
    accessorFn: (row) => row.is_enabled ? 'enabled' : 'disabled',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const enabled = row.original.is_enabled
      return enabled ? (
        <Badge variant='outline' className='text-xs gap-1 bg-green-500/10 text-green-600 border-green-200'>
          <Power className='h-3 w-3' />
          Enabled
        </Badge>
      ) : (
        <Badge variant='outline' className='text-xs gap-1'>
          <PowerOff className='h-3 w-3' />
          Disabled
        </Badge>
      )
    },
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: InstalledRowActions,
  },
]
