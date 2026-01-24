import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'
import type { CreateBatchRequest, ListBatchFilter, ListBatchTaskFilter } from '../types'

// List batches
export function useBatches(filter?: ListBatchFilter) {
  return useQuery({
    queryKey: ['batches', filter],
    queryFn: () => api.listBatches(filter),
    refetchInterval: 5000, // Auto refresh every 5s
  })
}

// Get single batch
export function useBatch(id: string) {
  return useQuery({
    queryKey: ['batch', id],
    queryFn: () => api.getBatch(id),
    enabled: !!id,
    refetchInterval: (query) => {
      // Auto refresh if running
      const status = query.state.data?.status
      if (status === 'running' || status === 'paused') {
        return 2000
      }
      return false
    },
  })
}

// Create batch
export function useCreateBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateBatchRequest) => api.createBatch(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
    },
  })
}

// Delete batch
export function useDeleteBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteBatch(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
    },
  })
}

// Start batch
export function useStartBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.startBatch(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', data.id] })
    },
  })
}

// Pause batch
export function usePauseBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.pauseBatch(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', data.id] })
    },
  })
}

// Resume batch
export function useResumeBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.resumeBatch(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', data.id] })
    },
  })
}

// Cancel batch
export function useCancelBatch() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.cancelBatch(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', data.id] })
    },
  })
}

// Retry failed tasks
export function useRetryBatchFailed() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.retryBatchFailed(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', data.id] })
    },
  })
}

// Get batch stats
export function useBatchStats(id: string) {
  return useQuery({
    queryKey: ['batch', id, 'stats'],
    queryFn: () => api.getBatchStats(id),
    enabled: !!id,
  })
}

// List batch tasks
export function useBatchTasks(batchId: string, filter?: ListBatchTaskFilter) {
  return useQuery({
    queryKey: ['batch', batchId, 'tasks', filter],
    queryFn: () => api.listBatchTasks(batchId, filter),
    enabled: !!batchId,
  })
}

// Get single batch task
export function useBatchTask(batchId: string, taskId: string) {
  return useQuery({
    queryKey: ['batch', batchId, 'task', taskId],
    queryFn: () => api.getBatchTask(batchId, taskId),
    enabled: !!batchId && !!taskId,
  })
}

// Get dead letter tasks
export function useBatchDeadTasks(batchId: string, limit = 100) {
  return useQuery({
    queryKey: ['batch', batchId, 'dead', limit],
    queryFn: () => api.listBatchDeadTasks(batchId, limit),
    enabled: !!batchId,
  })
}

// Retry dead letter tasks
export function useRetryBatchDead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ batchId, taskIds }: { batchId: string; taskIds?: string[] }) =>
      api.retryBatchDead(batchId, taskIds),
    onSuccess: (_, { batchId }) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] })
      queryClient.invalidateQueries({ queryKey: ['batch', batchId] })
      queryClient.invalidateQueries({ queryKey: ['batch', batchId, 'dead'] })
      queryClient.invalidateQueries({ queryKey: ['batch', batchId, 'stats'] })
    },
  })
}
