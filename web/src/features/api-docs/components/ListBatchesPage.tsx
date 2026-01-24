import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import type { BatchStatus } from '@/types'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
  TryItField,
} from './shared'

export function ListBatchesPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'status',
      label: 'status',
      type: 'select',
      required: false,
      placeholder: 'Filter by status',
      description: 'Filter batches by status',
      options: [
        { value: 'pending', label: 'Pending' },
        { value: 'running', label: 'Running' },
        { value: 'paused', label: 'Paused' },
        { value: 'completed', label: 'Completed' },
        { value: 'failed', label: 'Failed' },
        { value: 'cancelled', label: 'Cancelled' },
      ],
    },
    {
      name: 'limit',
      label: 'limit',
      type: 'text',
      required: false,
      placeholder: '20',
      description: 'Maximum number of results',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.listBatches({
      status: (values.status || undefined) as BatchStatus | undefined,
      limit: values.limit ? parseInt(values.limit) : undefined,
    })
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/batches"
        title="List Batches"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Returns a list of all batches with optional filtering by status.
      </p>

      {/* Query Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Query Parameters</h3>
        <ParamTable params={[
          { name: 'status', type: 'string', required: false, desc: 'Filter by batch status', example: '"running"' },
          { name: 'limit', type: 'number', required: false, desc: 'Maximum results to return (default: 20)', example: '50' },
          { name: 'offset', type: 'number', required: false, desc: 'Pagination offset', example: '0' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "batches": [
    {
      "id": "batch-abc123",
      "name": "Alert Analysis Batch",
      "status": "running",
      "total_tasks": 100,
      "completed": 45,
      "failed": 2,
      "concurrency": 5,
      "created_at": "2024-01-15T10:30:00Z",
      "started_at": "2024-01-15T10:31:00Z"
    }
  ],
  "total": 1
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# List all batches
curl ${BASE_URL}/batches

# Filter by status
curl "${BASE_URL}/batches?status=running&limit=10"`,
            python: `import requests

# List all batches
response = requests.get("${BASE_URL}/batches")
batches = response.json()["data"]["batches"]

# Filter running batches
response = requests.get(
    "${BASE_URL}/batches",
    params={"status": "running", "limit": 10}
)`,
            javascript: `// List all batches
const response = await fetch("${BASE_URL}/batches");
const { data } = await response.json();
console.log(\`Found \${data.total} batches\`);

// Filter by status
const filtered = await fetch("${BASE_URL}/batches?status=running");`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/batches"
        title="List Batches"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
