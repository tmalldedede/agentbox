import { Shield } from 'lucide-react'
import type { Profile } from '../../types'
import { Section } from './Section'
import { useLanguage } from '../../contexts/LanguageContext'

interface PermissionsSectionProps {
  profile: Profile
}

export function PermissionsSection({ profile }: PermissionsSectionProps) {
  const { t } = useLanguage()

  return (
    <Section title={t('permissions')} icon={<Shield className="w-5 h-5" />} defaultOpen={false}>
      <div className="mt-4 space-y-3">
        {profile.adapter === 'claude-code' && (
          <>
            <div className="flex items-center justify-between p-3 bg-secondary rounded-lg">
              <span className="text-secondary">{t('mode')}</span>
              <span className="text-primary font-medium">
                {profile.permissions.mode || t('default')}
              </span>
            </div>
            {profile.permissions.allowed_tools && (
              <div className="p-3 bg-secondary rounded-lg">
                <span className="text-secondary block mb-2">{t('allowedTools')}</span>
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
              <span className="text-secondary">{t('sandboxMode')}</span>
              <span className="text-primary font-medium">
                {profile.permissions.sandbox_mode || 'read-only'}
              </span>
            </div>
            <div className="flex items-center justify-between p-3 bg-secondary rounded-lg">
              <span className="text-secondary">{t('approvalPolicy')}</span>
              <span className="text-primary font-medium">
                {profile.permissions.approval_policy || 'on-failure'}
              </span>
            </div>
          </>
        )}
      </div>
    </Section>
  )
}
