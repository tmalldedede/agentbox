import { createFileRoute } from '@tanstack/react-router'
import CommandCenter from '@/features/command-center'

export const Route = createFileRoute('/_authenticated/')({
  component: CommandCenter,
})
