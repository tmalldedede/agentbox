import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateSkillRequest, UpdateSkillRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 Skills
 */
export function useSkills() {
  return useQuery({
    queryKey: ['skills'],
    queryFn: () => api.listSkills(),
    staleTime: 1000 * 60,
  })
}

/**
 * 查询单个 Skill
 */
export function useSkill(skillId: string | undefined) {
  return useQuery({
    queryKey: ['skill', skillId],
    queryFn: () => api.getSkill(skillId!),
    enabled: !!skillId,
    staleTime: 1000 * 30,
  })
}

/**
 * 创建 Skill
 */
export function useCreateSkill() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateSkillRequest) => api.createSkill(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
      toast.success('Skill 创建成功')
    },
    onError: error => {
      toast.error(`创建 Skill 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 更新 Skill
 */
export function useUpdateSkill() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSkillRequest }) => api.updateSkill(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['skill', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['skills'] })
      toast.success('Skill 更新成功')
    },
    onError: error => {
      toast.error(`更新 Skill 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 Skill
 */
export function useDeleteSkill() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (skillId: string) => api.deleteSkill(skillId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
      toast.success('Skill 已删除')
    },
    onError: error => {
      toast.error(`删除 Skill 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 克隆 Skill
 */
export function useCloneSkill() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, newId, newName }: { id: string; newId: string; newName: string }) =>
      api.cloneSkill(id, { new_id: newId, new_name: newName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
      toast.success('Skill 克隆成功')
    },
    onError: error => {
      toast.error(`克隆 Skill 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 导出 Skill (返回 markdown 内容)
 */
export function useExportSkill() {
  return useMutation({
    mutationFn: (skillId: string) => api.exportSkill(skillId),
    onError: error => {
      toast.error(`导出 Skill 失败: ${getErrorMessage(error)}`)
    },
  })
}
