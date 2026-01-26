import { useCallback, useEffect, useRef, useState } from 'react'
import { useChatStore } from '@/stores/chat-store'
import { useAgents } from '@/hooks/useAgents'
import { useCreateTask, useAppendTurn } from '@/hooks/useTasks'
import { api } from '@/services/api'
import type { TaskEvent } from '@/types'

export function useChat() {
  const {
    taskId,
    agentId,
    messages,
    isThinking,
    streamingText,
    attachments,
    setTaskId,
    setAgent,
    addMessage,
    setThinking,
    appendStreamingText,
    clearStreamingText,
    setConnected,
    clearChat,
    addAttachment,
    updateAttachment,
    removeAttachment,
    clearAttachments,
    getUploadedFileIds,
  } = useChatStore()

  const { data: agents = [] } = useAgents()
  const createTask = useCreateTask()
  const appendTurn = useAppendTurn()
  const eventSourceRef = useRef<EventSource | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  // Connect to SSE when taskId changes
  useEffect(() => {
    if (!taskId) {
      eventSourceRef.current?.close()
      setConnected(false)
      return
    }

    const es = api.streamTaskEvents(taskId)
    eventSourceRef.current = es

    es.onopen = () => {
      setConnected(true)
    }

    es.onmessage = (e) => {
      try {
        const event: TaskEvent = JSON.parse(e.data)
        handleEvent(event)
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      setConnected(false)
      es.close()
    }

    return () => {
      es.close()
      setConnected(false)
    }
  }, [taskId, setConnected])

  // Handle SSE events
  const handleEvent = useCallback(
    (event: TaskEvent) => {
      switch (event.type) {
        case 'task.started':
        case 'task.turn_started':
          setThinking(true)
          break

        case 'agent.thinking':
          setThinking(true)
          break

        case 'agent.message': {
          const data = event.data as { text?: string; content?: string }
          const text = data?.text || data?.content || ''
          if (text) {
            appendStreamingText(text)
          }
          break
        }

        case 'task.completed':
        case 'task.turn_completed': {
          // Finalize streaming text as a message
          const currentStreamingText = useChatStore.getState().streamingText
          if (currentStreamingText) {
            addMessage({
              role: 'assistant',
              content: currentStreamingText,
              status: 'sent',
            })
            clearStreamingText()
          }
          setThinking(false)
          break
        }

        case 'task.failed': {
          const data = event.data as { error?: string }
          // Finalize with error
          const currentStreamingText = useChatStore.getState().streamingText
          if (currentStreamingText) {
            addMessage({
              role: 'assistant',
              content: currentStreamingText,
              status: 'error',
            })
            clearStreamingText()
          } else {
            addMessage({
              role: 'assistant',
              content: data?.error || 'Task failed',
              status: 'error',
            })
          }
          setThinking(false)
          break
        }
      }
    },
    [addMessage, appendStreamingText, clearStreamingText, setThinking]
  )

  // Upload a single file
  const uploadFile = useCallback(
    async (attachmentId: string, file: File) => {
      updateAttachment(attachmentId, { status: 'uploading' })
      try {
        const uploadedFile = await api.uploadFile(file)
        updateAttachment(attachmentId, {
          status: 'uploaded',
          uploadedFile,
        })
        return uploadedFile
      } catch (error) {
        updateAttachment(attachmentId, {
          status: 'error',
          error: error instanceof Error ? error.message : 'Upload failed',
        })
        throw error
      }
    },
    [updateAttachment]
  )

  // Add files and start uploading
  const addFiles = useCallback(
    async (files: FileList) => {
      setIsUploading(true)
      const uploadPromises: Promise<void>[] = []

      for (let i = 0; i < files.length; i++) {
        const file = files[i]
        const attachmentId = addAttachment(file)

        // Start upload immediately
        const promise = uploadFile(attachmentId, file)
          .then(() => {})
          .catch(() => {
            // Error is already handled in uploadFile
          })
        uploadPromises.push(promise)
      }

      // Wait for all uploads to complete
      await Promise.all(uploadPromises)
      setIsUploading(false)
    },
    [addAttachment, uploadFile]
  )

  // Send message with attachments
  const sendMessage = useCallback(
    async (prompt: string) => {
      if (!agentId) {
        throw new Error('Please select an agent first')
      }

      // Get uploaded file IDs
      const attachmentIds = getUploadedFileIds()

      // Build message content with attachment info
      let messageContent = prompt
      if (attachmentIds.length > 0) {
        const currentAttachments = useChatStore.getState().attachments
        const fileNames = currentAttachments
          .filter((att) => att.status === 'uploaded')
          .map((att) => att.file.name)
        messageContent = `${prompt}\n\n📎 Attachments: ${fileNames.join(', ')}`
      }

      // Add user message immediately
      addMessage({
        role: 'user',
        content: messageContent,
        status: 'sent',
      })

      // Clear attachments after adding to message
      clearAttachments()

      setThinking(true)

      try {
        if (taskId) {
          // Append to existing task (multi-turn)
          // Note: attachments are only supported for the first turn
          await appendTurn.mutateAsync({ taskId, prompt })
        } else {
          // Create new task with attachments
          const task = await createTask.mutateAsync({
            agent_id: agentId,
            prompt,
            attachments: attachmentIds.length > 0 ? attachmentIds : undefined,
          })
          setTaskId(task.id)
        }
      } catch (error) {
        setThinking(false)
        addMessage({
          role: 'assistant',
          content: error instanceof Error ? error.message : 'Failed to send message',
          status: 'error',
        })
      }
    },
    [
      agentId,
      taskId,
      addMessage,
      setThinking,
      setTaskId,
      createTask,
      appendTurn,
      getUploadedFileIds,
      clearAttachments,
    ]
  )

  // Start new chat
  const newChat = useCallback(() => {
    eventSourceRef.current?.close()
    clearChat()
  }, [clearChat])

  return {
    // State
    taskId,
    agentId,
    messages,
    isThinking,
    streamingText,
    attachments,
    agents,
    isUploading,

    // Actions
    setAgent,
    sendMessage,
    newChat,
    clearChat,
    addFiles,
    removeAttachment,

    // Loading states
    isCreating: createTask.isPending,
    isAppending: appendTurn.isPending,
  }
}
