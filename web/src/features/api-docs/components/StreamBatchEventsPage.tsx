import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
  ResponseSection,
} from './shared'

export function StreamBatchEventsPage() {
  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/batches/:id/events"
        title="Stream Batch Events"
      />

      <p className="text-muted-foreground">
        Subscribe to real-time Server-Sent Events (SSE) for batch progress updates.
        Events include progress updates, worker status changes, and task completions.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The batch identifier', example: '"batch-abc123"' },
        ]} />
      </div>

      {/* Event Types */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Event Types</h3>
        <div className="space-y-4">
          <div>
            <p className="text-sm font-medium mb-2">batch.started</p>
            <ResponseSection code={`{
  "type": "batch.started",
  "batch_id": "batch-abc123",
  "workers": 5,
  "timestamp": "2024-01-15T10:31:00Z"
}`} />
          </div>

          <div>
            <p className="text-sm font-medium mb-2">batch.progress</p>
            <ResponseSection code={`{
  "type": "batch.progress",
  "batch_id": "batch-abc123",
  "completed": 45,
  "failed": 2,
  "total": 100,
  "percent": 47.0,
  "eta": "5m 30s",
  "tasks_per_sec": 0.85
}`} />
          </div>

          <div>
            <p className="text-sm font-medium mb-2">task.completed</p>
            <ResponseSection code={`{
  "type": "task.completed",
  "batch_id": "batch-abc123",
  "task_id": "batch-abc123-42",
  "worker_id": "worker-0",
  "duration_ms": 1523
}`} />
          </div>

          <div>
            <p className="text-sm font-medium mb-2">task.failed</p>
            <ResponseSection code={`{
  "type": "task.failed",
  "batch_id": "batch-abc123",
  "task_id": "batch-abc123-43",
  "worker_id": "worker-1",
  "error": "timeout after 300s",
  "attempts": 3
}`} />
          </div>

          <div>
            <p className="text-sm font-medium mb-2">worker.error</p>
            <ResponseSection code={`{
  "type": "worker.error",
  "batch_id": "batch-abc123",
  "worker_id": "worker-2",
  "error": "container crashed",
  "restarting": true
}`} />
          </div>

          <div>
            <p className="text-sm font-medium mb-2">batch.completed</p>
            <ResponseSection code={`{
  "type": "batch.completed",
  "batch_id": "batch-abc123",
  "completed": 98,
  "failed": 2,
  "total": 100,
  "duration_seconds": 3600
}`} />
          </div>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# Stream events with curl
curl -N ${BASE_URL}/batches/batch-abc123/events

# Output:
# event: batch.started
# data: {"type":"batch.started","batch_id":"batch-abc123","workers":5}
#
# event: batch.progress
# data: {"type":"batch.progress","completed":10,"total":100,"percent":10.0}`,
            python: `import sseclient
import requests

# Subscribe to batch events
response = requests.get(
    "${BASE_URL}/batches/batch-abc123/events",
    stream=True
)
client = sseclient.SSEClient(response)

for event in client.events():
    data = json.loads(event.data)

    if data["type"] == "batch.progress":
        print(f"Progress: {data['completed']}/{data['total']} ({data['percent']:.1f}%)")
    elif data["type"] == "task.failed":
        print(f"Task {data['task_id']} failed: {data['error']}")
    elif data["type"] == "batch.completed":
        print(f"Batch complete! {data['completed']} succeeded, {data['failed']} failed")
        break`,
            javascript: `// Subscribe to batch events
const eventSource = new EventSource(
  "${BASE_URL}/batches/batch-abc123/events"
);

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);

  switch (data.type) {
    case "batch.progress":
      console.log(\`Progress: \${data.completed}/\${data.total}\`);
      updateProgressBar(data.percent);
      break;
    case "task.failed":
      console.error(\`Task \${data.task_id} failed: \${data.error}\`);
      break;
    case "batch.completed":
      console.log("Batch complete!");
      eventSource.close();
      break;
  }
};

eventSource.onerror = (error) => {
  console.error("SSE connection error:", error);
  eventSource.close();
};`,
          }}
        />
      </div>

      {/* Notes */}
      <div className="bg-muted/50 rounded-lg p-4 space-y-2">
        <h4 className="font-medium">Notes</h4>
        <ul className="text-sm text-muted-foreground list-disc pl-5 space-y-1">
          <li>The connection remains open until the batch completes, fails, or is cancelled</li>
          <li>Progress events are sent every 2 seconds during active processing</li>
          <li>Individual task events (task.completed, task.failed) can be high-frequency for large batches</li>
          <li>Reconnect with <code>Last-Event-ID</code> header to resume from a specific point</li>
        </ul>
      </div>
    </div>
  )
}
