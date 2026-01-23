import { createFileRoute } from '@tanstack/react-router'
import { DownloadFilePage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/download-file')({
  component: DownloadFilePage,
})
