import {
  BASE_URL,
  EndpointHeader,
  ParamTable,
  ExamplesPanel,
} from './shared'

export function DownloadFilePage() {
  return (
    <div className="space-y-8">
      <EndpointHeader
        method="GET"
        path="/files/:id/download"
        title="Download File"
      />

      <p className="text-muted-foreground">
        Download a file by its ID. The response is the raw file content with appropriate
        <code className="bg-muted px-1.5 py-0.5 rounded text-sm mx-1">Content-Type</code>
        and
        <code className="bg-muted px-1.5 py-0.5 rounded text-sm mx-1">Content-Disposition</code>
        headers set for the browser to handle.
      </p>

      {/* Path Parameters */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Path Parameters</h3>
        <ParamTable params={[
          { name: 'id', type: 'string', required: true, desc: 'The file ID to download', example: '"550e8400-e29b-41d4-a716-446655440000"' },
        ]} />
      </div>

      {/* Response Headers */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response Headers</h3>
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr>
                <th className="text-left px-4 py-2 font-medium">Header</th>
                <th className="text-left px-4 py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-t">
                <td className="px-4 py-2 font-mono">Content-Type</td>
                <td className="px-4 py-2 text-muted-foreground">MIME type of the file (e.g., <code className="bg-muted px-1 rounded">application/pdf</code>)</td>
              </tr>
              <tr className="border-t">
                <td className="px-4 py-2 font-mono">Content-Disposition</td>
                <td className="px-4 py-2 text-muted-foreground">
                  Attachment directive with original filename<br />
                  <code className="text-xs bg-muted px-1 rounded">attachment; filename="document.pdf"</code>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Response */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Response</h3>
        <div className="rounded-lg border p-4 space-y-2 text-sm">
          <p className="text-muted-foreground">
            <span className="inline-block bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-mono font-bold px-1.5 py-0.5 rounded mr-2">200</span>
            Raw file content (binary)
          </p>
        </div>
      </div>

      {/* Code Examples */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Examples</h3>
        <ExamplesPanel
          examples={{
            curl: `# Download to file
curl ${BASE_URL}/files/550e8400-e29b-41d4-a716-446655440000/download \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -o downloaded_file.pdf

# Or view headers only
curl -I ${BASE_URL}/files/550e8400-e29b-41d4-a716-446655440000/download \\
  -H "Authorization: Bearer YOUR_API_KEY"`,
            python: `import requests

file_id = "550e8400-e29b-41d4-a716-446655440000"

response = requests.get(
    f"${BASE_URL}/files/{file_id}/download",
    headers={"Authorization": "Bearer YOUR_API_KEY"},
)

# Get filename from Content-Disposition header
content_disp = response.headers.get("Content-Disposition", "")
filename = "downloaded_file"
if 'filename="' in content_disp:
    filename = content_disp.split('filename="')[1].rstrip('"')

# Save to file
with open(filename, "wb") as f:
    f.write(response.content)

print(f"Downloaded: {filename}")`,
            javascript: `const fileId = "550e8400-e29b-41d4-a716-446655440000";

const response = await fetch(
  \`${BASE_URL}/files/\${fileId}/download\`,
  {
    headers: { "Authorization": "Bearer YOUR_API_KEY" },
  }
);

// Get filename from header
const contentDisp = response.headers.get("Content-Disposition") || "";
const match = contentDisp.match(/filename="(.+)"/);
const filename = match ? match[1] : "downloaded_file";

// Create download link
const blob = await response.blob();
const url = URL.createObjectURL(blob);
const a = document.createElement("a");
a.href = url;
a.download = filename;
a.click();
URL.revokeObjectURL(url);`,
          }}
        />
      </div>

      {/* Direct Link */}
      <div className="space-y-3">
        <h3 className="font-semibold text-lg">Direct Link</h3>
        <p className="text-sm text-muted-foreground">
          You can also construct a direct download URL and open it in a browser or embed it in applications:
        </p>
        <div className="rounded-lg bg-muted p-3 font-mono text-sm">
          {BASE_URL}/files/&#123;file_id&#125;/download
        </div>
      </div>
    </div>
  )
}
