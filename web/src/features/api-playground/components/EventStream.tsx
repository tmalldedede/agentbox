import { useEffect, useRef } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import {
  Loader2,
  Clock,
  MessageSquare,
  Wrench,
  CheckCircle2,
  XCircle,
  RefreshCw,
} from 'lucide-react'

type TaskEvent = {
  type: string
  data?: any
  timestamp?: string
}

type EventStreamProps = {
  events: TaskEvent[]
  isRunning: boolean
}

export function EventStream({ events, isRunning }: EventStreamProps) {
  const scrollRef = useRef<HTMLDivElement>(null)

  // 自动滚动到底部
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [events])

  if (events.length === 0) {
    return (
      <Card className='flex h-[500px] items-center justify-center border-dashed'>
        <div className='text-center text-muted-foreground'>
          <MessageSquare className='mx-auto mb-2 h-8 w-8 opacity-50' />
          <p className='text-sm'>Send a prompt to see live events</p>
        </div>
      </Card>
    )
  }

  return (
    <Card className='h-[500px] overflow-hidden'>
      <ScrollArea className='h-full p-4' ref={scrollRef}>
        <div className='space-y-3'>
          {events.map((event, index) => (
            <EventItem key={index} event={event} />
          ))}
          {isRunning && (
            <div className='flex items-center gap-2 text-sm text-muted-foreground'>
              <Loader2 className='h-4 w-4 animate-spin' />
              <span>Waiting for events...</span>
            </div>
          )}
        </div>
      </ScrollArea>
    </Card>
  )
}

function EventItem({ event }: { event: TaskEvent }) {
  const time = event.timestamp
    ? new Date(event.timestamp).toLocaleTimeString()
    : new Date().toLocaleTimeString()

  switch (event.type) {
    case 'task.started':
      return (
        <div className='flex items-start gap-3'>
          <Clock className='mt-0.5 h-4 w-4 text-blue-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-blue-500 text-blue-500'>
                Started
              </Badge>
            </div>
            <p className='mt-1 text-sm'>Task started</p>
          </div>
        </div>
      )

    case 'task.turn_started':
      return (
        <div className='flex items-start gap-3'>
          <RefreshCw className='mt-0.5 h-4 w-4 text-purple-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-purple-500 text-purple-500'>
                Turn #{event.data?.turn_number || 1}
              </Badge>
            </div>
          </div>
        </div>
      )

    case 'agent.thinking':
      return (
        <div className='flex items-start gap-3'>
          <Loader2 className='mt-0.5 h-4 w-4 animate-spin text-yellow-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-yellow-500 text-yellow-500'>
                Thinking
              </Badge>
            </div>
            <p className='mt-1 text-sm text-muted-foreground'>Agent is thinking...</p>
          </div>
        </div>
      )

    case 'agent.tool_call':
      return <ToolCallEvent time={time} data={event.data} />

    case 'agent.message':
      return (
        <div className='flex items-start gap-3'>
          <MessageSquare className='mt-0.5 h-4 w-4 text-green-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-green-500 text-green-500'>
                Message
              </Badge>
            </div>
            <div className='mt-2 rounded-lg bg-muted p-3 text-sm'>
              {event.data?.text || event.data}
            </div>
          </div>
        </div>
      )

    case 'task.completed':
      return (
        <div className='flex items-start gap-3'>
          <CheckCircle2 className='mt-0.5 h-4 w-4 text-green-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-green-500 text-green-500'>
                Completed
              </Badge>
            </div>
            {event.data?.usage && (
              <div className='mt-2 flex gap-4 text-xs text-muted-foreground'>
                <span>📊 Tokens: {event.data.usage.input_tokens} in</span>
                <span>{event.data.usage.output_tokens} out</span>
                {event.data.usage.cached_input_tokens > 0 && (
                  <span>{event.data.usage.cached_input_tokens} cached</span>
                )}
              </div>
            )}
          </div>
        </div>
      )

    case 'task.failed':
      return (
        <div className='flex items-start gap-3'>
          <XCircle className='mt-0.5 h-4 w-4 text-red-500' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline' className='border-red-500 text-red-500'>
                Failed
              </Badge>
            </div>
            <div className='mt-2 rounded-lg bg-red-50 p-3 text-sm text-red-700 dark:bg-red-950/20'>
              {event.data?.error || 'Task failed'}
            </div>
          </div>
        </div>
      )

    default:
      return (
        <div className='flex items-start gap-3'>
          <div className='mt-0.5 h-4 w-4' />
          <div className='flex-1'>
            <div className='flex items-center gap-2'>
              <span className='text-xs text-muted-foreground'>{time}</span>
              <Badge variant='outline'>{event.type}</Badge>
            </div>
            {event.data && (
              <pre className='mt-2 text-xs text-muted-foreground'>
                {JSON.stringify(event.data, null, 2)}
              </pre>
            )}
          </div>
        </div>
      )
  }
}

function ToolCallEvent({ time, data }: { time: string; data: any }) {
  return (
    <div className='flex items-start gap-3'>
      <Wrench className='mt-0.5 h-4 w-4 text-orange-500' />
      <div className='flex-1'>
        <div className='flex items-center gap-2'>
          <span className='text-xs text-muted-foreground'>{time}</span>
          <Badge variant='outline' className='border-orange-500 text-orange-500'>
            Tool
          </Badge>
          <code className='rounded bg-muted px-2 py-0.5 text-xs font-mono'>
            {data?.tool_name || data?.name || 'unknown'}
          </code>
        </div>

        {/* Tool Input */}
        {data?.input && (
          <details className='mt-2 group'>
            <summary className='cursor-pointer text-xs font-medium text-muted-foreground hover:text-foreground'>
              Input {data.input && typeof data.input === 'string' && `(${data.input.length} chars)`}
            </summary>
            <div className='mt-1 rounded-lg bg-muted p-3'>
              <pre className='overflow-x-auto text-xs'>
                {typeof data.input === 'string'
                  ? data.input
                  : JSON.stringify(data.input, null, 2)}
              </pre>
            </div>
          </details>
        )}

        {/* Tool Output */}
        {data?.output && (
          <details className='mt-2 group' open={typeof data.output === 'string' && data.output.length < 200}>
            <summary className='cursor-pointer text-xs font-medium text-muted-foreground hover:text-foreground'>
              Output {data.output && typeof data.output === 'string' && `(${data.output.length} chars)`}
            </summary>
            <div className='mt-1 rounded-lg bg-muted p-3'>
              <pre className='overflow-x-auto text-xs'>
                {typeof data.output === 'string'
                  ? data.output
                  : JSON.stringify(data.output, null, 2)}
              </pre>
            </div>
          </details>
        )}
      </div>
    </div>
  )
}
