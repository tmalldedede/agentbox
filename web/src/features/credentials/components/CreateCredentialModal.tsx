import { useState } from 'react'
import { X, Key, Loader2, Shield, AlertCircle } from 'lucide-react'
import type { CreateCredentialRequest, CredentialProvider, CredentialScope, CredentialType } from '@/types'
import { api } from '@/services/api'

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
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal max-w-lg" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-amber-500/20 flex items-center justify-center">
              <Key className="w-5 h-5 text-amber-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">New Credential</h2>
              <p className="text-sm text-muted-foreground">Add an API key or token</p>
            </div>
          </div>
          <button onClick={onClose} className="btn btn-ghost btn-icon">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit}>
          <div className="modal-body space-y-5">
            {error && (
              <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm flex items-center gap-2">
                <AlertCircle className="w-4 h-4" />
                {error}
              </div>
            )}

            {/* Provider Selection */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-3">
                Provider
              </label>
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
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                Name
              </label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="e.g., Production API Key"
                className="input"
                autoFocus
              />
              {name && (
                <p className="text-xs text-muted-foreground mt-1.5">
                  ID: <code className="text-amber-400">{generateId(name)}</code>
                </p>
              )}
            </div>

            {/* Type */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                Type
              </label>
              <select
                value={type}
                onChange={e => setType(e.target.value as CredentialType)}
                className="input"
              >
                <option value="api_key">API Key</option>
                <option value="token">Token</option>
                <option value="oauth">OAuth</option>
              </select>
            </div>

            {/* Value */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                Value
              </label>
              <input
                type="password"
                value={value}
                onChange={e => setValue(e.target.value)}
                placeholder="sk-..."
                className="input font-mono"
              />
              <p className="text-xs text-muted-foreground mt-1.5 flex items-center gap-1">
                <Shield className="w-3 h-3" />
                Value will be encrypted with AES-256
              </p>
            </div>

            {/* Scope */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                Scope
              </label>
              <select
                value={scope}
                onChange={e => setScope(e.target.value as CredentialScope)}
                className="input"
              >
                <option value="global">Global - Available to all sessions</option>
                <option value="profile">Profile - Tied to specific profile</option>
                <option value="session">Session - Single session only</option>
              </select>
            </div>

            {/* Env Var */}
            <div>
              <label className="block text-sm font-medium text-foreground/90 mb-2">
                Environment Variable (Optional)
              </label>
              <input
                type="text"
                value={envVar}
                onChange={e => setEnvVar(e.target.value)}
                placeholder="e.g., ANTHROPIC_API_KEY"
                className="input font-mono"
              />
              <p className="text-xs text-muted-foreground mt-1.5">
                Injected as this env var in agent sessions
              </p>
            </div>
          </div>

          {/* Footer */}
          <div className="modal-footer">
            <button type="button" onClick={onClose} className="btn btn-secondary">
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim() || !value.trim()}
              className="btn btn-primary"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Key className="w-4 h-4" />
                  Create Credential
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
