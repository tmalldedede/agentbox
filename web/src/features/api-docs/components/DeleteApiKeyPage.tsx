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

export function DeleteApiKeyPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'key-abc12345',
      description: 'The API key ID to delete',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.deleteAPIKey(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="DELETE"
        path="/auth/api-keys/:id"
        title="Delete API Key"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Permanently deletes an API key. Any applications using this key will immediately lose access.
        This action cannot be undone.
      </p>

      <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950/20 p-4">
        <p className="text-sm">
          <strong>Warning:</strong> Deleting an API key is immediate and irreversible.
          Any services or scripts using this key will stop working.
        </p>
      </div>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The API key ID', example: '"key-abc12345"' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "key-abc12345",
  "deleted": true
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X DELETE ${BASE_URL}/auth/api-keys/key-abc12345 \\
  -H "Authorization: Bearer YOUR_TOKEN"`,
            python: `import requests

response = requests.delete(
    "${BASE_URL}/auth/api-keys/key-abc12345",
    headers={"Authorization": "Bearer YOUR_TOKEN"},
)
result = response.json()["data"]
print(f"Deleted: {result['deleted']}")`,
            javascript: `const response = await fetch("${BASE_URL}/auth/api-keys/key-abc12345", {
  method: "DELETE",
  headers: { Authorization: "Bearer YOUR_TOKEN" },
});
const { data } = await response.json();
console.log(\`Deleted: \${data.deleted}\`);`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="DELETE"
        path="/auth/api-keys/:id"
        title="Delete API Key"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
