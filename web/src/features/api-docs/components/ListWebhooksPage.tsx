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

export function ListWebhooksPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = []

  const handleExecute = useCallback(async () => {
    const response = await api.listWebhooks()
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/webhooks"
        title="List Webhooks"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Returns a list of all configured webhook endpoints.
      </p>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <p className="text-sm text-muted-foreground">
          <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">200</span>
          Success
        </p>
        <ResponseSection code={`[
  {
    "id": "wh-a1b2c3d4-...",
    "url": "https://example.com/webhook",
    "events": ["task.created", "task.completed"],
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "wh-e5f6g7h8-...",
    "url": "https://another.com/notify",
    "events": [],
    "is_active": false,
    "created_at": "2024-01-14T09:00:00Z",
    "updated_at": "2024-01-15T08:00:00Z"
  }
]`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/webhooks`,
            python: `import requests

response = requests.get("${BASE_URL}/webhooks")
webhooks = response.json()["data"]

for wh in webhooks:
    print(f"{wh['id']} -> {wh['url']} ({'active' if wh['is_active'] else 'inactive'})")`,
            javascript: `const response = await fetch("${BASE_URL}/webhooks");
const { data: webhooks } = await response.json();

webhooks.forEach(wh => {
  console.log(\`\${wh.id} -> \${wh.url} (\${wh.is_active ? 'active' : 'inactive'})\`);
});`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/webhooks"
        title="List Webhooks"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
