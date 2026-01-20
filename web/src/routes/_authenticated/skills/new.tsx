import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/_authenticated/skills/new')({
  component: SkillNew,
})

function SkillNew() {
  const navigate = useNavigate()

  useEffect(() => {
    // Redirect to skills page, new skill is created via modal
    navigate({ to: '/skills' })
  }, [navigate])

  return null
}
