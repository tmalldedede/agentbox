import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
  TryItField,
} from './shared'

export function GetFilePage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: '550e8400-e29b-41d4-a716-446655440000',
      description: 'The file ID returned from Upload File',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.getFile(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/files/:id"
        title="Get File"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Get metadata for a specific file by its ID.
        Use this to check if a file exists and retrieve its details before downloading.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The file ID (UUID format)', example: '"550e8400-e29b-41d4-a716-446655440000"' },
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

      {/* Response Fields */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response Fields</h3>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Field</th>
                <th className="text-left px-4 py-2 font-medium">Type</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-t"><td className="px-4 py-2 font-mono">id</td><td className="px-4 py-2 text-muted-foreground">string</td><td className="px-4 py-2 text-muted-foreground">Unique file identifier (UUID)</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">name</td><td className="px-4 py-2 text-muted-foreground">string</td><td className="px-4 py-2 text-muted-foreground">Original filename</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">size</td><td className="px-4 py-2 text-muted-foreground">number</td><td className="px-4 py-2 text-muted-foreground">File size in bytes</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">mime_type</td><td className="px-4 py-2 text-muted-foreground">string</td><td className="px-4 py-2 text-muted-foreground">MIME type of the file</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">uploaded_at</td><td className="px-4 py-2 text-muted-foreground">string</td><td className="px-4 py-2 text-muted-foreground">ISO 8601 upload timestamp</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">expires_at</td><td className="px-4 py-2 text-muted-foreground">string</td><td className="px-4 py-2 text-muted-foreground">ISO 8601 expiration timestamp</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/files/550e8400-e29b-41d4-a716-446655440000 \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

file_id = "550e8400-e29b-41d4-a716-446655440000"

response = requests.get(
    f"${BASE_URL}/files/{file_id}",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
)

file_info = response.json()["data"]
print(f"Name: {file_info['name']}")
print(f"Size: {file_info['size']} bytes")
print(f"Type: {file_info['mime_type']}")`,
            javascript: `const fileId = "550e8400-e29b-41d4-a716-446655440000";

const response = await fetch(\`${BASE_URL}/files/\${fileId}\`, {
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
});

const { data } = await response.json();
console.log(\`Name: \${data.name}\`);
console.log(\`Size: \${data.size} bytes\`);`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/files/:id"
        title="Get File"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
