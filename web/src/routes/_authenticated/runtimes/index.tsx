import { createFileRoute } from '@tanstack/react-router'
import Runtimes from '@/features/runtimes'

export const Route = createFileRoute('/_authenticated/runtimes/')({
  component: Runtimes,
})
