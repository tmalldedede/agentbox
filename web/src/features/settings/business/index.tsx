import { useState } from 'react'
import { useSettings, useUpdateSettings, useResetSettings } from '@/hooks/useSettings'
import { useProviders, useRuntimes } from '@/hooks'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Loader2, Save, RotateCcw, Bot, ListTodo, Layers, Database, Bell } from 'lucide-react'
import { toast } from 'sonner'
import type { Settings } from '@/types'

export function BusinessSettings() {
  const { data: settings, isLoading } = useSettings()
  const { data: providers } = useProviders()
  const { data: runtimes } = useRuntimes()
  const updateSettings = useUpdateSettings()
  const resetSettings = useResetSettings()

  const [form, setForm] = useState<Settings | null>(null)

  // Initialize form when settings load
  if (settings && !form) {
    setForm(settings)
  }

  const handleSave = async () => {
    if (!form) return
    try {
      await updateSettings.mutateAsync(form)
      toast.success('Settings saved successfully')
    } catch (error) {
      toast.error('Failed to save settings')
    }
  }

  const handleReset = async () => {
    try {
      const defaults = await resetSettings.mutateAsync()
      setForm(defaults)
      toast.success('Settings reset to defaults')
    } catch (error) {
      toast.error('Failed to reset settings')
    }
  }

  if (isLoading || !form) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Settings</h2>
          <p className="text-muted-foreground">
            Configure default behaviors and system parameters
          </p>
        </div>
        <div className="flex items-center gap-2">
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="outline">
                <RotateCcw className="w-4 h-4 mr-2" />
                Reset
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Reset to Defaults?</AlertDialogTitle>
                <AlertDialogDescription>
                  This will reset all settings to their default values. This action cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={handleReset}>Reset</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
          <Button onClick={handleSave} disabled={updateSettings.isPending}>
            {updateSettings.isPending ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            Save Changes
          </Button>
        </div>
      </div>

      <Tabs defaultValue="agent" className="space-y-4">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="agent" className="flex items-center gap-2">
            <Bot className="w-4 h-4" />
            Agent
          </TabsTrigger>
          <TabsTrigger value="task" className="flex items-center gap-2">
            <ListTodo className="w-4 h-4" />
            Task
          </TabsTrigger>
          <TabsTrigger value="batch" className="flex items-center gap-2">
            <Layers className="w-4 h-4" />
            Batch
          </TabsTrigger>
          <TabsTrigger value="storage" className="flex items-center gap-2">
            <Database className="w-4 h-4" />
            Storage
          </TabsTrigger>
          <TabsTrigger value="notify" className="flex items-center gap-2">
            <Bell className="w-4 h-4" />
            Notify
          </TabsTrigger>
        </TabsList>

        {/* Agent Settings */}
        <TabsContent value="agent">
          <Card>
            <CardHeader>
              <CardTitle>Agent Defaults</CardTitle>
              <CardDescription>
                Default configuration for new agents and tasks
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label>Default Provider</Label>
                  <Select
                    value={form.agent.default_provider_id || 'none'}
                    onValueChange={(v) =>
                      setForm({
                        ...form,
                        agent: { ...form.agent, default_provider_id: v === 'none' ? '' : v },
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select provider" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">None</SelectItem>
                      {providers?.map((p) => (
                        <SelectItem key={p.id} value={p.id}>
                          {p.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Default Model</Label>
                  <Input
                    value={form.agent.default_model}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        agent: { ...form.agent, default_model: e.target.value },
                      })
                    }
                    placeholder="e.g., claude-3-5-sonnet-20241022"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Default Runtime</Label>
                  <Select
                    value={form.agent.default_runtime_id || 'none'}
                    onValueChange={(v) =>
                      setForm({
                        ...form,
                        agent: { ...form.agent, default_runtime_id: v === 'none' ? '' : v },
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select runtime" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">None</SelectItem>
                      {runtimes?.map((r) => (
                        <SelectItem key={r.id} value={r.id}>
                          {r.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Default Timeout (seconds)</Label>
                  <Input
                    type="number"
                    value={form.agent.default_timeout}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        agent: { ...form.agent, default_timeout: parseInt(e.target.value) || 3600 },
                      })
                    }
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label>Default System Prompt</Label>
                <Textarea
                  value={form.agent.system_prompt}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      agent: { ...form.agent, system_prompt: e.target.value },
                    })
                  }
                  placeholder="Enter default system prompt for agents..."
                  rows={4}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Task Settings */}
        <TabsContent value="task">
          <Card>
            <CardHeader>
              <CardTitle>Task Configuration</CardTitle>
              <CardDescription>
                Default parameters for task execution
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label>Default Idle Timeout (seconds)</Label>
                  <Input
                    type="number"
                    value={form.task.default_idle_timeout}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        task: { ...form.task, default_idle_timeout: parseInt(e.target.value) || 30 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Time to wait for agent output before completing task
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Default Poll Interval (ms)</Label>
                  <Input
                    type="number"
                    value={form.task.default_poll_interval}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        task: { ...form.task, default_poll_interval: parseInt(e.target.value) || 500 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Interval for checking agent output
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Max Turns</Label>
                  <Input
                    type="number"
                    value={form.task.max_turns}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        task: { ...form.task, max_turns: parseInt(e.target.value) || 100 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Maximum conversation turns per task
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Max Attachments</Label>
                  <Input
                    type="number"
                    value={form.task.max_attachments}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        task: { ...form.task, max_attachments: parseInt(e.target.value) || 10 },
                      })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Max Attachment Size (MB)</Label>
                  <Input
                    type="number"
                    value={form.task.max_attachment_size}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        task: { ...form.task, max_attachment_size: parseInt(e.target.value) || 100 },
                      })
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Batch Settings */}
        <TabsContent value="batch">
          <Card>
            <CardHeader>
              <CardTitle>Batch Processing</CardTitle>
              <CardDescription>
                Worker pool and retry configuration
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label>Default Workers</Label>
                  <Input
                    type="number"
                    value={form.batch.default_workers}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, default_workers: parseInt(e.target.value) || 5 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Default number of concurrent workers per batch
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Max Workers</Label>
                  <Input
                    type="number"
                    value={form.batch.max_workers}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, max_workers: parseInt(e.target.value) || 50 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Maximum workers allowed per batch
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Max Concurrent Batches</Label>
                  <Input
                    type="number"
                    value={form.batch.max_concurrent_batches}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, max_concurrent_batches: parseInt(e.target.value) || 10 },
                      })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label>Max Retries</Label>
                  <Input
                    type="number"
                    value={form.batch.max_retries}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, max_retries: parseInt(e.target.value) || 3 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Retries before moving to dead letter queue
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Retry Delay (seconds)</Label>
                  <Input
                    type="number"
                    value={form.batch.retry_delay}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, retry_delay: parseInt(e.target.value) || 5 },
                      })
                    }
                  />
                </div>
                <div className="flex items-center justify-between space-x-4 pt-6">
                  <div className="space-y-0.5">
                    <Label>Dead Letter Queue</Label>
                    <p className="text-xs text-muted-foreground">
                      Store failed tasks for manual review
                    </p>
                  </div>
                  <Switch
                    checked={form.batch.dead_letter_enabled}
                    onCheckedChange={(checked) =>
                      setForm({
                        ...form,
                        batch: { ...form.batch, dead_letter_enabled: checked },
                      })
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Storage Settings */}
        <TabsContent value="storage">
          <Card>
            <CardHeader>
              <CardTitle>Storage & Cleanup</CardTitle>
              <CardDescription>
                Data retention and automatic cleanup settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label>History Retention (days)</Label>
                  <Input
                    type="number"
                    value={form.storage.history_retention_days}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        storage: { ...form.storage, history_retention_days: parseInt(e.target.value) || 30 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    How long to keep execution history
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Session Retention (days)</Label>
                  <Input
                    type="number"
                    value={form.storage.session_retention_days}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        storage: { ...form.storage, session_retention_days: parseInt(e.target.value) || 7 },
                      })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    How long to keep inactive sessions
                  </p>
                </div>
              </div>
              <div className="flex items-center justify-between space-x-4">
                <div className="space-y-0.5">
                  <Label>Auto Cleanup</Label>
                  <p className="text-xs text-muted-foreground">
                    Automatically clean up expired data
                  </p>
                </div>
                <Switch
                  checked={form.storage.auto_cleanup}
                  onCheckedChange={(checked) =>
                    setForm({
                      ...form,
                      storage: { ...form.storage, auto_cleanup: checked },
                    })
                  }
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Notify Settings */}
        <TabsContent value="notify">
          <Card>
            <CardHeader>
              <CardTitle>Notifications</CardTitle>
              <CardDescription>
                Webhook and notification settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <Label>Webhook URL</Label>
                  <Input
                    value={form.notify.webhook_url}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        notify: { ...form.notify, webhook_url: e.target.value },
                      })
                    }
                    placeholder="https://your-webhook.example.com/hook"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Webhook Secret</Label>
                  <Input
                    type="password"
                    value={form.notify.webhook_secret}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        notify: { ...form.notify, webhook_secret: e.target.value },
                      })
                    }
                    placeholder="Secret for signing webhooks"
                  />
                </div>
              </div>
              <div className="space-y-4">
                <div className="flex items-center justify-between space-x-4">
                  <div className="space-y-0.5">
                    <Label>Notify on Task Complete</Label>
                    <p className="text-xs text-muted-foreground">
                      Send notification when a task completes successfully
                    </p>
                  </div>
                  <Switch
                    checked={form.notify.notify_on_complete}
                    onCheckedChange={(checked) =>
                      setForm({
                        ...form,
                        notify: { ...form.notify, notify_on_complete: checked },
                      })
                    }
                  />
                </div>
                <div className="flex items-center justify-between space-x-4">
                  <div className="space-y-0.5">
                    <Label>Notify on Task Failed</Label>
                    <p className="text-xs text-muted-foreground">
                      Send notification when a task fails
                    </p>
                  </div>
                  <Switch
                    checked={form.notify.notify_on_failed}
                    onCheckedChange={(checked) =>
                      setForm({
                        ...form,
                        notify: { ...form.notify, notify_on_failed: checked },
                      })
                    }
                  />
                </div>
                <div className="flex items-center justify-between space-x-4">
                  <div className="space-y-0.5">
                    <Label>Notify on Batch Complete</Label>
                    <p className="text-xs text-muted-foreground">
                      Send notification when a batch finishes
                    </p>
                  </div>
                  <Switch
                    checked={form.notify.notify_on_batch_complete}
                    onCheckedChange={(checked) =>
                      setForm({
                        ...form,
                        notify: { ...form.notify, notify_on_batch_complete: checked },
                      })
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
