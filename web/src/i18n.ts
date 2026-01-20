export type Language = 'en' | 'zh'

export const translations = {
  en: {
    // Header
    appName: 'AgentBox',
    tagline: 'AI Agent Container Platform',

    // Navigation
    dashboard: 'Dashboard',
    sessions: 'Sessions',
    agents: 'Agents',
    settings: 'Settings',

    // Dashboard
    overview: 'Overview',
    totalSessions: 'Total Sessions',
    runningSessions: 'Running',
    stoppedSessions: 'Stopped',
    availableAgents: 'Available Agents',
    successRate: 'Success Rate',

    // Sessions
    sessionList: 'Session List',
    newSession: 'New Session',
    noSessions: 'No sessions yet. Create one to get started.',
    noSessionsFiltered: 'No sessions found. Try adjusting your search or filters.',
    sessionId: 'Session ID',
    agent: 'Agent',
    status: 'Status',
    workspace: 'Workspace',
    created: 'Created',
    lastActivity: 'Last Activity',
    actions: 'Actions',
    searchSessions: 'Search sessions...',
    allStatus: 'All Status',

    // Status
    running: 'Running',
    stopped: 'Stopped',
    creating: 'Creating',
    error: 'Error',

    // Actions
    start: 'Start',
    stop: 'Stop',
    delete: 'Delete',
    refresh: 'Refresh',
    viewLogs: 'View Logs',
    execute: 'Execute',
    reconnect: 'Reconnect',

    // Create Modal
    createSession: 'Create Session',
    selectAgent: 'Select Agent',
    workspacePath: 'Workspace Path',
    workspacePathPlaceholder: '/path/to/project',
    apiKey: 'API Key',
    apiKeyPlaceholder: 'sk-...',
    cancel: 'Cancel',
    create: 'Create',
    creating_: 'Creating...',
    continue_: 'Continue',
    back: 'Back',
    step: 'Step',
    sessionName: 'Session Name',
    sessionNamePlaceholder: 'my-awesome-project',
    sessionNameHint: 'Leave empty for auto-generated ID',
    configureEnv: 'Configure Environment',

    // Agents
    supportedAgents: 'Supported Agents',
    requiredEnv: 'Required',
    image: 'Image',
    agentDescription: 'Description',

    // Messages
    confirmDelete: 'Are you sure you want to delete this session?',
    loading: 'Loading...',
    error_: 'Error',
    success: 'Success',
    justNow: 'Just now',
    minutesAgo: 'm ago',
    hoursAgo: 'h ago',
    daysAgo: 'd ago',
    never: 'Never',

    // Footer
    documentation: 'Documentation',

    // Language
    language: 'Language',
    english: 'English',
    chinese: '中文',

    // Theme
    theme: 'Theme',
    dark: 'Dark',
    light: 'Light',
    darkTheme: 'Dark theme',
    lightTheme: 'Light theme',

    // Sidebar Navigation
    quickStart: 'Quick Start',
    profiles: 'Profiles',
    tasks: 'Tasks',
    webhooks: 'Webhooks',
    explore: 'Explore',
    mcpServers: 'MCP Servers',
    skills: 'Skills',
    credentials: 'Credentials',
    images: 'Images',
    system: 'System',

    // Explore Page
    featured: 'Featured',
    mcpMarketplace: 'MCP Servers',
    skillMarketplace: 'Skills',
    profileTemplates: 'Profiles',
    applications: 'Applications',
    community: 'Community',
    tutorials: 'Tutorials',
    leaderboard: 'Leaderboard',
    myCreations: 'My Creations',

    // Actions
    newTask: 'New Task',
    newMCP: 'New MCP',
    newSkill: 'New Skill',
    newCredential: 'New Credential',

    // Dashboard
    activityLog: 'Activity Log',
    recentEvents: 'Recent events',
    available: 'available',
    noAgentsAvailable: 'No agents available',

    // Placeholders
    enterName: 'Enter name',
    enterDescription: 'Enter description',
    enterPrompt: 'Enter prompt',
    addTag: 'Add a tag...',
    addArgument: 'Add an argument...',
    describeTask: 'Describe what you want the agent to do...',
    fileContent: 'File content...',
    yourName: 'Your name',
    profileNamePlaceholder: 'Enter profile name',
    profileDescPlaceholder: 'Describe what this profile is for',
    systemPromptPlaceholder: 'Enter system prompt for the agent...',

    // Buttons & Labels
    verify: 'Verify',
    testConnection: 'Test Connection',
    exportSkill: 'Export as SKILL.md',
    advancedConfig: 'Advanced Model Configuration',

    // Common
    name: 'Name',
    description: 'Description',
    tags: 'Tags',
    author: 'Author',
    profile: 'Profile',
    prompt: 'Prompt',
    createTask: 'Create Task',
    createNewTask: 'Create New Task',
    logs: 'Logs',
    noLogsAvailable: 'No logs available',

    // Actions & Buttons
    add: 'Add',
    edit: 'Edit',
    save: 'Save',
    saving: 'Saving',
    clone: 'Clone',
    export: 'Export',
    import: 'Import',
    copy: 'Copy',
    copyUrl: 'Copy URL',

    // Filter Options
    all: 'All',
    enabled: 'Enabled',
    disabled: 'Disabled',

    // Scopes
    global: 'Global',
    session: 'Session',

    // Providers
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    github: 'GitHub',
    custom: 'Custom',
    allScopes: 'All Scopes',
    allProviders: 'All Providers',

    // Credential Types
    token: 'Token',

    // Section Titles
    basicInformation: 'Basic Information',
    systemPrompt: 'System Prompt',
    permissions: 'Permissions',
    resourceLimits: 'Resource Limits',

    // Field Labels
    command: 'Command',
    arguments: 'Arguments',
    workDir: 'Working Directory',
    environment: 'Environment Variables',
    category: 'Category',
    type: 'Type',
    version: 'Version',
    model: 'Model',
    adapter: 'Adapter',
    apiBaseUrl: 'API Base URL',
    optional: 'Optional',
    required: 'Required',

    // Messages
    leaveEmptyDefault: 'Leave empty to use the default API endpoint',
    useForProxies: 'Use for proxies or compatible APIs',
    uniqueIdentifier: 'Unique identifier. Cannot be changed later',

    // Credential
    credentialType: 'Credential Type',
    scope: 'Scope',
    provider: 'Provider',
    keyValue: 'Key/Value',

    // Permission Details
    mode: 'Mode',
    allowedTools: 'Allowed Tools',
    sandboxMode: 'Sandbox Mode',
    approvalPolicy: 'Approval Policy',

    // Resource Details
    maxBudget: 'Max Budget (USD)',
    unlimited: 'Unlimited',
    maxTurns: 'Max Turns',
    cpus: 'CPUs',
    memoryMB: 'Memory (MB)',
    default: 'Default',

    // Profile List
    newProfile: 'New Profile',
    agentProfiles: 'Agent Profiles',
    profilesDescription: 'Profiles are pre-configured templates that combine adapter settings, model selection, MCP servers, and permissions. Use them to quickly create sessions with your preferred setup.',
    noProfilesFound: 'No profiles found',
    createFirstProfile: 'Create your first profile to get started',
    failedToLoadProfiles: 'Failed to load profiles',
    claudeCode: 'Claude Code',
    codex: 'Codex',
    opencode: 'OpenCode',
    other: 'Other',

    // MCP Server Descriptions
    mcpServersDescription: 'Select MCP servers to enable for this profile. These servers will be started automatically when a session using this profile is created.',

    // Skills Descriptions
    skillsDescription: 'Select skills available for this profile. Skills provide reusable task templates that can be invoked with commands like /commit.',

    // Credentials Descriptions
    credentialsDescription: 'Select credentials to inject into sessions using this profile. Credentials are securely stored and will be available as environment variables.',

    // MCP Server List
    mcpServersTitle: 'MCP Servers',
    mcpServersDesc: 'Model Context Protocol (MCP) servers extend Agent capabilities with external tools, data sources, and integrations. Configure and manage your MCP servers here.',
    noMCPServersFound: 'No MCP servers found',
    tryChangingFilter: 'Try changing the filter or create a new server',
    createFirstMCPServer: 'Create your first MCP server to get started',
    failedToLoadMCPServers: 'Failed to load MCP servers',

    // Skill List
    skillsTitle: 'Skills',
    skillsDesc: 'Skills are reusable task templates that define how agents should handle specific tasks. Use commands like /commit or /review-pr to invoke them.',
    noSkillsFound: 'No skills found',
    tryChangingFilterOrCreateSkill: 'Try changing the filter or create a new skill',
    createFirstSkill: 'Create your first skill to get started',
    failedToLoadSkills: 'Failed to fetch skills',

    // Credential List
    credentialsTitle: 'Credentials',
    credentialsListDesc: 'Securely manage API keys and tokens for AI providers and services. Credentials are encrypted at rest and can be scoped to global, profile, or session level.',
    securityNotice: 'Security Notice',
    securityNoticeDesc: 'Credentials are stored with AES-256 encryption. Only masked values are shown in the UI. The actual values are only decrypted when injected into agent sessions.',
    noCredentialsFound: 'No credentials found',
    tryChangingFilterOrAddCredential: 'Try changing the filters or add a new credential',
    addFirstAPIKey: 'Add your first API key to get started',
    failedToLoadCredentials: 'Failed to load credentials',

    // Status badges
    valid: 'Valid',
    invalid: 'Invalid',
    builtIn: 'Built-in',

    // Image List
    imagesTitle: 'Images',
    imagesCount: 'images',
    totalSize: 'total',
    pullImage: 'Pull Image',
    pulling: 'Pulling...',
    pull: 'Pull',
    enterFullImageName: 'Enter full image name with registry and tag',
    filter: 'Filter',
    allImages: 'All Images',
    agentImagesOnly: 'Agent Images Only',
    noImagesFound: 'No images found',
    pullImageToGetStarted: 'Pull an image to get started',
    failedToFetchImages: 'Failed to fetch images',
    confirmDeleteImage: 'Are you sure you want to delete this image?',
    imagePlaceholder: 'e.g. anthropic/claude-code:latest',
    size: 'Size',

    // Settings
    settingsTitle: 'Settings',
    saved: 'Saved',
    apiKeys: 'API Keys',
    apiKeysDesc: 'Configure default API keys for agents. These will be used when creating new sessions.',
    anthropicApiKey: 'Anthropic API Key',
    anthropicApiKeyDesc: 'Used for Claude Code agent',
    openaiApiKey: 'OpenAI API Key',
    openaiApiKeyDesc: 'Used for Codex agent',
    defaultWorkspace: 'Default Workspace',
    defaultWorkspaceDesc: 'Default path to mount in container when creating new sessions',
    systemStatus: 'System Status',
    backendStatus: 'Backend Status',
    healthy: 'Healthy',
    unhealthy: 'Unhealthy',
    checking: 'Checking...',
    serverStatus: 'Server Status',
    connected: 'Connected',
    connectionFailed: 'Connection failed',
    defaultLanguage: 'Default language',
    chineseLanguage: 'Chinese',
    about: 'About',
    agentBoxTagline: 'AI Agent Container Platform',
    agentBoxDescription: 'Open-source solution for running AI agents in isolated containers.',
    githubRepository: 'GitHub Repository',

    // Task List
    tasksTitle: 'Tasks',
    allTasks: 'All',
    queuedTasks: 'Queued',
    completedTasks: 'Completed',
    failedTasks: 'Failed',
    duration: 'Duration',
    cost: 'Cost',
    promptText: 'Prompt',
    confirmCancelTask: 'Cancel task',
    failedToLoadTasks: 'Failed to load tasks',
    noTasksFound: 'No tasks found',
    createTaskToGetStarted: 'Create a task to get started',

    // Task Status
    pending: 'Pending',
    queued: 'Queued',
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',

    // Webhook List
    webhooksTitle: 'Webhooks',
    webhooksDesc: 'Configure webhooks to receive real-time notifications about events in your AgentBox instance.',
    newWebhook: 'New Webhook',
    noWebhooksFound: 'No webhooks found',
    createWebhookToGetStarted: 'Create a webhook to get started',
    failedToLoadWebhooks: 'Failed to load webhooks',
    signed: 'Signed',
    events: 'Events',
    lastTriggered: 'Last Triggered',
    confirmDeleteWebhook: 'Delete webhook',
    createWebhook: 'Create Webhook',
    editWebhook: 'Edit Webhook',
    webhookURL: 'Webhook URL',
    secretDesc: 'Used for HMAC-SHA256 signature',
    unchanged: 'unchanged',
    enterSecretPlaceholder: 'Enter secret for signing',
    selectAtLeastOneEvent: 'Please select at least one event',
    totalWebhooks: 'Total Webhooks',
    copyURL: 'Copy URL',

    // Event Types
    taskCreated: 'Task Created',
    taskCreatedDesc: 'When a new task is created',
    taskCompleted: 'Task Completed',
    taskCompletedDesc: 'When a task completes successfully',
    taskFailed: 'Task Failed',
    taskFailedDesc: 'When a task fails',
    sessionStarted: 'Session Started',
    sessionStartedDesc: 'When a session container starts',
    sessionStopped: 'Session Stopped',
    sessionStoppedDesc: 'When a session container stops',

    // System Maintenance
    systemMaintenance: 'System Maintenance',
    healthChecksStatsCleanup: 'Health checks, stats & cleanup',
    failedToFetchSystemData: 'Failed to fetch system data',
    systemHealth: 'System Health',
    overallStatus: 'Overall Status',
    uptime: 'Uptime',
    containerRuntime: 'Container Runtime',
    containers: 'Containers',
    resources: 'Resources',
    serverInfo: 'Server Info',
    memory: 'Memory',
    cpuCores: 'CPU Cores',
    goroutines: 'Goroutines',
    healthChecks: 'Health Checks',
    goVersion: 'Go Version',
    cleanupActions: 'Cleanup Actions',
    orphanContainersTitle: 'Orphan Containers',
    orphanContainersDesc: 'Remove containers that are managed by AgentBox but no longer have an associated session.',
    cleanupContainers: 'Cleanup Containers',
    cleaning: 'Cleaning...',
    unusedImagesTitle: 'Unused Images',
    unusedImagesDesc: 'Remove images that are not being used by any container (excludes Agent images).',
    cleanupImages: 'Cleanup Images',
    removed: 'Removed',
    orphanContainers: 'orphan container(s)',
    noOrphanContainersFound: 'No orphan containers found',
    failedToCleanupContainers: 'Failed to cleanup containers',
    freed: 'freed',
    noUnusedImagesToRemove: 'No unused images to remove',
    failedToCleanupImages: 'Failed to cleanup images',
  },
  zh: {
    // Header
    appName: 'AgentBox',
    tagline: 'AI Agent 容器化平台',

    // Navigation
    dashboard: '仪表盘',
    sessions: '会话',
    agents: 'Agent',
    settings: '设置',

    // Dashboard
    overview: '概览',
    totalSessions: '总会话数',
    runningSessions: '运行中',
    stoppedSessions: '已停止',
    availableAgents: '可用 Agent',
    successRate: '成功率',

    // Sessions
    sessionList: '会话列表',
    newSession: '新建会话',
    noSessions: '暂无会话，点击新建开始使用。',
    noSessionsFiltered: '未找到会话，请调整搜索条件。',
    sessionId: '会话 ID',
    agent: 'Agent',
    status: '状态',
    workspace: '工作目录',
    created: '创建时间',
    lastActivity: '最近活动',
    actions: '操作',
    searchSessions: '搜索会话...',
    allStatus: '全部状态',

    // Status
    running: '运行中',
    stopped: '已停止',
    creating: '创建中',
    error: '错误',

    // Actions
    start: '启动',
    stop: '停止',
    delete: '删除',
    refresh: '刷新',
    viewLogs: '查看日志',
    execute: '执行',
    reconnect: '重连',

    // Create Modal
    createSession: '创建会话',
    selectAgent: '选择 Agent',
    workspacePath: '工作目录路径',
    workspacePathPlaceholder: '/path/to/project',
    apiKey: 'API 密钥',
    apiKeyPlaceholder: 'sk-...',
    cancel: '取消',
    create: '创建',
    creating_: '创建中...',
    continue_: '继续',
    back: '返回',
    step: '步骤',
    sessionName: '会话名称',
    sessionNamePlaceholder: 'my-awesome-project',
    sessionNameHint: '留空将自动生成',
    configureEnv: '配置环境',

    // Agents
    supportedAgents: '支持的 Agent',
    requiredEnv: '需要配置',
    image: '镜像',
    agentDescription: '描述',

    // Messages
    confirmDelete: '确定要删除此会话吗？',
    loading: '加载中...',
    error_: '错误',
    success: '成功',
    justNow: '刚刚',
    minutesAgo: '分钟前',
    hoursAgo: '小时前',
    daysAgo: '天前',
    never: '从未',

    // Footer
    documentation: '文档',

    // Language
    language: '语言',
    english: 'English',
    chinese: '中文',

    // Theme
    theme: '主题',
    dark: '深色',
    light: '浅色',
    darkTheme: '深色主题',
    lightTheme: '浅色主题',

    // Sidebar Navigation
    quickStart: '快速开始',
    profiles: 'Profile配置',
    tasks: '任务',
    webhooks: 'Webhook',
    explore: '探索',
    mcpServers: 'MCP服务器',
    skills: '技能',
    credentials: '凭证',
    images: '镜像',
    system: '系统',

    // Explore Page
    featured: '精选',
    mcpMarketplace: 'MCP服务器',
    skillMarketplace: '技能',
    profileTemplates: 'Profile模板',
    applications: '应用',
    community: '社区',
    tutorials: '教程',
    leaderboard: '排行榜',
    myCreations: '我的创作',

    // Actions
    newTask: '新建任务',
    newMCP: '新建MCP',
    newSkill: '新建技能',
    newCredential: '新建凭证',

    // Dashboard
    activityLog: '活动日志',
    recentEvents: '最近事件',
    available: '可用',
    noAgentsAvailable: '暂无可用 Agent',

    // Placeholders
    enterName: '输入名称',
    enterDescription: '输入描述',
    enterPrompt: '输入提示词',
    addTag: '添加标签...',
    addArgument: '添加参数...',
    describeTask: '描述您希望 Agent 执行的任务...',
    fileContent: '文件内容...',
    yourName: '您的名字',
    profileNamePlaceholder: '输入 Profile 名称',
    profileDescPlaceholder: '描述此 Profile 的用途',
    systemPromptPlaceholder: '输入 Agent 的系统提示词...',

    // Buttons & Labels
    verify: '验证',
    testConnection: '测试连接',
    exportSkill: '导出为 SKILL.md',
    advancedConfig: '高级模型配置',

    // Common
    name: '名称',
    description: '描述',
    tags: '标签',
    author: '作者',
    profile: 'Profile',
    prompt: '提示词',
    createTask: '创建任务',
    createNewTask: '创建新任务',
    logs: '日志',
    noLogsAvailable: '暂无日志',

    // Actions & Buttons
    add: '添加',
    edit: '编辑',
    save: '保存',
    saving: '保存中',
    clone: '克隆',
    export: '导出',
    import: '导入',
    copy: '复制',
    copyUrl: '复制链接',

    // Filter Options
    all: '全部',
    enabled: '已启用',
    disabled: '已禁用',

    // Scopes
    global: '全局',
    session: '会话',

    // Providers
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    github: 'GitHub',
    custom: '自定义',
    allScopes: '所有作用域',
    allProviders: '所有提供商',

    // Credential Types
    token: '令牌',

    // Section Titles
    basicInformation: '基本信息',
    systemPrompt: '系统提示词',
    permissions: '权限',
    resourceLimits: '资源限制',

    // Field Labels
    command: '命令',
    arguments: '参数',
    workDir: '工作目录',
    environment: '环境变量',
    category: '分类',
    type: '类型',
    version: '版本',
    model: '模型',
    adapter: '适配器',
    apiBaseUrl: 'API 基础 URL',
    optional: '可选',
    required: '必填',

    // Messages
    leaveEmptyDefault: '留空将使用默认 API 端点',
    useForProxies: '可用于代理或兼容的 API',
    uniqueIdentifier: '唯一标识符，创建后不可更改',

    // Credential
    credentialType: '凭证类型',
    scope: '作用域',
    provider: '提供商',
    keyValue: '键/值',

    // Permission Details
    mode: '模式',
    allowedTools: '允许的工具',
    sandboxMode: '沙箱模式',
    approvalPolicy: '审批策略',

    // Resource Details
    maxBudget: '最大预算（美元）',
    unlimited: '无限制',
    maxTurns: '最大轮次',
    cpus: 'CPU数量',
    memoryMB: '内存（MB）',
    default: '默认',

    // Profile List
    newProfile: '新建Profile',
    agentProfiles: 'Agent Profile配置',
    profilesDescription: 'Profile是预配置的模板，结合了适配器设置、模型选择、MCP服务器和权限配置。使用它们可以快速创建具有您偏好设置的会话。',
    noProfilesFound: '未找到Profile',
    createFirstProfile: '创建您的第一个Profile以开始使用',
    failedToLoadProfiles: '加载Profile失败',
    claudeCode: 'Claude Code',
    codex: 'Codex',
    opencode: 'OpenCode',
    other: '其他',

    // MCP Server Descriptions
    mcpServersDescription: '选择要为此Profile启用的MCP服务器。当使用此Profile创建会话时，这些服务器将自动启动。',

    // Skills Descriptions
    skillsDescription: '选择此Profile可用的技能。技能提供可重用的任务模板，可以通过 /commit 等命令调用。',

    // Credentials Descriptions
    credentialsDescription: '选择要注入到使用此Profile的会话中的凭证。凭证将被安全存储，并作为环境变量提供。',

    // MCP Server List
    mcpServersTitle: 'MCP服务器',
    mcpServersDesc: '模型上下文协议（MCP）服务器通过外部工具、数据源和集成扩展Agent的能力。在此配置和管理您的MCP服务器。',
    noMCPServersFound: '未找到MCP服务器',
    tryChangingFilter: '尝试更改筛选条件或创建新服务器',
    createFirstMCPServer: '创建您的第一个MCP服务器以开始使用',
    failedToLoadMCPServers: '加载MCP服务器失败',

    // Skill List
    skillsTitle: '技能',
    skillsDesc: '技能是可重用的任务模板，定义了Agent如何处理特定任务。使用 /commit 或 /review-pr 等命令来调用它们。',
    noSkillsFound: '未找到技能',
    tryChangingFilterOrCreateSkill: '尝试更改筛选条件或创建新技能',
    createFirstSkill: '创建您的第一个技能以开始使用',
    failedToLoadSkills: '获取技能失败',

    // Credential List
    credentialsTitle: '凭证',
    credentialsListDesc: '安全管理AI提供商和服务的API密钥和令牌。凭证在静态时加密，可以设置为全局、Profile或会话级别的作用域。',
    securityNotice: '安全提示',
    securityNoticeDesc: '凭证使用AES-256加密存储。用户界面中仅显示掩码值。实际值仅在注入到Agent会话时才会解密。',
    noCredentialsFound: '未找到凭证',
    tryChangingFilterOrAddCredential: '尝试更改筛选条件或添加新凭证',
    addFirstAPIKey: '添加您的第一个API密钥以开始使用',
    failedToLoadCredentials: '加载凭证失败',

    // Status badges
    valid: '有效',
    invalid: '无效',
    builtIn: '内置',

    // Image List
    imagesTitle: '镜像',
    imagesCount: '个镜像',
    totalSize: '总大小',
    pullImage: '拉取镜像',
    pulling: '拉取中...',
    pull: '拉取',
    enterFullImageName: '输入完整的镜像名称（包含仓库和标签）',
    filter: '筛选',
    allImages: '所有镜像',
    agentImagesOnly: '仅Agent镜像',
    noImagesFound: '未找到镜像',
    pullImageToGetStarted: '拉取一个镜像以开始使用',
    failedToFetchImages: '获取镜像失败',
    confirmDeleteImage: '确定要删除此镜像吗？',
    imagePlaceholder: '例如：anthropic/claude-code:latest',
    size: '大小',

    // Settings
    settingsTitle: '设置',
    saved: '已保存',
    apiKeys: 'API密钥',
    apiKeysDesc: '配置Agent的默认API密钥。创建新会话时将使用这些密钥。',
    anthropicApiKey: 'Anthropic API密钥',
    anthropicApiKeyDesc: '用于Claude Code代理',
    openaiApiKey: 'OpenAI API密钥',
    openaiApiKeyDesc: '用于Codex代理',
    defaultWorkspace: '默认工作区',
    defaultWorkspaceDesc: '创建新会话时在容器中挂载的默认路径',
    systemStatus: '系统状态',
    backendStatus: '后端状态',
    healthy: '健康',
    unhealthy: '不健康',
    checking: '检查中...',
    serverStatus: '服务器状态',
    connected: '已连接',
    connectionFailed: '连接失败',
    defaultLanguage: '默认语言',
    chineseLanguage: '中文',
    about: '关于',
    agentBoxTagline: 'AI Agent容器化平台',
    agentBoxDescription: '在隔离容器中运行AI代理的开源解决方案。',
    githubRepository: 'GitHub仓库',

    // Task List
    tasksTitle: '任务',
    allTasks: '全部',
    queuedTasks: '队列中',
    completedTasks: '已完成',
    failedTasks: '失败',
    duration: '持续时间',
    cost: '花费',
    promptText: '提示词',
    confirmCancelTask: '取消任务',
    failedToLoadTasks: '加载任务失败',
    noTasksFound: '未找到任务',
    createTaskToGetStarted: '创建一个任务以开始使用',

    // Task Status
    pending: '等待中',
    queued: '队列中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',

    // Webhook List
    webhooksTitle: 'Webhook',
    webhooksDesc: '配置webhook以接收AgentBox实例中事件的实时通知。',
    newWebhook: '新建Webhook',
    noWebhooksFound: '未找到Webhook',
    createWebhookToGetStarted: '创建一个webhook以开始使用',
    failedToLoadWebhooks: '加载Webhook失败',
    signed: '已签名',
    events: '事件',
    lastTriggered: '最后触发',
    confirmDeleteWebhook: '删除webhook',
    createWebhook: '创建Webhook',
    editWebhook: '编辑Webhook',
    webhookURL: 'Webhook URL',
    secretDesc: '用于HMAC-SHA256签名',
    unchanged: '未更改',
    enterSecretPlaceholder: '输入签名密钥',
    selectAtLeastOneEvent: '请至少选择一个事件',
    totalWebhooks: 'Webhook总数',
    copyURL: '复制URL',

    // Event Types
    taskCreated: '任务已创建',
    taskCreatedDesc: '当创建新任务时',
    taskCompleted: '任务已完成',
    taskCompletedDesc: '当任务成功完成时',
    taskFailed: '任务失败',
    taskFailedDesc: '当任务失败时',
    sessionStarted: '会话已启动',
    sessionStartedDesc: '当会话容器启动时',
    sessionStopped: '会话已停止',
    sessionStoppedDesc: '当会话容器停止时',

    // System Maintenance
    systemMaintenance: '系统维护',
    healthChecksStatsCleanup: '健康检查、统计和清理',
    failedToFetchSystemData: '获取系统数据失败',
    systemHealth: '系统健康',
    overallStatus: '整体状态',
    uptime: '运行时长',
    containerRuntime: '容器运行时',
    containers: '容器',
    resources: '资源',
    serverInfo: '服务器信息',
    memory: '内存',
    cpuCores: 'CPU核心',
    goroutines: '协程',
    healthChecks: '健康检查',
    goVersion: 'Go版本',
    cleanupActions: '清理操作',
    orphanContainersTitle: '孤立容器',
    orphanContainersDesc: '删除由AgentBox管理但不再关联任何会话的容器。',
    cleanupContainers: '清理容器',
    cleaning: '清理中...',
    unusedImagesTitle: '未使用的镜像',
    unusedImagesDesc: '删除未被任何容器使用的镜像（不包括Agent镜像）。',
    cleanupImages: '清理镜像',
    removed: '已删除',
    orphanContainers: '个孤立容器',
    noOrphanContainersFound: '未找到孤立容器',
    failedToCleanupContainers: '清理容器失败',
    freed: '释放',
    noUnusedImagesToRemove: '没有可删除的未使用镜像',
    failedToCleanupImages: '清理镜像失败',
  },
} as const

export type TranslationKey = keyof typeof translations.en

export function t(key: TranslationKey, lang: Language = 'en'): string {
  return translations[lang][key] || key
}
