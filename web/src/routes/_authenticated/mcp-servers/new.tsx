import { createFileRoute } from '@tanstack/react-router'
import { MCPServerNewPage } from '@/features/mcp-servers/new'

export const Route = createFileRoute('/_authenticated/mcp-servers/new')({
  component: MCPServerNewPage,
})
