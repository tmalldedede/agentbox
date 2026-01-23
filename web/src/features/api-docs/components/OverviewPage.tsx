import { BASE_URL, CodeBlock } from './shared'

export function OverviewPage() {
  return (
    <div className="space-y-10">
      <div>
        <h2 className="text-2xl font-bold mb-4">Overview</h2>
        <p className="text-muted-foreground mb-6">
          The AgentBox API allows you to create and manage tasks executed by configured AI agents.
          Tasks are the core resource — you submit a prompt to an agent and receive results asynchronously.
        </p>
        <div className="rounded-lg border p-4 bg-muted/30">
          <p className="text-sm font-medium mb-2">Base URL</p>
          <CodeBlock code={BASE_URL} />
        </div>
      </div>

      {/* Authentication */}
      <section id="auth">
        <h3 className="text-xl font-bold mb-4">Authentication</h3>
        <p className="text-muted-foreground mb-4">
          Include your API key in the request header for agents configured with <code className="bg-muted px-1.5 py-0.5 rounded text-sm">api_key</code> access.
        </p>
        <CodeBlock code={`Authorization: Bearer YOUR_API_KEY`} />
        <div className="mt-4 rounded-lg border p-4 space-y-3">
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-bold px-2 py-0.5 rounded">public</span>
            <p className="text-sm text-muted-foreground">No authentication required. Anyone can call the endpoint.</p>
          </div>
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs font-bold px-2 py-0.5 rounded">api_key</span>
            <p className="text-sm text-muted-foreground">Requires <code className="bg-muted px-1 rounded">Authorization: Bearer</code> header with valid API key.</p>
          </div>
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400 text-xs font-bold px-2 py-0.5 rounded">private</span>
            <p className="text-sm text-muted-foreground">Admin-only access via dashboard. Not available for external API calls.</p>
          </div>
        </div>
      </section>

      {/* Errors */}
      <section id="errors">
        <h3 className="text-xl font-bold mb-4">Errors</h3>
        <p className="text-muted-foreground mb-4">
          The API uses standard HTTP status codes and returns JSON error responses.
        </p>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Code</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono">400</td>
                <td className="px-4 py-2.5 text-muted-foreground">Bad Request — Invalid parameters</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono">401</td>
                <td className="px-4 py-2.5 text-muted-foreground">Unauthorized — Missing or invalid API key</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono">404</td>
                <td className="px-4 py-2.5 text-muted-foreground">Not Found — Task or agent not found</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono">429</td>
                <td className="px-4 py-2.5 text-muted-foreground">Rate Limited — Too many requests</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2.5 font-mono">500</td>
                <td className="px-4 py-2.5 text-muted-foreground">Internal Error — Server-side failure</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div className="mt-4">
          <CodeBlock code={`{
  "error": "agent not found: invalid-id",
  "code": 400
}`} />
        </div>
      </section>
    </div>
  )
}
