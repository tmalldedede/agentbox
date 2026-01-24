import { createFileRoute } from '@tanstack/react-router'
import { CancelBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/cancel-batch')({
  component: CancelBatchPage,
})
