import { useState } from 'react'
import { Download, Loader2, Check, Terminal, Star, ArrowDownToLine } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { RemoteSkill, SkillCategory } from '@/types'
import { useInstallSkill } from '@/hooks'
import { categoryOptions, categoryBgColors, getRemoteSkillStats } from '../data/data'

type Props = {
  data: RemoteSkill[]
  isLoading: boolean
}

export function StoreGrid({ data, isLoading }: Props) {
  const [searchTerm, setSearchTerm] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [installingId, setInstallingId] = useState<string | null>(null)
  const installSkill = useInstallSkill()

  const handleInstall = (skill: RemoteSkill) => {
    setInstallingId(skill.id)
    installSkill.mutate(
      { sourceId: skill.source_id, skillId: skill.id },
      { onSettled: () => setInstallingId(null) }
    )
  }

  const filtered = data
    .filter((s) =>
      categoryFilter === 'all' ? true : s.category === categoryFilter
    )
    .filter(
      (s) =>
        s.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        s.description?.toLowerCase().includes(searchTerm.toLowerCase())
    )

  if (isLoading) {
    return (
      <div className='flex items-center justify-center h-48'>
        <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  return (
    <div>
      {/* Filter bar */}
      <div className='flex items-center gap-4 mb-4'>
        <Input
          placeholder='Search store...'
          className='h-9 w-40 lg:w-[250px]'
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className='w-[140px] h-9'>
            <SelectValue placeholder='Category' />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value='all'>All Categories</SelectItem>
            {categoryOptions.map((c) => (
              <SelectItem key={c.value} value={c.value}>
                {c.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className='ml-auto text-sm text-muted-foreground'>
          {filtered.length} skills available
        </div>
      </div>

      <Separator className='mb-4' />

      {/* Card grid */}
      {filtered.length === 0 ? (
        <div className='flex flex-col items-center justify-center h-48 text-center'>
          <p className='text-muted-foreground'>No skills found in store.</p>
        </div>
      ) : (
        <ul className='grid gap-4 md:grid-cols-2 lg:grid-cols-3'>
          {filtered.map((skill) => {
            const bgColor = categoryBgColors[skill.category as SkillCategory] || categoryBgColors.other
            const catIcon = categoryOptions.find((c) => c.value === skill.category)
            const mockStats = getRemoteSkillStats(skill.id)

            return (
              <li
                key={`${skill.source_id}-${skill.id}`}
                className='group rounded-xl border bg-card p-5 hover:shadow-lg hover:border-primary/20 transition-all'
              >
                {/* Top: icon + install button */}
                <div className='mb-4 flex items-center justify-between'>
                  <div className={`flex h-11 w-11 items-center justify-center rounded-xl ${bgColor} transition-transform group-hover:scale-110`}>
                    {catIcon?.icon && <catIcon.icon className='h-5 w-5' />}
                  </div>
                  {skill.is_installed ? (
                    <Button variant='outline' size='sm' disabled className='gap-1.5'>
                      <Check className='h-3.5 w-3.5 text-emerald-500' />
                      Installed
                    </Button>
                  ) : (
                    <Button
                      size='sm'
                      onClick={() => handleInstall(skill)}
                      disabled={installingId === skill.id}
                      className='gap-1.5'
                    >
                      {installingId === skill.id ? (
                        <Loader2 className='h-3.5 w-3.5 animate-spin' />
                      ) : (
                        <Download className='h-3.5 w-3.5' />
                      )}
                      Install
                    </Button>
                  )}
                </div>

                {/* Name + command */}
                <div className='mb-2'>
                  <h3 className='font-semibold text-base'>{skill.name}</h3>
                  <div className='flex items-center gap-1.5 mt-0.5'>
                    <Terminal className='h-3 w-3 text-muted-foreground' />
                    <code className='text-xs text-emerald-600 dark:text-emerald-400 font-mono'>
                      /{skill.command}
                    </code>
                  </div>
                </div>

                {/* Description */}
                <p className='text-sm text-muted-foreground line-clamp-2 mb-4'>
                  {skill.description || 'No description'}
                </p>

                {/* Stats Row */}
                <div className='flex items-center gap-4 mb-3 text-xs text-muted-foreground'>
                  <span className='flex items-center gap-1'>
                    <Star className='h-3 w-3 text-amber-500 fill-amber-500' />
                    <span className='font-medium text-foreground'>{mockStats.stars}</span>
                  </span>
                  <span className='flex items-center gap-1'>
                    <ArrowDownToLine className='h-3 w-3' />
                    <span className='font-medium text-foreground'>{mockStats.downloads.toLocaleString()}</span>
                  </span>
                  {skill.version && (
                    <span className='ml-auto font-mono'>v{skill.version}</span>
                  )}
                </div>

                {/* Meta badges */}
                <div className='flex items-center gap-1.5 flex-wrap'>
                  <Badge variant='outline' className='text-[10px] capitalize'>
                    {skill.category}
                  </Badge>
                  {skill.author && (
                    <Badge variant='secondary' className='text-[10px]'>
                      {skill.author}
                    </Badge>
                  )}
                  <Badge variant='secondary' className='text-[10px]'>
                    {skill.source_name}
                  </Badge>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
