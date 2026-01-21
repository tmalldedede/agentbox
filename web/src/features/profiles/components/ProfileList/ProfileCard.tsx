import { Copy, Trash2, Lock, MoreVertical, Edit, Cpu, Shield, Code2, Server } from 'lucide-react'
import type { Profile } from '@/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

interface ProfileCardProps {
  profile: Profile
  onClone: () => void
  onDelete: () => void
  onClick: () => void
  isCloning?: boolean
  isDeleting?: boolean
}

const getAdapterIcon = (adapter: string) => {
  switch (adapter) {
    case 'claude-code':
      return <Cpu className="w-5 h-5 text-purple-400" />
    case 'codex':
      return <Shield className="w-5 h-5 text-emerald-400" />
    case 'opencode':
      return <Code2 className="w-5 h-5 text-blue-400" />
    default:
      return <Server className="w-5 h-5 text-gray-400" />
  }
}

const getAdapterBgColor = (adapter: string) => {
  switch (adapter) {
    case 'claude-code':
      return 'bg-purple-500/20'
    case 'codex':
      return 'bg-emerald-500/20'
    case 'opencode':
      return 'bg-blue-500/20'
    default:
      return 'bg-gray-500/20'
  }
}

export function ProfileCard({
  profile,
  onClone,
  onDelete,
  onClick,
  isCloning,
  isDeleting,
}: ProfileCardProps) {
  return (
    <Card
      className="cursor-pointer hover:border-primary/50 transition-colors"
      onClick={onClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-lg ${getAdapterBgColor(profile.adapter)} flex items-center justify-center`}>
              {profile.icon ? (
                <span className="text-xl">{profile.icon}</span>
              ) : (
                getAdapterIcon(profile.adapter)
              )}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{profile.name}</CardTitle>
                {profile.is_built_in && (
                  <Badge variant="secondary" className="text-xs">
                    <Lock className="w-3 h-3 mr-1" />
                    Built-in
                  </Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground font-mono">{profile.id}</p>
            </div>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onClick()
              }}>
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation()
                onClone()
              }} disabled={isCloning}>
                <Copy className="w-4 h-4 mr-2" />
                {isCloning ? 'Cloning...' : 'Clone'}
              </DropdownMenuItem>
              {!profile.is_built_in && (
                <DropdownMenuItem
                  className="text-red-600"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete()
                  }}
                  disabled={isDeleting}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {isDeleting ? 'Deleting...' : 'Delete'}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent>
        {profile.description && (
          <CardDescription className="mb-3 line-clamp-2">
            {profile.description}
          </CardDescription>
        )}
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="outline" className="text-xs capitalize">
            {profile.adapter}
          </Badge>
          <Badge variant="outline" className="text-xs">
            {profile.model.name || 'default'}
          </Badge>
          {profile.permissions.mode && (
            <Badge variant="outline" className="text-xs">
              {profile.permissions.mode}
            </Badge>
          )}
          {profile.permissions.sandbox_mode && (
            <Badge variant="outline" className="text-xs">
              {profile.permissions.sandbox_mode}
            </Badge>
          )}
          {profile.mcp_servers && profile.mcp_servers.length > 0 && (
            <Badge variant="outline" className="text-xs text-amber-600">
              {profile.mcp_servers.length} MCP
            </Badge>
          )}
          {profile.skill_ids && profile.skill_ids.length > 0 && (
            <Badge variant="outline" className="text-xs text-blue-600">
              {profile.skill_ids.length} Skills
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
