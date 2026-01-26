import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Play, Trash2, Plus, Power, PowerOff } from 'lucide-react'
import { toast } from 'sonner'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api, CronJob, CreateCronJobRequest } from '@/services/api'

export const Route = createFileRoute('/_authenticated/crons/')({
  component: CronsPage,
})

function CronsPage() {
  const queryClient = useQueryClient()
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [formData, setFormData] = useState<CreateCronJobRequest>({
    name: '',
    schedule: '',
    agent_id: '',
    prompt: '',
    enabled: true,
  })

  const { data: cronJobs = [], isLoading } = useQuery({
    queryKey: ['cronJobs'],
    queryFn: api.listCronJobs,
  })

  const { data: agents = [] } = useQuery({
    queryKey: ['agents'],
    queryFn: api.listAgents,
  })

  const createMutation = useMutation({
    mutationFn: api.createCronJob,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cronJobs'] })
      setShowCreateDialog(false)
      setFormData({ name: '', schedule: '', agent_id: '', prompt: '', enabled: true })
      toast.success('Cron job created')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: api.deleteCronJob,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cronJobs'] })
      toast.success('Cron job deleted')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const triggerMutation = useMutation({
    mutationFn: api.triggerCronJob,
    onSuccess: () => {
      toast.success('Cron job triggered')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.updateCronJob(id, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cronJobs'] })
      toast.success('Cron job updated')
    },
    onError: (error: Error) => {
      toast.error(error.message)
    },
  })

  const handleCreate = () => {
    const errors: string[] = []
    if (!formData.name) errors.push('Name')
    if (!formData.schedule) errors.push('Schedule')
    if (!formData.agent_id) errors.push('Agent')
    if (!formData.prompt) errors.push('Prompt')

    if (errors.length > 0) {
      toast.error(`Missing required fields: ${errors.join(', ')}`)
      return
    }
    createMutation.mutate(formData)
  }

  const formatDate = (date?: string) => {
    if (!date) return '-'
    return new Date(date).toLocaleString()
  }

  return (
    <Main>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Cron Jobs</h1>
          <p className="text-muted-foreground">Schedule automated tasks</p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Cron Job
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Schedule</TableHead>
              <TableHead>Agent</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Last Run</TableHead>
              <TableHead>Next Run</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                  Loading...
                </TableCell>
              </TableRow>
            ) : cronJobs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                  No cron jobs configured
                </TableCell>
              </TableRow>
            ) : (
              cronJobs.map((job: CronJob) => (
                <TableRow key={job.id}>
                  <TableCell className="font-medium">{job.name}</TableCell>
                  <TableCell>
                    <code className="text-xs bg-muted px-1 py-0.5 rounded">
                      {job.schedule}
                    </code>
                  </TableCell>
                  <TableCell>{job.agent_id.slice(0, 8)}...</TableCell>
                  <TableCell>
                    <Badge variant={job.enabled ? 'default' : 'secondary'}>
                      {job.enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                    {job.last_status && (
                      <Badge
                        variant={job.last_status === 'success' ? 'default' : 'destructive'}
                        className="ml-1"
                      >
                        {job.last_status}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(job.last_run)}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(job.next_run)}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => toggleMutation.mutate({ id: job.id, enabled: !job.enabled })}
                      title={job.enabled ? 'Disable' : 'Enable'}
                    >
                      {job.enabled ? (
                        <PowerOff className="h-4 w-4" />
                      ) : (
                        <Power className="h-4 w-4" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => triggerMutation.mutate(job.id)}
                      title="Trigger Now"
                    >
                      <Play className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => deleteMutation.mutate(job.id)}
                      title="Delete"
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Create Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Create Cron Job</DialogTitle>
            <DialogDescription>
              Schedule a task to run automatically at specified intervals.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Daily Security Scan"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="schedule">Schedule (Cron Expression)</Label>
              <Input
                id="schedule"
                value={formData.schedule}
                onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                placeholder="0 9 * * *"
              />
              <p className="text-xs text-muted-foreground">
                Examples: "0 9 * * *" (daily at 9am), "*/30 * * * *" (every 30 min)
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="agent">Agent</Label>
              <Select
                value={formData.agent_id}
                onValueChange={(value) => setFormData({ ...formData, agent_id: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select an agent" />
                </SelectTrigger>
                <SelectContent>
                  {agents.map((agent) => (
                    <SelectItem key={agent.id} value={agent.id}>
                      {agent.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="prompt">Prompt</Label>
              <Textarea
                id="prompt"
                value={formData.prompt}
                onChange={(e) => setFormData({ ...formData, prompt: e.target.value })}
                placeholder="Run security scan and generate report..."
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Main>
  )
}
