import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有镜像
 */
export function useImages(options?: { agentOnly?: boolean }) {
  return useQuery({
    queryKey: ['images', options],
    queryFn: () => api.listImages(options || {}),
    staleTime: 1000 * 30, // 30秒
    select: data => {
      // Sort: agent images first, then by created time
      return [...data].sort((a, b) => {
        if (a.is_agent_image !== b.is_agent_image) {
          return a.is_agent_image ? -1 : 1
        }
        return b.created - a.created
      })
    },
  })
}

/**
 * 拉取镜像
 */
export function usePullImage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (imageName: string) => api.pullImage({ image: imageName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('镜像拉取成功')
    },
    onError: error => {
      toast.error(`拉取镜像失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除镜像
 */
export function useRemoveImage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (imageId: string) => api.removeImage(imageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('镜像已删除')
    },
    onError: error => {
      toast.error(`删除镜像失败: ${getErrorMessage(error)}`)
    },
  })
}
