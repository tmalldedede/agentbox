import { Bot } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ThinkingIndicatorProps {
  className?: string
}

export function ThinkingIndicator({ className }: ThinkingIndicatorProps) {
  return (
    <div className={cn('flex items-start gap-3', className)}>
      <div className='flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10'>
        <Bot className='h-4 w-4 text-primary' />
      </div>
      <div className='flex items-center gap-1 rounded-2xl rounded-tl-none bg-muted px-4 py-3'>
        <span className='sr-only'>Agent is thinking</span>
        <span className='animate-bounce text-lg' style={{ animationDelay: '0ms' }}>
          .
        </span>
        <span className='animate-bounce text-lg' style={{ animationDelay: '150ms' }}>
          .
        </span>
        <span className='animate-bounce text-lg' style={{ animationDelay: '300ms' }}>
          .
        </span>
      </div>
    </div>
  )
}
