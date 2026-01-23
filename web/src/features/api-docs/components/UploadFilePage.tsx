import { useState, useRef } from 'react'
import { Loader2, Play, X, Upload } from 'lucide-react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { Button } from '@/components/ui/button'
import { api } from '@/services/api'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
  MethodBadge,
  CodeBlock,
} from './shared'

export function UploadFilePage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/files"
        title="Upload File"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Upload a file to the server. Files can be attached to tasks or used as input for agent operations.
        The file will be stored temporarily and expires after 24 hours.
      </p>

      {/* Request Body */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <div className="rounded-lg border p-4 space-y-2 text-sm">
          <p className="text-muted-foreground">
            <code className="bg-muted px-1.5 py-0.5 rounded">Content-Type: multipart/form-data</code>
          </p>
        </div>
        <ParamTable params={[
          { name: 'file', type: 'file', required: true, desc: 'The file to upload (max 100MB)' },
        ]} />
      </div>

      {/* Response */}
      <ResponseSection code={`{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "document.pdf",
  "size": 1024000,
  "mime_type": "application/pdf",
  "uploaded_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-16T10:30:00Z"
}`} />

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/files \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -F "file=@/path/to/document.pdf"`,
            python: `import requests

with open("/path/to/document.pdf", "rb") as f:
    response = requests.post(
        "${BASE_URL}/files",
        headers={"Authorization": "Bearer YOUR_API_KEY"},
        files={"file": f},
    )

file_info = response.json()
print(f"File ID: {file_info['data']['id']}")
print(f"Name: {file_info['data']['name']}")
print(f"Size: {file_info['data']['size']} bytes")`,
            javascript: `const formData = new FormData();
formData.append("file", fileInput.files[0]);

const response = await fetch("${BASE_URL}/files", {
  method: "POST",
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
  body: formData,
});

const { data } = await response.json();
console.log(\`File ID: \${data.id}\`);
console.log(\`Name: \${data.name}\`);`,
          }}
        />
      </div>

      {/* Upload Try It Dialog */}
      <UploadTryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
      />
    </div>
  )
}

// Special dialog for file upload
function UploadTryItDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [response, setResponse] = useState<{ status: number; data: unknown } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [lang, setLang] = useState<'curl' | 'javascript'>('javascript')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0])
      setResponse(null)
      setError(null)
    }
  }

  const handleSend = async () => {
    if (!file) {
      setError('Please select a file')
      return
    }
    setLoading(true)
    setResponse(null)
    setError(null)
    try {
      const result = await api.uploadFile(file)
      setResponse({ status: 201, data: result })
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Upload failed')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    setFile(null)
    setResponse(null)
    setError(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const curlCode = `curl -X POST ${BASE_URL}/files \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -F "file=@${file?.name || 'your-file.pdf'}"`

  const jsCode = `const formData = new FormData();
formData.append("file", fileInput.files[0]); // ${file?.name || 'your-file.pdf'}

const response = await fetch("${BASE_URL}/files", {
  method: "POST",
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
  body: formData,
});

const data = await response.json();
console.log(data);`

  if (!open) return null

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/50 animate-in fade-in-0" />
        <DialogPrimitive.Content className="fixed inset-4 z-50 flex flex-col bg-background rounded-lg shadow-lg animate-in fade-in-0 zoom-in-95">
          {/* Header */}
          <div className="flex items-center gap-4 px-6 py-4 border-b shrink-0">
            <MethodBadge method="POST" />
            <code className="text-sm font-mono">/v1/files</code>
            <div className="flex-1" />
            <Button onClick={handleSend} disabled={loading || !file}>
              {loading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
              Send <Play className="w-3 h-3 ml-1" />
            </Button>
            <DialogPrimitive.Close className="p-2 rounded-md hover:bg-muted">
              <X className="w-4 h-4" />
            </DialogPrimitive.Close>
          </div>
          <DialogPrimitive.Title className="sr-only">Upload File</DialogPrimitive.Title>

          {/* Body: two columns */}
          <div className="flex flex-1 min-h-0 overflow-hidden">
            {/* Left: Form */}
            <div className="w-[420px] shrink-0 border-r overflow-y-auto p-6 space-y-6">
              <div>
                <h3 className="font-semibold mb-4">File Upload</h3>
                <div className="space-y-4">
                  <div
                    className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
                      file ? 'border-primary bg-primary/5' : 'border-muted-foreground/25 hover:border-muted-foreground/50'
                    }`}
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <input
                      ref={fileInputRef}
                      type="file"
                      className="hidden"
                      onChange={handleFileChange}
                    />
                    <Upload className="w-8 h-8 mx-auto mb-3 text-muted-foreground" />
                    {file ? (
                      <div>
                        <p className="font-medium">{file.name}</p>
                        <p className="text-sm text-muted-foreground mt-1">
                          {(file.size / 1024).toFixed(1)} KB
                        </p>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="mt-2"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleReset()
                          }}
                        >
                          Change file
                        </Button>
                      </div>
                    ) : (
                      <div>
                        <p className="text-muted-foreground">Click to select a file</p>
                        <p className="text-xs text-muted-foreground mt-1">Max 100MB</p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* Right: Code & Response */}
            <div className="flex-1 min-w-0 overflow-y-auto p-6 space-y-6 bg-muted/30">
              {/* Code Example */}
              <div className="min-w-0">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold">Upload File</h3>
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
                    Select a file and click "Send" to upload
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
