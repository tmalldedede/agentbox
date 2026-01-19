import { Settings } from 'lucide-react'
import type { Provider } from '../../types'
import { Section } from './Section'
import ProviderSelector from '../ProviderSelector'

interface BasicInfoSectionProps {
  name: string
  setName: (v: string) => void
  description: string
  setDescription: (v: string) => void
  adapter: string
  setAdapter: (v: string) => void
  modelName: string
  setModelName: (v: string) => void
  modelBaseUrl: string
  setModelBaseUrl: (v: string) => void
  modelProvider: string
  setModelProvider: (v: string) => void
  selectedProvider: Provider | null
  setSelectedProvider: (v: Provider | null) => void
  showProviderSelector: boolean
  setShowProviderSelector: (v: boolean) => void
  isNewProfile: boolean
  isBuiltIn: boolean
}

export function BasicInfoSection({
  name,
  setName,
  description,
  setDescription,
  adapter,
  setAdapter,
  modelName,
  setModelName,
  modelBaseUrl,
  setModelBaseUrl,
  modelProvider: _modelProvider,
  setModelProvider,
  selectedProvider,
  setSelectedProvider,
  showProviderSelector,
  setShowProviderSelector,
  isNewProfile,
  isBuiltIn,
}: BasicInfoSectionProps) {
  const disabled = !isNewProfile && isBuiltIn

  return (
    <Section title="Basic Information" icon={<Settings className="w-5 h-5" />}>
      <div className="space-y-4 mt-4">
        <div>
          <label className="block text-sm font-medium text-secondary mb-1">Name *</label>
          <input
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            className="input w-full"
            placeholder="Enter profile name"
            disabled={disabled}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-secondary mb-1">Description</label>
          <textarea
            value={description}
            onChange={e => setDescription(e.target.value)}
            className="input w-full"
            rows={3}
            placeholder="Describe what this profile is for"
            disabled={disabled}
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">Adapter</label>
            {isNewProfile ? (
              <select
                value={adapter}
                onChange={e => setAdapter(e.target.value)}
                className="select w-full"
              >
                <option value="claude-code">Claude Code</option>
                <option value="codex">Codex</option>
                <option value="opencode">OpenCode</option>
              </select>
            ) : (
              <input type="text" value={adapter} className="input w-full" disabled />
            )}
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary mb-1">Model</label>
            <input
              type="text"
              value={modelName}
              onChange={e => setModelName(e.target.value)}
              className="input w-full"
              placeholder="e.g., claude-sonnet-4-20250514"
              disabled={disabled}
            />
          </div>
        </div>
        <div>
          <label className="block text-sm font-medium text-secondary mb-1">
            API Base URL
            <span className="text-muted font-normal ml-2">(Optional)</span>
          </label>
          <input
            type="text"
            value={modelBaseUrl}
            onChange={e => setModelBaseUrl(e.target.value)}
            className="input w-full font-mono text-sm"
            placeholder="https://api.anthropic.com or custom proxy URL"
            disabled={disabled}
          />
          <p className="text-xs text-muted mt-1">
            Leave empty to use the default API endpoint. Use for proxies or compatible APIs.
          </p>
        </div>

        {/* Provider Quick Select */}
        <div className="pt-4 border-t border-primary">
          <div className="flex items-center justify-between mb-3">
            <label className="block text-sm font-medium text-secondary">Quick Select Provider</label>
            <button
              onClick={() => setShowProviderSelector(!showProviderSelector)}
              className="btn btn-ghost btn-sm"
              type="button"
            >
              {showProviderSelector ? 'Hide' : 'Show'} Providers
            </button>
          </div>
          {showProviderSelector && (
            <div className="mt-3">
              <ProviderSelector
                agent={adapter as 'claude-code' | 'codex' | 'opencode'}
                selectedProviderId={selectedProvider?.id}
                onSelect={provider => {
                  setSelectedProvider(provider)
                  setModelProvider(provider.id)
                  if (provider.base_url) {
                    setModelBaseUrl(provider.base_url)
                  }
                  if (provider.default_model) {
                    setModelName(provider.default_model)
                  }
                  setShowProviderSelector(false)
                }}
              />
            </div>
          )}
          {selectedProvider && (
            <div className="mt-3 p-3 bg-secondary rounded-lg flex items-center gap-3">
              <div
                className="w-8 h-8 rounded-lg flex items-center justify-center text-white text-sm font-bold"
                style={{ backgroundColor: selectedProvider.icon_color || '#6366F1' }}
              >
                {selectedProvider.name.charAt(0).toUpperCase()}
              </div>
              <div>
                <div className="font-medium text-primary">{selectedProvider.name}</div>
                <div className="text-xs text-muted">{selectedProvider.description}</div>
              </div>
              <button
                onClick={() => {
                  setSelectedProvider(null)
                  setModelProvider('')
                  setModelBaseUrl('')
                }}
                className="ml-auto text-muted hover:text-primary"
                type="button"
              >
                ✕
              </button>
            </div>
          )}
        </div>
      </div>
    </Section>
  )
}
