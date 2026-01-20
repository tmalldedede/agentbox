import { createFileRoute } from '@tanstack/react-router'
import { MCPServers } from '@/features/mcp-servers'

export const Route = createFileRoute('/_authenticated/mcp-servers/')({
  component: MCPServers,
})
