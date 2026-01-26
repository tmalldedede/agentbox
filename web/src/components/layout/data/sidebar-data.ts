import {
  HelpCircle,
  Server,
  Zap,
  History,
  Bot,
  Settings,
  Box,
  Cloud,
  TerminalSquare,
  FileCode2,
  ListTodo,
  File,
  Layers,
  LayoutDashboard,
  Wrench,
  Users,
  Clock,
  MessageSquare,
} from 'lucide-react'
import { AgentBoxLogo } from '@/components/icons/agentbox-logo'
import { type SidebarData } from '../types'

export function getSidebarData(role: string): SidebarData {
  const isAdmin = role === 'admin'

  const coreItems = [
    {
      title: 'Tasks',
      url: '/tasks',
      icon: ListTodo,
    },
    {
      title: 'Batches',
      url: '/batches',
      icon: Layers,
    },
    {
      title: 'Agents',
      url: '/agents',
      icon: Bot,
    },
    {
      title: 'Files',
      url: '/files',
      icon: File,
    },
  ]

  const navGroups = [
    {
      items: isAdmin
        ? [{ title: 'Dashboard', url: '/', icon: LayoutDashboard }]
        : [{ title: 'Tasks', url: '/', icon: ListTodo }],
    },
    {
      title: 'Core',
      items: coreItems,
    },
  ]

  if (isAdmin) {
    navGroups.push(
      {
        title: 'Platform',
        items: [
          { title: 'Sessions', url: '/sessions', icon: TerminalSquare },
          { title: 'History', url: '/history', icon: History },
          { title: 'Providers', url: '/providers', icon: Cloud },
          { title: 'Runtimes', url: '/runtimes', icon: Box },
        ],
      },
      {
        title: 'Config',
        items: [
          { title: 'MCP Servers', url: '/mcp-servers', icon: Server },
          { title: 'Skills', url: '/skills', icon: Zap },
          { title: 'Cron Jobs', url: '/crons', icon: Clock },
          { title: 'Channels', url: '/channels', icon: MessageSquare },
        ],
      },
      {
        title: 'System',
        items: [
          { title: 'Users', url: '/users', icon: Users },
          { title: 'Settings', url: '/settings', icon: Settings },
          { title: 'Maintenance', url: '/system', icon: Wrench },
          { title: 'API Docs', url: '/api-docs', icon: FileCode2 },
          { title: 'Help', url: '/help-center', icon: HelpCircle },
        ],
      },
    )
  } else {
    navGroups.push({
      title: 'Other',
      items: [
        { title: 'History', url: '/history', icon: History },
        { title: 'API Docs', url: '/api-docs', icon: FileCode2 },
        { title: 'Help', url: '/help-center', icon: HelpCircle },
      ],
    })
  }

  return {
    user: {
      name: '',
      email: '',
      avatar: '/avatars/shadcn.jpg',
    },
    teams: [
      {
        name: 'AgentBox',
        logo: AgentBoxLogo,
        plan: 'AI Agent Platform',
      },
    ],
    navGroups,
  }
}

// Keep backward compatibility - default to admin view
export const sidebarData: SidebarData = getSidebarData('admin')
