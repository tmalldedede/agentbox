import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  BookOpen,
  Key,
  Rocket,
  Code,
  Terminal,
  AlertCircle,
  CheckCircle2,
  Copy,
} from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

export default function Documentation() {
  const navigate = useNavigate()
  const [copiedStep, setCopiedStep] = useState<string | null>(null)

  const copyToClipboard = (text: string, step: string) => {
    navigator.clipboard.writeText(text)
    setCopiedStep(step)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopiedStep(null), 2000)
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <BookOpen className="w-5 h-5 text-emerald-400" />
            <span className="font-semibold">Getting Started with AgentBox</span>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto p-6 space-y-8">
        {/* Hero */}
        <div className="text-center py-8">
          <h1 className="text-4xl font-bold text-primary mb-4">
            Welcome to AgentBox! 🎉
          </h1>
          <p className="text-lg text-secondary max-w-2xl mx-auto">
            Run AI coding assistants in isolated containers with full control over tools,
            skills, and permissions. Let's get you set up in 5 minutes.
          </p>
        </div>

        {/* Quick Setup */}
        <div className="card p-8 bg-gradient-to-br from-emerald-500/10 to-blue-500/10 border-emerald-500/30">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <Rocket className="w-6 h-6 text-emerald-400" />
            </div>
            <h2 className="text-2xl font-bold text-primary">Quick Setup (5 minutes)</h2>
          </div>

          {/* Step 1: Get API Key */}
          <div className="space-y-6">
            <div className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-lg">
                1
              </div>
              <div className="flex-1">
                <h3 className="text-xl font-bold text-primary mb-2 flex items-center gap-2">
                  <Key className="w-5 h-5 text-amber-400" />
                  Get Your API Key
                </h3>
                <p className="text-secondary mb-4">
                  You need an API key from Anthropic (Claude) or OpenAI. Choose one based on your preference:
                </p>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Anthropic */}
                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <h4 className="font-semibold text-primary mb-2">Anthropic (Recommended)</h4>
                    <p className="text-sm text-secondary mb-3">
                      Get Claude API key for best reasoning and code understanding
                    </p>
                    <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                      <li>Visit <a href="https://console.anthropic.com" target="_blank" rel="noopener noreferrer" className="text-emerald-400 hover:underline">console.anthropic.com</a></li>
                      <li>Sign up or log in</li>
                      <li>Go to "API Keys" section</li>
                      <li>Click "Create Key"</li>
                      <li>Copy the key (starts with <code className="text-emerald-400">sk-ant-</code>)</li>
                    </ol>
                    <div className="mt-3 p-2 rounded bg-blue-500/10 border border-blue-500/20">
                      <p className="text-xs text-blue-400">
                        💡 New users get $5 free credits
                      </p>
                    </div>
                  </div>

                  {/* OpenAI */}
                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <h4 className="font-semibold text-primary mb-2">OpenAI</h4>
                    <p className="text-sm text-secondary mb-3">
                      Get OpenAI API key for GPT models
                    </p>
                    <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                      <li>Visit <a href="https://platform.openai.com" target="_blank" rel="noopener noreferrer" className="text-emerald-400 hover:underline">platform.openai.com</a></li>
                      <li>Sign up or log in</li>
                      <li>Go to "API keys" page</li>
                      <li>Click "Create new secret key"</li>
                      <li>Copy the key (starts with <code className="text-emerald-400">sk-</code>)</li>
                    </ol>
                    <div className="mt-3 p-2 rounded bg-amber-500/10 border border-amber-500/20">
                      <p className="text-xs text-amber-400">
                        ⚠️ Requires payment method
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Step 2: Configure API Key */}
            <div className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-lg">
                2
              </div>
              <div className="flex-1">
                <h3 className="text-xl font-bold text-primary mb-2">Configure Your API Key</h3>
                <p className="text-secondary mb-4">
                  Add your API key to AgentBox Settings:
                </p>

                <div className="space-y-3">
                  <div className="p-4 rounded-lg bg-secondary border border-emerald-500/30">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-primary mb-1">1. Open Settings</p>
                        <p className="text-sm text-secondary">
                          Click the <strong>⚙️ Settings</strong> button in the sidebar (bottom left)
                        </p>
                      </div>
                      <button
                        onClick={() => navigate('/settings')}
                        className="btn btn-primary btn-sm"
                      >
                        Go to Settings
                      </button>
                    </div>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <p className="text-sm font-medium text-primary mb-1">2. Scroll to "API Keys" section</p>
                    <p className="text-sm text-secondary">
                      Enter your API key in the appropriate field (ANTHROPIC_API_KEY or OPENAI_API_KEY)
                    </p>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <p className="text-sm font-medium text-primary mb-1">3. Click "Save"</p>
                    <p className="text-sm text-secondary">
                      Your key is stored locally in your browser (not sent to our servers)
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Step 3: Start a Session */}
            <div className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-lg">
                3
              </div>
              <div className="flex-1">
                <h3 className="text-xl font-bold text-primary mb-2">Start Your First Session</h3>
                <p className="text-secondary mb-4">
                  Use a pre-configured profile to launch an AI assistant:
                </p>

                <div className="space-y-3">
                  <div className="p-4 rounded-lg bg-secondary border border-emerald-500/30">
                    <p className="text-sm font-medium text-primary mb-2">1. Go to Profiles</p>
                    <button
                      onClick={() => navigate('/profiles')}
                      className="btn btn-secondary btn-sm"
                    >
                      View Profiles
                    </button>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <p className="text-sm font-medium text-primary mb-2">2. Pick a Profile</p>
                    <p className="text-sm text-secondary mb-2">
                      We recommend starting with <strong className="text-emerald-400">"Claude Code - Default"</strong>
                    </p>
                    <p className="text-xs text-muted">
                      This profile includes filesystem access and basic coding skills
                    </p>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <p className="text-sm font-medium text-primary mb-2">3. Click the profile card</p>
                    <p className="text-sm text-secondary">
                      This opens the profile details page
                    </p>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <p className="text-sm font-medium text-primary mb-2">4. Click "Start Session"</p>
                    <p className="text-sm text-secondary">
                      AgentBox will launch a Docker container with the AI assistant (takes ~10 seconds)
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="mt-8 p-4 rounded-lg bg-emerald-500/20 border border-emerald-500/30">
            <div className="flex items-start gap-3">
              <CheckCircle2 className="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
              <div>
                <p className="font-semibold text-emerald-400 mb-1">That's it! You're ready to go! 🎉</p>
                <p className="text-sm text-secondary">
                  Once your session is running, go to <strong>Tasks</strong> and submit your first coding task.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Real Use Cases */}
        <div className="card p-8">
          <div className="flex items-center gap-3 mb-6">
            <Code className="w-8 h-8 text-purple-400" />
            <h2 className="text-2xl font-bold text-primary">Real Use Cases</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Use Case 1 */}
            <div className="p-5 rounded-lg bg-secondary border border-default hover:border-emerald-500/50 transition-colors">
              <h3 className="font-semibold text-primary mb-2">🐛 Fix a Bug</h3>
              <p className="text-sm text-secondary mb-3">
                Submit a task: "There's a bug in users/api.ts where the login function doesn't handle network errors. Fix it and add error handling."
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => copyToClipboard("There's a bug in users/api.ts where the login function doesn't handle network errors. Fix it and add error handling.", "bug")}
                  className="btn btn-ghost btn-sm text-xs"
                >
                  {copiedStep === "bug" ? <CheckCircle2 className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                  Copy
                </button>
              </div>
            </div>

            {/* Use Case 2 */}
            <div className="p-5 rounded-lg bg-secondary border border-default hover:border-emerald-500/50 transition-colors">
              <h3 className="font-semibold text-primary mb-2">✨ Add a Feature</h3>
              <p className="text-sm text-secondary mb-3">
                Submit a task: "Add a dark mode toggle to the Settings page. Use React context for theme state."
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => copyToClipboard("Add a dark mode toggle to the Settings page. Use React context for theme state.", "feature")}
                  className="btn btn-ghost btn-sm text-xs"
                >
                  {copiedStep === "feature" ? <CheckCircle2 className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                  Copy
                </button>
              </div>
            </div>

            {/* Use Case 3 */}
            <div className="p-5 rounded-lg bg-secondary border border-default hover:border-emerald-500/50 transition-colors">
              <h3 className="font-semibold text-primary mb-2">🧪 Write Tests</h3>
              <p className="text-sm text-secondary mb-3">
                Submit a task: "Generate comprehensive unit tests for the utils/validation.ts file. Use Jest."
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => copyToClipboard("Generate comprehensive unit tests for the utils/validation.ts file. Use Jest.", "test")}
                  className="btn btn-ghost btn-sm text-xs"
                >
                  {copiedStep === "test" ? <CheckCircle2 className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                  Copy
                </button>
              </div>
            </div>

            {/* Use Case 4 */}
            <div className="p-5 rounded-lg bg-secondary border border-default hover:border-emerald-500/50 transition-colors">
              <h3 className="font-semibold text-primary mb-2">📝 Refactor Code</h3>
              <p className="text-sm text-secondary mb-3">
                Submit a task: "Refactor the Dashboard component to use React hooks instead of class components."
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => copyToClipboard("Refactor the Dashboard component to use React hooks instead of class components.", "refactor")}
                  className="btn btn-ghost btn-sm text-xs"
                >
                  {copiedStep === "refactor" ? <CheckCircle2 className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                  Copy
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Key Concepts */}
        <div className="card p-8">
          <div className="flex items-center gap-3 mb-6">
            <Terminal className="w-8 h-8 text-cyan-400" />
            <h2 className="text-2xl font-bold text-primary">Key Concepts (Optional Reading)</h2>
          </div>
          <p className="text-secondary mb-6">
            Understanding these concepts will help you get the most out of AgentBox:
          </p>

          <div className="space-y-4">
            <details className="p-4 rounded-lg bg-secondary border border-default">
              <summary className="font-semibold text-primary cursor-pointer">
                What are Profiles?
              </summary>
              <p className="text-sm text-secondary mt-3">
                A Profile is a template that bundles an AI agent with specific tools (MCP servers),
                skills (commands like /commit), and credentials. Think of it as a preset configuration
                you can reuse. For example, a "Full-Stack Dev" profile might include filesystem access,
                GitHub integration, and code review skills.
              </p>
            </details>

            <details className="p-4 rounded-lg bg-secondary border border-default">
              <summary className="font-semibold text-primary cursor-pointer">
                What are Sessions?
              </summary>
              <p className="text-sm text-secondary mt-3">
                A Session is a running instance of an AI agent in an isolated Docker container.
                Each session has its own workspace and can work on different projects simultaneously.
                Sessions stay alive until you stop them, so you can submit multiple tasks to the same session.
              </p>
            </details>

            <details className="p-4 rounded-lg bg-secondary border border-default">
              <summary className="font-semibold text-primary cursor-pointer">
                What are MCP Servers?
              </summary>
              <p className="text-sm text-secondary mt-3">
                MCP (Model Context Protocol) Servers give agents superpowers. They're plugins that
                let agents access your filesystem, databases, APIs, web browsers, and more.
                For example, the filesystem MCP lets the agent read and write files in your project.
              </p>
            </details>

            <details className="p-4 rounded-lg bg-secondary border border-default">
              <summary className="font-semibold text-primary cursor-pointer">
                What are Skills?
              </summary>
              <p className="text-sm text-secondary mt-3">
                Skills are reusable command templates invoked with slash commands (like /commit, /review-pr).
                They contain detailed instructions that guide the agent through specific workflows.
                You can create custom skills for your team's processes.
              </p>
            </details>
          </div>
        </div>

        {/* Troubleshooting */}
        <div className="card p-8 bg-amber-500/5 border-amber-500/20">
          <div className="flex items-center gap-3 mb-6">
            <AlertCircle className="w-8 h-8 text-amber-400" />
            <h2 className="text-2xl font-bold text-primary">Common Issues</h2>
          </div>

          <div className="space-y-4">
            <div className="p-4 rounded-lg bg-secondary border border-default">
              <h4 className="font-semibold text-primary mb-2">❌ "Failed to start session"</h4>
              <p className="text-sm text-secondary mb-2">
                <strong>Solution:</strong> Make sure Docker Desktop is running. On Mac, check the whale icon in the menu bar.
              </p>
            </div>

            <div className="p-4 rounded-lg bg-secondary border border-default">
              <h4 className="font-semibold text-primary mb-2">❌ "API key invalid"</h4>
              <p className="text-sm text-secondary mb-2">
                <strong>Solution:</strong> Double-check your API key in Settings. Make sure there are no extra spaces.
              </p>
            </div>

            <div className="p-4 rounded-lg bg-secondary border border-default">
              <h4 className="font-semibold text-primary mb-2">❌ "Agent is not responding"</h4>
              <p className="text-sm text-secondary mb-2">
                <strong>Solution:</strong> The task might be taking longer than expected. Check the task status page for updates.
              </p>
            </div>
          </div>
        </div>

        {/* Help */}
        <div className="card p-8 text-center">
          <h3 className="text-xl font-bold text-primary mb-3">Still Need Help?</h3>
          <p className="text-secondary mb-6">
            Visit our GitHub repository for detailed documentation, video tutorials, and community support.
          </p>
          <a
            href="https://github.com/tmalldedede/agentbox"
            target="_blank"
            rel="noopener noreferrer"
            className="btn btn-primary inline-flex items-center gap-2"
          >
            <BookOpen className="w-4 h-4" />
            View Documentation
          </a>
        </div>
      </div>
    </div>
  )
}
