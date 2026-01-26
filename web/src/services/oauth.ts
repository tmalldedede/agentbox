import { request, API_BASE } from './api'

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
  getSyncStatus: () =>
    request<OAuthSyncStatus>(`${API_BASE}/oauth/sync-status`),

  // Sync from Claude Code CLI
  syncFromClaudeCli: (providerId: string) =>
    request<SyncResponse>(`${API_BASE}/oauth/sync-from-claude-cli`, {
      method: 'POST',
      body: JSON.stringify({ provider_id: providerId }),
    }),

  // Sync from Codex CLI
  syncFromCodexCli: (providerId: string) =>
    request<SyncResponse>(`${API_BASE}/oauth/sync-from-codex-cli`, {
      method: 'POST',
      body: JSON.stringify({ provider_id: providerId }),
    }),

  // Sync to Claude Code CLI
  syncToClaudeCli: (providerId: string) =>
    request<SyncResponse>(`${API_BASE}/oauth/sync-to-claude-cli/${providerId}`, {
      method: 'POST',
    }),
}
