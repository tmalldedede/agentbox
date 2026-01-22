import { useParams } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { ProfileForm } from './components/ProfileForm/index'
import { api } from '@/services/api'
import type { Profile } from '@/types'
import { Loader2 } from 'lucide-react'

export function ProfileEdit() {
  const { id } = useParams({ strict: false })
  const [profile, setProfile] = useState<Profile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (id) {
      loadProfile(id)
    }
  }, [id])

  const loadProfile = async (profileId: string) => {
    try {
      const profile = await api.getProfile(profileId)
      setProfile(profile)
    } catch (err: unknown) {
      const error = err as { message?: string }
      setError(error.message || 'Failed to load profile')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-destructive">{error}</p>
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Profile not found</p>
      </div>
    )
  }

  return <ProfileForm profile={profile} />
}

export default ProfileEdit
