import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'

/**
 * 查询所有 Agent
 */
export function useAgents() {
  return useQuery({
    queryKey: ['agents'],
    queryFn: api.listAgents,
    staleTime: 1000 * 60 * 5, // 5 分钟（Agent 列表很少变化）
  })
}
