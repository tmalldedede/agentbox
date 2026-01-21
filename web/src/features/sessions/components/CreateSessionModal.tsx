import { useState } from 'react'
import { Terminal, Loader2, ChevronRight, Layers, Globe, ChevronDown } from 'lucide-react'
import type { Agent, CreateSessionRequest, Profile } from '@/types'
import { useLanguage } from '@/contexts/LanguageContext'
import { useProfiles, useCreateSession } from '@/hooks'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { InlineError } from '@/components/ErrorAlert'

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
    'claude-code': 'bg-purple-500/20 text-purple-400 border-purple-500/50',
    codex: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/50',
    opencode: 'bg-blue-500/20 text-blue-400 border-blue-500/50',
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
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <Terminal className="w-5 h-5 text-emerald-400" />
            </div>
            <div>
              <DialogTitle>{t('createSession')}</DialogTitle>
              <DialogDescription>{t('tagline')}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="space-y-5 py-4">
            {createSession.error && (
              <InlineError error={createSession.error} />
            )}

            {/* Agent Selection */}
            <div className="space-y-3">
              <Label>{t('selectAgent')}</Label>
              <div className="grid grid-cols-2 gap-3">
                {agents.map(a => {
                  const isSelected = agent === a.name
                  const colors =
                    agentColors[a.name] || 'bg-blue-500/20 text-blue-400 border-blue-500/50'
                  const initials = a.name.slice(0, 2).toUpperCase()

                  return (
                    <button
                      key={a.name}
                      type="button"
                      onClick={() => setAgent(a.name)}
                      className={`relative p-4 rounded-xl border-2 text-left transition-all ${
                        isSelected ? colors : 'border-border hover:border-muted-foreground/50'
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        <div
                          className={`w-10 h-10 rounded-lg flex items-center justify-center text-xs font-bold ${colors.split(' ').slice(0, 2).join(' ')}`}
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
                          <ChevronRight className="w-3 h-3 text-background" />
                        </div>
                      )}
                    </button>
                  )
                })}
              </div>
            </div>

            {/* Profile Selection */}
            {filteredProfiles.length > 0 && (
              <div className="space-y-2">
                <Label>Profile (Optional)</Label>
                <Select value={profileId} onValueChange={setProfileId}>
                  <SelectTrigger>
                    <SelectValue placeholder="No profile (use defaults)" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">No profile (use defaults)</SelectItem>
                    {filteredProfiles.map(p => (
                      <SelectItem key={p.id} value={p.id}>
                        {p.name} {p.is_built_in ? '(Built-in)' : ''}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  <Layers className="w-3 h-3 inline mr-1" />
                  Profiles provide pre-configured settings for the agent
                </p>
              </div>
            )}

            {/* Workspace */}
            <div className="space-y-2">
              <Label htmlFor="workspace">{t('workspacePath')}</Label>
              <Input
                id="workspace"
                type="text"
                value={workspace}
                onChange={e => setWorkspace(e.target.value)}
                placeholder={t('workspacePathPlaceholder')}
              />
              <p className="text-xs text-muted-foreground">Path to mount in container</p>
            </div>

            {/* API Key */}
            {envKey && (
              <div className="space-y-2">
                <Label htmlFor="apiKey">{envKey}</Label>
                <Input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={e => setApiKey(e.target.value)}
                  placeholder={t('apiKeyPlaceholder')}
                />
                <p className="text-xs text-muted-foreground">
                  Required for {selectedAgent?.display_name}
                </p>
              </div>
            )}

            {/* Advanced Options */}
            <Collapsible open={showAdvanced} onOpenChange={setShowAdvanced}>
              <div className="border-t border-border pt-4">
                <CollapsibleTrigger asChild>
                  <Button variant="ghost" size="sm" className="p-0 h-auto hover:bg-transparent">
                    <ChevronDown
                      className={`w-4 h-4 mr-2 transition-transform ${showAdvanced ? 'rotate-180' : ''}`}
                    />
                    Advanced Options
                  </Button>
                </CollapsibleTrigger>

                <CollapsibleContent className="mt-4 space-y-4">
                  {/* Base URL */}
                  <div className="space-y-2">
                    <Label htmlFor="baseUrl">
                      <Globe className="w-4 h-4 inline mr-1.5" />
                      API Base URL
                      <span className="text-muted-foreground font-normal ml-2">(Optional)</span>
                    </Label>
                    <Input
                      id="baseUrl"
                      type="text"
                      value={baseUrl}
                      onChange={e => setBaseUrl(e.target.value)}
                      placeholder={
                        agent === 'claude-code'
                          ? 'https://api.anthropic.com'
                          : 'https://api.openai.com/v1'
                      }
                      className="font-mono text-sm"
                    />
                    <p className="text-xs text-muted-foreground">
                      Custom API endpoint for proxies or compatible APIs. Leave empty for default.
                    </p>
                  </div>
                </CollapsibleContent>
              </div>
            </Collapsible>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              {t('cancel')}
            </Button>
            <Button
              type="submit"
              disabled={createSession.isPending || !agent || !workspace}
            >
              {createSession.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  {t('creating_')}
                </>
              ) : (
                <>
                  <Terminal className="w-4 h-4 mr-2" />
                  {t('create')}
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
