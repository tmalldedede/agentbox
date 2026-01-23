// Provider icon mapping: maps icon ID to display properties
export interface ProviderIconInfo {
  label: string   // 1-2 char short label
  color: string   // Background color (tailwind classes)
}

const iconMap: Record<string, ProviderIconInfo> = {
  anthropic: { label: 'A', color: 'bg-amber-600 text-white' },
  openai: { label: 'AI', color: 'bg-emerald-600 text-white' },
  deepseek: { label: 'DS', color: 'bg-blue-600 text-white' },
  zhipu: { label: 'Z', color: 'bg-blue-700 text-white' },
  qwen: { label: 'Q', color: 'bg-indigo-600 text-white' },
  kimi: { label: 'K', color: 'bg-gray-800 text-white dark:bg-gray-200 dark:text-gray-800' },
  minimax: { label: 'M', color: 'bg-orange-500 text-white' },
  doubao: { label: 'D', color: 'bg-blue-500 text-white' },
  openrouter: { label: 'OR', color: 'bg-indigo-500 text-white' },
  aihubmix: { label: 'AH', color: 'bg-emerald-500 text-white' },
  azure: { label: 'Az', color: 'bg-sky-600 text-white' },
  ollama: { label: 'OL', color: 'bg-gray-700 text-white' },
}

const defaultIcon: ProviderIconInfo = { label: '?', color: 'bg-gray-400 text-white' }

export function getProviderIcon(iconId: string | undefined): ProviderIconInfo {
  if (!iconId) return defaultIcon
  return iconMap[iconId] || defaultIcon
}
