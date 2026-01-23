import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { AgentRuntime } from '@/types'

type RuntimesDialogType = 'create' | 'edit' | 'delete'

type RuntimesContextType = {
  open: RuntimesDialogType | null
  setOpen: (str: RuntimesDialogType | null) => void
  currentRow: AgentRuntime | null
  setCurrentRow: React.Dispatch<React.SetStateAction<AgentRuntime | null>>
}

const RuntimesContext = React.createContext<RuntimesContextType | null>(null)

export function RuntimesProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<RuntimesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<AgentRuntime | null>(null)

  return (
    <RuntimesContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </RuntimesContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useRuntimesContext = () => {
  const ctx = React.useContext(RuntimesContext)
  if (!ctx) {
    throw new Error('useRuntimesContext must be used within <RuntimesProvider>')
  }
  return ctx
}
