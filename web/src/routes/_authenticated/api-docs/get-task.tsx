import { createFileRoute } from '@tanstack/react-router'
import { GetTaskPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/get-task')({
  component: GetTaskPage,
})
