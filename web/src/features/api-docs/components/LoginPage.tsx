import { useState, useCallback } from 'react'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
  TryItField,
} from './shared'

export function LoginPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'username',
      label: 'username',
      type: 'text',
      required: true,
      placeholder: 'admin',
      description: 'The username to authenticate',
    },
    {
      name: 'password',
      label: 'password',
      type: 'text',
      required: true,
      placeholder: 'admin123',
      description: 'The user password',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await fetch(`${BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: values.username,
        password: values.password,
      }),
    })
    const data = await response.json()
    return { status: response.status, data: data.data || data }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/auth/login"
        title="Login"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Authenticate with username and password to obtain a JWT token. The token expires after 24 hours.
      </p>

      <div className="rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950/20 p-4">
        <p className="text-sm">
          <strong>Tip:</strong> For programmatic access, consider using{' '}
          <a href="/api-docs/api-keys" className="text-primary underline">API Keys</a>{' '}
          instead of JWT tokens. API Keys don't expire and are easier to manage.
        </p>
      </div>

      {/* Body Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <ParamTable params={[
          { name: 'username', type: 'string', required: true, desc: 'The username to authenticate', example: '"admin"' },
          { name: 'password', type: 'string', required: true, desc: 'The user password', example: '"admin123"' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1705440000,
  "user": {
    "id": "user-abc123",
    "username": "admin",
    "role": "admin"
  }
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# Use the token in subsequent requests:
curl ${BASE_URL}/tasks \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN"`,
            python: `import requests

# Login
response = requests.post(
    "${BASE_URL}/auth/login",
    json={"username": "admin", "password": "admin123"},
)
token = response.json()["data"]["token"]

# Use token for authenticated requests
headers = {"Authorization": f"Bearer {token}"}
tasks = requests.get("${BASE_URL}/tasks", headers=headers)`,
            javascript: `// Login
const loginRes = await fetch("${BASE_URL}/auth/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    username: "admin",
    password: "admin123",
  }),
});
const { data: { token } } = await loginRes.json();

// Use token for authenticated requests
const tasksRes = await fetch("${BASE_URL}/tasks", {
  headers: { Authorization: \`Bearer \${token}\` },
});`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/auth/login"
        title="Login"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
