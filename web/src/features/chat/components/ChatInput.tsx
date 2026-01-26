import { useState, useRef, useEffect, type KeyboardEvent, type ChangeEvent } from 'react'
import { Send, Loader2, Paperclip, X, FileText, Image, File } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import type { PendingAttachment } from '@/stores/chat-store'

interface ChatInputProps {
  onSend: (message: string) => void
  onAddFiles: (files: FileList) => void
  onRemoveAttachment: (id: string) => void
  attachments: PendingAttachment[]
  disabled?: boolean
  loading?: boolean
  uploading?: boolean
  placeholder?: string
  className?: string
}

// 文件图标
function getFileIcon(file: File) {
  const type = file.type
  if (type.startsWith('image/')) {
    return Image
  }
  if (type.includes('pdf') || type.includes('document') || type.includes('text')) {
    return FileText
  }
  return File
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function ChatInput({
  onSend,
  onAddFiles,
  onRemoveAttachment,
  attachments,
  disabled,
  loading,
  uploading,
  placeholder = 'Type a message... (Enter to send, Shift+Enter for newline)',
  className,
}: ChatInputProps) {
  const [message, setMessage] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`
    }
  }, [message])

  const handleSend = () => {
    const trimmed = message.trim()
    if (!trimmed || disabled || loading) return

    onSend(trimmed)
    setMessage('')

    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleFileSelect = (e: ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      onAddFiles(files)
    }
    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleAttachClick = () => {
    fileInputRef.current?.click()
  }

  const hasAttachments = attachments.length > 0
  const hasPendingUploads = attachments.some(
    (att) => att.status === 'pending' || att.status === 'uploading'
  )

  return (
    <div className={cn('border-t bg-background', className)}>
      {/* Attachments preview */}
      {hasAttachments && (
        <div className='mx-auto max-w-3xl border-b px-4 py-2'>
          <div className='flex flex-wrap gap-2'>
            {attachments.map((att) => {
              const FileIcon = getFileIcon(att.file)
              const isUploading = att.status === 'uploading'
              const isError = att.status === 'error'
              const isUploaded = att.status === 'uploaded'

              return (
                <div
                  key={att.id}
                  className={cn(
                    'group relative flex items-center gap-2 rounded-lg border px-3 py-2 text-sm',
                    isError && 'border-destructive bg-destructive/10',
                    isUploaded && 'border-green-500 bg-green-50 dark:bg-green-950/20',
                    isUploading && 'animate-pulse'
                  )}
                >
                  <FileIcon className='h-4 w-4 shrink-0 text-muted-foreground' />
                  <div className='flex flex-col overflow-hidden'>
                    <span className='truncate max-w-[150px] font-medium'>
                      {att.file.name}
                    </span>
                    <span className='text-xs text-muted-foreground'>
                      {isUploading && 'Uploading...'}
                      {isError && (att.error || 'Upload failed')}
                      {isUploaded && formatFileSize(att.file.size)}
                      {att.status === 'pending' && 'Pending...'}
                    </span>
                  </div>
                  <Button
                    variant='ghost'
                    size='icon'
                    className='h-5 w-5 shrink-0 opacity-60 hover:opacity-100'
                    onClick={() => onRemoveAttachment(att.id)}
                    disabled={isUploading}
                  >
                    <X className='h-3 w-3' />
                  </Button>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Input area */}
      <div className='p-4'>
        <div className='mx-auto flex max-w-3xl items-end gap-2'>
          {/* Hidden file input */}
          <input
            ref={fileInputRef}
            type='file'
            multiple
            className='hidden'
            onChange={handleFileSelect}
            accept='*/*'
          />

          {/* Attachment button */}
          <Button
            variant='ghost'
            size='icon'
            className='shrink-0'
            onClick={handleAttachClick}
            disabled={disabled || loading}
            title='Add attachments'
          >
            <Paperclip className='h-5 w-5' />
          </Button>

          {/* Message input */}
          <div className='relative flex-1'>
            <Textarea
              ref={textareaRef}
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={placeholder}
              disabled={disabled || loading}
              className='min-h-[44px] max-h-[200px] resize-none pr-12'
              rows={1}
            />
          </div>

          {/* Send button */}
          <Button
            onClick={handleSend}
            disabled={!message.trim() || disabled || loading || uploading || hasPendingUploads}
            size='icon'
            className='h-11 w-11 shrink-0'
          >
            {loading || uploading ? (
              <Loader2 className='h-5 w-5 animate-spin' />
            ) : (
              <Send className='h-5 w-5' />
            )}
          </Button>
        </div>

        <p className='mx-auto mt-2 max-w-3xl text-center text-xs text-muted-foreground'>
          Press Enter to send, Shift+Enter for new line
          {hasAttachments && ` • ${attachments.length} file(s) attached`}
        </p>
      </div>
    </div>
  )
}
