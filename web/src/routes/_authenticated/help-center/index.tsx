import { createFileRoute } from '@tanstack/react-router'
import { DocumentationPage } from '@/features/documentation'

export const Route = createFileRoute('/_authenticated/help-center/')({
  component: DocumentationPage,
})
