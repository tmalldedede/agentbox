import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { oauthAPI, type OAuthSyncStatus, type SyncResponse } from '@/services/oauth'
import { toast } from 'sonner'

/**
 * 获取 OAuth 同步状态
 */
export function useOAuthSyncStatus() {
  return useQuery<OAuthSyncStatus>({
    queryKey: ['oauth', 'sync-status'],
    queryFn: () => oauthAPI.getSyncStatus(),
    refetchInterval: 10000, // 每 10 秒刷新一次
  })
}

/**
 * 从 Claude CLI 同步
 */
export function useSyncFromClaudeCli() {
  const queryClient = useQueryClient()

  return useMutation<SyncResponse, Error, string>({
    mutationFn: (providerId: string) => oauthAPI.syncFromClaudeCli(providerId),
    onSuccess: (data) => {
      toast.success('OAuth 令牌同步成功', {
        description: `已从 Claude Code CLI 导入令牌，过期时间：${new Date(
          data.expires_at || ''
        ).toLocaleString()}`,
      })
      // 刷新 auth profiles 列表
      queryClient.invalidateQueries({ queryKey: ['authProfiles'] })
      queryClient.invalidateQueries({ queryKey: ['oauth', 'sync-status'] })
    },
    onError: (error) => {
      toast.error('同步失败', {
        description: error.message || '无法从 Claude Code CLI 读取令牌',
      })
    },
  })
}

/**
 * 从 Codex CLI 同步
 */
export function useSyncFromCodexCli() {
  const queryClient = useQueryClient()

  return useMutation<SyncResponse, Error, string>({
    mutationFn: (providerId: string) => oauthAPI.syncFromCodexCli(providerId),
    onSuccess: (data) => {
      toast.success('OAuth 令牌同步成功', {
        description: `已从 Codex CLI 导入令牌，过期时间：${new Date(
          data.expires_at || ''
        ).toLocaleString()}`,
      })
      queryClient.invalidateQueries({ queryKey: ['authProfiles'] })
      queryClient.invalidateQueries({ queryKey: ['oauth', 'sync-status'] })
    },
    onError: (error) => {
      toast.error('同步失败', {
        description: error.message || '无法从 Codex CLI 读取令牌',
      })
    },
  })
}

/**
 * 同步到 Claude CLI
 */
export function useSyncToClaudeCli() {
  const queryClient = useQueryClient()

  return useMutation<SyncResponse, Error, string>({
    mutationFn: (providerId: string) => oauthAPI.syncToClaudeCli(providerId),
    onSuccess: (data) => {
      toast.success('OAuth 令牌已导出', {
        description: `已将刷新后的令牌写入 Claude Code CLI，过期时间：${new Date(
          data.expires_at || ''
        ).toLocaleString()}`,
      })
      queryClient.invalidateQueries({ queryKey: ['oauth', 'sync-status'] })
    },
    onError: (error) => {
      toast.error('导出失败', {
        description: error.message || '无法写入令牌到 Claude Code CLI',
      })
    },
  })
}
