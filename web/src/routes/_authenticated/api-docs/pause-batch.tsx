import { createFileRoute } from '@tanstack/react-router'
import { PauseBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/pause-batch')({
  component: PauseBatchPage,
})
