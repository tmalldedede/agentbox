import { useState, useEffect } from 'react'
import { Cloud, Globe, Sparkles, ExternalLink, Check } from 'lucide-react'
import type { Provider, ProviderCategory } from '@/types'
import { api } from '@/services/api'

interface ProviderSelectorProps {
  agent: 'claude-code' | 'codex' | 'opencode'
  selectedProviderId?: string
  onSelect: (provider: Provider) => void
}

const categoryLabels: Record<ProviderCategory, string> = {
  official: '官方',
  cn_official: '国产官方',
  aggregator: '聚合器',
  third_party: '第三方',
}

const categoryColors: Record<ProviderCategory, string> = {
  official: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  cn_official: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  aggregator: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
  third_party: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
}

export function ProviderSelector({ agent, selectedProviderId, onSelect }: ProviderSelectorProps) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<ProviderCategory | 'all'>('all')

  useEffect(() => {
    loadProviders()
  }, [agent])

  const loadProviders = async () => {
    try {
      setLoading(true)
      const data = await api.listProviders({ agent })
      setProviders(data)
    } catch (err) {
      console.error('Failed to load providers:', err)
    } finally {
      setLoading(false)
    }
  }

  const filteredProviders = filter === 'all'
    ? providers
    : providers.filter(p => p.category === filter)

  // Group providers by category
  const groupedProviders: Record<ProviderCategory, Provider[]> = {
    official: [],
    cn_official: [],
    aggregator: [],
    third_party: [],
  }

  filteredProviders.forEach(p => {
    if (groupedProviders[p.category]) {
      groupedProviders[p.category].push(p)
    }
  })

  if (loading) {
    return (
      <div className="animate-pulse space-y-4">
        <div className="h-10 bg-gray-200 dark:bg-gray-700 rounded"></div>
        <div className="grid grid-cols-2 gap-4">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="h-24 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Filter tabs */}
      <div className="flex space-x-2 border-b border-gray-200 dark:border-gray-700 pb-2">
        <button
          onClick={() => setFilter('all')}
          className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
            filter === 'all'
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900 dark:text-primary-200'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
          }`}
        >
          全部
        </button>
        {Object.entries(categoryLabels).map(([key, label]) => (
          <button
            key={key}
            onClick={() => setFilter(key as ProviderCategory)}
            className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
              filter === key
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900 dark:text-primary-200'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Provider grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredProviders.map(provider => (
          <ProviderCard
            key={provider.id}
            provider={provider}
            isSelected={provider.id === selectedProviderId}
            onSelect={() => onSelect(provider)}
          />
        ))}
      </div>

      {filteredProviders.length === 0 && (
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          没有找到匹配的 Provider
        </div>
      )}
    </div>
  )
}

interface ProviderCardProps {
  provider: Provider
  isSelected: boolean
  onSelect: () => void
}

function ProviderCard({ provider, isSelected, onSelect }: ProviderCardProps) {
  const getCategoryIcon = (category: ProviderCategory) => {
    switch (category) {
      case 'official':
        return <Sparkles className="h-4 w-4" />
      case 'cn_official':
        return <Globe className="h-4 w-4" />
      case 'aggregator':
        return <Cloud className="h-4 w-4" />
      default:
        return <Cloud className="h-4 w-4" />
    }
  }

  return (
    <div
      onClick={onSelect}
      className={`relative p-4 rounded-lg border-2 cursor-pointer transition-all ${
        isSelected
          ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
          : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
      }`}
    >
      {/* Selected indicator */}
      {isSelected && (
        <div className="absolute top-2 right-2 bg-primary-500 text-white rounded-full p-1">
          <Check className="h-3 w-3" />
        </div>
      )}

      {/* Partner badge */}
      {provider.is_partner && (
        <div className="absolute top-2 left-2">
          <span className="px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded-full">
            推荐
          </span>
        </div>
      )}

      <div className="flex items-start space-x-3 mt-4">
        {/* Icon */}
        <div
          className="w-10 h-10 rounded-lg flex items-center justify-center text-white text-lg font-bold"
          style={{ backgroundColor: provider.icon_color || '#6366F1' }}
        >
          {provider.name.charAt(0).toUpperCase()}
        </div>

        <div className="flex-1 min-w-0">
          <h3 className="font-medium text-gray-900 dark:text-white truncate">
            {provider.name}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
            {provider.description || provider.base_url || '官方 API'}
          </p>
        </div>
      </div>

      {/* Category badge */}
      <div className="mt-3 flex items-center justify-between">
        <span className={`inline-flex items-center space-x-1 px-2 py-0.5 text-xs rounded-full ${categoryColors[provider.category]}`}>
          {getCategoryIcon(provider.category)}
          <span>{categoryLabels[provider.category]}</span>
        </span>

        {provider.website_url && (
          <a
            href={provider.website_url}
            target="_blank"
            rel="noopener noreferrer"
            onClick={e => e.stopPropagation()}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <ExternalLink className="h-4 w-4" />
          </a>
        )}
      </div>

      {/* Default model */}
      {provider.default_model && (
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
          默认模型: <code className="bg-gray-100 dark:bg-gray-800 px-1 rounded">{provider.default_model}</code>
        </div>
      )}
    </div>
  )
}

export default ProviderSelector
