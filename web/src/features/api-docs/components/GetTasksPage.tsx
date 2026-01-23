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

export function GetTasksPage() {
  const [tryItOpen, setTryItOpen] = useState(false)
  const { data: agents = [] } = useAgents()

  const fields: TryItField[] = [
    {
      name: 'agent_id',
      label: 'agent_id',
      type: 'select',
      required: false,
      placeholder: 'All agents (optional)',
      description: 'Filter by agent ID',
      options: agents.map((a) => ({ value: a.id, label: `${a.name} (${a.id})` })),
    },
    {
      name: 'status',
      label: 'status',
      type: 'select',
      required: false,
      placeholder: 'All statuses (optional)',
      description: 'Filter by task status',
      options: [
        { value: 'queued', label: 'queued' },
        { value: 'running', label: 'running' },
        { value: 'completed', label: 'completed' },
        { value: 'failed', label: 'failed' },
        { value: 'cancelled', label: 'cancelled' },
      ],
    },
    {
      name: 'limit',
      label: 'limit',
      type: 'text',
      required: false,
      placeholder: '20',
      description: 'Number of results per page (default: 20, max: 100)',
    },
  ]

  const handleExecute = useCallback(async (values: Record<string, string>) => {
    const options: { status?: string; agent_id?: string; limit?: number } = {}
    if (values.agent_id) options.agent_id = values.agent_id
    if (values.status) options.status = values.status
    if (values.limit) options.limit = parseInt(values.limit, 10)
    const response = await api.listTasks(options)
    return { status: 200, data: response }
  }, [])

  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/tasks"
        title="Get Tasks"
        onTryIt={() => setTryItOpen(true)}
      />

      <p className="text-muted-foreground">
        List tasks with optional filters. Returns a paginated list sorted by creation time (newest first).
      </p>

      {/* Query Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Query Parameters</h3>
        <ParamTable params={[
          { name: 'agent_id', type: 'string', required: false, desc: 'Filter by agent ID' },
          { name: 'status', type: 'string', required: false, desc: 'Filter by status', example: 'running | completed | failed | cancelled' },
          { name: 'limit', type: 'number', required: false, desc: 'Number of results per page (default: 20, max: 100)' },
          { name: 'offset', type: 'number', required: false, desc: 'Offset for pagination' },
        ]} />
      </div>

      {/* Response */}
      <ResponseSection code={`{
  "tasks": [
    {
      "id": "task-a1b2c3d4",
      "agent_id": "code-helper-x3k9",
      "agent_name": "Code Helper",
      "status": "completed",
      "prompt": "Write a fibonacci function",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:45Z"
    }
  ],
  "total": 42
}`} />

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `curl "${BASE_URL}/tasks?agent_id=code-helper-x3k9&status=completed&limit=10" \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

response = requests.get(
    "${BASE_URL}/tasks",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
    params={
        "agent_id": "code-helper-x3k9",
        "status": "completed",
        "limit": 10,
    },
)

data = response.json()
print(f"Total: {data['total']}")
for task in data["tasks"]:
    print(f"  {task['id']}: {task['status']}")`,
            javascript: `const params = new URLSearchParams({
  agent_id: "code-helper-x3k9",
  status: "completed",
  limit: "10",
});

const response = await fetch(\`${BASE_URL}/tasks?\${params}\`, {
  headers: { "Authorization": "Bearer YOUR_API_KEY" },
});

const { tasks, total } = await response.json();
console.log(\`Total: \${total}\`);
tasks.forEach(t => console.log(\`  \${t.id}: \${t.status}\`));`,
          }}
        />
      </div>

      {/* Try It Dialog */}
      <TryItDialog
        open={tryItOpen}
        onOpenChange={setTryItOpen}
        method="GET"
        path="/tasks"
        title="Get Tasks"
        fields={fields}
        onExecute={handleExecute}
      />
    </div>
  )
}
