import type { Skill } from '../../types'
import { CheckboxItem } from './CheckboxItem'

interface SkillSelectorProps {
  selectedIds: string[]
  onChange: (ids: string[]) => void
  availableSkills: Skill[]
  disabled?: boolean
}

export function SkillSelector({
  selectedIds,
  onChange,
  availableSkills,
  disabled,
}: SkillSelectorProps) {
  const toggleSkill = (id: string) => {
    if (disabled) return
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter(s => s !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  const enabledSkills = availableSkills.filter(s => s.is_enabled)

  return (
    <div className="space-y-2 mt-3">
      {enabledSkills.length === 0 ? (
        <p className="text-muted text-sm">No skills available</p>
      ) : (
        enabledSkills.map(skill => (
          <CheckboxItem
            key={skill.id}
            checked={selectedIds.includes(skill.id)}
            onChange={() => toggleSkill(skill.id)}
            label={skill.name}
            sublabel={skill.command}
            badge={skill.category}
          />
        ))
      )}
    </div>
  )
}
