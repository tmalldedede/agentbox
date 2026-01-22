import { useEffect, useRef, useState } from 'react'
import { Terminal, Pause, Play, Trash2, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { useWebSocket, StreamMessage } from '@/hooks/useWebSocket'

interface LogViewerProps {
  sessionId: string
  className?: string
  autoScroll?: boolean
  maxLines?: number
}

export function LogViewer({
  sessionId,
  className,
  autoScroll = true,
  maxLines = 1000,
}: LogViewerProps) {
  const [isPaused, setIsPaused] = useState(false)
  const [filter, setFilter] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Build WebSocket URL
  const wsUrl = sessionId
    ? `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/sessions/${sessionId}/stream`
    : null

  const { isConnected, messages, clearMessages } = useWebSocket(wsUrl, {
    autoReconnect: true,
  })

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && !isPaused && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, autoScroll, isPaused])

  // Filter messages
  const filteredMessages = messages.filter((msg) => {
    if (!filter) return true
    return msg.content?.toLowerCase().includes(filter.toLowerCase())
  })

  // Limit messages
  const displayMessages = filteredMessages.slice(-maxLines)

  const getMessageStyle = (msg: StreamMessage) => {
    switch (msg.type) {
      case 'error':
        return 'text-red-500'
      case 'start':
        return 'text-blue-500 font-semibold'
      case 'done':
        return 'text-green-500 font-semibold'
      default:
        return 'text-foreground'
    }
  }

  const formatTimestamp = (ts: number) => {
    return new Date(ts).toLocaleTimeString()
  }

  const handleDownload = () => {
    const content = messages
      .map((m) => `[${formatTimestamp(m.timestamp)}] ${m.type}: ${m.content || ''}`)
      .join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${sessionId}-${Date.now()}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Card className={cn('flex flex-col', className)}>
      <CardHeader className="py-3 px-4 border-b">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <Terminal className="h-4 w-4" />
            Real-time Logs
            <Badge variant={isConnected ? 'default' : 'secondary'}>
              {isConnected ? 'Connected' : 'Disconnected'}
            </Badge>
          </CardTitle>
          <div className="flex items-center gap-2">
            <input
              type="text"
              placeholder="Filter logs..."
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="h-8 px-2 text-sm border rounded-md w-40"
            />
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setIsPaused(!isPaused)}
              title={isPaused ? 'Resume auto-scroll' : 'Pause auto-scroll'}
            >
              {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={clearMessages}
              title="Clear logs"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={handleDownload}
              title="Download logs"
            >
              <Download className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="p-0 flex-1 min-h-0" ref={containerRef}>
        <div
          ref={scrollRef}
          className="h-full overflow-auto bg-black/95 text-sm font-mono p-4"
          style={{ maxHeight: '400px' }}
        >
          {displayMessages.length === 0 ? (
            <div className="text-muted-foreground text-center py-8">
              Waiting for logs...
            </div>
          ) : (
            displayMessages.map((msg, idx) => (
              <div key={idx} className={cn('py-0.5', getMessageStyle(msg))}>
                <span className="text-muted-foreground mr-2">
                  [{formatTimestamp(msg.timestamp)}]
                </span>
                {msg.type === 'start' && (
                  <span className="mr-2">[START {msg.execution_id}]</span>
                )}
                {msg.type === 'done' && (
                  <span className="mr-2">[DONE {msg.execution_id}]</span>
                )}
                {msg.type === 'error' && <span className="mr-2">[ERROR]</span>}
                <span>{msg.content}</span>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export default LogViewer
