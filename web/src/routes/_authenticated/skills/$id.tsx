import { createFileRoute } from '@tanstack/react-router'
import { SkillDetailPage } from '@/features/skills/detail'

export const Route = createFileRoute('/_authenticated/skills/$id')({
  component: SkillDetail,
})

function SkillDetail() {
  const { id } = Route.useParams()
  return <SkillDetailPage skillId={id} />
}
