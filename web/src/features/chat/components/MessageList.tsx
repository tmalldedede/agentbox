import { useEffect, useRef } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { ChatMessage } from '@/types'
import { MessageItem } from './MessageItem'
import { ThinkingIndicator } from './ThinkingIndicator'

interface MessageListProps {
  messages: ChatMessage[]
  isThinking: boolean
  streamingText: string
}

export function MessageList({
  messages,
  isThinking,
  streamingText,
}: MessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  // Auto scroll to bottom when new messages arrive
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, streamingText, isThinking])

  // Show streaming message as a temporary message
  const displayMessages = [...messages]
  if (streamingText) {
    displayMessages.push({
      id: 'streaming',
      role: 'assistant',
      content: streamingText,
      timestamp: new Date(),
      status: 'streaming',
    })
  }

  if (displayMessages.length === 0 && !isThinking) {
    return (
      <div className='flex flex-1 items-center justify-center'>
        <div className='text-center'>
          <p className='text-lg font-medium text-muted-foreground'>
            Start a conversation
          </p>
          <p className='text-sm text-muted-foreground'>
            Select an agent and send a message to begin
          </p>
        </div>
      </div>
    )
  }

  return (
    <ScrollArea ref={scrollRef} className='flex-1 p-4'>
      <div className='mx-auto max-w-3xl space-y-4'>
        {displayMessages.map((message) => (
          <MessageItem
            key={message.id}
            message={message}
            isStreaming={message.status === 'streaming'}
          />
        ))}
        {isThinking && !streamingText && <ThinkingIndicator />}
        <div ref={bottomRef} />
      </div>
    </ScrollArea>
  )
}
