import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  MessageSquare,
  Plus,
  Trash2,
  Settings,
  Activity,
  Users,
  MessageCircle,
  ArrowDownLeft,
  ArrowUpRight,
  Clock,
  StopCircle,
  ExternalLink,
} from 'lucide-react'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api, FeishuConfig } from '@/services/api'
import type { ChannelSession, ChannelMessage } from '@/types'

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
  const [sessionFilter, setSessionFilter] = useState<{ status?: string; channel_type?: string }>({})
  const [messageFilter, setMessageFilter] = useState<{ direction?: string; channel_type?: string }>({})

  // Queries
  const { data: feishuConfigs = [], isLoading: configsLoading } = useQuery({
    queryKey: ['feishuConfigs'],
    queryFn: api.listFeishuConfigs,
  })

  const { data: agents = [] } = useQuery({
    queryKey: ['agents'],
    queryFn: api.listAgents,
  })

  const { data: stats } = useQuery({
    queryKey: ['channelStats'],
    queryFn: () => api.getChannelStats(),
    refetchInterval: 30000,
  })

  const { data: sessionsData, isLoading: sessionsLoading } = useQuery({
    queryKey: ['channelSessions', sessionFilter],
    queryFn: () => api.listChannelSessions({
      status: sessionFilter.status as 'active' | 'completed' | 'expired' | undefined,
      channel_type: sessionFilter.channel_type,
      limit: 50,
    }),
  })

  const { data: messagesData, isLoading: messagesLoading } = useQuery({
    queryKey: ['channelMessages', messageFilter],
    queryFn: () => api.listChannelMessages({
      direction: messageFilter.direction as 'inbound' | 'outbound' | undefined,
      channel_type: messageFilter.channel_type,
      limit: 50,
    }),
  })

  // Mutations
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

  const endSessionMutation = useMutation({
    mutationFn: api.endChannelSession,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channelSessions'] })
      queryClient.invalidateQueries({ queryKey: ['channelStats'] })
      toast.success('Session ended')
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
      app_secret: '',
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

  const formatTime = (dateStr: string | undefined) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleString()
  }

  const sessions = sessionsData?.sessions || []
  const messages = messagesData?.messages || []

  return (
    <Main>
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Channels</h1>
        <p className="text-muted-foreground">Manage messaging channels and view conversation history</p>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="feishu">Feishu</TabsTrigger>
          <TabsTrigger value="sessions">Sessions</TabsTrigger>
          <TabsTrigger value="messages">Messages</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Sessions</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.total_sessions || 0}</div>
                <p className="text-xs text-muted-foreground">
                  {stats?.active_sessions || 0} active
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.active_sessions || 0}</div>
                <p className="text-xs text-muted-foreground">Currently active</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Messages</CardTitle>
                <MessageCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.total_messages || 0}</div>
                <p className="text-xs text-muted-foreground">All time</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Messages Today</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats?.messages_today || 0}</div>
                <p className="text-xs text-muted-foreground">Last 24 hours</p>
              </CardContent>
            </Card>
          </div>

          {/* By Channel Stats */}
          {stats?.by_channel && Object.keys(stats.by_channel).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">By Channel</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-3">
                  {Object.entries(stats.by_channel).map(([channel, stat]) => (
                    <div key={channel} className="flex items-center justify-between p-4 border rounded-lg">
                      <div>
                        <p className="font-medium capitalize">{channel}</p>
                        <p className="text-sm text-muted-foreground">
                          {stat.sessions} sessions, {stat.messages} messages
                        </p>
                      </div>
                      <MessageSquare className="h-8 w-8 text-muted-foreground" />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Feishu Tab */}
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

          {configsLoading ? (
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
                      <Button variant="outline" size="sm" onClick={() => handleEdit(config)}>
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

        {/* Sessions Tab */}
        <TabsContent value="sessions" className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-lg font-semibold">Channel Sessions</h2>
              <p className="text-sm text-muted-foreground">
                View and manage active conversations
              </p>
            </div>
            <div className="flex gap-2">
              <Select
                value={sessionFilter.status || 'all'}
                onValueChange={(v) => setSessionFilter({ ...sessionFilter, status: v === 'all' ? undefined : v })}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="expired">Expired</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={sessionFilter.channel_type || 'all'}
                onValueChange={(v) => setSessionFilter({ ...sessionFilter, channel_type: v === 'all' ? undefined : v })}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Channel" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Channels</SelectItem>
                  <SelectItem value="feishu">Feishu</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Channel</TableHead>
                    <TableHead>Agent</TableHead>
                    <TableHead>Messages</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Last Active</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sessionsLoading ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                        Loading...
                      </TableCell>
                    </TableRow>
                  ) : sessions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                        No sessions found
                      </TableCell>
                    </TableRow>
                  ) : (
                    sessions.map((session: ChannelSession) => (
                      <TableRow key={session.id}>
                        <TableCell>
                          <div>
                            <p className="font-medium">{session.user_name || session.user_id}</p>
                            {session.is_group && (
                              <Badge variant="outline" className="text-xs">Group</Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="capitalize">
                            {session.channel_type}
                          </Badge>
                        </TableCell>
                        <TableCell>{session.agent_name || session.agent_id?.slice(0, 8)}</TableCell>
                        <TableCell>{session.message_count}</TableCell>
                        <TableCell>
                          <Badge
                            variant={
                              session.status === 'active'
                                ? 'default'
                                : session.status === 'completed'
                                  ? 'secondary'
                                  : 'destructive'
                            }
                          >
                            {session.status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatTime(session.last_message_at)}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            {session.task_id && (
                              <a href={`/tasks/${session.task_id}`}>
                                <Button variant="ghost" size="sm" title="View Task">
                                  <ExternalLink className="h-4 w-4" />
                                </Button>
                              </a>
                            )}
                            {session.status === 'active' && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => endSessionMutation.mutate(session.id)}
                                title="End Session"
                              >
                                <StopCircle className="h-4 w-4 text-destructive" />
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Messages Tab */}
        <TabsContent value="messages" className="space-y-4">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-lg font-semibold">Message Log</h2>
              <p className="text-sm text-muted-foreground">View all channel messages</p>
            </div>
            <div className="flex gap-2">
              <Select
                value={messageFilter.direction || 'all'}
                onValueChange={(v) => setMessageFilter({ ...messageFilter, direction: v === 'all' ? undefined : v })}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Direction" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="inbound">Inbound</SelectItem>
                  <SelectItem value="outbound">Outbound</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={messageFilter.channel_type || 'all'}
                onValueChange={(v) => setMessageFilter({ ...messageFilter, channel_type: v === 'all' ? undefined : v })}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Channel" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Channels</SelectItem>
                  <SelectItem value="feishu">Feishu</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[50px]">Dir</TableHead>
                    <TableHead>Sender</TableHead>
                    <TableHead>Content</TableHead>
                    <TableHead>Channel</TableHead>
                    <TableHead>Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {messagesLoading ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        Loading...
                      </TableCell>
                    </TableRow>
                  ) : messages.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        No messages found
                      </TableCell>
                    </TableRow>
                  ) : (
                    messages.map((msg: ChannelMessage) => (
                      <TableRow key={msg.id}>
                        <TableCell>
                          {msg.direction === 'inbound' ? (
                            <span title="Inbound">
                              <ArrowDownLeft className="h-4 w-4 text-blue-500" />
                            </span>
                          ) : (
                            <span title="Outbound">
                              <ArrowUpRight className="h-4 w-4 text-green-500" />
                            </span>
                          )}
                        </TableCell>
                        <TableCell>
                          <p className="font-medium">{msg.sender_name || msg.sender_id}</p>
                        </TableCell>
                        <TableCell>
                          <p className="max-w-[400px] truncate">{msg.content}</p>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="capitalize">
                            {msg.channel_type}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatTime(msg.received_at)}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
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
            <DialogDescription>Configure your Feishu application credentials</DialogDescription>
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
