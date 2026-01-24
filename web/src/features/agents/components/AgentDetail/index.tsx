import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  ArrowRight,
  Save,
  Bot,
  Loader2,
  AlertCircle,
  Play,
  Check,
  Copy,
  CheckCheck,
} from 'lucide-react'
import { useAgent, useCreateAgent, useUpdateAgent, useDockerAvailable } from '@/hooks'
import { useProviders } from '@/hooks/useProviders'
import { useRuntimes } from '@/hooks/useRuntimes'
import { useSkills } from '@/hooks/useSkills'
import { useMCPServers } from '@/hooks/useMCPServers'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import type { CreateAgentRequest, UpdateAgentRequest, AgentAPIAccess, AgentStatus, AdapterType, PermissionConfig } from '@/types'

interface AgentDetailProps {
  agentId: string
}

// Auto-generate ID from name: extract ASCII slug + 4-char random suffix
function generateId(name: string) {
  const ascii = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 30)
  const suffix = Math.random().toString(36).slice(2, 6)
  return ascii ? `${ascii}-${suffix}` : `agent-${suffix}`
}

interface FormData {
  name: string
  description: string
  adapter: AdapterType
  provider_id: string
  runtime_id: string
  model: string
  base_url_override: string
  system_prompt: string
  append_system_prompt: string
  permissions: PermissionConfig
  skill_ids: string[]
  mcp_server_ids: string[]
  workspace: string
  env: Record<string, string>
  api_access: AgentAPIAccess
  rate_limit: number
  webhook_url: string
  status: AgentStatus
  output_format: string
}

const defaultFormData: FormData = {
  name: '',
  description: '',
  adapter: 'claude-code',
  provider_id: '',
  runtime_id: '',
  model: '',
  base_url_override: '',
  system_prompt: '',
  append_system_prompt: '',
  permissions: {
    full_auto: true,
    sandbox_mode: 'workspace-write',
    approval_policy: 'on-failure',
  },
  skill_ids: [],
  mcp_server_ids: [],
  workspace: '',
  env: {},
  api_access: 'private',
  rate_limit: 0,
  webhook_url: '',
  status: 'active',
  output_format: '',
}

