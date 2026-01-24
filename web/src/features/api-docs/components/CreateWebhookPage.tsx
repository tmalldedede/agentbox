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

export function CreateWebhookPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'url',
      label: 'url',
      type: 'text',
      required: true,
      placeholder: 'https://example.com/webhook',
      description: 'The endpoint URL to receive webhook payloads',
    },
    {
      name: 'secret',
      label: 'secret',
      type: 'text',
      required: false,
      placeholder: 'your-secret-key',
      description: 'Secret for HMAC-SHA256 signature verification',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.createWebhook({
      url: values.url,
      secret: values.secret || undefined,
    })
    return { status: 201, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/webhooks"
        title="Create Webhook"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Creates a new webhook endpoint to receive real-time notifications for task lifecycle events.
        The webhook will start receiving events immediately after creation.
      </p>

      {/* Body Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <ParamTable params={[
          { name: 'url', type: 'string', required: true, desc: 'The endpoint URL where webhook payloads will be delivered via HTTP POST', example: '"https://example.com/webhook"' },
          { name: 'secret', type: 'string', required: false, desc: 'Secret key used to generate HMAC-SHA256 signature in X-Webhook-Signature header', example: '"whsec_abc123..."' },
          { name: 'events', type: 'string[]', required: false, desc: 'List of event types to subscribe to. If empty, subscribes to all events', example: '["task.created", "task.completed"]' },
        ]} />
      </div>

      {/* Events Reference */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Available Events</h3>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Event</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              {[
                { event: 'task.created', desc: 'Fired when a new task is created and queued' },
                { event: 'task.completed', desc: 'Fired when a task finishes execution successfully' },
                { event: 'task.failed', desc: 'Fired when a task encounters an error during execution' },
                { event: 'session.started', desc: 'Fired when an agent session starts' },
                { event: 'session.stopped', desc: 'Fired when an agent session ends' },
              ].map((row) => (
                <tr key={row.event} className="border-t">
                  <td className="px-4 py-2.5">
                    <code className="font-mono text-sm">{row.event}</code>
                  </td>
                  <td className="px-4 py-2.5 text-muted-foreground">{row.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Webhook Payload */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Webhook Payload</h3>
        <p className="text-sm text-muted-foreground">
          When an event occurs, a POST request is sent to your endpoint with the following JSON body:
        </p>
        <ResponseSection code={`{
  "id": "evt-a1b2c3d4",
  "event": "task.completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "task_id": "task-x7y8z9",
    "status": "completed",
    "result": "..."
  }
}`} />
        <div className="space-y-2 mt-3">
          <h4 className="font-medium text-sm">Headers</h4>
          <ParamTable params={[
            { name: 'Content-Type', type: 'string', required: true, desc: 'Always application/json' },
            { name: 'X-Webhook-ID', type: 'string', required: true, desc: 'The webhook configuration ID' },
            { name: 'X-Webhook-Signature', type: 'string', required: false, desc: 'HMAC-SHA256 hex digest of the payload body (only if secret is configured)' },
          ]} />
        </div>
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <p className="text-sm text-muted-foreground">
          <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">201</span>
          Created
        </p>
        <ResponseSection code={`{
  "id": "wh-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "url": "https://example.com/webhook",
  "events": ["task.created", "task.completed"],
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/webhooks \\
  -H "Content-Type: application/json" \\
  -d '{
    "url": "https://example.com/webhook",
    "secret": "whsec_your_secret_key",
    "events": ["task.created", "task.completed", "task.failed"]
  }'`,
            python: `import requests

response = requests.post(
    "${BASE_URL}/webhooks",
    json={
        "url": "https://example.com/webhook",
        "secret": "whsec_your_secret_key",
        "events": ["task.created", "task.completed", "task.failed"],
    },
)

webhook = response.json()["data"]
print(f"Webhook ID: {webhook['id']}")`,
            javascript: `const response = await fetch("${BASE_URL}/webhooks", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://example.com/webhook",
    secret: "whsec_your_secret_key",
    events: ["task.created", "task.completed", "task.failed"],
  }),
});

const { data: webhook } = await response.json();
console.log(\`Webhook ID: \${webhook.id}\`);`,
          }}
        />
      </div>

      {/* Signature Verification */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Signature Verification</h3>
        <p className="text-sm text-muted-foreground">
          If you configured a secret, verify incoming webhooks by computing the HMAC-SHA256 signature of the raw request body:
        </p>
        <ExamplesPanel
          examples={{
            curl: `# Verify signature using openssl
BODY='{"id":"evt-xxx","event":"task.completed",...}'
SECRET="whsec_your_secret_key"
EXPECTED=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | cut -d' ' -f2)
# Compare $EXPECTED with X-Webhook-Signature header`,
            python: `import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)

# In your webhook handler:
# signature = request.headers["X-Webhook-Signature"]
# is_valid = verify_webhook(request.body, signature, "whsec_...")`,
            javascript: `import crypto from "crypto";

function verifyWebhook(payload: string, signature: string, secret: string): boolean {
  const expected = crypto
    .createHmac("sha256", secret)
    .update(payload)
    .digest("hex");
  return crypto.timingSafeEqual(
    Buffer.from(expected),
    Buffer.from(signature)
  );
}

// In your webhook handler:
// const sig = req.headers["x-webhook-signature"];
// const valid = verifyWebhook(req.rawBody, sig, "whsec_...");`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/webhooks"
        title="Create Webhook"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
