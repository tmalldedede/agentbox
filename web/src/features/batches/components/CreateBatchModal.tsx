import { useState } from 'react'
import { Loader2, Upload, FileJson } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAgents, useCreateBatch } from '@/hooks'
import type { CreateBatchRequest } from '@/types'

interface CreateBatchModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateBatchModal({ open, onOpenChange }: CreateBatchModalProps) {
  const [name, setName] = useState('')
  const [agentId, setAgentId] = useState('')
  const [promptTemplate, setPromptTemplate] = useState('')
  const [inputsJson, setInputsJson] = useState('')
  const [concurrency, setConcurrency] = useState('5')
  const [timeout, setTimeout] = useState('300')
  const [maxRetries, setMaxRetries] = useState('2')
  const [autoStart, setAutoStart] = useState(true)
  const [inputTab, setInputTab] = useState<'json' | 'csv'>('json')
  const [error, setError] = useState('')

  const { data: agentsData } = useAgents()
  const createBatch = useCreateBatch()
  const agents = agentsData || []

  const resetForm = () => {
    setName('')
    setAgentId('')
    setPromptTemplate('')
    setInputsJson('')
    setConcurrency('5')
    setTimeout('300')
    setMaxRetries('2')
    setAutoStart(true)
    setError('')
  }

