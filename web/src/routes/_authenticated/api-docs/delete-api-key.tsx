import { createFileRoute } from '@tanstack/react-router'
import { DeleteApiKeyPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/delete-api-key')({
  component: DeleteApiKeyPage,
})
