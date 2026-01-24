import { createFileRoute } from '@tanstack/react-router'
import { ListBatchTasksPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/list-batch-tasks')({
  component: ListBatchTasksPage,
})
