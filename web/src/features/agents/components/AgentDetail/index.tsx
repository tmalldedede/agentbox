import { useState, useEffect } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import {
  ArrowLeft,
  Save,
  Bot,
  Loader2,
  AlertCircle,
  Play,
} from 'lucide-react'
import { useSmartAgent, useCreateSmartAgent, useUpdateSmartAgent, useProfiles } from '@/hooks'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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
import type { CreateSmartAgentRequest, UpdateSmartAgentRequest, SmartAgentAPIAccess, SmartAgentStatus } from '@/types'

export default function AgentDetail() {
  const navigate = useNavigate()
  const { id } = useParams({ from: '/_authenticated/agents/$id' })
  const isNew = id === 'new'

  // Form state
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    description: '',
    icon: '',
    profile_id: '',
    system_prompt: '',
    api_access: 'private' as SmartAgentAPIAccess,
    rate_limit: 0,
    webhook_url: '',
    status: 'active' as SmartAgentStatus,
  })

  // Queries
  const { data: agent, isLoading: agentLoading, error: agentError } = useSmartAgent(id)
  const { data: profiles = [], isLoading: profilesLoading } = useProfiles()

  // Mutations
  const createAgent = useCreateSmartAgent()
  const updateAgent = useUpdateSmartAgent()

  // Load agent data when editing
  useEffect(() => {
    if (agent && !isNew) {
      setFormData({
        id: agent.id,
        name: agent.name,
        description: agent.description || '',
        icon: agent.icon || '',
        profile_id: agent.profile_id,
        system_prompt: agent.system_prompt || '',
        api_access: agent.api_access,
        rate_limit: agent.rate_limit || 0,
        webhook_url: agent.webhook_url || '',
        status: agent.status,
      })
    }
  }, [agent, isNew])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (isNew) {
      const req: CreateSmartAgentRequest = {
        id: formData.id,
        name: formData.name,
        description: formData.description || undefined,
        icon: formData.icon || undefined,
        profile_id: formData.profile_id,
        system_prompt: formData.system_prompt || undefined,
        api_access: formData.api_access,
        rate_limit: formData.rate_limit || undefined,
        webhook_url: formData.webhook_url || undefined,
      }
      createAgent.mutate(req, {
        onSuccess: (data) => {
          navigate({ to: `/agents/${data.id}` })
        },
      })
    } else {
      const req: UpdateSmartAgentRequest = {
        name: formData.name,
        description: formData.description || undefined,
        icon: formData.icon || undefined,
        profile_id: formData.profile_id,
        system_prompt: formData.system_prompt || undefined,
        api_access: formData.api_access,
        rate_limit: formData.rate_limit || undefined,
        webhook_url: formData.webhook_url || undefined,
        status: formData.status,
      }
      updateAgent.mutate({ id, req })
    }
  }

  const isLoading = agentLoading || profilesLoading
  const isSaving = createAgent.isPending || updateAgent.isPending

  if (isLoading && !isNew) {
    return (
      <div className="flex items-center justify-center h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-blue-400" />
      </div>
    )
  }

  if (agentError && !isNew) {
    return (
      <div className="p-6">
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
          <AlertCircle className="w-5 h-5 text-red-400" />
          <span className="text-red-400">
            {agentError instanceof Error ? agentError.message : 'Failed to load agent'}
          </span>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/agents' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Bot className="w-6 h-6 text-blue-400" />
            <span className="text-lg font-bold">
              {isNew ? 'Create Agent' : `Edit: ${agent?.name}`}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {!isNew && (
            <Button variant="outline" onClick={() => navigate({ to: `/agents/${id}/run` })}>
              <Play className="w-4 h-4 mr-2" />
              Test Run
            </Button>
          )}
          <Button onClick={handleSubmit} disabled={isSaving}>
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            {isNew ? 'Create' : 'Save'}
          </Button>
        </div>
      </header>

      <form onSubmit={handleSubmit} className="p-6 max-w-4xl">
        <div className="space-y-6">
          {/* Basic Info */}
          <Card>
            <CardHeader>
              <CardTitle>Basic Information</CardTitle>
              <CardDescription>
                Configure the agent's identity and description
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="id">Agent ID</Label>
                  <Input
                    id="id"
                    value={formData.id}
                    onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                    placeholder="my-agent"
                    disabled={!isNew}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="My Agent"
                    required
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="icon">Icon (Emoji)</Label>
                  <Input
                    id="icon"
                    value={formData.icon}
                    onChange={(e) => setFormData({ ...formData, icon: e.target.value })}
                    placeholder="🤖"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="profile">Profile</Label>
                  <Select
                    value={formData.profile_id}
                    onValueChange={(value) => setFormData({ ...formData, profile_id: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a profile" />
                    </SelectTrigger>
                    <SelectContent>
                      {profiles.map((profile) => (
                        <SelectItem key={profile.id} value={profile.id}>
                          {profile.name} ({profile.adapter})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
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
            </CardContent>
          </Card>

          {/* System Prompt */}
          <Card>
            <CardHeader>
              <CardTitle>System Prompt</CardTitle>
              <CardDescription>
                Define the agent's behavior and personality
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Textarea
                value={formData.system_prompt}
                onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
                placeholder="You are a helpful assistant that..."
                rows={6}
                className="font-mono text-sm"
              />
            </CardContent>
          </Card>

          {/* API Settings */}
          <Card>
            <CardHeader>
              <CardTitle>API Settings</CardTitle>
              <CardDescription>
                Configure how the agent can be accessed externally
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="api_access">Access Level</Label>
                  <Select
                    value={formData.api_access}
                    onValueChange={(value) => setFormData({ ...formData, api_access: value as SmartAgentAPIAccess })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="private">Private (Internal only)</SelectItem>
                      <SelectItem value="api_key">API Key (Requires authentication)</SelectItem>
                      <SelectItem value="public">Public (No authentication)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="rate_limit">Rate Limit (req/min)</Label>
                  <Input
                    id="rate_limit"
                    type="number"
                    value={formData.rate_limit || ''}
                    onChange={(e) => setFormData({ ...formData, rate_limit: parseInt(e.target.value) || 0 })}
                    placeholder="0 = unlimited"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="webhook_url">Webhook URL</Label>
                <Input
                  id="webhook_url"
                  type="url"
                  value={formData.webhook_url}
                  onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
                  placeholder="https://your-server.com/webhook"
                />
                <p className="text-xs text-muted-foreground">
                  Optional: Receive notifications when agent runs complete
                </p>
              </div>
              {!isNew && (
                <div className="space-y-2">
                  <Label htmlFor="status">Status</Label>
                  <Select
                    value={formData.status}
                    onValueChange={(value) => setFormData({ ...formData, status: value as SmartAgentStatus })}
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
            </CardContent>
          </Card>

          {/* API Endpoint Info (for existing agents) */}
          {!isNew && (
            <Card>
              <CardHeader>
                <CardTitle>API Endpoint</CardTitle>
                <CardDescription>
                  Use this endpoint to invoke the agent programmatically
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="p-4 bg-muted rounded-lg font-mono text-sm">
                  <p className="text-muted-foreground mb-2"># Run this agent</p>
                  <p>POST /api/v1/agents/{id}/run</p>
                  <p className="text-muted-foreground mt-4 mb-2"># Request body</p>
                  <pre className="text-xs">{`{
  "prompt": "Your task description",
  "workspace": "optional-workspace-name",
  "options": {
    "max_turns": 10,
    "timeout": 300
  }
}`}</pre>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </form>
    </div>
  )
}
