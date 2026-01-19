import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api } from '../services/api'
import type { CreateMCPServerRequest, UpdateMCPServerRequest } from '../types'
import { getErrorMessage } from '../lib/errors'

/**
 * 查询所有 MCP Servers
 */
export function useMCPServers() {
  return useQuery({
    queryKey: ['mcp-servers'],
    queryFn: () => api.listMCPServers(),
    staleTime: 1000 * 60,
  })
}

/**
 * 查询单个 MCP Server
 */
export function useMCPServer(serverId: string | undefined) {
  return useQuery({
    queryKey: ['mcp-server', serverId],
    queryFn: () => api.getMCPServer(serverId!),
    enabled: !!serverId,
    staleTime: 1000 * 30,
  })
}

/**
 * 创建 MCP Server
 */
export function useCreateMCPServer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateMCPServerRequest) => api.createMCPServer(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      toast.success('MCP Server 创建成功')
    },
    onError: error => {
      toast.error(`创建 MCP Server 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 更新 MCP Server
 */
export function useUpdateMCPServer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateMCPServerRequest }) =>
      api.updateMCPServer(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['mcp-server', variables.id] })
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      toast.success('MCP Server 更新成功')
    },
    onError: error => {
      toast.error(`更新 MCP Server 失败: ${getErrorMessage(error)}`)
    },
  })
}

/**
 * 删除 MCP Server
 */
export function useDeleteMCPServer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (serverId: string) => api.deleteMCPServer(serverId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      toast.success('MCP Server 已删除')
    },
    onError: error => {
      toast.error(`删除 MCP Server 失败: ${getErrorMessage(error)}`)
    },
  })
}
