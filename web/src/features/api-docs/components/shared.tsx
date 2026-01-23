import { useState, useCallback, useEffect } from 'react'
import { Copy, CheckCheck, Play, Loader2, X } from 'lucide-react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export const BASE_URL = `${window.location.origin}/api/v1`

// ============================================================
// METHOD BADGE
// ============================================================

export function MethodBadge({ method, small }: { method: string; small?: boolean }) {
  const colors: Record<string, string> = {
    GET: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    POST: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    PUT: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    DEL: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    DELETE: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  }
  return (
    <span className={`inline-block font-mono font-bold rounded ${small ? 'text-[10px] px-1 py-0.5' : 'text-xs px-2 py-1'} ${colors[method] || ''}`}>
      {method}
    </span>
  )
}

// ============================================================
// ENDPOINT HEADER
// ============================================================

export function EndpointHeader({ method, path, title, onTryIt }: { method: string; path: string; title: string; onTryIt?: () => void }) {
  return (
    <div className="mb-6">
      <h2 className="text-2xl font-bold mb-3">{title}</h2>
      <div className="flex items-center gap-3 bg-muted rounded-lg px-4 py-2.5">
        <MethodBadge method={method} />
        <code className="text-sm font-mono flex-1">/v1{path}</code>
        {onTryIt && (
          <Button size="sm" onClick={onTryIt} className="shrink-0">
            Try it <Play className="w-3 h-3 ml-1" />
          </Button>
        )}
      </div>
    </div>
  )
}

// ============================================================
// PARAM TABLE
// ============================================================

export interface Param {
  name: string
  type: string
  required: boolean
  desc: string
  example?: string
}

