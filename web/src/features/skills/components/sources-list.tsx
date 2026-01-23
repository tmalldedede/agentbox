import { useState } from 'react'
import {
  Github,
  ExternalLink,
  RefreshCw,
  Trash2,
  Loader2,
  Plus,
  Star,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { SkillSource } from '@/types'
import {
  useRefreshSkillSource,
  useRemoveSkillSource,
} from '@/hooks'
import { AddSourceDialog } from './add-source-dialog'

type Props = {
  data: SkillSource[]
  isLoading: boolean
}

export function SourcesList({ data, isLoading }: Props) {
  const [refreshingId, setRefreshingId] = useState<string | null>(null)
  const [showAddDialog, setShowAddDialog] = useState(false)
  const refreshSource = useRefreshSkillSource()
  const removeSource = useRemoveSkillSource()

  const handleRefresh = (id: string) => {
    setRefreshingId(id)
    refreshSource.mutate(id, {
      onSettled: () => setRefreshingId(null),
    })
  }

  const handleDelete = (source: SkillSource) => {
    if (!confirm(`Remove source "${source.name}"?`)) return
    removeSource.mutate(source.id)
  }

  if (isLoading) {
    return (
      <div className='flex items-center justify-center h-48'>
        <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <div className='flex justify-end'>
        <Button size='sm' onClick={() => setShowAddDialog(true)}>
          <Plus className='mr-2 h-4 w-4' />
          Add Source
        </Button>
      </div>

      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Source</TableHead>
              <TableHead>Repository</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Stars</TableHead>
              <TableHead className='w-[100px]'>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className='h-24 text-center'>
                  No sources configured. Add a GitHub repository as a skill source.
                </TableCell>
              </TableRow>
            ) : (
              data.map((source) => (
                <TableRow key={source.id}>
                  <TableCell>
                    <div className='flex items-center gap-2'>
                      <Github className='h-4 w-4 text-muted-foreground' />
                      <div>
                        <div className='font-medium'>{source.name}</div>
                        {source.description && (
                          <div className='text-xs text-muted-foreground line-clamp-1'>
                            {source.description}
                          </div>
                        )}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <code className='text-xs font-mono'>
                      {source.owner}/{source.repo}:{source.branch}/{source.path}
                    </code>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={source.type === 'official' ? 'default' : 'secondary'}
                      className='text-xs'
                    >
                      {source.type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {source.stars != null && source.stars > 0 ? (
                      <span className='flex items-center gap-1 text-xs text-amber-500'>
                        <Star className='h-3 w-3' />
                        {source.stars}
                      </span>
                    ) : (
                      <span className='text-xs text-muted-foreground'>-</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className='flex items-center gap-1'>
                      <Button
                        variant='ghost'
                        size='icon'
                        className='h-7 w-7'
                        asChild
                      >
                        <a
                          href={`https://github.com/${source.owner}/${source.repo}`}
                          target='_blank'
                          rel='noopener noreferrer'
                        >
                          <ExternalLink className='h-3.5 w-3.5' />
                        </a>
                      </Button>
                      <Button
                        variant='ghost'
                        size='icon'
                        className='h-7 w-7'
                        onClick={() => handleRefresh(source.id)}
                        disabled={refreshingId === source.id}
                      >
                        <RefreshCw
                          className={`h-3.5 w-3.5 ${refreshingId === source.id ? 'animate-spin' : ''}`}
                        />
                      </Button>
                      {source.type !== 'official' && (
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-7 w-7 text-red-500 hover:text-red-600'
                          onClick={() => handleDelete(source)}
                          disabled={removeSource.isPending}
                        >
                          <Trash2 className='h-3.5 w-3.5' />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <AddSourceDialog
        open={showAddDialog}
        onOpenChange={setShowAddDialog}
      />
    </div>
  )
}
