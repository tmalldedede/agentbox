import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, Plus, Trash2, Settings } from 'lucide-react'
import { toast } from 'sonner'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import { api, FeishuConfig } from '@/services/api'

export const Route = createFileRoute('/_authenticated/channels/')({
  component: ChannelsPage,
})

function ChannelsPage() {
  const queryClient = useQueryClient()
  const [showConfigDialog, setShowConfigDialog] = useState(false)
  const [editingConfig, setEditingConfig] = useState<FeishuConfig | null>(null)
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    app_id: '',
    app_secret: '',
    encrypt_key: '',
    verification_token: '',
    default_agent_id: '',
  })

  const { data: feishuConfigs = [], isLoading } = useQuery({
    queryKey: ['feishuConfigs'],
    queryFn: api.listFeishuConfigs,
  })

  const { data: agents = [] } = useQuery({
    queryKey: ['agents'],
    queryFn: api.listAgents,
  })

  const saveMutation = useMutation({
    mutationFn: (data: typeof formData) => api.saveFeishuConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feishuConfigs'] })
      setShowConfigDialog(false)
      resetForm()
      toast.success(editingConfig ? 'Configuration updated' : 'Configuration created')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: api.deleteFeishuConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feishuConfigs'] })
      toast.success('Configuration deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const resetForm = () => {
    setFormData({
      id: '',
      name: '',
      app_id: '',
      app_secret: '',
      encrypt_key: '',
      verification_token: '',
      default_agent_id: '',
    })
    setEditingConfig(null)
  }

  const handleEdit = (config: FeishuConfig) => {
    setEditingConfig(config)
    setFormData({
      id: config.id,
      name: config.name,
      app_id: config.app_id,
      app_secret: '', // Don't show existing secret
      encrypt_key: '',
      verification_token: '',
      default_agent_id: config.default_agent_id || '',
    })
    setShowConfigDialog(true)
  }

  const handleCreate = () => {
    resetForm()
    setShowConfigDialog(true)
  }

  const handleSave = () => {
    if (!formData.name || !formData.app_id) {
      toast.error('Please fill required fields')
      return
    }
    if (!editingConfig && !formData.app_secret) {
      toast.error('App Secret is required for new configuration')
      return
    }
    saveMutation.mutate(formData)
  }

  return (
    <Main>
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Channels</h1>
        <p className="text-muted-foreground">Configure messaging channels for agent interaction</p>
      </div>

      <Tabs defaultValue="feishu" className="space-y-4">
        <TabsList>
          <TabsTrigger value="feishu">Feishu / Lark</TabsTrigger>
          <TabsTrigger value="slack" disabled>Slack (Coming Soon)</TabsTrigger>
          <TabsTrigger value="discord" disabled>Discord (Coming Soon)</TabsTrigger>
        </TabsList>

        <TabsContent value="feishu" className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-lg font-semibold">Feishu Applications</h2>
              <p className="text-sm text-muted-foreground">
                Configure Feishu bot applications to receive and respond to messages
              </p>
            </div>
            <Button onClick={handleCreate}>
              <Plus className="mr-2 h-4 w-4" />
              Add Application
            </Button>
          </div>

          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading...</div>
          ) : feishuConfigs.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <MessageSquare className="h-12 w-12 text-muted-foreground mb-4" />
                <h3 className="text-lg font-medium mb-2">No Feishu applications configured</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Add a Feishu application to enable bot messaging
                </p>
                <Button onClick={handleCreate}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Application
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {feishuConfigs.map((config) => (
                <Card key={config.id}>
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">{config.name}</CardTitle>
                      <Badge variant={config.enabled ? 'default' : 'secondary'}>
                        {config.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </div>
                    <CardDescription className="font-mono text-xs">
                      App ID: {config.app_id}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-sm text-muted-foreground mb-4">
                      {config.default_agent_id ? (
                        <span>Default Agent: {config.default_agent_id.slice(0, 8)}...</span>
                      ) : (
                        <span className="text-yellow-600">No default agent set</span>
                      )}
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleEdit(config)}
                      >
                        <Settings className="h-4 w-4 mr-1" />
                        Configure
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => deleteMutation.mutate(config.id)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}

          {/* Webhook URL Info */}
          <Card className="mt-6">
            <CardHeader>
              <CardTitle className="text-base">Webhook Configuration</CardTitle>
              <CardDescription>
                Configure these URLs in your Feishu application settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <div>
                <Label className="text-xs text-muted-foreground">Event Callback URL</Label>
                <code className="block text-sm bg-muted px-2 py-1 rounded mt-1">
                  {window.location.origin}/api/v1/webhooks/feishu
                </code>
              </div>
              <p className="text-xs text-muted-foreground">
                Set this URL in Feishu Open Platform &gt; Event Subscription &gt; Request URL
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Config Dialog */}
      <Dialog open={showConfigDialog} onOpenChange={setShowConfigDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingConfig ? 'Edit Feishu Application' : 'Add Feishu Application'}
            </DialogTitle>
            <DialogDescription>
              Configure your Feishu application credentials
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="My Feishu Bot"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="app_id">App ID *</Label>
              <Input
                id="app_id"
                value={formData.app_id}
                onChange={(e) => setFormData({ ...formData, app_id: e.target.value })}
                placeholder="cli_xxxxx"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="app_secret">
                App Secret {editingConfig ? '(leave empty to keep existing)' : '*'}
              </Label>
              <Input
                id="app_secret"
                type="password"
                value={formData.app_secret}
                onChange={(e) => setFormData({ ...formData, app_secret: e.target.value })}
                placeholder="••••••••"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="encrypt_key">Encrypt Key (optional)</Label>
              <Input
                id="encrypt_key"
                type="password"
                value={formData.encrypt_key}
                onChange={(e) => setFormData({ ...formData, encrypt_key: e.target.value })}
                placeholder="For encrypted events"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="verification_token">Verification Token (optional)</Label>
              <Input
                id="verification_token"
                value={formData.verification_token}
                onChange={(e) => setFormData({ ...formData, verification_token: e.target.value })}
                placeholder="For event verification"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="default_agent">Default Agent</Label>
              <select
                id="default_agent"
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
                value={formData.default_agent_id}
                onChange={(e) => setFormData({ ...formData, default_agent_id: e.target.value })}
              >
                <option value="">Select an agent</option>
                {agents.map((agent) => (
                  <option key={agent.id} value={agent.id}>
                    {agent.name}
                  </option>
                ))}
              </select>
              <p className="text-xs text-muted-foreground">
                The agent that will handle incoming messages
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowConfigDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Main>
  )
}
