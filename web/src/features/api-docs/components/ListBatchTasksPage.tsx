import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import type { BatchTaskStatus } from '@/types'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
  TryItDialog,
  TryItField,
} from './shared'

export function ListBatchTasksPage() {
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
    {
      name: 'status',
      label: 'status',
      type: 'select',
      required: false,
      placeholder: 'Filter by status',
      description: 'Filter tasks by status',
      options: [
        { value: 'pending', label: 'Pending' },
        { value: 'running', label: 'Running' },
        { value: 'completed', label: 'Completed' },
        { value: 'failed', label: 'Failed' },
      ],
    },
    {
      name: 'limit',
      label: 'limit',
      type: 'text',
      required: false,
      placeholder: '50',
      description: 'Maximum number of results',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.listBatchTasks(values.id, {
      status: (values.status || undefined) as BatchTaskStatus | undefined,
      limit: values.limit ? parseInt(values.limit) : undefined,
    })
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/batches/:id/tasks"
        title="List Batch Tasks"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Returns a paginated list of tasks within a batch. Filter by status to see pending, running, completed, or failed tasks.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      {/* Query Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Query Parameters</h3>
        <ParamTable params={[
          { name: 'status', type: 'string', required: false, desc: 'Filter by task status (pending, running, completed, failed)', example: '"failed"' },
          { name: 'limit', type: 'number', required: false, desc: 'Maximum results to return (default: 50)', example: '100' },
          { name: 'offset', type: 'number', required: false, desc: 'Pagination offset', example: '0' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "tasks": [
    {
      "id": "batch-abc123-0",
      "batch_id": "batch-abc123",
      "index": 0,
      "status": "completed",
      "input": {"data": "alert data 1"},
      "prompt": "Analyze: alert data 1",
      "result": "This appears to be a false positive...",
      "worker_id": "worker-0",
      "attempts": 1,
      "duration_ms": 1523,
      "created_at": "2024-01-15T10:30:00Z",
      "started_at": "2024-01-15T10:31:00Z"
    },
    {
      "id": "batch-abc123-1",
      "batch_id": "batch-abc123",
      "index": 1,
      "status": "failed",
      "input": {"data": "alert data 2"},
      "error": "timeout after 300s",
      "worker_id": "worker-1",
      "attempts": 3,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# List all tasks
curl ${BASE_URL}/batches/batch-abc123/tasks

# Get failed tasks only
curl "${BASE_URL}/batches/batch-abc123/tasks?status=failed"

# Paginate results
curl "${BASE_URL}/batches/batch-abc123/tasks?limit=50&offset=100"`,
            python: `import requests

# List all tasks
response = requests.get("${BASE_URL}/batches/batch-abc123/tasks")
tasks = response.json()["data"]["tasks"]

# Get failed tasks for debugging
failed = requests.get(
    "${BASE_URL}/batches/batch-abc123/tasks",
    params={"status": "failed"}
)
for task in failed.json()["data"]["tasks"]:
    print(f"Task {task['index']} failed: {task['error']}")`,
            javascript: `// List all tasks
const response = await fetch("${BASE_URL}/batches/batch-abc123/tasks");
const { data } = await response.json();

// Get completed tasks
const completed = await fetch(
  "${BASE_URL}/batches/batch-abc123/tasks?status=completed"
);`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/batches/:id/tasks"
        title="List Batch Tasks"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
