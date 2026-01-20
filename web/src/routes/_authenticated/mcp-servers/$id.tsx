import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_authenticated/mcp-servers/$id')({
  component: MCPServerDetail,
})

function MCPServerDetail() {
  const navigate = useNavigate()
  const { id } = Route.useParams()

  useEffect(() => {
    // Redirect to mcp-servers page with server selected
    navigate({ to: '/mcp-servers', search: { selected: id } })
  }, [navigate, id])

  return null
}
