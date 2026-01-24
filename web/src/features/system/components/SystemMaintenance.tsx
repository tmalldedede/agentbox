import { useState, useEffect, useRef, useCallback } from 'react'
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
  Users,
  Layers,
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

  const withTimeout = <T,>(promise: Promise<T>, ms = 5000): Promise<T> => {
    return Promise.race([
      promise,
      new Promise<T>((_, reject) => setTimeout(() => reject(new Error('Request timeout')), ms)),
    ])
  }

  const failCountRef = useRef(0)
  const isMountedRef = useRef(true)

  const fetchData = useCallback(async (isBackground = false) => {
    if (!isBackground) setLoading(true)
    try {
      const [healthResult, statsResult, gcResult] = await Promise.allSettled([
        withTimeout(api.getSystemHealth()),
        withTimeout(api.getSystemStats()),
        withTimeout(api.getGCStats()),
      ])
      if (!isMountedRef.current) return

      if (healthResult.status === 'fulfilled') setHealth(healthResult.value)
      if (statsResult.status === 'fulfilled') setStats(statsResult.value)
      if (gcResult.status === 'fulfilled') setGCStats(gcResult.value)

      const allFailed = healthResult.status === 'rejected' &&
        statsResult.status === 'rejected' &&
        gcResult.status === 'rejected'
      if (allFailed) {
        failCountRef.current++
        const reason = (healthResult as PromiseRejectedResult).reason
        setError(reason?.message || 'Failed to fetch system data')
      } else {
        failCountRef.current = 0
        setError(null)
      }
    } catch (err) {
      if (!isMountedRef.current) return
      failCountRef.current++
      setError(err instanceof Error ? err.message : 'Failed to fetch system data')
    } finally {
      if (isMountedRef.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    isMountedRef.current = true
    fetchData(false)
    const interval = setInterval(() => {
      if (failCountRef.current >= 3) return
      fetchData(true)
    }, 5000) // 更频繁刷新以追踪 batch 进度
    return () => {
      isMountedRef.current = false
      clearInterval(interval)
    }
  }, [fetchData])

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
      fetchData(true)
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
      fetchData(true)
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
      fetchData(true)
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
      fetchData(true)
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

  // 计算总体状态
  const getOverallStatus = () => {
    if (!health) return 'unknown'
    if (health.status === 'healthy' && health.docker.status === 'healthy') return 'healthy'
    if (health.status === 'degraded' || health.docker.status !== 'healthy') return 'degraded'
    return 'unhealthy'
  }

  const overallStatus = getOverallStatus()

  if (loading && !health) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-emerald-400 animate-spin" />
      </div>
    )
  }

  const batches = stats?.batches

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
              <h1 className="text-xl font-semibold text-foreground">Operations Dashboard</h1>
              <p className="text-sm text-muted-foreground">
                System health, worker pool & maintenance
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={() => { failCountRef.current = 0; fetchData(false) }} className="btn btn-secondary" disabled={loading}>
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </header>

      <div className="p-6 space-y-6">
        {/* Alerts */}
        {error && (
          <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <div>
              <span>{error}</span>
              {failCountRef.current >= 3 && (
                <p className="text-xs mt-1 opacity-70">Auto-refresh paused. Click Refresh to retry.</p>
              )}
            </div>
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

        {/* Top Status Indicators - 4 cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Overall Status */}
          <div className="card p-4">
            <div className="flex items-center gap-3">
              {overallStatus === 'healthy' ? (
                <div className="w-12 h-12 rounded-full bg-emerald-500/20 flex items-center justify-center">
                  <CheckCircle className="w-6 h-6 text-emerald-400" />
                </div>
              ) : overallStatus === 'degraded' ? (
                <div className="w-12 h-12 rounded-full bg-amber-500/20 flex items-center justify-center">
                  <AlertCircle className="w-6 h-6 text-amber-400" />
                </div>
              ) : (
                <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center">
                  <XCircle className="w-6 h-6 text-red-400" />
                </div>
              )}
              <div>
                <p className="text-lg font-semibold text-foreground capitalize">{overallStatus}</p>
                <p className="text-sm text-muted-foreground">Overall Status</p>
              </div>
            </div>
            {health && (
              <div className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
                <Clock className="w-3 h-3" />
                Uptime: {health.uptime}
              </div>
            )}
          </div>

          {/* Docker Status */}
          <div className="card p-4">
            <div className="flex items-center gap-3">
              {health?.docker.status === 'healthy' ? (
                <div className="w-12 h-12 rounded-full bg-blue-500/20 flex items-center justify-center">
                  <Box className="w-6 h-6 text-blue-400" />
                </div>
              ) : (
                <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center">
                  <Box className="w-6 h-6 text-red-400" />
                </div>
              )}
              <div>
                <p className="text-lg font-semibold text-foreground capitalize">
                  {health?.docker.status || 'Unknown'}
                </p>
                <p className="text-sm text-muted-foreground">Docker</p>
              </div>
            </div>
            {stats && (
              <div className="mt-3 flex items-center gap-4 text-xs text-muted-foreground">
                <span>{stats.containers.running} running</span>
                <span>{stats.images.total} images</span>
              </div>
            )}
          </div>

          {/* Worker Pool Status */}
          <div className="card p-4">
            <div className="flex items-center gap-3">
              <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
                batches && batches.running_batches > 0
                  ? 'bg-cyan-500/20'
                  : 'bg-gray-500/20'
              }`}>
                <Users className={`w-6 h-6 ${
                  batches && batches.running_batches > 0
                    ? 'text-cyan-400'
                    : 'text-gray-400'
                }`} />
              </div>
              <div>
                <p className="text-lg font-semibold text-foreground">
                  {batches?.total_workers || 0} Workers
                </p>
                <p className="text-sm text-muted-foreground">Worker Pool</p>
              </div>
            </div>
            {batches && (
              <div className="mt-3 flex items-center gap-4 text-xs text-muted-foreground">
                <span className="text-cyan-400">{batches.busy_workers} busy</span>
                <span>{batches.idle_workers} idle</span>
                <span>{batches.running_batches}/{batches.max_batches} batches</span>
              </div>
            )}
          </div>

          {/* Resources */}
          <div className="card p-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-purple-500/20 flex items-center justify-center">
                <Server className="w-6 h-6 text-purple-400" />
              </div>
              <div>
                <p className="text-lg font-semibold text-foreground">
                  {stats?.system.memory_usage_mb || 0} MB
                </p>
                <p className="text-sm text-muted-foreground">Memory</p>
              </div>
            </div>
            {stats && (
              <div className="mt-3 flex items-center gap-4 text-xs text-muted-foreground">
                <span>{stats.system.num_cpu} CPUs</span>
                <span>{stats.system.num_goroutines} goroutines</span>
              </div>
            )}
          </div>
        </div>

        {/* Running Batches Panel */}
        {batches && batches.batches.length > 0 && (
          <div className="card p-6">
            <h3 className="text-lg font-medium text-foreground mb-4 flex items-center gap-2">
              <Layers className="w-5 h-5 text-cyan-400" />
              Running Batches ({batches.running_batches})
            </h3>
            <div className="space-y-3">
              {batches.batches.map((batch) => (
                <div key={batch.id} className="p-4 rounded-lg bg-background/50 border border-default">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-3">
                      <span className="font-medium text-foreground">{batch.name || batch.id}</span>
                      <span className="text-xs text-muted-foreground px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-400">
                        {batch.workers} workers
                      </span>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <span className="text-emerald-400">{batch.completed} done</span>
                      {batch.failed > 0 && <span className="text-red-400">{batch.failed} failed</span>}
                      <span className="text-muted-foreground">/ {batch.total}</span>
                    </div>
                  </div>
                  {/* Progress bar */}
                  <div className="h-2 rounded-full bg-gray-700 overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-cyan-500 to-emerald-500 transition-all duration-300"
                      style={{ width: `${batch.percent}%` }}
                    />
                  </div>
                  <div className="flex items-center justify-between mt-2 text-xs text-muted-foreground">
                    <span>{batch.percent.toFixed(1)}% complete</span>
                    <span>{batch.tasks_per_sec.toFixed(1)} tasks/sec</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Middle Section: Sessions + Containers + Images + System */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Sessions */}
            <div className="card p-5">
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
            <div className="card p-5">
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
            <div className="card p-5">
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

            {/* System Info */}
            <div className="card p-5">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-10 h-10 rounded-lg bg-amber-500/20 flex items-center justify-center">
                  <Gauge className="w-5 h-5 text-amber-400" />
                </div>
                <div>
                  <p className="text-2xl font-bold text-foreground">{stats.system.num_cpu}</p>
                  <p className="text-sm text-muted-foreground">CPU Cores</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground flex items-center gap-1">
                    <MemoryStick className="w-3 h-3" /> Memory
                  </span>
                  <span className="text-foreground/80">{stats.system.memory_usage_mb} MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground flex items-center gap-1">
                    <Zap className="w-3 h-3" /> Goroutines
                  </span>
                  <span className="text-foreground/80">{stats.system.num_goroutines}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Go Version</span>
                  <span className="text-foreground/80">{stats.system.go_version.replace('go', '')}</span>
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
            Quick Actions
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
