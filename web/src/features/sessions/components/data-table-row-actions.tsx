import { useNavigate } from '@tanstack/react-router'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Play, Square, Trash2, Eye } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Session } from '@/types'
import { useStartSession, useStopSession } from '@/hooks'
import { useSessionsContext } from './sessions-provider'

type DataTableRowActionsProps = {
  row: Row<Session>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const navigate = useNavigate()
  const { setOpen, setCurrentRow } = useSessionsContext()
  const session = row.original
  const startSession = useStartSession()
  const stopSession = useStopSession()

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
          onClick={(e) => e.stopPropagation()}
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem
          onClick={(e) => {
            e.stopPropagation()
            navigate({ to: `/sessions/${session.id}` })
          }}
        >
          View Detail
          <DropdownMenuShortcut>
            <Eye size={16} />
          </DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {session.status === 'running' ? (
          <DropdownMenuItem
            onClick={(e) => {
              e.stopPropagation()
              stopSession.mutate(session.id)
            }}
            disabled={stopSession.isPending}
          >
            Stop
            <DropdownMenuShortcut>
              <Square size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        ) : (
          <DropdownMenuItem
            onClick={(e) => {
              e.stopPropagation()
              startSession.mutate(session.id)
            }}
            disabled={
              startSession.isPending || session.status === 'creating'
            }
          >
            Start
            <DropdownMenuShortcut>
              <Play size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        )}
        {session.status !== 'running' && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation()
                setCurrentRow(session)
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
