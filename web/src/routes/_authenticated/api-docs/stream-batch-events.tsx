import { createFileRoute } from '@tanstack/react-router'
import { StreamBatchEventsPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/stream-batch-events')({
  component: StreamBatchEventsPage,
})
