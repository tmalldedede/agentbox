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

export function GetTaskPage() {
  const [tryItOpen, setTryItOpen] = useState(false)

  const fields: TryItField[] = [
    {
      name: 'id',
      label: 'id',
      type: 'text',
      required: true,
      placeholder: 'task-a1b2c3d4',
      description: 'The task ID returned from Create Task',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.getTask(values.id)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/tasks/:id"
        title="Get Task"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Get a task's current status, result, turns, and metadata by its ID.
        You can poll this endpoint, or subscribe to SSE at <code className="bg-muted px-1.5 py-0.5 rounded text-sm">/tasks/:id/events</code> for real-time updates.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The task ID returned from Create Task', example: '"task-a1b2c3d4"' },
        ]} />
      </div>

      {/* Response */}
      <ResponseSection code={`{
  "id": "task-a1b2c3d4",
  "agent_id": "code-helper-x3k9",
  "agent_name": "Code Helper",
  "agent_type": "claude-code",
  "status": "completed",
  "prompt": "Now add memoization",
  "turn_count": 2,
  "turns": [
    {
      "id": "turn-e5f6g7h8",
      "prompt": "Write a fibonacci function in Python",
      "result": { "text": "def fib(n): ...", "usage": { "duration_seconds": 8 } },
      "created_at": "2024-01-15T10:30:01Z"
    },
    {
      "id": "turn-i9j0k1l2",
      "prompt": "Now add memoization",
      "result": { "text": "def fib(n, memo={}): ...", "usage": { "duration_seconds": 5 } },
      "created_at": "2024-01-15T10:31:00Z"
    }
  ],
  "result": { "text": "def fib(n, memo={}): ...", "usage": { "duration_seconds": 5 } },
  "attachments": ["file-abc123"],
  "session_id": "sess-xyz789",
  "created_at": "2024-01-15T10:30:00Z",
  "started_at": "2024-01-15T10:30:01Z",
  "completed_at": "2024-01-15T10:31:05Z"
}`} />

      {/* Status Values */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Task Status</h3>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Status</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-t"><td className="px-4 py-2 font-mono">pending</td><td className="px-4 py-2 text-muted-foreground">Task created, not yet queued</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">queued</td><td className="px-4 py-2 text-muted-foreground">In queue, waiting for execution</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">running</td><td className="px-4 py-2 text-muted-foreground">Currently executing</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">completed</td><td className="px-4 py-2 text-muted-foreground">Finished successfully, result available</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">failed</td><td className="px-4 py-2 text-muted-foreground">Execution failed, check error field</td></tr>
              <tr className="border-t"><td className="px-4 py-2 font-mono">cancelled</td><td className="px-4 py-2 text-muted-foreground">Cancelled by user via DELETE</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl ${BASE_URL}/tasks/task-a1b2c3d4 \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests
import time

task_id = "task-a1b2c3d4"

# Poll until complete
while True:
    response = requests.get(
        f"${BASE_URL}/tasks/{task_id}",
        headers={"Authorization": "Bearer YOUR_API_KEY"},
    )
    task = response.json()

    if task["status"] in ("completed", "failed", "cancelled"):
        break

    time.sleep(2)  # Wait 2 seconds between polls

if task["status"] == "completed":
    print(task["result"]["text"])
else:
    print(f"Task {task['status']}: {task.get('error', 'unknown')}")`,
            javascript: `async function waitForTask(taskId) {
  while (true) {
    const res = await fetch(\`${BASE_URL}/tasks/\${taskId}\`, {
      headers: { "Authorization": "Bearer YOUR_API_KEY" },
    });
    const task = await res.json();

    if (["completed", "failed", "cancelled"].includes(task.status)) {
      return task;
    }

    await new Promise(r => setTimeout(r, 2000)); // Poll every 2s
  }
}

const task = await waitForTask("task-a1b2c3d4");
if (task.status === "completed") {
  console.log(task.result.text);
}`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/tasks/:id"
        title="Get Task"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
