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

export function CreateApiKeyPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'name',
      label: 'name',
      type: 'text',
      required: true,
      placeholder: 'My API Key',
      description: 'A friendly name to identify this key',
    },
    {
      name: 'expires_in',
      label: 'expires_in',
      type: 'select',
      required: false,
      placeholder: 'Never expire',
      description: 'Number of days until expiration (0 = never)',
      options: [
        { value: '0', label: 'Never expire' },
        { value: '7', label: '7 days' },
        { value: '30', label: '30 days' },
        { value: '90', label: '90 days' },
        { value: '365', label: '1 year' },
      ],
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.createAPIKey({
      name: values.name,
      expires_in: values.expires_in ? parseInt(values.expires_in) : 0,
    })
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/auth/api-keys"
        title="Create API Key"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Creates a new API key for programmatic access. The full key is only returned once in the response -
        make sure to copy and store it securely.
      </p>

      <div className="rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950/20 p-4">
        <p className="text-sm">
          <strong>Important:</strong> The full API key is only shown once when created.
          Store it securely - you won't be able to retrieve it later.
        </p>
      </div>

      {/* Body Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <ParamTable params={[
          { name: 'name', type: 'string', required: true, desc: 'A friendly name to identify this key', example: '"Production API"' },
          { name: 'expires_in', type: 'number', required: false, desc: 'Days until expiration. 0 or omit for never expire', example: '90' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "key-abc12345",
  "name": "Production API",
  "key_prefix": "ab_8f3d2e1...",
  "key": "ab_8f3d2e1a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2",
  "expires_at": "2024-04-15T10:30:00Z",
  "created_at": "2024-01-15T10:30:00Z"
}`} />
        <p className="text-sm text-muted-foreground">
          Note: The <code className="bg-muted px-1 rounded">key</code> field containing the full API key
          is only included in the create response.
        </p>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# Create API key (never expires)
curl -X POST ${BASE_URL}/auth/api-keys \\
  -H "Authorization: Bearer YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Production API"
  }'

# Create API key that expires in 90 days
curl -X POST ${BASE_URL}/auth/api-keys \\
  -H "Authorization: Bearer YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Temporary Key",
    "expires_in": 90
  }'`,
            python: `import requests

response = requests.post(
    "${BASE_URL}/auth/api-keys",
    headers={"Authorization": "Bearer YOUR_TOKEN"},
    json={
        "name": "Production API",
        "expires_in": 0,  # Never expires
    },
)
data = response.json()["data"]
api_key = data["key"]  # Save this! Only shown once
print(f"API Key: {api_key}")

# Use the API key for future requests
response = requests.get(
    "${BASE_URL}/tasks",
    headers={"Authorization": f"Bearer {api_key}"},
)`,
            javascript: `const response = await fetch("${BASE_URL}/auth/api-keys", {
  method: "POST",
  headers: {
    "Authorization": "Bearer YOUR_TOKEN",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    name: "Production API",
    expires_in: 0,  // Never expires
  }),
});
const { data } = await response.json();
const apiKey = data.key;  // Save this! Only shown once
console.log(\`API Key: \${apiKey}\`);

// Use the API key for future requests
const tasksRes = await fetch("${BASE_URL}/tasks", {
  headers: { Authorization: \`Bearer \${apiKey}\` },
});`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/auth/api-keys"
        title="Create API Key"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
