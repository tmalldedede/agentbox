import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateWebhookRequest, UpdateWebhookRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Webhooks
 */
export function useWebhooks() {
  return useQuery({
    queryKey: ['webhooks'],
    queryFn: api.listWebhooks,
    staleTime: 1000 * 60, // 1分钟
  })
}

/**
 * 查询单个 Webhook
 */
export function useWebhook(webhookId: string | undefined) {
  return useQuery({
    queryKey: ['webhook', webhookId],
    queryFn: () => api.getWebhook(webhookId!),
    enabled: !!webhookId,
    staleTime: 1000 * 30,
  })
}

/**
 * 创建 Webhook
 */
export function useCreateWebhook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateWebhookRequest) => api.createWebhook(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      toast.success('Webhook 创建成功')
    },
    onError: error => {
      toast.error(`创建 Webhook 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 更新 Webhook
 */
export function useUpdateWebhook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWebhookRequest }) =>
      api.updateWebhook(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['webhook', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      toast.success('Webhook 更新成功')
    },
    onError: error => {
      toast.error(`更新 Webhook 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 Webhook
 */
export function useDeleteWebhook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (webhookId: string) => api.deleteWebhook(webhookId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      toast.success('Webhook 已删除')
    },
    onError: error => {
      toast.error(`删除 Webhook 失败: ${getErrorMessage(error)}`)
    },
  })
}
