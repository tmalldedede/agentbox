import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_authenticated/skills/$id')({
  component: SkillDetail,
})

function SkillDetail() {
  const navigate = useNavigate()
  const { id } = Route.useParams()

  useEffect(() => {
    // Redirect to skills page with skill selected
    navigate({ to: '/skills', search: { selected: id } })
  }, [navigate, id])

  return null
}
