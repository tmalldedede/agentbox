import { useMCPServersContext } from './mcp-servers-provider'
import { MCPServerDeleteDialog } from './mcp-server-delete-dialog'

export function MCPServersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useMCPServersContext()

  const closeDialog = () => {
    setOpen(null)
    setTimeout(() => setCurrentRow(null), 300)
  }

  return (
    <>
      {currentRow && (
        <MCPServerDeleteDialog
          key={`mcp-delete-${currentRow.id}`}
          open={open === 'delete'}
          onOpenChange={(isOpen) => {
            if (!isOpen) closeDialog()
          }}
          server={currentRow}
        />
      )}
    </>
  )
}
