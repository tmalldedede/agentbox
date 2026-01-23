import { useSessionsContext } from './sessions-provider'
import { SessionDeleteDialog } from './session-delete-dialog'

export function SessionsDialogs() {
  const { open, setOpen, currentRow } = useSessionsContext()

  return (
    <>
      {currentRow && (
        <SessionDeleteDialog
          open={open === 'delete'}
          onOpenChange={() => setOpen(null)}
          session={currentRow}
        />
      )}
    </>
  )
}
