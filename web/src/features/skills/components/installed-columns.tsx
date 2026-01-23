import { type ColumnDef } from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { Lock, Power, PowerOff } from 'lucide-react'
import type { Skill } from '@/types'
import { categoryOptions, categoryColorMap } from '../data/data'
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
    accessorKey: 'tags',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tags' />
    ),
    cell: ({ row }) => {
      const tags = row.original.tags
      if (!tags?.length) return <span className='text-muted-foreground text-xs'>-</span>
      return (
        <div className='flex flex-wrap gap-1'>
          {tags.slice(0, 3).map((tag) => (
            <Badge key={tag} variant='outline' className='text-xs'>{tag}</Badge>
          ))}
          {tags.length > 3 && (
            <Badge variant='outline' className='text-xs'>+{tags.length - 3}</Badge>
          )}
        </div>
      )
    },
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: InstalledRowActions,
  },
]
