import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  BookOpen,
  Key,
  Rocket,
  Terminal,
  AlertCircle,
  CheckCircle2,
  Copy,
  Bot,
  Layers,
  FileText,
  Zap,
  Shield,
  ExternalLink,
  ChevronRight,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'

export default function Documentation() {
  const [copiedCode, setCopiedCode] = useState<string | null>(null)

  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedCode(id)
      toast.success('Copied to clipboard')
      setTimeout(() => setCopiedCode(null), 2000)
    } catch {
      toast.error('Failed to copy')
    }
  }

  return (
    <div className="space-y-8">
      {/* Hero */}
      <div className="rounded-lg border bg-gradient-to-br from-primary/5 via-primary/10 to-primary/5 p-8">
        <div className="flex items-center gap-3 mb-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary">
            <BookOpen className="h-6 w-6 text-primary-foreground" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">Getting Started</h1>
            <p className="text-muted-foreground">Learn how to use AgentBox in 5 minutes</p>
          </div>
        </div>
      </div>

      {/* Quick Links */}
      <div className="grid gap-4 md:grid-cols-3">
        <QuickLink
          icon={Bot}
          title="Create Task"
          description="Run your first AI task"
          to="/tasks"
        />
        <QuickLink
          icon={Layers}
          title="Batch Processing"
          description="Process multiple inputs"
          to="/batches"
        />
        <QuickLink
          icon={FileText}
          title="API Reference"
          description="Integrate via REST API"
          to="/api-docs"
        />
      </div>

      {/* Setup Steps */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Rocket className="h-5 w-5 text-primary" />
            Quick Setup
          </CardTitle>
          <CardDescription>
            Follow these steps to start using AgentBox
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Step 1 */}
          <SetupStep
            number={1}
            title="Configure a Provider"
            icon={Key}
            iconColor="text-amber-500"
          >
            <p className="text-sm text-muted-foreground mb-3">
              Add your LLM API key in the Providers settings. AgentBox supports multiple providers:
            </p>
            <div className="grid gap-3 sm:grid-cols-2">
              <ProviderCard
                name="Anthropic (Claude)"
                description="Best for code understanding"
                url="https://console.anthropic.com"
                badge="Recommended"
              />
              <ProviderCard
                name="OpenAI"
                description="GPT-4 and GPT-3.5"
                url="https://platform.openai.com"
              />
              <ProviderCard
                name="智谱 AI"
                description="GLM-4 系列模型"
                url="https://open.bigmodel.cn"
              />
              <ProviderCard
                name="DeepSeek"
                description="DeepSeek Coder"
                url="https://platform.deepseek.com"
              />
            </div>
          </SetupStep>

          {/* Step 2 */}
          <SetupStep
            number={2}
            title="Create an Agent"
            icon={Bot}
            iconColor="text-blue-500"
          >
            <p className="text-sm text-muted-foreground mb-3">
              Agents define how tasks are executed. Configure:
            </p>
            <ul className="text-sm text-muted-foreground space-y-1.5 list-disc list-inside">
              <li><strong>Adapter</strong>: claude-code, codex, or custom</li>
              <li><strong>Model</strong>: Which LLM to use</li>
              <li><strong>System Prompt</strong>: Instructions for the agent</li>
              <li><strong>Permissions</strong>: What tools the agent can use</li>
            </ul>
          </SetupStep>

          {/* Step 3 */}
          <SetupStep
            number={3}
            title="Run a Task"
            icon={Zap}
            iconColor="text-green-500"
          >
            <p className="text-sm text-muted-foreground mb-3">
              Submit a prompt to your agent via UI or API:
            </p>
            <CodeBlock
              code={`curl -X POST http://localhost:18080/api/v1/tasks \\
  -H "Authorization: Bearer YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"agent_id": "agent-xxx", "prompt": "Hello, world!"}'`}
              onCopy={() => copyToClipboard('curl...', 'task')}
              copied={copiedCode === 'task'}
            />
          </SetupStep>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5 text-primary" />
            Key Features
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <FeatureCard
              icon={Bot}
              title="Multi-Engine Support"
              description="Use Claude Code, Codex, or custom adapters. Switch engines without changing your code."
            />
            <FeatureCard
              icon={Layers}
              title="Batch Processing"
              description="Process thousands of tasks in parallel with rate limiting, retries, and progress tracking."
            />
            <FeatureCard
              icon={Terminal}
              title="Real-time Streaming"
              description="Watch task execution in real-time via SSE. See agent thinking, tool calls, and results."
            />
            <FeatureCard
              icon={Shield}
              title="Enterprise Security"
              description="Role-based access control, isolated containers, and audit logging for compliance."
            />
          </div>
        </CardContent>
      </Card>

      {/* FAQ */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-primary" />
            FAQ
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="docker">
              <AccordionTrigger>Do I need Docker?</AccordionTrigger>
              <AccordionContent>
                Yes, AgentBox runs each task in an isolated Docker container for security.
                Make sure Docker Desktop is running before creating tasks.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="models">
              <AccordionTrigger>Which models are supported?</AccordionTrigger>
              <AccordionContent>
                Any OpenAI-compatible API works. We recommend Claude Sonnet 3.5/4 for best results.
                Also supports GPT-4, DeepSeek, GLM-4, and local models via Ollama.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="pricing">
              <AccordionTrigger>How much does it cost?</AccordionTrigger>
              <AccordionContent>
                AgentBox is open source and free. You only pay for the LLM API calls to your provider.
                Most tasks cost $0.01-0.10 depending on the model and complexity.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="security">
              <AccordionTrigger>Is it secure for production?</AccordionTrigger>
              <AccordionContent>
                Yes. Each task runs in an isolated container with restricted permissions.
                API keys are encrypted, and role-based access control prevents unauthorized access.
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardContent>
      </Card>

      {/* Next Steps */}
      <Card className="border-green-200 dark:border-green-900 bg-green-50/50 dark:bg-green-950/20">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="h-6 w-6 text-green-600 dark:text-green-400 shrink-0" />
            <div>
              <h3 className="font-semibold mb-2">Ready to Start!</h3>
              <p className="text-sm text-muted-foreground mb-4">
                You now have everything you need. Create your first task and watch the magic happen.
              </p>
              <div className="flex flex-wrap gap-2">
                <Button asChild>
                  <Link to="/tasks">
                    Create Task
                    <ChevronRight className="ml-1 h-4 w-4" />
                  </Link>
                </Button>
                <Button variant="outline" asChild>
                  <Link to="/api-docs">View API Docs</Link>
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function QuickLink({
  icon: Icon,
  title,
  description,
  to,
}: {
  icon: React.ElementType
  title: string
  description: string
  to: string
}) {
  return (
    <Link to={to}>
      <Card className="h-full transition-colors hover:bg-muted/50">
        <CardContent className="flex items-center gap-3 p-4">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
            <Icon className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h3 className="font-medium">{title}</h3>
            <p className="text-sm text-muted-foreground">{description}</p>
          </div>
          <ChevronRight className="ml-auto h-4 w-4 text-muted-foreground" />
        </CardContent>
      </Card>
    </Link>
  )
}

function SetupStep({
  number,
  title,
  icon: Icon,
  iconColor,
  children,
}: {
  number: number
  title: string
  icon: React.ElementType
  iconColor: string
  children: React.ReactNode
}) {
  return (
    <div className="flex gap-4">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">
        {number}
      </div>
      <div className="flex-1 space-y-3">
        <h3 className="font-semibold flex items-center gap-2">
          <Icon className={cn('h-4 w-4', iconColor)} />
          {title}
        </h3>
        {children}
      </div>
    </div>
  )
}

function ProviderCard({
  name,
  description,
  url,
  badge,
}: {
  name: string
  description: string
  url: string
  badge?: string
}) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="block rounded-lg border p-3 transition-colors hover:bg-muted/50"
    >
      <div className="flex items-center justify-between mb-1">
        <span className="font-medium text-sm">{name}</span>
        {badge && <Badge variant="secondary" className="text-xs">{badge}</Badge>}
      </div>
      <p className="text-xs text-muted-foreground flex items-center gap-1">
        {description}
        <ExternalLink className="h-3 w-3" />
      </p>
    </a>
  )
}

function FeatureCard({
  icon: Icon,
  title,
  description,
}: {
  icon: React.ElementType
  title: string
  description: string
}) {
  return (
    <div className="flex gap-3 p-3 rounded-lg border">
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
        <Icon className="h-4 w-4 text-primary" />
      </div>
      <div>
        <h4 className="font-medium text-sm">{title}</h4>
        <p className="text-xs text-muted-foreground mt-0.5">{description}</p>
      </div>
    </div>
  )
}

function CodeBlock({
  code,
  onCopy,
  copied,
}: {
  code: string
  onCopy: () => void
  copied: boolean
}) {
  return (
    <div className="relative">
      <pre className="rounded-lg bg-muted p-4 text-xs overflow-x-auto">
        <code>{code}</code>
      </pre>
      <Button
        size="icon"
        variant="ghost"
        className="absolute top-2 right-2 h-7 w-7"
        onClick={onCopy}
      >
        {copied ? (
          <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
        ) : (
          <Copy className="h-3.5 w-3.5" />
        )}
      </Button>
    </div>
  )
}
