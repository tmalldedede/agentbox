import { useNavigate } from '@tanstack/react-router'
import { Loader2, Plus, Zap, Store, Github } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Button } from '@/components/ui/button'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import {
  useSkills,
  useRemoteSkills,
  useSkillSources,
} from '@/hooks'
import { InstalledTable } from './installed-table'
import { StoreGrid } from './store-grid'
import { SourcesList } from './sources-list'

export default function SkillList() {
  const navigate = useNavigate()
  const { data: skills = [], isLoading } = useSkills()
  const { data: remoteSkills = [], isLoading: loadingRemote } = useRemoteSkills()
  const { data: sources = [], isLoading: loadingSources } = useSkillSources()

  return (
    <>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Skills</h2>
            <p className='text-muted-foreground'>
              Reusable task templates that define how agents handle specific tasks.
              Invoke via commands like{' '}
              <code className='text-emerald-600 dark:text-emerald-400'>/commit</code>
              {' '}or{' '}
              <code className='text-emerald-600 dark:text-emerald-400'>/review-pr</code>.
            </p>
          </div>
          <Button size='sm' onClick={() => navigate({ to: '/skills/new' })}>
            <Plus className='mr-2 h-4 w-4' />
            New Skill
          </Button>
        </div>

        <Tabs defaultValue='installed' className='flex-1'>
          <TabsList>
            <TabsTrigger value='installed' className='gap-1.5'>
              <Zap className='h-4 w-4' />
              Installed ({skills.length})
            </TabsTrigger>
            <TabsTrigger value='store' className='gap-1.5'>
              <Store className='h-4 w-4' />
              Store ({remoteSkills.filter((s) => !s.is_installed).length})
            </TabsTrigger>
            <TabsTrigger value='sources' className='gap-1.5'>
              <Github className='h-4 w-4' />
              Sources ({sources.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value='installed' className='mt-4'>
            {isLoading ? (
              <div className='flex items-center justify-center h-48'>
                <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
              </div>
            ) : (
              <InstalledTable data={skills} />
            )}
          </TabsContent>

          <TabsContent value='store' className='mt-4'>
            <StoreGrid data={remoteSkills} isLoading={loadingRemote} />
          </TabsContent>

          <TabsContent value='sources' className='mt-4'>
            <SourcesList data={sources} isLoading={loadingSources} />
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}
