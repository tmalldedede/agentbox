import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateProviderRequest, UpdateProviderRequest } from '@/types'

export const providersQueryKey = ['providers']

/**
 * Query all providers
 */
export function useProviders(options?: { agent?: string; category?: string; configured?: boolean }) {
  const query = useQuery({
    queryKey: [...providersQueryKey, options],
    queryFn: () => api.listProviders(options),
    staleTime: 1000 * 60 * 5,
  })

  return {
    ...query,
    data: query.data ?? [],
  }
}

/**
 * Query configured providers only
 */
export function useConfiguredProviders() {
  return useProviders({ configured: true })
}

/**
 * Query provider templates (for Add Provider dialog)
 */
export function useProviderTemplates() {
  const query = useQuery({
    queryKey: [...providersQueryKey, 'templates'],
    queryFn: () => api.listProviderTemplates(),
    staleTime: 1000 * 60 * 30,
  })

  return {
    ...query,
    data: query.data ?? [],
  }
}

/**
 * Query provider statistics
 */
export function useProviderStats() {
  return useQuery({
    queryKey: [...providersQueryKey, 'stats'],
    queryFn: () => api.getProviderStats(),
    staleTime: 1000 * 30,
  })
}

/**
 * Verify all configured providers
 */
export function useVerifyAllProviders() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.verifyAllProviders(),
    onSuccess: (results) => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      const valid = results.filter(r => r.valid).length
      const failed = results.filter(r => !r.valid).length
      if (failed === 0) {
        toast.success(`All ${valid} providers verified successfully`)
      } else {
        toast.warning(`${valid} passed, ${failed} failed`)
      }
    },
    onError: (error: Error) => {
      toast.error(`Verification failed: ${error.message}`)
    },
  })
}

/**
 * Query a single provider by ID
 */
export function useProvider(id: string) {
  return useQuery({
    queryKey: [...providersQueryKey, id],
    queryFn: () => api.getProvider(id),
    enabled: !!id,
  })
}

/**
 * Configure provider API key
 */
export function useConfigureProviderKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, apiKey }: { id: string; apiKey: string }) =>
      api.configureProviderKey(id, apiKey),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      queryClient.invalidateQueries({ queryKey: [...providersQueryKey, id] })
      toast.success('API key configured successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to configure API key: ${error.message}`)
    },
  })
}

/**
 * Verify provider API key
 */
export function useVerifyProviderKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.verifyProviderKey(id),
    onSuccess: (data, id) => {
      queryClient.invalidateQueries({ queryKey: [...providersQueryKey, id] })
      if (data.valid) {
        toast.success('API key is valid')
      } else {
        toast.error(`API key validation failed: ${data.message}`)
      }
    },
    onError: (error: Error) => {
      toast.error(`Failed to verify API key: ${error.message}`)
    },
  })
}

/**
 * Delete provider API key
 */
export function useDeleteProviderKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteProviderKey(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      queryClient.invalidateQueries({ queryKey: [...providersQueryKey, id] })
      toast.success('API key removed')
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove API key: ${error.message}`)
    },
  })
}

/**
 * Create a new provider
 */
export function useCreateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateProviderRequest) => api.createProvider(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      toast.success('Provider created successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to create provider: ${error.message}`)
    },
  })
}

/**
 * Update a provider
 */
export function useUpdateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateProviderRequest }) =>
      api.updateProvider(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      queryClient.invalidateQueries({ queryKey: [...providersQueryKey, id] })
      toast.success('Provider updated successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to update provider: ${error.message}`)
    },
  })
}

/**
 * Delete a provider
 */
export function useDeleteProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providersQueryKey })
      toast.success('Provider deleted successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete provider: ${error.message}`)
    },
  })
}

/**
 * Fetch models for a configured provider
 */
export function useFetchProviderModels() {
  return useMutation({
    mutationFn: (id: string) => api.fetchProviderModels(id),
    onError: (error: Error) => {
      toast.error(`Failed to fetch models: ${error.message}`)
    },
  })
}

/**
 * Probe models from a given API endpoint (for create flow)
 */
export function useProbeModels() {
  return useMutation({
    mutationFn: ({ baseURL, apiKey, agents }: { baseURL: string; apiKey: string; agents: string[] }) =>
      api.probeModels(baseURL, apiKey, agents),
    onError: (error: Error) => {
      toast.error(`Failed to fetch models: ${error.message}`)
    },
  })
}
