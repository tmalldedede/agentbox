import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  Save,
  ArrowLeft,
  Server,
  Zap,
  Shield,
  Cpu,
  MessageSquare,
  Settings2,
} from 'lucide-react'
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
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'sonner'
import { api } from '@/services/api'
import type { Profile, MCPServer, Skill, Credential } from '@/types'

interface ProfileFormProps {
  profile?: Profile
  isNew?: boolean
}

export function ProfileForm({ profile, isNew = false }: ProfileFormProps) {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [mcpServers, setMcpServers] = useState<MCPServer[]>([])
  const [skills, setSkills] = useState<Skill[]>([])
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [profiles, setProfiles] = useState<Profile[]>([])

  const [formData, setFormData] = useState({
    id: profile?.id || '',
    name: profile?.name || '',
    description: profile?.description || '',
    icon: profile?.icon || '',
    tags: profile?.tags?.join(', ') || '',
    adapter: profile?.adapter || 'claude-code',
    extends: profile?.extends || '',
    credential_id: profile?.credential_id || '',
    // Model config
    model_name: profile?.model?.name || 'sonnet',
    model_provider: profile?.model?.provider || '',
    model_base_url: profile?.model?.base_url || '',
    model_timeout_ms: profile?.model?.timeout_ms || 0,
    model_max_output_tokens: profile?.model?.max_output_tokens || 0,
    // MCP servers
    mcp_server_ids: profile?.mcp_servers?.map(s => s.name) || [],
    // Skills
    skill_ids: profile?.skill_ids || [],
    // Permissions - Claude Code
    permission_mode: profile?.permissions?.mode || 'default',
    allowed_tools: profile?.permissions?.allowed_tools?.join(', ') || '',
    disallowed_tools: profile?.permissions?.disallowed_tools?.join(', ') || '',
    skip_all: profile?.permissions?.skip_all || false,
    // Permissions - Codex
    sandbox_mode: profile?.permissions?.sandbox_mode || 'workspace-write',
    approval_policy: profile?.permissions?.approval_policy || 'on-failure',
    full_auto: profile?.permissions?.full_auto || false,
    // Resources
    max_budget_usd: profile?.resources?.max_budget_usd || 0,
    max_turns: profile?.resources?.max_turns || 0,
    max_tokens: profile?.resources?.max_tokens || 0,
    timeout: profile?.resources?.timeout || 0,
    cpus: profile?.resources?.cpus || 2,
    memory_mb: profile?.resources?.memory_mb || 4096,
    disk_gb: profile?.resources?.disk_gb || 10,
    // Prompts
    system_prompt: profile?.system_prompt || '',
    append_system_prompt: profile?.append_system_prompt || '',
    base_instructions: profile?.base_instructions || '',
    developer_instructions: profile?.developer_instructions || '',
    // Features
    web_search: profile?.features?.web_search || false,
  })

  useEffect(() => {
    loadRelatedData()
  }, [])

  const loadRelatedData = async () => {
    try {
      const [mcpList, skillList, credList, profileList] = await Promise.all([
        api.listMCPServers(),
        api.listSkills(),
        api.listCredentials(),
        api.listProfiles(),
      ])
      setMcpServers(mcpList)
      setSkills(skillList)
      setCredentials(credList)
      setProfiles(profileList.filter((p: Profile) => p.id !== profile?.id))
    } catch (error) {
      console.error('Failed to load related data:', error)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const payload = {
        id: formData.id,
        name: formData.name,
        description: formData.description,
        icon: formData.icon,
        tags: formData.tags.split(',').map(t => t.trim()).filter(Boolean),
        adapter: formData.adapter as 'claude-code' | 'codex' | 'opencode',
        extends: formData.extends || undefined,
        credential_id: formData.credential_id || undefined,
        model: {
          name: formData.model_name,
          provider: formData.model_provider || undefined,
          base_url: formData.model_base_url || undefined,
          timeout_ms: formData.model_timeout_ms || undefined,
          max_output_tokens: formData.model_max_output_tokens || undefined,
        },
        mcp_servers: formData.mcp_server_ids.map(name => {
          const server = mcpServers.find(s => s.name === name)
          return server ? { name: server.name, command: server.command, args: server.args } : null
        }).filter(Boolean) as { name: string; command: string; args?: string[] }[],
        skill_ids: formData.skill_ids,
        permissions: formData.adapter === 'claude-code' ? {
          mode: formData.permission_mode,
          allowed_tools: formData.allowed_tools.split(',').map(t => t.trim()).filter(Boolean),
          disallowed_tools: formData.disallowed_tools.split(',').map(t => t.trim()).filter(Boolean),
          skip_all: formData.skip_all,
        } : {
          sandbox_mode: formData.sandbox_mode as 'read-only' | 'workspace-write' | 'danger-full-access',
          approval_policy: formData.approval_policy as 'untrusted' | 'on-failure' | 'on-request' | 'never',
          full_auto: formData.full_auto,
        },
        resources: {
          max_budget_usd: formData.max_budget_usd || undefined,
          max_turns: formData.max_turns || undefined,
          max_tokens: formData.max_tokens || undefined,
          timeout: formData.timeout || undefined,
          cpus: formData.cpus,
          memory_mb: formData.memory_mb,
          disk_gb: formData.disk_gb,
        },
        system_prompt: formData.system_prompt || undefined,
        append_system_prompt: formData.append_system_prompt || undefined,
        base_instructions: formData.base_instructions || undefined,
        developer_instructions: formData.developer_instructions || undefined,
        features: {
          web_search: formData.web_search,
        },
      }

      if (isNew) {
        await api.createProfile(payload)
        toast.success('Profile created successfully')
      } else {
        await api.updateProfile(profile!.id, payload)
        toast.success('Profile updated successfully')
      }

      navigate({ to: '/profiles' })
    } catch (error: unknown) {
      const err = error as { message?: string }
      toast.error(err.message || 'Failed to save profile')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => navigate({ to: '/profiles' })}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">
              {isNew ? 'Create Profile' : 'Edit Profile'}
            </h1>
            <p className="text-muted-foreground">
              {isNew
                ? 'Create a new agent runtime configuration'
                : 'Modify agent runtime configuration'}
            </p>
          </div>
        </div>
        <Button type="submit" disabled={loading}>
          <Save className="mr-2 h-4 w-4" />
          {loading ? 'Saving...' : 'Save'}
        </Button>
      </div>

      <Tabs defaultValue="basic" className="space-y-4">
        <TabsList className="grid w-full grid-cols-6">
          <TabsTrigger value="basic">
            <Settings2 className="mr-2 h-4 w-4" />
            Basic
          </TabsTrigger>
          <TabsTrigger value="model">
            <Cpu className="mr-2 h-4 w-4" />
            Model
          </TabsTrigger>
          <TabsTrigger value="mcp">
            <Server className="mr-2 h-4 w-4" />
            MCP
          </TabsTrigger>
          <TabsTrigger value="skills">
            <Zap className="mr-2 h-4 w-4" />
            Skills
          </TabsTrigger>
          <TabsTrigger value="permissions">
            <Shield className="mr-2 h-4 w-4" />
            Permissions
          </TabsTrigger>
          <TabsTrigger value="prompts">
            <MessageSquare className="mr-2 h-4 w-4" />
            Prompts
          </TabsTrigger>
        </TabsList>

        {/* Basic Tab */}
        <TabsContent value="basic">
          <Card>
            <CardHeader>
              <CardTitle>Basic Information</CardTitle>
              <CardDescription>
                Profile identification and adapter selection
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="id">Profile ID</Label>
                  <Input
                    id="id"
                    value={formData.id}
                    onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                    placeholder="my-profile"
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
                    placeholder="My Profile"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Profile description..."
                  rows={3}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="adapter">Adapter</Label>
                  <Select
                    value={formData.adapter}
                    onValueChange={(value: 'claude-code' | 'codex' | 'opencode') => setFormData({ ...formData, adapter: value })}
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
                  <Label htmlFor="extends">Extends</Label>
                  <Select
                    value={formData.extends}
                    onValueChange={(value) => setFormData({ ...formData, extends: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="None" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">None</SelectItem>
                      {profiles.map((p) => (
                        <SelectItem key={p.id} value={p.id}>
                          {p.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="tags">Tags</Label>
                  <Input
                    id="tags"
                    value={formData.tags}
                    onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
                    placeholder="coding, review, security"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="credential">Credential</Label>
                  <Select
                    value={formData.credential_id}
                    onValueChange={(value) => setFormData({ ...formData, credential_id: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select credential" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">None</SelectItem>
                      {credentials.map((c) => (
                        <SelectItem key={c.id} value={c.id}>
                          {c.name} ({c.provider})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Model Tab */}
        <TabsContent value="model">
          <Card>
            <CardHeader>
              <CardTitle>Model Configuration</CardTitle>
              <CardDescription>
                Configure the AI model settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="model_name">Model Name</Label>
                  <Select
                    value={formData.model_name}
                    onValueChange={(value) => setFormData({ ...formData, model_name: value })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="sonnet">Sonnet</SelectItem>
                      <SelectItem value="opus">Opus</SelectItem>
                      <SelectItem value="haiku">Haiku</SelectItem>
                      <SelectItem value="o3">O3</SelectItem>
                      <SelectItem value="o4-mini">O4 Mini</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="model_provider">Provider</Label>
                  <Input
                    id="model_provider"
                    value={formData.model_provider}
                    onChange={(e) => setFormData({ ...formData, model_provider: e.target.value })}
                    placeholder="anthropic"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="model_base_url">Base URL</Label>
                <Input
                  id="model_base_url"
                  value={formData.model_base_url}
                  onChange={(e) => setFormData({ ...formData, model_base_url: e.target.value })}
                  placeholder="https://api.anthropic.com"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="model_timeout_ms">Timeout (ms)</Label>
                  <Input
                    id="model_timeout_ms"
                    type="number"
                    value={formData.model_timeout_ms || ''}
                    onChange={(e) => setFormData({ ...formData, model_timeout_ms: parseInt(e.target.value) || 0 })}
                    placeholder="60000"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="model_max_output_tokens">Max Output Tokens</Label>
                  <Input
                    id="model_max_output_tokens"
                    type="number"
                    value={formData.model_max_output_tokens || ''}
                    onChange={(e) => setFormData({ ...formData, model_max_output_tokens: parseInt(e.target.value) || 0 })}
                    placeholder="16000"
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* MCP Tab */}
        <TabsContent value="mcp">
          <Card>
            <CardHeader>
              <CardTitle>MCP Servers</CardTitle>
              <CardDescription>
                Select MCP servers to enable for this profile
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                {mcpServers.map((server) => (
                  <div
                    key={server.id}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      formData.mcp_server_ids.includes(server.name)
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                    onClick={() => {
                      const ids = formData.mcp_server_ids.includes(server.name)
                        ? formData.mcp_server_ids.filter(id => id !== server.name)
                        : [...formData.mcp_server_ids, server.name]
                      setFormData({ ...formData, mcp_server_ids: ids })
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <Server className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="font-medium">{server.name}</p>
                        <p className="text-sm text-muted-foreground">{server.description}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              {mcpServers.length === 0 && (
                <p className="text-center text-muted-foreground py-8">
                  No MCP servers available
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Skills Tab */}
        <TabsContent value="skills">
          <Card>
            <CardHeader>
              <CardTitle>Skills</CardTitle>
              <CardDescription>
                Select skills to enable for this profile
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                {skills.map((skill) => (
                  <div
                    key={skill.id}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      formData.skill_ids.includes(skill.id)
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                    onClick={() => {
                      const ids = formData.skill_ids.includes(skill.id)
                        ? formData.skill_ids.filter(id => id !== skill.id)
                        : [...formData.skill_ids, skill.id]
                      setFormData({ ...formData, skill_ids: ids })
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <Zap className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="font-medium">{skill.name}</p>
                        <p className="text-sm text-muted-foreground">{skill.description}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              {skills.length === 0 && (
                <p className="text-center text-muted-foreground py-8">
                  No skills available
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Permissions Tab */}
        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <CardTitle>Permissions</CardTitle>
              <CardDescription>
                Configure security and permission settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {formData.adapter === 'claude-code' ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="permission_mode">Permission Mode</Label>
                    <Select
                      value={formData.permission_mode}
                      onValueChange={(value) => setFormData({ ...formData, permission_mode: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="default">Default</SelectItem>
                        <SelectItem value="acceptEdits">Accept Edits</SelectItem>
                        <SelectItem value="bypassPermissions">Bypass Permissions</SelectItem>
                        <SelectItem value="plan">Plan Mode</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="allowed_tools">Allowed Tools</Label>
                    <Input
                      id="allowed_tools"
                      value={formData.allowed_tools}
                      onChange={(e) => setFormData({ ...formData, allowed_tools: e.target.value })}
                      placeholder="Bash, Read, Write"
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label>Skip All Permissions</Label>
                      <p className="text-sm text-muted-foreground">
                        Dangerously skip all permission checks
                      </p>
                    </div>
                    <Switch
                      checked={formData.skip_all}
                      onCheckedChange={(checked) => setFormData({ ...formData, skip_all: checked })}
                    />
                  </div>
                </>
              ) : (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="sandbox_mode">Sandbox Mode</Label>
                    <Select
                      value={formData.sandbox_mode}
                      onValueChange={(value: 'read-only' | 'workspace-write' | 'danger-full-access') => setFormData({ ...formData, sandbox_mode: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="read-only">Read Only</SelectItem>
                        <SelectItem value="workspace-write">Workspace Write</SelectItem>
                        <SelectItem value="danger-full-access">Danger Full Access</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="approval_policy">Approval Policy</Label>
                    <Select
                      value={formData.approval_policy}
                      onValueChange={(value: 'untrusted' | 'on-failure' | 'on-request' | 'never') => setFormData({ ...formData, approval_policy: value })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="untrusted">Untrusted</SelectItem>
                        <SelectItem value="on-failure">On Failure</SelectItem>
                        <SelectItem value="on-request">On Request</SelectItem>
                        <SelectItem value="never">Never</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label>Full Auto Mode</Label>
                      <p className="text-sm text-muted-foreground">
                        Enable full automation without confirmations
                      </p>
                    </div>
                    <Switch
                      checked={formData.full_auto}
                      onCheckedChange={(checked) => setFormData({ ...formData, full_auto: checked })}
                    />
                  </div>
                </>
              )}

              {/* Resource Limits */}
              <div className="pt-4 border-t">
                <h4 className="font-medium mb-4">Resource Limits</h4>
                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="cpus">CPUs</Label>
                    <Input
                      id="cpus"
                      type="number"
                      value={formData.cpus}
                      onChange={(e) => setFormData({ ...formData, cpus: parseFloat(e.target.value) || 2 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="memory_mb">Memory (MB)</Label>
                    <Input
                      id="memory_mb"
                      type="number"
                      value={formData.memory_mb}
                      onChange={(e) => setFormData({ ...formData, memory_mb: parseInt(e.target.value) || 4096 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="disk_gb">Disk (GB)</Label>
                    <Input
                      id="disk_gb"
                      type="number"
                      value={formData.disk_gb}
                      onChange={(e) => setFormData({ ...formData, disk_gb: parseInt(e.target.value) || 10 })}
                    />
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Prompts Tab */}
        <TabsContent value="prompts">
          <Card>
            <CardHeader>
              <CardTitle>System Prompts</CardTitle>
              <CardDescription>
                Configure system prompts and instructions
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="system_prompt">System Prompt</Label>
                <Textarea
                  id="system_prompt"
                  value={formData.system_prompt}
                  onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
                  placeholder="You are a helpful assistant..."
                  rows={4}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="append_system_prompt">Append System Prompt</Label>
                <Textarea
                  id="append_system_prompt"
                  value={formData.append_system_prompt}
                  onChange={(e) => setFormData({ ...formData, append_system_prompt: e.target.value })}
                  placeholder="Additional instructions..."
                  rows={3}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="base_instructions">Base Instructions</Label>
                <Textarea
                  id="base_instructions"
                  value={formData.base_instructions}
                  onChange={(e) => setFormData({ ...formData, base_instructions: e.target.value })}
                  placeholder="Base instructions for the agent..."
                  rows={3}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="developer_instructions">Developer Instructions</Label>
                <Textarea
                  id="developer_instructions"
                  value={formData.developer_instructions}
                  onChange={(e) => setFormData({ ...formData, developer_instructions: e.target.value })}
                  placeholder="Developer-specific instructions..."
                  rows={3}
                />
              </div>

              <div className="flex items-center justify-between pt-4 border-t">
                <div>
                  <Label>Web Search</Label>
                  <p className="text-sm text-muted-foreground">
                    Enable web search capability
                  </p>
                </div>
                <Switch
                  checked={formData.web_search}
                  onCheckedChange={(checked) => setFormData({ ...formData, web_search: checked })}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </form>
  )
}

export default ProfileForm
