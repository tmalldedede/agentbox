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
} from 'lucide-react'
import { AgentBoxLogo } from '@/components/icons/agentbox-logo'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Admin',
    email: 'admin@agentbox.dev',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'AgentBox',
      logo: AgentBoxLogo,
      plan: 'AI Agent Platform',
    },
  ],
  navGroups: [
    {
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
      ],
    },
    {
      title: 'Core',
      items: [
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
      ],
    },
    {
      title: 'Platform',
      items: [
        {
          title: 'Sessions',
          url: '/sessions',
          icon: TerminalSquare,
        },
        {
          title: 'History',
          url: '/history',
          icon: History,
        },
        {
          title: 'Providers',
          url: '/providers',
          icon: Cloud,
        },
        {
          title: 'Runtimes',
          url: '/runtimes',
          icon: Box,
        },
      ],
    },
    {
      title: 'Config',
      items: [
        {
          title: 'MCP Servers',
          url: '/mcp-servers',
          icon: Server,
        },
        {
          title: 'Skills',
          url: '/skills',
          icon: Zap,
        },
      ],
    },
    {
      title: 'System',
      items: [
        {
          title: 'Settings',
          url: '/settings',
          icon: Settings,
        },
        {
          title: 'Maintenance',
          url: '/system',
          icon: Wrench,
        },
        {
          title: 'API Docs',
          url: '/api-docs',
          icon: FileCode2,
        },
        {
          title: 'Help',
          url: '/help-center',
          icon: HelpCircle,
        },
      ],
    },
  ],
}
