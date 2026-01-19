import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateTaskRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Tasks
 */
export function useTasks(options?: { limit?: number; status?: string }) {
  return useQuery({
    queryKey: ['tasks', options],
    queryFn: () => api.listTasks(options || { limit: 100 }),
    staleTime: 1000 * 5, // 5秒
    refetchInterval: 5000, // 自动轮询
  })
}

/**
 * 查询单个 Task
 */
export function useTask(taskId: string | undefined) {
  return useQuery({
    queryKey: ['task', taskId],
    queryFn: () => api.getTask(taskId!),
    enabled: !!taskId,
    staleTime: 1000 * 5,
    refetchInterval: (query) => {
      const task = query.state.data
      // 如果任务还在运行，继续轮询
      if (task && ['running', 'queued', 'pending'].includes(task.status)) {
        return 3000
      }
      return false
    },
  })
}

/**
 * 创建 Task
 */
export function useCreateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateTaskRequest) => api.createTask(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      toast.success('任务创建成功')
    },
    onError: error => {
      toast.error(`创建任务失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 取消 Task
 */
export function useCancelTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (taskId: string) => api.cancelTask(taskId),
    onSuccess: (_data, taskId) => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      toast.success('任务已取消')
    },
    onError: error => {
      toast.error(`取消任务失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 获取 Task 日志
 */
export function useTaskLogs(taskId: string | undefined) {
  return useQuery({
    queryKey: ['taskLogs', taskId],
    queryFn: () => api.getTaskLogs(taskId!),
    enabled: !!taskId,
    staleTime: 1000 * 10, // 10秒
  })
}
