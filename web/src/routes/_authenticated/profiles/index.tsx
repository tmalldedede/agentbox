import { createFileRoute } from '@tanstack/react-router'
import { Profiles } from '@/features/profiles'

export const Route = createFileRoute('/_authenticated/profiles/')({
  component: Profiles,
})
