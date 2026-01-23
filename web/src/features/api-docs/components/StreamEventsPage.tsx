import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
} from './shared'

export function StreamEventsPage() {
  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/tasks/:id/events"
        title="Stream Events (SSE)"
      />

      <p className="text-muted-foreground">
        Subscribe to real-time Server-Sent Events (SSE) for a task.
        Events are pushed as the agent executes, providing live feedback on thinking, tool calls, and messages.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The task ID to subscribe to', example: '"task-a1b2c3d4"' },
        ]} />
      </div>

      {/* Event Types */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Event Types</h3>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Type</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
                <th className="text-left px-4 py-2 font-medium">Data</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">task.started</td>
                <td className="px-4 py-2.5 text-muted-foreground">Task execution began</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ task_id, agent_id }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">task.turn_started</td>
                <td className="px-4 py-2.5 text-muted-foreground">A new turn started executing</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ turn_id, prompt }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">agent.thinking</td>
                <td className="px-4 py-2.5 text-muted-foreground">Agent is processing</td>
                <td className="px-4 py-2.5 text-muted-foreground">-</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">agent.tool_call</td>
                <td className="px-4 py-2.5 text-muted-foreground">Agent invoked a tool</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ tool, args }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">agent.message</td>
                <td className="px-4 py-2.5 text-muted-foreground">Agent produced output</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ text, turn_id }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">task.completed</td>
                <td className="px-4 py-2.5 text-muted-foreground">Task finished (idle timeout or all turns done)</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ reason }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">task.failed</td>
                <td className="px-4 py-2.5 text-muted-foreground">Task execution failed</td>
                <td className="px-4 py-2.5 font-mono text-xs">{"{ error }"}</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono text-xs">task.cancelled</td>
                <td className="px-4 py-2.5 text-muted-foreground">Task was cancelled by user</td>
                <td className="px-4 py-2.5 text-muted-foreground">-</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Response Format */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Event Format</h3>
        <p className="text-sm text-muted-foreground">
          Each event is a JSON object sent as an SSE <code className="bg-muted px-1.5 py-0.5 rounded">data:</code> line:
        </p>
        <ResponseSection code={`data: {"type":"task.turn_started","data":{"turn_id":"turn-e5f6g7h8","prompt":"Write a function"}}

data: {"type":"agent.thinking"}

data: {"type":"agent.message","data":{"text":"Here is the function...","turn_id":"turn-e5f6g7h8"}}

data: {"type":"task.completed","data":{"reason":"idle timeout"}}`} />
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# Stream events (curl stays connected)
curl -N ${BASE_URL}/tasks/task-a1b2c3d4/events`,
            python: `import sseclient
import requests

response = requests.get(
    "${BASE_URL}/tasks/task-a1b2c3d4/events",
    stream=True,
)

client = sseclient.SSEClient(response)
for event in client.events():
    data = json.loads(event.data)
    print(f"[{data['type']}] {data.get('data', '')}")

    if data["type"] in ("task.completed", "task.failed"):
        break`,
            javascript: `const es = new EventSource("${BASE_URL}/tasks/task-a1b2c3d4/events");

es.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(\`[\${data.type}]\`, data.data);

  if (data.type === "task.completed" || data.type === "task.failed") {
    es.close();
  }
};

es.onerror = () => {
  console.error("SSE connection error");
  es.close();
};`,
          }}
        />
      </div>

      {/* Notes */}
      <div className="rounded-lg border p-4 bg-muted/30 space-y-2">
        <p className="text-sm font-medium">Notes</p>
        <ul className="text-sm text-muted-foreground list-disc pl-5 space-y-1">
          <li>The connection stays open until the task reaches a terminal state or the client disconnects.</li>
          <li>If the task is already completed when you connect, you will immediately receive the final event.</li>
          <li>Events are buffered (up to 100) — fast-producing agents won't drop messages under normal conditions.</li>
          <li>Use this endpoint alongside multi-turn <code className="bg-muted px-1 rounded">POST /tasks</code> to get real-time feedback on each turn's execution.</li>
        </ul>
      </div>
    </div>
  )
}
