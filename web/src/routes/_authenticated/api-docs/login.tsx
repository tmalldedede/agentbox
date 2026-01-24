import { createFileRoute } from '@tanstack/react-router'
import { LoginPage } from '@/features/api-docs'

export const Route = createFileRoute('/_authenticated/api-docs/login')({
  component: LoginPage,
})
