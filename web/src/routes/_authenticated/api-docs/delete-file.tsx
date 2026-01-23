import { createFileRoute } from '@tanstack/react-router'
import { DeleteFilePage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/delete-file')({
  component: DeleteFilePage,
})
