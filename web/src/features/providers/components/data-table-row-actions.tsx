import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Trash2, ExternalLink, Pencil, KeyRound } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Provider } from '@/types'
import { useProvidersContext } from './providers-provider'

type DataTableRowActionsProps = {
  row: Row<Provider>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const { setOpen, setCurrentRow } = useProvidersContext()
  const provider = row.original

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
      <DropdownMenuContent align='end' className='w-[180px]'>
        {/* API Key actions */}
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(provider)
            setOpen('manage-keys')
          }}
        >
          Manage Keys
          <DropdownMenuShortcut>
            <KeyRound size={16} />
          </DropdownMenuShortcut>
        </DropdownMenuItem>

        {/* Edit / Delete (only for non-built-in) */}
        {!provider.is_built_in && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(provider)
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

        {/* Danger zone */}
        <DropdownMenuSeparator />
        {provider.is_configured && (
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(provider)
              setOpen('delete-key')
            }}
            className='text-red-500!'
          >
            Remove Key
            <DropdownMenuShortcut>
              <Trash2 size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        )}
        {!provider.is_built_in && (
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(provider)
              setOpen('delete')
            }}
            className='text-red-500!'
          >
            Delete Provider
            <DropdownMenuShortcut>
              <Trash2 size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        )}

        {/* External links */}
        {provider.api_key_url && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <a
                href={provider.api_key_url}
                target='_blank'
                rel='noopener noreferrer'
              >
                Get API Key
                <DropdownMenuShortcut>
                  <ExternalLink size={16} />
                </DropdownMenuShortcut>
              </a>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
