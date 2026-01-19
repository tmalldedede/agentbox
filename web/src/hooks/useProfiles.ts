import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateProfileRequest, CloneProfileRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Profile
 */
export function useProfiles() {
  return useQuery({
    queryKey: ['profiles'],
    queryFn: api.listProfiles,
    staleTime: 1000 * 60, // 1 分钟
  })
}

/**
 * 查询单个 Profile
 */
export function useProfile(profileId: string | undefined) {
  return useQuery({
    queryKey: ['profile', profileId],
    queryFn: () => api.getProfile(profileId!),
    enabled: !!profileId,
    staleTime: 1000 * 30,
  })
}

/**
 * 查询解析后的 Profile（包含继承）
 */
export function useProfileResolved(profileId: string | undefined) {
  return useQuery({
    queryKey: ['profile', profileId, 'resolved'],
    queryFn: () => api.getProfileResolved(profileId!),
    enabled: !!profileId,
    staleTime: 1000 * 30,
  })
}

/**
 * 创建 Profile
 */
export function useCreateProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: CreateProfileRequest) => api.createProfile(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      toast.success('Profile 创建成功')
    },
    onError: error => {
      toast.error(`创建 Profile 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 更新 Profile
 */
export function useUpdateProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: CreateProfileRequest }) =>
      api.updateProfile(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['profile', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      toast.success('Profile 更新成功')
    },
    onError: error => {
      toast.error(`更新 Profile 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 Profile
 */
export function useDeleteProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (profileId: string) => api.deleteProfile(profileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      toast.success('Profile 已删除')
    },
    onError: error => {
      toast.error(`删除 Profile 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 克隆 Profile
 */
export function useCloneProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, request }: { id: string; request: CloneProfileRequest }) =>
      api.cloneProfile(id, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['profiles'] })
      toast.success('Profile 克隆成功')
    },
    onError: error => {
      toast.error(`克隆 Profile 失败: ${getErrorMessage(error)}`)
    },
  })
}