export default function AgentDetail({ agentId }: AgentDetailProps) {
  const navigate = useNavigate()
  const id = agentId
  const isNew = id === 'new'

  const [formData, setFormData] = useState<FormData>(defaultFormData)
  const [step, setStep] = useState(0) // wizard step for new agents

  const dockerAvailable = useDockerAvailable()

  // Queries
  const { data: agent, isLoading: agentLoading, error: agentError } = useAgent(id)
  const { data: providers = [] } = useProviders()
  const { data: runtimes = [] } = useRuntimes()
  const { data: skills = [] } = useSkills()
  const { data: mcpServers = [] } = useMCPServers()

  // Mutations
  const createAgent = useCreateAgent()
  const updateAgent = useUpdateAgent()

  // Load agent data when editing
  useEffect(() => {
    if (agent && !isNew) {
      setFormData({
        name: agent.name,
        description: agent.description || '',
        adapter: agent.adapter,
        provider_id: agent.provider_id,
        runtime_id: agent.runtime_id || '',
        model: agent.model || '',
        base_url_override: agent.base_url_override || '',
        system_prompt: agent.system_prompt || '',
        append_system_prompt: agent.append_system_prompt || '',
        permissions: agent.permissions || defaultFormData.permissions,
        skill_ids: agent.skill_ids || [],
        mcp_server_ids: agent.mcp_server_ids || [],
        workspace: agent.workspace || '',
        env: agent.env || {},
        api_access: agent.api_access || 'private',
        rate_limit: agent.rate_limit || 0,
        webhook_url: agent.webhook_url || '',
        status: agent.status,
        output_format: agent.output_format || '',
      })
    }
  }, [agent, isNew])

  const handleSubmit = () => {
    if (!formData.name.trim() || !formData.provider_id) return

    if (isNew) {
      const req: CreateAgentRequest = {
        id: generateId(formData.name),
        name: formData.name,
        description: formData.description || undefined,
        adapter: formData.adapter,
        provider_id: formData.provider_id,
        runtime_id: formData.runtime_id || undefined,
        model: formData.model || undefined,
        base_url_override: formData.base_url_override || undefined,
        system_prompt: formData.system_prompt || undefined,
        append_system_prompt: formData.append_system_prompt || undefined,
        permissions: formData.permissions,
        skill_ids: formData.skill_ids.length > 0 ? formData.skill_ids : undefined,
        mcp_server_ids: formData.mcp_server_ids.length > 0 ? formData.mcp_server_ids : undefined,
        workspace: formData.workspace || undefined,
        env: Object.keys(formData.env).length > 0 ? formData.env : undefined,
        api_access: formData.api_access,
        rate_limit: formData.rate_limit || undefined,
        webhook_url: formData.webhook_url || undefined,
        output_format: formData.output_format || undefined,
      }
      createAgent.mutate(req, {
        onSuccess: (data) => {
          navigate({ to: `/agents/${data.id}` })
        },
      })
    } else {
      const req: UpdateAgentRequest = {
        name: formData.name,
        description: formData.description || undefined,
        adapter: formData.adapter,
        provider_id: formData.provider_id,
        runtime_id: formData.runtime_id || undefined,
        model: formData.model || undefined,
        base_url_override: formData.base_url_override || undefined,
        system_prompt: formData.system_prompt || undefined,
        append_system_prompt: formData.append_system_prompt || undefined,
        permissions: formData.permissions,
        skill_ids: formData.skill_ids,
        mcp_server_ids: formData.mcp_server_ids,
        workspace: formData.workspace || undefined,
        env: formData.env,
        api_access: formData.api_access,
        rate_limit: formData.rate_limit || undefined,
        webhook_url: formData.webhook_url || undefined,
        output_format: formData.output_format || undefined,
        status: formData.status,
      }
      updateAgent.mutate({ id, req })
    }
  }

  const isSaving = createAgent.isPending || updateAgent.isPending

  // Filter providers by selected adapter
  const filteredProviders = providers.filter(
    (p) => p.agents?.includes(formData.adapter)
  )

  if (agentLoading && !isNew) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (agentError && !isNew) {
    return (
      <div className="p-6">
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-destructive" />
          <span className="text-destructive">
            {agentError instanceof Error ? agentError.message : 'Failed to load agent'}
          </span>
        </div>
      </div>
    )
  }

  // === NEW AGENT: Step Wizard ===
  if (isNew) {
    return <NewAgentWizard
      formData={formData}
      setFormData={setFormData}
      step={step}
      setStep={setStep}
      filteredProviders={filteredProviders}
      runtimes={runtimes}
      skills={skills}
      onSubmit={handleSubmit}
      isSaving={isSaving}
      onBack={() => navigate({ to: '/agents' })}
    />
  }

  // === EDIT AGENT: Tab Layout ===
  return <EditAgentTabs
    formData={formData}
    setFormData={setFormData}
    filteredProviders={filteredProviders}
    runtimes={runtimes}
    skills={skills}
    mcpServers={mcpServers}
    agentName={agent?.name || ''}
    agentId={id}
    onSubmit={handleSubmit}
    isSaving={isSaving}
    onBack={() => navigate({ to: '/agents' })}
    onTestRun={() => navigate({ to: '/api-playground', search: { agent: id } })}
    dockerAvailable={dockerAvailable}
  />
}

// ============================================================
// NEW AGENT WIZARD
// ============================================================

const STEPS = [
  { title: 'Basic', description: 'Name and engine' },
  { title: 'Behavior', description: 'Prompts and permissions' },
  { title: 'Deploy', description: 'Runtime and access' },
]

interface WizardProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  step: number
  setStep: (s: number) => void
  filteredProviders: any[]
  runtimes: any[]
  skills: any[]
  onSubmit: () => void
  isSaving: boolean
  onBack: () => void
}

