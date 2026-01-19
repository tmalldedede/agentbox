import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Settings as SettingsIcon,
  Key,
  Globe,
  Save,
  CheckCircle,
  Server,
  RefreshCw,
  Sun,
  Moon,
  Monitor,
} from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { useTheme } from '../contexts/ThemeContext'
import { api } from '../services/api'

const STORAGE_KEY = 'agentbox-settings'

interface AppSettings {
  anthropicApiKey: string
  openaiApiKey: string
  defaultWorkspace: string
}

function loadSettings(): AppSettings {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch {
    // ignore
  }
  return {
    anthropicApiKey: '',
    openaiApiKey: '',
    defaultWorkspace: '/tmp/myproject',
  }
}

function saveSettings(settings: AppSettings) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
}

export default function Settings() {
  const navigate = useNavigate()
  const { t, language, setLanguage } = useLanguage()
  const { theme, setTheme } = useTheme()

  const [settings, setSettings] = useState<AppSettings>(loadSettings)
  const [saved, setSaved] = useState(false)
  const [healthStatus, setHealthStatus] = useState<'ok' | 'error' | 'loading'>('loading')
  const [version, setVersion] = useState('')

  useEffect(() => {
    checkHealth()
  }, [])

  const checkHealth = async () => {
    setHealthStatus('loading')
    try {
      const data = await api.health()
      setHealthStatus('ok')
      setVersion(data.version)
    } catch {
      setHealthStatus('error')
    }
  }

  const handleSave = () => {
    saveSettings(settings)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const updateSetting = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setSettings(prev => ({ ...prev, [key]: value }))
    setSaved(false)
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="app-header">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <SettingsIcon className="w-5 h-5 text-emerald-400" />
            <span className="font-semibold">{t('settings')}</span>
          </div>
        </div>

        <button
          onClick={handleSave}
          className={`btn ${saved ? 'btn-secondary' : 'btn-primary'}`}
        >
          {saved ? (
            <>
              <CheckCircle className="w-4 h-4" />
              Saved
            </>
          ) : (
            <>
              <Save className="w-4 h-4" />
              Save
            </>
          )}
        </button>
      </header>

      <div className="max-w-2xl mx-auto p-6 space-y-6">
        {/* Server Status */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <Server className="w-5 h-5 text-blue-400" />
              <h3 className="font-semibold">Server Status</h3>
            </div>
            <button onClick={checkHealth} className="btn btn-ghost btn-icon">
              <RefreshCw className={`w-4 h-4 ${healthStatus === 'loading' ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <div className="flex items-center gap-4">
            <div className={`w-3 h-3 rounded-full ${
              healthStatus === 'ok' ? 'bg-emerald-400' :
              healthStatus === 'error' ? 'bg-red-400' :
              'bg-amber-400 animate-pulse'
            }`} />
            <span className="text-secondary">
              {healthStatus === 'ok' && `Connected (v${version})`}
              {healthStatus === 'error' && 'Connection failed'}
              {healthStatus === 'loading' && 'Checking...'}
            </span>
          </div>
        </div>

        {/* Language */}
        <div className="card p-6">
          <div className="flex items-center gap-3 mb-4">
            <Globe className="w-5 h-5 text-purple-400" />
            <h3 className="font-semibold">{t('language')}</h3>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <button
              onClick={() => setLanguage('en')}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                language === 'en'
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : 'border-default'
              }`}
              style={{ borderColor: language !== 'en' ? 'var(--border-color)' : undefined }}
            >
              <p className="font-medium">English</p>
              <p className="text-sm text-muted">Default language</p>
            </button>
            <button
              onClick={() => setLanguage('zh')}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                language === 'zh'
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : 'border-default'
              }`}
              style={{ borderColor: language !== 'zh' ? 'var(--border-color)' : undefined }}
            >
              <p className="font-medium">中文</p>
              <p className="text-sm text-muted">Chinese</p>
            </button>
          </div>
        </div>

        {/* Theme */}
        <div className="card p-6">
          <div className="flex items-center gap-3 mb-4">
            <Monitor className="w-5 h-5 text-cyan-400" />
            <h3 className="font-semibold">{t('theme')}</h3>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <button
              onClick={() => setTheme('dark')}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                theme === 'dark'
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : 'border-default hover:border-[var(--border-hover)]'
              }`}
              style={{ borderColor: theme !== 'dark' ? 'var(--border-color)' : undefined }}
            >
              <div className="flex items-center gap-2 mb-1">
                <Moon className="w-4 h-4" />
                <p className="font-medium">Dark</p>
              </div>
              <p className="text-sm text-muted">深色主题</p>
            </button>
            <button
              onClick={() => setTheme('light')}
              className={`p-4 rounded-xl border-2 text-left transition-all ${
                theme === 'light'
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : 'border-default hover:border-[var(--border-hover)]'
              }`}
              style={{ borderColor: theme !== 'light' ? 'var(--border-color)' : undefined }}
            >
              <div className="flex items-center gap-2 mb-1">
                <Sun className="w-4 h-4" />
                <p className="font-medium">Light</p>
              </div>
              <p className="text-sm text-muted">浅色主题</p>
            </button>
          </div>
        </div>

        {/* API Keys */}
        <div className="card p-6">
          <div className="flex items-center gap-3 mb-4">
            <Key className="w-5 h-5 text-amber-400" />
            <h3 className="font-semibold">API Keys</h3>
          </div>
          <p className="text-sm text-muted mb-4">
            Configure default API keys for agents. These will be used when creating new sessions.
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-secondary mb-2">
                ANTHROPIC_API_KEY
              </label>
              <input
                type="password"
                value={settings.anthropicApiKey}
                onChange={(e) => updateSetting('anthropicApiKey', e.target.value)}
                placeholder="sk-ant-..."
                className="input"
              />
              <p className="text-xs text-muted mt-1">Used for Claude Code agent</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-secondary mb-2">
                OPENAI_API_KEY
              </label>
              <input
                type="password"
                value={settings.openaiApiKey}
                onChange={(e) => updateSetting('openaiApiKey', e.target.value)}
                placeholder="sk-..."
                className="input"
              />
              <p className="text-xs text-muted mt-1">Used for Codex agent</p>
            </div>
          </div>
        </div>

        {/* Default Workspace */}
        <div className="card p-6">
          <div className="flex items-center gap-3 mb-4">
            <Server className="w-5 h-5 text-cyan-400" />
            <h3 className="font-semibold">Default Workspace</h3>
          </div>
          <input
            type="text"
            value={settings.defaultWorkspace}
            onChange={(e) => updateSetting('defaultWorkspace', e.target.value)}
            placeholder="/tmp/myproject"
            className="input"
          />
          <p className="text-xs text-muted mt-1">
            Default path to mount in container when creating new sessions
          </p>
        </div>

        {/* About */}
        <div className="card p-6">
          <h3 className="font-semibold mb-4">About</h3>
          <div className="space-y-2 text-sm text-secondary">
            <p><strong className="text-primary">AgentBox</strong> - AI Agent Container Platform</p>
            <p>Open-source solution for running AI agents in isolated containers.</p>
            <p className="pt-2">
              <a
                href="https://github.com/tmalldedede/agentbox"
                target="_blank"
                rel="noopener noreferrer"
                className="text-emerald-400 hover:underline"
              >
                GitHub Repository
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

// Export settings getter for other components
export function getAppSettings(): AppSettings {
  return loadSettings()
}
