import { createFileRoute } from '@tanstack/react-router'
import { GetFilePage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/get-file')({
  component: GetFilePage,
})
