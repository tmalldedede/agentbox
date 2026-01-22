import { useState, useEffect } from 'react'
import { RefreshCw, Server, Cpu, HardDrive, Activity, Container } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { api } from '@/services/api'
import type { SystemHealth, SystemStats } from '@/types'

export function ResourceMonitor() {
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const [healthData, statsData] = await Promise.all([
        api.getSystemHealth(),
        api.getSystemStats(),
      ])
      setHealth(healthData)
      setStats(statsData)
    } catch (error) {
      console.error('Failed to load system data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleRefresh = async () => {
    setRefreshing(true)
    await loadData()
    setRefreshing(false)
    toast.success('Data refreshed')
  }

  const handleCleanupContainers = async () => {
    try {
      const result = await api.cleanupContainers()
      const count = result.removed ? result.removed.length : 0
      toast.success(`Cleaned up ${count} containers`)
      loadData()
    } catch (error) {
      toast.error('Failed to cleanup containers')
    }
  }

  const handleCleanupImages = async () => {
    try {
      const result = await api.cleanupImages({ unused_only: true })
      const count = result.removed ? result.removed.length : 0
      const freedMB = ((result.space_freed || 0) / 1024 / 1024).toFixed(1)
      toast.success(`Cleaned up ${count} images, freed ${freedMB} MB`)
      loadData()
    } catch (error) {
      toast.error('Failed to cleanup images')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Activity className="h-6 w-6" />
            Resource Monitor
          </h2>
          <p className="text-muted-foreground">System health and resource usage</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={handleCleanupContainers}>
            <Container className="mr-2 h-4 w-4" />
            Cleanup Containers
          </Button>
          <Button variant="outline" onClick={handleCleanupImages}>
            <HardDrive className="mr-2 h-4 w-4" />
            Cleanup Images
          </Button>
          <Button onClick={handleRefresh} disabled={refreshing}>
            <RefreshCw className={`mr-2 h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Health Status */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">System Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge className={health?.status === 'ok' ? 'bg-green-500' : 'bg-red-500'}>
              {health?.status || 'Unknown'}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Docker Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge className={health?.docker?.status === 'connected' ? 'bg-green-500' : 'bg-red-500'}>
              {health?.docker?.status || 'Unknown'}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Uptime</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">{health?.uptime || '-'}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Docker Version</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xl font-bold">{health?.docker?.version || '-'}</p>
          </CardContent>
        </Card>
      </div>

      {/* Resource Usage */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Cpu className="h-5 w-5" />
              System Resources
            </CardTitle>
            <CardDescription>Memory and CPU information</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span>Memory Usage:</span>
              <span className="font-bold">{health?.resources?.memory_usage_mb || 0} MB</span>
            </div>
            <div className="flex justify-between">
              <span>CPUs:</span>
              <span className="font-bold">{health?.resources?.num_cpu || 0}</span>
            </div>
            <div className="flex justify-between">
              <span>Goroutines:</span>
              <span className="font-bold">{health?.resources?.num_goroutines || 0}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Docker Resources
            </CardTitle>
            <CardDescription>Container and image counts</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span>Containers:</span>
              <span className="font-bold">{health?.docker?.containers || 0}</span>
            </div>
            <div className="flex justify-between">
              <span>Images:</span>
              <span className="font-bold">{health?.docker?.images || 0}</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Sessions</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats?.sessions?.total || 0}</p>
            <p className="text-xs text-muted-foreground">{stats?.sessions?.running || 0} running</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Containers</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats?.containers?.total || 0}</p>
            <p className="text-xs text-muted-foreground">{stats?.containers?.running || 0} running</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Images</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{stats?.images?.total || 0}</p>
            <p className="text-xs text-muted-foreground">{stats?.images?.agent_images || 0} agent images</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default ResourceMonitor
