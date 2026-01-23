import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useDeleteMCPServer } from '@/hooks'
import type { MCPServer } from '@/types'

type MCPServerDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  server: MCPServer
}

export function MCPServerDeleteDialog({
  open,
  onOpenChange,
  server,
}: MCPServerDeleteDialogProps) {
  const deleteMCPServer = useDeleteMCPServer()

  const handleDelete = () => {
    deleteMCPServer.mutate(server.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete MCP Server</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete MCP server &quot;{server.name}&quot;?
            This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant='destructive'
            onClick={handleDelete}
            disabled={deleteMCPServer.isPending}
          >
            {deleteMCPServer.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Delete
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