export function ParamTable({ params }: { params: Param[] }) {
  return (
    <div className="rounded-lg border overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-muted/50">
          <tr>
            <th className="text-left px-4 py-2 font-medium">Name</th>
            <th className="text-left px-4 py-2 font-medium">Type</th>
            <th className="text-left px-4 py-2 font-medium">Description</th>
          </tr>
        </thead>
        <tbody>
          {params.map((p) => (
            <tr key={p.name} className="border-t">
              <td className="px-4 py-2.5">
                <code className="font-mono text-sm">{p.name}</code>
                {p.required && <span className="ml-2 text-xs text-red-500 font-medium">required</span>}
              </td>
              <td className="px-4 py-2.5 text-muted-foreground">{p.type}</td>
              <td className="px-4 py-2.5">
                <span className="text-muted-foreground">{p.desc}</span>
                {p.example && (
                  <div className="mt-1">
                    <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{p.example}</code>
                  </div>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ============================================================
// CODE BLOCK
// ============================================================

export function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const copy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative group min-w-0">
      <pre className="bg-muted rounded-lg p-4 pr-12 text-sm font-mono overflow-x-auto whitespace-pre leading-relaxed max-w-full">
        {code}
      </pre>
      <button
        onClick={copy}
        className="absolute top-3 right-3 p-1.5 rounded-md opacity-0 group-hover:opacity-100 bg-background/80 hover:bg-background border text-muted-foreground hover:text-foreground transition-all"
        title="Copy"
      >
        {copied ? <CheckCheck className="w-3.5 h-3.5 text-green-500" /> : <Copy className="w-3.5 h-3.5" />}
      </button>
    </div>
  )
}

// ============================================================
// EXAMPLES PANEL (language tabs)
// ============================================================

interface ExamplesProps {
  examples: {
    curl: string
    python: string
    javascript: string
  }
}

export function ExamplesPanel({ examples }: ExamplesProps) {
  const [lang, setLang] = useState<'curl' | 'python' | 'javascript'>('curl')

  const tabs = [
    { id: 'curl' as const, label: 'cURL' },
    { id: 'python' as const, label: 'Python' },
    { id: 'javascript' as const, label: 'JavaScript' },
  ]

  return (
    <div className="rounded-lg border overflow-hidden">
      <div className="flex items-center gap-0 border-b bg-muted/30">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setLang(t.id)}
            className={`px-4 py-2 text-sm font-medium transition-colors ${
              lang === t.id
                ? 'bg-background border-b-2 border-primary text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>
      <CodeBlock code={examples[lang]} />
    </div>
  )
}

// ============================================================
// SECTION WRAPPER
// ============================================================

export function Section({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-4">
      {title && <h3 className="font-semibold text-lg">{title}</h3>}
      {children}
    </div>
  )
}

export function ResponseSection({ code }: { code: string }) {
  return (
    <Section title="Response">
      <div className="space-y-2">
        <p className="text-sm text-muted-foreground">
          <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">200</span>
          Success
        </p>
        <CodeBlock code={code} />
      </div>
    </Section>
  )
}

// ============================================================
// TRY IT DIALOG
// ============================================================

export interface TryItField {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  required?: boolean
  placeholder?: string
  description?: string
  options?: { value: string; label: string }[]
}

interface TryItDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  method: string
  path: string
  title: string
  fields: TryItField[]
  onExecute: (values: Record<string, string>) => Promise<{ status: number; data: unknown }>
  generateCode?: (values: Record<string, string>) => string
}

export function TryItDialog({
  open,
  onOpenChange,
  method,
  path,
  title,
  fields,
  onExecute,
  generateCode,
}: TryItDialogProps) {
  const [values, setValues] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)
  const [response, setResponse] = useState<{ status: number; data: unknown } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [lang, setLang] = useState<'curl' | 'javascript'>('javascript')

  const updateValue = useCallback((name: string, value: string) => {
    setValues(prev => ({ ...prev, [name]: value }))
  }, [])

  const handleSend = async () => {
    setLoading(true)
    setResponse(null)
    setError(null)
    try {
      const result = await onExecute(values)
      setResponse(result)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Request failed')
    } finally {
      setLoading(false)
    }
  }

  // Generate curl command
  const curlCode = (() => {
    const bodyFields = fields.filter(f => f.name !== 'id')
    const body: Record<string, string> = {}
    bodyFields.forEach(f => {
      if (values[f.name]) body[f.name] = values[f.name]
    })
    const pathWithId = values.id ? path.replace(':id', values.id) : path
    if (method === 'GET' || method === 'DELETE') {
      return `curl ${method === 'DELETE' ? '-X DELETE ' : ''}${BASE_URL}${pathWithId} \\
  -H "Authorization: Bearer YOUR_API_KEY"`
    }
    return `curl -X ${method} ${BASE_URL}${pathWithId} \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '${JSON.stringify(body, null, 2)}'`
  })()

  const jsCode = generateCode?.(values) || (() => {
    const bodyFields = fields.filter(f => f.name !== 'id')
    const body: Record<string, string> = {}
    bodyFields.forEach(f => {
      if (values[f.name]) body[f.name] = values[f.name]
    })
    const pathWithId = values.id ? path.replace(':id', values.id) : path
    if (method === 'GET' || method === 'DELETE') {
      return `const response = await fetch("${BASE_URL}${pathWithId}", {
  ${method === 'DELETE' ? 'method: "DELETE",\n  ' : ''}headers: { "Authorization": "Bearer YOUR_API_KEY" },
});

const data = await response.json();
console.log(data);`
    }
    return `const response = await fetch("${BASE_URL}${pathWithId}", {
  method: "${method}",
  headers: {
    "Authorization": "Bearer YOUR_API_KEY",
    "Content-Type": "application/json",
  },
  body: JSON.stringify(${JSON.stringify(body, null, 2)}),
});

const data = await response.json();
console.log(data);`
  })()

  // Reset state when dialog opens
  useEffect(() => {
    if (open) {
      setValues({})
      setResponse(null)
      setError(null)
    }
  }, [open])

  if (!open) return null

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/50 animate-in fade-in-0" />
        <DialogPrimitive.Content className="fixed inset-4 z-50 flex flex-col bg-background rounded-lg shadow-lg animate-in fade-in-0 zoom-in-95">
          {/* Header */}
          <div className="flex items-center gap-4 px-6 py-4 border-b shrink-0">
            <MethodBadge method={method} />
            <code className="text-sm font-mono">/v1{path}</code>
            <div className="flex-1" />
            <Button onClick={handleSend} disabled={loading}>
              {loading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
              Send <Play className="w-3 h-3 ml-1" />
            </Button>
            <DialogPrimitive.Close className="p-2 rounded-md hover:bg-muted">
              <X className="w-4 h-4" />
            </DialogPrimitive.Close>
          </div>
          <DialogPrimitive.Title className="sr-only">{title}</DialogPrimitive.Title>

          {/* Body: two columns */}
          <div className="flex flex-1 min-h-0 overflow-hidden">
            {/* Left: Form */}
            <div className="w-[420px] shrink-0 border-r overflow-y-auto p-6 space-y-6">
              <div>
                <h3 className="font-semibold mb-4 flex items-center gap-2">
                  <span className="w-2 h-2 bg-red-500 rounded-full" />
                  Authorization
                </h3>
                <div className="space-y-2">
                  <Label className="text-sm">
                    API_KEY <span className="text-xs bg-muted px-1.5 py-0.5 rounded ml-2">string</span>
                    <span className="text-xs text-red-500 ml-2">required</span>
                  </Label>
                  <Input placeholder="Enter API_KEY" disabled className="font-mono text-sm" />
                  <p className="text-xs text-muted-foreground">Auto-filled from current session</p>
                </div>
              </div>

              {fields.length > 0 && (
                <div>
                  <h3 className="font-semibold mb-4">Body</h3>
                  <div className="space-y-4">
                    {fields.map((field) => (
                      <div key={field.name} className="space-y-2">
                        <Label className="text-sm">
                          {field.name}
                          <span className="text-xs bg-muted px-1.5 py-0.5 rounded ml-2">{field.type === 'select' ? 'enum' : 'string'}</span>
                          {field.required && <span className="text-xs text-red-500 ml-2">required</span>}
                        </Label>
                        {field.type === 'textarea' ? (
                          <Textarea
                            value={values[field.name] || ''}
                            onChange={(e) => updateValue(field.name, e.target.value)}
                            placeholder={field.placeholder}
                            rows={3}
                            className="font-mono text-sm"
                          />
                        ) : field.type === 'select' && field.options ? (
                          <Select
                            value={values[field.name] || ''}
                            onValueChange={(v) => updateValue(field.name, v)}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder={field.placeholder || `Select ${field.name}`} />
                            </SelectTrigger>
                            <SelectContent>
                              {field.options.map((opt) => (
                                <SelectItem key={opt.value} value={opt.value}>
                                  {opt.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        ) : (
                          <Input
                            value={values[field.name] || ''}
                            onChange={(e) => updateValue(field.name, e.target.value)}
                            placeholder={field.placeholder}
                            className="font-mono text-sm"
                          />
                        )}
                        {field.description && (
                          <p className="text-xs text-muted-foreground">{field.description}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Right: Code & Response */}
            <div className="flex-1 min-w-0 overflow-y-auto p-6 space-y-6 bg-muted/30">
              {/* Code Example */}
              <div className="min-w-0">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold">{title}</h3>
                  <div className="flex items-center gap-1 text-xs">
                    <button
                      onClick={() => setLang('javascript')}
                      className={`px-2 py-1 rounded ${lang === 'javascript' ? 'bg-background border' : 'text-muted-foreground hover:text-foreground'}`}
                    >
                      JavaScript
                    </button>
                    <button
                      onClick={() => setLang('curl')}
                      className={`px-2 py-1 rounded ${lang === 'curl' ? 'bg-background border' : 'text-muted-foreground hover:text-foreground'}`}
                    >
                      cURL
                    </button>
                  </div>
                </div>
                <div className="min-w-0 overflow-hidden">
                  <CodeBlock code={lang === 'curl' ? curlCode : jsCode} />
                </div>
              </div>

              {/* Response */}
              <div className="min-w-0">
                <div className="flex items-center gap-2 mb-3">
                  {response && (
                    <span className={`text-xs font-mono font-bold px-2 py-0.5 rounded ${
                      response.status >= 200 && response.status < 300
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                    }`}>
                      {response.status}
                    </span>
                  )}
                  {error && (
                    <span className="text-xs font-mono font-bold px-2 py-0.5 rounded bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400">
                      Error
                    </span>
                  )}
                  {!response && !error && (
                    <span className="text-xs text-muted-foreground">Response will appear here</span>
                  )}
                </div>
                {response && (
                  <div className="min-w-0 overflow-hidden">
                    <CodeBlock code={JSON.stringify(response.data, null, 2)} />
                  </div>
                )}
                {error && (
                  <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3">
                    <p className="text-sm text-destructive font-mono">{error}</p>
                  </div>
                )}
                {!response && !error && (
                  <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground text-sm">
                    Click "Send" to execute the request
                  </div>
                )}
              </div>
            </div>
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
