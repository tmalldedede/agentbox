import { useState, useRef, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { useTask, useAppendTurn, useTaskEvents, useCancelTask, useRetryTask, useDeleteTask } from '@/hooks/useTasks'
import { useDockerAvailable } from '@/hooks/useSystemHealth'
import {
  ArrowLeft,
  Loader2,
  Send,
  Square,
  FileText,
  Clock,
  Bot,
  User,
  Zap,
  RotateCw,
  Trash2,
  Download,
} from 'lucide-react'
import type { TaskEvent } from '@/types'

const statusColors: Record<string, string> = {
  pending: 'bg-yellow-500/10 text-yellow-600 border-yellow-200',
  queued: 'bg-blue-500/10 text-blue-600 border-blue-200',
  running: 'bg-green-500/10 text-green-600 border-green-200',
  completed: 'bg-gray-500/10 text-gray-600 border-gray-200',
  failed: 'bg-red-500/10 text-red-600 border-red-200',
  cancelled: 'bg-orange-500/10 text-orange-600 border-orange-200',
}

export function TaskDetail({ taskId }: { taskId: string }) {
  const navigate = useNavigate()
  const dockerAvailable = useDockerAvailable()
  const { data: task, isLoading } = useTask(taskId)
  const appendTurn = useAppendTurn()
  const cancelTask = useCancelTask()
  const retryTask = useRetryTask()
  const deleteTask = useDeleteTask()
  const isRunning = task?.status === 'running' || task?.status === 'queued'
  const isTerminal = task?.status === 'completed' || task?.status === 'failed' || task?.status === 'cancelled'
  const canRetry = task?.status === 'failed' || task?.status === 'cancelled'
  const { events } = useTaskEvents(taskId, isRunning)
  const [prompt, setPrompt] = useState('')
  const eventsEndRef = useRef<HTMLDivElement>(null)

  // 自动滚动到最新事件
  useEffect(() => {
    eventsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [events.length])

  const handleSubmit = () => {
    if (!prompt.trim() || appendTurn.isPending) return
    appendTurn.mutate({ taskId, prompt: prompt.trim() })
    setPrompt('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  if (isLoading) {
    return (
      <>
        <Header fixed>
          <div className='ms-auto flex items-center space-x-4'>
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
        </Header>
        <Main className='flex items-center justify-center'>
          <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
        </Main>
      </>
    )
  }

  if (!task) {
    return (
      <>
        <Header fixed>
          <div className='ms-auto flex items-center space-x-4'>
            <ThemeSwitch />
            <ProfileDropdown />
          </div>
        </Header>
        <Main className='flex items-center justify-center'>
          <p className='text-muted-foreground'>Task not found</p>
        </Main>
      </>
    )
  }

  return (
    <>
      <Header fixed>
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* 顶部导航 + Task 信息 */}
        <div className='flex items-center gap-4'>
          <Button
            variant='ghost'
            size='icon'
            onClick={() => navigate({ to: '/tasks' })}
          >
            <ArrowLeft className='h-4 w-4' />
          </Button>
          <div className='flex-1'>
            <div className='flex items-center gap-3'>
              <h2 className='text-xl font-bold tracking-tight'>
                {task.id}
              </h2>
              <Badge
                variant='outline'
                className={`text-xs capitalize ${statusColors[task.status] || ''}`}
              >
                {task.status}
              </Badge>
            </div>
            <div className='flex items-center gap-4 mt-1 text-sm text-muted-foreground'>
              <span className='flex items-center gap-1'>
                <Bot className='h-3 w-3' />
                {task.agent_name || task.agent_id}
              </span>
              <span className='flex items-center gap-1'>
                <Clock className='h-3 w-3' />
                {new Date(task.created_at).toLocaleString()}
              </span>
              <span className='flex items-center gap-1'>
                <Zap className='h-3 w-3' />
                {task.turn_count} turn(s)
              </span>
            </div>
          </div>
          <div className='flex items-center gap-2'>
            {isRunning && (
              <Button
                variant='destructive'
                size='sm'
                onClick={() => cancelTask.mutate(taskId)}
                disabled={cancelTask.isPending}
              >
                <Square className='h-3 w-3 mr-1' />
                Cancel
              </Button>
            )}
            {canRetry && (
              <Button
                variant='outline'
                size='sm'
                onClick={() => retryTask.mutate(taskId, {
                  onSuccess: (newTask) => navigate({ to: `/tasks/${newTask.id}` }),
                })}
                disabled={retryTask.isPending || !dockerAvailable}
              >
                <RotateCw className='h-3 w-3 mr-1' />
                Retry
              </Button>
            )}
            {isTerminal && (
              <Button
                variant='ghost'
                size='sm'
                className='text-destructive hover:text-destructive'
                onClick={() => deleteTask.mutate(taskId, {
                  onSuccess: () => navigate({ to: '/tasks' }),
                })}
                disabled={deleteTask.isPending}
              >
                <Trash2 className='h-3 w-3 mr-1' />
                Delete
              </Button>
            )}
          </div>
        </div>

        {/* Turns 列表 */}
        <div className='flex-1 space-y-4 overflow-auto'>
          {task.turns?.map((turn) => (
            <Card key={turn.id} className='shadow-sm'>
              <CardHeader className='pb-2'>
                <CardTitle className='text-sm flex items-center gap-2'>
                  <User className='h-4 w-4 text-blue-500' />
                  <span className='font-mono text-xs text-muted-foreground'>
                    {turn.id}
                  </span>
                  <span className='text-xs text-muted-foreground'>
                    {new Date(turn.created_at).toLocaleString()}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className='space-y-3'>
                {/* Prompt */}
                <div className='rounded-md bg-muted/50 p-3'>
                  <p className='text-sm whitespace-pre-wrap'>{turn.prompt}</p>
                </div>
                {/* Result */}
                {turn.result ? (
                  <div className='rounded-md border p-3'>
                    <div className='flex items-center gap-2 mb-2'>
                      <Bot className='h-4 w-4 text-green-500' />
                      <span className='text-xs font-medium text-muted-foreground'>
                        Response
                      </span>
                    </div>
                    <p className='text-sm whitespace-pre-wrap'>
                      {turn.result.text || turn.result.summary || 'No output'}
                    </p>
                    {turn.result.usage && (
                      <p className='text-xs text-muted-foreground mt-2'>
                        Duration: {turn.result.usage.duration_seconds}s
                        {turn.result.usage.total_tokens &&
                          ` | Tokens: ${turn.result.usage.total_tokens}`}
                      </p>
                    )}
                  </div>
                ) : isRunning ? (
                  <div className='flex items-center gap-2 text-sm text-muted-foreground'>
                    <Loader2 className='h-4 w-4 animate-spin' />
                    Executing...
                  </div>
                ) : null}
              </CardContent>
            </Card>
          ))}

          {/* SSE 实时事件 */}
          {events.length > 0 && (
            <Card className='shadow-sm border-dashed'>
              <CardHeader className='pb-2'>
                <CardTitle className='text-sm flex items-center gap-2'>
                  <Zap className='h-4 w-4 text-orange-500' />
                  Live Events
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className='space-y-1 max-h-[200px] overflow-auto text-xs font-mono'>
                  {events.map((event: TaskEvent, i: number) => (
                    <div key={i} className='text-muted-foreground'>
                      <span className='text-orange-500'>[{event.type}]</span>{' '}
                      {typeof event.data === 'object'
                        ? JSON.stringify(event.data)
                        : String(event.data || '')}
                    </div>
                  ))}
                  <div ref={eventsEndRef} />
                </div>
              </CardContent>
            </Card>
          )}

          {/* Error 信息 */}
          {task.error_message && (
            <Card className='shadow-sm border-red-200'>
              <CardContent className='pt-4'>
                <p className='text-sm text-red-600'>{task.error_message}</p>
              </CardContent>
            </Card>
          )}

          {/* 附件和输出文件 */}
          {((task.attachments && task.attachments.length > 0) ||
            (task.output_files && task.output_files.length > 0)) && (
            <Card className='shadow-sm'>
              <CardHeader className='pb-2'>
                <CardTitle className='text-sm flex items-center gap-2'>
                  <FileText className='h-4 w-4' />
                  Files
                </CardTitle>
              </CardHeader>
              <CardContent className='space-y-2'>
                {task.attachments && task.attachments.length > 0 && (
                  <div>
                    <p className='text-xs font-medium text-muted-foreground mb-1'>
                      Attachments
                    </p>
                    <div className='flex flex-wrap gap-2'>
                      {task.attachments.map((fileId) => (
                        <Badge key={fileId} variant='secondary' className='text-xs'>
                          {fileId}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
                {task.output_files && task.output_files.length > 0 && (
                  <div>
                    <p className='text-xs font-medium text-muted-foreground mb-1'>
                      Output Files
                    </p>
                    <div className='flex flex-wrap gap-2'>
                      {task.output_files.map((file) => (
                        file.url ? (
                          <a
                            key={file.path}
                            href={file.url}
                            download={file.name}
                            className='inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-semibold hover:bg-accent transition-colors'
                          >
                            <Download className='h-3 w-3' />
                            {file.name} ({(file.size / 1024).toFixed(1)} KB)
                          </a>
                        ) : (
                          <Badge key={file.path} variant='outline' className='text-xs'>
                            {file.name} ({(file.size / 1024).toFixed(1)} KB)
                          </Badge>
                        )
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        {/* 多轮输入框 */}
        {(task.status === 'running' || task.status === 'completed') && (
          <div className='sticky bottom-0 bg-background border-t pt-4'>
            <div className='flex gap-2'>
              <Textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder='Enter next prompt... (Enter to send, Shift+Enter for newline)'
                className='min-h-[60px] resize-none'
                disabled={appendTurn.isPending}
              />
              <Button
                onClick={handleSubmit}
                disabled={!prompt.trim() || appendTurn.isPending || !dockerAvailable}
                size='icon'
                className='h-[60px] w-[60px]'
              >
                {appendTurn.isPending ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Send className='h-4 w-4' />
                )}
              </Button>
            </div>
          </div>
        )}
      </Main>
    </>
  )
}
