import { useState } from 'react'
import { X, Terminal, Loader2, ChevronRight, Layers, Globe, ChevronDown } from 'lucide-react'
import type { Agent, CreateSessionRequest, Profile } from '@/types'
import { useLanguage } from '@/contexts/LanguageContext'
import { useProfiles, useCreateSession } from '@/hooks'

interface Props {
  agents: Agent[]
  profiles?: Profile[]
  defaultAgent?: string
  onClose: () => void
  onCreated: () => void
}

export default function CreateSessionModal({
  agents,
  profiles: propProfiles,
  defaultAgent,
  onClose,
  onCreated,
}: Props) {
  const { t } = useLanguage()
  const [agent, setAgent] = useState(defaultAgent || agents[0]?.name || '')
  const [profileId, setProfileId] = useState<string>('')
  const [workspace, setWorkspace] = useState('/tmp/myproject')
  const [apiKey, setApiKey] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)

  const selectedAgent = agents.find(a => a.name === agent)
  const envKey = selectedAgent?.required_env[0] || ''

  // React Query hooks
  const { data: queryProfiles = [] } = useProfiles()
  const createSession = useCreateSession()

  // Use prop profiles if provided, otherwise use query profiles
  const profiles = propProfiles || queryProfiles

  // Filter profiles by selected agent adapter
  const filteredProfiles = profiles.filter(p => {
    if (agent === 'claude-code') return p.adapter === 'claude-code'
    if (agent === 'codex') return p.adapter === 'codex'
    if (agent === 'opencode') return p.adapter === 'opencode'
    return true
  })

  const agentColors: Record<string, string> = {
    'claude-code': 'bg-purple-500/20 text-purple-400 border-purple-500/30',
    codex: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    opencode: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const req: CreateSessionRequest = {
      agent,
      workspace,
    }

    if (profileId) {
      req.profile_id = profileId
    }

    // Build env with API key and optional base URL
    const env: Record<string, string> = {}
    if (apiKey && envKey) {
      env[envKey] = apiKey
    }
    if (baseUrl.trim()) {
      if (agent === 'claude-code') {
        env['ANTHROPIC_BASE_URL'] = baseUrl.trim()
      } else if (agent === 'codex') {
        env['OPENAI_BASE_URL'] = baseUrl.trim()
      }
    }
    if (Object.keys(env).length > 0) {
      req.env = env
    }

    createSession.mutate(req, {
      onSuccess: () => onCreated(),
    })
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal max-w-xl" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <Terminal className="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">{t('createSession')}</h2>
              <p className="text-sm text-muted-foreground">{t('tagline')}</p>
            </div>
          </div>
          <button onClick={onClose} className="btn btn-ghost btn-icon">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit}>
          <div className="modal-body space-y-5">
            {createSession.error && (
              <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
                {createSession.error instanceof Error
                  ? createSession.error.message
                  : 'Failed to create session'}
              </div>
            )}

            {/* Agent Selection */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-3">
                {t('selectAgent')}
              </label>
              <div className="grid grid-cols-2 gap-3">
                {agents.map(a => {
                  const isSelected = agent === a.name
                  const colors =
                    agentColors[a.name] || 'bg-blue-500/20 text-blue-400 border-blue-500/30'
                  const initials = a.name.slice(0, 2).toUpperCase()

                  return (
                    <button
                      key={a.name}
                      type="button"
                      onClick={() => setAgent(a.name)}
                      className={`relative p-4 rounded-xl border-2 text-left transition-all ${
                        isSelected ? `${colors} border-current` : 'border-default'
                      }`}
                      style={!isSelected ? { borderColor: 'var(--border-color)' } : undefined}
                    >
                      <div className="flex items-start gap-3">
                        <div
                          className={`agent-avatar ${colors.split(' ').slice(0, 2).join(' ')}`}
                        >
                          {initials}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p
                            className={`font-medium ${isSelected ? 'text-foreground' : 'text-foreground/90'}`}
                          >
                            {a.display_name}
                          </p>
                          <p className="text-xs text-muted-foreground mt-0.5 truncate">{a.description}</p>
                        </div>
                      </div>
                      {isSelected && (
                        <div className="absolute top-2 right-2 w-5 h-5 bg-current rounded-full flex items-center justify-center">
                          <ChevronRight className="w-3 h-3" style={{ color: 'var(--bg-card)' }} />
                        </div>
                      )}
                    </button>
                  )
                })}
              </div>
            </div>

            {/* Profile Selection */}
            {filteredProfiles.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-foreground/90 mb-2">
                  Profile (Optional)
                </label>
                <select
                  value={profileId}
                  onChange={e => setProfileId(e.target.value)}
                  className="input"
                >
                  <option value="">No profile (use defaults)</option>
                  {filteredProfiles.map(p => (
                    <option key={p.id} value={p.id}>
                      {p.name} {p.is_built_in ? '(Built-in)' : ''}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted-foreground mt-1.5">
                  <Layers className="w-3 h-3 inline mr-1" />
                  Profiles provide pre-configured settings for the agent
                </p>
              </div>
            )}

            {/* Workspace */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                {t('workspacePath')}
              </label>
              <input
                type="text"
                value={workspace}
                onChange={e => setWorkspace(e.target.value)}
                placeholder={t('workspacePathPlaceholder')}
                className="input"
              />
              <p className="text-xs text-muted-foreground mt-1.5">Path to mount in container</p>
            </div>

            {/* API Key */}
            {envKey && (
              <div>
                <label className="block text-sm font-medium text-foreground/90 mb-2">{envKey}</label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={e => setApiKey(e.target.value)}
                  placeholder={t('apiKeyPlaceholder')}
                  className="input"
                />
                <p className="text-xs text-muted-foreground mt-1.5">
                  Required for {selectedAgent?.display_name}
                </p>
              </div>
            )}

            {/* Advanced Options */}
            <div className="border-t border-border pt-4">
              <button
                type="button"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground/90 transition-colors"
              >
                <ChevronDown
                  className={`w-4 h-4 transition-transform ${showAdvanced ? 'rotate-180' : ''}`}
                />
                Advanced Options
              </button>

              {showAdvanced && (
                <div className="mt-4 space-y-4">
                  {/* Base URL */}
                  <div>
                    <label className="block text-sm font-medium text-foreground/90 mb-2">
                      <Globe className="w-4 h-4 inline mr-1.5" />
                      API Base URL
                      <span className="text-muted-foreground font-normal ml-2">(Optional)</span>
                    </label>
                    <input
                      type="text"
                      value={baseUrl}
                      onChange={e => setBaseUrl(e.target.value)}
                      placeholder={
                        agent === 'claude-code'
                          ? 'https://api.anthropic.com'
                          : 'https://api.openai.com/v1'
                      }
                      className="input font-mono text-sm"
                    />
                    <p className="text-xs text-muted-foreground mt-1.5">
                      Custom API endpoint for proxies or compatible APIs. Leave empty for default.
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="modal-footer">
            <button type="button" onClick={onClose} className="btn btn-secondary">
              {t('cancel')}
            </button>
            <button
              type="submit"
              disabled={createSession.isPending || !agent || !workspace}
              className="btn btn-primary"
            >
              {createSession.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  {t('creating_')}
                </>
              ) : (
                <>
                  <Terminal className="w-4 h-4" />
                  {t('create')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
