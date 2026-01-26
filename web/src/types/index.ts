// Engine Types (底层引擎适配器信息)
export interface Engine {
  name: string
  display_name: string
  description: string
  image: string
  required_env: string[]
}

// Provider Types
export type ProviderCategory = 'official' | 'cn_official' | 'aggregator' | 'third_party'

export interface Provider {
  id: string
  name: string
  description?: string
  template_id?: string
  agents: AdapterType[]
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
  // API Key management
  api_key_masked?: string
  is_configured: boolean
  is_valid: boolean
  last_validated_at?: string
}

export interface ProviderStats {
  total: number
  configured: number
  valid: number
  failed: number
  not_configured: number
}

export interface VerifyResult {
  id: string
  name: string
  valid: boolean
  error?: string
}

export interface CreateProviderRequest {
  id: string
  name: string
  // Template-based creation
  template_id?: string
  api_key?: string
  models?: string[]
  // Common
  base_url?: string
  // Custom creation
  description?: string
  agents?: AdapterType[]
  category?: ProviderCategory
  website_url?: string
  api_key_url?: string
  docs_url?: string
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

// AgentRuntime Types
export interface AgentRuntime {
  id: string
  name: string
  description?: string
  image: string
  cpus: number
  memory_mb: number
  network: string
  privileged: boolean
  is_built_in: boolean
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface CreateRuntimeRequest {
  id: string
  name: string
  description?: string
  image: string
  cpus?: number
  memory_mb?: number
  network?: string
  privileged?: boolean
}

export interface UpdateRuntimeRequest {
  name?: string
  description?: string
  image?: string
  cpus?: number
  memory_mb?: number
  network?: string
  privileged?: boolean
}

// Agent Types (合并 Profile + SmartAgent)
export type AgentStatus = 'active' | 'inactive'
export type AgentAPIAccess = 'public' | 'api_key' | 'private'
export type AdapterType = 'claude-code' | 'codex' | 'opencode'

export interface Agent {
  id: string
  name: string
  description?: string
  icon?: string
  adapter: AdapterType
  provider_id: string
  runtime_id?: string
  model?: string
  base_url_override?: string
  model_config?: ModelConfig
  skill_ids?: string[]
  mcp_server_ids?: string[]
  system_prompt?: string
  append_system_prompt?: string
  permissions: PermissionConfig
  workspace?: string
  env?: Record<string, string>
  api_access?: AgentAPIAccess
  rate_limit?: number
  webhook_url?: string
  output_format?: string
  features?: FeatureConfig
  config_overrides?: Record<string, string>
  status: AgentStatus
  is_built_in: boolean
  created_at: string
  updated_at: string
}

export interface CreateAgentRequest {
  id: string
  name: string
  description?: string
  icon?: string
  adapter: AdapterType
  provider_id: string
  runtime_id?: string
  model?: string
  base_url_override?: string
  model_config?: ModelConfig
  skill_ids?: string[]
  mcp_server_ids?: string[]
  system_prompt?: string
  append_system_prompt?: string
  permissions?: PermissionConfig
  workspace?: string
  env?: Record<string, string>
  api_access?: AgentAPIAccess
  rate_limit?: number
  webhook_url?: string
  output_format?: string
  features?: FeatureConfig
  config_overrides?: Record<string, string>
}

export interface UpdateAgentRequest {
  name?: string
  description?: string
  icon?: string
  adapter?: AdapterType
  provider_id?: string
  runtime_id?: string
  model?: string
  base_url_override?: string
  model_config?: ModelConfig
  skill_ids?: string[]
  mcp_server_ids?: string[]
  system_prompt?: string
  append_system_prompt?: string
  permissions?: PermissionConfig
  workspace?: string
  env?: Record<string, string>
  api_access?: AgentAPIAccess
  rate_limit?: number
  webhook_url?: string
  output_format?: string
  features?: FeatureConfig
  config_overrides?: Record<string, string>
  status?: AgentStatus
}

export interface RunAgentRequest {
  prompt: string
  workspace?: string
  input?: {
    files?: string[]
    text?: string
  }
  options?: {
    max_turns?: number
    timeout?: number
  }
  metadata?: Record<string, string>
}

export interface AgentRunResult {
  run_id: string
  agent_id: string
  agent_name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  output?: string
  error?: string
  usage?: {
    input_tokens?: number
    output_tokens?: number
    duration_ms?: number
    cost_usd?: number
  }
  started_at: string
  ended_at?: string
}

// Model Config Types
export interface ModelConfig {
  name: string
  provider?: string
  base_url?: string
  reasoning_effort?: 'low' | 'medium' | 'high'
  haiku_model?: string
  sonnet_model?: string
  opus_model?: string
  timeout_ms?: number
  max_output_tokens?: number
  disable_traffic?: boolean
  wire_api?: 'chat' | 'responses'
}

export interface PermissionConfig {
  mode?: string
  allowed_tools?: string[]
  disallowed_tools?: string[]
  tools?: string[]
  skip_all?: boolean
  sandbox_mode?: 'read-only' | 'workspace-write' | 'danger-full-access'
  approval_policy?: 'untrusted' | 'on-failure' | 'on-request' | 'never'
  full_auto?: boolean
  additional_dirs?: string[]
}

export interface FeatureConfig {
  web_search?: boolean
}

// Session Types
export interface Session {
  id: string
  agent_id?: string
  agent: string
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

export interface CreateSessionRequest {
  agent_id?: string
  agent?: string
  workspace: string
  env?: Record<string, string>
}

export interface ExecRequest {
  prompt: string
  max_turns?: number
  timeout?: number
  allowed_tools?: string[]
  disallowed_tools?: string[]
  include_events?: boolean
}

export interface ExecResponse {
  execution_id: string
  message: string
  output: string
  events?: ExecEvent[]
  usage?: TokenUsage
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
  is_configured: boolean
  created_at: string
  updated_at: string
}

export interface MCPTestResult {
  status: 'ok' | 'error'
  latency_ms: number
  server_info?: Record<string, unknown>
  capabilities?: Record<string, unknown>
  error?: string
}

export interface MCPServerStats {
  total: number
  enabled: number
  configured: number
  not_configured: number
  built_in: number
  custom: number
  by_category: Record<string, number>
  by_type: Record<string, number>
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
export type SkillOrigin = 'extra' | 'bundled' | 'managed' | 'workspace'
export type SkillLoadLevel = 'metadata' | 'body' | 'full'

export interface SkillFile {
  path: string
  content: string
}

export interface SkillRequirements {
  bins?: string[]
  any_bins?: string[]
  env?: string[]
  config?: string[]
  pip?: string[]
  npm?: string[]
  os?: string[]
}

// 安装方式
export type InstallKind = 'brew' | 'node' | 'go' | 'uv' | 'pip' | 'download'

// 安装规范
export interface InstallSpec {
  id?: string
  kind: InstallKind
  label?: string
  bins?: string[]
  formula?: string
  package?: string
  module?: string
  url?: string
  os?: string[]
}

// Skill 配置（用户可配置）
export interface SkillConfig {
  enabled: boolean
  api_key?: string
  env?: Record<string, string>
}

export interface SkillInvocationPolicy {
  user_invocable?: boolean
  auto_invocable?: boolean
  hook_invocable?: string[]
}

export interface SkillRuntimeConfig {
  python?: string
  node?: string
  memory?: string
  timeout?: string
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
  source?: SkillOrigin
  source_path?: string
  source_dir?: string
  is_built_in: boolean
  is_enabled: boolean
  // 渐进式加载
  load_level?: SkillLoadLevel
  body_loaded?: boolean
  // 依赖与运行时
  requirements?: SkillRequirements
  runtime?: SkillRuntimeConfig
  invocation?: SkillInvocationPolicy
  install?: InstallSpec[]
  primary_env?: string
  always?: boolean
  emoji?: string
  homepage?: string
  created_at: string
  updated_at: string
}

// Skill 元数据（轻量级，用于列表）
export interface SkillMetadata {
  id: string
  name: string
  description?: string
  command: string
  category: SkillCategory
  tags?: string[]
  author?: string
  version?: string
  source: SkillOrigin
  is_built_in: boolean
  is_enabled: boolean
  has_deps: boolean
  updated_at: string
}

// 依赖检查结果
export interface SkillCheckResult {
  skill_id: string
  satisfied: boolean
  missing?: {
    bins?: string[]
    any_bins?: string[]
    env?: string[]
    config?: string[]
    pip?: string[]
    npm?: string[]
    os?: string[]
  }
  error?: string
}

// 配置检查结果
export interface ConfigCheck {
  path: string
  value: unknown
  satisfied: boolean
}

// 安装选项
export interface InstallOption {
  id: string
  kind: InstallKind
  label: string
  bins?: string[]
}

// Skill 状态条目（完整状态报告）
export interface SkillStatusEntry {
  name: string
  description: string
  source: string
  file_path: string
  skill_key: string
  primary_env?: string
  emoji?: string
  homepage?: string
  always: boolean
  disabled: boolean
  blocked_by_allowlist: boolean
  eligible: boolean
  requirements: {
    bins?: string[]
    any_bins?: string[]
    env?: string[]
    config?: string[]
    os?: string[]
  }
  missing: {
    bins?: string[]
    any_bins?: string[]
    env?: string[]
    config?: string[]
    os?: string[]
  }
  config_checks?: ConfigCheck[]
  install?: InstallOption[]
}

// Skill 状态报告
export interface SkillStatusReport {
  workspace_dir: string
  managed_skills_dir: string
  skills: SkillStatusEntry[]
}

// Skill 统计信息
export interface SkillStats {
  total: number
  extra: number
  bundled: number
  managed: number
  workspace: number
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
  source_dir?: string
  requirements?: SkillRequirements
  runtime?: SkillRuntimeConfig
  invocation?: SkillInvocationPolicy
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
  requirements?: SkillRequirements
  runtime?: SkillRuntimeConfig
  invocation?: SkillInvocationPolicy
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

// Task Types (Task-Centric, 对齐 Manus)
export type TaskStatus = 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface TaskResult {
  summary?: string
  text?: string
  files?: OutputFile[]
  usage?: {
    duration_seconds: number
    input_tokens?: number
    output_tokens?: number
    total_tokens?: number
  }
  logs?: string
}

export interface OutputFile {
  name: string
  path: string
  size: number
  mime_type?: string
  url?: string
}

export interface Turn {
  id: string
  prompt: string
  result?: TaskResult
  created_at: string
}

export interface Task {
  id: string
  agent_id: string
  agent_name?: string
  agent_type?: string
  prompt: string
  attachments?: string[]
  output_files?: OutputFile[]
  turns?: Turn[]
  turn_count: number
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
  agent_id?: string
  prompt: string
  task_id?: string          // 多轮时传入
  attachments?: string[]    // file IDs
  webhook_url?: string
  timeout?: number
  metadata?: Record<string, string>
}

export interface TaskEvent {
  type: string
  data?: unknown
}

export interface TaskStats {
  total: number
  by_status: Record<TaskStatus, number>
  by_agent: Record<string, number>
  avg_duration_seconds: number
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

// File Types (Public API)
export type FilePurpose = 'attachment' | 'output' | 'general'

export interface UploadedFile {
  id: string
  name: string
  size: number
  mime_type: string
  task_id?: string
  purpose: FilePurpose
  uploaded_at: string
  expires_at?: string
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

export interface RunningBatchInfo {
  id: string
  name: string
  workers: number
  completed: number
  failed: number
  total: number
  percent: number
  tasks_per_sec: number
}

export interface BatchPoolStats {
  max_batches: number
  running_batches: number
  total_workers: number
  busy_workers: number
  idle_workers: number
  batches: RunningBatchInfo[]
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
  batches?: BatchPoolStats
  system: {
    uptime: string
    memory_usage_mb: number
    go_version: string
    num_cpu: number
    num_goroutines: number
  }
}

export interface GCStats {
  running: boolean
  last_run_at: string
  next_run_at: string
  containers_removed: number
  total_runs: number
  errors: string[]
  config: {
    interval_seconds: number
    container_ttl_seconds: number
    idle_timeout_seconds: number
  }
}

export interface GCCandidate {
  container_id: string
  name: string
  image: string
  status: string
  created_at: number
  reason: string
}

export interface UpdateGCConfigRequest {
  interval_seconds?: number
  container_ttl_seconds?: number
  idle_timeout_seconds?: number
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

// History Types
export type HistorySourceType = 'session' | 'agent'
export type HistoryStatus = 'pending' | 'running' | 'completed' | 'failed'

export interface HistoryEntry {
  id: string
  source_type: HistorySourceType
  source_id: string
  source_name: string
  engine?: string
  prompt: string
  status: HistoryStatus
  output?: string
  error?: string
  exit_code: number
  usage?: {
    input_tokens: number
    cached_input_tokens?: number
    output_tokens: number
  }
  metadata?: Record<string, string>
  started_at: string
  ended_at?: string
}

export interface HistoryStats {
  total_executions: number
  completed_count: number
  failed_count: number
  total_input_tokens: number
  total_output_tokens: number
  by_source: Record<string, number>
  by_engine: Record<string, number>
}

export interface HistoryListResponse {
  entries: HistoryEntry[]
  total: number
  limit: number
  offset: number
}

export interface HistoryFilter {
  source_type?: HistorySourceType
  source_id?: string
  agent_id?: string
  engine?: string
  status?: HistoryStatus
  limit?: number
  offset?: number
}

// Dashboard Types (态势感知大屏)
export interface DashboardAgentDetail {
  id: string
  name: string
  adapter: string
  model: string
  status: string
  running: number
  queued: number
  completed: number
  failed: number
}

export interface DashboardProviderInfo {
  id: string
  name: string
  status: 'online' | 'offline' | 'degraded'
  is_configured: boolean
  is_valid: boolean
  category: string
  icon?: string
  icon_color?: string
}

export interface DashboardRecentTask {
  id: string
  agent_id: string
  agent_name: string
  adapter: string
  prompt: string
  status: TaskStatus
  duration_seconds: number
  created_at: string
}

export interface DashboardStats {
  agents: {
    total: number
    active: number
    by_adapter: Record<string, number>
    details: DashboardAgentDetail[]
  }
  tasks: {
    total: number
    today: number
    by_status: Record<string, number>
    avg_duration_seconds: number
    success_rate: number
  }
  sessions: {
    total: number
    running: number
    creating: number
    stopped: number
    error: number
  }
  tokens: {
    total_input: number
    total_output: number
    total_tokens: number
  }
  containers: {
    total: number
    running: number
    stopped: number
  }
  providers: DashboardProviderInfo[]
  system: {
    uptime: string
    started_at: string
  }
  recent_tasks: DashboardRecentTask[]
}

// API Response
export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

// Batch Types (批量任务处理)
export type BatchStatus = 'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'cancelled'
export type BatchTaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'dead'

export interface BatchTemplate {
  prompt_template: string
  timeout: number
  max_retries: number
  runtime_id?: string
}

export interface WorkerInfo {
  id: string
  session_id: string
  container_id?: string
  status: string  // idle, busy, error, stopped
  current_task?: string
  completed: number
  last_error?: string
}

export interface Batch {
  id: string
  name: string
  agent_id: string
  template: BatchTemplate
  concurrency: number
  status: BatchStatus
  total_tasks: number
  completed: number
  failed: number
  progress_percent?: number
  estimated_eta?: string
  tasks_per_sec?: number
  workers?: WorkerInfo[]
  error_summary?: Record<string, number>
  created_at: string
  started_at?: string
  completed_at?: string
}

export interface BatchTask {
  id: string
  batch_id: string
  index: number
  input: Record<string, unknown>
  prompt?: string
  status: BatchTaskStatus
  worker_id?: string
  result?: string
  error?: string
  attempts: number
  claimed_at?: string
  claimed_by?: string
  dead_at?: string
  dead_reason?: string
  created_at: string
  started_at?: string
  duration_ms?: number
}

export interface CreateBatchRequest {
  name: string
  agent_id: string
  prompt_template: string
  inputs: Record<string, unknown>[]
  concurrency?: number
  timeout?: number
  max_retries?: number
  runtime_id?: string
  auto_start?: boolean
}

export interface BatchStats {
  total_tasks: number
  pending: number
  running: number
  completed: number
  failed: number
  dead?: number
  by_worker?: Record<string, number>
  avg_duration_ms: number
  error_types?: Record<string, number>
}

export interface BatchEvent {
  type: string
  batch_id: string
  timestamp: string
  data?: unknown
}

export interface BatchProgressData {
  completed: number
  failed: number
  total: number
  percent: number
  eta: string
  tasks_per_sec: number
}

export interface ListBatchFilter {
  status?: BatchStatus
  agent_id?: string
  limit?: number
  offset?: number
}

export interface ListBatchTaskFilter {
  status?: BatchTaskStatus
  worker_id?: string
  limit?: number
  offset?: number
}

// ==================== Settings Types ====================

export interface Settings {
  agent: AgentSettings
  task: TaskSettings
  batch: BatchSettings
  storage: StorageSettings
  notify: NotifySettings
}

export interface AgentSettings {
  default_provider_id: string
  default_model: string
  default_runtime_id: string
  default_timeout: number
  system_prompt: string
}

export interface TaskSettings {
  default_idle_timeout: number
  default_poll_interval: number
  max_turns: number
  max_attachments: number
  max_attachment_size: number
}

export interface BatchSettings {
  default_workers: number
  max_workers: number
  max_concurrent_batches: number
  max_retries: number
  retry_delay: number
  dead_letter_enabled: boolean
}

export interface StorageSettings {
  history_retention_days: number
  session_retention_days: number
  auto_cleanup: boolean
}

export interface NotifySettings {
  webhook_url: string
  webhook_secret: string
  notify_on_complete: boolean
  notify_on_failed: boolean
  notify_on_batch_complete: boolean
}

// ==================== Auth Profile Types (API Key Rotation) ====================

export interface AuthProfile {
  id: string
  provider_id: string
  key_masked: string          // 脱敏的 Key（显示用）
  priority: number            // 优先级（0=最高）
  is_enabled: boolean
  cooldown_until?: string     // 冷却结束时间
  fail_count: number
  success_count: number
  last_used_at?: string
  created_at: string
  updated_at: string
}

export interface RotatorStats {
  total_profiles: number
  active_profiles: number
  all_in_cooldown: boolean
  next_available_in: string
}

export interface AddAuthProfileRequest {
  api_key: string
  priority?: number
}

// ==================== Feishu Channel Types ====================

export interface FeishuConfig {
  id: string
  name: string
  app_id: string
  app_secret: string           // 脱敏显示
  verification_token?: string  // 脱敏显示
  encrypt_key?: string         // 脱敏显示
  bot_name?: string
  default_agent_id?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateFeishuConfigRequest {
  name: string
  app_id: string
  app_secret: string
  verification_token?: string
  encrypt_key?: string
  bot_name?: string
  default_agent_id?: string
}

export interface UpdateFeishuConfigRequest {
  id: string
  name?: string
  app_id?: string
  app_secret?: string          // 可选，不提供则保持原值
  verification_token?: string
  encrypt_key?: string
  bot_name?: string
  default_agent_id?: string
}

export interface FeishuMessageLog {
  id: string
  chat_id: string
  sender_id: string
  sender_name: string
  content: string
  message_type: string
  reply_id?: string
  task_id?: string
  received_at: string
  created_at: string
}

// Chat Types (Chat 界面)
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  turnId?: string
  timestamp: Date
  status?: 'sending' | 'sent' | 'streaming' | 'error'
}

export interface MessageGroup {
  role: 'user' | 'assistant' | 'system'
  messages: ChatMessage[]
}

// ==================== Channel Session & Message Types ====================

export type ChannelSessionStatus = 'active' | 'completed' | 'expired'

export interface ChannelSession {
  id: string
  channel_type: string
  chat_id: string
  user_id: string
  user_name: string
  is_group: boolean
  task_id: string
  agent_id: string
  agent_name: string
  status: ChannelSessionStatus
  message_count: number
  last_message_at?: string
  created_at: string
  updated_at: string
}

export type ChannelMessageDirection = 'inbound' | 'outbound'

export interface ChannelMessage {
  id: string
  session_id: string
  channel_type: string
  chat_id: string
  sender_id: string
  sender_name: string
  content: string
  direction: ChannelMessageDirection
  task_id: string
  turn_id?: string
  status: string
  metadata?: Record<string, string>
  received_at: string
  created_at: string
}

export interface ChannelStat {
  sessions: number
  messages: number
}

export interface ChannelStats {
  total_sessions: number
  active_sessions: number
  total_messages: number
  messages_today: number
  by_channel: Record<string, ChannelStat>
}

export interface ChannelSessionFilter {
  channel_type?: string
  status?: ChannelSessionStatus
  agent_id?: string
  limit?: number
  offset?: number
}

export interface ChannelMessageFilter {
  channel_type?: string
  direction?: ChannelMessageDirection
  task_id?: string
  limit?: number
  offset?: number
}
