import { useRuntimesContext } from './runtimes-provider'
import { RuntimeFormDialog } from './runtime-form-dialog'
import { RuntimeDeleteDialog } from './runtime-delete-dialog'

export function RuntimesDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useRuntimesContext()

  return (
    <>
      <RuntimeFormDialog
        key='runtime-create'
        open={open === 'create'}
        onOpenChange={(isOpen) => {
          if (!isOpen) setOpen(null)
        }}
      />

      {currentRow && (
        <>
          <RuntimeFormDialog
            key={`runtime-edit-${currentRow.id}`}
            open={open === 'edit'}
            onOpenChange={(isOpen) => {
              if (!isOpen) {
                setOpen(null)
                setTimeout(() => setCurrentRow(null), 300)
              }
            }}
            runtime={currentRow}
          />
          <RuntimeDeleteDialog
            key={`runtime-delete-${currentRow.id}`}
            open={open === 'delete'}
            onOpenChange={(isOpen) => {
              if (!isOpen) {
                setOpen(null)
                setTimeout(() => setCurrentRow(null), 300)
              }
            }}
            runtime={currentRow}
          />
        </>
      )}
    </>
  )
}
