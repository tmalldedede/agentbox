import { createFileRoute } from '@tanstack/react-router'
import { ListBatchesPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/list-batches')({
  component: ListBatchesPage,
})
