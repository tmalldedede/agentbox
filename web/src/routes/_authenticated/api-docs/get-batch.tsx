import { createFileRoute } from '@tanstack/react-router'
import { GetBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/get-batch')({
  component: GetBatchPage,
})
