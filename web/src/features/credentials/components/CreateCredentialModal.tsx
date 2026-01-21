import { useState } from 'react'
import { Key, Loader2, Shield, AlertCircle } from 'lucide-react'
import type { CreateCredentialRequest, CredentialProvider, CredentialScope, CredentialType } from '@/types'
import { api } from '@/services/api'
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
  Alert,
  AlertDescription,
} from '@/components/ui/alert'

interface Props {
  onClose: () => void
  onCreated: () => void
}

export default function CreateCredentialModal({ onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [provider, setProvider] = useState<CredentialProvider>('anthropic')
  const [type, setType] = useState<CredentialType>('api_key')
  const [value, setValue] = useState('')
  const [scope, setScope] = useState<CredentialScope>('global')
  const [envVar, setEnvVar] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Auto-generate ID from name
  const generateId = (name: string) => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '')
  }

  // Auto-suggest env var based on provider
  const suggestEnvVar = (provider: CredentialProvider) => {
    const envVars: Record<CredentialProvider, string> = {
      anthropic: 'ANTHROPIC_API_KEY',
      openai: 'OPENAI_API_KEY',
      github: 'GITHUB_TOKEN',
      custom: '',
    }
    return envVars[provider]
  }

  const handleProviderChange = (newProvider: CredentialProvider) => {
    setProvider(newProvider)
    if (!envVar || envVar === suggestEnvVar(provider)) {
      setEnvVar(suggestEnvVar(newProvider))
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !value.trim()) {
      setError('Name and value are required')
      return
    }

    try {
      setLoading(true)
      setError(null)

      const req: CreateCredentialRequest = {
        id: generateId(name),
        name: name.trim(),
        provider,
        type,
        value: value.trim(),
        scope,
      }

      if (envVar.trim()) {
        req.env_var = envVar.trim()
      }

      await api.createCredential(req)
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create credential')
    } finally {
      setLoading(false)
    }
  }

  const providerOptions: { value: CredentialProvider; label: string; icon: string }[] = [
    { value: 'anthropic', label: 'Anthropic', icon: '🧠' },
    { value: 'openai', label: 'OpenAI', icon: '🤖' },
    { value: 'github', label: 'GitHub', icon: '🐙' },
    { value: 'custom', label: 'Custom', icon: '🔧' },
  ]

  const providerColors: Record<CredentialProvider, string> = {
    anthropic: 'border-orange-500/50 bg-orange-500/10',
    openai: 'border-emerald-500/50 bg-emerald-500/10',
    github: 'border-purple-500/50 bg-purple-500/10',
    custom: 'border-gray-500/50 bg-gray-500/10',
  }

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-amber-500/20 flex items-center justify-center">
              <Key className="w-5 h-5 text-amber-400" />
            </div>
            <div>
              <DialogTitle>New Credential</DialogTitle>
              <DialogDescription>Add an API key or token</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className="space-y-5 py-4">
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="w-4 h-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* Provider Selection */}
            <div className="space-y-3">
              <Label>Provider</Label>
              <div className="grid grid-cols-4 gap-2">
                {providerOptions.map(opt => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => handleProviderChange(opt.value)}
                    className={`p-3 rounded-xl border-2 text-center transition-all ${
                      provider === opt.value
                        ? providerColors[opt.value]
                        : 'border-border hover:border-muted'
                    }`}
                  >
                    <span className="text-2xl block mb-1">{opt.icon}</span>
                    <span className="text-xs font-medium">{opt.label}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="e.g., Production API Key"
                autoFocus
              />
              {name && (
                <p className="text-xs text-muted-foreground">
                  ID: <code className="text-amber-400">{generateId(name)}</code>
                </p>
              )}
            </div>

            {/* Type */}
            <div className="space-y-2">
              <Label htmlFor="type">Type</Label>
              <Select value={type} onValueChange={(v) => setType(v as CredentialType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="api_key">API Key</SelectItem>
                  <SelectItem value="token">Token</SelectItem>
                  <SelectItem value="oauth">OAuth</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Value */}
            <div className="space-y-2">
              <Label htmlFor="value">Value</Label>
              <Input
                id="value"
                type="password"
                value={value}
                onChange={e => setValue(e.target.value)}
                placeholder="sk-..."
                className="font-mono"
              />
              <p className="text-xs text-muted-foreground flex items-center gap-1">
                <Shield className="w-3 h-3" />
                Value will be encrypted with AES-256
              </p>
            </div>

            {/* Scope */}
            <div className="space-y-2">
              <Label htmlFor="scope">Scope</Label>
              <Select value={scope} onValueChange={(v) => setScope(v as CredentialScope)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="global">Global - Available to all sessions</SelectItem>
                  <SelectItem value="profile">Profile - Tied to specific profile</SelectItem>
                  <SelectItem value="session">Session - Single session only</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Env Var */}
            <div className="space-y-2">
              <Label htmlFor="envVar">Environment Variable (Optional)</Label>
              <Input
                id="envVar"
                value={envVar}
                onChange={e => setEnvVar(e.target.value)}
                placeholder="e.g., ANTHROPIC_API_KEY"
                className="font-mono"
              />
              <p className="text-xs text-muted-foreground">
                Injected as this env var in agent sessions
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading || !name.trim() || !value.trim()}
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Key className="w-4 h-4 mr-2" />
                  Create Credential
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
