import { createFileRoute } from '@tanstack/react-router'
import { StartBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/start-batch')({
  component: StartBatchPage,
})
