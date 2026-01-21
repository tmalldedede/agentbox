import { createFileRoute } from '@tanstack/react-router'
import { SkillStore } from '@/features/skill-store'

export const Route = createFileRoute('/_authenticated/skill-store/')({
  component: SkillStore,
})