function NewAgentWizard({
  formData, setFormData, step, setStep,
  filteredProviders, runtimes, skills,
  onSubmit, isSaving, onBack,
}: WizardProps) {
  const canNext = () => {
    if (step === 0) return formData.name.trim() && formData.provider_id
    return true
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button onClick={onBack} className="p-2 rounded-lg hover:bg-muted">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div>
          <h1 className="text-xl font-bold flex items-center gap-2">
            <Bot className="w-5 h-5 text-primary" />
            Create Agent
          </h1>
          <p className="text-sm text-muted-foreground">
            Step {step + 1} of {STEPS.length}: {STEPS[step].description}
          </p>
        </div>
      </div>

      {/* Step Indicators */}
      <div className="flex items-center gap-2 mb-8">
        {STEPS.map((s, i) => (
          <div key={i} className="flex items-center gap-2 flex-1">
            <div className={`flex items-center justify-center w-7 h-7 rounded-full text-xs font-medium transition-colors
              ${i < step ? 'bg-primary text-primary-foreground' :
                i === step ? 'bg-primary text-primary-foreground' :
                'bg-muted text-muted-foreground'}`}
            >
              {i < step ? <Check className="w-4 h-4" /> : i + 1}
            </div>
            <span className={`text-sm hidden sm:block ${i === step ? 'font-medium' : 'text-muted-foreground'}`}>
              {s.title}
            </span>
            {i < STEPS.length - 1 && <div className="flex-1 h-px bg-border" />}
          </div>
        ))}
      </div>

      {/* Step Content */}
      <Card>
        <CardContent className="pt-6">
          {step === 0 && (
            <StepBasic formData={formData} setFormData={setFormData} filteredProviders={filteredProviders} />
          )}
          {step === 1 && (
            <StepBehavior formData={formData} setFormData={setFormData} skills={skills} />
          )}
          {step === 2 && (
            <StepDeploy formData={formData} setFormData={setFormData} runtimes={runtimes} />
          )}
        </CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex justify-between mt-6">
        <Button
          variant="outline"
          onClick={() => step > 0 ? setStep(step - 1) : onBack()}
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          {step > 0 ? 'Back' : 'Cancel'}
        </Button>

        {step < STEPS.length - 1 ? (
          <Button onClick={() => setStep(step + 1)} disabled={!canNext()}>
            Next
            <ArrowRight className="w-4 h-4 ml-2" />
          </Button>
        ) : (
          <Button onClick={onSubmit} disabled={isSaving || !canNext()}>
            {isSaving ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Save className="w-4 h-4 mr-2" />}
            Create Agent
          </Button>
        )}
      </div>

      {/* Preview ID */}
      {formData.name && (
        <p className="text-xs text-muted-foreground mt-4 text-center">
          Agent ID: <code className="bg-muted px-1.5 py-0.5 rounded">{generateId(formData.name)}</code>
        </p>
      )}
    </div>
  )
}

// ============================================================
// EDIT AGENT TABS
// ============================================================

interface EditProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  filteredProviders: any[]
  runtimes: any[]
  skills: any[]
  mcpServers: any[]
  agentName: string
  agentId: string
  onSubmit: () => void
  isSaving: boolean
  onBack: () => void
  onTestRun: () => void
  dockerAvailable: boolean
}

