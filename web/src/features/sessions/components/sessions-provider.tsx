import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { Session } from '@/types'

type SessionsDialogType = 'create' | 'delete'

type SessionsContextType = {
  open: SessionsDialogType | null
  setOpen: (str: SessionsDialogType | null) => void
  currentRow: Session | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Session | null>>
}

const SessionsContext = React.createContext<SessionsContextType | null>(null)

export function SessionsProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<SessionsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Session | null>(null)

  return (
    <SessionsContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </SessionsContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useSessionsContext = () => {
  const ctx = React.useContext(SessionsContext)
  if (!ctx) {
    throw new Error('useSessionsContext must be used within <SessionsProvider>')
  }
  return ctx
}
