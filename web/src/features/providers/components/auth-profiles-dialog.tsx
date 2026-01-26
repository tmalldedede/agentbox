import { useState, useCallback } from 'react'
import { Loader2, Plus, Trash2, Clock, Check, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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
  useAuthProfiles,
  useAddAuthProfile,
  useRemoveAuthProfile,
  useRotationStats,
} from '@/hooks/useAuthProfiles'
import type { Provider, AuthProfile } from '@/types'

type AuthProfilesDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  provider: Provider
}

export function AuthProfilesDialog({
  open,
  onOpenChange,
  provider,
}: AuthProfilesDialogProps) {
  const [showAddForm, setShowAddForm] = useState(false)
  const [newApiKey, setNewApiKey] = useState('')
  const [newPriority, setNewPriority] = useState(0)
  const [deleteProfile, setDeleteProfile] = useState<AuthProfile | null>(null)

  const { data: profiles, isLoading } = useAuthProfiles(provider.id)
  const { data: stats } = useRotationStats(provider.id)
  const addProfile = useAddAuthProfile()
  const removeProfile = useRemoveAuthProfile()

  const handleAddProfile = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newApiKey.trim()) return

    addProfile.mutate(
      {
        providerId: provider.id,
        req: { api_key: newApiKey.trim(), priority: newPriority },
      },
      {
        onSuccess: () => {
          setNewApiKey('')
          setNewPriority(0)
          setShowAddForm(false)
        },
      }
    )
  }

  const handleDeleteProfile = () => {
    if (!deleteProfile) return
    removeProfile.mutate(
      { providerId: provider.id, profileId: deleteProfile.id },
      {
        onSuccess: () => setDeleteProfile(null),
      }
    )
  }

  const getProfileStatus = useCallback((profile: AuthProfile) => {
    if (!profile.is_enabled) {
      return { icon: AlertCircle, label: 'Disabled', variant: 'secondary' as const }
    }
    if (profile.cooldown_until) {
      const cooldownEnd = new Date(profile.cooldown_until)
      const now = new Date()
      if (cooldownEnd > now) {
        const remaining = Math.ceil((cooldownEnd.getTime() - now.getTime()) / 1000)
        return { icon: Clock, label: `${remaining}s`, variant: 'warning' as const }
      }
    }
    return { icon: Check, label: 'Active', variant: 'default' as const }
  }, [])

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className='sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <span className='text-lg'>{provider.icon || '☁️'}</span>
              API Keys - {provider.name}
            </DialogTitle>
            <DialogDescription>
              Manage multiple API keys with automatic rotation. Lower priority numbers are used first.
            </DialogDescription>
          </DialogHeader>

          {/* Rotation Stats Summary */}
          {stats && (stats.total_profiles > 0) && (
            <div className='grid grid-cols-4 gap-4 p-4 bg-muted/50 rounded-lg text-sm'>
              <div>
                <span className='text-muted-foreground'>Total Keys</span>
                <p className='text-lg font-semibold'>{stats.total_profiles}</p>
              </div>
              <div>
                <span className='text-muted-foreground'>Active</span>
                <p className='text-lg font-semibold text-green-600'>{stats.active_profiles}</p>
              </div>
              <div>
                <span className='text-muted-foreground'>In Cooldown</span>
                <p className='text-lg font-semibold text-orange-600'>
                  {stats.total_profiles - stats.active_profiles}
                </p>
              </div>
              <div>
                <span className='text-muted-foreground'>Status</span>
                <p className='text-lg font-semibold'>
                  {stats.all_in_cooldown ? (
                    <span className='text-red-600'>All Cooling</span>
                  ) : (
                    <span className='text-green-600'>Healthy</span>
                  )}
                </p>
              </div>
            </div>
          )}

          {/* Profiles Table */}
          <div className='border rounded-lg'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='w-16'>Priority</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead className='w-24'>Status</TableHead>
                  <TableHead className='w-20 text-right'>Success</TableHead>
                  <TableHead className='w-20 text-right'>Fails</TableHead>
                  <TableHead className='w-16'></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell colSpan={6} className='text-center py-8'>
                      <Loader2 className='mx-auto h-6 w-6 animate-spin' />
                    </TableCell>
                  </TableRow>
                ) : !profiles || profiles.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className='text-center py-8 text-muted-foreground'>
                      No API keys configured. Add your first key below.
                    </TableCell>
                  </TableRow>
                ) : (
                  profiles.map((profile) => {
                    const status = getProfileStatus(profile)
                    const StatusIcon = status.icon
                    return (
                      <TableRow key={profile.id}>
                        <TableCell className='font-mono'>{profile.priority}</TableCell>
                        <TableCell className='font-mono text-sm'>
                          {profile.key_masked}
                        </TableCell>
                        <TableCell>
                          <Badge variant={status.variant} className='gap-1'>
                            <StatusIcon className='h-3 w-3' />
                            {status.label}
                          </Badge>
                        </TableCell>
                        <TableCell className='text-right font-mono text-green-600'>
                          {profile.success_count}
                        </TableCell>
                        <TableCell className='text-right font-mono text-red-600'>
                          {profile.fail_count}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant='ghost'
                            size='icon'
                            onClick={() => setDeleteProfile(profile)}
                            className='h-8 w-8 text-muted-foreground hover:text-red-600'
                          >
                            <Trash2 className='h-4 w-4' />
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </div>

          {/* Add Key Form */}
          {showAddForm ? (
            <form onSubmit={handleAddProfile} className='space-y-4 p-4 border rounded-lg'>
              <div className='grid grid-cols-[1fr_100px] gap-4'>
                <div className='space-y-2'>
                  <Label htmlFor='new-api-key'>API Key</Label>
                  <Input
                    id='new-api-key'
                    type='password'
                    value={newApiKey}
                    onChange={(e) => setNewApiKey(e.target.value)}
                    placeholder='sk-...'
                    className='font-mono'
                    autoFocus
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='priority'>Priority</Label>
                  <Input
                    id='priority'
                    type='number'
                    min={0}
                    value={newPriority}
                    onChange={(e) => setNewPriority(parseInt(e.target.value) || 0)}
                    placeholder='0'
                  />
                </div>
              </div>
              <div className='flex justify-end gap-2'>
                <Button
                  type='button'
                  variant='outline'
                  onClick={() => {
                    setShowAddForm(false)
                    setNewApiKey('')
                    setNewPriority(0)
                  }}
                >
                  Cancel
                </Button>
                <Button type='submit' disabled={!newApiKey.trim() || addProfile.isPending}>
                  {addProfile.isPending && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
                  Add Key
                </Button>
              </div>
            </form>
          ) : (
            <Button
              variant='outline'
              className='w-full'
              onClick={() => setShowAddForm(true)}
            >
              <Plus className='mr-2 h-4 w-4' />
              Add API Key
            </Button>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteProfile} onOpenChange={() => setDeleteProfile(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove API Key?</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove the API key{' '}
              <span className='font-mono'>{deleteProfile?.key_masked}</span> from the
              rotation pool.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteProfile}
              className='bg-red-600 hover:bg-red-700'
            >
              {removeProfile.isPending && (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              )}
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
