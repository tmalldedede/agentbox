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

// Start Batch
export function StartBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'batch-abc123',
      description: 'The batch ID to start',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.startBatch(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/batches/:id/start"
        title="Start Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Starts a pending or paused batch. Creates worker containers and begins processing tasks from the queue.
      </p>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "batch-abc123",
  "status": "running",
  "workers": [
    {"id": "worker-0", "status": "idle", "completed": 0},
    {"id": "worker-1", "status": "idle", "completed": 0}
  ],
  "started_at": "2024-01-15T10:31:00Z"
}`} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/batches/batch-abc123/start`,
            python: `response = requests.post("${BASE_URL}/batches/batch-abc123/start")
batch = response.json()["data"]
print(f"Batch started with {len(batch['workers'])} workers")`,
            javascript: `const response = await fetch("${BASE_URL}/batches/batch-abc123/start", {
  method: "POST"
});
const { data: batch } = await response.json();`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/batches/:id/start"
        title="Start Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}

// Pause Batch
export function PauseBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'batch-abc123',
      description: 'The batch ID to pause',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.pauseBatch(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/batches/:id/pause"
        title="Pause Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Pauses a running batch. Workers finish their current task then stop. The batch can be resumed later.
      </p>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "batch-abc123",
  "status": "paused",
  "completed": 45,
  "total_tasks": 100
}`} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/batches/batch-abc123/pause`,
            python: `response = requests.post("${BASE_URL}/batches/batch-abc123/pause")`,
            javascript: `await fetch("${BASE_URL}/batches/batch-abc123/pause", { method: "POST" });`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/batches/:id/pause"
        title="Pause Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}

// Cancel Batch
export function CancelBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'batch-abc123',
      description: 'The batch ID to cancel',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.cancelBatch(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/batches/:id/cancel"
        title="Cancel Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Cancels a running or paused batch. All workers are stopped and destroyed. Pending tasks remain unprocessed.
      </p>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <ResponseSection code={`{
  "id": "batch-abc123",
  "status": "cancelled",
  "completed": 45,
  "failed": 2,
  "total_tasks": 100,
  "completed_at": "2024-01-15T11:00:00Z"
}`} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/batches/batch-abc123/cancel`,
            python: `response = requests.post("${BASE_URL}/batches/batch-abc123/cancel")`,
            javascript: `await fetch("${BASE_URL}/batches/batch-abc123/cancel", { method: "POST" });`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/batches/:id/cancel"
        title="Cancel Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}

// Delete Batch
export function DeleteBatchPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'batch-abc123',
      description: 'The batch ID to delete',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    await api.deleteBatch(values.id)
    return { status: 204, data: null }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="DELETE"
        path="/batches/:id"
        title="Delete Batch"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Permanently deletes a batch and all its tasks. The batch must be in a terminal state (completed, failed, or cancelled).
      </p>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <p className="text-sm text-muted-foreground">Returns 204 No Content on success.</p>
      </div>

      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X DELETE ${BASE_URL}/batches/batch-abc123`,
            python: `response = requests.delete("${BASE_URL}/batches/batch-abc123")
if response.status_code == 204:
    print("Batch deleted successfully")`,
            javascript: `const response = await fetch("${BASE_URL}/batches/batch-abc123", {
  method: "DELETE"
});
if (response.status === 204) {
  console.log("Batch deleted");
}`,
          }}
        />
      </div>

      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="DELETE"
        path="/batches/:id"
        title="Delete Batch"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
