import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有已上传的文件
 */
export function useFiles() {
  return useQuery({
    queryKey: ['files'],
    queryFn: () => api.listFiles(),
    staleTime: 1000 * 10,
  })
}

/**
 * 上传文件
 */
export function useUploadFile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (file: File) => api.uploadFile(file),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['files'] })
      toast.success(`文件 ${data.name} 上传成功`)
    },
    onError: error => {
      toast.error(`上传失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除文件
 */
export function useDeleteFile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (fileId: string) => api.deleteFile(fileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files'] })
      toast.success('文件已删除')
    },
    onError: error => {
      toast.error(`删除失败: ${getErrorMessage(error)}`)
    },
  })
}
