import { useEffect, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Save,
  Zap,
  Upload,
  Sparkles,
  Loader2,
  FileArchive,
  CheckCircle2,
  AlertCircle,
  ChevronRight,
} from 'lucide-react'
import { api } from '@/services/api'
import { toast } from 'sonner'
import { useSessions } from '@/hooks'

type TabType = 'import' | 'generate'

export default function SkillForm() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<TabType>('import')

  return (
    <div className="min-h-screen">
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/skills' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Zap className="w-6 h-6 text-emerald-400" />
            <span className="text-lg font-bold">Add New Skill</span>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto p-6">
        <div className="flex gap-4 mb-8 border-b border-default">
          <button
            onClick={() => setActiveTab('import')}
            className={`px-6 py-3 font-medium transition-colors relative ${
              activeTab === 'import'
                ? 'text-emerald-400'
                : 'text-secondary hover:text-primary'
            }`}
          >
            <div className="flex items-center gap-2">
              <Upload className="w-4 h-4" />
              Import Skill Package
            </div>
            {activeTab === 'import' && (
              <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-emerald-400" />
            )}
          </button>

          <button
            onClick={() => setActiveTab('generate')}
            className={`px-6 py-3 font-medium transition-colors relative ${
              activeTab === 'generate'
                ? 'text-emerald-400'
                : 'text-secondary hover:text-primary'
            }`}
          >
            <div className="flex items-center gap-2">
              <Sparkles className="w-4 h-4" />
              Generate with AI
            </div>
            {activeTab === 'generate' && (
              <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-emerald-400" />
            )}
          </button>
        </div>

        {activeTab === 'import' ? <ImportTab /> : <GenerateTab />}
      </div>
    </div>
  )
}

function ImportTab() {
  const [file, setFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [dragActive, setDragActive] = useState(false)

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0])
    }
  }

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0])
    }
  }

  const handleFile = (file: File) => {
    if (!file.name.endsWith('.skill')) {
      toast.error('Please upload a .skill file')
      return
    }
    setFile(file)
  }

  const handleUpload = async () => {
    if (!file) return

    setUploading(true)
    try {
      toast.info('Upload functionality coming soon - this requires backend implementation')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to import skill')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="card p-6 bg-blue-500/5 border-blue-500/20">
        <div className="flex items-start gap-3">
          <FileArchive className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="font-semibold text-primary mb-2">Import a Skill Package</h3>
            <p className="text-sm text-secondary mb-3">
              Upload a <code className="text-emerald-400">.skill</code> file (packaged skill) to import it into AgentBox.
            </p>
            <p className="text-sm text-secondary">
              The .skill file contains SKILL.md with metadata, optional scripts, references, and assets.
            </p>
          </div>
        </div>
      </div>

      <div
        className={`card p-12 border-2 border-dashed transition-colors ${
          dragActive
            ? 'border-emerald-500 bg-emerald-500/10'
            : 'border-default hover:border-emerald-500/50'
        }`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        <div className="flex flex-col items-center text-center">
          {file ? (
            <>
              <CheckCircle2 className="w-16 h-16 text-emerald-400 mb-4" />
              <p className="text-lg font-semibold text-primary mb-2">{file.name}</p>
              <p className="text-sm text-secondary mb-4">{(file.size / 1024).toFixed(2)} KB</p>
              <div className="flex gap-3">
                <button onClick={() => setFile(null)} className="btn btn-ghost">
                  Remove
                </button>
                <button onClick={handleUpload} className="btn btn-primary" disabled={uploading}>
                  {uploading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Importing...
                    </>
                  ) : (
                    <>
                      <Upload className="w-4 h-4" />
                      Import Skill
                    </>
                  )}
                </button>
              </div>
            </>
          ) : (
            <>
              <Upload className="w-16 h-16 text-muted mb-4" />
              <p className="text-lg font-semibold text-primary mb-2">Drop your .skill file here</p>
              <p className="text-sm text-secondary mb-4">or click to browse</p>
              <label className="btn btn-primary cursor-pointer">
                <Upload className="w-4 h-4" />
                Select File
                <input type="file" className="hidden" accept=".skill" onChange={handleFileInput} />
              </label>
            </>
          )}
        </div>
      </div>

      <div className="card p-6">
        <h3 className="font-semibold text-primary mb-3">Expected Format</h3>
        <p className="text-sm text-secondary mb-4">
          A .skill file is a zip archive with the following structure:
        </p>
        <pre className="text-xs text-secondary bg-secondary p-4 rounded-lg overflow-x-auto">
{`skill-name/
├── SKILL.md              # Required: Main documentation
│   ├── --- (YAML frontmatter)
│   │   ├── name: skill-id
│   │   └── description: ...
│   └── (Markdown body)
├── scripts/              # Optional: Executable code
│   └── tool.py
├── references/           # Optional: Reference docs
│   └── examples.md
└── assets/              # Optional: Output resources
    └── template.html`}
        </pre>
      </div>
    </div>
  )
}

