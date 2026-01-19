import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateSessionRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有会话
 */
export function useSessions() {
  return useQuery({
    queryKey: ['sessions'],
    queryFn: api.listSessions,
    staleTime: 1000 * 10, // 10 秒内数据被认为是新鲜的
    refetchInterval: 5000, // 每 5 秒自动刷新
  })
}

/**
 * 查询单个会话
 */
export function useSession(sessionId: string | undefined) {
  return useQuery({
    queryKey: ['session', sessionId],
    queryFn: () => api.getSession(sessionId!),
    enabled: !!sessionId, // 只有当 sessionId 存在时才执行查询
    staleTime: 1000 * 5,
    refetchInterval: 5000,
  })
}

/**
 * 创建会话
 */
export function useCreateSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: CreateSessionRequest) => api.createSession(request),
    onSuccess: () => {
      // 使会话列表缓存失效，触发重新获取
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('会话创建成功')
    },
    onError: error => {
      toast.error(`创建会话失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除会话
 */
export function useDeleteSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => api.deleteSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('会话已删除')
    },
    onError: error => {
      toast.error(`删除会话失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 停止会话
 */
export function useStopSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => api.stopSession(sessionId),
    onSuccess: (_data, sessionId) => {
      // 使特定会话的缓存失效
      queryClient.invalidateQueries({ queryKey: ['session', sessionId] })
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('会话已停止')
    },
    onError: error => {
      toast.error(`停止会话失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 重启会话 (停止后重新启动)
 */
export function useRestartSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (sessionId: string) => {
      // Stop first, then start
      await api.stopSession(sessionId)
      return api.startSession(sessionId)
    },
    onSuccess: (_data, sessionId) => {
      queryClient.invalidateQueries({ queryKey: ['session', sessionId] })
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('会话已重启')
    },
    onError: error => {
      toast.error(`重启会话失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 启动会话
 */
export function useStartSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sessionId: string) => api.startSession(sessionId),
    onSuccess: (_data, sessionId) => {
      queryClient.invalidateQueries({ queryKey: ['session', sessionId] })
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('会话已启动')
    },
    onError: error => {
      toast.error(`启动会话失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 执行命令
 */
export function useExecSession(sessionId: string | undefined) {
  return useMutation({
    mutationFn: (prompt: string) => {
      if (!sessionId) throw new Error('Session ID is required')
      return api.execSession(sessionId, { prompt })
    },
    onError: error => {
      toast.error(`执行失败: ${getErrorMessage(error)}`)
    },
  })
}
