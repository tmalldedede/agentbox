import { useEffect, useRef, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ChevronDown, ChevronRight, Radio } from 'lucide-react'

type ContainerLogsProps = {
  logs: string[]
  connected: boolean
}

type LogEntry = {
  raw: string
  content: string
  parsed?: any
  isJson: boolean
  timestamp?: string
}

export function ContainerLogs({ logs, connected }: ContainerLogsProps) {
  const logsEndRef = useRef<HTMLDivElement>(null)
  const [expandedLines, setExpandedLines] = useState<Set<number>>(new Set())
  const [parsedLogs, setParsedLogs] = useState<LogEntry[]>([])

  // 解析日志
  useEffect(() => {
    const entries: LogEntry[] = logs
      .filter(log => log != null && log !== '') // 过滤空日志
      .map(log => {
        // 尝试提取时间戳 (Docker 格式: 2024-01-27T10:23:45.123456789Z ...)
        const timestampMatch = log.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(.*)/)
        const timestamp = timestampMatch ? timestampMatch[1] : undefined
        const content = timestampMatch ? timestampMatch[2] : log

        // 检查是否为 JSON（查找第一个 { 的位置）
        const jsonStart = content.indexOf('{')
        if (jsonStart >= 0) {
          const jsonContent = content.substring(jsonStart)
          try {
            const parsed = JSON.parse(jsonContent)
            return {
              raw: log,
              content,
              parsed,
              isJson: true,
              timestamp
            }
          } catch {
            // 不是有效的 JSON，按普通文本处理
          }
        }

        return {
          raw: log,
          content,
          isJson: false,
          timestamp
        }
      })
    setParsedLogs(entries)
  }, [logs])

  // 自动滚动到底部
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const toggleExpand = (index: number) => {
    setExpandedLines(prev => {
      const next = new Set(prev)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  const formatTimestamp = (ts: string) => {
    try {
      return new Date(ts).toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    } catch {
      return ts
    }
  }

  const getEventTypeColor = (type: string) => {
    if (type.includes('error') || type.includes('failed')) return 'text-red-500'
    if (type.includes('completed') || type.includes('success')) return 'text-green-500'
    if (type.includes('started') || type.includes('thinking')) return 'text-blue-500'
    if (type.includes('tool') || type.includes('command')) return 'text-orange-500'
    return 'text-muted-foreground'
  }

  if (parsedLogs.length === 0) {
    return (
      <div className='space-y-2'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
            <span className='text-sm text-muted-foreground'>
              {connected ? 'SSE Connected' : 'Disconnected'}
            </span>
          </div>
        </div>
        <div className='h-[500px] overflow-y-auto bg-muted rounded-lg p-4 font-mono text-xs'>
          <div className='text-muted-foreground text-center py-8'>
            Container logs will appear here during execution
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className='space-y-2'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Radio className={`h-4 w-4 ${connected ? 'text-green-500 animate-pulse' : 'text-gray-400'}`} />
          <span className='text-sm text-muted-foreground'>
            {connected ? 'Streaming' : 'Disconnected'}
          </span>
        </div>
        <span className='text-xs text-muted-foreground'>{parsedLogs.length} entries</span>
      </div>

      <div className='h-[500px] overflow-y-auto bg-muted rounded-lg p-3 font-mono text-xs space-y-1'>
        {parsedLogs.map((entry, index) => {
          if (!entry.isJson) {
            // 普通文本日志
            const content = entry.content || entry.raw || ''
            return (
              <div
                key={index}
                className={`py-0.5 ${
                  content.startsWith('[System]') ? 'text-blue-500' :
                  content.startsWith('[Connected]') ? 'text-green-500' :
                  content.startsWith('[Error]') ? 'text-red-500' :
                  'text-foreground/80'
                }`}
              >
                {entry.timestamp && (
                  <span className='text-muted-foreground mr-2'>
                    [{formatTimestamp(entry.timestamp)}]
                  </span>
                )}
                {content}
              </div>
            )
          }

          // JSON 日志
          const isExpanded = expandedLines.has(index)
          const eventType = entry.parsed?.type || 'unknown'

          return (
            <div key={index} className='border-l-2 border-primary/30 pl-2 py-1'>
              <Button
                variant='ghost'
                size='sm'
                className='h-5 w-full justify-start gap-2 px-1 hover:bg-primary/10'
                onClick={() => toggleExpand(index)}
              >
                {isExpanded ? (
                  <ChevronDown className='h-3 w-3' />
                ) : (
                  <ChevronRight className='h-3 w-3' />
                )}
                {entry.timestamp && (
                  <span className='text-muted-foreground text-[10px]'>
                    {formatTimestamp(entry.timestamp)}
                  </span>
                )}
                <Badge variant='outline' className={`text-[10px] px-1.5 py-0 ${getEventTypeColor(eventType)}`}>
                  {eventType}
                </Badge>
                {entry.parsed?.item?.type && (
                  <span className='text-muted-foreground text-[10px]'>
                    {entry.parsed.item.type}
                  </span>
                )}
              </Button>

              {isExpanded && (
                <pre className='mt-1 p-2 bg-background rounded text-[10px] overflow-x-auto'>
                  {JSON.stringify(entry.parsed, null, 2)}
                </pre>
              )}
            </div>
          )
        })}
        <div ref={logsEndRef} />
      </div>
    </div>
  )
}
