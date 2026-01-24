import { createFileRoute } from '@tanstack/react-router'
import { CreateWebhookPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/create-webhook')({
  component: CreateWebhookPage,
})
