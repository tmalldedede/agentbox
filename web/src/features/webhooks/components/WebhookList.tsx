import { useState } from 'react'
import {
  Plus,
  Trash2,
  CheckCircle,
  XCircle,
  Webhook as WebhookIcon,
  Globe,
  Activity,
  Power,
  PowerOff,
  MoreHorizontal,
  Pencil,
  Copy,
  Loader2,
  Shield,
} from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { toast } from 'sonner'
import {
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useDeleteWebhook,
} from '@/hooks/useWebhooks'
import type { Webhook, CreateWebhookRequest, UpdateWebhookRequest } from '@/types'

// 所有支持的事件类型
const ALL_EVENTS = [
  { value: 'task.created', label: 'Task Created', desc: 'When a new task is created' },
  { value: 'task.completed', label: 'Task Completed', desc: 'When a task finishes successfully' },
  { value: 'task.failed', label: 'Task Failed', desc: 'When a task encounters an error' },
  { value: 'session.started', label: 'Session Started', desc: 'When an agent session starts' },
  { value: 'session.stopped', label: 'Session Stopped', desc: 'When an agent session ends' },
]

const eventColorMap: Record<string, string> = {
  'task.created': 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  'task.completed': 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
  'task.failed': 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  'session.started': 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  'session.stopped': 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
}

interface WebhookFormData {
  url: string
  secret: string
  events: string[]
  is_active: boolean
}

