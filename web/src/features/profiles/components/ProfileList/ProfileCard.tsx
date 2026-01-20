import { Copy, Trash2, ChevronRight, Lock } from 'lucide-react'
import type { Profile } from '@/types'

interface ProfileCardProps {
  profile: Profile
  onClone: () => void
  onDelete: () => void
  onClick: () => void
  isCloning?: boolean
  isDeleting?: boolean
}

const adapterColors: Record<string, string> = {
  'claude-code': 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  codex: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  opencode: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
}

export function ProfileCard({
  profile,
  onClone,
  onDelete,
  onClick,
  isCloning,
  isDeleting,
}: ProfileCardProps) {
  const colors = adapterColors[profile.adapter] || 'bg-blue-500/20 text-blue-400 border-blue-500/30'
  const initials = profile.adapter.slice(0, 2).toUpperCase()

  return (
    <div
      className="card p-4 cursor-pointer group hover:border-emerald-500/50 transition-colors"
      onClick={onClick}
    >
      <div className="flex items-start gap-4">
        {/* Avatar */}
        <div
          className={`w-12 h-12 rounded-xl flex items-center justify-center text-sm font-bold ${colors}`}
        >
          {initials}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{profile.name}</span>
            {profile.is_built_in && (
              <span className="badge badge-scaling text-xs">
                <Lock className="w-3 h-3" />
                Built-in
              </span>
            )}
          </div>
          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
            {profile.description || `${profile.adapter} profile`}
          </p>

          {/* Tags */}
          <div className="flex items-center gap-2 mt-3 flex-wrap">
            <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/80">
              {profile.model.name || 'default'}
            </span>
            {profile.permissions.mode && (
              <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/80">
                {profile.permissions.mode}
              </span>
            )}
            {profile.permissions.sandbox_mode && (
              <span className="text-xs px-2 py-0.5 rounded bg-muted text-foreground/80">
                {profile.permissions.sandbox_mode}
              </span>
            )}
            {profile.mcp_servers && profile.mcp_servers.length > 0 && (
              <span className="text-xs px-2 py-0.5 rounded bg-amber-500/20 text-amber-600 dark:text-amber-400">
                {profile.mcp_servers.length} MCP
              </span>
            )}
          </div>
        </div>

        {/* Actions */}
        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          <button
            onClick={onClone}
            className="btn btn-ghost btn-icon"
            title="Clone"
            disabled={isCloning}
          >
            <Copy className={`w-4 h-4 ${isCloning ? 'animate-pulse' : ''}`} />
          </button>
          {!profile.is_built_in && (
            <button
              onClick={onDelete}
              className="btn btn-ghost btn-icon text-red-400"
              title="Delete"
              disabled={isDeleting}
            >
              <Trash2 className={`w-4 h-4 ${isDeleting ? 'animate-pulse' : ''}`} />
            </button>
          )}
        </div>

        {/* Arrow */}
        <ChevronRight className="w-5 h-5 text-muted-foreground group-hover:text-emerald-400 transition-colors" />
      </div>
    </div>
  )
}
