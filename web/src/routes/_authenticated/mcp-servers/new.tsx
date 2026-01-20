import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_authenticated/mcp-servers/new')({
  component: MCPServerNew,
})

function MCPServerNew() {
  const navigate = useNavigate()

  useEffect(() => {
    // Redirect to mcp-servers page, new server is created via modal
    navigate({ to: '/mcp-servers' })
  }, [navigate])

  return null
}
