import { createFileRoute } from '@tanstack/react-router'
import { QuickStartPage } from '@/features/quick-start'

export const Route = createFileRoute('/_authenticated/')({
  component: QuickStartPage,
})
