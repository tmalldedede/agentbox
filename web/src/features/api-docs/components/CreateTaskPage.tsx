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

export function CreateTaskPage() {
  const [tryItOpen, setTryItOpen] = useState(false)
  const { data: agents = [] } = useAgents()

  const fields: TryItField[] = [
    {
      name: 'prompt',
      label: 'prompt',
      type: 'textarea',
      required: true,
      placeholder: 'Write a fibonacci function in Python',
      description: 'The task prompt or instruction for the agent',
    },
    {
      name: 'agent_id',
      label: 'agent_id',
      type: 'select',
      required: false,
      placeholder: 'Select an agent (for new task)',
      description: 'Required for new task. Leave empty when appending a turn.',
      options: agents.map((a) => ({ value: a.id, label: `${a.name} (${a.id})` })),
    },
    {
      name: 'task_id',
      label: 'task_id',
      type: 'text',
      required: false,
      placeholder: 'task-xxxxxxxx (for multi-turn)',
      description: 'Existing task ID to append a new turn. Mutually exclusive with agent_id.',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const response = await api.createTask({
      agent_id: values.agent_id || undefined,
      task_id: values.task_id || undefined,
      prompt: values.prompt,
    })
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="POST"
        path="/tasks"
        title="Create Task"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        Creates a new task or appends a turn to an existing task (multi-turn conversation).
      </p>
      <ul className="text-sm text-muted-foreground list-disc pl-5 space-y-1">
        <li><strong>New task</strong>: provide <code>agent_id</code> + <code>prompt</code>. Task is queued and executed asynchronously.</li>
        <li><strong>Append turn</strong>: provide <code>task_id</code> + <code>prompt</code>. The HTTP response returns immediately; the turn executes in the background. Subscribe to SSE events at <code>/tasks/:id/events</code> to receive real-time updates.</li>
      </ul>

      {/* Body Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Body</h3>
        <ParamTable params={[
          { name: 'prompt', type: 'string', required: true, desc: 'The task prompt or instruction for the agent', example: '"Write a fibonacci function"' },
          { name: 'agent_id', type: 'string', required: false, desc: 'Required for new task. The agent identifier to execute this task', example: '"code-helper-x3k9"' },
          { name: 'task_id', type: 'string', required: false, desc: 'For multi-turn: existing task ID to append a new turn. Mutually exclusive with agent_id', example: '"task-a1b2c3d4"' },
          { name: 'attachments', type: 'string[]', required: false, desc: 'List of uploaded file IDs to attach to the task workspace', example: '["file-abc123"]' },
          { name: 'webhook_url', type: 'string', required: false, desc: 'URL to receive POST notification on task completion' },
          { name: 'timeout', type: 'number', required: false, desc: 'Task timeout in seconds (default: 1800)' },
          { name: 'metadata', type: 'object', required: false, desc: 'Key-value pairs attached to the task for your reference' },
        ]} />
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <p className="text-sm text-muted-foreground">New task (status: queued):</p>
        <ResponseSection code={`{
  "id": "task-a1b2c3d4",
  "agent_id": "code-helper-x3k9",
  "agent_name": "Code Helper",
  "status": "queued",
  "prompt": "Write a fibonacci function",
  "turn_count": 1,
  "turns": [{ "id": "turn-e5f6g7h8", "prompt": "Write a fibonacci function", "created_at": "..." }],
  "created_at": "2024-01-15T10:30:00Z"
}`} />
        <p className="text-sm text-muted-foreground">Append turn (status: running, returns immediately):</p>
        <ResponseSection code={`{
  "id": "task-a1b2c3d4",
  "status": "running",
  "prompt": "Now add memoization",
  "turn_count": 2,
  "turns": [
    { "id": "turn-e5f6g7h8", "prompt": "Write a fibonacci function", "result": {...}, "created_at": "..." },
    { "id": "turn-i9j0k1l2", "prompt": "Now add memoization", "created_at": "..." }
  ]
}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <p className="text-sm text-muted-foreground font-medium">Create new task:</p>
        <ExamplesPanel
          examples={{
            curl: `curl -X POST ${BASE_URL}/tasks \\
  -H "Content-Type: application/json" \\
  -d '{
    "prompt": "Write a fibonacci function in Python",
    "agent_id": "YOUR_AGENT_ID"
  }'`,
            python: `import requests

response = requests.post(
    "${BASE_URL}/tasks",
    json={
        "prompt": "Write a fibonacci function in Python",
        "agent_id": "YOUR_AGENT_ID",
    },
)

task = response.json()["data"]
print(f"Task ID: {task['id']}, Status: {task['status']}")`,
            javascript: `const response = await fetch("${BASE_URL}/tasks", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    prompt: "Write a fibonacci function in Python",
    agent_id: "YOUR_AGENT_ID",
  }),
});

const { data: task } = await response.json();
console.log(\`Task ID: \${task.id}, Status: \${task.status}\`);`,
          }}
        />
        <p className="text-sm text-muted-foreground font-medium mt-4">Append turn (multi-turn):</p>
        <ExamplesPanel
          examples={{
            curl: `# Append a follow-up turn (returns immediately, executes async)
curl -X POST ${BASE_URL}/tasks \\
  -H "Content-Type: application/json" \\
  -d '{
    "task_id": "task-a1b2c3d4",
    "prompt": "Now add memoization to the function"
  }'`,
            python: `# Append turn + listen SSE events
response = requests.post(
    "${BASE_URL}/tasks",
    json={
        "task_id": "task-a1b2c3d4",
        "prompt": "Now add memoization",
    },
)
# Response returns immediately; turn executes in background

# Subscribe to SSE for real-time updates
import sseclient
events = sseclient.SSEClient(f"${BASE_URL}/tasks/task-a1b2c3d4/events")
for event in events:
    print(event.data)`,
            javascript: `// Append turn
const res = await fetch("${BASE_URL}/tasks", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    task_id: "task-a1b2c3d4",
    prompt: "Now add memoization",
  }),
});
// Response returns immediately

// Subscribe to SSE events
const es = new EventSource("${BASE_URL}/tasks/task-a1b2c3d4/events");
es.onmessage = (e) => console.log(JSON.parse(e.data));`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="POST"
        path="/tasks"
        title="Create Task"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
