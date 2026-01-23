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
import { useDeleteRuntime } from '@/hooks/useRuntimes'
import type { AgentRuntime } from '@/types'

type RuntimeDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  runtime: AgentRuntime
}

export function RuntimeDeleteDialog({
  open,
  onOpenChange,
  runtime,
}: RuntimeDeleteDialogProps) {
  const deleteRuntime = useDeleteRuntime()

  const handleDelete = () => {
    deleteRuntime.mutate(runtime.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Runtime</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete runtime &quot;{runtime.name}&quot;?
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
            disabled={deleteRuntime.isPending}
          >
            {deleteRuntime.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Delete
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
