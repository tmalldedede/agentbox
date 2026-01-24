import { createFileRoute } from '@tanstack/react-router'
import { BatchList } from '@/features/batches'

export const Route = createFileRoute('/_authenticated/batches/')({
  component: BatchList,
})
