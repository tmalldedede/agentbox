import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Activity,
  RefreshCw,
  Loader2,
  AlertCircle,
  CheckCircle,
  XCircle,
  HardDrive,
  Cpu,
  Box,
  Terminal,
  Trash2,
  Clock,
  Server,
  MemoryStick,
  Gauge,
  Zap,
} from 'lucide-react'
import type { SystemHealth, SystemStats } from '../types'
import { api } from '../services/api'

export default function SystemMaintenance() {
  const navigate = useNavigate()
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cleaningContainers, setCleaningContainers] = useState(false)
  const [cleaningImages, setCleaningImages] = useState(false)
  const [cleanupResult, setCleanupResult] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      const [healthData, statsData] = await Promise.all([
        api.getSystemHealth(),
        api.getSystemStats(),
      ])
      setHealth(healthData)
      setStats(statsData)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch system data')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 10000)
    return () => clearInterval(interval)
  }, [])

  const handleCleanupContainers = async () => {
    try {
      setCleaningContainers(true)
      setCleanupResult(null)
      const result = await api.cleanupContainers()
      if (result.removed.length > 0) {
        setCleanupResult(`Removed ${result.removed.length} orphan container(s)`)
      } else {
        setCleanupResult('No orphan containers found')
      }
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cleanup containers')
    } finally {
      setCleaningContainers(false)
    }
  }

  const handleCleanupImages = async () => {
    try {
      setCleaningImages(true)
      setCleanupResult(null)
      const result = await api.cleanupImages({ unused_only: true })
      if (result.removed.length > 0) {
        const spaceMB = (result.space_freed / 1024 / 1024).toFixed(1)
        setCleanupResult(`Removed ${result.removed.length} image(s), freed ${spaceMB} MB`)
      } else {
        setCleanupResult('No unused images to remove')
      }
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cleanup images')
    } finally {
      setCleaningImages(false)
    }
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
  }

  if (loading && !health) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
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
            <Activity className="w-6 h-6 text-emerald-400" />
            <div>
              <h1 className="text-xl font-semibold text-primary">System Maintenance</h1>
              <p className="text-sm text-muted">
                Health checks, stats & cleanup
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={fetchData} className="btn btn-secondary" disabled={loading}>
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </header>

      <div className="p-6 space-y-6">
        {error && (
          <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
            <button onClick={() => setError(null)} className="ml-auto text-red-400 hover:text-red-300">
              &times;
            </button>
          </div>
        )}

        {cleanupResult && (
          <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center gap-3">
            <CheckCircle className="w-5 h-5 flex-shrink-0" />
            <span>{cleanupResult}</span>
            <button onClick={() => setCleanupResult(null)} className="ml-auto text-emerald-400 hover:text-emerald-300">
              &times;
            </button>
          </div>
        )}

        {/* Health Status */}
        {health && (
          <div className="card p-6">
            <h3 className="text-lg font-medium text-primary mb-4 flex items-center gap-2">
              <Gauge className="w-5 h-5 text-emerald-400" />
              System Health
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Overall Status */}
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  {health.status === 'healthy' ? (
                    <div className="w-10 h-10 rounded-full bg-emerald-500/20 flex items-center justify-center">
                      <CheckCircle className="w-5 h-5 text-emerald-400" />
                    </div>
                  ) : health.status === 'degraded' ? (
                    <div className="w-10 h-10 rounded-full bg-amber-500/20 flex items-center justify-center">
                      <AlertCircle className="w-5 h-5 text-amber-400" />
                    </div>
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-red-500/20 flex items-center justify-center">
                      <XCircle className="w-5 h-5 text-red-400" />
                    </div>
                  )}
                  <div>
                    <p className="font-medium text-primary capitalize">{health.status}</p>
                    <p className="text-sm text-muted">Overall Status</p>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-sm text-muted">
                  <Clock className="w-4 h-4" />
                  Uptime: {health.uptime}
                </div>
              </div>

              {/* Docker Status */}
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  {health.docker.status === 'healthy' ? (
                    <div className="w-10 h-10 rounded-full bg-blue-500/20 flex items-center justify-center">
                      <Box className="w-5 h-5 text-blue-400" />
                    </div>
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-red-500/20 flex items-center justify-center">
                      <Box className="w-5 h-5 text-red-400" />
                    </div>
                  )}
                  <div>
                    <p className="font-medium text-primary capitalize">Docker {health.docker.status}</p>
                    <p className="text-sm text-muted">Container Runtime</p>
                  </div>
                </div>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted">Containers</span>
                    <span className="text-secondary">{health.docker.containers}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted">Images</span>
                    <span className="text-secondary">{health.docker.images}</span>
                  </div>
                </div>
              </div>

              {/* Resources */}
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-purple-500/20 flex items-center justify-center">
                    <Server className="w-5 h-5 text-purple-400" />
                  </div>
                  <div>
                    <p className="font-medium text-primary">Resources</p>
                    <p className="text-sm text-muted">Server Info</p>
                  </div>
                </div>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted flex items-center gap-1">
                      <MemoryStick className="w-3 h-3" /> Memory
                    </span>
                    <span className="text-secondary">{health.resources.memory_usage_mb} MB</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted flex items-center gap-1">
                      <Cpu className="w-3 h-3" /> CPU Cores
                    </span>
                    <span className="text-secondary">{health.resources.num_cpu}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted flex items-center gap-1">
                      <Zap className="w-3 h-3" /> Goroutines
                    </span>
                    <span className="text-secondary">{health.resources.num_goroutines}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Health Checks */}
            <div className="mt-6 pt-4 border-t border-default">
              <p className="text-sm font-medium text-muted mb-3">Health Checks</p>
              <div className="flex flex-wrap gap-2">
                {Object.entries(health.checks).map(([name, status]) => (
                  <span
                    key={name}
                    className={`px-3 py-1 rounded-full text-xs flex items-center gap-1.5 ${
                      status === 'ok'
                        ? 'bg-emerald-500/20 text-emerald-400'
                        : 'bg-red-500/20 text-red-400'
                    }`}
                  >
                    {status === 'ok' ? (
                      <CheckCircle className="w-3 h-3" />
                    ) : (
                      <XCircle className="w-3 h-3" />
                    )}
                    {name}
                  </span>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Statistics */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Sessions */}
            <div className="card p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-emerald-500/20 flex items-center justify-center">
                  <Terminal className="w-5 h-5 text-emerald-400" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-primary">{stats.sessions.total}</p>
                  <p className="text-sm text-muted">Sessions</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted">Running</span>
                  <span className="text-emerald-400">{stats.sessions.running}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Stopped</span>
                  <span className="text-gray-400">{stats.sessions.stopped}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Error</span>
                  <span className="text-red-400">{stats.sessions.error}</span>
                </div>
              </div>
            </div>

            {/* Containers */}
            <div className="card p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                  <Box className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-primary">{stats.containers.total}</p>
                  <p className="text-sm text-muted">Containers</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted">Running</span>
                  <span className="text-emerald-400">{stats.containers.running}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Stopped</span>
                  <span className="text-gray-400">{stats.containers.stopped}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Other</span>
                  <span className="text-amber-400">{stats.containers.other}</span>
                </div>
              </div>
            </div>

            {/* Images */}
            <div className="card p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-purple-500/20 flex items-center justify-center">
                  <HardDrive className="w-5 h-5 text-purple-400" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-primary">{stats.images.total}</p>
                  <p className="text-sm text-muted">Images</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted">Agent Images</span>
                  <span className="text-purple-400">{stats.images.agent_images}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">In Use</span>
                  <span className="text-emerald-400">{stats.images.in_use}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Total Size</span>
                  <span className="text-secondary">{formatSize(stats.images.total_size)}</span>
                </div>
              </div>
            </div>

            {/* System */}
            <div className="card p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-amber-500/20 flex items-center justify-center">
                  <Server className="w-5 h-5 text-amber-400" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-primary">{stats.system.num_cpu}</p>
                  <p className="text-sm text-muted">CPU Cores</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted">Memory</span>
                  <span className="text-secondary">{stats.system.memory_usage_mb} MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Go Version</span>
                  <span className="text-secondary">{stats.system.go_version.replace('go', '')}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted">Uptime</span>
                  <span className="text-secondary">{stats.system.uptime}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Cleanup Actions */}
        <div className="card p-6">
          <h3 className="text-lg font-medium text-primary mb-4 flex items-center gap-2">
            <Trash2 className="w-5 h-5 text-red-400" />
            Cleanup Actions
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Cleanup Containers */}
            <div className="p-4 rounded-lg bg-card border border-default">
              <h4 className="font-medium text-primary mb-2">Orphan Containers</h4>
              <p className="text-sm text-muted mb-4">
                Remove containers that are managed by AgentBox but no longer have an associated session.
              </p>
              <button
                onClick={handleCleanupContainers}
                disabled={cleaningContainers}
                className="btn btn-secondary"
              >
                {cleaningContainers ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Cleaning...
                  </>
                ) : (
                  <>
                    <Trash2 className="w-4 h-4" />
                    Cleanup Containers
                  </>
                )}
              </button>
            </div>

            {/* Cleanup Images */}
            <div className="p-4 rounded-lg bg-card border border-default">
              <h4 className="font-medium text-primary mb-2">Unused Images</h4>
              <p className="text-sm text-muted mb-4">
                Remove images that are not being used by any container (excludes Agent images).
              </p>
              <button
                onClick={handleCleanupImages}
                disabled={cleaningImages}
                className="btn btn-secondary"
              >
                {cleaningImages ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Cleaning...
                  </>
                ) : (
                  <>
                    <Trash2 className="w-4 h-4" />
                    Cleanup Images
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
