import { useEffect } from 'react'
import { useVerifyProviderKey } from '@/hooks/useProviders'
import { useProvidersContext } from './providers-provider'
import { ProviderActionDialog } from './provider-action-dialog'
import { DeleteProviderDialog } from './delete-provider-dialog'
import { ConfigureKeyDialog } from './configure-key-dialog'
import { DeleteKeyDialog } from './delete-key-dialog'
import { AuthProfilesDialog } from './auth-profiles-dialog'

export function ProvidersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useProvidersContext()
  const verifyKey = useVerifyProviderKey()

  // Handle verify action immediately (no dialog needed)
  useEffect(() => {
    if (open === 'verify' && currentRow) {
      verifyKey.mutate(currentRow.id)
      setOpen(null)
      setCurrentRow(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, currentRow])

  const closeDialog = () => {
    setOpen(null)
    setTimeout(() => setCurrentRow(null), 300)
  }

  return (
    <>
      {/* Add Provider (no currentRow needed) */}
      <ProviderActionDialog
        key='provider-add'
        open={open === 'add'}
        onOpenChange={(isOpen) => {
          if (!isOpen) closeDialog()
        }}
      />

      {currentRow && (
        <>
          {/* Edit Provider */}
          <ProviderActionDialog
            key={`provider-edit-${currentRow.id}`}
            open={open === 'edit'}
            onOpenChange={(isOpen) => {
              if (!isOpen) closeDialog()
            }}
            currentRow={currentRow}
          />

          {/* Delete Provider */}
          <DeleteProviderDialog
            key={`provider-delete-${currentRow.id}`}
            open={open === 'delete'}
            onOpenChange={(isOpen) => {
              if (!isOpen) closeDialog()
            }}
            provider={currentRow}
          />

          {/* Configure API Key */}
          <ConfigureKeyDialog
            key={`configure-${currentRow.id}`}
            open={open === 'configure'}
            onOpenChange={(isOpen) => {
              if (!isOpen) closeDialog()
            }}
            provider={currentRow}
          />

          {/* Delete API Key */}
          <DeleteKeyDialog
            key={`delete-key-${currentRow.id}`}
            open={open === 'delete-key'}
            onOpenChange={(isOpen) => {
              if (!isOpen) closeDialog()
            }}
            provider={currentRow}
          />

          {/* Manage API Keys (Multi-Key Rotation) */}
          <AuthProfilesDialog
            key={`manage-keys-${currentRow.id}`}
            open={open === 'manage-keys'}
            onOpenChange={(isOpen) => {
              if (!isOpen) closeDialog()
            }}
            provider={currentRow}
          />
        </>
      )}
    </>
  )
}
