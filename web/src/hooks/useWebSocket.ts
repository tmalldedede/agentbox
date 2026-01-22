import { useState, useEffect, useRef, useCallback } from 'react'

export interface StreamMessage {
  type: 'start' | 'message' | 'error' | 'done' | 'ping'
  content?: string
  timestamp: number
  execution_id?: string
}

export interface UseWebSocketOptions {
  onMessage?: (msg: StreamMessage) => void
  onError?: (error: Event) => void
  onOpen?: () => void
  onClose?: () => void
  autoReconnect?: boolean
  reconnectInterval?: number
}

export function useWebSocket(url: string | null, options: UseWebSocketOptions = {}) {
  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<StreamMessage[]>([])
  const [error, setError] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const {
    onMessage,
    onError,
    onOpen,
    onClose,
    autoReconnect = false,
    reconnectInterval = 3000,
  } = options

  const connect = useCallback(() => {
    if (!url) return

    try {
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        setError(null)
        onOpen?.()
      }

      ws.onmessage = (event) => {
        try {
          const msg: StreamMessage = JSON.parse(event.data)
          if (msg.type !== 'ping') {
            setMessages((prev) => [...prev, msg])
          }
          onMessage?.(msg)
        } catch {
          console.error('Failed to parse WebSocket message:', event.data)
        }
      }

      ws.onerror = (event) => {
        setError('WebSocket connection error')
        onError?.(event)
      }

      ws.onclose = () => {
        setIsConnected(false)
        onClose?.()

        if (autoReconnect) {
          reconnectTimeoutRef.current = setTimeout(connect, reconnectInterval)
        }
      }
    } catch (err) {
      setError(`Failed to connect: ${err}`)
    }
  }, [url, onMessage, onError, onOpen, onClose, autoReconnect, reconnectInterval])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }, [])

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  const clearMessages = useCallback(() => {
    setMessages([])
  }, [])

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  return {
    isConnected,
    messages,
    error,
    send,
    connect,
    disconnect,
    clearMessages,
  }
}

export default useWebSocket
