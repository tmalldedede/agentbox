import { createFileRoute } from '@tanstack/react-router'
import { Webhooks } from '@/features/webhooks'

export const Route = createFileRoute('/_authenticated/webhooks/')({
  component: Webhooks,
})
