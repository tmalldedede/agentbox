import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Key,
  RefreshCw,
  AlertCircle,
  Loader2,
  CheckCircle,
  XCircle,
  Globe,
  User,
  Monitor,
  Shield,
  Eye,
  EyeOff,
} from 'lucide-react'
import type { Credential, CredentialProvider, CredentialScope } from '@/types'
import { useCredentials, useDeleteCredential } from '@/hooks'
import { api } from '@/services/api'
import { toast } from 'sonner'
import CreateCredentialModal from './CreateCredentialModal'

// Provider icon mapping
const providerIcons: Record<CredentialProvider, React.ReactNode> = {
  anthropic: <span className="text-xl">🧠</span>,
  openai: <span className="text-xl">🤖</span>,
  github: <span className="text-xl">🐙</span>,
  custom: <Key className="w-5 h-5" />,
}

// Provider color mapping
const providerColors: Record<CredentialProvider, string> = {
  anthropic: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
  openai: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
  github: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  custom: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
}

// Scope icon mapping
const scopeIcons: Record<CredentialScope, React.ReactNode> = {
  global: <Globe className="w-4 h-4" />,
  profile: <User className="w-4 h-4" />,
  session: <Monitor className="w-4 h-4" />,
}

// Scope color mapping
const scopeColors: Record<CredentialScope, string> = {
  global: 'bg-blue-500/20 text-blue-400',
  profile: 'bg-purple-500/20 text-purple-400',
  session: 'bg-amber-500/20 text-amber-400',
}

