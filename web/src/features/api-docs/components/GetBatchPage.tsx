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

export function GetBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'batch-abc123',
      description: 'The batch ID',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.getBatch(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/batches/:id"
        title="Get Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Retrieves detailed information about a specific batch, including progress, worker status, and configuration.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "batch-abc123",
  "name": "Alert Analysis Batch",
  "agent_id": "analyzer-agent",
  "status": "running",
  "total_tasks": 100,
  "completed": 45,
  "failed": 2,
  "concurrency": 5,
  "template": {
    "prompt_template": "Analyze: {{.data}}",
    "timeout": 300,
    "max_retries": 2
  },
  "workers": [
    {"id": "worker-0", "status": "busy", "completed": 10, "current_task": "batch-abc123-42"},
    {"id": "worker-1", "status": "idle", "completed": 9},
    {"id": "worker-2", "status": "busy", "completed": 8, "current_task": "batch-abc123-43"}
  ],
  "error_summary": {"timeout": 1, "api_error": 1},
  "created_at": "2024-01-15T10:30:00Z",
  "started_at": "2024-01-15T10:31:00Z"
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/batches/batch-abc123`,
            python: `import requests

response = requests.get("${BASE_URL}/batches/batch-abc123")
batch = response.json()["data"]

print(f"Status: {batch['status']}")
print(f"Progress: {batch['completed']}/{batch['total_tasks']}")
print(f"Workers: {len(batch.get('workers', []))}")`,
            javascript: `const response = await fetch("${BASE_URL}/batches/batch-abc123");
const { data: batch } = await response.json();

console.log(\`Status: \${batch.status}\`);
console.log(\`Progress: \${batch.completed}/\${batch.total_tasks}\`);`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/batches/:id"
        title="Get Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
