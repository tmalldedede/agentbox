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

export function DeleteFilePage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: '550e8400-e29b-41d4-a716-446655440000',
      description: 'The file ID to delete',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.deleteFile(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="DELETE"
        path="/files/:id"
        title="Delete File"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Permanently delete a file from the server.
        This action cannot be undone. Files are also automatically deleted after 24 hours.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The file ID to delete', example: '"550e8400-e29b-41d4-a716-446655440000"' },
        ]} />
      </div>

      {/* Response */}
      <ResponseSection code={`{
  "deleted": "550e8400-e29b-41d4-a716-446655440000"
}`} />

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X DELETE ${BASE_URL}/files/550e8400-e29b-41d4-a716-446655440000 \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

file_id = "550e8400-e29b-41d4-a716-446655440000"

response = requests.delete(
    f"${BASE_URL}/files/{file_id}",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
)

if response.status_code == 200:
    print("File deleted successfully")`,
            javascript: `const fileId = "550e8400-e29b-41d4-a716-446655440000";

const response = await fetch(\`${BASE_URL}/files/\${fileId}\`, {
  method: "DELETE",
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
});

if (response.ok) {
  console.log("File deleted successfully");
}`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="DELETE"
        path="/files/:id"
        title="Delete File"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
