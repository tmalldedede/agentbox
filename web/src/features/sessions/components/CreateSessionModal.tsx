import { useState } from 'react'
import { Terminal, Loader2, ChevronDown } from 'lucide-react'
import type { Agent, CreateSessionRequest } from '@/types'
import { useLanguage } from '@/contexts/LanguageContext'
import { useCreateSession } from '@/hooks'
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
  defaultAgentId?: string
  onClose: () => void
  onCreated: () => void
}

export default function CreateSessionModal({
  agents,
  defaultAgentId,
  onClose,
  onCreated,
}: Props) {
  const { t } = useLanguage()
  const [agentId, setAgentId] = useState(defaultAgentId || agents[0]?.id || '')
  const [workspace, setWorkspace] = useState('/tmp/myproject')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [envKey, setEnvKey] = useState('')
  const [envValue, setEnvValue] = useState('')

  const selectedAgent = agents.find(a => a.id === agentId)

  const createSession = useCreateSession()

  const adapterColors: Record<string, string> = {
    'claude-code': 'bg-purple-500/20 text-purple-400 border-purple-500/50',
    codex: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/50',
    opencode: 'bg-blue-500/20 text-blue-400 border-blue-500/50',
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const req: CreateSessionRequest = {
      agent_id: agentId,
      workspace,
    }

    // Build env overrides
    const env: Record<string, string> = {}
    if (envKey.trim() && envValue.trim()) {
      env[envKey.trim()] = envValue.trim()
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
              <DialogDescription>Select an agent and workspace to start a session</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="space-y-5 py-4">
            {createSession.error && (
              <InlineError error={createSession.error} />
            )}

            {/* Agent Selection */}
            <div className="space-y-2">
              <Label>Agent</Label>
              <Select value={agentId} onValueChange={setAgentId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select an agent" />
                </SelectTrigger>
                <SelectContent>
                  {agents.map(a => (
                    <SelectItem key={a.id} value={a.id}>
                      <div className="flex items-center gap-2">
                        {a.icon && <span>{a.icon}</span>}
                        <span>{a.name}</span>
                        <span className="text-xs text-muted-foreground">({a.adapter})</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {selectedAgent && (
                <p className="text-xs text-muted-foreground">
                  <span className={`inline-block px-1.5 py-0.5 rounded text-xs ${adapterColors[selectedAgent.adapter] || ''}`}>
                    {selectedAgent.adapter}
                  </span>
                  {selectedAgent.description && (
                    <span className="ml-2">{selectedAgent.description}</span>
                  )}
                </p>
              )}
            </div>

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
                  {/* Extra Environment Variable */}
                  <div className="space-y-2">
                    <Label>Extra Environment Variable (Optional)</Label>
                    <div className="flex gap-2">
                      <Input
                        value={envKey}
                        onChange={e => setEnvKey(e.target.value)}
                        placeholder="KEY"
                        className="flex-1 font-mono text-sm"
                      />
                      <Input
                        value={envValue}
                        onChange={e => setEnvValue(e.target.value)}
                        placeholder="value"
                        type="password"
                        className="flex-1 font-mono text-sm"
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Override or add environment variables for this session
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
              disabled={createSession.isPending || !agentId || !workspace}
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
