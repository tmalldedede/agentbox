import { Cpu } from 'lucide-react'
import type { Profile } from '../../types'
import { Section } from './Section'
import { useLanguage } from '../../contexts/LanguageContext'

interface ResourcesSectionProps {
  profile: Profile
}

export function ResourcesSection({ profile }: ResourcesSectionProps) {
  const { t } = useLanguage()

  return (
    <Section title={t('resourceLimits')} icon={<Cpu className="w-5 h-5" />} defaultOpen={false}>
      <div className="mt-4 grid grid-cols-2 gap-4">
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">{t('maxBudget')}</span>
          <span className="text-lg font-semibold text-primary">
            ${profile.resources.max_budget_usd || t('unlimited')}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">{t('maxTurns')}</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.max_turns || t('unlimited')}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">{t('cpus')}</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.cpus || t('default')}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">{t('memoryMB')}</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.memory_mb || t('default')}
          </span>
        </div>
      </div>
    </Section>
  )
}