function EditAgentTabs({
  formData, setFormData, filteredProviders, runtimes, skills, mcpServers,
  agentName, agentId, onSubmit, isSaving, onBack, onTestRun, dockerAvailable,
}: EditProps) {
  const [idCopied, setIdCopied] = useState(false)
  const copyAgentId = () => {
    navigator.clipboard.writeText(agentId)
    setIdCopied(true)
    setTimeout(() => setIdCopied(false), 2000)
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="p-2 rounded-lg hover:bg-muted">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-xl font-bold flex items-center gap-2">
              <Bot className="w-5 h-5 text-primary" />
              {agentName}
            </h1>
            <button
              onClick={copyAgentId}
              className="flex items-center gap-1.5 text-sm text-muted-foreground font-mono hover:text-foreground transition-colors"
              title="Copy Agent ID"
            >
              {agentId}
              {idCopied ? <CheckCheck className="w-3.5 h-3.5 text-green-500" /> : <Copy className="w-3.5 h-3.5" />}
            </button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={onTestRun} disabled={!dockerAvailable}>
            <Play className="w-4 h-4 mr-2" />
            Test Run
          </Button>
          <Button onClick={onSubmit} disabled={isSaving}>
            {isSaving ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Save className="w-4 h-4 mr-2" />}
            Save
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="basic">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="basic">Basic</TabsTrigger>
          <TabsTrigger value="behavior">Behavior</TabsTrigger>
          <TabsTrigger value="capabilities">Capabilities</TabsTrigger>
          <TabsTrigger value="deploy">Deploy</TabsTrigger>
          <TabsTrigger value="advanced">Advanced</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Basic Configuration</CardTitle>
              <CardDescription>Agent identity and engine settings</CardDescription>
            </CardHeader>
            <CardContent>
              <StepBasic formData={formData} setFormData={setFormData} filteredProviders={filteredProviders} showDescription />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="behavior" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Behavior</CardTitle>
              <CardDescription>System prompts and permission settings</CardDescription>
            </CardHeader>
            <CardContent>
              <StepBehavior formData={formData} setFormData={setFormData} skills={skills} showAppendPrompt />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="capabilities" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Capabilities</CardTitle>
              <CardDescription>Skills and MCP servers available to this agent</CardDescription>
            </CardHeader>
            <CardContent>
              <CapabilitiesSection formData={formData} setFormData={setFormData} skills={skills} mcpServers={mcpServers} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="deploy" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Deployment</CardTitle>
              <CardDescription>Runtime, access control and status</CardDescription>
            </CardHeader>
            <CardContent>
              <StepDeploy formData={formData} setFormData={setFormData} runtimes={runtimes} showStatus />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="advanced" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Advanced Settings</CardTitle>
              <CardDescription>Override URLs, environment variables, and output format</CardDescription>
            </CardHeader>
            <CardContent>
              <AdvancedSection formData={formData} setFormData={setFormData} />
            </CardContent>
          </Card>
        </TabsContent>

      </Tabs>
    </div>
  )
}

// ============================================================
// SHARED FORM SECTIONS
// ============================================================

interface StepBasicProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  filteredProviders: any[]
  showDescription?: boolean
}

