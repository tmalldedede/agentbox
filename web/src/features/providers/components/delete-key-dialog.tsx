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
import { useDeleteProviderKey } from '@/hooks/useProviders'
import type { Provider } from '@/types'

type DeleteKeyDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  provider: Provider
}

export function DeleteKeyDialog({
  open,
  onOpenChange,
  provider,
}: DeleteKeyDialogProps) {
  const deleteKey = useDeleteProviderKey()

  const handleDelete = () => {
    deleteKey.mutate(provider.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove API Key</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to remove the API key for{' '}
            <span className='font-semibold'>{provider.name}</span>? Agents using
            this provider will no longer be able to make API calls.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            disabled={deleteKey.isPending}
          >
            {deleteKey.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Remove
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
