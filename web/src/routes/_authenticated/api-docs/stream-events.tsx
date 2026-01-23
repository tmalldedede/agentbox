import { createFileRoute } from '@tanstack/react-router'
import { StreamEventsPage } from '@/features/api-docs/components/StreamEventsPage'

export const Route = createFileRoute('/_authenticated/api-docs/stream-events')({
  component: StreamEventsPage,
})
