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

export function GetWebhookPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'webhook_id',
      type: 'text',
      required: true,
      placeholder: 'wh-a1b2c3d4-...',
      description: 'The unique identifier of the webhook',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.getWebhook(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/webhooks/:id"
        title="Get Webhook"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Retrieves the details of a specific webhook by its ID.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The unique identifier of the webhook to retrieve', example: '"wh-a1b2c3d4-..."' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <p className="text-sm text-muted-foreground">
          <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">200</span>
          Success
        </p>
        <ResponseSection code={`{
  "id": "wh-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "url": "https://example.com/webhook",
  "events": ["task.created", "task.completed"],
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}`} />
        <p className="text-sm text-muted-foreground mt-4">
          <span className="inline-block bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">404</span>
          Not Found
        </p>
        <ResponseSection code={`{
  "error": "webhook not found"
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/webhooks/WEBHOOK_ID`,
            python: `import requests

webhook_id = "wh-a1b2c3d4-..."
response = requests.get(f"${BASE_URL}/webhooks/{webhook_id}")
webhook = response.json()["data"]

print(f"URL: {webhook['url']}")
print(f"Events: {webhook['events']}")
print(f"Active: {webhook['is_active']}")`,
            javascript: `const webhookId = "wh-a1b2c3d4-...";
const response = await fetch(\`${BASE_URL}/webhooks/\${webhookId}\`);
const { data: webhook } = await response.json();

console.log(\`URL: \${webhook.url}\`);
console.log(\`Events: \${webhook.events.join(", ")}\`);
console.log(\`Active: \${webhook.is_active}\`);`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/webhooks/:id"
        title="Get Webhook"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
