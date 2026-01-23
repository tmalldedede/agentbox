import { Cloud, Globe, Layers, Users } from 'lucide-react'

export const categories = [
  { label: 'Official', value: 'official', icon: Cloud },
  { label: 'CN Official', value: 'cn_official', icon: Globe },
  { label: 'Aggregator', value: 'aggregator', icon: Layers },
  { label: 'Third Party', value: 'third_party', icon: Users },
] as const

export const agents = [
  { label: 'Claude Code', value: 'claude-code' },
  { label: 'Codex', value: 'codex' },
  { label: 'OpenCode', value: 'opencode' },
] as const

export const categoryColorMap: Record<string, string> = {
  official: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  cn_official: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  aggregator: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  third_party: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
}

export const agentColorMap: Record<string, string> = {
  'claude-code': 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  codex: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  opencode: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
}
