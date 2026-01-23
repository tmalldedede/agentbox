import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateRuntimeRequest, UpdateRuntimeRequest } from '../types'

export const runtimesQueryKey = ['runtimes']

/**
 * Query all runtimes
 */
export function useRuntimes() {
  return useQuery({
    queryKey: runtimesQueryKey,
    queryFn: api.listRuntimes,
    staleTime: 1000 * 60 * 5,
  })
}

/**
 * Query a single runtime by ID
 */
export function useRuntime(id: string) {
  return useQuery({
    queryKey: [...runtimesQueryKey, id],
    queryFn: () => api.getRuntime(id),
    enabled: !!id && id !== 'new',
  })
}

/**
 * Create a new runtime
 */
export function useCreateRuntime() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateRuntimeRequest) => api.createRuntime(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: runtimesQueryKey })
      toast.success('Runtime created successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to create runtime: ${error.message}`)
    },
  })
}

/**
 * Update an existing runtime
 */
export function useUpdateRuntime() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateRuntimeRequest }) =>
      api.updateRuntime(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: runtimesQueryKey })
      queryClient.invalidateQueries({ queryKey: [...runtimesQueryKey, id] })
      toast.success('Runtime updated successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to update runtime: ${error.message}`)
    },
  })
}

/**
 * Delete a runtime
 */
export function useDeleteRuntime() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteRuntime(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: runtimesQueryKey })
      toast.success('Runtime deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete runtime: ${error.message}`)
    },
  })
}

/**
 * Set a runtime as the default
 */
export function useSetDefaultRuntime() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.setDefaultRuntime(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: runtimesQueryKey })
      toast.success('Default runtime updated')
    },
    onError: (error: Error) => {
      toast.error(`Failed to set default runtime: ${error.message}`)
    },
  })
}
