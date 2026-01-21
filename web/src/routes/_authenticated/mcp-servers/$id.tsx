import { createFileRoute } from '@tanstack/react-router'
import { MCPServerDetailPage } from '@/features/mcp-servers/detail'

export const Route = createFileRoute('/_authenticated/mcp-servers/$id')({
  component: MCPServerDetail,
})

function MCPServerDetail() {
  const { id } = Route.useParams()
  return <MCPServerDetailPage serverId={id} />
}
