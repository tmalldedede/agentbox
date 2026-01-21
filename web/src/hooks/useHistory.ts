import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/services/api'
import type { HistoryFilter } from '@/types'

// Query keys
export const historyKeys = {
  all: ['history'] as const,
  lists: () => [...historyKeys.all, 'list'] as const,
  list: (filter?: HistoryFilter) => [...historyKeys.lists(), filter] as const,
  details: () => [...historyKeys.all, 'detail'] as const,
  detail: (id: string) => [...historyKeys.details(), id] as const,
  stats: (filter?: HistoryFilter) => [...historyKeys.all, 'stats', filter] as const,
}

// List history entries
export function useHistory(filter?: HistoryFilter) {
  return useQuery({
    queryKey: historyKeys.list(filter),
    queryFn: () => api.listHistory(filter),
  })
}

// Get single history entry
export function useHistoryEntry(id: string) {
  return useQuery({
    queryKey: historyKeys.detail(id),
    queryFn: () => api.getHistoryEntry(id),
    enabled: !!id,
  })
}

// Get history statistics
export function useHistoryStats(filter?: HistoryFilter) {
  return useQuery({
    queryKey: historyKeys.stats(filter),
    queryFn: () => api.getHistoryStats(filter),
  })
}

// Delete history entry
export function useDeleteHistoryEntry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => api.deleteHistoryEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: historyKeys.all })
    },
  })
}
