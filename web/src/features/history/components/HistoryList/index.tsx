import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  History,
  RefreshCw,
  AlertCircle,
  Loader2,
  Bot,
  Terminal,
  CheckCircle2,
  XCircle,
  Clock,
  Trash2,
  ChevronDown,
  ChevronRight,
  Zap,
} from 'lucide-react'
import { useHistory, useHistoryStats, useDeleteHistoryEntry } from '@/hooks'
import type { HistoryEntry, HistoryFilter, HistorySourceType, HistoryStatus } from '@/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'

export default function HistoryList() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<HistoryFilter>({ limit: 50 })
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const {
    data: historyData,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useHistory(filter)

  const { data: stats } = useHistoryStats()
  const deleteEntry = useDeleteHistoryEntry()

  const entries = historyData?.entries || []
  const total = historyData?.total || 0

  const handleDelete = (entry: HistoryEntry) => {
    if (!confirm(`Delete this execution record?`)) return
    deleteEntry.mutate(entry.id)
  }

  const getStatusBadge = (status: HistoryStatus) => {
    switch (status) {
      case 'completed':
        return (
          <Badge variant="default" className="bg-green-500">
            <CheckCircle2 className="w-3 h-3 mr-1" />
            Completed
          </Badge>
        )
      case 'failed':
        return (
          <Badge variant="destructive">
            <XCircle className="w-3 h-3 mr-1" />
            Failed
          </Badge>
        )
      case 'running':
        return (
          <Badge variant="default" className="bg-blue-500">
            <Loader2 className="w-3 h-3 mr-1 animate-spin" />
            Running
          </Badge>
        )
      case 'pending':
        return (
          <Badge variant="secondary">
            <Clock className="w-3 h-3 mr-1" />
            Pending
          </Badge>
        )
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getSourceIcon = (sourceType: HistorySourceType) => {
    switch (sourceType) {
      case 'agent':
        return <Bot className="w-4 h-4 text-blue-400" />
      case 'session':
        return <Terminal className="w-4 h-4 text-green-400" />
      default:
        return <Zap className="w-4 h-4 text-gray-400" />
    }
  }

  const formatDuration = (startedAt: string, endedAt?: string) => {
    const start = new Date(startedAt)
    const end = endedAt ? new Date(endedAt) : new Date()
    const ms = end.getTime() - start.getTime()
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    return `${(ms / 60000).toFixed(1)}m`
  }

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleString()
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <History className="w-6 h-6 text-amber-400" />
            <span className="text-lg font-bold">History</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="btn btn-ghost btn-icon"
            disabled={isFetching}
          >
            <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load history'}
            </span>
          </div>
        )}

        {/* Description & Stats */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">Execution History</h1>
          <p className="text-muted-foreground mb-4">
            View all past agent executions, task runs, and session activities in one place.
          </p>

          {stats && (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold">{stats.total_executions}</div>
                  <div className="text-sm text-muted-foreground">Total Executions</div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-green-500">{stats.completed_count}</div>
                  <div className="text-sm text-muted-foreground">Completed</div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-red-500">{stats.failed_count}</div>
                  <div className="text-sm text-muted-foreground">Failed</div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold">
                    {((stats.total_input_tokens + stats.total_output_tokens) / 1000).toFixed(1)}k
                  </div>
                  <div className="text-sm text-muted-foreground">Total Tokens</div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>

        {/* Filters */}
        <div className="flex gap-4 mb-6">
          <Select
            value={filter.source_type || 'all'}
            onValueChange={(value) => setFilter({
              ...filter,
              source_type: value === 'all' ? undefined : value as HistorySourceType
            })}
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Source" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Sources</SelectItem>
              <SelectItem value="agent">Agent</SelectItem>
              <SelectItem value="session">Session</SelectItem>
            </SelectContent>
          </Select>

          <Select
            value={filter.status || 'all'}
            onValueChange={(value) => setFilter({
              ...filter,
              status: value === 'all' ? undefined : value as HistoryStatus
            })}
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="failed">Failed</SelectItem>
              <SelectItem value="running">Running</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-amber-400 animate-spin" />
          </div>
        ) : entries.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <History className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg">No execution history</p>
            <p className="text-muted-foreground mt-2">
              Run an agent or execute a session to see history here
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            <div className="text-sm text-muted-foreground mb-2">
              Showing {entries.length} of {total} entries
            </div>

            {entries.map((entry) => (
              <Collapsible
                key={entry.id}
                open={expandedId === entry.id}
                onOpenChange={(open) => setExpandedId(open ? entry.id : null)}
              >
                <Card className="overflow-hidden">
                  <CollapsibleTrigger asChild>
                    <CardHeader className="cursor-pointer hover:bg-muted/50 py-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {expandedId === entry.id ? (
                            <ChevronDown className="w-4 h-4 text-muted-foreground" />
                          ) : (
                            <ChevronRight className="w-4 h-4 text-muted-foreground" />
                          )}
                          {getSourceIcon(entry.source_type)}
                          <div>
                            <CardTitle className="text-sm font-medium">
                              {entry.source_name}
                            </CardTitle>
                            <CardDescription className="text-xs">
                              {formatTime(entry.started_at)}
                              {entry.ended_at && (
                                <span className="ml-2 text-muted-foreground">
                                  ({formatDuration(entry.started_at, entry.ended_at)})
                                </span>
                              )}
                            </CardDescription>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {entry.usage && (
                            <Badge variant="outline" className="text-xs">
                              {entry.usage.input_tokens + entry.usage.output_tokens} tokens
                            </Badge>
                          )}
                          {getStatusBadge(entry.status)}
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDelete(entry)
                            }}
                          >
                            <Trash2 className="w-4 h-4 text-muted-foreground hover:text-red-500" />
                          </Button>
                        </div>
                      </div>
                    </CardHeader>
                  </CollapsibleTrigger>

                  <CollapsibleContent>
                    <CardContent className="pt-0 pb-4 space-y-4">
                      {/* Prompt */}
                      <div>
                        <div className="text-xs font-medium text-muted-foreground mb-1">Prompt</div>
                        <div className="bg-muted p-3 rounded-md text-sm font-mono whitespace-pre-wrap max-h-32 overflow-auto">
                          {entry.prompt}
                        </div>
                      </div>

                      {/* Output */}
                      {entry.output && (
                        <div>
                          <div className="text-xs font-medium text-muted-foreground mb-1">Output</div>
                          <div className="bg-muted p-3 rounded-md text-sm font-mono whitespace-pre-wrap max-h-64 overflow-auto">
                            {entry.output}
                          </div>
                        </div>
                      )}

                      {/* Error */}
                      {entry.error && (
                        <div>
                          <div className="text-xs font-medium text-red-400 mb-1">Error</div>
                          <div className="bg-red-500/10 border border-red-500/20 p-3 rounded-md text-sm text-red-400">
                            {entry.error}
                          </div>
                        </div>
                      )}

                      {/* Metadata */}
                      <div className="flex flex-wrap gap-2 text-xs">
                        {entry.engine && (
                          <Badge variant="outline">Engine: {entry.engine}</Badge>
                        )}
                        {entry.source_name && (
                          <Badge variant="outline">Agent: {entry.source_name}</Badge>
                        )}
                        <Badge variant="outline">Source: {entry.source_type}</Badge>
                        <Badge variant="outline">ID: {entry.id.slice(0, 8)}...</Badge>
                      </div>
                    </CardContent>
                  </CollapsibleContent>
                </Card>
              </Collapsible>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