function WebhookFormDialog({
  open,
  onOpenChange,
  webhook,
  onSubmit,
  loading,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  webhook?: Webhook | null
  onSubmit: (data: WebhookFormData) => void
  loading: boolean
}) {
  const isEdit = !!webhook
  const [form, setForm] = useState<WebhookFormData>({
    url: webhook?.url || '',
    secret: '',
    events: webhook?.events || [],
    is_active: webhook?.is_active ?? true,
  })

  const toggleEvent = (event: string) => {
    setForm((prev) => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter((e) => e !== event)
        : [...prev.events, event],
    }))
  }

  const selectAll = () => {
    setForm((prev) => ({
      ...prev,
      events: ALL_EVENTS.map((e) => e.value),
    }))
  }

  const clearAll = () => {
    setForm((prev) => ({ ...prev, events: [] }))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Webhook' : 'Create Webhook'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update webhook configuration. Leave secret empty to keep existing.'
              : 'Configure a new webhook endpoint to receive event notifications.'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5 py-4">
          {/* URL */}
          <div className="space-y-2">
            <Label htmlFor="webhook-url">
              Endpoint URL <span className="text-red-500">*</span>
            </Label>
            <Input
              id="webhook-url"
              placeholder="https://example.com/webhook"
              value={form.url}
              onChange={(e) => setForm((p) => ({ ...p, url: e.target.value }))}
            />
            <p className="text-xs text-muted-foreground">
              The URL where webhook payloads will be delivered via POST.
            </p>
          </div>

          {/* Secret */}
          <div className="space-y-2">
            <Label htmlFor="webhook-secret" className="flex items-center gap-1.5">
              <Shield className="h-3.5 w-3.5" />
              Secret (optional)
            </Label>
            <Input
              id="webhook-secret"
              type="password"
              placeholder={isEdit ? '(unchanged)' : 'your-webhook-secret'}
              value={form.secret}
              onChange={(e) => setForm((p) => ({ ...p, secret: e.target.value }))}
            />
            <p className="text-xs text-muted-foreground">
              Used to generate HMAC-SHA256 signature in <code className="text-xs">X-Webhook-Signature</code> header.
            </p>
          </div>

          {/* Events */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Events</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-6 text-xs"
                  onClick={selectAll}
                >
                  Select all
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-6 text-xs"
                  onClick={clearAll}
                >
                  Clear
                </Button>
              </div>
            </div>
            <div className="rounded-lg border divide-y">
              {ALL_EVENTS.map((evt) => (
                <label
                  key={evt.value}
                  className="flex items-center gap-3 px-4 py-3 hover:bg-muted/50 cursor-pointer transition-colors"
                >
                  <Checkbox
                    checked={form.events.includes(evt.value)}
                    onCheckedChange={() => toggleEvent(evt.value)}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <code className="text-xs font-mono">{evt.value}</code>
                    </div>
                    <p className="text-xs text-muted-foreground">{evt.desc}</p>
                  </div>
                </label>
              ))}
            </div>
            <p className="text-xs text-muted-foreground">
              Leave empty to subscribe to all events.
            </p>
          </div>

          {/* Active Toggle (edit only) */}
          {isEdit && (
            <div className="flex items-center justify-between rounded-lg border px-4 py-3">
              <div>
                <Label>Active</Label>
                <p className="text-xs text-muted-foreground">
                  Inactive webhooks will not receive events.
                </p>
              </div>
              <Switch
                checked={form.is_active}
                onCheckedChange={(v) => setForm((p) => ({ ...p, is_active: v }))}
              />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={() => onSubmit(form)}
            disabled={!form.url || loading}
          >
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {isEdit ? 'Update' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function WebhookList() {
  const { data: webhooks = [], isLoading } = useWebhooks()
  const createWebhook = useCreateWebhook()
  const updateWebhook = useUpdateWebhook()
  const deleteWebhook = useDeleteWebhook()

  const [createOpen, setCreateOpen] = useState(false)
  const [editWebhook, setEditWebhook] = useState<Webhook | null>(null)
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const handleCreate = (data: WebhookFormData) => {
    const req: CreateWebhookRequest = {
      url: data.url,
      events: data.events.length > 0 ? data.events : undefined,
    }
    if (data.secret) req.secret = data.secret
    createWebhook.mutate(req, {
      onSuccess: () => setCreateOpen(false),
    })
  }

  const handleUpdate = (data: WebhookFormData) => {
    if (!editWebhook) return
    const req: UpdateWebhookRequest = {
      url: data.url,
      events: data.events.length > 0 ? data.events : undefined,
      is_active: data.is_active,
    }
    if (data.secret) req.secret = data.secret
    updateWebhook.mutate(
      { id: editWebhook.id, data: req },
      { onSuccess: () => setEditWebhook(null) },
    )
  }

  const handleDelete = () => {
    if (!deleteId) return
    deleteWebhook.mutate(deleteId, {
      onSuccess: () => setDeleteId(null),
    })
  }

  const copyId = (id: string) => {
    navigator.clipboard.writeText(id)
    toast.success('Webhook ID copied')
  }

  const activeCount = webhooks.filter((w) => w.is_active).length

  if (isLoading) {
    return (
      <>
        <Header fixed className="md:hidden" />
        <Main className="flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </Main>
      </>
    )
  }

  return (
    <>
      <Header fixed className="md:hidden" />

      <Main className="flex flex-1 flex-col gap-4 sm:gap-6">
        {/* Header */}
        <div className="flex flex-wrap items-end justify-between gap-2">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Webhooks</h2>
            <p className="text-muted-foreground">
              Manage webhook endpoints for real-time event notifications.
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            New Webhook
          </Button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <Card>
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-500/10">
                <WebhookIcon className="h-5 w-5 text-blue-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{webhooks.length}</p>
                <p className="text-xs text-muted-foreground">Total Webhooks</p>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-green-500/10">
                <Activity className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{activeCount}</p>
                <p className="text-xs text-muted-foreground">Active</p>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-purple-500/10">
                <Globe className="h-5 w-5 text-purple-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">
                  {new Set(webhooks.flatMap((w) => w.events || [])).size}
                </p>
                <p className="text-xs text-muted-foreground">Event Types</p>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-amber-500/10">
                <Shield className="h-5 w-5 text-amber-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">
                  {webhooks.filter((w) => w.secret).length}
                </p>
                <p className="text-xs text-muted-foreground">With Secret</p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Webhook List */}
        <Card>
          <CardHeader>
            <CardTitle>Configured Webhooks</CardTitle>
            <CardDescription>
              {webhooks.length} webhook{webhooks.length !== 1 ? 's' : ''} configured
            </CardDescription>
          </CardHeader>
          <CardContent>
            {webhooks.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <WebhookIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
                <h3 className="font-semibold text-lg mb-1">No webhooks yet</h3>
                <p className="text-muted-foreground text-sm mb-4">
                  Create a webhook to receive real-time notifications for task events.
                </p>
                <Button size="sm" onClick={() => setCreateOpen(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create First Webhook
                </Button>
              </div>
            ) : (
              <div className="space-y-3">
                {webhooks.map((w) => (
                  <div
                    key={w.id}
                    className="group flex items-center gap-4 rounded-lg border p-4 hover:bg-muted/50 transition-colors"
                  >
                    {/* Status indicator */}
                    <div className="shrink-0">
                      {w.is_active ? (
                        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-green-500/10">
                          <Power className="h-4 w-4 text-green-500" />
                        </div>
                      ) : (
                        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-muted">
                          <PowerOff className="h-4 w-4 text-muted-foreground" />
                        </div>
                      )}
                    </div>

                    {/* Info */}
                    <div className="flex-1 min-w-0 space-y-1">
                      <div className="flex items-center gap-2">
                        <code className="text-sm font-mono truncate">{w.url}</code>
                        {w.is_active ? (
                          <Badge variant="outline" className="text-xs gap-1 bg-green-500/10 text-green-600 border-green-200 shrink-0">
                            <CheckCircle className="h-3 w-3" />
                            Active
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="text-xs gap-1 shrink-0">
                            <XCircle className="h-3 w-3" />
                            Inactive
                          </Badge>
                        )}
                      </div>
                      <div className="flex items-center gap-1.5 flex-wrap">
                        {(w.events && w.events.length > 0) ? (
                          w.events.map((e) => (
                            <Badge
                              key={e}
                              variant="secondary"
                              className={`text-[10px] ${eventColorMap[e] || ''}`}
                            >
                              {e}
                            </Badge>
                          ))
                        ) : (
                          <Badge variant="secondary" className="text-[10px]">
                            All events
                          </Badge>
                        )}
                      </div>
                    </div>

                    {/* Actions */}
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => copyId(w.id)}>
                          <Copy className="mr-2 h-4 w-4" />
                          Copy ID
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setEditWebhook(w)}>
                          <Pencil className="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => setDeleteId(w.id)}
                          className="text-destructive"
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </Main>

      {/* Create Dialog */}
      <WebhookFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSubmit={handleCreate}
        loading={createWebhook.isPending}
      />

      {/* Edit Dialog */}
      <WebhookFormDialog
        open={!!editWebhook}
        onOpenChange={(v) => { if (!v) setEditWebhook(null) }}
        webhook={editWebhook}
        onSubmit={handleUpdate}
        loading={updateWebhook.isPending}
      />

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Webhook?</AlertDialogTitle>
            <AlertDialogDescription>
              This webhook will be permanently deleted and will no longer receive event notifications.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

export default WebhookList
