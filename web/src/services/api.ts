import type {
  Session,
  Agent,
  ApiResponse,
  CreateSessionRequest,
  ExecRequest,
  ExecResponse,
  Execution,
  Profile,
  CreateProfileRequest,
  CloneProfileRequest,
  Provider,
  CreateProviderRequest,
  UpdateProviderRequest,
  MCPServer,
  CreateMCPServerRequest,
  UpdateMCPServerRequest,
  CloneMCPServerRequest,
  Skill,
  CreateSkillRequest,
  UpdateSkillRequest,
  CloneSkillRequest,
  SkillSource,
  RemoteSkill,
  InstallSkillRequest,
  AddSourceRequest,
  Credential,
  CreateCredentialRequest,
  UpdateCredentialRequest,
  Image,
  PullImageRequest,
  SystemHealth,
  SystemStats,
  CleanupContainersResponse,
  CleanupImagesRequest,
  CleanupImagesResponse,
  Webhook,
  CreateWebhookRequest,
  UpdateWebhookRequest,
  Task,
  CreateTaskRequest,
} from '../types'

const API_BASE = '/api/v1'
const ADMIN_BASE = '/api/v1/admin'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  })

  const data: ApiResponse<T> = await response.json()

  if (data.code !== 0) {
    throw new Error(data.message)
  }

  return data.data as T
}

