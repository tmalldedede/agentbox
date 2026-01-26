import { api } from './api'

export interface OAuthSyncStatus {
  claude_cli_available: boolean
  codex_cli_available: boolean
  last_sync_at?: string
  platform: string
}

export interface SyncResponse {
  success: boolean
  message: string
  expires_at?: string
}

export const oauthAPI = {
  // Get sync status
  getSyncStatus: async (): Promise<OAuthSyncStatus> => {
    const response = await api.get('/api/v1/oauth/sync-status')
    return response.data
  },

  // Sync from Claude Code CLI
  syncFromClaudeCli: async (providerId: string): Promise<SyncResponse> => {
    const response = await api.post('/api/v1/oauth/sync-from-claude-cli', {
      provider_id: providerId,
    })
    return response.data
  },

  // Sync from Codex CLI
  syncFromCodexCli: async (providerId: string): Promise<SyncResponse> => {
    const response = await api.post('/api/v1/oauth/sync-from-codex-cli', {
      provider_id: providerId,
    })
    return response.data
  },

  // Sync to Claude Code CLI
  syncToClaudeCli: async (providerId: string): Promise<SyncResponse> => {
    const response = await api.post(`/api/v1/oauth/sync-to-claude-cli/${providerId}`)
    return response.data
  },
}
