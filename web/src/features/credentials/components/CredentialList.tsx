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
  MoreVertical,
  Play,
} from 'lucide-react'
import type { Credential, CredentialProvider, CredentialScope } from '@/types'
import { useCredentials, useDeleteCredential } from '@/hooks'
import { api } from '@/services/api'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
  AlertTitle,
} from '@/components/ui/alert'
import CreateCredentialModal from './CreateCredentialModal'

// Provider icon mapping
const providerIcons: Record<CredentialProvider, React.ReactNode> = {
  anthropic: <span className="text-xl">🧠</span>,
  openai: <span className="text-xl">🤖</span>,
  github: <span className="text-xl">🐙</span>,
  custom: <Key className="w-5 h-5" />,
}

// Provider color mapping
const providerBgColors: Record<CredentialProvider, string> = {
  anthropic: 'bg-orange-500/20 text-orange-400',
  openai: 'bg-emerald-500/20 text-emerald-400',
  github: 'bg-purple-500/20 text-purple-400',
  custom: 'bg-gray-500/20 text-gray-400',
}

// Scope icon mapping
const scopeIcons: Record<CredentialScope, React.ReactNode> = {
  global: <Globe className="w-3 h-3" />,
  profile: <User className="w-3 h-3" />,
  session: <Monitor className="w-3 h-3" />,
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
  const bgColor = providerBgColors[credential.provider] || providerBgColors.custom
  const icon = providerIcons[credential.provider] || providerIcons.custom

  return (
    <Card className="transition-colors hover:border-primary/50">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-lg ${bgColor} flex items-center justify-center`}>
              {icon}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <CardTitle className="text-base">{credential.name}</CardTitle>
                {credential.is_valid ? (
                  <Badge variant="default" className="bg-green-500 text-xs">
                    <CheckCircle className="w-3 h-3 mr-1" />
                    Valid
                  </Badge>
                ) : (
                  <Badge variant="destructive" className="text-xs">
                    <XCircle className="w-3 h-3 mr-1" />
                    Invalid
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <Shield className="w-3 h-3 text-muted-foreground" />
                <code className="text-xs text-muted-foreground font-mono">
                  {showValue ? credential.value_masked : '••••••••••••'}
                </code>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-5 w-5"
                  onClick={(e) => {
                    e.stopPropagation()
                    setShowValue(!showValue)
                  }}
                >
                  {showValue ? (
                    <EyeOff className="w-3 h-3" />
                  ) : (
                    <Eye className="w-3 h-3" />
                  )}
                </Button>
              </div>
            </div>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={(e) => {
                  e.stopPropagation()
                  onVerify()
                }}
                disabled={isVerifying}
              >
                {isVerifying ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <Play className="w-4 h-4 mr-2" />
                )}
                Verify
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-red-600"
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete()
                }}
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent>
        {credential.env_var && (
          <CardDescription className="mb-3">
            ENV: <code className="text-amber-400 font-mono">{credential.env_var}</code>
          </CardDescription>
        )}
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="outline" className="text-xs capitalize">
            {credential.provider}
          </Badge>
          <Badge variant="outline" className="text-xs capitalize">
            {scopeIcons[credential.scope]}
            <span className="ml-1">{credential.scope}</span>
          </Badge>
          <Badge variant="outline" className="text-xs">
            {credential.type}
          </Badge>
          {credential.expires_at && (
            <Badge variant="outline" className="text-xs text-amber-600">
              Expires: {new Date(credential.expires_at).toLocaleDateString()}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
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
          <Select value={scopeFilter} onValueChange={(v) => setScopeFilter(v as typeof scopeFilter)}>
            <SelectTrigger className="w-[130px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Scopes</SelectItem>
              <SelectItem value="global">Global</SelectItem>
              <SelectItem value="profile">Profile</SelectItem>
              <SelectItem value="session">Session</SelectItem>
            </SelectContent>
          </Select>

          <Select value={providerFilter} onValueChange={(v) => setProviderFilter(v as typeof providerFilter)}>
            <SelectTrigger className="w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Providers</SelectItem>
              <SelectItem value="anthropic">Anthropic</SelectItem>
              <SelectItem value="openai">OpenAI</SelectItem>
              <SelectItem value="github">GitHub</SelectItem>
              <SelectItem value="custom">Custom</SelectItem>
            </SelectContent>
          </Select>

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

        <Alert className="mb-6 border-amber-500/30 bg-amber-500/10">
          <Shield className="w-4 h-4 text-amber-400" />
          <AlertTitle className="text-amber-400">Security Notice</AlertTitle>
          <AlertDescription className="text-amber-400/80">
            Credentials are stored with AES-256 encryption. Only masked values are shown in the
            UI. The actual values are only decrypted when injected into agent sessions.
          </AlertDescription>
        </Alert>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-amber-400 animate-spin" />
          </div>
        ) : filteredCredentials.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Key className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg">No credentials found</p>
            <p className="text-muted-foreground mt-2">
              {scopeFilter !== 'all' || providerFilter !== 'all'
                ? 'Try changing the filters or add a new credential'
                : 'Add your first API key to get started'}
            </p>
            <Button className="mt-4" onClick={() => setShowCreate(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Credential
            </Button>
          </div>
        ) : (
          <div className="space-y-8">
            {providers.map(provider => (
              <div key={provider}>
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className={`w-8 h-8 rounded-lg flex items-center justify-center ${providerBgColors[provider]}`}
                  >
                    {providerIcons[provider]}
                  </div>
                  <h2 className="text-lg font-semibold text-foreground capitalize">{provider}</h2>
                  <span className="text-sm text-muted-foreground">({groupedCredentials[provider].length})</span>
                </div>
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
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
