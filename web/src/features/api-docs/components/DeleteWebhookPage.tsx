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

export function DeleteWebhookPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'webhook_id',
      type: 'text',
      required: true,
      placeholder: 'wh-a1b2c3d4-...',
      description: 'The unique identifier of the webhook to delete',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.deleteWebhook(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="DELETE"
        path="/webhooks/:id"
        title="Delete Webhook"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Permanently deletes a webhook. The webhook will immediately stop receiving event notifications.
        This action cannot be undone.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The unique identifier of the webhook to delete', example: '"wh-a1b2c3d4-..."' },
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
  "deleted": true
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
            curl: `curl -X DELETE ${BASE_URL}/webhooks/WEBHOOK_ID`,
            python: `import requests

webhook_id = "wh-a1b2c3d4-..."
response = requests.delete(f"${BASE_URL}/webhooks/{webhook_id}")

if response.status_code == 200:
    print("Webhook deleted successfully")
else:
    print(f"Error: {response.json()}")`,
            javascript: `const webhookId = "wh-a1b2c3d4-...";
const response = await fetch(\`${BASE_URL}/webhooks/\${webhookId}\`, {
  method: "DELETE",
});

const { data } = await response.json();
if (data.deleted) {
  console.log("Webhook deleted successfully");
}`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="DELETE"
        path="/webhooks/:id"
        title="Delete Webhook"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
