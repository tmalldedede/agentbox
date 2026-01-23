import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { Agent } from '@/types'

type AgentsDialogType = 'delete'

type AgentsContextType = {
  open: AgentsDialogType | null
  setOpen: (str: AgentsDialogType | null) => void
  currentRow: Agent | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Agent | null>>
}

const AgentsContext = React.createContext<AgentsContextType | null>(null)

export function AgentsProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<AgentsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Agent | null>(null)

  return (
    <AgentsContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </AgentsContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useAgentsContext = () => {
  const ctx = React.useContext(AgentsContext)
  if (!ctx) {
    throw new Error('useAgentsContext must be used within <AgentsProvider>')
  }
  return ctx
}
