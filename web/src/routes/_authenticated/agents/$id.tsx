import { createFileRoute } from '@tanstack/react-router'
import { AgentDetail } from '@/features/agents'

export const Route = createFileRoute('/_authenticated/agents/$id')({
  component: AgentDetailRoute,
})

function AgentDetailRoute() {
  const { id } = Route.useParams()
  return <AgentDetail agentId={id} />
}
