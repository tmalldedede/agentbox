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
import { useDeleteAgent } from '@/hooks'
import type { Agent } from '@/types'

type AgentDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  agent: Agent
}

export function AgentDeleteDialog({
  open,
  onOpenChange,
  agent,
}: AgentDeleteDialogProps) {
  const deleteAgent = useDeleteAgent()

  const handleDelete = () => {
    deleteAgent.mutate(agent.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Agent</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete agent{' '}
            <span className='font-semibold'>"{agent.name}"</span>? This action
            cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            disabled={deleteAgent.isPending}
          >
            {deleteAgent.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
