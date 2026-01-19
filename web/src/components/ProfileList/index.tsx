import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  Cpu,
  Shield,
  Code2,
  Server,
  Zap,
  RefreshCw,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import type { Profile } from '../../types'
import { useProfiles, useDeleteProfile, useCloneProfile } from '../../hooks'
import { ProfileGroup } from './ProfileGroup'

export default function ProfileList() {
  const navigate = useNavigate()
  const [deletingId, setDeletingId] = useState<string | undefined>()
  const [cloningId, setCloningId] = useState<string | undefined>()

  // React Query hooks
  const {
    data: profiles = [],
    isLoading,
    isFetching,
    error,
    refetch,
  } = useProfiles()

  const deleteProfile = useDeleteProfile()
  const cloneProfile = useCloneProfile()

  const handleClone = (profile: Profile) => {
    const newId = `${profile.id}-copy-${Date.now()}`
    const newName = `${profile.name} (Copy)`
    setCloningId(profile.id)
    cloneProfile.mutate(
      { id: profile.id, request: { new_id: newId, new_name: newName } },
      {
        onSettled: () => setCloningId(undefined),
      }
    )
  }

  const handleDelete = (profile: Profile) => {
    if (!confirm(`Delete profile "${profile.name}"?`)) return
    setDeletingId(profile.id)
    deleteProfile.mutate(profile.id, {
      onSettled: () => setDeletingId(undefined),
    })
  }

  const handleClick = (profile: Profile) => {
    navigate(`/profiles/${profile.id}`)
  }

  // Group profiles by adapter
  const claudeProfiles = profiles.filter(p => p.adapter === 'claude-code')
  const codexProfiles = profiles.filter(p => p.adapter === 'codex')
  const opencodeProfiles = profiles.filter(p => p.adapter === 'opencode')
  const otherProfiles = profiles.filter(
    p => !['claude-code', 'codex', 'opencode'].includes(p.adapter)
  )

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Zap className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Profiles</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="btn btn-ghost btn-icon"
            disabled={isFetching}
          >
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button className="btn btn-primary" onClick={() => navigate('/profiles/new')}>
            <Plus className="w-4 h-4" />
            New Profile
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load profiles'}
            </span>
          </div>
        )}

        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-primary mb-2">Agent Profiles</h1>
          <p className="text-secondary">
            Profiles are pre-configured templates that combine adapter settings, model selection, MCP
            servers, and permissions. Use them to quickly create sessions with your preferred setup.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
          </div>
        ) : profiles.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Server className="w-16 h-16 text-muted mb-4" />
            <p className="text-secondary text-lg">No profiles found</p>
            <p className="text-muted mt-2">Create your first profile to get started</p>
          </div>
        ) : (
          <div className="space-y-8">
            <ProfileGroup
              title="Claude Code"
              icon={<Cpu className="w-4 h-4 text-purple-400" />}
              iconBgColor="bg-purple-500/20"
              profiles={claudeProfiles}
              onClone={handleClone}
              onDelete={handleDelete}
              onClick={handleClick}
              cloningId={cloningId}
              deletingId={deletingId}
            />

            <ProfileGroup
              title="Codex"
              icon={<Shield className="w-4 h-4 text-emerald-400" />}
              iconBgColor="bg-emerald-500/20"
              profiles={codexProfiles}
              onClone={handleClone}
              onDelete={handleDelete}
              onClick={handleClick}
              cloningId={cloningId}
              deletingId={deletingId}
            />

            <ProfileGroup
              title="OpenCode"
              icon={<Code2 className="w-4 h-4 text-blue-400" />}
              iconBgColor="bg-blue-500/20"
              profiles={opencodeProfiles}
              onClone={handleClone}
              onDelete={handleDelete}
              onClick={handleClick}
              cloningId={cloningId}
              deletingId={deletingId}
            />

            <ProfileGroup
              title="Other"
              icon={<Server className="w-4 h-4 text-gray-400" />}
              iconBgColor="bg-gray-500/20"
              profiles={otherProfiles}
              onClone={handleClone}
              onDelete={handleDelete}
              onClick={handleClick}
              cloningId={cloningId}
              deletingId={deletingId}
            />
          </div>
        )}
      </div>
    </div>
  )
}
