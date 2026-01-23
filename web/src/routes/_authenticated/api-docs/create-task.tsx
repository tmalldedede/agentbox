import { createFileRoute } from '@tanstack/react-router'
import { CreateTaskPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/create-task')({
  component: CreateTaskPage,
})
