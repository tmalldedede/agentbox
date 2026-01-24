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
          AgentBox supports two authentication methods: <strong>API Keys</strong> (recommended for programmatic access)
          and <strong>JWT Tokens</strong> (for interactive sessions).
        </p>
        <CodeBlock code={`Authorization: Bearer YOUR_API_KEY_OR_JWT_TOKEN`} />

        {/* Auth Methods */}
        <div className="mt-6 space-y-4">
          <div className="rounded-lg border p-4">
            <div className="flex items-center gap-2 mb-2">
              <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-bold px-2 py-0.5 rounded">Recommended</span>
              <h4 className="font-semibold">API Keys</h4>
            </div>
            <p className="text-sm text-muted-foreground mb-3">
              Long-lived credentials starting with <code className="bg-muted px-1 rounded">ab_</code>.
              Ideal for scripts, CI/CD pipelines, and backend integrations.
            </p>
            <ul className="text-sm text-muted-foreground list-disc pl-5 space-y-1">
              <li>Create via <a href="/api-keys" className="text-primary underline">API Keys page</a> or <a href="/api-docs/create-api-key" className="text-primary underline">API</a></li>
              <li>Never expire (unless configured)</li>
              <li>Can be revoked instantly</li>
              <li>Inherit your user permissions</li>
            </ul>
          </div>

          <div className="rounded-lg border p-4">
            <div className="flex items-center gap-2 mb-2">
              <h4 className="font-semibold">JWT Tokens</h4>
            </div>
            <p className="text-sm text-muted-foreground mb-3">
              Short-lived tokens obtained via <a href="/api-docs/login" className="text-primary underline">login endpoint</a>.
              Suitable for web applications and interactive sessions.
            </p>
            <ul className="text-sm text-muted-foreground list-disc pl-5 space-y-1">
              <li>Expire after 24 hours</li>
              <li>Obtained via username/password login</li>
              <li>Used internally by the web UI</li>
            </ul>
          </div>
        </div>

        {/* Access Levels */}
        <h4 className="font-semibold mt-6 mb-3">Access Levels</h4>
        <div className="rounded-lg border p-4 space-y-3">
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-bold px-2 py-0.5 rounded">public</span>
            <p className="text-sm text-muted-foreground">No authentication required. Only login and health endpoints.</p>
          </div>
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs font-bold px-2 py-0.5 rounded">user</span>
            <p className="text-sm text-muted-foreground">Requires valid API Key or JWT. Access to tasks, batches, files, and webhooks.</p>
          </div>
          <div className="flex items-start gap-3">
            <span className="shrink-0 inline-block bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 text-xs font-bold px-2 py-0.5 rounded">admin</span>
            <p className="text-sm text-muted-foreground">Full system access including user management, agents, providers, and settings.</p>
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
