import { createFileRoute } from '@tanstack/react-router'
import { ListFilesPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/list-files')({
  component: ListFilesPage,
})
