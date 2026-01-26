import { createFileRoute } from '@tanstack/react-router'
import { APIPlayground } from '@/features/api-playground'

interface PlaygroundSearch {
  agent?: string
  prompt?: string
}

export const Route = createFileRoute('/_authenticated/api-playground/')({
  validateSearch: (search: Record<string, unknown>): PlaygroundSearch => ({
    agent: typeof search.agent === 'string' ? search.agent : undefined,
    prompt: typeof search.prompt === 'string' ? search.prompt : undefined,
  }),
  component: PlaygroundRoute,
})

function PlaygroundRoute() {
  const { agent, prompt } = Route.useSearch()
  return <APIPlayground preselectedAgentId={agent} initialPrompt={prompt} />
}
