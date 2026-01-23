import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
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
  Recycle,
  Play,
} from 'lucide-react'
import type { SystemHealth, SystemStats, GCStats, GCCandidate } from '@/types'
import { api } from '@/services/api'

export default function SystemMaintenance() {
  const navigate = useNavigate()
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [gcStats, setGCStats] = useState<GCStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cleaningContainers, setCleaningContainers] = useState(false)
  const [cleaningImages, setCleaningImages] = useState(false)
  const [triggeringGC, setTriggeringGC] = useState(false)
  const [previewingGC, setPreviewingGC] = useState(false)
  const [gcCandidates, setGCCandidates] = useState<GCCandidate[] | null>(null)
  const [editingGCConfig, setEditingGCConfig] = useState(false)
  const [gcConfigForm, setGCConfigForm] = useState({ interval: 60, ttl: 7200, idle: 600 })
  const [savingGCConfig, setSavingGCConfig] = useState(false)
  const [cleanupResult, setCleanupResult] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      const [healthData, statsData, gcData] = await Promise.all([
        api.getSystemHealth(),
        api.getSystemStats(),
        api.getGCStats(),
      ])
      setHealth(healthData)
      setStats(statsData)
      setGCStats(gcData)
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

  const handleTriggerGC = async () => {
    try {
      setTriggeringGC(true)
      setCleanupResult(null)
      const result = await api.triggerGC()
      if (result.removed > 0) {
        setCleanupResult(`GC completed: removed ${result.removed} container(s)`)
      } else {
        setCleanupResult('GC completed: no containers to remove')
      }
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to trigger GC')
    } finally {
      setTriggeringGC(false)
    }
  }

  const handlePreviewGC = async () => {
    try {
      setPreviewingGC(true)
      const candidates = await api.previewGC()
      setGCCandidates(candidates)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to preview GC')
    } finally {
      setPreviewingGC(false)
    }
  }

  const handleEditGCConfig = () => {
    if (gcStats) {
      setGCConfigForm({
        interval: gcStats.config.interval_seconds,
        ttl: gcStats.config.container_ttl_seconds,
        idle: gcStats.config.idle_timeout_seconds,
      })
    }
    setEditingGCConfig(true)
  }

  const handleSaveGCConfig = async () => {
    try {
      setSavingGCConfig(true)
      await api.updateGCConfig({
        interval_seconds: gcConfigForm.interval,
        container_ttl_seconds: gcConfigForm.ttl,
        idle_timeout_seconds: gcConfigForm.idle,
      })
      setEditingGCConfig(false)
      setCleanupResult('GC configuration updated')
      fetchData()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update GC config')
    } finally {
      setSavingGCConfig(false)
    }
  }

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
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
          <button onClick={() => navigate({ to: '/' })} className="btn btn-ghost btn-icon">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-3">
            <Activity className="w-6 h-6 text-emerald-400" />
            <div>
              <h1 className="text-xl font-semibold text-foreground">System Maintenance</h1>
              <p className="text-sm text-muted-foreground">
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
            <h3 className="text-lg font-medium text-foreground mb-4 flex items-center gap-2">
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
                    <p className="font-medium text-foreground capitalize">{health.status}</p>
                    <p className="text-sm text-muted-foreground">Overall Status</p>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
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
                    <p className="font-medium text-foreground capitalize">Docker {health.docker.status}</p>
                    <p className="text-sm text-muted-foreground">Container Runtime</p>
                  </div>
                </div>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Containers</span>
                    <span className="text-foreground/80">{health.docker.containers}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Images</span>
                    <span className="text-foreground/80">{health.docker.images}</span>
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
                    <p className="font-medium text-foreground">Resources</p>
                    <p className="text-sm text-muted-foreground">Server Info</p>
                  </div>
                </div>
                <div className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <MemoryStick className="w-3 h-3" /> Memory
                    </span>
                    <span className="text-foreground/80">{health.resources.memory_usage_mb} MB</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Cpu className="w-3 h-3" /> CPU Cores
                    </span>
                    <span className="text-foreground/80">{health.resources.num_cpu}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Zap className="w-3 h-3" /> Goroutines
                    </span>
                    <span className="text-foreground/80">{health.resources.num_goroutines}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Health Checks */}
            <div className="mt-6 pt-4 border-t border-default">
              <p className="text-sm font-medium text-muted-foreground mb-3">Health Checks</p>
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
                  <p className="text-2xl font-bold text-foreground">{stats.sessions.total}</p>
                  <p className="text-sm text-muted-foreground">Sessions</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Running</span>
                  <span className="text-emerald-400">{stats.sessions.running}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Stopped</span>
                  <span className="text-gray-400">{stats.sessions.stopped}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Error</span>
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
                  <p className="text-2xl font-bold text-foreground">{stats.containers.total}</p>
                  <p className="text-sm text-muted-foreground">Containers</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Running</span>
                  <span className="text-emerald-400">{stats.containers.running}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Stopped</span>
                  <span className="text-gray-400">{stats.containers.stopped}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Other</span>
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
                  <p className="text-2xl font-bold text-foreground">{stats.images.total}</p>
                  <p className="text-sm text-muted-foreground">Images</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Agent Images</span>
                  <span className="text-purple-400">{stats.images.agent_images}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">In Use</span>
                  <span className="text-emerald-400">{stats.images.in_use}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Total Size</span>
                  <span className="text-foreground/80">{formatSize(stats.images.total_size)}</span>
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
                  <p className="text-2xl font-bold text-foreground">{stats.system.num_cpu}</p>
                  <p className="text-sm text-muted-foreground">CPU Cores</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Memory</span>
                  <span className="text-foreground/80">{stats.system.memory_usage_mb} MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Go Version</span>
                  <span className="text-foreground/80">{stats.system.go_version.replace('go', '')}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Uptime</span>
                  <span className="text-foreground/80">{stats.system.uptime}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* GC Status */}
        {gcStats && (
          <div className="card p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-foreground flex items-center gap-2">
                <Recycle className="w-5 h-5 text-cyan-400" />
                Container GC
              </h3>
              <div className="flex items-center gap-3">
                <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${
                  gcStats.running
                    ? 'bg-emerald-500/20 text-emerald-400'
                    : 'bg-gray-500/20 text-gray-400'
                }`}>
                  <span className={`w-2 h-2 rounded-full ${gcStats.running ? 'bg-emerald-400 animate-pulse' : 'bg-gray-400'}`} />
                  {gcStats.running ? 'Running' : 'Stopped'}
                </span>
                <button
                  onClick={handlePreviewGC}
                  disabled={previewingGC}
                  className="btn btn-secondary btn-sm"
                >
                  {previewingGC ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <AlertCircle className="w-3.5 h-3.5" />
                  )}
                  Preview
                </button>
                <button
                  onClick={handleTriggerGC}
                  disabled={triggeringGC}
                  className="btn btn-secondary btn-sm"
                >
                  {triggeringGC ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Play className="w-3.5 h-3.5" />
                  )}
                  Trigger GC
                </button>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* GC Stats */}
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Total Runs</span>
                  <span className="text-foreground/80">{gcStats.total_runs}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Containers Removed</span>
                  <span className="text-foreground/80">{gcStats.containers_removed}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Last Run</span>
                  <span className="text-foreground/80">
                    {gcStats.last_run_at && gcStats.last_run_at !== '0001-01-01T00:00:00Z'
                      ? new Date(gcStats.last_run_at).toLocaleTimeString()
                      : 'Never'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Next Run</span>
                  <span className="text-foreground/80">
                    {gcStats.next_run_at && gcStats.next_run_at !== '0001-01-01T00:00:00Z'
                      ? new Date(gcStats.next_run_at).toLocaleTimeString()
                      : '-'}
                  </span>
                </div>
              </div>

              {/* GC Config */}
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between mb-1">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Configuration</p>
                  {!editingGCConfig && (
                    <button onClick={handleEditGCConfig} className="text-xs text-cyan-400 hover:text-cyan-300">
                      Edit
                    </button>
                  )}
                </div>
                {editingGCConfig ? (
                  <div className="space-y-2">
                    <div>
                      <label className="text-xs text-muted-foreground">Interval (sec)</label>
                      <input
                        type="number"
                        min={10}
                        value={gcConfigForm.interval}
                        onChange={e => setGCConfigForm(f => ({ ...f, interval: parseInt(e.target.value) || 10 }))}
                        className="w-full px-2 py-1 text-xs rounded bg-background border border-default text-foreground"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground">TTL (sec)</label>
                      <input
                        type="number"
                        min={60}
                        value={gcConfigForm.ttl}
                        onChange={e => setGCConfigForm(f => ({ ...f, ttl: parseInt(e.target.value) || 60 }))}
                        className="w-full px-2 py-1 text-xs rounded bg-background border border-default text-foreground"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground">Idle Timeout (sec)</label>
                      <input
                        type="number"
                        min={30}
                        value={gcConfigForm.idle}
                        onChange={e => setGCConfigForm(f => ({ ...f, idle: parseInt(e.target.value) || 30 }))}
                        className="w-full px-2 py-1 text-xs rounded bg-background border border-default text-foreground"
                      />
                    </div>
                    <div className="flex gap-2 pt-1">
                      <button onClick={handleSaveGCConfig} disabled={savingGCConfig} className="btn btn-primary btn-sm text-xs">
                        {savingGCConfig ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Save'}
                      </button>
                      <button onClick={() => setEditingGCConfig(false)} className="btn btn-secondary btn-sm text-xs">Cancel</button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Scan Interval</span>
                      <span className="text-foreground/80">{formatDuration(gcStats.config.interval_seconds)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Container TTL</span>
                      <span className="text-foreground/80">{formatDuration(gcStats.config.container_ttl_seconds)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Idle Timeout</span>
                      <span className="text-foreground/80">{formatDuration(gcStats.config.idle_timeout_seconds)}</span>
                    </div>
                  </>
                )}
              </div>

              {/* GC Errors */}
              <div className="space-y-2 text-sm">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">Recent Errors</p>
                {gcStats.errors.length === 0 ? (
                  <p className="text-emerald-400 text-xs">No errors</p>
                ) : (
                  <div className="space-y-1 max-h-24 overflow-y-auto">
                    {gcStats.errors.slice(-5).map((err, i) => (
                      <p key={i} className="text-xs text-red-400 truncate" title={err}>{err}</p>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* GC Preview Results */}
            {gcCandidates !== null && (
              <div className="mt-4 pt-4 border-t border-default">
                <div className="flex items-center justify-between mb-3">
                  <p className="text-sm font-medium text-foreground">
                    GC Preview — {gcCandidates.length} container(s) to remove
                  </p>
                  <button onClick={() => setGCCandidates(null)} className="text-xs text-muted-foreground hover:text-foreground">
                    Dismiss
                  </button>
                </div>
                {gcCandidates.length === 0 ? (
                  <p className="text-sm text-emerald-400">No containers eligible for removal</p>
                ) : (
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {gcCandidates.map((c) => (
                      <div key={c.container_id} className="flex items-center justify-between p-2 rounded bg-background/50 border border-default text-xs">
                        <div className="flex items-center gap-3">
                          <Box className="w-3.5 h-3.5 text-muted-foreground" />
                          <div>
                            <span className="text-foreground font-mono">{c.container_id.slice(0, 12)}</span>
                            {c.name && <span className="text-muted-foreground ml-2">{c.name}</span>}
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className={`px-1.5 py-0.5 rounded text-xs ${
                            c.status === 'running' ? 'bg-emerald-500/20 text-emerald-400' :
                            c.status === 'exited' ? 'bg-gray-500/20 text-gray-400' :
                            'bg-amber-500/20 text-amber-400'
                          }`}>{c.status}</span>
                          <span className="text-amber-400">{c.reason}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Cleanup Actions */}
        <div className="card p-6">
          <h3 className="text-lg font-medium text-foreground mb-4 flex items-center gap-2">
            <Trash2 className="w-5 h-5 text-red-400" />
            Cleanup Actions
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Cleanup Containers */}
            <div className="p-4 rounded-lg bg-card border border-default">
              <h4 className="font-medium text-foreground mb-2">Orphan Containers</h4>
              <p className="text-sm text-muted-foreground mb-4">
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
              <h4 className="font-medium text-foreground mb-2">Unused Images</h4>
              <p className="text-sm text-muted-foreground mb-4">
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
