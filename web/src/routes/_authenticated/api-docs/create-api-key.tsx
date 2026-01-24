import { createFileRoute } from '@tanstack/react-router'
import { CreateApiKeyPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/create-api-key')({
  component: CreateApiKeyPage,
})
