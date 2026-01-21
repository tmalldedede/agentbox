import { createFileRoute } from '@tanstack/react-router'
import { AgentDetail } from '@/features/agents'

export const Route = createFileRoute('/_authenticated/agents/$id')({
  component: AgentDetail,
})
