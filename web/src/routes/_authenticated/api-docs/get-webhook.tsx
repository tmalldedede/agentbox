import { createFileRoute } from '@tanstack/react-router'
import { GetWebhookPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/get-webhook')({
  component: GetWebhookPage,
})
