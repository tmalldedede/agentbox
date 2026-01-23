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
import { useDeleteProvider } from '@/hooks/useProviders'
import type { Provider } from '@/types'

type DeleteProviderDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  provider: Provider
}

export function DeleteProviderDialog({
  open,
  onOpenChange,
  provider,
}: DeleteProviderDialogProps) {
  const deleteProvider = useDeleteProvider()

  const handleDelete = () => {
    deleteProvider.mutate(provider.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Provider</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete{' '}
            <span className='font-semibold'>{provider.name}</span>? This action
            cannot be undone. Agents referencing this provider will need to be
            updated.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            disabled={deleteProvider.isPending}
          >
            {deleteProvider.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
