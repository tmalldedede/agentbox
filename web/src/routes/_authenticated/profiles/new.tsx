import { createFileRoute } from '@tanstack/react-router'
import { ProfileDetail } from '@/features/profiles'

export const Route = createFileRoute('/_authenticated/profiles/new')({
  component: ProfileDetail,
})
