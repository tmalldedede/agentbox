import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateSmartAgentRequest, UpdateSmartAgentRequest, RunSmartAgentRequest } from '../types'

/**
 * Query key for smart agents
 */
export const smartAgentsQueryKey = ['smartAgents']

/**
 * Query all smart agents
 */
export function useSmartAgents() {
  return useQuery({
    queryKey: smartAgentsQueryKey,
    queryFn: api.listSmartAgents,
    staleTime: 1000 * 60, // 1 minute
  })
}

/**
 * Query a single smart agent by ID
 */
export function useSmartAgent(id: string) {
  return useQuery({
    queryKey: [...smartAgentsQueryKey, id],
    queryFn: () => api.getSmartAgent(id),
    enabled: !!id && id !== 'new',
  })
}

/**
 * Create a new smart agent
 */
export function useCreateSmartAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateSmartAgentRequest) => api.createSmartAgent(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: smartAgentsQueryKey })
      toast.success('Agent created successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to create agent: ${error.message}`)
    },
  })
}

/**
 * Update an existing smart agent
 */
export function useUpdateSmartAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateSmartAgentRequest }) =>
      api.updateSmartAgent(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: smartAgentsQueryKey })
      queryClient.invalidateQueries({ queryKey: [...smartAgentsQueryKey, id] })
      toast.success('Agent updated successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to update agent: ${error.message}`)
    },
  })
}

/**
 * Delete a smart agent
 */
export function useDeleteSmartAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteSmartAgent(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: smartAgentsQueryKey })
      toast.success('Agent deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete agent: ${error.message}`)
    },
  })
}

/**
 * Run a smart agent
 */
export function useRunSmartAgent() {
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: RunSmartAgentRequest }) =>
      api.runSmartAgent(id, req),
    onError: (error: Error) => {
      toast.error(`Failed to run agent: ${error.message}`)
    },
  })
}
