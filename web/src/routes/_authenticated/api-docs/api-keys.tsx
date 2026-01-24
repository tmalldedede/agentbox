import { createFileRoute } from '@tanstack/react-router'
import { ApiKeysPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/api-keys')({
  component: ApiKeysPage,
})
