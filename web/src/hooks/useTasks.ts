import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateTaskRequest, TaskEvent } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Tasks
 */
export function useTasks(options?: { limit?: number; status?: string; search?: string; agent_id?: string }) {
  return useQuery({
    queryKey: ['tasks', options],
    queryFn: () => api.listTasks(options || { limit: 100 }),
    staleTime: 1000 * 5, // 5秒
    refetchInterval: 5000, // 自动轮询
  })
}

/**
 * 获取任务统计
 */
export function useTaskStats() {
  return useQuery({
    queryKey: ['taskStats'],
    queryFn: () => api.getTaskStats(),
    staleTime: 1000 * 10, // 10秒
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
 * 重试 Task（从失败/取消的任务重新创建）
 */
export function useRetryTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (taskId: string) => api.retryTask(taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['taskStats'] })
      toast.success('任务已重试')
    },
    onError: error => {
      toast.error(`重试失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 Task
 */
export function useDeleteTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (taskId: string) => api.deleteTask(taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['taskStats'] })
      toast.success('任务已删除')
    },
    onError: error => {
      toast.error(`删除失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 批量清理 Tasks
 */
export function useCleanupTasks() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (options?: { before_days?: number; statuses?: string[] }) =>
      api.cleanupTasks(options),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['taskStats'] })
      toast.success(`已清理 ${data.deleted} 个任务`)
    },
    onError: error => {
      toast.error(`清理失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 获取 Task 输出
 */
export function useTaskOutput(taskId: string | undefined) {
  return useQuery({
    queryKey: ['taskOutput', taskId],
    queryFn: () => api.getTaskOutput(taskId!),
    enabled: !!taskId,
    staleTime: 1000 * 10, // 10秒
  })
}

/**
 * 追加 Turn（多轮对话）
 */
export function useAppendTurn() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ taskId, prompt }: { taskId: string; prompt: string }) =>
      api.createTask({ task_id: taskId, prompt }),
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: ['task', vars.taskId] })
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
    onError: (error) => {
      toast.error(`追加轮次失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * SSE 实时事件流
 */
export function useTaskEvents(taskId: string | undefined, enabled: boolean) {
  const [events, setEvents] = useState<TaskEvent[]>([])
  const queryClient = useQueryClient()

  const clearEvents = useCallback(() => setEvents([]), [])

  useEffect(() => {
    if (!enabled || !taskId) return

    const es = api.streamTaskEvents(taskId)

    es.onmessage = (e) => {
      try {
        const event: TaskEvent = JSON.parse(e.data)
        setEvents((prev) => [...prev, event])

        // 任务完成/失败时刷新数据
        if (event.type === 'task.completed' || event.type === 'task.failed') {
          queryClient.invalidateQueries({ queryKey: ['task', taskId] })
          queryClient.invalidateQueries({ queryKey: ['tasks'] })
        }
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      es.close()
    }

    return () => {
      es.close()
    }
  }, [taskId, enabled, queryClient])

  return { events, clearEvents }
}
