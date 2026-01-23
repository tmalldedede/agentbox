import { useState } from 'react'
import { Loader2, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import type { AddSourceRequest } from '@/types'
import { useAddSkillSource } from '@/hooks'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AddSourceDialog({ open, onOpenChange }: Props) {
  const addSource = useAddSkillSource()
  const [formData, setFormData] = useState<AddSourceRequest>({
    id: '',
    name: '',
    owner: '',
    repo: '',
    branch: 'main',
    path: 'skills',
    type: 'community',
    description: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    addSource.mutate(formData, {
      onSuccess: () => {
        onOpenChange(false)
        setFormData({
          id: '', name: '', owner: '', repo: '',
          branch: 'main', path: 'skills', type: 'community', description: '',
        })
      },
    })
  }

  const handleGitHubUrlParse = (url: string) => {
    const match = url.match(/github\.com\/([^/]+)\/([^/]+)/)
    if (match) {
      const owner = match[1]
      const repo = match[2].replace(/\.git$/, '')
      setFormData((prev) => ({
        ...prev,
        owner,
        repo,
        id: `${owner}-${repo}`,
        name: repo,
      }))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Add Skill Source</DialogTitle>
          <DialogDescription>
            Add a GitHub repository as a skill source.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className='space-y-4'>
          <div className='space-y-2'>
            <Label>GitHub URL (Quick Fill)</Label>
            <Input
              placeholder='https://github.com/owner/repo'
              onChange={(e) => handleGitHubUrlParse(e.target.value)}
            />
            <p className='text-xs text-muted-foreground'>
              Paste GitHub URL to auto-fill owner and repo
            </p>
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label>Owner <span className='text-red-500'>*</span></Label>
              <Input
                value={formData.owner}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, owner: e.target.value }))
                }
                placeholder='owner'
                required
              />
            </div>
            <div className='space-y-2'>
              <Label>Repo <span className='text-red-500'>*</span></Label>
              <Input
                value={formData.repo}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, repo: e.target.value }))
                }
                placeholder='repo'
                required
              />
            </div>
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label>Branch</Label>
              <Input
                value={formData.branch}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, branch: e.target.value }))
                }
                placeholder='main'
              />
            </div>
            <div className='space-y-2'>
              <Label>Skills Path</Label>
              <Input
                value={formData.path}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, path: e.target.value }))
                }
                placeholder='skills'
              />
            </div>
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label>ID <span className='text-red-500'>*</span></Label>
              <Input
                value={formData.id}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, id: e.target.value }))
                }
                placeholder='my-skills'
                required
              />
            </div>
            <div className='space-y-2'>
              <Label>Name <span className='text-red-500'>*</span></Label>
              <Input
                value={formData.name}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, name: e.target.value }))
                }
                placeholder='My Skills'
                required
              />
            </div>
          </div>

          <div className='space-y-2'>
            <Label>Description</Label>
            <Input
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
              placeholder='Optional description'
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
            <Button
              type='submit'
              disabled={
                addSource.isPending ||
                !formData.id ||
                !formData.owner ||
                !formData.repo
              }
            >
              {addSource.isPending ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <Plus className='mr-2 h-4 w-4' />
              )}
              Add Source
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
