export interface Session {
  id: string
  agent: string
  profile_id?: string
  status: 'creating' | 'running' | 'stopped' | 'error'
  workspace: string
  container_id?: string
  config: SessionConfig
  env?: Record<string, string>
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
  profile_id?: string
  workspace: string
  env?: Record<string, string>
}

export interface ExecRequest {
  prompt: string
  max_turns?: number
  timeout?: number
  allowed_tools?: string[]
  disallowed_tools?: string[]
  include_events?: boolean // 是否返回完整事件列表
}

export interface ExecResponse {
  execution_id: string
  message: string // Agent 最终回复
  output: string // 原始输出 (兼容旧版)
  events?: ExecEvent[] // 完整事件列表 (当 include_events=true)
  usage?: TokenUsage // Token 使用统计
  exit_code: number
  error?: string
}

export interface TokenUsage {
  input_tokens: number
  cached_input_tokens?: number
  output_tokens: number
}

export interface ExecEvent {
  type: string
  raw?: unknown
}

// Profile Types
export interface Profile {
  id: string
  name: string
  description?: string
  icon?: string
  tags?: string[]
  adapter: 'claude-code' | 'codex' | 'opencode'
  extends?: string
  model: ModelConfig
  mcp_servers?: MCPServerConfig[]
  permissions: PermissionConfig
  resources: ResourceConfig
  system_prompt?: string
  append_system_prompt?: string
  base_instructions?: string
  developer_instructions?: string
  features: FeatureConfig
  config_overrides?: Record<string, string>
  output_format?: string
  output_schema?: string
  debug?: DebugConfig
  created_at: string
  updated_at: string
  created_by?: string
  is_built_in: boolean
  is_public: boolean
}

export interface ModelConfig {
  // Basic configuration
  name: string
  provider?: string
  base_url?: string  // Custom API endpoint (proxy/compatible API)
  reasoning_effort?: 'low' | 'medium' | 'high'

  // Model tier configuration (Claude Code)
  haiku_model?: string   // ANTHROPIC_DEFAULT_HAIKU_MODEL
  sonnet_model?: string  // ANTHROPIC_DEFAULT_SONNET_MODEL
  opus_model?: string    // ANTHROPIC_DEFAULT_OPUS_MODEL

  // Advanced configuration
  timeout_ms?: number        // API_TIMEOUT_MS
  max_output_tokens?: number // CLAUDE_CODE_MAX_OUTPUT_TOKENS
  disable_traffic?: boolean  // CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC

  // Codex specific
  wire_api?: 'chat' | 'responses'  // Codex config.toml wire_api
}

export interface MCPServerConfig {
  name: string
  command: string
  args?: string[]
  env?: Record<string, string>
  description?: string
}

export interface PermissionConfig {
  // Claude Code specific
  mode?: string
  allowed_tools?: string[]
  disallowed_tools?: string[]
  tools?: string[]
  skip_all?: boolean
  // Codex specific
  sandbox_mode?: 'read-only' | 'workspace-write' | 'danger-full-access'
  approval_policy?: 'untrusted' | 'on-failure' | 'on-request' | 'never'
  full_auto?: boolean
  // Common
  additional_dirs?: string[]
}

export interface ResourceConfig {
  max_budget_usd?: number
  max_turns?: number
  max_tokens?: number
  timeout?: number
  cpus?: number
  memory_mb?: number
  disk_gb?: number
}

export interface FeatureConfig {
  web_search?: boolean
}

export interface DebugConfig {
  verbose?: boolean
}

export interface CreateProfileRequest {
  id: string
  name: string
  description?: string
  icon?: string
  tags?: string[]
  adapter: 'claude-code' | 'codex' | 'opencode'
  extends?: string
  model: ModelConfig
  mcp_servers?: MCPServerConfig[]
  permissions: PermissionConfig
  resources: ResourceConfig
  system_prompt?: string
  append_system_prompt?: string
  base_instructions?: string
  developer_instructions?: string
  features?: FeatureConfig
  config_overrides?: Record<string, string>
  output_format?: string
  output_schema?: string
  debug?: DebugConfig
}

