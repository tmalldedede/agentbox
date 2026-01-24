import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import {
  BASE_URL,
  EndpointHeader,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
} from './shared'

export function ApiKeysPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const handleExecute = useCallback(async () => {
    const response = await api.listAPIKeys()
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/auth/api-keys"
        title="List API Keys"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        List all API keys belonging to the authenticated user. Returns key metadata only -
        the full key value is only shown once when created.
      </p>

      <div className="rounded-lg border border-blue-200 bg-blue-50 dark:border-blue-900 dark:bg-blue-950/20 p-4">
        <p className="text-sm">
          <strong>API Keys vs JWT Tokens:</strong> API keys are long-lived credentials ideal for
          programmatic access. They start with <code className="bg-muted px-1 rounded">ab_</code> prefix
          and inherit the permissions of the user who created them.
        </p>
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`[
  {
    "id": "key-abc12345",
    "name": "Production API",
    "key_prefix": "ab_8f3d2e1...",
    "last_used_at": "2024-01-15T10:30:00Z",
    "expires_at": null,
    "created_at": "2024-01-01T00:00:00Z"
  },
  {
    "id": "key-def67890",
    "name": "CI/CD Pipeline",
    "key_prefix": "ab_9c4e3f2...",
    "last_used_at": null,
    "expires_at": "2024-04-01T00:00:00Z",
    "created_at": "2024-01-10T12:00:00Z"
  }
]`} />
      </div>

      {/* Response Fields */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response Fields</h3>
        <div className="overflow-x-auto rounded-lg border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-2 text-left font-medium">Field</th>
                <th className="px-4 py-2 text-left font-medium">Type</th>
                <th className="px-4 py-2 text-left font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b">
                <td className="px-4 py-2 font-mono text-xs">id</td>
                <td className="px-4 py-2 text-muted-foreground">string</td>
                <td className="px-4 py-2">Unique key identifier</td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-2 font-mono text-xs">name</td>
                <td className="px-4 py-2 text-muted-foreground">string</td>
                <td className="px-4 py-2">User-defined name for the key</td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-2 font-mono text-xs">key_prefix</td>
                <td className="px-4 py-2 text-muted-foreground">string</td>
                <td className="px-4 py-2">First 10 characters of the key (for identification)</td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-2 font-mono text-xs">last_used_at</td>
                <td className="px-4 py-2 text-muted-foreground">string | null</td>
                <td className="px-4 py-2">ISO timestamp of last API call using this key</td>
              </tr>
              <tr className="border-b">
                <td className="px-4 py-2 font-mono text-xs">expires_at</td>
                <td className="px-4 py-2 text-muted-foreground">string | null</td>
                <td className="px-4 py-2">ISO timestamp when key expires (null = never)</td>
              </tr>
              <tr>
                <td className="px-4 py-2 font-mono text-xs">created_at</td>
                <td className="px-4 py-2 text-muted-foreground">string</td>
                <td className="px-4 py-2">ISO timestamp when key was created</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/auth/api-keys \\
  -H "Authorization: Bearer YOUR_TOKEN"`,
            python: `import requests

response = requests.get(
    "${BASE_URL}/auth/api-keys",
    headers={"Authorization": "Bearer YOUR_TOKEN"},
)
keys = response.json()["data"]
for key in keys:
    print(f"{key['name']}: {key['key_prefix']}")`,
            javascript: `const response = await fetch("${BASE_URL}/auth/api-keys", {
  headers: { Authorization: "Bearer YOUR_TOKEN" },
});
const { data: keys } = await response.json();
keys.forEach(key => console.log(\`\${key.name}: \${key.key_prefix}\`));`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/auth/api-keys"
        title="List API Keys"
        fields={[]}
        onExecute={handleExecute}
      />
    </div>
  )
}
