import { Loader2 } from 'lucide-react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useDeleteSession } from '@/hooks'
import type { Session } from '@/types'

type SessionDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  session: Session
}

export function SessionDeleteDialog({
  open,
  onOpenChange,
  session,
}: SessionDeleteDialogProps) {
  const deleteSession = useDeleteSession()

  const handleDelete = () => {
    deleteSession.mutate(session.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Session</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete session{' '}
            <code className='font-mono text-sm'>{session.id.slice(0, 12)}...</code>?
            This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            disabled={deleteSession.isPending}
          >
            {deleteSession.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
