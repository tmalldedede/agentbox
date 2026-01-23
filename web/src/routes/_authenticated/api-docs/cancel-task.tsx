import { createFileRoute } from '@tanstack/react-router'
import { CancelTaskPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/cancel-task')({
  component: CancelTaskPage,
})