function GenerateTab() {
  const navigate = useNavigate()
  const { data: sessions = [] } = useSessions()
  const runningSessions = sessions.filter(s => s.status === 'running')
  const pollIntervalRef = useRef<number | null>(null)

  const [step, setStep] = useState<'input' | 'generating' | 'preview'>('input')
  const [requirements, setRequirements] = useState('')
  const [selectedSession, setSelectedSession] = useState('')
  const [taskId, setTaskId] = useState('')
  const [generatedSkill, setGeneratedSkill] = useState<any>(null)

  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) {
        window.clearInterval(pollIntervalRef.current)
      }
    }
  }, [])

  const handleGenerate = async () => {
    if (!requirements.trim()) {
      toast.error('Please describe your skill requirements')
      return
    }
    if (!selectedSession) {
      toast.error('Please select a session')
      return
    }

    const session = sessions.find(s => s.id === selectedSession)
    if (!session?.profile_id) {
      toast.error('Selected session does not have a profile')
      return
    }

    setStep('generating')

    try {
      const prompt = `You are a skill creator. Generate a complete, production-ready skill based on the following requirements.

## User Requirements:
${requirements}

## Your Task:
Follow the skill-creator workflow to generate a complete skill:

1. **Analyze the requirements** and identify what scripts, references, and assets are needed
2. **Generate SKILL.md** with proper YAML frontmatter (name, description)
3. **Generate any necessary files** (scripts/, references/, assets/)
4. **Output as JSON** in this exact format:

\`\`\`json
{
  "id": "skill-id",
  "name": "Skill Display Name",
  "description": "Complete description including what the skill does and when to use it. Be specific about triggers.",
  "command": "/skill-command",
  "prompt": "The complete SKILL.md body (markdown instructions)",
  "files": [
    {"path": "scripts/tool.py", "content": "#!/usr/bin/env python3\\n..."},
    {"path": "references/guide.md", "content": "# Guide\\n..."},
    {"path": "requirements.txt", "content": "requests>=2.28.0\\n"}
  ],
  "category": "coding",
  "tags": ["tag1", "tag2"],
  "version": "1.0.0",
  "author": "AgentBox User"
}
\`\`\`

**Important Guidelines:**
- Make the description comprehensive and include specific trigger phrases
- Keep SKILL.md concise (under 500 lines)
- Include complete, runnable code in scripts
- Only create files that are actually needed
- Follow the progressive disclosure principle`

      const task = await api.createTask({
        profile_id: session.profile_id,
        prompt,
      })

      setTaskId(task.id)

      if (pollIntervalRef.current) {
        window.clearInterval(pollIntervalRef.current)
      }

      pollIntervalRef.current = window.setInterval(async () => {
        try {
          const taskStatus = await api.getTask(task.id)

          if (taskStatus.status === 'completed') {
            if (pollIntervalRef.current) {
              window.clearInterval(pollIntervalRef.current)
            }

            const resultText = taskStatus.result?.text || taskStatus.result?.summary || ''

            try {
              const jsonMatch =
                resultText.match(/```json\n([\s\S]*?)\n```/) || resultText.match(/({[\s\S]*})/)

              if (jsonMatch) {
                const skillData = JSON.parse(jsonMatch[1])
                setGeneratedSkill(skillData)
                setStep('preview')
                toast.success('Skill generated successfully!')
              } else {
                throw new Error('Could not parse skill JSON from response')
              }
            } catch (parseErr) {
              console.error('Parse error:', parseErr)
              toast.error('Failed to parse generated skill. Please try again.')
              setStep('input')
            }
          } else if (taskStatus.status === 'failed') {
            if (pollIntervalRef.current) {
              window.clearInterval(pollIntervalRef.current)
            }
            toast.error('Skill generation failed: ' + (taskStatus.error_message || 'Unknown error'))
            setStep('input')
          }
        } catch (err) {
          if (pollIntervalRef.current) {
            window.clearInterval(pollIntervalRef.current)
          }
          toast.error('Failed to check task status')
          setStep('input')
        }
      }, 3000)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to generate skill')
      setStep('input')
    }
  }

  const handleSave = async () => {
    if (!generatedSkill) return

    try {
      await api.createSkill(generatedSkill)
      toast.success('Skill created successfully!')
      navigate({ to: '/skills' })
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create skill')
    }
  }

  if (step === 'generating') {
    return (
      <div className="space-y-6">
        <div className="card p-12 text-center">
          <Loader2 className="w-16 h-16 text-emerald-400 animate-spin mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-primary mb-2">Generating Your Skill...</h3>
          <p className="text-secondary mb-4">
            The AI agent is creating your skill based on your requirements.
            This may take 1-3 minutes.
          </p>
          {taskId && (
            <p className="text-sm text-muted">
              Task ID: <code className="text-emerald-400">{taskId}</code>
            </p>
          )}
        </div>
      </div>
    )
  }

  if (step === 'preview' && generatedSkill) {
    return (
      <div className="space-y-6">
        <div className="card p-6 bg-emerald-500/10 border-emerald-500/30">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="w-5 h-5 text-emerald-400 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <h3 className="font-semibold text-emerald-400 mb-1">Skill Generated Successfully!</h3>
              <p className="text-sm text-secondary">
                Review the generated skill below and make any necessary edits before saving.
              </p>
            </div>
          </div>
        </div>

        <div className="card p-6">
          <h3 className="font-semibold text-primary mb-4">Generated Skill Preview</h3>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-secondary mb-1">Skill ID</label>
                <p className="text-primary font-mono text-sm">{generatedSkill.id}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-secondary mb-1">Name</label>
                <p className="text-primary">{generatedSkill.name}</p>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-1">Command</label>
              <p className="text-emerald-400 font-mono">{generatedSkill.command}</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-1">Description</label>
              <p className="text-secondary text-sm">{generatedSkill.description}</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">SKILL.md Instructions</label>
              <pre className="text-xs text-secondary bg-secondary p-4 rounded-lg overflow-x-auto max-h-[300px]">
                {generatedSkill.prompt}
              </pre>
            </div>

            {generatedSkill.files && generatedSkill.files.length > 0 && (
              <div>
                <label className="block text-sm font-medium text-secondary mb-2">
                  Bundled Files ({generatedSkill.files.length})
                </label>
                <div className="space-y-2">
                  {generatedSkill.files.map((file: any, idx: number) => (
                    <details key={idx} className="p-3 rounded-lg bg-secondary">
                      <summary className="font-mono text-sm text-emerald-400 cursor-pointer">
                        {file.path}
                      </summary>
                      <pre className="text-xs text-secondary mt-2 overflow-x-auto max-h-[200px]">
                        {file.content}
                      </pre>
                    </details>
                  ))}
                </div>
              </div>
            )}

            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <label className="block text-muted mb-1">Category</label>
                <p className="text-primary">{generatedSkill.category}</p>
              </div>
              <div>
                <label className="block text-muted mb-1">Version</label>
                <p className="text-primary">{generatedSkill.version}</p>
              </div>
              <div>
                <label className="block text-muted mb-1">Tags</label>
                <p className="text-primary">{generatedSkill.tags?.join(', ') || 'None'}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="flex gap-3">
          <button onClick={() => setStep('input')} className="btn btn-ghost">
            <ArrowLeft className="w-4 h-4" />
            Start Over
          </button>
          <button onClick={handleSave} className="btn btn-primary flex-1">
            <Save className="w-4 h-4" />
            Save Skill
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="card p-6 bg-purple-500/5 border-purple-500/20">
        <div className="flex items-start gap-3">
          <Sparkles className="w-5 h-5 text-purple-400 flex-shrink-0 mt-0.5" />
          <div>
            <h3 className="font-semibold text-primary mb-2">Generate a Skill with AI</h3>
            <p className="text-sm text-secondary mb-3">
              Describe what you want your skill to do, and an AI agent will generate a complete skill package for you,
              including SKILL.md, scripts, references, and assets.
            </p>
            <p className="text-sm text-secondary">
              This uses the <code className="text-emerald-400">skill-creator</code> workflow to ensure best practices.
            </p>
          </div>
        </div>
      </div>

      {runningSessions.length === 0 ? (
        <div className="card p-6 bg-amber-500/10 border-amber-500/20">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <h4 className="font-semibold text-amber-400 mb-1">No Running Sessions</h4>
              <p className="text-sm text-secondary mb-3">
                You need a running session to generate skills. Start a session first.
              </p>
              <button onClick={() => navigate({ to: '/profiles' })} className="btn btn-secondary btn-sm">
                Go to Profiles
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className="card p-6">
          <label className="block text-sm font-medium text-secondary mb-2">Select Session *</label>
          <select
            value={selectedSession}
            onChange={e => setSelectedSession(e.target.value)}
            className="input"
            required
          >
            <option value="">Choose a session...</option>
            {runningSessions.map(session => (
              <option key={session.id} value={session.id}>
                {session.profile_id ?? session.id.slice(0, 8)} ({session.id.slice(0, 8)})
              </option>
            ))}
          </select>
          <p className="text-xs text-muted mt-1">The selected agent will generate your skill</p>
        </div>
      )}

      <div className="card p-6">
        <label className="block text-sm font-medium text-secondary mb-2">Describe Your Skill Requirements *</label>
        <p className="text-xs text-muted mb-3">
          Be specific about what the skill should do, when it should be used, and any examples of usage.
        </p>
        <textarea
          value={requirements}
          onChange={e => setRequirements(e.target.value)}
          className="input min-h-[300px] font-mono text-sm"
          placeholder={`Example:

I need a skill for automated git commit messages:

**Functionality:**
- Analyze git diff --cached to see staged changes
- Generate commit message following Conventional Commits format
- Support both English and Chinese
- Include type (feat/fix/refactor/docs/test/chore)
- Brief description (50 chars max)
- Optional detailed body

**When to use:**
- When user says "commit" or "create commit"
- When user asks to "write commit message"

**Example usage:**
- "Create a commit for these changes"
- "Generate commit message"
- "Commit with proper message"`}
          required
        />
      </div>

      <button
        onClick={handleGenerate}
        className="btn btn-primary w-full"
        disabled={!requirements.trim() || !selectedSession}
      >
        <Sparkles className="w-4 h-4" />
        Generate Skill with AI
      </button>

      <div className="card p-6">
        <h4 className="font-semibold text-primary mb-3">💡 Tips for Better Results</h4>
        <ul className="space-y-2 text-sm text-secondary">
          <li className="flex items-start gap-2">
            <span className="text-emerald-400 mt-0.5">•</span>
            <span>Include specific examples of how the skill will be used</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-emerald-400 mt-0.5">•</span>
            <span>Describe what should trigger the skill (keywords, phrases)</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-emerald-400 mt-0.5">•</span>
            <span>Mention if you need scripts, references, or assets</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-emerald-400 mt-0.5">•</span>
            <span>Be clear about the expected output format</span>
          </li>
        </ul>
      </div>
    </div>
  )
}
