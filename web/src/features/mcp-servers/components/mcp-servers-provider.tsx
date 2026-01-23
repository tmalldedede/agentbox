import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { MCPServer } from '@/types'

type MCPServersDialogType = 'delete' | 'clone'

type MCPServersContextType = {
  open: MCPServersDialogType | null
  setOpen: (str: MCPServersDialogType | null) => void
  currentRow: MCPServer | null
  setCurrentRow: React.Dispatch<React.SetStateAction<MCPServer | null>>
}

const MCPServersContext = React.createContext<MCPServersContextType | null>(null)

export function MCPServersProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<MCPServersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<MCPServer | null>(null)

  return (
    <MCPServersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </MCPServersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useMCPServersContext = () => {
  const ctx = React.useContext(MCPServersContext)
  if (!ctx) {
    throw new Error('useMCPServersContext must be used within <MCPServersProvider>')
  }
  return ctx
}
