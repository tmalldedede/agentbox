import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'
import type {
  Settings,
  AgentSettings,
  TaskSettings,
  BatchSettings,
  StorageSettings,
  NotifySettings,
} from '../types'

// Get all settings
export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  })
}

// Update all settings
export function useUpdateSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: Settings) => api.updateSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Reset settings to defaults
export function useResetSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.resetSettings(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Agent settings
export function useAgentSettings() {
  return useQuery({
    queryKey: ['settings', 'agent'],
    queryFn: () => api.getAgentSettings(),
  })
}

export function useUpdateAgentSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: AgentSettings) => api.updateAgentSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Task settings
export function useTaskSettings() {
  return useQuery({
    queryKey: ['settings', 'task'],
    queryFn: () => api.getTaskSettings(),
  })
}

export function useUpdateTaskSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: TaskSettings) => api.updateTaskSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Batch settings
export function useBatchSettings() {
  return useQuery({
    queryKey: ['settings', 'batch'],
    queryFn: () => api.getBatchSettings(),
  })
}

export function useUpdateBatchSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: BatchSettings) => api.updateBatchSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Storage settings
export function useStorageSettings() {
  return useQuery({
    queryKey: ['settings', 'storage'],
    queryFn: () => api.getStorageSettings(),
  })
}

export function useUpdateStorageSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: StorageSettings) => api.updateStorageSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}

// Notify settings
export function useNotifySettings() {
  return useQuery({
    queryKey: ['settings', 'notify'],
    queryFn: () => api.getNotifySettings(),
  })
}

export function useUpdateNotifySettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (settings: NotifySettings) => api.updateNotifySettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}
