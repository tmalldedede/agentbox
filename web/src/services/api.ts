import type {
  Session,
  Agent,
  ApiResponse,
  CreateSessionRequest,
  ExecRequest,
  ExecResponse,
  Execution
} from '../types'

const API_BASE = '/api'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
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
  // Health
  health: () => request<{ status: string; version: string }>('/health'),

  // Agents
  listAgents: () => request<Agent[]>('/agents'),

  // Sessions
  listSessions: () => request<Session[]>('/sessions'),

  getSession: (id: string) => request<Session>(`/sessions/${id}`),

  createSession: (req: CreateSessionRequest) =>
    request<Session>('/sessions', {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteSession: (id: string) =>
    request<{ deleted: string }>(`/sessions/${id}`, {
      method: 'DELETE',
    }),

  startSession: (id: string) =>
    request<Session>(`/sessions/${id}/start`, {
      method: 'POST',
    }),

  stopSession: (id: string) =>
    request<Session>(`/sessions/${id}/stop`, {
      method: 'POST',
    }),

  execSession: (id: string, req: ExecRequest) =>
    request<ExecResponse>(`/sessions/${id}/exec`, {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  getExecutions: (id: string) =>
    request<Execution[]>(`/sessions/${id}/executions`),
}
