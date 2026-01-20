import { Shield } from 'lucide-react'
import type { Profile } from '@/types'
import { Section } from './Section'

interface PermissionsSectionProps {
  profile: Profile
}

export function PermissionsSection({ profile }: PermissionsSectionProps) {
  return (
    <Section title="Permissions" icon={<Shield className="w-5 h-5" />} defaultOpen={false}>
      <div className="mt-4 space-y-3">
        {profile.adapter === 'claude-code' && (
          <>
            <div className="flex items-center justify-between p-3 bg-secondary rounded-lg">
              <span className="text-muted-foreground">Mode</span>
              <span className="text-foreground font-medium">
                {profile.permissions.mode || 'default'}
              </span>
            </div>
            {profile.permissions.allowed_tools && (
              <div className="p-3 bg-secondary rounded-lg">
                <span className="text-muted-foreground block mb-2">Allowed Tools</span>
                <div className="flex flex-wrap gap-2">
                  {profile.permissions.allowed_tools.map(tool => (
                    <span
                      key={tool}
                      className="text-xs px-2 py-1 rounded bg-emerald-500/20 text-emerald-400"
                    >
                      {tool}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </>
        )}
        {profile.adapter === 'codex' && (
          <>
            <div className="flex items-center justify-between p-3 bg-secondary rounded-lg">
              <span className="text-muted-foreground">Sandbox Mode</span>
              <span className="text-foreground font-medium">
                {profile.permissions.sandbox_mode || 'read-only'}
              </span>
            </div>
            <div className="flex items-center justify-between p-3 bg-secondary rounded-lg">
              <span className="text-muted-foreground">Approval Policy</span>
              <span className="text-foreground font-medium">
                {profile.permissions.approval_policy || 'on-failure'}
              </span>
            </div>
          </>
        )}
      </div>
    </Section>
  )
}
