import { createFileRoute } from '@tanstack/react-router'
import { ListWebhooksPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/list-webhooks')({
  component: ListWebhooksPage,
})
