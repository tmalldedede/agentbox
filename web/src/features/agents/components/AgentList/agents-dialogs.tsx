import { useAgentsContext } from './agents-provider'
import { AgentDeleteDialog } from './agent-delete-dialog'

export function AgentsDialogs() {
  const { open, setOpen, currentRow } = useAgentsContext()

  return (
    <>
      {currentRow && (
        <AgentDeleteDialog
          open={open === 'delete'}
          onOpenChange={() => setOpen(null)}
          agent={currentRow}
        />
      )}
    </>
  )
}
