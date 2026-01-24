import { createFileRoute } from '@tanstack/react-router'
import { DeleteWebhookPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/delete-webhook')({
  component: DeleteWebhookPage,
})
