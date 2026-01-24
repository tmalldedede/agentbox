import { createFileRoute } from '@tanstack/react-router'
import { DeleteBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/delete-batch')({
  component: DeleteBatchPage,
})
