import type { Credential } from '@/types'
import { CheckboxItem } from './CheckboxItem'

interface CredentialSelectorProps {
  selectedIds: string[]
  onChange: (ids: string[]) => void
  availableCredentials: Credential[]
  disabled?: boolean
}

export function CredentialSelector({
  selectedIds,
  onChange,
  availableCredentials,
  disabled,
}: CredentialSelectorProps) {
  const toggleCredential = (id: string) => {
    if (disabled) return
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter(s => s !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  return (
    <div className="space-y-2 mt-3">
      {availableCredentials.length === 0 ? (
        <p className="text-muted-foreground text-sm">No credentials available</p>
      ) : (
        availableCredentials.map(credential => (
          <CheckboxItem
            key={credential.id}
            checked={selectedIds.includes(credential.id)}
            onChange={() => toggleCredential(credential.id)}
            label={credential.name}
            sublabel={credential.env_var || credential.provider}
            badge={credential.provider}
            activeColor="amber"
          />
        ))
      )}
    </div>
  )
}
