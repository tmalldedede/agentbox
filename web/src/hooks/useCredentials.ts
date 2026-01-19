import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateCredentialRequest, UpdateCredentialRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Credentials
 */
export function useCredentials() {
  return useQuery({
    queryKey: ['credentials'],
    queryFn: () => api.listCredentials(),
    staleTime: 1000 * 60,
  })
}

/**
 * 查询单个 Credential
 */
export function useCredential(credentialId: string | undefined) {
  return useQuery({
    queryKey: ['credential', credentialId],
    queryFn: () => api.getCredential(credentialId!),
    enabled: !!credentialId,
    staleTime: 1000 * 30,
  })
}

/**
 * 创建 Credential
 */
export function useCreateCredential() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateCredentialRequest) => api.createCredential(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials'] })
      toast.success('Credential 创建成功')
    },
    onError: error => {
      toast.error(`创建 Credential 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 更新 Credential
 */
export function useUpdateCredential() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateCredentialRequest }) =>
      api.updateCredential(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['credential', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['credentials'] })
      toast.success('Credential 更新成功')
    },
    onError: error => {
      toast.error(`更新 Credential 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 Credential
 */
export function useDeleteCredential() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (credentialId: string) => api.deleteCredential(credentialId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials'] })
      toast.success('Credential 已删除')
    },
    onError: error => {
      toast.error(`删除 Credential 失败: ${getErrorMessage(error)}`)
    },
  })
}
