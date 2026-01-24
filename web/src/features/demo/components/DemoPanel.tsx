import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Play, Loader2, CheckCircle2, ArrowRight, Sparkles } from 'lucide-react'
import { useDockerAvailable } from '@/hooks/useSystemHealth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'sonner'
import { api } from '@/services/api'

interface DemoTask {
  id: string
  name: string
  description: string
  agent_id: string
  prompt: string
  icon: string
}

const DEMO_TASKS: DemoTask[] = [
  {
    id: 'code-review',
    name: 'Code Review',
    description: 'Analyze code quality and suggest improvements',
    agent_id: 'claude-code-default',
    prompt: 'Review the code in this repository. Identify potential bugs, security issues, and suggest improvements for code quality and maintainability.',
    icon: '🔍',
  },
  {
    id: 'doc-generator',
    name: 'Documentation Generator',
    description: 'Generate comprehensive documentation',
    agent_id: 'claude-code-default',
    prompt: 'Analyze this codebase and generate comprehensive documentation including: 1) Project overview, 2) Architecture diagram in mermaid, 3) API reference, 4) Getting started guide.',
    icon: '📚',
  },
  {
    id: 'security-scan',
    name: 'Security Scan',
    description: 'Perform security vulnerability analysis',
    agent_id: 'security-research',
    prompt: 'Perform a security audit of this codebase. Check for: 1) OWASP Top 10 vulnerabilities, 2) Hardcoded secrets, 3) Insecure dependencies, 4) Authentication/authorization issues. Provide a detailed report.',
    icon: '🔒',
  },
  {
    id: 'test-generator',
    name: 'Test Generator',
    description: 'Generate unit tests for the codebase',
    agent_id: 'claude-code-auto',
    prompt: 'Analyze the codebase and generate comprehensive unit tests. Focus on: 1) Critical business logic, 2) Edge cases, 3) Error handling. Use the existing test framework if present.',
    icon: '🧪',
  },
]

export function DemoPanel() {
  const navigate = useNavigate()
  const dockerAvailable = useDockerAvailable()
  const [selectedDemo, setSelectedDemo] = useState<DemoTask | null>(null)
  const [customPrompt, setCustomPrompt] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<string | null>(null)

  const runDemo = async (demo: DemoTask) => {
    setLoading(true)
    setSelectedDemo(demo)
    setResult(null)

    try {
      // Create a session using agent_id
      const session = await api.createSession({
        agent_id: demo.agent_id,
        workspace: '/tmp/demo',
      })

      const sessionId = session.id

      // Start the session
      await api.startSession(sessionId)

      // Execute the demo prompt
      const execRes = await api.execSession(sessionId, {
        prompt: customPrompt || demo.prompt,
        max_turns: 10,
        timeout: 120,
      })

      setResult(execRes.output || execRes.message || 'Demo completed successfully!')
      toast.success('Demo completed successfully')

      // Stop the session
      await api.stopSession(sessionId)
    } catch (error: unknown) {
      const err = error as { message?: string }
      const errorMessage = err.message || 'Demo failed'
      setResult(`Error: ${errorMessage}`)
      toast.error(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Sparkles className="h-6 w-6 text-yellow-500" />
            Try Demo
          </h2>
          <p className="text-muted-foreground">
            Experience AgentBox with pre-configured demo tasks
          </p>
        </div>
        <Button variant="outline" onClick={() => navigate({ to: '/agents' })}>
          Create Your Own
          <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </div>

      {/* Demo Cards */}
      <div className="grid grid-cols-2 gap-4">
        {DEMO_TASKS.map((demo) => (
          <Card
            key={demo.id}
            className={`cursor-pointer transition-all hover:shadow-md ${
              selectedDemo?.id === demo.id ? 'ring-2 ring-primary' : ''
            }`}
            onClick={() => {
              setSelectedDemo(demo)
              setCustomPrompt(demo.prompt)
            }}
          >
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2">
                <span className="text-2xl">{demo.icon}</span>
                {demo.name}
              </CardTitle>
              <CardDescription>{demo.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground line-clamp-2">
                {demo.prompt}
              </p>
              <div className="mt-3 flex items-center justify-between">
                <span className="text-xs bg-secondary px-2 py-1 rounded">
                  {demo.agent_id}
                </span>
                <Button
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    runDemo(demo)
                  }}
                  disabled={loading || !dockerAvailable}
                >
                  {loading && selectedDemo?.id === demo.id ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Play className="mr-2 h-4 w-4" />
                  )}
                  Run
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Custom Prompt Editor */}
      {selectedDemo && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className="text-xl">{selectedDemo.icon}</span>
              {selectedDemo.name}
            </CardTitle>
            <CardDescription>
              Customize the prompt before running
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Textarea
              value={customPrompt}
              onChange={(e) => setCustomPrompt(e.target.value)}
              rows={5}
              placeholder="Enter your custom prompt..."
            />
            <Button
              className="w-full"
              onClick={() => runDemo(selectedDemo)}
              disabled={loading || !dockerAvailable}
            >
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Running...
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  Run Demo
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Result */}
      {result && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle2 className="h-5 w-5 text-green-500" />
              Result
            </CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="bg-secondary p-4 rounded-lg overflow-auto max-h-96 text-sm whitespace-pre-wrap">
              {result}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default DemoPanel
