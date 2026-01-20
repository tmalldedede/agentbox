import { createFileRoute } from '@tanstack/react-router'
import { SessionDetail } from '@/features/sessions'

export const Route = createFileRoute('/_authenticated/sessions/$id')({
  component: SessionDetail,
})
