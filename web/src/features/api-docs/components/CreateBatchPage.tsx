import { useState, useCallback } from 'react'
import { useAgents } from '@/hooks'
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

export function CreateBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)
  const { data: agents = [] } = useAgents()

  const fields: TryItField[] = [
    {
      name: 'name',
      label: 'name',
      type: 'text',
      required: true,
      placeholder: 'Alert Analysis Batch',
      description: 'Human-readable name for this batch',
    },
    {
      name: 'agent_id',
      label: 'agent_id',
      type: 'select',
      required: true,
      placeholder: 'Select an agent',
      description: 'The agent to execute batch tasks',
      options: agents.map((a) => ({ value: a.id, label: `${a.name} (${a.id})` })),
    },
    {
      name: 'prompt_template',
      label: 'prompt_template',
      type: 'textarea',
      required: true,
      placeholder: 'Analyze this alert: {{.alert_data}}',
      description: 'Go template with {{.field}} placeholders',
    },
    {
      name: 'concurrency',
      label: 'concurrency',
      type: 'text',
      required: false,
      placeholder: '5',
      description: 'Number of parallel workers (default: 5)',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.createBatch({
      name: values.name,
      agent_id: values.agent_id,
      prompt_template: values.prompt_template,
      inputs: [{ example: 'data' }],
      concurrency: parseInt(values.concurrency) || 5,
      timeout: 300,
      max_retries: 2,
    })
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/batches"
        title="Create Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Creates a new batch for processing multiple tasks efficiently using a worker pool.
        Tasks share the same prompt template and are executed in parallel.
      </p>

      {/* Body Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <ParamTable params={[
          { name: 'name', type: 'string', required: true, desc: 'Human-readable name for the batch', example: '"Alert Analysis Batch"' },
          { name: 'agent_id', type: 'string', required: true, desc: 'The agent identifier to execute tasks', example: '"analyzer-agent"' },
          { name: 'prompt_template', type: 'string', required: true, desc: 'Go template with {{.field}} placeholders for variable substitution', example: '"Analyze: {{.data}}"' },
          { name: 'inputs', type: 'object[]', required: true, desc: 'Array of input objects, each becomes one task', example: '[{"data": "alert1"}, {"data": "alert2"}]' },
          { name: 'concurrency', type: 'number', required: false, desc: 'Number of parallel workers (default: 5)', example: '10' },
          { name: 'timeout', type: 'number', required: false, desc: 'Per-task timeout in seconds (default: 300)', example: '600' },
          { name: 'max_retries', type: 'number', required: false, desc: 'Max retry attempts for failed tasks (default: 2)', example: '3' },
          { name: 'auto_start', type: 'boolean', required: false, desc: 'Start batch immediately after creation (default: false)', example: 'true' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "batch-abc123",
  "name": "Alert Analysis Batch",
  "agent_id": "analyzer-agent",
  "status": "pending",
  "total_tasks": 100,
  "completed": 0,
  "failed": 0,
  "concurrency": 5,
  "template": {
    "prompt_template": "Analyze: {{.data}}",
    "timeout": 300,
    "max_retries": 2
  },
  "created_at": "2024-01-15T10:30:00Z"
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/batches \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Alert Analysis Batch",
    "agent_id": "analyzer-agent",
    "prompt_template": "Analyze this security alert: {{.alert_data}}",
    "inputs": [
      {"alert_data": "Suspicious login from IP 1.2.3.4"},
      {"alert_data": "Port scan detected on port 22"}
    ],
    "concurrency": 5,
    "auto_start": true
  }'`,
            python: `import requests

response = requests.post(
    "${BASE_URL}/batches",
    json={
        "name": "Alert Analysis Batch",
        "agent_id": "analyzer-agent",
        "prompt_template": "Analyze this security alert: {{.alert_data}}",
        "inputs": [
            {"alert_data": "Suspicious login from IP 1.2.3.4"},
            {"alert_data": "Port scan detected on port 22"},
        ],
        "concurrency": 5,
        "auto_start": True,
    },
)

batch = response.json()["data"]
print(f"Batch ID: {batch['id']}, Tasks: {batch['total_tasks']}")`,
            javascript: `const response = await fetch("${BASE_URL}/batches", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    name: "Alert Analysis Batch",
    agent_id: "analyzer-agent",
    prompt_template: "Analyze this security alert: {{.alert_data}}",
    inputs: [
      { alert_data: "Suspicious login from IP 1.2.3.4" },
      { alert_data: "Port scan detected on port 22" },
    ],
    concurrency: 5,
    auto_start: true,
  }),
});

const { data: batch } = await response.json();
console.log(\`Batch ID: \${batch.id}, Tasks: \${batch.total_tasks}\`);`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/batches"
        title="Create Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
