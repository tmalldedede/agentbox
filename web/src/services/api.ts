import type {
  Engine,
  Session,
  Agent,
  CreateAgentRequest,
  UpdateAgentRequest,
  RunAgentRequest,
  AgentRunResult,
  AgentRuntime,
  CreateRuntimeRequest,
  UpdateRuntimeRequest,
  ApiResponse,
  CreateSessionRequest,
  ExecRequest,
  ExecResponse,
  Execution,
  Provider,
  ProviderStats,
  VerifyResult,
  CreateProviderRequest,
  UpdateProviderRequest,
  AuthProfile,
  RotatorStats,
  AddAuthProfileRequest,
  MCPServer,
  MCPServerStats,
  MCPTestResult,
  CreateMCPServerRequest,
  UpdateMCPServerRequest,
  CloneMCPServerRequest,
  Skill,
  SkillMetadata,
  SkillCheckResult,
  SkillStats,
  SkillLoadLevel,
  SkillStatusReport,
  SkillConfig,
  CreateSkillRequest,
  UpdateSkillRequest,
  CloneSkillRequest,
  SkillSource,
  SkillSourceType,
  RemoteSkill,
  InstallSkillRequest,
  AddSourceRequest,
  Image,
  PullImageRequest,
  SystemHealth,
  SystemStats,
  GCStats,
  GCCandidate,
  UpdateGCConfigRequest,
  CleanupContainersResponse,
  CleanupImagesRequest,
  CleanupImagesResponse,
  Webhook,
  CreateWebhookRequest,
  UpdateWebhookRequest,
  Task,
  TaskStats,
  CreateTaskRequest,
  UploadedFile,
  HistoryEntry,
  HistoryStats,
  HistoryListResponse,
  HistoryFilter,
  DashboardStats,
  Batch,
  BatchTask,
  BatchStats,
  CreateBatchRequest,
  ListBatchFilter,
  ListBatchTaskFilter,
  Settings,
  AgentSettings,
  TaskSettings,
  BatchSettings,
  StorageSettings,
  NotifySettings,
  ChannelSession,
  ChannelMessage,
  ChannelStats,
  ChannelSessionFilter,
  ChannelMessageFilter,
} from '../types'

export const API_BASE = '/api/v1'
const ADMIN_BASE = '/api/v1/admin'
const TOKEN_KEY = 'agentbox_token'

function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }
  }
  return { 'Content-Type': 'application/json' }
}

export async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: getAuthHeaders(),
    ...options,
  })

  // Handle 401 - redirect to login
  if (response.status === 401) {
    localStorage.removeItem(TOKEN_KEY)
    window.location.href = '/sign-in'
    throw new Error('Unauthorized')
  }

  // Handle 403 - forbidden
  if (response.status === 403) {
    throw new Error('Access denied: admin privileges required')
  }

  const data: ApiResponse<T> = await response.json()

  if (data.code !== 0) {
    throw new Error(data.message)
  }

  return data.data as T
}

// Auth types
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  user: {
    id: string
    username: string
    role: string
  }
}

export interface UserInfo {
  id: string
  username: string
  role: string
}

export interface UserResponse {
  id: string
  username: string
  role: string
  is_active: boolean
  created_at: string
}

export interface CreateUserRequest {
  username: string
  password: string
  role?: string
}

export interface APIKeyResponse {
  id: string
  name: string
  key_prefix: string
  key?: string // 仅创建时返回
  last_used_at?: string
  expires_at?: string
  created_at: string
}