// Credential Card component
function CredentialCard({
  credential,
  onDelete,
  onVerify,
  isVerifying,
}: {
  credential: Credential
  onDelete: () => void
  onVerify: () => void
  isVerifying: boolean
}) {
  const [showValue, setShowValue] = useState(false)
  const colors = providerColors[credential.provider] || providerColors.custom
  const icon = providerIcons[credential.provider] || providerIcons.custom
  const scopeIcon = scopeIcons[credential.scope] || scopeIcons.global
  const scopeColor = scopeColors[credential.scope] || scopeColors.global

  return (
    <div className="card p-4 group transition-colors">
      <div className="flex items-start gap-4">
        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${colors}`}>
          {icon}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground truncate">{credential.name}</span>
            {credential.is_valid ? (
              <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400">
                <CheckCircle className="w-3 h-3" />
                Valid
              </span>
            ) : (
              <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-red-500/20 text-red-400">
                <XCircle className="w-3 h-3" />
                Invalid
              </span>
            )}
          </div>

          <div className="flex items-center gap-2 mt-1">
            <Shield className="w-3 h-3 text-muted-foreground" />
            <code className="text-sm text-muted-foreground font-mono">
              {showValue ? credential.value_masked : '••••••••••••'}
            </code>
            <button
              onClick={e => {
                e.stopPropagation()
                setShowValue(!showValue)
              }}
              className="p-1 hover:bg-secondary rounded"
            >
              {showValue ? (
                <EyeOff className="w-3 h-3 text-muted-foreground" />
              ) : (
                <Eye className="w-3 h-3 text-muted-foreground" />
              )}
            </button>
          </div>

          {credential.env_var && (
            <div className="flex items-center gap-2 mt-1">
              <span className="text-xs text-muted-foreground">ENV:</span>
              <code className="text-xs text-amber-400 font-mono">{credential.env_var}</code>
            </div>
          )}

          <div className="flex items-center gap-2 mt-3 flex-wrap">
            <span className={`text-xs px-2 py-0.5 rounded border ${colors}`}>
              {credential.provider}
            </span>
            <span className={`text-xs px-2 py-0.5 rounded ${scopeColor} flex items-center gap-1`}>
              {scopeIcon}
              {credential.scope}
            </span>
            <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">
              {credential.type}
            </span>
            {credential.expires_at && (
              <span className="text-xs px-2 py-0.5 rounded bg-amber-500/20 text-amber-400">
                Expires: {new Date(credential.expires_at).toLocaleDateString()}
              </span>
            )}
          </div>
        </div>

        <div
          className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={e => e.stopPropagation()}
        >
          <button
            onClick={e => {
              e.stopPropagation()
              onVerify()
            }}
            className="btn btn-ghost btn-icon"
            title="Verify Credential"
            disabled={isVerifying}
          >
            {isVerifying ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <CheckCircle className="w-4 h-4" />
            )}
          </button>

          <button
            onClick={e => {
              e.stopPropagation()
              onDelete()
            }}
            className="btn btn-ghost btn-icon text-red-400"
            title="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

export default function CredentialList() {
  const navigate = useNavigate()
  const [scopeFilter, setScopeFilter] = useState<CredentialScope | 'all'>('all')
  const [providerFilter, setProviderFilter] = useState<CredentialProvider | 'all'>('all')
  const [showCreate, setShowCreate] = useState(false)
  const [verifyingId, setVerifyingId] = useState<string | undefined>()

  // React Query hooks
  const { data: credentials = [], isLoading, isFetching, error, refetch } = useCredentials()
  const deleteCredential = useDeleteCredential()

  const handleDelete = (credential: Credential) => {
    if (!confirm(`Delete credential "${credential.name}"? This action cannot be undone.`)) return
    deleteCredential.mutate(credential.id)
  }

  const handleVerify = async (credential: Credential) => {
    setVerifyingId(credential.id)
    try {
      const result = await api.verifyCredential(credential.id)
      if (result.valid) {
        toast.success(`Credential "${credential.name}" is valid`)
      } else {
        toast.error(`Credential "${credential.name}" is invalid or expired`)
      }
      refetch()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to verify credential')
    } finally {
      setVerifyingId(undefined)
    }
  }

  // Filter credentials
  const filteredCredentials = credentials.filter(c => {
    if (scopeFilter !== 'all' && c.scope !== scopeFilter) return false
    if (providerFilter !== 'all' && c.provider !== providerFilter) return false
    return true
  })

  // Group by provider
  const providers = Array.from(new Set(filteredCredentials.map(c => c.provider)))
  const groupedCredentials = providers.reduce(
    (acc, provider) => {
      acc[provider] = filteredCredentials.filter(c => c.provider === provider)
      return acc
    },
    {} as Record<CredentialProvider, Credential[]>
  )

  return (
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Key className="w-6 h-6 text-amber-400" />
            <span className="text-lg font-bold">Credentials</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={scopeFilter}
            onChange={e => setScopeFilter(e.target.value as typeof scopeFilter)}
            className="input py-2 px-3 text-sm"
          >
            <option value="all">All Scopes</option>
            <option value="global">Global</option>
            <option value="profile">Profile</option>
            <option value="session">Session</option>
          </select>

          <select
            value={providerFilter}
            onChange={e => setProviderFilter(e.target.value as typeof providerFilter)}
            className="input py-2 px-3 text-sm"
          >
            <option value="all">All Providers</option>
            <option value="anthropic">Anthropic</option>
            <option value="openai">OpenAI</option>
            <option value="github">GitHub</option>
            <option value="custom">Custom</option>
          </select>

          <button onClick={() => refetch()} className="btn btn-ghost btn-icon" disabled={isFetching}>
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
            <Plus className="w-4 h-4" />
            New Credential
          </button>
        </div>
      </header>

      <div className="p-6">
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load credentials'}
            </span>
          </div>
        )}

        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">Credentials</h1>
          <p className="text-muted-foreground">
            Securely manage API keys and tokens for AI providers and services. Credentials are
            encrypted at rest and can be scoped to global, profile, or session level.
          </p>
        </div>

        <div className="mb-6 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-start gap-3">
          <Shield className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-amber-400 font-medium">Security Notice</p>
            <p className="text-sm text-amber-400/80 mt-1">
              Credentials are stored with AES-256 encryption. Only masked values are shown in the
              UI. The actual values are only decrypted when injected into agent sessions.
            </p>
          </div>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-amber-400 animate-spin" />
          </div>
        ) : filteredCredentials.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Key className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-foreground/90 text-lg">No credentials found</p>
            <p className="text-muted-foreground mt-2">
              {scopeFilter !== 'all' || providerFilter !== 'all'
                ? 'Try changing the filters or add a new credential'
                : 'Add your first API key to get started'}
            </p>
          </div>
        ) : (
          <div className="space-y-8">
            {providers.map(provider => (
              <div key={provider}>
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center ${providerColors[provider]}`}
                  >
                    {providerIcons[provider]}
                  </div>
                  <h2 className="text-lg font-semibold text-foreground capitalize">{provider}</h2>
                  <span className="text-sm text-muted-foreground">({groupedCredentials[provider].length})</span>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                  {groupedCredentials[provider].map(credential => (
                    <CredentialCard
                      key={credential.id}
                      credential={credential}
                      onDelete={() => handleDelete(credential)}
                      onVerify={() => handleVerify(credential)}
                      isVerifying={verifyingId === credential.id}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {showCreate && (
        <CreateCredentialModal
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false)
            refetch()
          }}
        />
      )}
    </div>
  )
}
