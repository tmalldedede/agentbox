import { createFileRoute } from '@tanstack/react-router'
import { GetTasksPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/get-tasks')({
  component: GetTasksPage,
})
