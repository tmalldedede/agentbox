import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Pencil, Trash2, Star } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { AgentRuntime } from '@/types'
import { useRuntimesContext } from './runtimes-provider'
import { useSetDefaultRuntime } from '@/hooks/useRuntimes'

type DataTableRowActionsProps = {
  row: Row<AgentRuntime>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const { setOpen, setCurrentRow } = useRuntimesContext()
  const setDefault = useSetDefaultRuntime()
  const runtime = row.original

  const showEdit = !runtime.is_built_in
  const showDelete = !runtime.is_built_in
  const showSetDefault = !runtime.is_default

  if (!showEdit && !showDelete && !showSetDefault) {
    return null
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        {showSetDefault && (
          <DropdownMenuItem
            onClick={() => setDefault.mutate(runtime.id)}
          >
            Set as Default
            <DropdownMenuShortcut>
              <Star size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        )}
        {showEdit && (
          <>
            {showSetDefault && <DropdownMenuSeparator />}
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(runtime)
                setOpen('edit')
              }}
            >
              Edit
              <DropdownMenuShortcut>
                <Pencil size={16} />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          </>
        )}
        {showDelete && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(runtime)
                setOpen('delete')
              }}
              className='text-red-500!'
            >
              Delete
              <DropdownMenuShortcut>
                <Trash2 size={16} />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
