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
  KeyRound,
  Image,
  Webhook,
  Layers,
  Construction,
  Command,
  History,
  Bot,
  Store,
  FlaskConical,
  Settings,
  UserCog,
  Wrench,
  Bell,
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
          title: 'Agents',
          url: '/agents',
          icon: Bot,
        },
        {
          title: 'History',
          url: '/history',
          icon: History,
        },
        {
          title: 'API Playground',
          url: '/api-playground',
          icon: FlaskConical,
        },
        {
          title: 'Profiles',
          url: '/profiles',
          icon: Layers,
        },
      ],
    },
    {
      title: 'Resources',
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
        {
          title: 'Credentials',
          url: '/credentials',
          icon: KeyRound,
        },
        {
          title: 'Skill Store',
          url: '/skill-store',
          icon: Store,
        },
      ],
    },
    {
      title: 'Settings',
      items: [
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
        {
          title: 'Appearance',
          url: '/settings/appearance',
          icon: Palette,
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
