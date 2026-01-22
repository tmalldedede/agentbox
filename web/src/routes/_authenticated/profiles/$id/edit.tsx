import { createFileRoute } from '@tanstack/react-router'
import { ProfileEdit } from '@/features/profiles/edit'

export const Route = createFileRoute('/_authenticated/profiles/$id/edit')({
  component: ProfileEdit,
})
