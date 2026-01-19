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
    github: 'GitHub',
    version: 'Version',

    // Language
    language: 'Language',
    english: 'English',
    chinese: '中文',

    // Theme
    theme: 'Theme',
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
    github: 'GitHub',
    version: '版本',

    // Language
    language: '语言',
    english: 'English',
    chinese: '中文',

    // Theme
    theme: '主题',
  },
} as const

export type TranslationKey = keyof typeof translations.en

export function t(key: TranslationKey, lang: Language = 'en'): string {
  return translations[lang][key] || key
}
