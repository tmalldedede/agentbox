import type { ReactNode } from 'react'
import type { Profile } from '@/types'
import { ProfileCard } from './ProfileCard'

interface ProfileGroupProps {
  title: string
  icon: ReactNode
  iconBgColor: string
  profiles: Profile[]
  onClone: (profile: Profile) => void
  onDelete: (profile: Profile) => void
  onClick: (profile: Profile) => void
  cloningId?: string
  deletingId?: string
}

export function ProfileGroup({
  title,
  icon,
  iconBgColor,
  profiles,
  onClone,
  onDelete,
  onClick,
  cloningId,
  deletingId,
}: ProfileGroupProps) {
  if (profiles.length === 0) return null

  return (
    <div>
      <div className="flex items-center gap-3 mb-4">
        <div className={`w-8 h-8 rounded-lg ${iconBgColor} flex items-center justify-center`}>
          {icon}
        </div>
        <h2 className="text-lg font-semibold text-foreground">{title}</h2>
        <span className="text-sm text-muted-foreground">({profiles.length})</span>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {profiles.map(profile => (
          <ProfileCard
            key={profile.id}
            profile={profile}
            onClone={() => onClone(profile)}
            onDelete={() => onDelete(profile)}
            onClick={() => onClick(profile)}
            isCloning={cloningId === profile.id}
            isDeleting={deletingId === profile.id}
          />
        ))}
      </div>
    </div>
  )
}
