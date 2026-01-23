import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateAgentRequest, UpdateAgentRequest, RunAgentRequest } from '../types'

export const agentsQueryKey = ['agents']

/**
 * Query all agents
 */
export function useAgents() {
  return useQuery({
    queryKey: agentsQueryKey,
    queryFn: api.listAgents,
    staleTime: 1000 * 60,
  })
}

/**
 * Query a single agent by ID
 */
export function useAgent(id: string) {
  return useQuery({
    queryKey: [...agentsQueryKey, id],
    queryFn: () => api.getAgent(id),
    enabled: !!id && id !== 'new',
  })
}

/**
 * Create a new agent
 */
export function useCreateAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateAgentRequest) => api.createAgent(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentsQueryKey })
      toast.success('Agent created successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to create agent: ${error.message}`)
    },
  })
}

/**
 * Update an existing agent
 */
export function useUpdateAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateAgentRequest }) =>
      api.updateAgent(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: agentsQueryKey })
      queryClient.invalidateQueries({ queryKey: [...agentsQueryKey, id] })
      toast.success('Agent updated successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to update agent: ${error.message}`)
    },
  })
}

/**
 * Delete an agent
 */
export function useDeleteAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteAgent(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: agentsQueryKey })
      toast.success('Agent deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete agent: ${error.message}`)
    },
  })
}

/**
 * Run an agent
 */
export function useRunAgent() {
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: RunAgentRequest }) =>
      api.runAgent(id, req),
    onError: (error: Error) => {
      toast.error(`Failed to run agent: ${error.message}`)
    },
  })
}
