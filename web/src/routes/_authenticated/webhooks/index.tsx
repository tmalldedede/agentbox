import { createFileRoute } from '@tanstack/react-router'
import { WebhookList } from '@/features/webhooks/components/WebhookList'

export const Route = createFileRoute('/_authenticated/webhooks/')({
  component: WebhookList,
})
