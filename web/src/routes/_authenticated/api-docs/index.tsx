import { createFileRoute } from '@tanstack/react-router'
import { OverviewPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/')({
  component: OverviewPage,
})