export const api = {
  // ==================== Public API ====================
  // 参考 Manus API 设计

  // Health
  health: () => request<{ status: string; version: string }>(`${API_BASE}/health`),

  // Agents (只读)
  listAgents: () => request<Agent[]>(`${API_BASE}/agents`),

  // Profiles (CRUD)
  listProfiles: () => request<Profile[]>(`${API_BASE}/profiles`),

  getProfile: (id: string) => request<Profile>(`${API_BASE}/profiles/${id}`),

  getProfileResolved: (id: string) => request<Profile>(`${API_BASE}/profiles/${id}/resolved`),

  createProfile: (req: CreateProfileRequest) =>
    request<Profile>(`${API_BASE}/profiles`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateProfile: (id: string, req: CreateProfileRequest) =>
    request<Profile>(`${API_BASE}/profiles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteProfile: (id: string) =>
    request<void>(`${API_BASE}/profiles/${id}`, {
      method: 'DELETE',
    }),

  cloneProfile: (id: string, req: CloneProfileRequest) =>
    request<Profile>(`${API_BASE}/profiles/${id}/clone`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  // Providers (CRUD) - API 提供商配置
  listProviders: (options?: { agent?: string; category?: string }) => {
    const params = new URLSearchParams()
    if (options?.agent) params.set('agent', options.agent)
    if (options?.category) params.set('category', options.category)
    const query = params.toString()
    return request<Provider[]>(`${API_BASE}/providers${query ? `?${query}` : ''}`)
  },

  getProvider: (id: string) => request<Provider>(`${API_BASE}/providers/${id}`),

  createProvider: (req: CreateProviderRequest) =>
    request<Provider>(`${API_BASE}/providers`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateProvider: (id: string, req: UpdateProviderRequest) =>
    request<Provider>(`${API_BASE}/providers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteProvider: (id: string) =>
    request<void>(`${API_BASE}/providers/${id}`, {
      method: 'DELETE',
    }),

  // Sessions (CRUD) - 类似 Manus Projects
  listSessions: () => request<Session[]>(`${API_BASE}/sessions`),

  getSession: (id: string) => request<Session>(`${API_BASE}/sessions/${id}`),

  createSession: (req: CreateSessionRequest) =>
    request<Session>(`${API_BASE}/sessions`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteSession: (id: string) =>
    request<{ deleted: string }>(`${API_BASE}/sessions/${id}`, {
      method: 'DELETE',
    }),

  startSession: (id: string) =>
    request<Session>(`${API_BASE}/sessions/${id}/start`, {
      method: 'POST',
    }),

  stopSession: (id: string) =>
    request<Session>(`${API_BASE}/sessions/${id}/stop`, {
      method: 'POST',
    }),

  execSession: (id: string, req: ExecRequest) =>
    request<ExecResponse>(`${API_BASE}/sessions/${id}/exec`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  getExecutions: (id: string) =>
    request<Execution[]>(`${API_BASE}/sessions/${id}/executions`),

  getExecution: (sessionId: string, execId: string) =>
    request<Execution>(`${API_BASE}/sessions/${sessionId}/executions/${execId}`),

  getSessionLogs: (id: string) =>
    request<{ logs: string }>(`${API_BASE}/sessions/${id}/logs`),

  // Tasks (CRUD) - 异步任务队列
  listTasks: (options?: { status?: string; profile_id?: string; limit?: number; offset?: number }) => {
    const params = new URLSearchParams()
    if (options?.status) params.set('status', options.status)
    if (options?.profile_id) params.set('profile_id', options.profile_id)
    if (options?.limit) params.set('limit', options.limit.toString())
    if (options?.offset) params.set('offset', options.offset.toString())
    const query = params.toString()
    return request<{ tasks: Task[]; total: number }>(`${API_BASE}/tasks${query ? `?${query}` : ''}`)
  },

  getTask: (id: string) => request<Task>(`${API_BASE}/tasks/${id}`),

  createTask: (req: CreateTaskRequest) =>
    request<Task>(`${API_BASE}/tasks`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  cancelTask: (id: string) =>
    request<Task>(`${API_BASE}/tasks/${id}`, {
      method: 'DELETE',
    }),

  getTaskOutput: (id: string) =>
    request<unknown>(`${API_BASE}/tasks/${id}/output`),

  getTaskLogs: (id: string) =>
    request<{ task_id: string; status: string; logs: string }>(`${API_BASE}/tasks/${id}/logs`),

  // Webhooks (CRUD)
  listWebhooks: () => request<Webhook[]>(`${API_BASE}/webhooks`),

  getWebhook: (id: string) => request<Webhook>(`${API_BASE}/webhooks/${id}`),

  createWebhook: (req: CreateWebhookRequest) =>
    request<Webhook>(`${API_BASE}/webhooks`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateWebhook: (id: string, req: UpdateWebhookRequest) =>
    request<Webhook>(`${API_BASE}/webhooks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteWebhook: (id: string) =>
    request<{ id: string; deleted: boolean }>(`${API_BASE}/webhooks/${id}`, {
      method: 'DELETE',
    }),

  // ==================== Admin API ====================
  // 平台管理接口

  // MCP Servers (管理接口)
  listMCPServers: (options?: { category?: string; enabled?: boolean }) => {
    const params = new URLSearchParams()
    if (options?.category) params.set('category', options.category)
    if (options?.enabled) params.set('enabled', 'true')
    const query = params.toString()
    return request<MCPServer[]>(`${ADMIN_BASE}/mcp-servers${query ? `?${query}` : ''}`)
  },

  getMCPServer: (id: string) => request<MCPServer>(`${ADMIN_BASE}/mcp-servers/${id}`),

  createMCPServer: (req: CreateMCPServerRequest) =>
    request<MCPServer>(`${ADMIN_BASE}/mcp-servers`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateMCPServer: (id: string, req: UpdateMCPServerRequest) =>
    request<MCPServer>(`${ADMIN_BASE}/mcp-servers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteMCPServer: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/mcp-servers/${id}`, {
      method: 'DELETE',
    }),

  cloneMCPServer: (id: string, req: CloneMCPServerRequest) =>
    request<MCPServer>(`${ADMIN_BASE}/mcp-servers/${id}/clone`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  testMCPServer: (id: string) =>
    request<{ status: string; message: string }>(`${ADMIN_BASE}/mcp-servers/${id}/test`, {
      method: 'POST',
    }),

  // Skills (管理接口)
  listSkills: (options?: { category?: string; enabled?: boolean }) => {
    const params = new URLSearchParams()
    if (options?.category) params.set('category', options.category)
    if (options?.enabled) params.set('enabled', 'true')
    const query = params.toString()
    return request<Skill[]>(`${ADMIN_BASE}/skills${query ? `?${query}` : ''}`)
  },

  getSkill: (id: string) => request<Skill>(`${ADMIN_BASE}/skills/${id}`),

  createSkill: (req: CreateSkillRequest) =>
    request<Skill>(`${ADMIN_BASE}/skills`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateSkill: (id: string, req: UpdateSkillRequest) =>
    request<Skill>(`${ADMIN_BASE}/skills/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteSkill: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/skills/${id}`, {
      method: 'DELETE',
    }),

  cloneSkill: (id: string, req: CloneSkillRequest) =>
    request<Skill>(`${ADMIN_BASE}/skills/${id}/clone`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  exportSkill: (id: string) =>
    fetch(`${ADMIN_BASE}/skills/${id}/export`).then(res => res.text()),

  // Skill Store API
  listSkillSources: () => request<SkillSource[]>(`${ADMIN_BASE}/skill-store/sources`),

  addSkillSource: (req: AddSourceRequest) =>
    request<SkillSource>(`${ADMIN_BASE}/skill-store/sources`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  removeSkillSource: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/skill-store/sources/${id}`, {
      method: 'DELETE',
    }),

  listRemoteSkills: () => request<RemoteSkill[]>(`${ADMIN_BASE}/skill-store/skills`),

  listSourceSkills: (sourceId: string) =>
    request<RemoteSkill[]>(`${ADMIN_BASE}/skill-store/skills/${sourceId}`),

  installSkill: (req: InstallSkillRequest) =>
    request<Skill>(`${ADMIN_BASE}/skill-store/install`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  uninstallSkill: (id: string) =>
    request<{ uninstalled: string }>(`${ADMIN_BASE}/skill-store/uninstall/${id}`, {
      method: 'DELETE',
    }),

  refreshSkillSource: (sourceId: string) =>
    request<{ refreshed: string }>(`${ADMIN_BASE}/skill-store/refresh/${sourceId}`, {
      method: 'POST',
    }),

  // Credentials (管理接口)
  listCredentials: (options?: { scope?: string; provider?: string }) => {
    const params = new URLSearchParams()
    if (options?.scope) params.set('scope', options.scope)
    if (options?.provider) params.set('provider', options.provider)
    const query = params.toString()
    return request<Credential[]>(`${ADMIN_BASE}/credentials${query ? `?${query}` : ''}`)
  },

  getCredential: (id: string) => request<Credential>(`${ADMIN_BASE}/credentials/${id}`),

  createCredential: (req: CreateCredentialRequest) =>
    request<Credential>(`${ADMIN_BASE}/credentials`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateCredential: (id: string, req: UpdateCredentialRequest) =>
    request<Credential>(`${ADMIN_BASE}/credentials/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteCredential: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/credentials/${id}`, {
      method: 'DELETE',
    }),

  verifyCredential: (id: string) =>
    request<{ valid: boolean; message: string }>(`${ADMIN_BASE}/credentials/${id}/verify`, {
      method: 'POST',
    }),

  // Images (管理接口)
  listImages: (options?: { agentOnly?: boolean }) => {
    const params = new URLSearchParams()
    if (options?.agentOnly) params.set('agent_only', 'true')
    const query = params.toString()
    return request<Image[]>(`${ADMIN_BASE}/images${query ? `?${query}` : ''}`)
  },

  pullImage: (req: PullImageRequest) =>
    request<{ message: string; image: string }>(`${ADMIN_BASE}/images/pull`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  removeImage: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/images/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),

  // System (管理接口)
  getSystemHealth: () => request<SystemHealth>(`${ADMIN_BASE}/system/health`),

  getSystemStats: () => request<SystemStats>(`${ADMIN_BASE}/system/stats`),

  cleanupContainers: () =>
    request<CleanupContainersResponse>(`${ADMIN_BASE}/system/cleanup/containers`, {
      method: 'POST',
    }),

  cleanupImages: (req?: CleanupImagesRequest) =>
    request<CleanupImagesResponse>(`${ADMIN_BASE}/system/cleanup/images`, {
      method: 'POST',
      body: JSON.stringify(req || { unused_only: true }),
    }),
}
