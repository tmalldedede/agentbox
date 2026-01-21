import { AlertCircle } from 'lucide-react'
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from '@/components/ui/alert'

interface ErrorAlertProps {
  error: Error | unknown
  title?: string
  className?: string
}

export function ErrorAlert({ error, title = '错误', className }: ErrorAlertProps) {
  const message = error instanceof Error
    ? error.message
    : typeof error === 'string'
      ? error
      : '发生了一个未知错误'

  return (
    <Alert variant="destructive" className={className}>
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  )
}

// 简化版本，用于内联错误提示
export function InlineError({ error }: { error: Error | unknown }) {
  const message = error instanceof Error
    ? error.message
    : typeof error === 'string'
      ? error
      : '发生了一个未知错误'

  return (
    <div className="flex items-center gap-2 p-3 rounded-lg bg-destructive/10 text-destructive">
      <AlertCircle className="w-4 h-4 flex-shrink-0" />
      <span className="text-sm">{message}</span>
    </div>
  )
}
