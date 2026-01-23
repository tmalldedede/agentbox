import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { Provider } from '@/types'

type ProvidersDialogType = 'add' | 'edit' | 'delete' | 'configure' | 'verify' | 'delete-key'

type ProvidersContextType = {
  open: ProvidersDialogType | null
  setOpen: (str: ProvidersDialogType | null) => void
  currentRow: Provider | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Provider | null>>
}

const ProvidersContext = React.createContext<ProvidersContextType | null>(null)

export function ProvidersProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<ProvidersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Provider | null>(null)

  return (
    <ProvidersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </ProvidersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useProvidersContext = () => {
  const ctx = React.useContext(ProvidersContext)
  if (!ctx) {
    throw new Error('useProvidersContext must be used within <ProvidersProvider>')
  }
  return ctx
}