  const handleSubmit = async () => {
    setError('')

    // Validate
    if (!agentId) {
      setError('Please select an agent')
      return
    }
    if (!promptTemplate) {
      setError('Please enter a prompt template')
      return
    }
    if (!inputsJson) {
      setError('Please provide inputs')
      return
    }

    // Parse inputs
    let inputs: Record<string, unknown>[]
    try {
      inputs = JSON.parse(inputsJson)
      if (!Array.isArray(inputs)) {
        throw new Error('Inputs must be an array')
      }
      if (inputs.length === 0) {
        throw new Error('Inputs array cannot be empty')
      }
    } catch (e) {
      setError(`Invalid JSON: ${(e as Error).message}`)
      return
    }

    const req: CreateBatchRequest = {
      name: name || `Batch ${new Date().toLocaleString()}`,
      agent_id: agentId,
      prompt_template: promptTemplate,
      inputs,
      concurrency: parseInt(concurrency, 10),
      timeout: parseInt(timeout, 10),
      max_retries: parseInt(maxRetries, 10),
      auto_start: autoStart,
    }

    try {
      await createBatch.mutateAsync(req)
      resetForm()
      onOpenChange(false)
    } catch (e) {
      setError((e as Error).message)
    }
  }

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (event) => {
      const content = event.target?.result as string
      if (file.name.endsWith('.csv')) {
        // Parse CSV to JSON
        try {
          const lines = content.split('\n').filter(l => l.trim())
          if (lines.length < 2) {
            setError('CSV must have header row and at least one data row')
            return
          }
          const headers = lines[0].split(',').map(h => h.trim())
          const data = lines.slice(1).map(line => {
            const values = line.split(',').map(v => v.trim())
            const obj: Record<string, string> = {}
            headers.forEach((h, i) => {
              obj[h] = values[i] || ''
            })
            return obj
          })
          setInputsJson(JSON.stringify(data, null, 2))
        } catch (e) {
          setError(`Failed to parse CSV: ${(e as Error).message}`)
        }
      } else {
        setInputsJson(content)
      }
    }
    reader.readAsText(file)
  }

  const exampleInputs = [
    { alert_id: 'A001', severity: 'high', source_ip: '192.168.1.100' },
    { alert_id: 'A002', severity: 'medium', source_ip: '10.0.0.50' },
  ]

  const exampleTemplate = `Analyze the following security alert:
- Alert ID: {{.alert_id}}
- Severity: {{.severity}}
- Source IP: {{.source_ip}}

Provide a brief analysis and recommended actions.`

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create Batch</DialogTitle>
          <DialogDescription>
            Create a batch job to process multiple tasks with worker pools.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Batch Name</Label>
            <Input
              id="name"
              placeholder="e.g., Alert Analysis Batch"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          {/* Agent */}
          <div className="space-y-2">
            <Label htmlFor="agent">Agent</Label>
            <Select value={agentId} onValueChange={setAgentId}>
              <SelectTrigger>
                <SelectValue placeholder="Select an agent" />
              </SelectTrigger>
              <SelectContent>
                {agents.map((agent) => (
                  <SelectItem key={agent.id} value={agent.id}>
                    {agent.name} ({agent.adapter})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Prompt Template */}
          <div className="space-y-2">
            <Label htmlFor="template">Prompt Template</Label>
            <Textarea
              id="template"
              placeholder="Enter prompt template with {{.field}} placeholders..."
              className="min-h-[120px] font-mono text-sm"
              value={promptTemplate}
              onChange={(e) => setPromptTemplate(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Use {'{{.field_name}}'} to reference input fields.
              <Button
                variant="link"
                size="sm"
                className="h-auto p-0 ml-2"
                onClick={() => setPromptTemplate(exampleTemplate)}
              >
                Load example
              </Button>
            </p>
          </div>

          {/* Inputs */}
          <div className="space-y-2">
            <Label>Inputs</Label>
            <Tabs value={inputTab} onValueChange={(v) => setInputTab(v as 'json' | 'csv')}>
              <div className="flex items-center justify-between">
                <TabsList>
                  <TabsTrigger value="json">JSON</TabsTrigger>
                  <TabsTrigger value="csv">CSV Upload</TabsTrigger>
                </TabsList>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setInputsJson(JSON.stringify(exampleInputs, null, 2))}
                  >
                    <FileJson className="h-4 w-4 mr-1" />
                    Load Example
                  </Button>
                </div>
              </div>
              <TabsContent value="json" className="mt-2">
                <Textarea
                  placeholder='[{"field1": "value1"}, {"field1": "value2"}]'
                  className="min-h-[150px] font-mono text-sm"
                  value={inputsJson}
                  onChange={(e) => setInputsJson(e.target.value)}
                />
              </TabsContent>
              <TabsContent value="csv" className="mt-2">
                <div className="border-2 border-dashed rounded-lg p-6 text-center">
                  <Upload className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                  <p className="text-sm mb-2">Upload a CSV or JSON file</p>
                  <Input
                    type="file"
                    accept=".csv,.json"
                    className="max-w-xs mx-auto"
                    onChange={handleFileUpload}
                  />
                </div>
                {inputsJson && (
                  <div className="mt-2">
                    <p className="text-sm text-muted-foreground mb-1">Parsed data:</p>
                    <pre className="text-xs bg-muted p-2 rounded max-h-32 overflow-auto">
                      {inputsJson}
                    </pre>
                  </div>
                )}
              </TabsContent>
            </Tabs>
          </div>

          {/* Configuration */}
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="concurrency">Concurrency</Label>
              <Input
                id="concurrency"
                type="number"
                min="1"
                max="20"
                value={concurrency}
                onChange={(e) => setConcurrency(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Worker count</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="timeout">Timeout (s)</Label>
              <Input
                id="timeout"
                type="number"
                min="10"
                max="3600"
                value={timeout}
                onChange={(e) => setTimeout(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Per task</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="retries">Max Retries</Label>
              <Input
                id="retries"
                type="number"
                min="0"
                max="5"
                value={maxRetries}
                onChange={(e) => setMaxRetries(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">On failure</p>
            </div>
          </div>

          {/* Auto-start */}
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="autoStart"
              checked={autoStart}
              onChange={(e) => setAutoStart(e.target.checked)}
              className="rounded border-gray-300"
            />
            <Label htmlFor="autoStart" className="text-sm font-normal">
              Start immediately after creation
            </Label>
          </div>

          {error && (
            <div className="text-sm text-red-500 bg-red-50 dark:bg-red-950 p-3 rounded">
              {error}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={createBatch.isPending}>
            {createBatch.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Create Batch
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
