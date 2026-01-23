import { useState, useCallback } from 'react'
import { api } from '@/services/api'
import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  TryItDialog,
  TryItField,
} from './shared'

export function CancelTaskPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'task-a1b2c3d4',
      description: 'The task ID to cancel or delete',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.cancelTask(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="DELETE"
        path="/tasks/:id"
        title="Cancel Task"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Cancel a running or queued task, or delete a completed/failed task.
        If the task is currently running, the agent process will be terminated.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The task ID to cancel or delete', example: '"task-a1b2c3d4"' },
        ]} />
      </div>

      {/* Behavior */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Behavior</h3>
        <div className="rounded-lg border p-4 space-y-3 text-sm">
          <div className="flex items-start gap-3">
            <code className="shrink-0 bg-muted px-2 py-0.5 rounded font-mono">queued</code>
            <p className="text-muted-foreground">Task is removed from queue. Status changes to <code className="bg-muted px-1 rounded">cancelled</code>.</p>
          </div>
          <div className="flex items-start gap-3">
            <code className="shrink-0 bg-muted px-2 py-0.5 rounded font-mono">running</code>
            <p className="text-muted-foreground">Agent process is terminated, container stopped. Status changes to <code className="bg-muted px-1 rounded">cancelled</code>.</p>
          </div>
          <div className="flex items-start gap-3">
            <code className="shrink-0 bg-muted px-2 py-0.5 rounded font-mono">completed/failed</code>
            <p className="text-muted-foreground">Task record is permanently deleted.</p>
          </div>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl -X DELETE ${BASE_URL}/tasks/task-a1b2c3d4 \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

response = requests.delete(
    f"${BASE_URL}/tasks/task-a1b2c3d4",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
)

if response.status_code == 200:
    print("Task cancelled/deleted successfully")`,
            javascript: `const response = await fetch(
  \`${BASE_URL}/tasks/task-a1b2c3d4\`,
  {
    method: "DELETE",
    headers: { "Authorization": "Bearer YOUR_API_KEY" },
  }
);

if (response.ok) {
  console.log("Task cancelled/deleted successfully");
}`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="DELETE"
        path="/tasks/:id"
        title="Cancel Task"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
