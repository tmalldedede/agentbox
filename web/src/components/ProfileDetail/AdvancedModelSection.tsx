import { Cpu } from 'lucide-react'
import { useLanguage } from '../../contexts/LanguageContext'
import { Section } from './Section'

interface AdvancedModelSectionProps {
  haikuModel: string
  setHaikuModel: (v: string) => void
  sonnetModel: string
  setSonnetModel: (v: string) => void
  opusModel: string
  setOpusModel: (v: string) => void
  timeoutMs: number | undefined
  setTimeoutMs: (v: number | undefined) => void
  maxOutputTokens: number | undefined
  setMaxOutputTokens: (v: number | undefined) => void
  disableTraffic: boolean
  setDisableTraffic: (v: boolean) => void
  disabled?: boolean
}

export function AdvancedModelSection({
  haikuModel,
  setHaikuModel,
  sonnetModel,
  setSonnetModel,
  opusModel,
  setOpusModel,
  timeoutMs,
  setTimeoutMs,
  maxOutputTokens,
  setMaxOutputTokens,
  disableTraffic,
  setDisableTraffic,
  disabled,
}: AdvancedModelSectionProps) {
  const { t } = useLanguage()

  return (
    <Section
      title={t('advancedConfig')}
      icon={<Cpu className="w-5 h-5" />}
      defaultOpen={false}
    >
      <div className="space-y-4 mt-4">
        <p className="text-sm text-muted">
          Configure model tier settings and advanced options for Claude Code. These settings are
          applied via environment variables.
        </p>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">Haiku Model</label>
            <input
              type="text"
              value={haikuModel}
              onChange={e => setHaikuModel(e.target.value)}
              className="input w-full text-sm"
              placeholder="claude-haiku-3-5-20241022"
              disabled={disabled}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">Sonnet Model</label>
            <input
              type="text"
              value={sonnetModel}
              onChange={e => setSonnetModel(e.target.value)}
              className="input w-full text-sm"
              placeholder="claude-sonnet-4-20250514"
              disabled={disabled}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">Opus Model</label>
            <input
              type="text"
              value={opusModel}
              onChange={e => setOpusModel(e.target.value)}
              className="input w-full text-sm"
              placeholder="claude-opus-4-20250514"
              disabled={disabled}
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">API Timeout (ms)</label>
            <input
              type="number"
              value={timeoutMs || ''}
              onChange={e => setTimeoutMs(e.target.value ? parseInt(e.target.value) : undefined)}
              className="input w-full"
              placeholder="120000"
              disabled={disabled}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">
              Max Output Tokens
            </label>
            <input
              type="number"
              value={maxOutputTokens || ''}
              onChange={e =>
                setMaxOutputTokens(e.target.value ? parseInt(e.target.value) : undefined)
              }
              className="input w-full"
              placeholder="16000"
              disabled={disabled}
            />
          </div>
        </div>

        <div className="flex items-center gap-3 p-3 bg-secondary rounded-lg">
          <input
            type="checkbox"
            id="disableTraffic"
            checked={disableTraffic}
            onChange={e => setDisableTraffic(e.target.checked)}
            className="w-4 h-4 rounded border-muted"
            disabled={disabled}
          />
          <label htmlFor="disableTraffic" className="flex-1">
            <span className="font-medium text-primary">Disable Non-essential Traffic</span>
            <span className="block text-xs text-muted">
              Disable telemetry and other non-essential network requests
            </span>
          </label>
        </div>
      </div>
    </Section>
  )
}
