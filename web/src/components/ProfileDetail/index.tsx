import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  Save,
  Trash2,
  RefreshCw,
  AlertCircle,
  Loader2,
  Lock,
  Server,
  Zap,
  Key,
  Shield,
  Cpu,
  FileText,
} from 'lucide-react'
import type { Profile, Provider } from '../../types'
import {
  useProfile,
  useCreateProfile,
  useUpdateProfile,
  useDeleteProfile,
  useMCPServers,
  useSkills,
  useCredentials,
} from '../../hooks'
import { Section } from './Section'
import { BasicInfoSection } from './BasicInfoSection'
import { AdvancedModelSection } from './AdvancedModelSection'
import { PermissionsSection } from './PermissionsSection'
import { ResourcesSection } from './ResourcesSection'
import { MCPServerSelector } from './MCPServerSelector'
import { SkillSelector } from './SkillSelector'
import { CredentialSelector } from './CredentialSelector'

export default function ProfileDetail() {
  const navigate = useNavigate()
  const { profileId } = useParams<{ profileId: string }>()
  const isNewProfile = profileId === 'new'

  // React Query hooks
  const {
    data: profile,
    isLoading: profileLoading,
    error: profileError,
    refetch: refetchProfile,
  } = useProfile(isNewProfile ? undefined : profileId)

  const { data: mcpServers = [] } = useMCPServers()
  const { data: skills = [] } = useSkills()
  const { data: credentials = [] } = useCredentials()

  // Mutations
  const createProfile = useCreateProfile()
  const updateProfile = useUpdateProfile()
  const deleteProfile = useDeleteProfile()

  // Selected resources
  const [selectedMCPIds, setSelectedMCPIds] = useState<string[]>([])
  const [selectedSkillIds, setSelectedSkillIds] = useState<string[]>([])
  const [selectedCredentialIds, setSelectedCredentialIds] = useState<string[]>([])

  // Editable fields
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [adapter, setAdapter] = useState('claude-code')
  const [modelName, setModelName] = useState('')
  const [modelBaseUrl, setModelBaseUrl] = useState('')
  const [modelProvider, setModelProvider] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')

  // Extended ModelConfig fields
  const [haikuModel, setHaikuModel] = useState('')
  const [sonnetModel, setSonnetModel] = useState('')
  const [opusModel, setOpusModel] = useState('')
  const [timeoutMs, setTimeoutMs] = useState<number | undefined>()
  const [maxOutputTokens, setMaxOutputTokens] = useState<number | undefined>()
  const [disableTraffic, setDisableTraffic] = useState(false)

  // Provider state
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null)
  const [showProviderSelector, setShowProviderSelector] = useState(false)

  // Initialize form when profile loads
  useEffect(() => {
    if (profile) {
      setName(profile.name)
      setDescription(profile.description || '')
      setAdapter(profile.adapter)
      setModelName(profile.model.name || '')
      setModelBaseUrl(profile.model.base_url || '')
      setModelProvider(profile.model.provider || '')
      setSystemPrompt(profile.system_prompt || '')

      // Extended ModelConfig fields
      setHaikuModel(profile.model.haiku_model || '')
      setSonnetModel(profile.model.sonnet_model || '')
      setOpusModel(profile.model.opus_model || '')
      setTimeoutMs(profile.model.timeout_ms)
      setMaxOutputTokens(profile.model.max_output_tokens)
      setDisableTraffic(profile.model.disable_traffic || false)

      // Extract MCP server IDs
      if (profile.mcp_servers) {
        setSelectedMCPIds(profile.mcp_servers.map(s => s.name))
      }
    }
  }, [profile])

  const handleSave = async () => {
    if (!name.trim()) {
      return
    }

    // Build MCP server configs from selected IDs
    const mcpServerConfigs = selectedMCPIds
      .map(id => {
        const server = mcpServers.find(s => s.id === id)
        if (!server) return null
        return {
          name: server.id,
          command: server.command,
          args: server.args,
          env: server.env,
          description: server.description,
        }
      })
      .filter(Boolean)

    const profileData = {
      id: isNewProfile
        ? `profile-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
        : profileId!,
      name,
      description,
      adapter: adapter as 'claude-code' | 'codex',
      model: {
        name: modelName,
        provider: modelProvider || (adapter === 'claude-code' ? 'anthropic' : 'openai'),
        base_url: modelBaseUrl || undefined,
        haiku_model: haikuModel || undefined,
        sonnet_model: sonnetModel || undefined,
        opus_model: opusModel || undefined,
        timeout_ms: timeoutMs,
        max_output_tokens: maxOutputTokens,
        disable_traffic: disableTraffic || undefined,
      },
      system_prompt: systemPrompt,
      permissions: profile?.permissions || {},
      resources: profile?.resources || {},
      mcp_servers: mcpServerConfigs as any,
    }

    if (isNewProfile) {
      createProfile.mutate(profileData as any, {
        onSuccess: created => {
          setTimeout(() => navigate(`/profiles/${created.id}`), 500)
        },
      })
    } else {
      updateProfile.mutate({ id: profileId!, data: profileData as any })
    }
  }

  const handleDelete = () => {
    if (!profile || !profileId) return
    if (!confirm(`Delete profile "${profile.name}"? This action cannot be undone.`)) return

    deleteProfile.mutate(profileId, {
      onSuccess: () => navigate('/profiles'),
    })
  }

  const isLoading = profileLoading && !isNewProfile
  const isSaving = createProfile.isPending || updateProfile.isPending

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  if (profileError && !isNewProfile) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center">
        <AlertCircle className="w-16 h-16 text-red-400 mb-4" />
        <p className="text-lg text-primary">Profile not found</p>
        <button onClick={() => navigate('/profiles')} className="btn btn-primary mt-4">
          Back to Profiles
        </button>
      </div>
    )
  }

  // Create empty profile for new profile mode
  const currentProfile: Profile = profile || {
    id: '',
    name: '',
    description: '',
    adapter: 'claude-code',
    model: { name: '', provider: '' },
    system_prompt: '',
    permissions: {},
    resources: {},
    features: {},
    mcp_servers: [],
    is_built_in: false,
    is_public: false,
    created_at: '',
    updated_at: '',
  }

  const adapterColors: Record<string, string> = {
    'claude-code': 'bg-purple-500/20 text-purple-400',
    codex: 'bg-emerald-500/20 text-emerald-400',
    opencode: 'bg-blue-500/20 text-blue-400',
  }

  const isBuiltIn = !isNewProfile && currentProfile.is_built_in
  const canEdit = isNewProfile || !isBuiltIn

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/profiles')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <div
              className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                adapterColors[isNewProfile ? adapter : currentProfile.adapter] ||
                'bg-blue-500/20 text-blue-400'
              }`}
            >
              {(isNewProfile ? adapter : currentProfile.adapter) === 'claude-code' ? (
                <Cpu className="w-5 h-5" />
              ) : (
                <Shield className="w-5 h-5" />
              )}
            </div>
            <div>
              <span className="text-lg font-bold">
                {isNewProfile ? 'New Profile' : currentProfile.name}
              </span>
              {isBuiltIn && (
                <span className="ml-2 text-xs px-2 py-0.5 rounded bg-secondary text-muted">
                  <Lock className="w-3 h-3 inline mr-1" />
                  Built-in
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {!isNewProfile && (
            <button onClick={() => refetchProfile()} className="btn btn-ghost btn-icon">
              <RefreshCw className={`w-4 h-4 ${profileLoading ? 'animate-spin' : ''}`} />
            </button>
          )}
          {canEdit && (
            <>
              {!isNewProfile && (
                <button
                  onClick={handleDelete}
                  className="btn btn-ghost text-red-400"
                  disabled={deleteProfile.isPending}
                >
                  <Trash2 className="w-4 h-4" />
                  Delete
                </button>
              )}
              <button onClick={handleSave} className="btn btn-primary" disabled={isSaving}>
                {isSaving ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Save className="w-4 h-4" />
                )}
                {isNewProfile ? 'Create' : 'Save'}
              </button>
            </>
          )}
        </div>
      </header>

      <div className="p-6 max-w-4xl mx-auto">
        <div className="space-y-4">
          {/* Basic Info */}
          <BasicInfoSection
            name={name}
            setName={setName}
            description={description}
            setDescription={setDescription}
            adapter={adapter}
            setAdapter={setAdapter}
            modelName={modelName}
            setModelName={setModelName}
            modelBaseUrl={modelBaseUrl}
            setModelBaseUrl={setModelBaseUrl}
            modelProvider={modelProvider}
            setModelProvider={setModelProvider}
            selectedProvider={selectedProvider}
            setSelectedProvider={setSelectedProvider}
            showProviderSelector={showProviderSelector}
            setShowProviderSelector={setShowProviderSelector}
            isNewProfile={isNewProfile}
            isBuiltIn={isBuiltIn}
          />

          {/* Advanced Model Config (Claude Code only) */}
          {adapter === 'claude-code' && (
            <AdvancedModelSection
              haikuModel={haikuModel}
              setHaikuModel={setHaikuModel}
              sonnetModel={sonnetModel}
              setSonnetModel={setSonnetModel}
              opusModel={opusModel}
              setOpusModel={setOpusModel}
              timeoutMs={timeoutMs}
              setTimeoutMs={setTimeoutMs}
              maxOutputTokens={maxOutputTokens}
              setMaxOutputTokens={setMaxOutputTokens}
              disableTraffic={disableTraffic}
              setDisableTraffic={setDisableTraffic}
              disabled={!canEdit}
            />
          )}

          {/* System Prompt */}
          <Section title="System Prompt" icon={<FileText className="w-5 h-5" />} defaultOpen={false}>
            <div className="mt-4">
              <textarea
                value={systemPrompt}
                onChange={e => setSystemPrompt(e.target.value)}
                className="input w-full font-mono text-sm"
                rows={10}
                placeholder="Enter system prompt for the agent..."
                disabled={!canEdit}
              />
            </div>
          </Section>

          {/* Permissions */}
          <PermissionsSection profile={currentProfile} />

          {/* MCP Servers */}
          <Section title="MCP Servers" icon={<Server className="w-5 h-5" />}>
            <p className="text-sm text-muted mt-3">
              Select MCP servers to enable for this profile. These servers will be started
              automatically when a session using this profile is created.
            </p>
            <MCPServerSelector
              selectedIds={selectedMCPIds}
              onChange={setSelectedMCPIds}
              availableServers={mcpServers}
              disabled={!canEdit}
            />
          </Section>

          {/* Skills */}
          <Section title="Skills" icon={<Zap className="w-5 h-5" />} defaultOpen={false}>
            <p className="text-sm text-muted mt-3">
              Select skills available for this profile. Skills provide reusable task templates that
              can be invoked with commands like <code>/commit</code>.
            </p>
            <SkillSelector
              selectedIds={selectedSkillIds}
              onChange={setSelectedSkillIds}
              availableSkills={skills}
              disabled={!canEdit}
            />
          </Section>

          {/* Credentials */}
          <Section title="Credentials" icon={<Key className="w-5 h-5" />} defaultOpen={false}>
            <p className="text-sm text-muted mt-3">
              Select credentials to inject into sessions using this profile. Credentials are
              securely stored and will be available as environment variables.
            </p>
            <CredentialSelector
              selectedIds={selectedCredentialIds}
              onChange={setSelectedCredentialIds}
              availableCredentials={credentials}
              disabled={!canEdit}
            />
          </Section>

          {/* Resources */}
          <ResourcesSection profile={currentProfile} />
        </div>
      </div>
    </div>
  )
}
