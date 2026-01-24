import { useSearch } from '@tanstack/react-router'
import { Bot, Shield, Zap, Cpu } from 'lucide-react'
import { cn } from '@/lib/utils'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })

  return (
    <div className='relative container grid h-svh flex-col items-center justify-center lg:max-w-none lg:grid-cols-2 lg:px-0'>
      {/* Left side - Login form */}
      <div className='flex h-full items-center justify-center lg:p-8'>
        <div className='mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[380px]'>
          {/* Logo and brand */}
          <div className='flex flex-col items-center space-y-2'>
            <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-primary'>
              <Bot className='h-7 w-7 text-primary-foreground' />
            </div>
            <h1 className='text-2xl font-bold tracking-tight'>AgentBox</h1>
            <p className='text-sm text-muted-foreground'>
              AI Agent Orchestration Platform
            </p>
          </div>

          {/* Login card */}
          <div className='rounded-lg border bg-card p-6 shadow-sm'>
            <div className='flex flex-col space-y-1.5 pb-4'>
              <h2 className='text-lg font-semibold'>Welcome back</h2>
              <p className='text-sm text-muted-foreground'>
                Sign in to your account to continue
              </p>
            </div>
            <UserAuthForm redirectTo={redirect} />
          </div>

          {/* Footer */}
          <p className='text-center text-xs text-muted-foreground'>
            Secure authentication powered by JWT
          </p>
        </div>
      </div>

      {/* Right side - Feature showcase */}
      <div
        className={cn(
          'relative hidden h-full flex-col bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 p-10 text-white lg:flex',
          'dark:from-slate-950 dark:via-purple-950 dark:to-slate-950'
        )}
      >
        {/* Grid pattern overlay */}
        <div
          className='absolute inset-0 opacity-20'
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.15'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
          }}
        />

        {/* Content */}
        <div className='relative z-10 flex h-full flex-col justify-between'>
          {/* Top - Logo */}
          <div className='flex items-center gap-2'>
            <div className='flex h-8 w-8 items-center justify-center rounded-lg bg-white/10'>
              <Bot className='h-5 w-5' />
            </div>
            <span className='font-semibold'>AgentBox</span>
          </div>

          {/* Center - Features */}
          <div className='space-y-8'>
            <div>
              <h2 className='text-3xl font-bold leading-tight'>
                Orchestrate AI Agents
                <br />
                <span className='text-purple-300'>at Scale</span>
              </h2>
              <p className='mt-4 text-lg text-white/70'>
                Build, deploy, and manage intelligent AI agents with enterprise-grade security and reliability.
              </p>
            </div>

            <div className='grid gap-4'>
              <FeatureItem
                icon={Cpu}
                title='Multi-Engine Support'
                description='Claude Code, Codex, and custom adapters'
              />
              <FeatureItem
                icon={Zap}
                title='Batch Processing'
                description='Run thousands of tasks in parallel'
              />
              <FeatureItem
                icon={Shield}
                title='Enterprise Security'
                description='Role-based access control & audit logs'
              />
            </div>
          </div>

          {/* Bottom - Stats */}
          <div className='flex gap-8 border-t border-white/10 pt-6'>
            <Stat value='10K+' label='Tasks/Day' />
            <Stat value='99.9%' label='Uptime' />
            <Stat value='<100ms' label='Latency' />
          </div>
        </div>
      </div>
    </div>
  )
}

function FeatureItem({
  icon: Icon,
  title,
  description,
}: {
  icon: React.ElementType
  title: string
  description: string
}) {
  return (
    <div className='flex items-start gap-3'>
      <div className='flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white/10'>
        <Icon className='h-5 w-5' />
      </div>
      <div>
        <h3 className='font-medium'>{title}</h3>
        <p className='text-sm text-white/60'>{description}</p>
      </div>
    </div>
  )
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div>
      <div className='text-2xl font-bold'>{value}</div>
      <div className='text-sm text-white/60'>{label}</div>
    </div>
  )
}
