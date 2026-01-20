import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  Rocket,
  Terminal,
  ArrowRight,
  Sparkles,
  Box,
  ListTodo,
  Layers,
  RefreshCw,
  Cpu,
  BookOpen,
  Zap,
  Globe,
  Sun,
  Moon,
} from 'lucide-react'
import type { Agent, Profile } from '@/types'
import { api } from '@/services/api'
import { useLanguage } from '@/contexts/LanguageContext'
import { useTheme } from '@/context/theme-provider'
import CreateSessionModal from '@/features/sessions/components/CreateSessionModal'

// Agent 颜色配置
const agentColors: Record<string, { bg: string; text: string; border: string; icon: string }> = {
  'claude-code': {
    bg: 'bg-purple-500/10',
    text: 'text-purple-400',
    border: 'border-purple-500/30 hover:border-purple-400',
    icon: 'bg-purple-500/20',
  },
  'codex': {
    bg: 'bg-emerald-500/10',
    text: 'text-emerald-400',
    border: 'border-emerald-500/30 hover:border-emerald-400',
    icon: 'bg-emerald-500/20',
  },
  'opencode': {
    bg: 'bg-blue-500/10',
    text: 'text-blue-400',
    border: 'border-blue-500/30 hover:border-blue-400',
    icon: 'bg-blue-500/20',
  },
}

const defaultColors = {
  bg: 'bg-gray-500/10',
  text: 'text-gray-400',
  border: 'border-gray-500/30 hover:border-gray-400',
  icon: 'bg-gray-500/20',
}

