import { Terminal, Sun, Moon, Globe } from 'lucide-react'
import { useLanguage } from '../../contexts/LanguageContext'
import { useTheme } from '../../contexts/ThemeContext'

interface DashboardHeaderProps {
  totalSessions: number
  runningSessions: number
  currentTime: Date
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: true,
  })
}

export function DashboardHeader({
  totalSessions,
  runningSessions,
  currentTime,
}: DashboardHeaderProps) {
  const { t, language, setLanguage } = useLanguage()
  const { theme, toggleTheme } = useTheme()

  return (
    <header className="app-header flex-shrink-0">
      <div className="flex items-center gap-3">
        <Terminal className="w-5 h-5 text-emerald-400" />
        <span className="text-lg font-semibold">{t('sessions')}</span>
      </div>

      <div className="flex items-center gap-4 text-sm">
        {/* Live Status */}
        <div className="flex items-center gap-2">
          <span className="status-dot status-dot-live" />
          <span className="text-emerald-400 font-semibold">{totalSessions}</span>
          <span className="text-muted">total</span>
        </div>

        <div className="flex items-center gap-2">
          <span className="status-dot status-dot-running" />
          <span className="text-emerald-400">{runningSessions}</span>
          <span className="text-muted">running</span>
        </div>

        {/* Theme Toggle */}
        <button
          onClick={toggleTheme}
          className="btn btn-ghost btn-icon"
          title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
        >
          {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
        </button>

        {/* Language Toggle */}
        <button
          onClick={() => setLanguage(language === 'en' ? 'zh' : 'en')}
          className="btn btn-ghost text-xs"
        >
          <Globe className="w-4 h-4" />
          {language === 'en' ? '中文' : 'EN'}
        </button>

        {/* Time */}
        <span className="text-muted font-mono">{formatTime(currentTime)}</span>
      </div>
    </header>
  )
}
