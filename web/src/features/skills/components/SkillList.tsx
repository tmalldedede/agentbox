import { useNavigate } from '@tanstack/react-router'
import {
  Loader2,
  Plus,
  Zap,
  Store,
  Github,
  CheckCircle,
  Clock,
  BarChart3,
} from 'lucide-react'
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
import { getAggregateStats } from '../data/data'

function StatCard({
  icon: Icon,
  label,
  value,
  sub,
  color,
}: {
  icon: React.ElementType
  label: string
  value: string | number
  sub?: string
  color: string
}) {
  return (
    <div className='flex items-center gap-3 rounded-xl border bg-card p-4'>
      <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${color}`}>
        <Icon className='h-5 w-5' />
      </div>
      <div className='min-w-0'>
        <p className='text-2xl font-bold tracking-tight'>{value}</p>
        <p className='text-xs text-muted-foreground truncate'>
          {label}
          {sub && <span className='ml-1 text-emerald-600 dark:text-emerald-400'>{sub}</span>}
        </p>
      </div>
    </div>
  )
}

export default function SkillList() {
  const navigate = useNavigate()
  const { data: skills = [], isLoading } = useSkills()
  const { data: remoteSkills = [], isLoading: loadingRemote } = useRemoteSkills()
  const { data: sources = [], isLoading: loadingSources } = useSkillSources()

  const stats = getAggregateStats(skills)

  return (
    <>
      <Header fixed className='md:hidden' />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>Skills</h2>
            <p className='text-muted-foreground'>
              Reusable task templates that define how agents handle specific tasks.
            </p>
          </div>
          <Button size='sm' onClick={() => navigate({ to: '/skills/new' })}>
            <Plus className='mr-2 h-4 w-4' />
            New Skill
          </Button>
        </div>

        {/* Stats KPI Cards */}
        {!isLoading && skills.length > 0 && (
          <div className='grid grid-cols-2 gap-3 lg:grid-cols-4'>
            <StatCard
              icon={Zap}
              label='Total Skills'
              value={stats.totalSkills}
              sub={`${stats.enabledSkills} active`}
              color='bg-blue-500/10 text-blue-500'
            />
            <StatCard
              icon={BarChart3}
              label='Total Invocations'
              value={stats.totalUsage.toLocaleString()}
              sub='+12% this week'
              color='bg-purple-500/10 text-purple-500'
            />
            <StatCard
              icon={CheckCircle}
              label='Avg Success Rate'
              value={`${stats.avgSuccess}%`}
              color='bg-emerald-500/10 text-emerald-500'
            />
            <StatCard
              icon={Clock}
              label='Avg Duration'
              value={`${stats.avgDuration}s`}
              color='bg-amber-500/10 text-amber-500'
            />
          </div>
        )}

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