// Agent 卡片
function AgentCard({
  agent,
  onClick,
  language,
}: {
  agent: Agent
  onClick: () => void
  language: string
}) {
  const colors = agentColors[agent.name] || defaultColors

  return (
    <button
      onClick={onClick}
      className={`w-full p-6 rounded-2xl border-2 ${colors.border} ${colors.bg} text-left transition-all hover:scale-[1.02] group`}
    >
      <div className="flex items-start justify-between mb-4">
        <div className={`w-14 h-14 rounded-xl ${colors.icon} flex items-center justify-center`}>
          <Terminal className={`w-7 h-7 ${colors.text}`} />
        </div>
        <ArrowRight className={`w-5 h-5 text-muted-foreground opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all ${colors.text}`} />
      </div>
      <h3 className="text-lg font-bold text-foreground mb-1">{agent.display_name}</h3>
      <p className="text-sm text-muted-foreground line-clamp-2 mb-4">{agent.description}</p>
      <div className="flex items-center justify-between">
        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${colors.bg} ${colors.text}`}>
          <Sparkles className="w-3 h-3" />
          Ready
        </span>
        <span className={`text-sm font-medium ${colors.text} opacity-0 group-hover:opacity-100 transition-opacity`}>
          {language === 'zh' ? '点击开始 →' : 'Click to start →'}
        </span>
      </div>
    </button>
  )
}

// 概念卡片
function ConceptCard({
  icon,
  title,
  description,
  color,
}: {
  icon: React.ReactNode
  title: string
  description: string
  color: string
}) {
  return (
    <div className={`p-4 rounded-xl border border-border bg-card`}>
      <div className="flex items-start gap-3">
        <div className={`w-10 h-10 rounded-lg ${color} flex items-center justify-center flex-shrink-0`}>
          {icon}
        </div>
        <div>
          <h4 className="font-medium text-foreground mb-1">{title}</h4>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      </div>
    </div>
  )
}

export default function QuickStart() {
  const navigate = useNavigate()
  const { language, setLanguage } = useLanguage()
  const { theme, setTheme } = useTheme()
  const toggleTheme = () => setTheme(theme === 'dark' ? 'light' : 'dark')
  const [agents, setAgents] = useState<Agent[]>([])
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [agentsData, profilesData] = await Promise.all([
          api.listAgents(),
          api.listProfiles(),
        ])
        setAgents(agentsData || [])
        setProfiles(profilesData || [])
      } catch {
        // Ignore
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  const handleAgentClick = (agent: Agent) => {
    setSelectedAgent(agent)
    setShowCreate(true)
  }

  const concepts = language === 'zh' ? [
    {
      icon: <Terminal className="w-5 h-5 text-emerald-400" />,
      title: 'Session（会话）',
      description: '一个运行中的 Agent 容器，包含独立的工作目录和执行环境',
      color: 'bg-emerald-500/10',
    },
    {
      icon: <ListTodo className="w-5 h-5 text-blue-400" />,
      title: 'Task（任务）',
      description: '异步执行的后台任务，可以批量处理，自动通知结果',
      color: 'bg-blue-500/10',
    },
    {
      icon: <Layers className="w-5 h-5 text-purple-400" />,
      title: 'Profile（配置）',
      description: 'Agent 的预设配置模板，包含模型、MCP 服务器、权限等',
      color: 'bg-purple-500/10',
    },
  ] : [
    {
      icon: <Terminal className="w-5 h-5 text-emerald-400" />,
      title: 'Session',
      description: 'A running Agent container with isolated workspace and execution environment',
      color: 'bg-emerald-500/10',
    },
    {
      icon: <ListTodo className="w-5 h-5 text-blue-400" />,
      title: 'Task',
      description: 'Async background job that can be batched and auto-notifies on completion',
      color: 'bg-blue-500/10',
    },
    {
      icon: <Layers className="w-5 h-5 text-purple-400" />,
      title: 'Profile',
      description: 'Pre-configured template with model, MCP servers, permissions, etc.',
      color: 'bg-purple-500/10',
    },
  ]

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Header */}
      <header className="app-header flex-shrink-0">
        <div className="flex items-center gap-3">
          <Rocket className="w-5 h-5 text-emerald-400" />
          <span className="text-lg font-semibold">
            {language === 'zh' ? '快速开始' : 'Quick Start'}
          </span>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={toggleTheme} className="btn btn-ghost btn-icon">
            {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
          </button>
          <button
            onClick={() => setLanguage(language === 'en' ? 'zh' : 'en')}
            className="btn btn-ghost text-xs"
          >
            <Globe className="w-4 h-4" />
            {language === 'en' ? '中文' : 'EN'}
          </button>
        </div>
      </header>

      <div className="flex-1 overflow-auto p-6">
        {/* Hero Section */}
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-8">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 text-emerald-400 text-sm font-medium mb-4">
              <Zap className="w-4 h-4" />
              AgentBox v0.1.0
            </div>
            <h1 className="text-3xl font-bold text-foreground mb-3">
              {language === 'zh' ? '选择一个 Agent，开始你的第一个任务' : 'Choose an Agent to Start Your First Task'}
            </h1>
            <p className="text-muted-foreground text-lg">
              {language === 'zh'
                ? '点击下方卡片，创建一个新的工作会话'
                : 'Click a card below to create a new work session'}
            </p>
          </div>

          {/* Agent Cards */}
          <div className="mb-10">
            {loading ? (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {[1, 2, 3].map(i => (
                  <div key={i} className="h-52 skeleton rounded-2xl" />
                ))}
              </div>
            ) : agents.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-center">
                <Cpu className="w-16 h-16 text-muted-foreground mb-4" />
                <p className="text-foreground text-lg">
                  {language === 'zh' ? '没有可用的 Agent' : 'No agents available'}
                </p>
                <p className="text-muted-foreground text-sm mt-1">
                  {language === 'zh' ? '请检查 Docker 连接' : 'Check your Docker connection'}
                </p>
                <button
                  onClick={() => window.location.reload()}
                  className="btn btn-secondary mt-4"
                >
                  <RefreshCw className="w-4 h-4" />
                  {language === 'zh' ? '刷新' : 'Refresh'}
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {agents.map(agent => (
                  <AgentCard
                    key={agent.name}
                    agent={agent}
                    language={language}
                    onClick={() => handleAgentClick(agent)}
                  />
                ))}
              </div>
            )}
          </div>

          {/* Divider */}
          <div className="flex items-center gap-4 mb-8">
            <div className="flex-1 border-t border-border" />
            <div className="flex items-center gap-2 text-muted-foreground">
              <BookOpen className="w-4 h-4" />
              <span className="text-sm font-medium">
                {language === 'zh' ? '概念说明' : 'Key Concepts'}
              </span>
            </div>
            <div className="flex-1 border-t border-border" />
          </div>

          {/* Concepts */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
            {concepts.map((concept, i) => (
              <ConceptCard key={i} {...concept} />
            ))}
          </div>

          {/* Quick Links */}
          <div className="flex items-center justify-center gap-4">
            <button
              onClick={() => navigate({ to: '/' })}
              className="btn btn-secondary"
            >
              <Terminal className="w-4 h-4" />
              {language === 'zh' ? '查看会话列表' : 'View Sessions'}
            </button>
            <button
              onClick={() => navigate({ to: '/tasks' })}
              className="btn btn-secondary"
            >
              <ListTodo className="w-4 h-4" />
              {language === 'zh' ? '异步任务' : 'Async Tasks'}
            </button>
            <button
              onClick={() => navigate({ to: '/profiles' })}
              className="btn btn-secondary"
            >
              <Box className="w-4 h-4" />
              {language === 'zh' ? '配置模板' : 'Profiles'}
            </button>
          </div>
        </div>
      </div>

      {/* Create Modal */}
      {showCreate && (
        <CreateSessionModal
          agents={agents}
          profiles={profiles}
          defaultAgent={selectedAgent?.name}
          onClose={() => {
            setShowCreate(false)
            setSelectedAgent(null)
          }}
          onCreated={() => {
            setShowCreate(false)
            setSelectedAgent(null)
            navigate({ to: '/' })
          }}
        />
      )}
    </div>
  )
}
