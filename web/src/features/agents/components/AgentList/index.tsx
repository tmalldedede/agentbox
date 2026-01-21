import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowLeft,
  Plus,
  Bot,
  RefreshCw,
  AlertCircle,
  Loader2,
  MoreVertical,
  Edit,
  Trash2,
} from 'lucide-react'
import type { SmartAgent } from '@/types'
import { useSmartAgents, useDeleteSmartAgent } from '@/hooks'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export default function AgentList() {
  const navigate = useNavigate()
  const [deletingId, setDeletingId] = useState<string | undefined>()

  const {
    data: agents = [],
    isLoading,
    isFetching,
    error,
    refetch,
  } = useSmartAgents()

  const deleteAgent = useDeleteSmartAgent()

  const handleDelete = (agent: SmartAgent) => {
    if (!confirm(`Delete agent "${agent.name}"?`)) return
    setDeletingId(agent.id)
    deleteAgent.mutate(agent.id, {
      onSettled: () => setDeletingId(undefined),
    })
  }

  const handleClick = (agent: SmartAgent) => {
    navigate({ to: `/agents/${agent.id}` })
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge variant="default" className="bg-green-500">Active</Badge>
      case 'inactive':
        return <Badge variant="secondary">Inactive</Badge>
      case 'error':
        return <Badge variant="destructive">Error</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getAccessBadge = (access: string) => {
    switch (access) {
      case 'public':
        return <Badge variant="outline" className="text-green-600">Public</Badge>
      case 'api_key':
        return <Badge variant="outline" className="text-blue-600">API Key</Badge>
      case 'private':
        return <Badge variant="outline" className="text-gray-600">Private</Badge>
      default:
        return <Badge variant="outline">{access}</Badge>
    }
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
            <Bot className="w-6 h-6 text-blue-400" />
            <span className="text-lg font-bold">Agents</span>
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
          <button className="btn btn-primary" onClick={() => navigate({ to: '/agents/new' })}>
            <Plus className="w-4 h-4" />
            New Agent
          </button>
        </div>
      </header>

      <div className="p-6">
        {/* Error */}
        {error && (
          <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-400" />
            <span className="text-red-400">
              {error instanceof Error ? error.message : 'Failed to load agents'}
            </span>
          </div>
        )}

        {/* Description */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">Smart Agents</h1>
          <p className="text-muted-foreground">
            Agents are AI assistants that combine a Profile with system prompts and environment
            variables. Each agent exposes an API endpoint for external invocation.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
          </div>
        ) : agents.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center">
            <Bot className="w-16 h-16 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-lg">No agents found</p>
            <p className="text-muted-foreground mt-2">Create your first agent to get started</p>
            <Button className="mt-4" onClick={() => navigate({ to: '/agents/new' })}>
              <Plus className="w-4 h-4 mr-2" />
              Create Agent
            </Button>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {agents.map((agent) => (
              <Card
                key={agent.id}
                className="cursor-pointer hover:border-primary/50 transition-colors"
                onClick={() => handleClick(agent)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                        {agent.icon ? (
                          <span className="text-xl">{agent.icon}</span>
                        ) : (
                          <Bot className="w-5 h-5 text-blue-400" />
                        )}
                      </div>
                      <div>
                        <CardTitle className="text-base">{agent.name}</CardTitle>
                        <p className="text-xs text-muted-foreground font-mono">{agent.id}</p>
                      </div>
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={(e) => {
                          e.stopPropagation()
                          navigate({ to: `/agents/${agent.id}` })
                        }}>
                          <Edit className="w-4 h-4 mr-2" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleDelete(agent)
                          }}
                          disabled={deletingId === agent.id}
                        >
                          <Trash2 className="w-4 h-4 mr-2" />
                          {deletingId === agent.id ? 'Deleting...' : 'Delete'}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardHeader>
                <CardContent>
                  {agent.description && (
                    <CardDescription className="mb-3 line-clamp-2">
                      {agent.description}
                    </CardDescription>
                  )}
                  <div className="flex items-center gap-2 flex-wrap">
                    {getStatusBadge(agent.status)}
                    {getAccessBadge(agent.api_access)}
                    <Badge variant="outline" className="text-xs">
                      Profile: {agent.profile_id}
                    </Badge>
                  </div>
                  {agent.rate_limit && (
                    <p className="text-xs text-muted-foreground mt-2">
                      Rate limit: {agent.rate_limit} req/min
                    </p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
