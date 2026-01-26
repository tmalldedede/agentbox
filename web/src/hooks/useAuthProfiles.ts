import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { AddAuthProfileRequest } from '@/types'
import { providersQueryKey } from './useProviders'

export const authProfilesQueryKey = (providerId: string) => [...providersQueryKey, providerId, 'profiles']
export const rotationStatsQueryKey = (providerId: string) => [...providersQueryKey, providerId, 'rotation-stats']

/**
 * Query auth profiles for a provider
 */
export function useAuthProfiles(providerId: string) {
  return useQuery({
    queryKey: authProfilesQueryKey(providerId),
    queryFn: () => api.listAuthProfiles(providerId),
    enabled: !!providerId,
    staleTime: 1000 * 30, // 30 seconds
  })
}

/**
 * Query rotation stats for a provider
 */
export function useRotationStats(providerId: string) {
  return useQuery({
    queryKey: rotationStatsQueryKey(providerId),
    queryFn: () => api.getRotationStats(providerId),
    enabled: !!providerId,
    staleTime: 1000 * 10, // 10 seconds (more frequently updated)
    refetchInterval: 1000 * 30, // Auto-refresh every 30 seconds
  })
}

/**
 * Add an auth profile to a provider
 */
export function useAddAuthProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ providerId, req }: { providerId: string; req: AddAuthProfileRequest }) =>
      api.addAuthProfile(providerId, req),
    onSuccess: (_, { providerId }) => {
      queryClient.invalidateQueries({ queryKey: authProfilesQueryKey(providerId) })
      queryClient.invalidateQueries({ queryKey: rotationStatsQueryKey(providerId) })
      toast.success('API key added successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to add API key: ${error.message}`)
    },
  })
}

/**
 * Remove an auth profile from a provider
 */
export function useRemoveAuthProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ providerId, profileId }: { providerId: string; profileId: string }) =>
      api.removeAuthProfile(providerId, profileId),
    onSuccess: (_, { providerId }) => {
      queryClient.invalidateQueries({ queryKey: authProfilesQueryKey(providerId) })
      queryClient.invalidateQueries({ queryKey: rotationStatsQueryKey(providerId) })
      toast.success('API key removed')
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove API key: ${error.message}`)
    },
  })
}
