import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useCreateRuntime, useUpdateRuntime } from '@/hooks/useRuntimes'
import type { AgentRuntime } from '@/types'

type RuntimeFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  runtime?: AgentRuntime | null
}

export function RuntimeFormDialog({
  open,
  onOpenChange,
  runtime,
}: RuntimeFormDialogProps) {
  const isEdit = !!runtime
  const createRuntime = useCreateRuntime()
  const updateRuntime = useUpdateRuntime()
  const isPending = createRuntime.isPending || updateRuntime.isPending

  const [form, setForm] = useState({
    id: runtime?.id || '',
    name: runtime?.name || '',
    description: runtime?.description || '',
    image: runtime?.image || 'agentbox/agent:latest',
    cpus: runtime?.cpus || 2,
    memory_mb: runtime?.memory_mb || 4096,
    network: runtime?.network || 'bridge',
    privileged: runtime?.privileged || false,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isEdit && runtime) {
      updateRuntime.mutate(
        {
          id: runtime.id,
          req: {
            name: form.name,
            description: form.description || undefined,
            image: form.image,
            cpus: form.cpus,
            memory_mb: form.memory_mb,
            network: form.network || undefined,
            privileged: form.privileged,
          },
        },
        { onSuccess: () => onOpenChange(false) }
      )
    } else {
      createRuntime.mutate(
        {
          id: form.id,
          name: form.name,
          description: form.description || undefined,
          image: form.image,
          cpus: form.cpus,
          memory_mb: form.memory_mb,
          network: form.network || undefined,
          privileged: form.privileged,
        },
        { onSuccess: () => onOpenChange(false) }
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Runtime' : 'Create Runtime'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the runtime configuration.'
              : 'Define a new runtime environment for agents.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className='space-y-4'>
          <div className='grid grid-cols-2 gap-4'>
            {!isEdit && (
              <div className='space-y-2'>
                <Label htmlFor='rt-id'>ID</Label>
                <Input
                  id='rt-id'
                  value={form.id}
                  onChange={(e) => setForm((f) => ({ ...f, id: e.target.value }))}
                  placeholder='my-runtime'
                  required
                />
              </div>
            )}
            <div className={`space-y-2 ${isEdit ? 'col-span-2' : ''}`}>
              <Label htmlFor='rt-name'>Name</Label>
              <Input
                id='rt-name'
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                placeholder='My Runtime'
                required
              />
            </div>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='rt-image'>Image</Label>
            <Input
              id='rt-image'
              value={form.image}
              onChange={(e) => setForm((f) => ({ ...f, image: e.target.value }))}
              placeholder='agentbox/agent:latest'
              className='font-mono text-sm'
              required
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='rt-desc'>Description</Label>
            <Input
              id='rt-desc'
              value={form.description}
              onChange={(e) =>
                setForm((f) => ({ ...f, description: e.target.value }))
              }
              placeholder='Optional description'
            />
          </div>

          <div className='grid grid-cols-3 gap-4'>
            <div className='space-y-2'>
              <Label htmlFor='rt-cpus'>CPUs</Label>
              <Input
                id='rt-cpus'
                type='number'
                value={form.cpus}
                onChange={(e) =>
                  setForm((f) => ({ ...f, cpus: Number(e.target.value) }))
                }
                min={0.5}
                step={0.5}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='rt-mem'>Memory (MB)</Label>
              <Input
                id='rt-mem'
                type='number'
                value={form.memory_mb}
                onChange={(e) =>
                  setForm((f) => ({ ...f, memory_mb: Number(e.target.value) }))
                }
                min={512}
                step={512}
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='rt-net'>Network</Label>
              <Select
                value={form.network}
                onValueChange={(v) => setForm((f) => ({ ...f, network: v }))}
              >
                <SelectTrigger id='rt-net'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='bridge'>bridge</SelectItem>
                  <SelectItem value='host'>host</SelectItem>
                  <SelectItem value='none'>none</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className='flex items-center justify-between rounded-lg border p-3'>
            <div className='space-y-0.5'>
              <Label htmlFor='rt-privileged'>Privileged Mode</Label>
              <p className='text-xs text-muted-foreground'>
                Enable privileged mode (required for Codex landlock)
              </p>
            </div>
            <Switch
              id='rt-privileged'
              checked={form.privileged}
              onCheckedChange={(checked) =>
                setForm((f) => ({ ...f, privileged: checked }))
              }
            />
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type='submit' disabled={isPending}>
              {isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
              {isEdit ? 'Update' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