export interface CloneProfileRequest {
  new_id: string
  new_name: string
}

// MCP Server Types
export type MCPServerType = 'stdio' | 'sse' | 'http'
export type MCPCategory = 'filesystem' | 'database' | 'api' | 'tool' | 'browser' | 'memory' | 'other'

export interface MCPServer {
  id: string
  name: string
  description?: string
  command: string
  args?: string[]
  env?: Record<string, string>
  work_dir?: string
  type: MCPServerType
  category: MCPCategory
  tags?: string[]
  url?: string
  is_built_in: boolean
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateMCPServerRequest {
  id: string
  name: string
  description?: string
  command: string
  args?: string[]
  env?: Record<string, string>
  work_dir?: string
  type?: MCPServerType
  category?: MCPCategory
  tags?: string[]
  url?: string
}

export interface UpdateMCPServerRequest {
  name?: string
  description?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  work_dir?: string
  type?: MCPServerType
  category?: MCPCategory
  tags?: string[]
  url?: string
  is_enabled?: boolean
}

export interface CloneMCPServerRequest {
  new_id: string
  new_name: string
}

// Skill Types
export type SkillCategory = 'coding' | 'review' | 'docs' | 'security' | 'testing' | 'other'

export interface SkillFile {
  path: string
  content: string
}

export interface Skill {
  id: string
  name: string
  description?: string
  command: string
  prompt: string
  files?: SkillFile[]
  allowed_tools?: string[]
  required_mcp?: string[]
  category: SkillCategory
  tags?: string[]
  author?: string
  version?: string
  is_built_in: boolean
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateSkillRequest {
  id: string
  name: string
  description?: string
  command: string
  prompt: string
  files?: SkillFile[]
  allowed_tools?: string[]
  required_mcp?: string[]
  category?: SkillCategory
  tags?: string[]
  author?: string
  version?: string
}

export interface UpdateSkillRequest {
  name?: string
  description?: string
  command?: string
  prompt?: string
  files?: SkillFile[]
  allowed_tools?: string[]
  required_mcp?: string[]
  category?: SkillCategory
  tags?: string[]
  author?: string
  version?: string
  is_enabled?: boolean
}

export interface CloneSkillRequest {
  new_id: string
  new_name: string
}

// Skill Store Types
export type SkillSourceType = 'official' | 'community' | 'custom'

export interface SkillSource {
  id: string
  name: string
  owner: string
  repo: string
  branch: string
  path: string
  type: SkillSourceType
  description?: string
  stars?: number
  updated_at?: string
  is_enabled: boolean
}

export interface RemoteSkill {
  id: string
  name: string
  description?: string
  command: string
  category: string
  author?: string
  version?: string
  source_id: string
  source_name: string
  path: string
  stars?: number
  is_installed: boolean
}

export interface InstallSkillRequest {
  source_id: string
  skill_id: string
}

export interface AddSourceRequest {
  id: string
  name: string
  owner: string
  repo: string
  branch?: string
  path?: string
  type?: SkillSourceType
  description?: string
}

// Credential Types
export type CredentialType = 'api_key' | 'token' | 'oauth'
export type CredentialProvider = 'anthropic' | 'openai' | 'github' | 'custom'
export type CredentialScope = 'global' | 'profile' | 'session'

export interface Credential {
  id: string
  name: string
  type: CredentialType
  provider: CredentialProvider
  value_masked?: string
  scope: CredentialScope
  profile_id?: string
  env_var?: string
  is_valid: boolean
  last_used_at?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface CreateCredentialRequest {
  id: string
  name: string
  type?: CredentialType
  provider: CredentialProvider
  value: string
  scope?: CredentialScope
  profile_id?: string
  env_var?: string
}

export interface UpdateCredentialRequest {
  name?: string
  type?: CredentialType
  provider?: CredentialProvider
  value?: string
  scope?: CredentialScope
  profile_id?: string
  env_var?: string
}

// Image Types
export interface Image {
  id: string
  tags: string[]
  size: number
  created: number
  in_use: boolean
  is_agent_image: boolean
}

export interface PullImageRequest {
  image: string
}

// System Types
export interface SystemHealth {
  status: string
  uptime: string
  docker: {
    status: string
    version?: string
    containers: number
    images: number
    error?: string
  }
  resources: {
    memory_usage_mb: number
    num_goroutines: number
    num_cpu: number
  }
  checks: Record<string, string>
}

export interface SystemStats {
  sessions: {
    total: number
    running: number
    stopped: number
    error: number
    creating: number
  }
  containers: {
    total: number
    running: number
    stopped: number
    other: number
  }
  images: {
    total: number
    agent_images: number
    total_size: number
    in_use: number
  }
  system: {
    uptime: string
    memory_usage_mb: number
    go_version: string
    num_cpu: number
    num_goroutines: number
  }
}

export interface CleanupContainersResponse {
  removed: string[]
  errors?: string[]
}

export interface CleanupImagesRequest {
  unused_only?: boolean
}

export interface CleanupImagesResponse {
  removed: string[]
  space_freed: number
  errors?: string[]
}

// Task Types
export type TaskStatus = 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface TaskInput {
  files?: string[]
  urls?: string[]
  text?: string
}

export interface TaskOutputConfig {
  format?: string
  path?: string
  include_files?: boolean
}

export interface TaskResult {
  summary?: string
  text?: string
  files?: string[]
  usage?: {
    duration_seconds: number
    tokens_used?: number
    cost_usd?: number
  }
  logs?: string
}

export interface Task {
  id: string
  profile_id: string
  profile_name: string
  agent_type: string
  prompt: string
  input?: TaskInput
  output?: TaskOutputConfig
  webhook_url?: string
  timeout?: number
  status: TaskStatus
  session_id?: string
  result?: TaskResult
  error_message?: string
  metadata?: Record<string, string>
  created_at: string
  queued_at?: string
  started_at?: string
  completed_at?: string
}

export interface CreateTaskRequest {
  profile_id: string
  prompt: string
  input?: TaskInput
  output?: TaskOutputConfig
  webhook_url?: string
  timeout?: number
  metadata?: Record<string, string>
}

// Webhook Types
export interface Webhook {
  id: string
  url: string
  secret?: string
  events: string[]
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateWebhookRequest {
  url: string
  secret?: string
  events?: string[]
}

export interface UpdateWebhookRequest {
  url?: string
  secret?: string
  events?: string[]
  is_active?: boolean
}

// Provider Types
export type ProviderCategory = 'official' | 'cn_official' | 'aggregator' | 'third_party'
export type ProviderAgent = 'claude-code' | 'codex' | 'opencode' | 'all'

export interface Provider {
  id: string
  name: string
  description?: string
  agent: ProviderAgent
  category: ProviderCategory
  website_url?: string
  api_key_url?: string
  docs_url?: string
  base_url?: string
  env_config?: Record<string, string>
  default_model?: string
  default_models?: string[]
  icon?: string
  icon_color?: string
  is_built_in: boolean
  is_partner?: boolean
  requires_ak?: boolean
  is_enabled: boolean
}

export interface CreateProviderRequest {
  id: string
  name: string
  description?: string
  agent: ProviderAgent
  category?: ProviderCategory
  website_url?: string
  api_key_url?: string
  docs_url?: string
  base_url?: string
  env_config?: Record<string, string>
  default_model?: string
  default_models?: string[]
  icon?: string
  icon_color?: string
}

export interface UpdateProviderRequest {
  name?: string
  description?: string
  base_url?: string
  env_config?: Record<string, string>
  default_model?: string
  default_models?: string[]
}
