import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
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
import { toast } from 'sonner'

export default function Documentation() {
  const navigate = useNavigate()
  const [copiedStep, setCopiedStep] = useState<string | null>(null)

  const copyToClipboard = async (text: string, step: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedStep(step)
      toast.success('Copied to clipboard')
      setTimeout(() => setCopiedStep(null), 2000)
    } catch {
      toast.error('Failed to copy to clipboard')
    }
  }

  return (
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <BookOpen className="w-5 h-5 text-emerald-400" />
            <span className="font-semibold">Getting Started with AgentBox</span>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto p-6 space-y-8">
        <div className="text-center py-8">
          <h1 className="text-4xl font-bold text-primary mb-4">Welcome to AgentBox! 🎉</h1>
          <p className="text-lg text-secondary max-w-2xl mx-auto">
            Run AI coding assistants in isolated containers with full control over tools,
            skills, and permissions. Let's get you set up in 5 minutes.
          </p>
        </div>

        <div className="card p-8 bg-gradient-to-br from-emerald-500/10 to-blue-500/10 border-emerald-500/30">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <Rocket className="w-6 h-6 text-emerald-400" />
            </div>
            <h2 className="text-2xl font-bold text-primary">Quick Setup (5 minutes)</h2>
          </div>

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
                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <h4 className="font-semibold text-primary mb-2">Anthropic (Recommended)</h4>
                    <p className="text-sm text-secondary mb-3">
                      Get Claude API key for best reasoning and code understanding
                    </p>
                    <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                      <li>
                        Visit{' '}
                        <a
                          href="https://console.anthropic.com"
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-emerald-400 hover:underline"
                        >
                          console.anthropic.com
                        </a>
                      </li>
                      <li>Sign up or log in</li>
                      <li>Go to "API Keys" section</li>
                      <li>Click "Create Key"</li>
                      <li>
                        Copy the key (starts with{' '}
                        <code className="text-emerald-400">sk-ant-</code>)
                      </li>
                    </ol>
                    <div className="mt-3 p-2 rounded bg-blue-500/10 border border-blue-500/20">
                      <p className="text-xs text-blue-400">💡 New users get $5 free credits</p>
                    </div>
                  </div>

                  <div className="p-4 rounded-lg bg-secondary border border-default">
                    <h4 className="font-semibold text-primary mb-2">OpenAI</h4>
                    <p className="text-sm text-secondary mb-3">
                      Get OpenAI API key for GPT models
                    </p>
                    <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                      <li>
                        Visit{' '}
                        <a
                          href="https://platform.openai.com"
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-emerald-400 hover:underline"
                        >
                          platform.openai.com
                        </a>
                      </li>
                      <li>Sign up or log in</li>
                      <li>Navigate to "API Keys"</li>
                      <li>Create a new secret key</li>
                      <li>
                        Copy the key (starts with{' '}
                        <code className="text-emerald-400">sk-</code>)
                      </li>
                    </ol>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-lg">
                2
              </div>
              <div className="flex-1">
                <h3 className="text-xl font-bold text-primary mb-2 flex items-center gap-2">
                  <Terminal className="w-5 h-5 text-blue-400" />
                  Add Your API Key
                </h3>
                <p className="text-secondary mb-4">Go to Settings and add your API key:</p>
                <div className="bg-secondary p-4 rounded-lg border border-default">
                  <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                    <li>Click the Settings gear icon in the sidebar</li>
                    <li>Find the API Keys section</li>
                    <li>Paste your key in the appropriate field</li>
                    <li>Click Save</li>
                  </ol>
                </div>
              </div>
            </div>

            <div className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-lg">
                3
              </div>
              <div className="flex-1">
                <h3 className="text-xl font-bold text-primary mb-2 flex items-center gap-2">
                  <Code className="w-5 h-5 text-purple-400" />
                  Create Your First Session
                </h3>
                <p className="text-secondary mb-4">Start a new coding session:</p>
                <div className="bg-secondary p-4 rounded-lg border border-default">
                  <ol className="space-y-2 text-sm text-secondary list-decimal list-inside">
                    <li>Go to the Sessions page</li>
                    <li>Click "New Session"</li>
                    <li>Choose your agent (Claude Code recommended)</li>
                    <li>Select your workspace directory</li>
                    <li>Click Create</li>
                  </ol>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="card p-8">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-blue-500/20 flex items-center justify-center">
              <Terminal className="w-6 h-6 text-blue-400" />
            </div>
            <h2 className="text-2xl font-bold text-primary">Common Commands</h2>
          </div>

          <div className="space-y-4">
            <CommandCard
              title="Create a commit"
              description="Generate a conventional commit message based on your staged changes"
              command="/commit"
              onCopy={() => copyToClipboard('/commit', 'commit')}
              copied={copiedStep === 'commit'}
            />
            <CommandCard
              title="Review a PR"
              description="Analyze a pull request and provide feedback"
              command="/review-pr"
              onCopy={() => copyToClipboard('/review-pr', 'review-pr')}
              copied={copiedStep === 'review-pr'}
            />
            <CommandCard
              title="Summarize changes"
              description="Get a summary of your git diff"
              command="/summarize"
              onCopy={() => copyToClipboard('/summarize', 'summarize')}
              copied={copiedStep === 'summarize'}
            />
          </div>
        </div>

        <div className="card p-8 bg-amber-500/5 border-amber-500/20">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-6 h-6 text-amber-400 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-primary mb-2">Need Help?</h3>
              <p className="text-secondary mb-3">
                If you run into issues, check the following:
              </p>
              <ul className="space-y-2 text-sm text-secondary">
                <li>• Make sure Docker is running</li>
                <li>• Verify your API key is valid</li>
                <li>• Check that your workspace path exists</li>
                <li>• Review logs in the session detail page</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="card p-8 bg-emerald-500/5 border-emerald-500/20">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="w-6 h-6 text-emerald-400 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-primary mb-2">You're Ready to Go!</h3>
              <p className="text-secondary">
                Start by creating a session and asking your agent to help with your coding tasks.
                You can customize profiles, add MCP servers, and create custom skills as you go.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function CommandCard({
  title,
  description,
  command: _command,
  onCopy,
  copied,
}: {
  title: string
  description: string
  command: string
  onCopy: () => void
  copied: boolean
}) {
  return (
    <div className="p-4 rounded-lg bg-secondary border border-default flex items-center justify-between">
      <div>
        <h4 className="font-semibold text-primary mb-1">{title}</h4>
        <p className="text-sm text-secondary">{description}</p>
      </div>
      <button onClick={onCopy} className="btn btn-ghost btn-icon">
        {copied ? <CheckCircle2 className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
      </button>
    </div>
  )
}
