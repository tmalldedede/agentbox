import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import {
  BASE_URL,
  EndpointHeader,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
  TryItField,
} from './shared'

export function ListFilesPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = []

  const handleExecute = useCallback(async () => {
    const response = await api.listFiles()
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/files"
        title="List Files"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Returns a list of all uploaded files, sorted by upload time (newest first).
        Files expire after 24 hours and are automatically deleted.
      </p>

      {/* Response */}
      <ResponseSection code={`[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "document.pdf",
    "size": 1024000,
    "mime_type": "application/pdf",
    "uploaded_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-01-16T10:30:00Z"
  },
  {
    "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "name": "data.json",
    "size": 2048,
    "mime_type": "application/json",
    "uploaded_at": "2024-01-15T09:15:00Z",
    "expires_at": "2024-01-16T09:15:00Z"
  }
]`} />

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/files \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

response = requests.get(
    "${BASE_URL}/files",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
)

files = response.json()["data"]
for f in files:
    print(f"{f['name']} ({f['size']} bytes) - ID: {f['id']}")`,
            javascript: `const response = await fetch("${BASE_URL}/files", {
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
});

const { data: files } = await response.json();
files.forEach(f => {
  console.log(\`\${f.name} (\${f.size} bytes) - ID: \${f.id}\`);
});`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/files"
        title="List Files"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
