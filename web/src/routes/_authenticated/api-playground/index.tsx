import { createFileRoute } from '@tanstack/react-router'
import { APIPlayground } from '@/features/api-playground'

export const Route = createFileRoute('/_authenticated/api-playground/')({
  component: APIPlayground,
})
