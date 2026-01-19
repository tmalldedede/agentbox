import { Cpu } from 'lucide-react'
import type { Profile } from '../../types'
import { Section } from './Section'

interface ResourcesSectionProps {
  profile: Profile
}

export function ResourcesSection({ profile }: ResourcesSectionProps) {
  return (
    <Section title="Resource Limits" icon={<Cpu className="w-5 h-5" />} defaultOpen={false}>
      <div className="mt-4 grid grid-cols-2 gap-4">
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">Max Budget (USD)</span>
          <span className="text-lg font-semibold text-primary">
            ${profile.resources.max_budget_usd || 'unlimited'}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">Max Turns</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.max_turns || 'unlimited'}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">CPUs</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.cpus || 'default'}
          </span>
        </div>
        <div className="p-3 bg-secondary rounded-lg">
          <span className="text-sm text-muted block">Memory (MB)</span>
          <span className="text-lg font-semibold text-primary">
            {profile.resources.memory_mb || 'default'}
          </span>
        </div>
      </div>
    </Section>
  )
}
