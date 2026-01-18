export interface Session {
  id: string
  agent: string
  status: 'creating' | 'running' | 'stopped' | 'error'
  workspace: string
  container_id?: string
  config: SessionConfig
  created_at: string
  updated_at: string
}

export interface SessionConfig {
  cpu_limit: number
  memory_limit: number
}

export interface Agent {
  name: string
  display_name: string
  description: string
  image: string
  required_env: string[]
}

export interface Execution {
  id: string
  session_id: string
  prompt: string
  status: 'pending' | 'running' | 'success' | 'failed'
  output?: string
  error?: string
  exit_code: number
  started_at: string
  ended_at?: string
}

export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

export interface CreateSessionRequest {
  agent: string
  workspace: string
  env?: Record<string, string>
}

export interface ExecRequest {
  prompt: string
}

export interface ExecResponse {
  execution_id: string
  output: string
  exit_code: number
  error?: string
}