export const api = {
  // ==================== Auth API ====================
  login: (req: LoginRequest) =>
    request<LoginResponse>(`${API_BASE}/auth/login`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  me: () => request<UserInfo>(`${API_BASE}/auth/me`),

  changePassword: (req: { old_password: string; new_password: string }) =>
    request<{ message: string }>(`${API_BASE}/auth/password`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  // API Keys
  listAPIKeys: () => request<APIKeyResponse[]>(`${API_BASE}/auth/api-keys`),

  createAPIKey: (req: { name: string; expires_in?: number }) =>
    request<APIKeyResponse>(`${API_BASE}/auth/api-keys`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteAPIKey: (id: string) =>
    request<{ id: string; deleted: boolean }>(`${API_BASE}/auth/api-keys/${id}`, {
      method: 'DELETE',
    }),

  // Admin: User management
  listUsers: () => request<UserResponse[]>(`${ADMIN_BASE}/users`),

  createUser: (req: CreateUserRequest) =>
    request<UserResponse>(`${ADMIN_BASE}/users`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteUser: (id: string) =>
    request<{ id: string; deleted: boolean }>(`${ADMIN_BASE}/users/${id}`, {
      method: 'DELETE',
    }),

  // ==================== Public API (Task-Centric) ====================

  // Health
  health: () => request<{ status: string; version: string }>(`${API_BASE}/health`),

  // Engines (只读)
  listEngines: () => request<Engine[]>(`${API_BASE}/engines`),

  // Agents (只读 - 只暴露 active agents)
  listAgents: () => request<Agent[]>(`${API_BASE}/agents`),
  getAgent: (id: string) => request<Agent>(`${API_BASE}/agents/${id}`),

  // Tasks (核心 API) - 创建/多轮/取消/SSE
  listTasks: (options?: { status?: string; agent_id?: string; search?: string; limit?: number; offset?: number }) => {
    const params = new URLSearchParams()
    if (options?.status) params.set('status', options.status)
    if (options?.agent_id) params.set('agent_id', options.agent_id)
    if (options?.search) params.set('search', options.search)
    if (options?.limit) params.set('limit', options.limit.toString())
    if (options?.offset) params.set('offset', options.offset.toString())
    const query = params.toString()
    return request<{ tasks: Task[]; total: number }>(`${API_BASE}/tasks${query ? `?${query}` : ''}`)
  },

  getTask: (id: string) => request<Task>(`${API_BASE}/tasks/${id}`),

  getTaskStats: () => request<TaskStats>(`${API_BASE}/tasks/stats`),

  createTask: (req: CreateTaskRequest) =>
    request<Task>(`${API_BASE}/tasks`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  cancelTask: (id: string) =>
    request<Task>(`${API_BASE}/tasks/${id}/cancel`, {
      method: 'POST',
    }),

  retryTask: (id: string) =>
    request<Task>(`${API_BASE}/tasks/${id}/retry`, {
      method: 'POST',
    }),

  deleteTask: (id: string) =>
    request<{ deleted: boolean }>(`${API_BASE}/tasks/${id}`, {
      method: 'DELETE',
    }),

  cleanupTasks: (options?: { before_days?: number; statuses?: string[] }) =>
    request<{ deleted: number }>(`${API_BASE}/tasks/cleanup`, {
      method: 'POST',
      body: JSON.stringify(options || {}),
    }),

  getTaskOutput: (id: string) =>
    request<unknown>(`${API_BASE}/tasks/${id}/output`),

  // SSE 事件流
  streamTaskEvents: (id: string): EventSource => {
    const token = localStorage.getItem(TOKEN_KEY)
    const url = token
      ? `${API_BASE}/tasks/${id}/events?token=${token}`
      : `${API_BASE}/tasks/${id}/events`
    return new EventSource(url)
  },

  // Batches (批量任务 API) - Worker 池模式
  listBatches: (filter?: ListBatchFilter) => {
    const params = new URLSearchParams()
    if (filter?.status) params.set('status', filter.status)
    if (filter?.agent_id) params.set('agent_id', filter.agent_id)
    if (filter?.limit) params.set('limit', filter.limit.toString())
    if (filter?.offset) params.set('offset', filter.offset.toString())
    const query = params.toString()
    return request<{ batches: Batch[]; total: number }>(`${API_BASE}/batches${query ? `?${query}` : ''}`)
  },

  getBatch: (id: string) => request<Batch>(`${API_BASE}/batches/${id}`),

  createBatch: (req: CreateBatchRequest) =>
    request<Batch>(`${API_BASE}/batches`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteBatch: (id: string) =>
    request<{ deleted: boolean }>(`${API_BASE}/batches/${id}`, {
      method: 'DELETE',
    }),

  startBatch: (id: string) =>
    request<Batch>(`${API_BASE}/batches/${id}/start`, {
      method: 'POST',
    }),

  pauseBatch: (id: string) =>
    request<Batch>(`${API_BASE}/batches/${id}/pause`, {
      method: 'POST',
    }),

  resumeBatch: (id: string) =>
    request<Batch>(`${API_BASE}/batches/${id}/resume`, {
      method: 'POST',
    }),

  cancelBatch: (id: string) =>
    request<Batch>(`${API_BASE}/batches/${id}/cancel`, {
      method: 'POST',
    }),

  retryBatchFailed: (id: string) =>
    request<Batch>(`${API_BASE}/batches/${id}/retry`, {
      method: 'POST',
    }),

  listBatchTasks: (batchId: string, filter?: ListBatchTaskFilter) => {
    const params = new URLSearchParams()
    if (filter?.status) params.set('status', filter.status)
    if (filter?.worker_id) params.set('worker_id', filter.worker_id)
    if (filter?.limit) params.set('limit', filter.limit.toString())
    if (filter?.offset) params.set('offset', filter.offset.toString())
    const query = params.toString()
    return request<{ tasks: BatchTask[]; total: number }>(`${API_BASE}/batches/${batchId}/tasks${query ? `?${query}` : ''}`)
  },

  getBatchTask: (batchId: string, taskId: string) =>
    request<BatchTask>(`${API_BASE}/batches/${batchId}/tasks/${taskId}`),

  getBatchStats: (id: string) =>
    request<BatchStats>(`${API_BASE}/batches/${id}/stats`),

  // Batch SSE 事件流
  streamBatchEvents: (id: string): EventSource => {
    const token = localStorage.getItem(TOKEN_KEY)
    const url = token
      ? `${API_BASE}/batches/${id}/events?token=${token}`
      : `${API_BASE}/batches/${id}/events`
    return new EventSource(url)
  },

  // Batch 结果导出
  getBatchExportUrl: (id: string, format: 'json' | 'csv' = 'json') =>
    `${API_BASE}/batches/${id}/export?format=${format}`,

  // Batch 死信队列
  listBatchDeadTasks: (batchId: string, limit = 100) =>
    request<{ batch_id: string; tasks: BatchTask[]; count: number }>(`${API_BASE}/batches/${batchId}/dead?limit=${limit}`),

  retryBatchDead: (batchId: string, taskIds?: string[]) =>
    request<{ batch_id: string; retried_count: number }>(`${API_BASE}/batches/${batchId}/dead/retry`, {
      method: 'POST',
      body: JSON.stringify({ task_ids: taskIds }),
    }),

  // Files (附件 API) - 文件上传/下载
  listFiles: () => request<UploadedFile[]>(`${API_BASE}/files`),

  getFile: (id: string) => request<UploadedFile>(`${API_BASE}/files/${id}`),

  uploadFile: async (file: File): Promise<UploadedFile> => {
    const formData = new FormData()
    formData.append('file', file)
    const token = localStorage.getItem(TOKEN_KEY)
    const headers: Record<string, string> = {}
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    const response = await fetch(`${API_BASE}/files`, {
      method: 'POST',
      headers,
      body: formData,
    })
    if (response.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      window.location.href = '/sign-in'
      throw new Error('Unauthorized')
    }
    const data: ApiResponse<UploadedFile> = await response.json()
    if (data.code !== 0) {
      throw new Error(data.message)
    }
    return data.data as UploadedFile
  },

  deleteFile: (id: string) =>
    request<{ deleted: string }>(`${API_BASE}/files/${id}`, {
      method: 'DELETE',
    }),

  getFileDownloadUrl: (id: string) => `${API_BASE}/files/${id}/download`,

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

  // History (认证用户可用) - 统一执行历史
  listHistory: (filter?: HistoryFilter) => {
    const params = new URLSearchParams()
    if (filter?.source_type) params.set('source_type', filter.source_type)
    if (filter?.source_id) params.set('source_id', filter.source_id)
    if (filter?.agent_id) params.set('agent_id', filter.agent_id)
    if (filter?.engine) params.set('engine', filter.engine)
    if (filter?.status) params.set('status', filter.status)
    if (filter?.limit) params.set('limit', filter.limit.toString())
    if (filter?.offset) params.set('offset', filter.offset.toString())
    const query = params.toString()
    return request<HistoryListResponse>(`${API_BASE}/history${query ? `?${query}` : ''}`)
  },

  getHistoryEntry: (id: string) => request<HistoryEntry>(`${API_BASE}/history/${id}`),

  deleteHistoryEntry: (id: string) =>
    request<{ id: string; deleted: boolean }>(`${API_BASE}/history/${id}`, {
      method: 'DELETE',
    }),

  getHistoryStats: (filter?: HistoryFilter) => {
    const params = new URLSearchParams()
    if (filter?.source_type) params.set('source_type', filter.source_type)
    if (filter?.agent_id) params.set('agent_id', filter.agent_id)
    if (filter?.engine) params.set('engine', filter.engine)
    const query = params.toString()
    return request<HistoryStats>(`${API_BASE}/history/stats${query ? `?${query}` : ''}`)
  },

  // ==================== Admin API ====================

  // Sessions (admin 调试用)
  listSessions: () => request<Session[]>(`${ADMIN_BASE}/sessions`),

  getSession: (id: string) => request<Session>(`${ADMIN_BASE}/sessions/${id}`),

  createSession: (req: CreateSessionRequest) =>
    request<Session>(`${ADMIN_BASE}/sessions`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteSession: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/sessions/${id}`, {
      method: 'DELETE',
    }),

  startSession: (id: string) =>
    request<Session>(`${ADMIN_BASE}/sessions/${id}/start`, {
      method: 'POST',
    }),

  stopSession: (id: string) =>
    request<Session>(`${ADMIN_BASE}/sessions/${id}/stop`, {
      method: 'POST',
    }),

  execSession: (id: string, req: ExecRequest) =>
    request<ExecResponse>(`${ADMIN_BASE}/sessions/${id}/exec`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  getExecutions: (id: string) =>
    request<Execution[]>(`${ADMIN_BASE}/sessions/${id}/executions`),

  getExecution: (sessionId: string, execId: string) =>
    request<Execution>(`${ADMIN_BASE}/sessions/${sessionId}/executions/${execId}`),

  getSessionLogs: (id: string) =>
    request<{ logs: string }>(`${ADMIN_BASE}/sessions/${id}/logs`),

  // Agents (admin CRUD + Run)
  listAdminAgents: () => request<Agent[]>(`${ADMIN_BASE}/agents`),

  createAgent: (req: CreateAgentRequest) =>
    request<Agent>(`${ADMIN_BASE}/agents`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateAgent: (id: string, req: UpdateAgentRequest) =>
    request<Agent>(`${ADMIN_BASE}/agents/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteAgent: (id: string) =>
    request<{ id: string; deleted: boolean }>(`${ADMIN_BASE}/agents/${id}`, {
      method: 'DELETE',
    }),

  runAgent: (id: string, req: RunAgentRequest) =>
    request<AgentRunResult>(`${ADMIN_BASE}/agents/${id}/run`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  // Providers (admin CRUD + Key 管理)
  listProviders: (options?: { agent?: string; category?: string; configured?: boolean }) => {
    const params = new URLSearchParams()
    if (options?.agent) params.set('agent', options.agent)
    if (options?.category) params.set('category', options.category)
    if (options?.configured) params.set('configured', 'true')
    const query = params.toString()
    return request<Provider[]>(`${ADMIN_BASE}/providers${query ? `?${query}` : ''}`)
  },

  listProviderTemplates: () =>
    request<Provider[]>(`${ADMIN_BASE}/providers/templates`),

  getProviderStats: () =>
    request<ProviderStats>(`${ADMIN_BASE}/providers/stats`),

  verifyAllProviders: () =>
    request<VerifyResult[]>(`${ADMIN_BASE}/providers/verify-all`, {
      method: 'POST',
    }),

  getProvider: (id: string) => request<Provider>(`${ADMIN_BASE}/providers/${id}`),

  createProvider: (req: CreateProviderRequest) =>
    request<Provider>(`${ADMIN_BASE}/providers`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateProvider: (id: string, req: UpdateProviderRequest) =>
    request<Provider>(`${ADMIN_BASE}/providers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteProvider: (id: string) =>
    request<void>(`${ADMIN_BASE}/providers/${id}`, {
      method: 'DELETE',
    }),

  configureProviderKey: (id: string, apiKey: string) =>
    request<Provider>(`${ADMIN_BASE}/providers/${id}/key`, {
      method: 'PUT',
      body: JSON.stringify({ api_key: apiKey }),
    }),

  verifyProviderKey: (id: string) =>
    request<{ valid: boolean; message: string }>(`${ADMIN_BASE}/providers/${id}/verify`, {
      method: 'POST',
    }),

  deleteProviderKey: (id: string) =>
    request<Provider>(`${ADMIN_BASE}/providers/${id}/key`, {
      method: 'DELETE',
    }),

  fetchProviderModels: (id: string) =>
    request<string[]>(`${ADMIN_BASE}/providers/${id}/models`),

  probeModels: (baseURL: string, apiKey: string, agents: string[]) =>
    request<string[]>(`${ADMIN_BASE}/providers/probe-models`, {
      method: 'POST',
      body: JSON.stringify({ base_url: baseURL, api_key: apiKey, agents }),
    }),

  // Auth Profile Rotation
  listAuthProfiles: (providerId: string) =>
    request<AuthProfile[]>(`${ADMIN_BASE}/providers/${providerId}/profiles`),

  addAuthProfile: (providerId: string, req: AddAuthProfileRequest) =>
    request<AuthProfile>(`${ADMIN_BASE}/providers/${providerId}/profiles`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  removeAuthProfile: (providerId: string, profileId: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/providers/${providerId}/profiles/${profileId}`, {
      method: 'DELETE',
    }),

  getRotationStats: (providerId: string) =>
    request<RotatorStats>(`${ADMIN_BASE}/providers/${providerId}/rotation-stats`),

  // Runtimes (管理接口) - 运行环境配置
  listRuntimes: () => request<AgentRuntime[]>(`${ADMIN_BASE}/runtimes`),

  getRuntime: (id: string) => request<AgentRuntime>(`${ADMIN_BASE}/runtimes/${id}`),

  createRuntime: (req: CreateRuntimeRequest) =>
    request<AgentRuntime>(`${ADMIN_BASE}/runtimes`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateRuntime: (id: string, req: UpdateRuntimeRequest) =>
    request<AgentRuntime>(`${ADMIN_BASE}/runtimes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteRuntime: (id: string) =>
    request<{ deleted: string }>(`${ADMIN_BASE}/runtimes/${id}`, {
      method: 'DELETE',
    }),

  setDefaultRuntime: (id: string) =>
    request<AgentRuntime>(`${ADMIN_BASE}/runtimes/${id}/set-default`, {
      method: 'POST',
    }),

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
    request<MCPTestResult>(`${ADMIN_BASE}/mcp-servers/${id}/test`, {
      method: 'POST',
    }),

  getMCPServerStats: () =>
    request<MCPServerStats>(`${ADMIN_BASE}/mcp-servers/stats`),

  // Skills (管理接口)
  listSkills: (options?: {
    category?: string
    enabled?: boolean
    level?: SkillLoadLevel
    source?: SkillSourceType
  }) => {
    const params = new URLSearchParams()
    if (options?.category) params.set('category', options.category)
    if (options?.enabled) params.set('enabled', 'true')
    if (options?.level) params.set('level', options.level)
    if (options?.source) params.set('source', options.source)
    const query = params.toString()
    return request<Skill[]>(`${ADMIN_BASE}/skills${query ? `?${query}` : ''}`)
  },

  // 快速列出元数据（用于列表页）
  listSkillMetadata: (options?: {
    category?: string
    source?: SkillSourceType
  }) => {
    const params = new URLSearchParams()
    params.set('level', 'metadata')
    if (options?.category) params.set('category', options.category)
    if (options?.source) params.set('source', options.source)
    const query = params.toString()
    return request<SkillMetadata[]>(`${ADMIN_BASE}/skills?${query}`)
  },

  getSkill: (id: string, level?: SkillLoadLevel) => {
    const params = level ? `?level=${level}` : ''
    return request<Skill>(`${ADMIN_BASE}/skills/${id}${params}`)
  },

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

  // 检查依赖
  checkSkillDeps: (id: string, containerId?: string) =>
    request<SkillCheckResult>(`${ADMIN_BASE}/skills/${id}/check-deps`, {
      method: 'POST',
      body: JSON.stringify({ container_id: containerId }),
    }),

  // 获取统计信息
  getSkillStats: () => request<SkillStats>(`${ADMIN_BASE}/skills/stats`),

  // 获取完整状态报告（借鉴 Clawdbot）
  getSkillStatus: () => request<SkillStatusReport>(`${ADMIN_BASE}/skills/status`),

  // 列出所有需要的二进制
  listRequiredBins: () => request<{ bins: string[] }>(`${ADMIN_BASE}/skills/bins`),

  // 更新 Skill 配置
  updateSkillConfig: (id: string, config: Partial<SkillConfig>) =>
    request<{ skill_id: string; config: SkillConfig }>(`${ADMIN_BASE}/skills/${id}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  exportSkill: (id: string) => {
    const token = localStorage.getItem(TOKEN_KEY)
    const headers: Record<string, string> = {}
    if (token) headers['Authorization'] = `Bearer ${token}`
    return fetch(`${ADMIN_BASE}/skills/${id}/export`, { headers }).then(res => res.text())
  },

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

  // Dashboard (态势感知大屏)
  getDashboardStats: () => request<DashboardStats>(`${ADMIN_BASE}/dashboard/stats`),

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

  // GC (管理接口)
  getGCStats: () => request<GCStats>(`${ADMIN_BASE}/system/gc/stats`),

  triggerGC: () =>
    request<{ removed: number }>(`${ADMIN_BASE}/system/gc/trigger`, {
      method: 'POST',
    }),

  previewGC: () =>
    request<GCCandidate[]>(`${ADMIN_BASE}/system/gc/preview`, {
      method: 'POST',
    }),

  updateGCConfig: (req: UpdateGCConfigRequest) =>
    request<GCStats>(`${ADMIN_BASE}/system/gc/config`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  // Settings (业务配置)
  getSettings: () => request<Settings>(`${ADMIN_BASE}/settings`),

  updateSettings: (settings: Settings) =>
    request<Settings>(`${ADMIN_BASE}/settings`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  resetSettings: () =>
    request<Settings>(`${ADMIN_BASE}/settings/reset`, {
      method: 'POST',
    }),

  getAgentSettings: () => request<AgentSettings>(`${ADMIN_BASE}/settings/agent`),

  updateAgentSettings: (settings: AgentSettings) =>
    request<AgentSettings>(`${ADMIN_BASE}/settings/agent`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  getTaskSettings: () => request<TaskSettings>(`${ADMIN_BASE}/settings/task`),

  updateTaskSettings: (settings: TaskSettings) =>
    request<TaskSettings>(`${ADMIN_BASE}/settings/task`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  getBatchSettings: () => request<BatchSettings>(`${ADMIN_BASE}/settings/batch`),

  updateBatchSettings: (settings: BatchSettings) =>
    request<BatchSettings>(`${ADMIN_BASE}/settings/batch`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  getStorageSettings: () => request<StorageSettings>(`${ADMIN_BASE}/settings/storage`),

  updateStorageSettings: (settings: StorageSettings) =>
    request<StorageSettings>(`${ADMIN_BASE}/settings/storage`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  getNotifySettings: () => request<NotifySettings>(`${ADMIN_BASE}/settings/notify`),

  updateNotifySettings: (settings: NotifySettings) =>
    request<NotifySettings>(`${ADMIN_BASE}/settings/notify`, {
      method: 'PUT',
      body: JSON.stringify(settings),
    }),

  // ==================== Cron Jobs API ====================
  listCronJobs: () => request<CronJob[]>(`${ADMIN_BASE}/crons`),

  getCronJob: (id: string) => request<CronJob>(`${ADMIN_BASE}/crons/${id}`),

  createCronJob: (req: CreateCronJobRequest) =>
    request<CronJob>(`${ADMIN_BASE}/crons`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  updateCronJob: (id: string, req: UpdateCronJobRequest) =>
    request<CronJob>(`${ADMIN_BASE}/crons/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  deleteCronJob: (id: string) =>
    request<void>(`${ADMIN_BASE}/crons/${id}`, {
      method: 'DELETE',
    }),

  triggerCronJob: (id: string) =>
    request<{ message: string }>(`${ADMIN_BASE}/crons/${id}/trigger`, {
      method: 'POST',
    }),

  // ==================== Channels API ====================
  listChannels: () => request<string[]>(`${ADMIN_BASE}/channels`),

  sendChannelMessage: (req: SendChannelMessageRequest) =>
    request<{ message_id: string }>(`${ADMIN_BASE}/channels/send`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  // Channel Sessions
  listChannelSessions: (filter?: ChannelSessionFilter) => {
    const params = new URLSearchParams()
    if (filter?.channel_type) params.set('channel_type', filter.channel_type)
    if (filter?.status) params.set('status', filter.status)
    if (filter?.agent_id) params.set('agent_id', filter.agent_id)
    if (filter?.limit) params.set('limit', filter.limit.toString())
    if (filter?.offset) params.set('offset', filter.offset.toString())
    const query = params.toString()
    return request<{ sessions: ChannelSession[]; total: number }>(`${ADMIN_BASE}/channel-sessions${query ? `?${query}` : ''}`)
  },

  getChannelSession: (id: string) =>
    request<ChannelSession>(`${ADMIN_BASE}/channel-sessions/${id}`),

  getSessionMessages: (id: string, limit = 50, offset = 0) =>
    request<{ messages: ChannelMessage[]; total: number }>(`${ADMIN_BASE}/channel-sessions/${id}/messages?limit=${limit}&offset=${offset}`),

  endChannelSession: (id: string) =>
    request<{ message: string }>(`${ADMIN_BASE}/channel-sessions/${id}/end`, {
      method: 'POST',
    }),

  // Channel Messages
  listChannelMessages: (filter?: ChannelMessageFilter) => {
    const params = new URLSearchParams()
    if (filter?.channel_type) params.set('channel_type', filter.channel_type)
    if (filter?.direction) params.set('direction', filter.direction)
    if (filter?.task_id) params.set('task_id', filter.task_id)
    if (filter?.limit) params.set('limit', filter.limit.toString())
    if (filter?.offset) params.set('offset', filter.offset.toString())
    const query = params.toString()
    return request<{ messages: ChannelMessage[]; total: number }>(`${ADMIN_BASE}/channel-messages${query ? `?${query}` : ''}`)
  },

  // Channel Stats
  getChannelStats: (channelType?: string) => {
    const query = channelType ? `?channel_type=${channelType}` : ''
    return request<ChannelStats>(`${ADMIN_BASE}/channel-stats${query}`)
  },

  // Feishu Config
  getFeishuConfig: () => request<FeishuConfig>(`${ADMIN_BASE}/feishu/config`),

  saveFeishuConfig: (req: SaveFeishuConfigRequest) =>
    request<{ id: string; message: string }>(`${ADMIN_BASE}/feishu/config`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  listFeishuConfigs: () => request<FeishuConfigItem[]>(`${ADMIN_BASE}/feishu/configs`),

  deleteFeishuConfig: (id: string) =>
    request<void>(`${ADMIN_BASE}/feishu/config/${id}`, {
      method: 'DELETE',
    }),

  // ==================== Coordinate API ====================
  listActiveSessions: () => request<CoordinateSession[]>(`${API_BASE}/coordinate/sessions`),

  getSessionHistory: (taskId: string) =>
    request<CoordinateMessage[]>(`${API_BASE}/coordinate/sessions/${taskId}/history`),

  sendSessionMessage: (taskId: string, message: string) =>
    request<{ message: string }>(`${API_BASE}/coordinate/sessions/${taskId}/send`, {
      method: 'POST',
      body: JSON.stringify({ message }),
    }),
}

// ==================== Cron Types ====================
export interface CronJob {
  id: string
  name: string
  schedule: string
  enabled: boolean
  agent_id: string
  prompt: string
  metadata?: Record<string, string>
  last_run?: string
  next_run?: string
  last_status?: string
  last_error?: string
  created_at: string
  updated_at: string
}

export interface CreateCronJobRequest {
  name: string
  schedule: string
  agent_id: string
  prompt: string
  enabled?: boolean
  metadata?: Record<string, string>
}

export interface UpdateCronJobRequest {
  name?: string
  schedule?: string
  agent_id?: string
  prompt?: string
  enabled?: boolean
  metadata?: Record<string, string>
}

// ==================== Channel Types ====================
export interface SendChannelMessageRequest {
  channel_type: string
  channel_id: string
  content: string
  reply_to?: string
}

export interface FeishuConfig {
  id: string
  name: string
  app_id: string
  app_secret: string
  verification_token: string
  encrypt_key: string
  bot_name: string
  default_agent_id?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface SaveFeishuConfigRequest {
  id?: string
  name: string
  app_id: string
  app_secret?: string
  verification_token?: string
  encrypt_key?: string
  bot_name?: string
  default_agent_id?: string
}

export interface FeishuConfigItem {
  id: string
  name: string
  app_id: string
  app_secret: string
  verification_token: string
  encrypt_key: string
  bot_name: string
  default_agent_id?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

// ==================== Coordinate Types ====================
export interface CoordinateSession {
  task_id: string
  agent_id: string
  agent_name: string
  status: string
  prompt: string
  turn_count: number
  started_at: string
  updated_at: string
}

export interface CoordinateMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}
