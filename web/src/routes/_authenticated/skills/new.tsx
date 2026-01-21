import { createFileRoute } from '@tanstack/react-router'
import { SkillNewPage } from '@/features/skills/new'

export const Route = createFileRoute('/_authenticated/skills/new')({
  component: SkillNewPage,
})
