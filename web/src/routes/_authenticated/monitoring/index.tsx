import { createFileRoute } from '@tanstack/react-router'
import { ResourceMonitor } from '@/features/monitoring/components/ResourceMonitor'

export const Route = createFileRoute('/_authenticated/monitoring/')({
  component: ResourceMonitor,
})
