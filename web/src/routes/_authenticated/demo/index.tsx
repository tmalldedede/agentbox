import { createFileRoute } from '@tanstack/react-router'
import { DemoPanel } from '@/features/demo/components/DemoPanel'

export const Route = createFileRoute('/_authenticated/demo/')({
  component: DemoPanel,
})
