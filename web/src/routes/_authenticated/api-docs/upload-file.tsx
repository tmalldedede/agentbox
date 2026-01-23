import { createFileRoute } from '@tanstack/react-router'
import { UploadFilePage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/upload-file')({
  component: UploadFilePage,
})
