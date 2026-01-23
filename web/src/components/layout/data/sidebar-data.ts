import {
  Monitor,
  Bug,
  FileX,
  HelpCircle,
  Lock,
  Palette,
  ServerOff,
  UserX,
  Server,
  Zap,
  Image,
  Webhook,
  Construction,
  Command,
  History,
  Bot,
  Store,
  Settings,
  UserCog,
  Wrench,
  Bell,
  Box,
  Cloud,
  TerminalSquare,
  FileCode2,
  ListTodo,
  File,
  Activity,
} from 'lucide-react'
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
      logo: Command,
      plan: 'AI Agent Platform',
    },
  ],
  navGroups: [
    {
      title: 'Main',
      items: [
        {
          title: 'Command Center',
          url: '/command-center',
          icon: Activity,
        },
        {
          title: 'Tasks',
          url: '/tasks',
          icon: ListTodo,
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
        {
          title: 'API Docs',
          url: '/api-docs',
          icon: FileCode2,
        },
      ],
    },
    {
      title: 'Admin',
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
        {
          title: 'Skill Store',
          url: '/skill-store',
          icon: Store,
        },
        {
          title: 'Images',
          url: '/images',
          icon: Image,
        },
        {
          title: 'Webhooks',
          url: '/webhooks',
          icon: Webhook,
        },
        {
          title: 'System',
          url: '/system',
          icon: Monitor,
        },
      ],
    },
    {
      title: 'Pages',
      items: [
        {
          title: 'Errors',
          icon: Bug,
          items: [
            {
              title: 'Unauthorized',
              url: '/errors/unauthorized',
              icon: Lock,
            },
            {
              title: 'Forbidden',
              url: '/errors/forbidden',
              icon: UserX,
            },
            {
              title: 'Not Found',
              url: '/errors/not-found',
              icon: FileX,
            },
            {
              title: 'Internal Server Error',
              url: '/errors/internal-server-error',
              icon: ServerOff,
            },
            {
              title: 'Maintenance Error',
              url: '/errors/maintenance-error',
              icon: Construction,
            },
          ],
        },
      ],
    },
    {
      title: 'Other',
      items: [
        {
          title: 'Settings',
          icon: Settings,
          items: [
            {
              title: 'Profile',
              url: '/settings',
              icon: UserCog,
            },
            {
              title: 'Account',
              url: '/settings/account',
              icon: Wrench,
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
              icon: Palette,
            },
            {
              title: 'Notifications',
              url: '/settings/notifications',
              icon: Bell,
            },
            {
              title: 'Display',
              url: '/settings/display',
              icon: Monitor,
            },
          ],
        },
        {
          title: 'Documentation',
          url: '/help-center',
          icon: HelpCircle,
        },
      ],
    },
  ],
}
