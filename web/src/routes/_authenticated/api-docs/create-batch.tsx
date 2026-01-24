import { createFileRoute } from '@tanstack/react-router'
import { CreateBatchPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/create-batch')({
  component: CreateBatchPage,
})
