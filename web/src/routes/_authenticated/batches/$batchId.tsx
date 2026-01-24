import { createFileRoute } from '@tanstack/react-router'
import { BatchDetail } from '@/features/batches'

export const Route = createFileRoute('/_authenticated/batches/$batchId')({
  component: () => {
    const { batchId } = Route.useParams()
    return <BatchDetail batchId={batchId} />
  },
})