function StepBasic({ formData, setFormData, filteredProviders, showDescription }: StepBasicProps) {
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Name *</Label>
        <Input
          id="name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          placeholder="My Agent"
        />
      </div>

      {showDescription && (
        <div className="space-y-2">
          <Label htmlFor="description">Description</Label>
          <Textarea
            id="description"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="What does this agent do?"
            rows={2}
          />
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Adapter *</Label>
          <Select
            value={formData.adapter}
            onValueChange={(value) => setFormData({ ...formData, adapter: value as AdapterType, provider_id: '' })}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="claude-code">Claude Code</SelectItem>
              <SelectItem value="codex">Codex</SelectItem>
              <SelectItem value="opencode">OpenCode</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label>Provider *</Label>
          <Select
            value={formData.provider_id}
            onValueChange={(value) => setFormData({ ...formData, provider_id: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select provider" />
            </SelectTrigger>
            <SelectContent>
              {filteredProviders.map((provider) => (
                <SelectItem key={provider.id} value={provider.id}>
                  <div className="flex items-center gap-2">
                    <span>{provider.name}</span>
                    {provider.is_configured && <Check className="w-3 h-3 text-green-500" />}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {formData.provider_id && !filteredProviders.find(p => p.id === formData.provider_id)?.is_configured && (
            <p className="text-xs text-amber-500">API key not configured for this provider</p>
          )}
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="model">Model</Label>
        <Input
          id="model"
          value={formData.model}
          onChange={(e) => setFormData({ ...formData, model: e.target.value })}
          placeholder="Leave empty for provider default"
        />
        {formData.provider_id && (() => {
          const provider = filteredProviders.find(p => p.id === formData.provider_id)
          if (provider?.default_models?.length) {
            return (
              <div className="flex flex-wrap gap-1 mt-1">
                {provider.default_models.map((m: string) => (
                  <button
                    key={m}
                    type="button"
                    onClick={() => setFormData({ ...formData, model: m })}
                    className={`text-xs px-2 py-0.5 rounded-full border transition-colors
                      ${formData.model === m ? 'bg-primary text-primary-foreground border-primary' : 'hover:bg-muted border-border'}`}
                  >
                    {m}
                  </button>
                ))}
              </div>
            )
          }
          return null
        })()}
      </div>
    </div>
  )
}

interface StepBehaviorProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  skills: any[]
  showAppendPrompt?: boolean
}

function StepBehavior({ formData, setFormData, skills, showAppendPrompt }: StepBehaviorProps) {
  return (
    <div className="space-y-6">
      {/* System Prompt */}
      <div className="space-y-2">
        <Label>System Prompt</Label>
        <Textarea
          value={formData.system_prompt}
          onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
          placeholder="You are a helpful assistant that..."
          rows={5}
          className="font-mono text-sm"
        />
      </div>

      {showAppendPrompt && (
        <div className="space-y-2">
          <Label>Append System Prompt</Label>
          <Textarea
            value={formData.append_system_prompt}
            onChange={(e) => setFormData({ ...formData, append_system_prompt: e.target.value })}
            placeholder="Additional instructions appended after main prompt..."
            rows={3}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">Appended after the main system prompt</p>
        </div>
      )}

      {/* Permissions */}
      <div className="space-y-4">
        <Label className="text-base font-medium">Permissions</Label>

        <div className="flex items-center justify-between rounded-lg border p-3">
          <div>
            <p className="text-sm font-medium">Full Auto</p>
            <p className="text-xs text-muted-foreground">Skip all confirmation prompts</p>
          </div>
          <Switch
            checked={formData.permissions.full_auto ?? false}
            onCheckedChange={(checked) => setFormData({
              ...formData,
              permissions: { ...formData.permissions, full_auto: checked }
            })}
          />
        </div>

        {formData.adapter === 'codex' && (
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Sandbox Mode</Label>
              <Select
                value={formData.permissions.sandbox_mode || 'workspace-write'}
                onValueChange={(value) => setFormData({
                  ...formData,
                  permissions: { ...formData.permissions, sandbox_mode: value as PermissionConfig['sandbox_mode'] }
                })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="read-only">Read Only</SelectItem>
                  <SelectItem value="workspace-write">Workspace Write</SelectItem>
                  <SelectItem value="danger-full-access">Full Access</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Approval Policy</Label>
              <Select
                value={formData.permissions.approval_policy || 'on-failure'}
                onValueChange={(value) => setFormData({
                  ...formData,
                  permissions: { ...formData.permissions, approval_policy: value as PermissionConfig['approval_policy'] }
                })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="never">Never ask</SelectItem>
                  <SelectItem value="on-failure">On Failure</SelectItem>
                  <SelectItem value="on-request">On Request</SelectItem>
                  <SelectItem value="untrusted">Always (Untrusted)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        )}

        {formData.adapter === 'claude-code' && (
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div>
              <p className="text-sm font-medium">Skip All Permissions</p>
              <p className="text-xs text-muted-foreground">Equivalent to --dangerously-skip-permissions</p>
            </div>
            <Switch
              checked={formData.permissions.skip_all ?? false}
              onCheckedChange={(checked) => setFormData({
                ...formData,
                permissions: { ...formData.permissions, skip_all: checked }
              })}
            />
          </div>
        )}
      </div>

      {/* Quick Skills/MCP in wizard (compact) */}
      {skills.length > 0 && (
        <div className="space-y-2">
          <Label>Skills (optional)</Label>
          <div className="flex flex-wrap gap-2">
            {skills.slice(0, 8).map((skill: any) => (
              <label key={skill.id} className="flex items-center gap-2 text-sm border rounded-lg px-3 py-1.5 cursor-pointer hover:bg-muted transition-colors">
                <Checkbox
                  checked={formData.skill_ids.includes(skill.id)}
                  onCheckedChange={(checked) => {
                    const ids = checked
                      ? [...formData.skill_ids, skill.id]
                      : formData.skill_ids.filter(id => id !== skill.id)
                    setFormData({ ...formData, skill_ids: ids })
                  }}
                />
                {skill.name}
              </label>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

interface StepDeployProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  runtimes: any[]
  showStatus?: boolean
}

function StepDeploy({ formData, setFormData, runtimes, showStatus }: StepDeployProps) {
  return (
    <div className="space-y-4">
      {/* Runtime */}
      <div className="space-y-2">
        <Label>Runtime</Label>
        <Select
          value={formData.runtime_id || '__default__'}
          onValueChange={(value) => setFormData({ ...formData, runtime_id: value === '__default__' ? '' : value })}
        >
          <SelectTrigger>
            <SelectValue placeholder="Default runtime" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__default__">Default</SelectItem>
            {runtimes.map((rt: any) => (
              <SelectItem key={rt.id} value={rt.id}>
                {rt.name} ({rt.cpus} CPU, {rt.memory_mb}MB)
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-xs text-muted-foreground">Container resources for running this agent</p>
      </div>

      {/* Workspace */}
      <div className="space-y-2">
        <Label>Workspace</Label>
        <Input
          value={formData.workspace}
          onChange={(e) => setFormData({ ...formData, workspace: e.target.value })}
          placeholder="Auto-generated per task (e.g. /workspace)"
        />
        <p className="text-xs text-muted-foreground">
          Default working directory inside the container. Leave empty for auto-generated unique path per task.
        </p>
      </div>

      {/* API Access */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Access Level</Label>
          <Select
            value={formData.api_access}
            onValueChange={(value) => setFormData({ ...formData, api_access: value as AgentAPIAccess })}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="private">Private</SelectItem>
              <SelectItem value="api_key">API Key</SelectItem>
              <SelectItem value="public">Public</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Rate Limit (req/min)</Label>
          <Input
            type="number"
            value={formData.rate_limit || ''}
            onChange={(e) => setFormData({ ...formData, rate_limit: parseInt(e.target.value) || 0 })}
            placeholder="0 = unlimited"
          />
        </div>
      </div>

      {/* Webhook */}
      <div className="space-y-2">
        <Label>Webhook URL</Label>
        <Input
          type="url"
          value={formData.webhook_url}
          onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
          placeholder="https://your-server.com/webhook"
        />
        <p className="text-xs text-muted-foreground">Receive notifications when agent runs complete</p>
      </div>

      {/* Status (edit only) */}
      {showStatus && (
        <div className="space-y-2">
          <Label>Status</Label>
          <Select
            value={formData.status}
            onValueChange={(value) => setFormData({ ...formData, status: value as AgentStatus })}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  )
}

// Capabilities section (edit only - full skill/mcp list)
interface CapabilitiesProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
  skills: any[]
  mcpServers: any[]
}

function CapabilitiesSection({ formData, setFormData, skills, mcpServers }: CapabilitiesProps) {
  return (
    <div className="space-y-6">
      {/* Skills */}
      <div className="space-y-3">
        <Label className="text-base font-medium">Skills ({formData.skill_ids.length} selected)</Label>
        {skills.length === 0 ? (
          <p className="text-sm text-muted-foreground">No skills configured. Add skills in the Skills section.</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-60 overflow-y-auto">
            {skills.map((skill: any) => (
              <label key={skill.id} className="flex items-start gap-3 border rounded-lg p-3 cursor-pointer hover:bg-muted/50 transition-colors">
                <Checkbox
                  className="mt-0.5"
                  checked={formData.skill_ids.includes(skill.id)}
                  onCheckedChange={(checked) => {
                    const ids = checked
                      ? [...formData.skill_ids, skill.id]
                      : formData.skill_ids.filter(id => id !== skill.id)
                    setFormData({ ...formData, skill_ids: ids })
                  }}
                />
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{skill.name}</p>
                  {skill.description && (
                    <p className="text-xs text-muted-foreground line-clamp-2">{skill.description}</p>
                  )}
                </div>
              </label>
            ))}
          </div>
        )}
      </div>

      {/* MCP Servers */}
      <div className="space-y-3">
        <Label className="text-base font-medium">MCP Servers ({formData.mcp_server_ids.length} selected)</Label>
        {mcpServers.length === 0 ? (
          <p className="text-sm text-muted-foreground">No MCP servers configured. Add servers in the MCP Servers section.</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-60 overflow-y-auto">
            {mcpServers.map((server: any) => (
              <label key={server.id} className="flex items-start gap-3 border rounded-lg p-3 cursor-pointer hover:bg-muted/50 transition-colors">
                <Checkbox
                  className="mt-0.5"
                  checked={formData.mcp_server_ids.includes(server.id)}
                  onCheckedChange={(checked) => {
                    const ids = checked
                      ? [...formData.mcp_server_ids, server.id]
                      : formData.mcp_server_ids.filter(id => id !== server.id)
                    setFormData({ ...formData, mcp_server_ids: ids })
                  }}
                />
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{server.name}</p>
                  {server.description && (
                    <p className="text-xs text-muted-foreground line-clamp-2">{server.description}</p>
                  )}
                </div>
              </label>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Advanced section (edit only)
interface AdvancedProps {
  formData: FormData
  setFormData: (fn: FormData | ((prev: FormData) => FormData)) => void
}

function AdvancedSection({ formData, setFormData }: AdvancedProps) {
  const [envText, setEnvText] = useState(() => {
    return Object.entries(formData.env).map(([k, v]) => `${k}=${v}`).join('\n')
  })

  const parseEnv = (text: string) => {
    const env: Record<string, string> = {}
    text.split('\n').forEach(line => {
      const idx = line.indexOf('=')
      if (idx > 0) {
        const key = line.slice(0, idx).trim()
        const val = line.slice(idx + 1).trim()
        if (key) env[key] = val
      }
    })
    return env
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>Base URL Override</Label>
        <Input
          value={formData.base_url_override}
          onChange={(e) => setFormData({ ...formData, base_url_override: e.target.value })}
          placeholder="Override provider's base URL for this agent"
        />
        <p className="text-xs text-muted-foreground">
          Same provider may need different URLs per adapter (e.g. /anthropic vs /v1)
        </p>
      </div>

      <div className="space-y-2">
        <Label>Environment Variables</Label>
        <Textarea
          value={envText}
          onChange={(e) => {
            setEnvText(e.target.value)
            setFormData({ ...formData, env: parseEnv(e.target.value) })
          }}
          placeholder={"KEY=value\nANOTHER_KEY=another_value"}
          rows={4}
          className="font-mono text-sm"
        />
        <p className="text-xs text-muted-foreground">One KEY=value per line. Injected into the agent container.</p>
      </div>

      <div className="space-y-2">
        <Label>Output Format</Label>
        <Select
          value={formData.output_format || '__default__'}
          onValueChange={(value) => setFormData({ ...formData, output_format: value === '__default__' ? '' : value })}
        >
          <SelectTrigger>
            <SelectValue placeholder="Default" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__default__">Default</SelectItem>
            <SelectItem value="text">Text</SelectItem>
            <SelectItem value="json">JSON</SelectItem>
            <SelectItem value="stream-json">Stream JSON</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
